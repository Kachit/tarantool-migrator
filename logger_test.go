package tarantool_migrator

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type LoggerTestSuite struct {
	suite.Suite
	ctx       context.Context
	logWriter *testLogWriter
}

func (suite *LoggerTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.logWriter = &testLogWriter{Logs: make([]testLogMsg, 0)}
}

func (suite *LoggerTestSuite) TestGetPrefix() {
	testable := LoggerConfig{
		LogLevel: LogLevelSilent,
		Prefix:   LogPrefixDefault,
	}
	assert.Equal(suite.T(), LogPrefixDefault+": ", testable.getPrefix())
	testable.Prefix = ""
	assert.Equal(suite.T(), "", testable.getPrefix())
}

func (suite *LoggerTestSuite) TestLogLevelSilent() {
	testable := NewLogger(suite.logWriter, LoggerConfig{
		LogLevel: LogLevelSilent,
	})
	testable.Info(suite.ctx, "info", 123, 456)
	testable.Debug(suite.ctx, "debug", 123, 456)
	assert.Len(suite.T(), suite.logWriter.Logs, 0)
}

func (suite *LoggerTestSuite) TestLogLevelInfo() {
	testable := NewLogger(suite.logWriter, LoggerConfig{
		LogLevel: LogLevelInfo,
	})
	testable.Info(suite.ctx, "info", 123, 456)
	testable.Debug(suite.ctx, "debug", 123, 456)
	assert.Len(suite.T(), suite.logWriter.Logs, 1)
	assert.Equal(suite.T(), "info", suite.logWriter.Logs[0].msg)
}

func (suite *LoggerTestSuite) TestLogLevelDebug() {
	testable := NewLogger(suite.logWriter, LoggerConfig{
		LogLevel: LogLevelDebug,
	})
	testable.Info(suite.ctx, "info", 123, 456)
	testable.Debug(suite.ctx, "debug", 123, 456)
	assert.Len(suite.T(), suite.logWriter.Logs, 2)
	assert.Equal(suite.T(), "info", suite.logWriter.Logs[0].msg)
	assert.Equal(suite.T(), "debug [123,456]", suite.logWriter.Logs[1].msg)
}

func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

type testLogMsg struct {
	msg  string
	args []interface{}
}

type testLogWriter struct {
	Logs []testLogMsg
}

func (lw *testLogWriter) Printf(msg string, args ...interface{}) {
	lw.Logs = append(lw.Logs, testLogMsg{
		msg:  msg,
		args: args,
	})
}
