package tarantool_migrator

import (
	"context"
	"log"
	"os"
)

type LogLevel int

const LogLevelSilent LogLevel = iota
const LogLevelInfo LogLevel = iota
const LogLevelDebug LogLevel = iota

type LoggerConfig struct {
	LogLevel LogLevel
	Prefix   string
}

// LogWriter log writer interface
type LogWriter interface {
	Printf(string, ...interface{})
}

type Logger interface {
	Info(ctx context.Context, msg string, args ...interface{})
	Debug(ctx context.Context, msg string, args ...interface{})
	SetLogLevel(level LogLevel) Logger
}

var DefaultLogger = NewLogger(log.New(os.Stdout, "", log.LstdFlags), LoggerConfig{
	LogLevel: LogLevelInfo,
	Prefix:   "Tarantool-Migrator:",
})

type logger struct {
	LogWriter
	LoggerConfig
}

func (l *logger) SetLogLevel(level LogLevel) Logger {
	l.LogLevel = level
	return l
}

func (l *logger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= LogLevelInfo {
		l.Printf(l.Prefix+" "+msg, args...)
	}
}

func (l *logger) Debug(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= LogLevelDebug {
		l.Printf(l.Prefix+" "+msg, args...)
	}
}

func NewLogger(writer LogWriter, config LoggerConfig) Logger {
	return &logger{LogWriter: writer, LoggerConfig: config}
}
