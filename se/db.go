// Copyright 2014

// 'se' short for storage engine.
package se

import (
   "sync"
)

type KVDB interface {
   Get(key []byte) []byte
   Set(key []byte, val []byte)
}

func NewKVDB() KVDB {
   kvdb := &kv_eng1{
              kvmap:make(map[string][]byte),
           }
   return kvdb
}

//a kvdb implemented by golang builtin map
type kv_eng1 struct {
   kvmap map[string][]byte
   lock sync.RWMutex
}


func (db *kv_eng1) Get(key []byte) []byte {
   db.lock.RLock()
   defer db.lock.RUnlock()
   return db.kvmap[string(key)]
}

func (db *kv_eng1) Set(key []byte, val []byte) {
   db.lock.Lock()
   defer db.lock.Unlock()
   db.kvmap[string(key)] = val
}

