package configserver

import (
   "encoding/json"
   "errors"
   "fmt"
   log "github.com/golang/glog"
   "github.com/goraft/raft"
   "net/http"
   "os"
   "time"
   rc "yy.com/ramcloud"
   //"bytes"
)

func init() {
   raft.RegisterCommand(&UpdateTblCmd{})
   raft.RegisterCommand(&UpdatePropCmd{})
}

//------------raft commands-------------
/*
 * Commands used by raft.
 */
type UpdateTblCmd struct { //keep fields public for json marshaller
   BuckTbl  BucketTbl
}

func (t UpdateTblCmd) CommandName() string {
   return "UpdateBucketLocTblCommand"
}

func (t *UpdateTblCmd) Apply(server raft.Server) (interface{}, error) {
   db := server.StateMachine().(*ConfigServerDb)
   db.tbl_lock.Lock()
   defer db.tbl_lock.Unlock()
   db.buck_tbl = t.BuckTbl
   log.V(1).Infoln("UpdateTblCmd version", t.BuckTbl.Ver)
   if log.V(2) {
      log.V(2).Infoln("Buck", t.BuckTbl.BucketNum(), "Repl", t.BuckTbl.ReplicaNum(), "Tbl", rc.DumpTbl(t.BuckTbl.Placement))
   }
   return nil, nil
}

type UpdatePropCmd struct { //keep fileds public for json marshaller
   Key   string
   Val string
}

func (t UpdatePropCmd) CommandName() string {
   return "UpdatePropCommand"
}

func (t *UpdatePropCmd) Apply(server raft.Server) (interface{}, error) {
   db := server.StateMachine().(*ConfigServerDb)
   db.prop_lock.Lock()
   defer db.prop_lock.Unlock()
   db.prop[t.Key] = t.Val
   log.V(1).Info("UpdatePropCmd key", t.Key, "value", t.Val)
   return nil, nil
}

// functions used to adjust dataserver array
// when there's peer change in raft.
func (s *Server) onAddPeer(e raft.Event) {
   if e.Type() == raft.AddPeerEventType {
      pn := e.Value().(string) //only remote peers have notification.
      log.V(0).Info("Add Peer ", pn)
   }
}

func (s *Server) onRemovePeer(e raft.Event) {
   if e.Type() == raft.RemovePeerEventType {
      pn := e.Value().(string)
      log.V(0).Info("Remove Peer ", pn)
   }
}

func (s *Server) onLeaderChange(e raft.Event) {
   if e.Type() == raft.LeaderChangeEventType {
      log.V(0).Info("Leader changed. ", e.PrevValue(), "->", e.Value())
      if e.Value() == s.raftServer.Name() {
         //been promoted as leader.
         log.V(0).Info("Been promoted as leader. Start monitor.")
         //s.ml.Start()
         //s.mon.Start()
      } else if e.PrevValue() == s.raftServer.Name() {
         log.V(0).Info("Been depromoted, stop monitor.")
         //s.ml.Stop()
         //s.mon.Stop()
      }
   }
}

func (s *Server) onCommit(e raft.Event) {
   cmd_name := e.Value().(*raft.LogEntry).CommandName()
   cmd_value := e.Value().(*raft.LogEntry).Command()
   log.V(2).Infoln("onCommit cmd", cmd_name, "value", string(cmd_value))
}

//return true if redirected.
func (s Server) redirectToLeader(path string, w http.ResponseWriter, req *http.Request) bool {
   if !s.isLeader() {
      url := fmt.Sprintf("http://%s%s", s.leaderConnectionString(), path)
      log.V(2).Info("Redirect request to ", url)
      http.Redirect(w, req, url, http.StatusTemporaryRedirect)
      return true
   }
   return false
}

