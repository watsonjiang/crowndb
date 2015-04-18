//Package cfs maintains configuration data of ramcloud system. 
//It holds the data, provides interface to change data and distributes them to all
//servers in the system.
package cfs

import (
   "bytes"
   "encoding/json"
   "errors"
   "fmt"
   log "github.com/golang/glog"
   "github.com/goraft/raft"
   "io"
   "io/ioutil"
   "net"
   "net/http"
   "os"
   "sync"
   "time"
   rc "yy.com/ramcloud"
)

var g_cfg_server *ConfigServer

/*
 * the interface for accessing config data
 */
type DataProvider interface {
   //return the config value specified by the key
   Get(key string) (string, error)
}

type DataSetter interface {
   //change the val of data entry specified by the key
   Set(key string, val string) error
}


//return all keyspace in system
func GetAllKeyspace() []string {
   return nil
}

//return bucket tbl of a keyspace
func GetBucketTbl(keyspace string) *BucketTbl {
   return nil
}

//=============BucketTbl===========
type BucketTbl struct {  //keep field public for json mashaller
   BuckNum  int
   ReplNum  int
   //array arrangement:  [[repla1, repla2],[replb1, replb2],...] (bucket_num * repl_num)
   Placement []uint64
   Ver       int
}

//CTbl is the current active server tbl
//non-zero field indicates this server is providing 
//services for the bucket.
func (t *BucketTbl) CTbl(bid int) []uint64 {
   if t.BuckNum == 0 {
      return []uint64{}
   }
   return t.Placement[0 : t.BuckNum*t.ReplNum]
}

//DTbl is the expected(designed) server distribution tbl.
//non-zero field indicates this server is designed 
//to hold the data of the bucket.
func (t *BucketTbl) DTbl(bid int) []uint64 {
   if t.BuckNum == 0{
      return []uint64{}
   }
   start := t.BuckNum * t.ReplNum * 2
   return t.Placement[start:]
}

//a container for config data. used as raft statemachine.
type config_db struct {  //internal use only
   data          map[string]string
   data_lock     sync.RWMutex
}

func (db *config_db) Save() ([]byte, error) {
   db.data_lock.RLock()
   defer db.data_lock.RUnlock()
   var tmp []byte
   var err error
   if tmp, err=json.Marshal(db.data); err!=nil{
      return nil, err
   }
   return tmp, nil
}

func (db *config_db) Recovery(b []byte) error {
   db.data_lock.RLock()
   defer db.data_lock.RUnlock()
   if err:=json.Unmarshal(b, &db.data);err!=nil{
      return err
   }
   return nil
}

//------------config server -----------
type Server struct {
   name         string
   host         string
   port         int
   path         string
   httpServeMux *http.ServeMux
   httpListener net.Listener
   raftServer   raft.Server
   db           config_db      //should hold the reference
   snapshotsched *SnapshotScheduler  //this used to compact raft log in this node.
   l_csid       uint64         //local cs id. host+port
}

func (s *Server) LocalCsId() uint64 {
   return s.l_csid
}

func (s *Server) GetProperty(key string) string {
   s.db.prop_lock.RLock()
   defer s.db.prop_lock.RUnlock()
   return s.db.prop[key]
}

func (s *Server) SetProperty(key string, value string) error {
   cmd := UpdatePropCmd{
      Key:   key,
      Val:   value,
   }
   if s.isLeader() {
      _, err := s.raftServer.Do(&cmd)
      return err
   }

   var b bytes.Buffer
   json.NewEncoder(&b).Encode(cmd)
   if err := postToRaft(fmt.Sprintf("%s/raft/update_prop",
      s.leaderConnectionString()),
      "application/json", &b); err != nil {
      return err
   }
   return nil
}

func (s *Server) BucketTbl() BucketTbl {
   s.db.tbl_lock.RLock()
   defer s.db.tbl_lock.RUnlock()
   return s.db.buck_tbl
}

//return ref of BucketTbl instead of a clone copy.
//this can avoid make slice which is expensive under heavy load.
//DO NOT try to use this interface to modify the BucketLoc table.
func (s *Server) BucketTblRef() *BucketTbl {
   return &s.db.buck_tbl
}

