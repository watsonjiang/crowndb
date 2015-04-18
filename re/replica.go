package replica

import (
   "time"
   "sync/atomic"
   "github.com/watsonjiang/crowndb/db"
)

type Replicator struct {
   db *db.DB
   uid uint64
}

func NewReplicaDb(db *db.DB) *Replicator {
   rep := &Replicator{
             db : db,
          }
   return rep
}

func (r *Replicator) Get(key string) string {
   return db.Get(key)
}

func (r *Replicator) get_next_uid() {
   return atomic.AddUint64(&r.uid, 1)
}

func (r *Replicator) Put(key string, val string) bool {
   //TODO: gen oplog and put into replication queue
   if db.Put(key, val) {
      oplog = new_oplog_entry()
      oplog.id = r.get_next_uid()
      oplog.ts = time.Now()
      oplog.data.
   }
   return false
}
