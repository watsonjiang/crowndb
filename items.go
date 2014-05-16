package mempool

import (
//      "fmt"
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
   nbytes int        /*size of data*/
   refcount int
   nsuffix uint8     /*lenth of flags-and-length string*/
   it_flags uint8
   slabs_clsid uint8 /*which slab class the item belongs to*/
   nkey uint8        /*key length*/
   data byte         /*start of data*/
}


