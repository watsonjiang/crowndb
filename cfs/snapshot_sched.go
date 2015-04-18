package configserver

import (
    "time"
    log "github.com/golang/glog"
)

type SnapshotScheduler struct {
   server *Server
   stop chan int
}

func (ss *SnapshotScheduler) Start() {
   go ss.schedule_loop()
}

func (ss *SnapshotScheduler) Stop() {
   ss.stop <- 1
}

func (ss *SnapshotScheduler) schedule_loop() {
   for {
      select {
      case <-time.After(1*time.Hour):
         if err:=ss.server.raftServer.TakeSnapshot();err!=nil {
            log.Errorln("Fail to take snapshot.", err.Error())
         }
         log.V(0).Infoln("Snapshot done.")
      case <-ss.stop:
         break
      }
   }
}
