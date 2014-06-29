package mempool

import (
   //"sync"
   //"container/list"
   "fmt"
   "bytes"
   "strconv"
   log "github.com/watsonjiang/crowndb/logging"
)

const (
   MAX_ALLOCATOR_IDX = 256
   MEMPOOL_LOGGER_ID = 0x0001
)

var logger log.Logger = log.GetLogger(MEMPOOL_LOGGER_ID)

type mempool_t struct {
   allocators [MAX_ALLOCATOR_IDX]allocator_t
   alloc_factor float32
   idx_smallest int
   idx_largest int
   mem_limit size_t
   mem_allocated size_t
   mem_requested size_t
   free_slab_list *slab_t //free slab list
   is_prealloc  bool      //true - prealloc mode.
}

type MemPool interface {
   ItemAlloc(nkey int, nval int) Item
   ItemRealloc(it Item, new_nkey int, new_nval int) Item
   ItemFree(it Item)
}

const (
   ITEM_TYPE_STR = iota
)

type Item interface {
   SetKV(key string, data []byte)
   GetKV() (string, []byte)
}

func NewPool(limit size_t, factor float32, prealloc bool) MemPool {
   m := &mempool_t{}
   m.mem_limit = limit
   m.init_allocators(factor)
   if prealloc {
      m.prealloc_mem()
   }
   return m
}

//prealloc memory
func (m *mempool_t) prealloc_mem() {
   /* prealloc slabs */
   num_slabs := int(m.mem_limit / size_t(SLAB_SIZE))
   if m.mem_limit % SLAB_SIZE != 0 {
      num_slabs++
   }
   for i:=0; i<num_slabs; i++ {
      m.slab_free(m.do_slab_alloc())
   }
}

//init internal data
func (m *mempool_t) init_allocators(factor float32) {
   m.alloc_factor = factor
   size := ITEM_HEAD_SIZE + CHUNK_SIZE  //init item size
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
func (m mempool_t) allocator_idx(size size_t) int {
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

func (m *mempool_t) ItemAlloc(nkey int, nval int) Item {
   nsize := item_size(size_t(nkey), size_t(nval))
   idx := m.allocator_idx(nsize)
   it := m.allocators[idx].alloc_item()
   if it == nil {
      slab := m.slab_alloc()
      if slab == nil {
         //out of memory
         logger.Warn("MemPool out of memory")
         return nil
      }
      m.allocators[idx].add_slab(slab)
      it = m.allocators[idx].alloc_item()
   }
   it.nkey = size_t(nkey)
   it.nval = size_t(nval)
   m.mem_requested += size_t(nkey+nval)
   return it
}

func (m *mempool_t) ItemRealloc(it Item, new_nkey int, new_nval int) Item {
   return nil
}

func (m *mempool_t) ItemFree(it Item) {
   item := it.(*item_t)
   size := item_size(item.nkey, item.nval)
   idx := m.allocator_idx(size)
   m.mem_requested -= size_t(item.nkey+item.nval)
   m.allocators[idx].free_item(item)
}

/* dump the allocator_class table for debug purpose. */
func (m *mempool_t) Info(verbose int) string{
   var buf bytes.Buffer
   buf.WriteString("----MemPool----\n")
   buf.WriteString("size_limit:"+strconv.Itoa(int(m.mem_limit))+"\n")
   buf.WriteString("alloc_factor:"+fmt.Sprintf("%.2f", m.alloc_factor)+"\n")
   per := float32(m.mem_requested)/float32(m.mem_allocated)*100.0
   buf.WriteString("requested/allocated:"+strconv.Itoa(int(m.mem_requested))+
                   "/"+strconv.Itoa(int(m.mem_allocated))+fmt.Sprintf("(%.2f%%)\n", per))
   if verbose > 0 {
      for i:=m.idx_smallest;i<m.idx_largest;i++ {
         buf.WriteString(fmt.Sprintf("  allocator %d size %d\n", i, m.allocators[i].size))
      }
   }
   return buf.String()
}


