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
type Item struct {
   next *Item
   prev *Item
   //h_next *Item      /*hash chain next*/
   time uint         /*least recent access*/
   exptime uint      /*expire time*/
   refcount int
   it_flags uint8    /*item flags*/
   nkey size_t       /*key length*/
   nval size_t     /*lenth of value*/
   data byte         /*start of data*/
}

const (
   ITEM_HEADER_SIZE = size_t(unsafe.Sizeof(Item{}))
)
/* return the size of item needed to store the kv pair*/
func item_size(nkey size_t, nvalue size_t) size_t {
   size := ITEM_HEADER_SIZE + nkey + nvalue
   return size
}