func (s *Server) UpdateBucketTbl(tbl *BucketTbl) error {
   s.db.tbl_lock.RLock()
   cur_tbl := s.db.buck_tbl
   s.db.tbl_lock.RUnlock()
   if tbl.Version() != cur_tbl.Version() {
      return errors.New("Can not update, version not correct.")
   }
   tbl.Ver++
   cmd := UpdateTblCmd{
      BuckTbl:     *tbl,
   }
   if s.leaderConnectionString()=="" {
      return errors.New("No leader found.")
   }
   if s.isLeader() {
      _, err := s.raftServer.Do(&cmd)
      return err
   }

   var b bytes.Buffer
   json.NewEncoder(&b).Encode(cmd)
   if err := postToRaft(fmt.Sprintf("%s/raft/update_tbl",
      s.leaderConnectionString()),
      "application/json", &b); err != nil {
      return err
   }
   return nil
}

func (s *Server) GetOrmData() ORMappingData {
   value := s.GetProperty(DS_ORM_DATA)
   var orm_data ORMappingData
   err := json.Unmarshal([]byte(value), &orm_data)
   if err != nil {
       orm_data.Init()
   }
   return orm_data
}

func (s *Server) BindOrmInfo(key string, dsn string, t ORMappingType, f ORMappingFormat) error {
   tmp_data := s.GetOrmData()
   // bind
   tmp_data.Data[key] = ORMappingInfo{
       Dsn : dsn,
       T : t,
       Format : f,
   }
   tmp_data.Ver++
   jsonStr, err := json.Marshal(tmp_data)
   if err != nil {
       return err
   }

   s.SetProperty(DS_ORM_DATA, string(jsonStr))

   return nil
}

//start the service.
func (s *Server) ListenAndServe(asLeader bool) error {
   mux := http.NewServeMux()
   mux.HandleFunc("/update_tbl", s.updateTblHandler)
   mux.HandleFunc("/update_prop", s.updatePropHandler)
   mux.HandleFunc("/add_peer", s.addPeerHandler)
   mux.HandleFunc("/rm_peer", s.rmPeerHandler)
   mux.HandleFunc("/list_peer", s.listPeerHandler)
   mux.HandleFunc("/take_snapshot", s.snapshotHandler)
   mux.HandleFunc("/list_data", s.listDataHandler)
   mux.HandleFunc("/stop", s.stopHandler)
   mux.HandleFunc("/dump_tbl", s.dumpTblHandler)
   mux.HandleFunc("/upload_tbl", s.uploadTblHandler)
   s.httpServeMux = mux
   s.startRaft(asLeader)
   lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
   if err != nil {
      log.Fatal("configserver fail to listen at ", s.connectionString(),
         "err:", err.Error())
      os.Exit(1)
      return err
   }
   s.httpListener = lis
   log.Info("configserver listening at ", s.connectionString())
   return http.Serve(s.httpListener, s.httpServeMux)
}

func validate_post_req(w http.ResponseWriter, req *http.Request) error {
   if req.Method != "POST" {
      http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
      return errors.New("Method notallowed")
   }
   return nil
}

