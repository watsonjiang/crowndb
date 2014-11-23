package mempool

import (
      "fmt"
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
   //h_next *Item      /*hash chain next*/
   time uint         /*least recent access*/
   exptime uint      /*expire time*/
   it_flags uint8    /*item flags*/
   nkey size_t       /*key length*/
   nval size_t     /*lenth of value*/
   data byte         /*start of data*/
}

const (
   ITEM_HEAD_SIZE = size_t(unsafe.Sizeof(item_t{}))
   ITEM_PAYLOAD_MIN = 32           //init item payload size
   ITEM_PAYLOAD_MAX = (1024*1024)-ITEM_HEAD_SIZE
)
/* return the size of item needed to store the kv pair*/
func item_size(nkey size_t, nvalue size_t) size_t {
   size := ITEM_HEAD_SIZE + nkey + nvalue
   return size
}

func (it *item_t) SetKV(k []byte, v []byte) {
   it.nkey = size_t(len(k))
   it.nval = size_t(len(v))
   if it.nkey + it.nval > ITEM_PAYLOAD_MAX {
      panic(
         fmt.Sprintf("item overflow. max payload size:%d, actual:%d", 
                     ITEM_PAYLOAD_MAX, it.nkey + it.nval))
   }
   pdata := (*[ITEM_PAYLOAD_MAX]byte)(unsafe.Pointer(&it.data))
   copy(pdata[0:], k[0:])
   copy(pdata[it.nkey:], v[0:])
}

func (it *item_t) GetKV() ([]byte , []byte) {
   pdata := (*[ITEM_PAYLOAD_MAX]byte)(unsafe.Pointer(&it.data))
   k := pdata[0:it.nkey]
   v := pdata[it.nkey:it.nval]
   return k, v
}

