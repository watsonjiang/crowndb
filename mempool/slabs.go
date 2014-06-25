package mempool

/* a golang copy of memchached slab allocator */

import (
      "unsafe"
      )
//#cgo LDFLAGS:-fPIC -m64
//#include <stdlib.h>
//#include <string.h>
import "C"

type size_t uint64

type slab_t struct {
   next *slab_t         //next slab item
   //prev *slab         //prev slab item
   ptr  unsafe.Pointer     //pointer to slab
}

const (
   SLAB_SIZE = size_t(1024 * 1024)
   CHUNK_SIZE = size_t(32)
   CHUNK_ALIGN_BYTES = 8
)

//get slab from free list or system malloc
//return nil if out of memory
func (m *mempool_t) slab_alloc() *slab_t {
   if m.free_slab_list == nil {
      if m.is_prealloc {
         return nil
      }else{
         if m.mem_allocated + size_t(SLAB_SIZE) > m.mem_limit {
            return nil
         }
      }
      newslab := m.do_slab_alloc()
      newslab.next = m.free_slab_list
      m.free_slab_list = newslab
   }
   t := m.free_slab_list
   m.free_slab_list = t.next
   return t
}

//free slab to free list.
func (m *mempool_t) slab_free(s *slab_t) {
   s.next = m.free_slab_list
   m.free_slab_list = s
}

func (m *mempool_t) do_slab_alloc() *slab_t {
   s := &slab_t{}
   s.ptr = C.malloc(C.size_t(SLAB_SIZE))
   m.mem_allocated += size_t(SLAB_SIZE)
   return s
}
