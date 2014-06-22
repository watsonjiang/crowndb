package logging

import (
   "log"
)

type Logger interface {
   Fatal(v ...interface{})
   Fatalln(v ...interface{})
   Fatalf(format string, v ...interface{})
   Error(v ...interface{})
   Errorln(v ...interface{})
   Errorf(format string, v ...interface{})
   Warn(v ...interface{})
   Warnln(v ...interface{})
   Warnf(format string, v ...interface{}) 
   Info(v ...interface{})
   Infoln(v ...interface{})
   Infof(format string, v ...interface{})
   Debug(v ...interface{})
   Debugln(v ...interface{})
   Debugf(format string, v ...interface{})
   Trace(v ...interface{})
   Traceln(v ...interface{})
   Tracef(format string, v ...interface{})
}

type LoggerFacMethod func(id int) Logger

var logger_pool map[int] Logger
var logger_factory_func LoggerFacMethod

func init() {
   logger_pool = make(map[int] Logger)
   logger_factory_func = buildin_new_logger_func
}

func GetLogger(id int) Logger{
   logger, found := logger_pool[id]
   if found {
      return logger
   }else{
      logger = logger_factory_func(id)
      logger_pool[id] = logger
      return logger
   }
}

func RegisterLoggerFactory(f LoggerFacMethod){
   logger_factory_func = f
}

func buildin_new_logger_func(id int) Logger {
   return &std_logger{id : id}
}

type std_logger struct {
   id int
}

func (l *std_logger) Fatal(v ...interface{}) {
   log.Fatal(v...)
}

func (l *std_logger) Fatalln(v ...interface{}) {
   log.Fatalln(v...)
}

func (l *std_logger) Fatalf(format string, v ...interface{}) {
   log.Fatalf(format, v...)
}

func (l *std_logger) Error(v ...interface{}) {
   log.Print(v...)
}

func (l *std_logger) Errorln(v ...interface{}) {
   log.Println(v...)
}

func (l *std_logger) Errorf(format string, v ...interface{}) {
   log.Printf(format, v...)
}

func (l *std_logger) Warn(v ...interface{}) {
   log.Print(v...)
}

func (l *std_logger) Warnln(v ...interface{}) {
   log.Println(v...)
}

func (l *std_logger) Warnf(format string, v ...interface{}) {
   log.Printf(format, v...)
}

func (l *std_logger) Info(v ...interface{}) {
   log.Print(v...)
}

func (l *std_logger) Infoln(v ...interface{}) {
   log.Println(v...)
}

func (l *std_logger) Infof(format string, v ...interface{}) {
   log.Printf(format, v...)
}

func (l *std_logger) Debug(v ...interface{}) {
   log.Print(v...)
}

func (l *std_logger) Debugln(v ...interface{}) {
   log.Println(v...)
}

func (l *std_logger) Debugf(format string, v ...interface{}) {
   log.Printf(format, v...)
}

func (l *std_logger) Trace(v ...interface{}) {
   log.Print(v...)
}

func (l *std_logger) Traceln(v ...interface{}) {
   log.Println(v...)
}

func (l *std_logger) Tracef(format string, v ...interface{}) {
   log.Printf(format, v...)
}


