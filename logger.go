package tarantool_migrator

import (
	"io"
	"log/slog"
	"os"
	"time"
)

var DefaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
var DebugLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
var SilentLogger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError + 1}))

func formatDurationToMs(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1e6
}
