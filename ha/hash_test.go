package ha
import (
   "testing"
)

func TestHash(t *testing.T) {
   s := "hello"
   e := uint32(2066056305)
   a := murmur3_32(s)
   if e!=a{
      t.Error("Exp:", e, "Act:", a)
   }
}
