package mempool

import (
   "testing"
   //"unsafe"
   "fmt"
   "strconv"
)

func countSlabList(list *slab_t) int {
   count := 0
   p := list
   for ;p != nil; {
      count += 1
      p = p.next
   }
   return count
}

func TestPreallocMem(t *testing.T) {
   mp := &mempool_t{}
   mp.mem_limit = 3*SLAB_SIZE+8   //should have 4 slabs
   mp.prealloc_mem()
   count := countSlabList(mp.free_slab_list)
   if count != 4 {
      t.Log("count", count)
      t.Error("mempool.prealloc_mem not correct")
   }
   if mp.mem_allocated != 4*SLAB_SIZE {
      t.Log("mp.mem_allocated", mp.mem_allocated)
      t.Error("mempool.prealloc_mem not correct")
   }
}

func TestInitAllocators(t *testing.T) {
   mp := &mempool_t{}
   mp.init_allocators(1024)      //start from 2^5, up to 2^20?? 3 allocators
   if mp.allocators[0].size == 0 {
      t.Error("mempool.init_allocators not correct")
   }
   if mp.allocators[2].size != 0 {
      t.Error("mempool.init_allocators not correct")
   }
}

func TestAllocatorIdx(t *testing.T) {
   mp := &mempool_t{}
   mp.init_allocators(2)
   if mp.allocator_idx(size_t(0))!=-1 {
      t.Log("idx 0", mp.allocator_idx(size_t(0)))
      t.Error("mempool.allocator_idx not correct")
   }
   if mp.allocator_idx(item_size(0, 16))!=0 {
      t.Log("idx 32", mp.allocator_idx(size_t(32)))
      t.Error("mempool.allocator_idx not correct")
   }
   if mp.allocator_idx(size_t(SLAB_SIZE))!=-1 {
      t.Log("idx ", SLAB_SIZE, " ", mp.allocator_idx(size_t(SLAB_SIZE)))
      t.Error("mempool.allocator_idx not correct")
   }
}

//an example of how to use mempool
func TestExample(t *testing.T) {
   var tmp []byte
   for i:=0;i<1024;i++ {
      tmp = append(tmp, byte(i % 128))
   }
   mp := NewPool(size_t(100*1024*1024), 1.2, false)
   fmt.Println(mp.(*mempool_t).Info(0))
   for i:=0;i<102400;i++ {
      key := "testkey"+strconv.Itoa(i)
      value := "testvalue"+string(tmp[0:i % 1024])
      it := mp.ItemAlloc(len(key), len(value))
      if it == nil {
         t.Error("mp.ItemAlloc not correct")
         return
      }
      it.SetKV(key, []byte(value))
      if i % 100 == 0 {
         fmt.Println(mp.(*mempool_t).Info(0))
      }
   }
}
