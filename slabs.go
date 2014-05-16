package mempool

/* a golang copy of memchached slab allocator */

import (
      "fmt"
      "unsafe"
      "os"
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
   CHUNK_ALIGN_BYTES = 8
   ITEM_SIZE_MAX = 1024 * 1024
   MAX_NUMBER_OF_SLAB_CLASSES = 256
)

var slabclass [MAX_NUMBER_OF_SLAB_CLASSES]slabclass_t
var power_smallest uint = 0      /*the smallest slabcalss id*/
var power_largest uint = 0       /*the largest slabclass id*/
var mem_limit size_t
var mem_malloced size_t

/* used in prealloc mode */
var mem_base unsafe.Pointer
var mem_current unsafe.Pointer
var mem_avail size_t

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
      if res ==  MAX_NUMBER_OF_SLAB_CLASSES {
         return 0
      }
   }
   return res
}

/* Determines the chunk sizes and initializes the slab calss */
func SlabInit(limit size_t, factor float32, prealloc bool) {
   item := item_t{} 
   size := size_t(unsafe.Sizeof(item))
   mem_limit = limit

   if prealloc {
      mem_base = C.malloc(C.size_t(mem_limit))
      if mem_base != nil {
         mem_current = mem_base
         mem_avail = mem_limit
      } else {
         fmt.Println("Warning: Failed to allocate required memory in",
                     "one large chunk.")
         os.Exit(1)
      }
   }

   C.memset(unsafe.Pointer(&slabclass[0]), 
            0, C.size_t(unsafe.Sizeof(slabclass)))
   var i uint
   for i=1;i<MAX_NUMBER_OF_SLAB_CLASSES;i++ {
      /*make sure items are always n-byte aligned */
      if size % CHUNK_ALIGN_BYTES != 0 {
         size += CHUNK_ALIGN_BYTES - (size % CHUNK_ALIGN_BYTES)
      }
      slabclass[i].size = size
      slabclass[i].perslab = size_t(ITEM_SIZE_MAX) / slabclass[i].size
      size = size_t(float64(size) * float64(factor))
      if size > ITEM_SIZE_MAX {
         break
      }
      fmt.Println("slab class", i, "chunk size", slabclass[i].size,
                  "perslab", slabclass[i].perslab)
   }
   power_largest = i
   slabclass[i].size = ITEM_SIZE_MAX
   slabclass[i].perslab = 1
   fmt.Println("slab class", i, "chunk size", slabclass[i].size,
               "perslab", slabclass[i].perslab)

   if prealloc {
      slabs_preallocate(MAX_NUMBER_OF_SLAB_CLASSES)
   }
}

func slabs_preallocate(maxslabs uint) {
   /* pre-allocate a 1mb slab in every size class.*/
   var prealloc uint = 1
   var i uint
   for i=1;i<=MAX_NUMBER_OF_SLAB_CLASSES;i++ {
      if prealloc > maxslabs {
         return
      }
      if do_slabs_newslab(i) == 0 {
         fmt.Println("Error while preallocating slab memory")
         os.Exit(1)
      }
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
   for x=0;x<p.perslab;x++ {
       do_slabs_free(ptr, 0, id)
       tmp = unsafe.Pointer(&C.GoBytes(tmp, 1024*1024)[p.size]);
   }
}

func memory_allocate(size size_t) unsafe.Pointer {
   var ret unsafe.Pointer = nil
   if mem_base == nil {
      /* not in prealloc mode, using system malloc. */
      ret = C.malloc(C.size_t(size))
   } else {
      ret = mem_current
      if size > mem_avail {
         panic("Not enough memory")
         return nil
      }
      /* mem_current pointer must be aligned!!! */
      if size % CHUNK_ALIGN_BYTES != 0 {
         size += CHUNK_ALIGN_BYTES - (size % CHUNK_ALIGN_BYTES)
      }

      mem_current = unsafe.Pointer(&C.GoBytes(mem_current, 255)[size])
      if size < mem_avail {
         mem_avail -= size
      }else{
         mem_avail = 0
      }
   }
   return ret
}

func do_slabs_newslab(id uint) int {
   p := &slabclass[id]
   len := p.size * p.perslab

   ptr:=memory_allocate(len)

   C.memset(ptr, 0, C.size_t(len))
   split_slab_page_into_freelist(ptr, id)

   p.slabs++
   k := p.list_size
   tmp := ([k]*unsafe.Pointer)(p.slab_list)
   tmp[p.slabs] = ptr
   return 0
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
