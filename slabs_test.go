package mempool

import (
   "fmt"
   "testing"
   "unsafe"
)

func Test_slabs_clsid(t *testing.T) {
   item := item_t{}
   fmt.Println("sizeof item", unsafe.Sizeof(item))
}

func Test_SlabInit(t *testing.T) {
   SlabInit(1024*1024, 1, true)
   fmt.Println("mem_current", mem_current, "mem_avail", mem_avail)
}

func Test_demo(t *testing.T) {
   fmt.Println("sizeof slabclass_t", unsafe.Sizeof(slabclass[0]))
   fmt.Println("sizeof slabclass", unsafe.Sizeof(slabclass))
}

