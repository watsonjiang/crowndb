package mempool

import (
   "testing"
   //"unsafe"
   "fmt"
   "strconv"
   "github.com/watsonjiang/crowndb/log"
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

func BenchmarkAllocFree(t *testing.B) {

}

//an example of how to use mempool
func TestExample(t *testing.T) {
   l := logging.GetLogger(MEMPOOL_LOGGER_ID)
   l.SetLevel(logging.L_DEBUG)
   var tmp []byte
   for i:=0;i<1024;i++ {
      tmp = append(tmp, byte('a'))
   }
   mp := NewPool(size_t(1024*1024*1024), 1.2, false)
   //print pool statistic info
   fmt.Println(mp.(*mempool_t).Info(1))
   //allocate items
   var items []Item
   for i:=0;i<1024*1024*10;i++ {
      key := "testkey"+strconv.Itoa(i)
      value := "testvalue"+string(tmp[0:i % 1024])
      it := mp.ItemAlloc(len(key), len(value))
      if it == nil {
         t.Error("mp.ItemAlloc not correct")
         return
      }
      it.SetKV([]byte(key), []byte(value))
      items = append(items, it)
   }
   fmt.Println(mp.(*mempool_t).Info(1))
   //free items
   for _, it:=range items {
      mp.ItemFree(it)
   }
   fmt.Println(mp.(*mempool_t).Info(1))
}
