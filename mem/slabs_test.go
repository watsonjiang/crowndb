package mempool

import (
   "testing"
   //"unsafe"
)

func TestDoSlabAlloc(t *testing.T) {
   mp := &mempool_t{}
   slab := mp.do_slab_alloc()
   if slab == nil ||
      mp.mem_allocated != SLAB_SIZE {
         t.Log("slab ", slab, "mem_allocated", mp.mem_allocated)
         t.Error("do_slab_alloc not correct")
   }
}

//test slab alloc with empty free slab list
func TestSlabsAlloc(t *testing.T) {
   mp := &mempool_t{}
   mp.mem_limit = 1024*1024
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
   mp := &mempool_t{}
   mp.mem_limit = 1024 * 1024
   t0 := &slab_t{}
   t1 := &slab_t{}
   t0.next = t1
   t0.ptr = nil
   mp.free_slab_list = t0
   s := mp.slab_alloc()
   if s!=t0 || mp.mem_allocated != 0 || mp.free_slab_list != t1 {
      t.Log("s", s, "t", t0)
      t.Log("mem_allocated", mp.mem_allocated)
      t.Log("free_slab_list", mp.free_slab_list)
      t.Error("slab_alloc not correct")
   }
}

//test slab alloc with OOM, non-prealloc
func TestSlabsAlloc2(t *testing.T) {
   mp := &mempool_t{}
   mp.free_slab_list = nil
   mp.mem_allocated = 0
   mp.mem_limit = 0
   s := mp.slab_alloc()
   if s!=nil {
      t.Log("s", s)
      t.Error("slab_alloc not correct")
   }
}

//test slab alloc with OOM, prealloc
func TestSlabsAlloc3(t *testing.T) {
   mp := &mempool_t{}
   mp.is_prealloc = true
   mp.free_slab_list = nil 
   mp.mem_allocated = 0
   mp.mem_limit = 1024*1024
   s := mp.slab_alloc()
   if s!=nil {
      t.Log("s", s)
      t.Error("slab_alloc not correct")
   }
}


func Test_SlabsAllocFree(t *testing.T) {
   s := &slab_t{}
   mp := &mempool_t{}
   mp.slab_free(s)
   if mp.free_slab_list != s {
      t.Log("mp.free_slab_list", mp.free_slab_list, "s", s)
      t.Error("slab_free not correct")
   }
}

