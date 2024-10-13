package tarantool_migrator

import "github.com/tarantool/go-tarantool/v2/pool"

const createMigrationsSpacePath = "lua/migrations/create_migrations_space.up.lua"

// Options define options for all migrations.
type Options struct {
	// Migrations space
	MigrationsSpace string
	// Tarantool instances list
	Instances []string
	// Transactions enabled flag
	TransactionsEnabled bool
	// Default mode for read requests
	ReadMode pool.Mode
	// Default mode for read-write requests
	WriteMode pool.Mode
}

var DefaultOptions = &Options{
	MigrationsSpace: "migrations",
	ReadMode:        pool.ANY,
	WriteMode:       pool.RW,
}

func WithLogger(lg Logger) func(migrator *Migrator) {
	return func(m *Migrator) {
		m.logger = lg
	}
}

func WithOptions(op *Options) func(migrator *Migrator) {
	return func(m *Migrator) {
		m.opts = op
	}
}
