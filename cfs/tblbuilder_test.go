package configserver

import (
   "testing"
   "fmt"
)

func TestBuildBucketLocTbl(t *testing.T) {
   live_svrs := []uint64{0x1, 0x2, 0x3}
   repl_num := 2
   buck_num := 3
   rst,err := BuildBucketLocTbl(live_svrs, buck_num, repl_num)
   if err!=nil {
      t.Error(err.Error())
      return
   }
   fmt.Println("rst", rst)
}

