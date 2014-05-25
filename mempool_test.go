package mempool

import (
   //"fmt"
   "testing"
   //"unsafe"
)

func Test_item_alloc_free(t *testing.T) {
   init_slabs()
   key := "Hello"
   value := "world"
   it := ItemAlloc(key, uint(len(value)))
   ItemFree(it)
}


