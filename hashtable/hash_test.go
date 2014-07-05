package hash

import (
   "testing"
   "fmt"
)

func TestHash(t *testing.T) {
   s := "hello"
   fmt.Println("hash ", murmur3_32(s)) 
}
