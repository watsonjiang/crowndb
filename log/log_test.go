package logging

import (
   "testing"
)

func TestLog(t *testing.T) {
   l := GetLogger(1)
   l.Debugln("Hello, logging is working correctly!")
}

func TestLogLevel(t *testing.T) {
   l := GetLogger(2)
   l.SetLevel(L_INFO)
   if l.IsDebugEnabled() {
      l.Debugln("ERROR! you should not see these text.")
   }
}
