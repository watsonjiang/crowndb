package proto
/*
  the message used for internal communication
*/
type Message struct {
   Opcode uint8
   Args [][]byte
}
