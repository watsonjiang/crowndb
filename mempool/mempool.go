package mempool

import (
   "sync"
   "container/list"
   "unsafe"
)

type MemPool struct {
   slabclass [MAX_NUM_OF_SLAB_CLS]slabclass_t
   power_smallest int
   power_largest int
   mem_limit size_t
   mem_alloced size_t
   free_slab_list list.List
   slabs_lock sync.Mutex
}

func NewPool(limit size_t, factor float32, prealloc bool) *MemPool {
   m := &MemPool{}
   m.slab_init(limit, factor, prealloc)
   return m
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
