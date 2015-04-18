package configserver

import (
   "testing"
   "time"
)

//server is up
func TestUpdateServer(t *testing.T) {
   ms := &mock_cfs{}
   ms.getDataServers = func()[]DataServer {
      var ret []DataServer
      ret = append(ret, DataServerImpl{Addrs:0x7f00000100000000, lat:time.Now(),})
      return ret
   }
   var tbl BucketTbl
   tbl.Placement = []uint64{0x7f00000100000000,
                          0x7f00000100000000,
                          0x7f00000100000000}
   tbl.Ver = 0
   tbl.BuckNum = 1
   tbl.ReplNum = 1
   ms.bucketTbl = func() BucketTbl {
                       return tbl
                  }
   ms.updateBucketTbl = func(*BucketTbl) error {t.Error("Should not update tbl.");return nil}
   mon := &monitor{cfs : ms,}
   mon.updateStatus()
}

//server is down 
func TestUpdateServer1(t *testing.T) {
   ms := &mock_cfs{}
   ms.getDataServers = func()[]DataServer {
      var ret []DataServer
      ret = append(ret, DataServerImpl{Addrs:0x7f00000100000000, lat:time.Now().Add(-SERVER_DOWNTIME_THRESHOLD*2),})
      return ret
   }
   var tbl BucketTbl
   tbl.Placement = []uint64{0x7f00000100000000,
                            0x7f00000100000000,
                            0x7f00000100000000}
   tbl.BuckNum = 1
   tbl.ReplNum = 1
   ms.bucketTbl = func() BucketTbl {
                        return tbl
                     }
   ms.updateBucketTbl = func(tbl1 *BucketTbl)error {tbl = *tbl1; return nil}
   mon := &monitor{cfs : ms,}
   mon.updateStatus()
   if tbl.CTbl()[0] != 0x0 {
      t.Error("update status does not correct. tbl[0]", tbl.CTbl()[0])
   }
}

