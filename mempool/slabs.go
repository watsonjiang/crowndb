package mempool

/* a golang copy of memchached slab allocator */

import (
      "unsafe"
      log "github.com/watsonjiang/crowndb/logging"
      )
//#include <stdlib.h>
//#include <string.h>
import "C"

const (
   MEMPOOL_LOGGER_ID = 0x0001
)

var logger log.Logger = log.GetLogger(MEMPOOL_LOGGER_ID)

type size_t uint64

/* allocation structure */
type slabclass_t struct {
   size size_t          /*size of items*/
   perslab int        /*how many items per slab*/
   slots *Item        /*list of item ptrs*/
   sl_curr uint         /*total free itemIs in list*/
   slabs uint           /*the number of slabs for this class*/
   slab_list unsafe.Pointer /*array of slab pointers*/
   list_size uint       /*size of array above*/
   requested size_t     /*number of requested bytes*/
}

type slab struct {
   next *slab         //next slab item
   prev *slab         //prev slab item
   ptr  unsafe.Pointer     //pointer to slab
}

const (
   SLAB_SIZE = 1024 * 1024
   CHUNK_SIZE = 32
   CHUNK_ALIGN_BYTES = 8
)

func (m *MemPool) slab_init(limit size_t, base_item_size size_t, factor float32, prealloc bool) {
   if prealloc {
      /* prealloc slabs */
      num_slabs := int(limit / size_t(SLAB_SIZE))
      if limit % SLAB_SIZE != 0 {
         num_slabs++
      }
      for i:=0; i<num_slabs; i++ {
         tmp := C.malloc(C.size_t(SLAB_SIZE))
         m.free_slab_list.PushBack(tmp)
         m.mem_alloced += SLAB_SIZE
      }
   }

   C.memset(unsafe.Pointer(&m.slabclass[0]),
            0, C.size_t(unsafe.Sizeof(m.slabclass)))

   for i:=0;i<MAX_NUM_OF_SLAB_CLS;i++ {
      /*make sure items are always n-byte aligned */
      if size % CHUNK_ALIGN_BYTES != 0 {
         size += CHUNK_ALIGN_BYTES - (size % CHUNK_ALIGN_BYTES)
      }
      if size > SLAB_SIZE {
         /*max chunk size is SLAB_SIZE*/
         m.slabclass[i].size = 0
         m.slabclass[i].perslab = 0   //use 0 as end flag
         m.power_largest = i-1
         break
      }
      m.slabclass[i].size = size
      m.slabclass[i].perslab = int(size_t(SLAB_SIZE) / m.slabclass[i].size)
      size = size_t(float64(size) * float64(factor))
   }


}

func (m * MemPool) slab_alloc(size size_t, id int) unsafe.Pointer {
   m.slabs_lock.Lock()
   defer m.slabs_lock.Unlock()

   return m.do_slab_alloc(size, id)
}

func (m * MemPool) slab_free(ptr unsafe.Pointer, size size_t, id int) {
   m.slabs_lock.Lock()
   defer m.slabs_lock.Unlock()

   m.do_slab_free(ptr, size, id)
}

/* dump the slabclass table for debug purpose. */
func (m *MemPool) slab_cls_info_dump(){
   for i:=m.power_smallest;i<m.power_largest;i++ {
      logger.Debugln("class", i, "size", m.slabclass[i].size)
   }
}

func (m *MemPool) do_slab_alloc(size size_t, id int) unsafe.Pointer {
   if (id < m.power_smallest && id > m.power_largest) {
      panic("Invalid class id")
   }
   var ret unsafe.Pointer
   p:=&m.slabclass[id]
   if p.sl_curr == 0 {
      m.do_slab_newslab(id)
   }

   if p.sl_curr != 0 {
      it := p.slots;
      p.slots = it.next
      if it.next != nil {
         it.next.prev = nil
      }
      p.sl_curr--
      ret = unsafe.Pointer(it)
   }
   p.requested += size
   return ret
}

func (m *MemPool) slab_preallocate() {
   /* pre-allocate a 1mb slab in every size class.*/
   for i:=m.power_smallest;i<=m.power_largest;i++ {
      m.do_slab_newslab(i)
   }
}

func (m *MemPool) grow_slab_list(id int) {
   p := &m.slabclass[id]
   if p.slabs == p.list_size {
      var new_size uint
      if p.list_size == 0 {
         new_size = 16
      }else {
         new_size = p.list_size * 2
      }
      new_list := C.realloc(p.slab_list,
                            C.size_t(new_size*uint(unsafe.Sizeof(p.slab_list))))
      p.list_size = new_size
      p.slab_list = new_list
   }
}

func (m *MemPool) split_slab_page_into_freelist(ptr unsafe.Pointer, id int) {
   p := &m.slabclass[id]
   tmp := ptr
   logger.Debugln("p.id", id, "p.perslab", p.perslab)
   for x:=0;x<p.perslab;x++ {
       m.do_slab_free(tmp, 0, id)
       tmp = unsafe.Pointer(&C.GoBytes(tmp, SLAB_SIZE)[p.size]);
   }
}

func (m *MemPool) memory_allocate_newslab() unsafe.Pointer {
   if m.free_slab_list.Len() == 0 {
      /* no available slab in free slab list */
      m.mem_alloced += size_t(SLAB_SIZE)
      if m.mem_alloced > m.mem_limit {
         panic("Out of memory")
      }
      tmp := C.malloc(C.size_t(SLAB_SIZE))
      m.free_slab_list.PushBack(tmp)
      m.mem_alloced += SLAB_SIZE
   }
   ret := m.free_slab_list.Front().Value.(unsafe.Pointer)
   return ret
}

/* new slab from system memory */
func (m *MemPool) do_slab_newslab(id int) {
   p := &m.slabclass[id]
   if p.slabs >= p.list_size {
      m.grow_slab_list(id)
   }
   ptr := m.memory_allocate_newslab()

   C.memset(ptr, 0, C.size_t(SLAB_SIZE))
   m.split_slab_page_into_freelist(ptr, id)

   p.slabs++
   tmp := (*[1<<16]unsafe.Pointer)(p.slab_list)
   tmp[p.slabs] = ptr
}

func (m *MemPool) do_slab_free(ptr unsafe.Pointer, size size_t, id int) {
  p := &m.slabclass[id]
  it := (*Item)(ptr)
  it.it_flags |= ITEM_SLABBED
  it.prev = nil
  it.next = p.slots;
  if it.next != nil {
     it.next.prev = it
  }
  p.slots = it

  p.sl_curr++
  p.requested -= size
}
