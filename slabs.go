package mempool

/* a golang copy of memchached slab allocator */

import (
      "fmt"
      "unsafe"
      "sync"
      "container/list"
      )
//#include <stdlib.h>
//#include <string.h>
import "C"

type size_t uint64

/* allocation structure */
type slabclass_t struct {
   size size_t          /*size of items*/
   perslab size_t       /*how many items per slab*/
   slots *item_t        /*list of item ptrs*/
   sl_curr uint         /*total free items in list*/
   slabs uint           /*the number of slabs for this class*/
   slab_list unsafe.Pointer /*array of slab pointers*/
   list_size uint       /*size of array above*/
   killing uint         /*index+1 of dying slab*/
   requested size_t     /*number of requested bytes*/
}

const (
   SLAB_SIZE = 1024 * 1024
   CHUNK_SIZE = 32
   CHUNK_ALIGN_BYTES = 8
   MAX_NUM_OF_SLAB_CLS = 256
)

var slabclass [MAX_NUM_OF_SLAB_CLS]slabclass_t
var power_smallest uint = 0      /*the smallest slabcalss id*/
var power_largest uint = 0       /*the largest slabclass id*/
var mem_limit size_t
var mem_alloced size_t
var free_slab_list list.List
var slabs_lock sync.Mutex

/* Figures out which slab class is required to store an item
   of given size
*/
func slabs_clsid(size size_t) uint {
   if size == 0 {
      return 0
   }
   res := uint(0)
   for ;size > slabclass[res].size; {
      res++
      if res ==  MAX_NUM_OF_SLAB_CLS {
         return 0
      }
   }
   return res
}

/* Determines the chunk sizes and initializes the slab calss */
func SlabInit(limit size_t, factor float32, prealloc bool) {
   item := item_t{}
   size := size_t(unsafe.Sizeof(item)) + CHUNK_SIZE
   mem_limit = limit

   if prealloc {
      /* prealloc 100 free slabs */
      num_slabs := limit / size_t(SLAB_SIZE)
      if limit % SLAB_SIZE != 0 {
         num_slabs++
      }
      for i:=size_t(0); i<num_slabs; i++ {
         tmp := C.malloc(C.size_t(SLAB_SIZE))
         free_slab_list.PushBack(tmp)
         mem_alloced += SLAB_SIZE
      }
   }

   C.memset(unsafe.Pointer(&slabclass[0]),
            0, C.size_t(unsafe.Sizeof(slabclass)))

   for i:=uint(0);i<MAX_NUM_OF_SLAB_CLS;i++ {
      /*make sure items are always n-byte aligned */
      if size % CHUNK_ALIGN_BYTES != 0 {
         size += CHUNK_ALIGN_BYTES - (size % CHUNK_ALIGN_BYTES)
      }
      if size > SLAB_SIZE {
         /*max chunk size is SLAB_SIZE*/
         slabclass[i].size = 0
         slabclass[i].perslab = 0   //use 0 as end flag
         power_largest = i-1
         break
      }
      slabclass[i].size = size
      slabclass[i].perslab = size_t(SLAB_SIZE) / slabclass[i].size
      size = size_t(float64(size) * float64(factor))
   }

   // may take 6x seconds to init, :(
   //if prealloc {
   //   slabs_preallocate()
   //}
}

func SlabsAlloc(size size_t, id uint) unsafe.Pointer {
   slabs_lock.Lock()
   defer slabs_lock.Unlock()

   return do_slabs_alloc(size, id)
}

func SlabsFree(ptr unsafe.Pointer, size size_t, id uint) {
   slabs_lock.Lock()
   defer slabs_lock.Unlock()

   do_slabs_free(ptr, size, id)
}

/* dump the slabclass table for debug purpose. */
func slabs_cls_info_dump(){
   for i:=uint(0);i<power_largest;i++ {
      fmt.Println("class", i, "size", slabclass[i].size)
   }
}

func do_slabs_alloc(size size_t, id uint) unsafe.Pointer {
   if (id < power_smallest && id > power_largest) {
      panic("Invalid class id")
   }
   var ret unsafe.Pointer
   p:=&slabclass[id]
   if p.sl_curr == 0 {
      do_slabs_newslab(id)
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

func slabs_preallocate() {
   /* pre-allocate a 1mb slab in every size class.*/
   for i:=uint(0);i<=power_largest;i++ {
      do_slabs_newslab(i)
   }
}

func grow_slab_list(id uint) {
   p := &slabclass[id]
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

func split_slab_page_into_freelist(ptr unsafe.Pointer, id uint) {
   p := &slabclass[id]
   tmp := ptr
   var x size_t = 0
   fmt.Println("p.id", id, "p.perslab", p.perslab)
   for x=0;x<p.perslab;x++ {
       do_slabs_free(tmp, 0, id)
       tmp = unsafe.Pointer(&C.GoBytes(tmp, SLAB_SIZE)[p.size]);
   }
}

func memory_allocate_newslab() unsafe.Pointer {
   fmt.Println("!!!!!!free_slab_list len", free_slab_list, free_slab_list.Len())
   if free_slab_list.Len() == 0 {
      /* no available slab in free slab list */
      mem_alloced += size_t(SLAB_SIZE)
      if mem_alloced > mem_limit {
         panic("Out of memory")
      }
      tmp := C.malloc(C.size_t(SLAB_SIZE))
      fmt.Println("!!!!!malloc mem", tmp)
      free_slab_list.PushBack(tmp)
      mem_alloced += SLAB_SIZE
   }
   ret := free_slab_list.Front().Value.(unsafe.Pointer)
   fmt.Println("newslab", ret) 
   return ret
}

/* new slab from system memory */
func do_slabs_newslab(id uint) {
   p := &slabclass[id]
   if p.slabs >= p.list_size {
      grow_slab_list(id)
   }
   ptr := memory_allocate_newslab()

   C.memset(ptr, 0, C.size_t(SLAB_SIZE))
   split_slab_page_into_freelist(ptr, id)

   p.slabs++
   tmp := (*[1<<16]unsafe.Pointer)(p.slab_list)
   tmp[p.slabs] = ptr
}

func do_slabs_free(ptr unsafe.Pointer, size size_t, id uint) {
  p := &slabclass[id]
  it := (*item_t)(ptr)
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