func (s *Server) snapshotHandler(w http.ResponseWriter, req *http.Request) {
   if err := validate_post_req(w, req); err != nil {
      return
   }

   if err:=s.raftServer.TakeSnapshot(); err != nil {
      log.Errorln("Fail to take snapshot", err.Error())
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
}

func (s *Server) updateTblHandler(w http.ResponseWriter, req *http.Request) {
   if err := validate_post_req(w, req); err != nil {
      return
   }

   command := &UpdateTblCmd{}
   if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
      log.Error(err.Error())
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
   if err := postToRaft(fmt.Sprintf("%s/raft/update_tbl",
      s.leaderConnectionString()), "application/json",
      req.Body); err != nil {
      log.Error(err.Error())
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
}

func (s *Server) updatePropHandler(w http.ResponseWriter, req *http.Request) {
   if err := validate_post_req(w, req); err != nil {
      return
   }

   command := &UpdatePropCmd{}
   if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
      log.Error(err.Error())
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }

   if err := postToRaft(fmt.Sprintf("%s/raft/update_prop",
      s.leaderConnectionString()), "application/json",
      req.Body); err != nil {
      log.Error(err.Error())
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
}

func (s *Server) dumpTblHandler(w http.ResponseWriter, req *http.Request) {
   if err := validate_post_req(w, req); err != nil {
      return
   }
   tbl := s.BucketTbl()
   var b bytes.Buffer
   json.NewEncoder(&b).Encode(tbl)
   w.Header().Set("Content-Type", "application/json")
   w.Write(b.Bytes())
}

//called by rcctl
func (s *Server) uploadTblHandler(w http.ResponseWriter, req *http.Request) {
   if err:=validate_post_req(w, req); err!=nil {
      return
   }
   var tbl BucketTbl
   if err:=json.NewDecoder(req.Body).Decode(&tbl); err!=nil {
      log.Warningln("Fail to decode data.", err.Error())
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
   cur_tbl := s.BucketTblRef() //update to latest version
   tbl.Ver = cur_tbl.Version()
   if err:=s.UpdateBucketTbl(&tbl);err!=nil {
      log.Warningln("Fail to update bucket tbl.", err.Error())
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
}


func (s *Server) checkPeerExist(peer string) bool {
   if peer==s.name {
      return true
   }
   for _, p :=range s.raftServer.Peers() {
      if p.Name==peer {
         return true
      }
   }
   return false
}

//from func call
func (s *Server) AddPeer(peer string) error {

   if s.checkPeerExist(peer) {
      return fmt.Errorf("Peer already exist!", peer)
   }

   command := &raft.DefaultJoinCommand{
      Name:             peer,
      ConnectionString: fmt.Sprintf("http://%s", peer),
   }

   var b bytes.Buffer
   json.NewEncoder(&b).Encode(command)
   return postToRaft(fmt.Sprintf("%s/raft/join", s.leaderConnectionString()),
      "application/json", &b)
}

//from http
func (s *Server) addPeerHandler(w http.ResponseWriter, req *http.Request) {
   log.V(2).Info("add peer req", req)
   if err := validate_post_req(w, req); err != nil {
      return
   }

   var peer string
   if err := json.NewDecoder(req.Body).Decode(&peer); err != nil {
      log.Error("Fail to decode peer.")
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }

   if s.checkPeerExist(peer) {
      log.Errorln("Peer already exist!", peer)
      http.Error(w, "Peer already exist!", http.StatusInternalServerError)
      return
   }

   command := &raft.DefaultJoinCommand{
      Name:             peer,
      ConnectionString: fmt.Sprintf("http://%s", peer),
   }

   var b bytes.Buffer
   json.NewEncoder(&b).Encode(command)
   if err := postToRaft(fmt.Sprintf("%s/raft/join", s.leaderConnectionString()),
      "application/json", &b); err != nil {
      log.Error(err.Error())
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
}

func (s *Server) rmPeerHandler(w http.ResponseWriter, req *http.Request) {
   if err := validate_post_req(w, req); err != nil {
      return
   }

   var peer string
   if err := json.NewDecoder(req.Body).Decode(&peer); err != nil {
      log.V(2).Info("Fail to decode peer.")
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }

   if !s.checkPeerExist(peer) {
      log.Errorln("Peer not exist!", peer)
      http.Error(w, "Peer not exist!", http.StatusInternalServerError)
      return
   }

   command := &raft.DefaultLeaveCommand{
      Name: peer,
   }
   var b bytes.Buffer
   json.NewEncoder(&b).Encode(command)
   if err := postToRaft(fmt.Sprintf("%s/raft/leave",
      s.leaderConnectionString()), "application/json", &b); err != nil {
      log.Error(err.Error())
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
}

func (s *Server) getPeerList() []map[string]string {
   var peers []map[string]string
   leader := s.raftServer.Leader()
   //1 add self first. raft peer list does not contain itself
   p:=make(map[string]string)
   p["name"] = s.raftServer.Name()
   p["lat"] = time.Now().String()
   if leader == s.raftServer.Name() {
      p["leader"] = "L"
   }
   peers = append(peers, p)
   //add other peers in raft
   for _, v := range s.raftServer.Peers() {
      p = make(map[string]string)
      p["name"] = v.Name
      p["lat"] = v.LastActivity().String()
      if leader == v.Name {
         p["leader"] = "L"
      }
      peers = append(peers, p)
   }
   return peers
}

func (s *Server) listPeerHandler(w http.ResponseWriter, req *http.Request) {
   if err := validate_post_req(w, req); err != nil {
      return
   }

   peers := s.getPeerList()
   var b bytes.Buffer
   json.NewEncoder(&b).Encode(peers)
   w.Header().Set("Content-Type", "application/json")
   w.Write(b.Bytes())
}

func postToRaft(url string, cnt_type string, body io.Reader) error {
   log.V(2).Infoln("post to raft leader", url, cnt_type, body)
   rsp, err := http.Post(url, cnt_type, body)
   if err != nil {
      return err
   }
   defer rsp.Body.Close()
   if rsp.StatusCode == http.StatusTemporaryRedirect {
      url, err := rsp.Location()
      if err != nil {
         return err
      }
      return postToRaft(url.String(), cnt_type, body)
   } else if rsp.StatusCode != http.StatusOK {
      buf, _ := ioutil.ReadAll(rsp.Body)
      return errors.New(string(buf))
   }
   return nil
}

func (s *Server) Start(asLeader bool) error {
   return s.ListenAndServe(asLeader)
}

func removeDir(path string) {
   os.RemoveAll(path)
}

func (s *Server) stopHandler(w http.ResponseWriter, req *http.Request) {
   if req.Method != "POST" {
      http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
   }
   fmt.Println("stop!!!!")
   os.Exit(0)
}

func (s *Server) listDataHandler(w http.ResponseWriter, req *http.Request) {
   if req.Method != "GET" {
      http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
   }

   ret := map[string]string{}
   tbl := s.BucketTbl()
   tmp, err := json.Marshal(tbl)
   if err!=nil {
      log.Errorln("Fail to marshal bucket tbl.", err)
      http.Error(w, "Fail to marshal bucket tbl.", http.StatusInternalServerError)
      return
   }
   ret["tbl"] = string(tmp)
   //prop
   tmp, err = json.Marshal(s.db.prop)
   if err!=nil {
      log.Errorln("Fail to marshal property map.", err)
      http.Error(w, "Fail to marshal property map.", http.StatusInternalServerError)
      return
   }
   ret["prop"] = string(tmp)
   var b bytes.Buffer
   json.NewEncoder(&b).Encode(ret)
   w.Header().Set("Content-Type", "application/json")
   w.Write(b.Bytes())
}

func (s *Server) Stop() {
   s.httpListener.Close()
   //s.ml.Stop()
   //s.mon.Stop()
   s.snapshotsched.Stop()
}

func (s Server) connectionString() string {
   return fmt.Sprintf("http://%s:%d", s.host, s.port)
}

func (s Server) leaderConnectionString() string {
   if s.isLeader() {
      return s.connectionString()
   }
   p, found := s.raftServer.Peers()[s.raftServer.Leader()]
   if found {
      return p.ConnectionString
   }
   return ""
}

func (s Server) isLeader() bool {
   return s.raftServer.Leader() == s.raftServer.Name()
}

func initConfigServerDb(db *ConfigServerDb) {
   db.buck_tbl.Placement = []uint64{}
   db.prop = make(map[string]string)
}

func NewConfigServer(host string, port int, data_path string) *Server {
   name := fmt.Sprintf("%s:%d", host, port)
   server := Server{
      name:         name,
      host:         host,
      port:         port,
      path:         data_path,
      clusterState: NOT_IN_CLUSTER,
      l_csid:       rc.ConnStr2uint64(name),
   }
   initConfigServerDb(&server.db)
   //server.ml = NewMasterLogic(&server)
   //server.mon = NewMonitor(&server)
   server.snapshotsched = &SnapshotScheduler{
                               server : &server,
                          }
   //raft.SetLogLevel(raft.Trace)
   return &server
}
