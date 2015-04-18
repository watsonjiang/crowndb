package configserver

import (
   "testing"
   //   "fmt"
   //   "yy.com/ramcloud"
   //"time"
)

func newBucketLocTbl() BucketTbl {
   ret := BucketTbl{
      BuckNum : 3,
      ReplNum : 1,
      Ver  : 0,
      Placement : make([]uint64, 9),
   }
   a := uint64(0x0a00000000000000)
   for i := 0; i < 9; i++ {
      ret.Placement[i] = a
   }
   return ret
}

func newPropList() map[string]string {
   ret := map[string]string{}
   ret["key"] = "test"
   return ret
}

func newCSDb() ConfigServerDb {
   ret := ConfigServerDb{
      buck_tbl: newBucketLocTbl(),
      prop:     newPropList(),
   }
   return ret
}

func TestCSDbSaveRecovery(t *testing.T) {
   db := newCSDb()
   b, err := db.Save()
   if err != nil {
      t.Error("Fail to Save db.", err)
      return
   }
   db1 := ConfigServerDb{}
   err = db1.Recovery(b)
   if err != nil {
      t.Error("Fail to Recover db.", err)
      return
   }
   tmp := db1.buck_tbl
   if tmp.Ver != 0 ||
      tmp.BuckNum != 3 ||
      tmp.ReplNum != 1 {
      t.Error("BucketTbl not currect.", tmp)
      return
   }
   if tmp.CTbl()[0] != 0x0a00000000000000 {
      t.Error("BucketLoc not correct.", tmp)
      return
   }
   if db1.prop["key"] != "test" {
      t.Error("prop cnot correct.expecting Prop[\"key\"]=\"test\"", db1)
      return
   }
}
/*
func TestUpdateTbl(t *testing.T) {
   cs1 := NewConfigServer("127.0.0.1", 9901, "test.1", 3, 1)
   go func() {
      cs1.Start(true)
   }()
   time.Sleep(1000 * time.Millisecond)
   defer cs1.Stop()
   cs2 := NewConfigServer("127.0.0.1", 9902, "test.2", 3, 1)
   go func() {
      cs2.Start(false)
   }()
   time.Sleep(1000 * time.Millisecond)
   defer cs2.Stop()
   cs1.AddPeer("127.0.0.1:9902")
   t.Log("join success")
   time.Sleep(1000 * time.Millisecond)
   //cs1.UpdateBucketLocTbl(tmp)
   tmp, ver := cs1.BucketLocTbl()
   tmp[0] = 0x7f00000126ae0000
   if err := cs2.UpdateBucketLocTbl(tmp, ver); err != nil {
      t.Error(err.Error())
      return
   }
   t.Log("updateBucketLocTbl success")
   //cs1.SetProperty("auto_rebalance", "off")
   cs2.SetProperty("auto_rebalance", "off")
   time.Sleep(1000 * time.Millisecond)
   tmp, ver = cs1.BucketLocTbl()
   if len(tmp) != 3*3 || tmp[0] != 0x7f00000126ad0000 {
      t.Log("actual:", len(tmp), "expect:", 3*1024)
      t.Log("actual:", tmp[0], "expect:", 0x7f00000126ad0000)
      t.Error("UpdateTbl Fail.")
   }
   if "off" != cs1.GetProperty("auto_rebalance") {
      t.Error("SetProperty fail.")
   }
   tmp1 := cs1.GetDataServers()
   if len(tmp1) != 2 {
      t.Error("GetDataServers not correct")
   }
}
*/
