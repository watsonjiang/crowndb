package se

import (
   "testing"
)

func TestSetGet(t *testing.T) {
   db := NewKVDB()
   db.Set([]byte("key1"), []byte("hello"))
   t.Error(string(db.Get([]byte("key1"))))
}

