package tarantool_migrator

import "time"

func FormatDurationToMs(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / 1e6
}
