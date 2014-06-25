package mempool

import (
   "unsafe"
)

/* allocator structure */
type allocator_t struct {
   size size_t              //size of items in this allocator
   used_item_list *item_t     //used item list
   free_item_list *item_t     //free item list
   slab_list *slab_t        //used slab list 
}

// split an empty slab into piceses, link them into free
// item list
func (a *allocator_t) add_slab(s *slab_t) {
   perslab := int(SLAB_SIZE / a.size)
   var tmp *[SLAB_SIZE]byte = (*[SLAB_SIZE]byte)(s.ptr)
   for x:=0;x<perslab;x++ {
      var item *item_t = (*item_t)(unsafe.Pointer(&tmp[int(a.size) * x]))
      item.next = a.free_item_list
      if a.free_item_list != nil {
         a.free_item_list.prev = item
      }
      a.free_item_list = item
      item.prev = nil
   }
   s.next = a.slab_list
   a.slab_list = s
}

func (a *allocator_t)alloc_item() *item_t{
   var it *item_t
   //get one from free list
   if a.free_item_list == nil {
      return nil    //return nil when full
   }else{
      it = a.free_item_list
      a.free_item_list=it.next
   }
   //put it into used list
   it.next = a.used_item_list
   if a.used_item_list != nil {
      a.used_item_list.prev = it
   }
   it.prev = nil
   a.used_item_list = it
   /* initialize the item */
   it.refcount = 1
   return it
}

func (a *allocator_t)free_item(it *item_t) {
  //unlink the item from used list
  if it.prev == nil { //head of used list
     a.used_item_list = it.next
     if a.used_item_list != nil {
        a.used_item_list.prev = nil
     }
  }else {  //in the middle of used list
     it.prev.next = it.next
     if it.next != nil {
        it.next.prev = it.prev
     }
  }
  //link it into free list
  it.next = a.free_item_list
  a.free_item_list.prev = it.next
  a.free_item_list = it
  a.free_item_list.prev = nil
}
