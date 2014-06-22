package mempool

import (
   "testing"
   //"unsafe"
)

//test slab alloc with empty free slab list
func TestSlabsAlloc(t *testing.T) {
   mp := &MemPool{}
   mp.free_slab_list = nil
   s := mp.slab_alloc()
   if s==nil || mp.mem_allocated != SLAB_SIZE || mp.free_slab_list != nil {
      t.Log("s ", s)
      t.Log("mem_allocated ", mp.mem_allocated)
      t.Log("free_slab_list ", mp.free_slab_list)
      t.Error("slab_alloc not correct")
   }
}

//test slab alloc with non-empty free slab list
func TestSlabsAlloc1(t *testing.T) {
   mp := &MemPool{}
   t0 := &slab_t{}
   t1 := &slab_t{}
   t0.next = t1
   t0.ptr = nil
   mp.free_slab_list = t0
   s := mp.slab_alloc()
   if s!=t0 || mp.mem_allocated != SLAB_SIZE || mp.free_slab_list != t1 {
      t.Log("s ", s, "t ", t)
      t.Log("mem_allocated ", mp.mem_allocated)
      t.Log("free_slab_list ", mp.free_slab_list)
      t.Error("slab_alloc not correct")
   }
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
