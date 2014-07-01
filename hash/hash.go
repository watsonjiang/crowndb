package hash

import (
    "unsafe"
)

func murmur3_32(key string) uint32{
   var c1 uint32 = 0xcc9e2d51;
   var c2 uint32 = 0x1b873593;
   var r1 uint32 = 15;
   var r2 uint32 = 13;
   var m uint32 = 5;
   var n uint32 = 0xe6546b64;

   var hash uint32  = uint32(0 ^ len(key));
   var buf []byte = []byte(key)
   nblocks := len(key) / 4;
   for i:=0;i<nblocks;i++ {
      k := *((*uint32)(unsafe.Pointer(&buf[i*4])))
      k *= c1;
      k = (k << r1) | (k >> (32 - r1));
      k *= c2;

      hash ^= k;
      hash = ((hash << r2) | (hash >> (32 - r2))) * m + n;
   }

   tail := key[nblocks*4:];
   var k1 uint32 = 0;

   switch (len(key) & 3) {
   case 3:
      k1 ^= uint32(tail[2]) << 16;
   case 2:
      k1 ^= uint32(tail[1]) << 8;
   case 1:
      k1 ^= uint32(tail[0]);

      k1 *= c1;
      k1 = (k1 << r1) | (k1 >> (32 - r1));
      k1 *= c2;
      hash ^= k1;
   }
   hash ^= uint32(len(key));
   hash ^= (hash >> 16);
   hash *= 0x85ebca6b;
   hash ^= (hash >> 13);
   hash *= 0xc2b2ae35;
   hash ^= (hash >> 16);

   return hash;
}
