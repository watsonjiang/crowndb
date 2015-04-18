package replica

import (
   "time"
   "bytes"
)

type oplog_entry struct {
   id uint64
   ts time.Time
   data bytes.Buffer
}

func init_oplog() {

}

func write_oplog(e *oplog_entry){

}

func new_oplog_entry() *oplog_entry {

}

func release_oplog_entry(e *oplog_entry) {

}
