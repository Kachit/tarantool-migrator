package tarantool_migrator

import "github.com/tarantool/go-tarantool/v2/pool"

// Options define options for all migrations.
type Options struct {
	// MigrationsSpace is the migrations space.
	MigrationsSpace string
	Instances       []string

	TransactionEnabled bool
	LogLevel           LogLevel
	ReadMode           pool.Mode
	WriteMode          pool.Mode
}

var DefaultOptions = &Options{
	MigrationsSpace: "migrations",
	LogLevel:        LogLevelInfo,
	ReadMode:        pool.ANY,
	WriteMode:       pool.RW,
}
