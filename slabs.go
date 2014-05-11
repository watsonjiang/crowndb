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
   size size_t         /*size of items*/
   perslab uint64      /*how many items per slab*/
   slots uintptr     /*list of item ptrs*/
   sl_curr uint      /*total free items in list*/
   slabs uint        /*the number of slabs for this class*/
   slab_list uintptr /*array of slab pointers*/
   list_size uint    /*size of array above*/
   killing uint      /*index+1 of dying slab*/
   requested uint    /*number of requested bytes*/
}

/* structure for storing items in the pool*/
type item_t struct {
   next *item_t
   prev *item_t
   h_next *item_t    /*hash chain next*/
   time uint         /*least recent access*/
   exptime uint      /*expire time*/
   nbytes int        /*size of data*/
   refcount int
   nsuffix uint8     /*lenth of flags-and-length string*/
   it_flags uint8
   slabs_clsid uint8 /*which slab class the item belongs to*/
   nkey uint8        /*key length*/
}

const (
   CHUNK_ALIGN_BYTES = 8
   ITEM_SIZE_MAX = 1024 * 1024
   MAX_NUMBER_OF_SLAB_CLASSES = 256
)



var slabclass [MAX_NUMBER_OF_SLAB_CLASSES]slabclass_t
var power_smallest uint = 0
var power_largest uint = 0
var mem_limit size_t
var mem_malloced size_t
var mem_base uintptr
var mem_current uintptr
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
      mem_base = uintptr(C.malloc(C.size_t(mem_limit)))
      if mem_base != 0 {
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
      slabclass[i].perslab = uint64(ITEM_SIZE_MAX) / uint64(slabclass[i].size)
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
      if DoSlabsNewSlab(i) == 0 {
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
     new_list := uintptr(C.realloc(unsafe.Pointer(p.slab_list), C.size_t(new_size*uint(unsafe.Sizeof(p.slab_list)))))
     p.list_size = new_size
     p.slab_list = new_list
  }
}

func DoSlabsNewSlab(id uint) int {
   //p := &slabclass[id]
   //len := 
  return 0
}
