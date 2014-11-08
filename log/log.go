// Copyright, 2014 watson (watson.jiang@gmail.com)

// Package 'logging' here is not another implementation of log or glog, 
// the design purpose of this package is to provide a standard interface
// for other modules. by using this package, it would be easy to switch 
// between different log libraries without changing too much of the code.
package logging

import (
   "fmt"
)

const (
   L_FATAL = iota
   L_ERROR
   L_WARN
   L_INFO
   L_DEBUG
   L_TRACE
)

// The standard interface for logging.
// note: to improve the performance, IsXXXEnabled is added into the interface.
// it recommended to detect the current log level before logging anything.
type Logger interface {
   SetLevel(l int)
   IsFatalEnabled() bool
   Fatal(v ...interface{})
   Fatalln(v ...interface{})
   Fatalf(format string, v ...interface{})
   IsErrorEnabled() bool
   Error(v ...interface{})
   Errorln(v ...interface{})
   Errorf(format string, v ...interface{})
   IsWarnEnabled() bool
   Warn(v ...interface{})
   Warnln(v ...interface{})
   Warnf(format string, v ...interface{})
   IsInfoEnabled() bool
   Info(v ...interface{})
   Infoln(v ...interface{})
   Infof(format string, v ...interface{})
   IsDebugEnabled() bool
   Debug(v ...interface{})
   Debugln(v ...interface{})
   Debugf(format string, v ...interface{})
   IsTraceEnabled() bool
   Trace(v ...interface{})
   Traceln(v ...interface{})
   Tracef(format string, v ...interface{})
}

type LoggerFacMethod func() Logger

var logger_pool map[int] Logger
var logger_factory_func LoggerFacMethod

func init() {
   logger_pool = make(map[int] Logger)
   logger_factory_func = builtin_new_logger_func
}

// Get a logger for logging. you can use the id to classify log streams.
// by default, logging uses golang log to output log stream to os.Sysout
func GetLogger(id int) Logger{
   logger, found := logger_pool[id]
   if found {
      return logger
   }else{
      logger = logger_factory_func()
      logger_pool[id] = logger
      return logger
   }
}

// Register new logger factory.
func RegisterLoggerFactory(f LoggerFacMethod){
   logger_factory_func = f
}

func builtin_new_logger_func() Logger {
   return &default_logger{level : L_INFO}
}

// Default log implementation. 
// fmt.print everything to screen directly
type default_logger struct {
   level int
}

func (l *default_logger) SetLevel(level int) {
   l.level = level
}

func (l *default_logger) IsFatalEnabled() bool {
   if l.level >= L_FATAL {
      return true
   }
   return false
}

func (l *default_logger) Fatal(v ...interface{}) {
   fmt.Print(v...)
}

func (l *default_logger) Fatalln(v ...interface{}) {
   fmt.Println(v...)
}

func (l *default_logger) Fatalf(format string, v ...interface{}) {
   fmt.Printf(format, v...)
}

func (l *default_logger) IsErrorEnabled() bool {
   if l.level >= L_ERROR {
      return true
   }
   return false
}

func (l *default_logger) Error(v ...interface{}) {
   fmt.Print(v...)
}

func (l *default_logger) Errorln(v ...interface{}) {
   fmt.Println(v...)
}

func (l *default_logger) Errorf(format string, v ...interface{}) {
   fmt.Printf(format, v...)
}

func (l *default_logger) IsWarnEnabled() bool {
   if l.level >= L_WARN {
      return true
   }
   return false
}

func (l *default_logger) Warn(v ...interface{}) {
   fmt.Print(v...)
}

func (l *default_logger) Warnln(v ...interface{}) {
   fmt.Println(v...)
}

func (l *default_logger) Warnf(format string, v ...interface{}) {
   fmt.Printf(format, v...)
}

func (l *default_logger) IsInfoEnabled() bool {
   if l.level >= L_INFO {
      return true
   }
   return false
}

func (l *default_logger) Info(v ...interface{}) {
   fmt.Print(v...)
}

func (l *default_logger) Infoln(v ...interface{}) {
   fmt.Println(v...)
}

func (l *default_logger) Infof(format string, v ...interface{}) {
   fmt.Printf(format, v...)
}

func (l *default_logger) IsDebugEnabled() bool {
   if l.level >= L_DEBUG {
      return true
   }
   return false
}

func (l *default_logger) Debug(v ...interface{}) {
   fmt.Print(v...)
}

func (l *default_logger) Debugln(v ...interface{}) {
   fmt.Println(v...)
}

func (l *default_logger) Debugf(format string, v ...interface{}) {
   fmt.Printf(format, v...)
}

func (l *default_logger) IsTraceEnabled() bool {
   if l.level >= L_TRACE {
      return true
   }
   return false
}

func (l *default_logger) Trace(v ...interface{}) {
   fmt.Print(v...)
}

func (l *default_logger) Traceln(v ...interface{}) {
   fmt.Println(v...)
}

func (l *default_logger) Tracef(format string, v ...interface{}) {
   fmt.Printf(format, v...)
}


