package tarantool_migrator

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPresetLoggers(t *testing.T) {
	assert.NotNil(t, DefaultLogger)
	assert.NotNil(t, DebugLogger)
	assert.NotNil(t, SilentLogger)
}

func TestPresetLoggersLevels(t *testing.T) {
	assert.True(t, DefaultLogger.Enabled(nil, slog.LevelInfo))
	assert.False(t, DefaultLogger.Enabled(nil, slog.LevelDebug))

	assert.True(t, DebugLogger.Enabled(nil, slog.LevelDebug))

	assert.False(t, SilentLogger.Enabled(nil, slog.LevelInfo))
	assert.False(t, SilentLogger.Enabled(nil, slog.LevelDebug))
}
