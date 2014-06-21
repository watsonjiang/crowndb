package mempool

import (
   "sync"
   "container/list"
   "unsafe"
)

const (
   MAX_POOL_CLS = 256
)
/* pool allocation structure */
type poolclass_t struct {
   size size_t         //size of items in this pool
   used_item_list *Item     //used item list
   free_item_list *Item     //free item list
   slab_list *slab     //slab list 
}

type MemPool struct {
   poolclass [MAX_POOL_CLS]poolclass_t
   poolcls_smallest int
   poolcls_largest int
   mem_limit size_t
   mem_alloced size_t
   free_slab_list *slab   //free slab list
}

func NewPool(limit size_t, factor float32, prealloc bool) *MemPool {
   m := &MemPool{}
   m.slab_init(limit, factor, prealloc)
   return m
}

/* Figures out which pool class is required to store an item
   of given size
*/
func (m MemPool) pool_clsid(size size_t) int {
   if size == 0 {
      return 0
   }
   res := int(0)
   for ;size > m.poolclass[res].size; {
      res++
      if res == MAX_POOL_CLS {
         return 0
      }
   }
   return res
}

/* Determines the chunk sizes and initializes the pool calss */
func (m *MemPool) Init(limit size_t, factor float32, prealloc bool) {
   item := Item{}
   size := size_t(unsafe.Sizeof(item)) + CHUNK_SIZE
   m.mem_limit = limit

   m.slab_init(limit, size, factor, prealloc)
   if prealloc {
      m.slab_preallocate()
   }
}



func (m *MemPool) ItemAlloc(key string, nvalue int) *Item {
   size := item_size(size_t(len(key)), size_t(nvalue))
   clsid := m.slab_clsid(size)
   it := (*Item)(m.slab_alloc(size, clsid))
   /* initialize the item */
   it.refcount = 1
   it.next = nil
   it.prev = nil
   it.h_next = nil
   it.slabs_clsid = clsid
   it.nkey = size_t(len(key))
   return it
}

func (m *MemPool) ItemFree(it *Item) {
   clsid := it.slabs_clsid
   it.slabs_clsid = 0
   size := item_size(it.nkey, it.nvalue)
   m.slab_free(unsafe.Pointer(it), size, clsid)
}
