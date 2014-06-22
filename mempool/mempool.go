package mempool

import (
   //"sync"
   //"container/list"
   "unsafe"
   log "github.com/watsonjiang/crowndb/logging"
)

const (
   MAX_ALLOCATOR_IDX = 256
   MEMPOOL_LOGGER_ID = 0x0001
)

var logger log.Logger = log.GetLogger(MEMPOOL_LOGGER_ID)

/* allocator structure */
type allocator_t struct {
   size size_t              //size of items in this allocator
   used_item_list *Item     //used item list
   free_item_list *Item     //free item list
   slab_list *slab_t        //used slab list 
}

type MemPool struct {
   allocators [MAX_ALLOCATOR_IDX]allocator_t
   idx_smallest int
   idx_largest int
   mem_limit size_t
   mem_allocated size_t
   mem_requested size_t
   free_slab_list *slab_t //free slab list
   is_prealloc  bool      //true - prealloc mode.
}

func NewPool(limit size_t, factor float32, prealloc bool) *MemPool {
   m := &MemPool{}
   m.init(limit, factor)
   if prealloc {
      /* prealloc slabs */
      num_slabs := int(limit / size_t(SLAB_SIZE))
      if limit % SLAB_SIZE != 0 {
         num_slabs++
      }
      for i:=0; i<num_slabs; i++ {
         m.slab_free(m.slab_alloc())
      }
   }
   return m
}

func (m *MemPool) init(limit size_t, factor float32) {
   size := item_size(0, CHUNK_SIZE)  //init item size
   for i:=0;i<MAX_ALLOCATOR_IDX;i++ {
      /*make sure items are always n-byte aligned */
      if size % CHUNK_ALIGN_BYTES != 0 {
         size += CHUNK_ALIGN_BYTES - (size % CHUNK_ALIGN_BYTES)
      }
      if size > SLAB_SIZE {
         /*max chunk size is SLAB_SIZE*/
         m.allocators[i].size = 0  //use 0 size as end flag
         m.idx_largest = i-1
         break
      }
      m.allocators[i].size = size
      size = size_t(float64(size) * float64(factor))
   }
}

/* Figures out which pool class is required to store an item
   of given size
*/
func (m MemPool) allocator_idx(size size_t) int {
   if size == 0 {
      return -1
   }
   res := 0
   for ;size > m.allocators[res].size; {
      res++
      if res == MAX_ALLOCATOR_IDX {
         return -1
      }
   }
   return res
}

func (m *MemPool) ItemAlloc(nkey int, nvalue int) *Item {
   size := item_size(size_t(nkey), size_t(nvalue))
   idx := m.allocator_idx(size)
   it := m.allocators[idx].item_alloc(nkey, nvalue)
   if it == nil {
      slab := m.slab_alloc()
      if slab == nil {
         //out of memory
         logger.Warn("MemPool out of memory")
         return nil
      }
      m.allocators[idx].add_slab(slab)
      it = m.allocators[idx].item_alloc(nkey, nvalue)
   }
   return it
}

func (m *MemPool) ItemFree(it *Item) {
   size := item_size(it.nkey, it.nval)
   idx := m.allocator_idx(size)
   m.allocators[idx].item_free(it)
}

/* dump the allocator_class table for debug purpose. */
func (m *MemPool) dump(){
   for i:=m.idx_smallest;i<m.idx_largest;i++ {
      logger.Debugln("allocator", i, "size", m.allocators[i].size)
   }
}

// split an empty slab into piceses, link them into free
// item list
func (a *allocator_t) add_slab(s *slab_t) {
   perslab := SLAB_SIZE / int(a.size)
   var tmp *[SLAB_SIZE]byte = (*[SLAB_SIZE]byte)(s.ptr)
   for x:=0;x<perslab;x++ {
      var item *Item = (*Item)(unsafe.Pointer(&tmp[int(a.size) * x]))
      item.next = a.free_item_list
      a.free_item_list.prev = item.next
      a.free_item_list = item
      item.prev = nil
   }
   s.next = a.slab_list
   a.slab_list = s
}

func (a *allocator_t)item_alloc(nkey int, nvalue int) *Item{
   var it *Item
   //get one from free list
   if a.free_item_list == nil {
      return nil    //return nil when full
   }else{
      it = a.free_item_list
      a.free_item_list=it.next
   }
   //put it into used list
   it.next = a.used_item_list
   if a.used_item_list != nil {
      a.used_item_list.prev = it
   }
   it.prev = nil
   a.used_item_list = it
   /* initialize the item */
   it.refcount = 1
   it.nkey = size_t(nkey)
   it.nval = size_t(nvalue)
   return it
}

func (a *allocator_t)item_free(it *Item) {
  //unlink the item from used list
  if it.prev == nil { //head of used list
     a.used_item_list = it.next
     if a.used_item_list != nil {
        a.used_item_list.prev = nil
     }
  }else {  //in the middle of used list
     it.prev.next = it.next
     if it.next != nil {
        it.next.prev = it.prev
     }
  }
  //link it into free list
  it.next = a.free_item_list
  a.free_item_list.prev = it.next
  a.free_item_list = it
  a.free_item_list.prev = nil
}
