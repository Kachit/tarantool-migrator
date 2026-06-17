package tarantool_migrator

import "github.com/tarantool/go-tarantool/v3/pool"

const createMigrationsSpacePath = "lua/migrations/create_migrations_space.up.lua"

// Options define options for all migrations.
type Options struct {
	// Migrations space
	MigrationsSpace string `json:"migrations_space"`
	// Dry run enabled flag
	DryRun bool `json:"dry_run"`
	// Transactions enabled flag
	TransactionsEnabled bool `json:"transactions_enabled"`
	// Default mode for read requests
	ReadMode pool.Mode `json:"read_mode"`
	// Default mode for read-write requests
	WriteMode pool.Mode `json:"write_mode"`
}

var DefaultOptions = Options{
	TransactionsEnabled: true,
	MigrationsSpace:     "migrations",
	ReadMode:            pool.ModeAny,
	WriteMode:           pool.ModeRW,
}
