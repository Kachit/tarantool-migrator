package tarantool_migrator

import "errors"

// ErrMissingID is returned when the ID of migration is equal to ""
var ErrMissingID = errors.New("missing ID in migration")
var ErrMissingMigrateFunc = errors.New("missing migrate function in migration")
var ErrMissingRollbackFunc = errors.New("missing rollback function in migration")

// ErrNoDefinedMigrations is returned when no migrations are defined.
var ErrNoDefinedMigrations = errors.New("no defined migrations")
var ErrNoAppliedMigrations = errors.New("no applied migrations")
var ErrMigrationIDDoesNotExist = errors.New("tried to migrate to an ID that doesn't exist")
var ErrWrongMigrationFileFormat = errors.New("wrong migration file format")
var ErrWrongMigrationCmdFormat = errors.New("wrong migration cmd format")
