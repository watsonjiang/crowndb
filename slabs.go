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
   perslab uint      /*how many items per slab*/
   slots uintptr     /*list of item ptrs*/
   sl_curr uint      /*total free items in list*/
   slabs uint        /*the number of slabs for this class*/
   slab_list uintptr /*array of slab pointers*/
   list_size uint    /*size of prev array*/
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
   ITEM_SIZE_MAX = 1024
   MAX_NUMBER_OF_SLAB_CLASSES = 256
)



var slabclass [MAX_NUMBER_OF_SLAB_CLASSES]slabclass_t
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
   i := 0
   item := item_t{} 
   size := unsafe.Sizeof(item)
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

   i++
   while( i < MAX_NUMBER_OF_SLAB_CLASSES && size <= ITEM_SIZE_MAX/factor) {
      /*make sure items are always n-byte aligned */
      if size % CHUNK_ALIGN_BYTES {
         size += CHUNK_ALIGN_BYTES - (size % CHUNK_ALIGN_BYTES)

      slabclass[i].size = size
      slabclass[i].perslab = ITEM_SIZE_MAX / slabclass[i].size

   

}
