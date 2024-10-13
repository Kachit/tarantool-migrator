package tarantool_migrator

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type LoggerTestSuite struct {
	suite.Suite
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

func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}
