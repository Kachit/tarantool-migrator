package tarantool_migrator

import (
	"context"
	"log"
	"os"
	"time"
)

type LogLevel int

const LogLevelSilent LogLevel = iota
const LogLevelInfo LogLevel = iota
const LogLevelDebug LogLevel = iota

const LogPrefixDefault string = "Tarantool-Migrator"

type LoggerConfig struct {
	LogLevel LogLevel
	Prefix   string
}

func (c *LoggerConfig) getPrefix() string {
	prefix := c.Prefix
	if prefix != "" {
		prefix = prefix + ": "
	}
	return prefix
}

// LogWriter log writer interface
type LogWriter interface {
	Printf(string, ...interface{})
}

type Logger interface {
	Info(ctx context.Context, msg string, args ...interface{})
	Debug(ctx context.Context, msg string, args ...interface{})
}

var DefaultLogger = NewLogger(log.New(os.Stdout, "", log.LstdFlags), LoggerConfig{
	LogLevel: LogLevelInfo,
	Prefix:   LogPrefixDefault,
})

var SilentLogger = NewLogger(log.New(os.Stdout, "", log.LstdFlags), LoggerConfig{
	LogLevel: LogLevelSilent,
	Prefix:   LogPrefixDefault,
})

type logger struct {
	LogWriter
	LoggerConfig
}

func (l *logger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= LogLevelInfo {
		l.Printf(l.getPrefix()+msg, args...)
	}
}

func (l *logger) Debug(ctx context.Context, msg string, args ...interface{}) {
	if l.LogLevel >= LogLevelDebug {
		l.Printf(l.getPrefix()+msg, args...)
	}
}

func NewLogger(writer LogWriter, config LoggerConfig) Logger {
	return &logger{LogWriter: writer, LoggerConfig: config}
}

func formatDurationToMs(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1e6
}
