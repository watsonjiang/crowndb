package mempool

import (
      //"fmt"
      "unsafe"
      )

const (
   ITEM_LINKED  = 1
   ITEM_CAS     = 2
   ITEM_SLABBED = 4
   ITEM_FETCHED = 8
   )

/* structure for storing items in the pool*/
type item_t struct {
   next *item_t
   prev *item_t
   h_next *item_t    /*hash chain next*/
   time uint         /*least recent access*/
   exptime uint      /*expire time*/
   refcount int
   it_flags uint8    /*item flags*/
   slabs_clsid uint  /*which slab class the item belongs to*/
   nkey uint16       /*key length*/
   nvalue uint16     /*lenth of value*/
   data byte         /*start of data*/
}

/* return the size of item needed to store the kv pair*/
func item_size(nkey uint16, nvalue uint16) size_t {
   size := size_t(unsafe.Sizeof(item_t{})) + size_t(nkey) + size_t(nvalue)
   return size
}

func ItemAlloc(key string, nvalue uint16) *item_t {
   size := item_size(uint16(len(key)), nvalue)
   clsid := slabs_clsid(size)
   it := (*item_t)(SlabsAlloc(size, clsid))
   /* initialize the item */
   it.refcount = 1
   it.next = nil
   it.prev = nil
   it.h_next = nil
   it.slabs_clsid = clsid
   it.nkey = uint16(len(key))
   return it
}

func ItemFree(it *item_t) {

}