func (s *Server) joinHandler(w http.ResponseWriter, req *http.Request) {
   log.V(2).Info("raft join req", req)
   if s.redirectToLeader("/raft/join", w, req) {
      return
   }

   command := &raft.DefaultJoinCommand{}

   if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
      log.Error(err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }

   if _, err := s.raftServer.Do(command); err != nil {
      log.Error(err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }

}

func (s *Server) leaveHandler(w http.ResponseWriter, req *http.Request) {
   log.V(2).Info("raft leave req", req)
   if s.redirectToLeader("/raft/leave", w, req) {
      return
   }

   command := &raft.DefaultLeaveCommand{}
   if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
      log.Error(err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
   if _, err := s.raftServer.Do(command); err != nil {
      log.Error(err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
}

func (s *Server) raftUpdateTblHandler(w http.ResponseWriter, req *http.Request) {
   log.V(2).Info("raftUpdateTbl req", req)
   if s.redirectToLeader("/raft/update_tbl", w, req) {
      return
   }

   command := &UpdateTblCmd{}
   if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
      log.Error(err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
   if _, err := s.raftServer.Do(command); err != nil {
      log.Error(err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
}

func (s *Server) raftUpdatePropHandler(w http.ResponseWriter, req *http.Request) {
   log.V(2).Info("raftUpdateProp req", req)
   if s.redirectToLeader("/raft/update_prop", w, req) {
      return
   }

   command := &UpdatePropCmd{}
   if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
      log.Error(err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
   if _, err := s.raftServer.Do(command); err != nil {
      log.Error(err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
}

func (s *Server) raftSnapshotHandler(w http.ResponseWriter, req *http.Request) {
   log.V(2).Info("raftSnapshot req", req)
   if s.redirectToLeader("/raft/take_snapshot", w, req) {
      return
   }

   if err := s.raftServer.TakeSnapshot(); err != nil {
      log.Error(err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
   }
}

func (s *Server) startRaft(asLeader bool) error {
   var err error
   mux := s.httpServeMux
   if s.clusterState == IN_CLUSTER {
      return errors.New("Already in cluster")
   }
   if _,err:=os.Stat(s.path);os.IsNotExist(err){ //create folder if not exist
      if err=os.MkdirAll(s.path, 0755);err!=nil {
         return fmt.Errorf("Fail to create folder %v", s.path)
      }
   }
   transporter := raft.NewHTTPTransporter("/raft", 200*time.Millisecond)
   s.raftServer, err = raft.NewServer(s.name, s.path, transporter, &s.db, nil, "")
   if err != nil {
      log.Errorf("Fail to create new server. %v", err)
      return err
   }
   transporter.Install(s.raftServer, mux)
   mux.HandleFunc("/raft/update_tbl", s.raftUpdateTblHandler)
   mux.HandleFunc("/raft/update_prop", s.raftUpdatePropHandler)
   mux.HandleFunc("/raft/take_snapshot", s.raftSnapshotHandler)
   mux.HandleFunc("/raft/join", s.joinHandler)
   mux.HandleFunc("/raft/leave", s.leaveHandler)
   s.raftServer.AddEventListener(raft.AddPeerEventType, s.onAddPeer)
   s.raftServer.AddEventListener(raft.RemovePeerEventType, s.onRemovePeer)
   s.raftServer.AddEventListener(raft.LeaderChangeEventType, s.onLeaderChange)
   s.raftServer.AddEventListener(raft.CommitEventType, s.onCommit)
   s.raftServer.Start()
   s.snapshotsched.Start() //all node should start this schedule.
                           //it compacts log every 1 hour
   if asLeader {
      if !s.raftServer.IsLogEmpty() {
         log.Fatalln("Can not start as leader with an existing log.")
      }
      log.Info("Initializing new cluster")
      _, err = s.raftServer.Do(&raft.DefaultJoinCommand{
         Name:             s.raftServer.Name(),
         ConnectionString: s.connectionString(),
      })
      if err != nil {
         log.Error(err)
         return err
      }
   }else {
      if !s.raftServer.IsLogEmpty() {
         log.Infoln("Recovering bucket tbl from log.")
      }
   }
   s.clusterState = IN_CLUSTER
   return nil
}
