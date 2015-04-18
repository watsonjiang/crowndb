package configserver
//this file contains the logic that monitor the server
//and update C table accordingly.
//this logic only be activated on leader node.
/*
import (
   log "github.com/golang/glog"
   "time"
)

const (
   SERVER_DOWNTIME_THRESHOLD = 4 * time.Second
   MONITOR_CHECK_INTERVAL = 1 * time.Second
)

type monitor struct {
   cfs ConfigServer
   stop chan int
}

func NewMonitor(cfs ConfigServer) *monitor{
   m := &monitor{
           cfs : cfs,
           stop : make(chan int, 1),
        }
   return m
}

func (m *monitor) Start() {
   go m.run()
}

func (m *monitor) Stop() {
   m.stop<-1
}

func (m *monitor) isServerDown(p map[string]string, t time.Time) bool {
   return t.Sub(ds.LastActTime()) > SERVER_DOWNTIME_THRESHOLD
}

func (m *monitor) updateStatus() {
   now := time.Now()
   //get all server status
   var downservers []uint64
   for _, p := range m.cfs.getPeerList() {
      if m.isServerDown(p, now) {
         downservers = append(downservers, ds.Addr())
      }
   }
   //mark all down server to status 0 in C table. 
   tbl := m.cfs.BucketTbl()
   C := tbl.CTbl()
   var tbl_changed = false
   for _, ds:=range downservers {
      for i,s:=range C{
         if s==ds {
            C[i] = 0
            tbl_changed = true
         }
      }
   }
   if tbl_changed {
      if err:=m.cfs.UpdateBucketTbl(&tbl);err==nil{
         log.Infoln("change tbl success. ver", tbl.Ver+1)
      }else{
         log.Warning("change tbl fail.", err.Error())
      }
   }
}

func (m *monitor) run() {
   for {
     select {
        case <-m.stop:
           return
        case <-time.After(MONITOR_CHECK_INTERVAL):
           //m.updateStatus()  disable update ctbl, ctbl change should be manually
     }
   }
}
*/
