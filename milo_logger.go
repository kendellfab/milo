package milo

import (
	"log"
	"os"
	"runtime"
)

// Methods to specify a logger for the milo app.
type MiloLogger interface {
	Log(message string)
	LogError(err error)
	LogInterfaces(items ...interface{})
	LogFatal(items ...interface{})
	LogStackTrace()
}

// An internal constructor for a default logger.
func newDefaultLogger() *defaultLogger {
	l := &defaultLogger{info: log.New(os.Stdout, "MILO:", log.Ldate|log.Ltime)}
	return l
}

// Default logger to only be registered inside of the milo package.
type defaultLogger struct {
	info *log.Logger
}

// Log a simple message.
func (l *defaultLogger) Log(message string) {
	l.info.Println(message)
}

// Log a simple error.
func (l *defaultLogger) LogError(err error) {
	l.info.Println(err)
	l.LogStackTrace()
}

// Have more data, send it to an interface logger and it'll be output for you.
func (l *defaultLogger) LogInterfaces(items ...interface{}) {
	l.info.Println(items...)
}

// Want to die on an error, send it here.
func (l *defaultLogger) LogFatal(items ...interface{}) {
	l.info.Fatalln(items...)
}

// Write out a stack trace for the current pc.
func (l *defaultLogger) LogStackTrace() {
	buf := make([]byte, 2048)
	runtime.Stack(buf, false)
	l.info.Println(string(buf))
}
