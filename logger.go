package tarantool_migrator

type LogLevel int

const LogLevelSilent LogLevel = iota
const LogLevelInfo LogLevel = iota
const LogLevelDebug LogLevel = iota

type Logger interface {
	Log(msg string)
}
