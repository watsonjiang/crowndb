package configserver

import (
   "github.com/goraft/raft"
   "testing"
   "time"
)

func TestUpdateTblCmdApply(t *testing.T) {
   //t.Skip()
   var tbl BucketTbl
   tbl.Placement = make([]uint64, 3)
   tbl.BuckNum = 1
   tbl.ReplNum = 1
   tbl.CTbl()[0] = 3
   cmd := UpdateTblCmd{
      buck_tbl:     tbl,
   }
   db := ConfigServerDb{}
   trs := raft.NewHTTPTransporter("/raft", 200*time.Millisecond)  //mock
   s, _ := raft.NewServer("test", "test", trs, &db, nil, "")
   cmd.Apply(s)
   if len(db.buck_tbl.Placement) != 3 || db.buck_tbl.Placement[0] != 3 {
      t.Error("UpdateTblCmd.Apply not correct")
      return
   }
}

func TestUpdatePropCmdApply(t *testing.T) {
   //t.Skip()
   cmd := UpdatePropCmd{}
   cmd.key = "test"
   cmd.value = "testvalue"
   db := ConfigServerDb{
      prop: map[string]string{},
   }
   trs := raft.NewHTTPTransporter("/raft", 200*time.Millisecond) //mock
   s, _ := raft.NewServer("test", "test", trs, &db, nil, "")
   cmd.Apply(s)
   if len(db.prop) != 1 || db.prop["test"] != "testvalue" {
      t.Error("UpdatePropCmd.Apply not correct")
      return
   }
}
