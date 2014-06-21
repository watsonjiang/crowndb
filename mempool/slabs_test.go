package mempool

import (
   "fmt"
   "testing"
   //"unsafe"
)

var _slabs_inited bool = false

func Test_slabs_clsid(t *testing.T) {
   mp := NewPool(1024 * 1024, 1.2, false)
   fmt.Println("clsid for 32 bytes:", mp.slab_clsid(32))
   fmt.Println("clsid for 64 bytes:", mp.slab_clsid(64))
   fmt.Println("clsid for 128 bytes:", mp.slab_clsid(128))
   mp.slab_cls_info_dump()
}
/*
func Test_SlabsAllocFree(t *testing.T) {
   mp := NewPool(1024 * 1024, 1.2, false)
   p1 := mp.SlabsAlloc(32, 0)
   p2 := SlabsAlloc(128, 1)
   fmt.Println("clsid", 0, "requested", slabclass[0].requested)
   fmt.Println("clsid", 1, "requested", slabclass[1].requested)
   SlabsFree(p1, 32, 0)
   SlabsFree(p2, 128, 1)
   fmt.Println("clsid", 0, "requested", slabclass[0].requested)
   fmt.Println("clsid", 1, "requested", slabclass[1].requested)
}
*/
