package mempool

import (
   "fmt"
   "testing"
   //"unsafe"
)

var _slabs_inited bool = false


func init_slabs() {
   if !_slabs_inited {
      SlabInit(1024 * 1024, 1.2, true)
      slabs_cls_info_dump()
      _slabs_inited = true
   }
}

func Test_slabs_clsid(t *testing.T) {
   init_slabs()
   fmt.Println("clsid for 32 bytes:", slabs_clsid(32))
   fmt.Println("clsid for 64 bytes:", slabs_clsid(64))
   fmt.Println("clsid for 128 bytes:", slabs_clsid(128))
}

func Test_SlabsAllocFree(t *testing.T) {
   init_slabs()
   p1 := SlabsAlloc(32, 0)
   p2 := SlabsAlloc(128, 1)
   fmt.Println("clsid", 0, "requested", slabclass[0].requested)
   fmt.Println("clsid", 1, "requested", slabclass[1].requested)
   SlabsFree(p1, 32, 0)
   SlabsFree(p2, 128, 1)
   fmt.Println("clsid", 0, "requested", slabclass[0].requested)
   fmt.Println("clsid", 1, "requested", slabclass[1].requested)
}

