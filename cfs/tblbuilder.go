package configserver

import (
   log "github.com/golang/glog"
   "sort"
   "errors"
   rc "yy.com/ramcloud"
)

type svr_load struct {
   svr   uint64         //server id
   load  int            //number of replica it holds
}

type svr_load_ary []svr_load

func (s svr_load_ary) Len() int {
   return len(s)
}

func (s svr_load_ary) Less(i, j int) bool {
   return s[i].load < s[j].load
}

func (s svr_load_ary) Swap(i, j int) {
   t := s[i]
   s[i] = s[j]
   s[j] = t
}

type TblBuilder struct {
   live_svrs  []uint64
   repl_num   int
   svr_repl_load   svr_load_ary
   svr_prim_repl_load svr_load_ary
}

func NewTblBuilder(svrs []uint64, repl_num int) *TblBuilder {
   b := &TblBuilder{
      repl_num : repl_num,
      }
   b.live_svrs = append(b.live_svrs, svrs...)
   for _,s :=range b.live_svrs {
      b.svr_repl_load = append(b.svr_repl_load, svr_load{s, 0})
      b.svr_prim_repl_load = append(b.svr_prim_repl_load, svr_load{s, 0})
   }
   return b
}

//return a bucket placement plan
//plan meets:
// 1 no two replicas distributed on the same servers
// 2 all servers hold almost the same number of 'primary' bucket replica
// 3 all servers hold almost the same number of replicas
// reutrn nil if not possible to make a plan
func (b *TblBuilder) alloc_bucket_placement() []uint64 {
   var rst []uint64
   //alloc primary repl
   sort.Sort(b.svr_prim_repl_load)
   b.svr_prim_repl_load[0].load += 1
   s := b.svr_prim_repl_load[0]
   rst = append(rst, s.svr)
   //update svr repl load accordingly
   for _,i:=range b.svr_repl_load {
      if i.svr==s.svr {
         i.load += 1
      }
   }
   //alloc non-primary repl
   var tmp svr_load_ary
   for _,i:=range b.svr_repl_load {
      if i.svr==s.svr {
         continue          //eliminate already used server
      }
      tmp = append(tmp, i)
   }
   for i:=1;i<b.repl_num;i++ {
      if len(tmp)==0 {
         return nil        //no enough server
      }
      sort.Sort(tmp)
      s := tmp[0]      //the smallest load server
      rst = append(rst, s.svr)
      var tmp1 svr_load_ary
      for _,i:=range tmp {
         if i.svr==s.svr {
            continue      //eleminate already used server
         }
         tmp1 = append(tmp1, i)
      }
      tmp = tmp1
   }
   return rst
}

//build a new bucket tbl.
//return bucket tbl. 
func BuildBucketLocTbl(svrs []uint64, buck_num int, repl_num int) ([]uint64, error) {
   b := NewTblBuilder(svrs, repl_num)
   var rst []uint64
   for bucket := 0; bucket < buck_num; bucket++ {
      placement := b.alloc_bucket_placement()
      if placement==nil {
         return nil, errors.New("Fail to build tbl. server not enough")
      }
      if log.V(1) {
         log.V(1).Infoln("place bucket", bucket, "on server", rc.DumpTbl(placement))
      }
      rst = append(rst, placement...)
   }
   return rst, nil
}
