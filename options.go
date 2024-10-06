package tarantool_migrator

import "github.com/tarantool/go-tarantool/v2/pool"

// Options define options for all migrations.
type Options struct {
	// SpaceName is the migrations space.
	SpaceName string
	Instances []string

	UseTransaction bool
	ReadMode       pool.Mode
	WriteMode      pool.Mode
}

var DefaultOptions = &Options{
	SpaceName: "migrations",
	ReadMode:  pool.ANY,
	WriteMode: pool.RW,
}
