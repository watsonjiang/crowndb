package hashtable

import "github.com/watsonjiang/crowndb/mempool"

type HashTable interface {
   Put(key string, val string)
   Get(key string) string
}

type hashfunc_t func (key string) uint32 

type htelem_t struct {
   next *htelem_t
   val Item
}

type htable_t struct {
   hfunc hashfunc_t
   key_count int
   array_size int
   array []*htelem_t
   
}

func New() {
   ret := &htable_t{}
   ret.hashfunc_t = murmur3_32
}

func (ht *htable_t) Put(key string, val string) {
 
}

func (ht *htable_t) Get(key string) {

}

