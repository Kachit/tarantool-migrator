package tarantool_migrator

import "errors"

// ErrMissingID is returned when the ID od migration is equal to ""
var ErrMissingID = errors.New("missing ID in migration")
var ErrMissingMigrateFunc = errors.New("missing migrate function in migration")
var ErrMissingRollbackFunc = errors.New("missing rollback function in migration")

// ErrNoMigrationsDefined is returned when no migrations are defined.
var ErrNoMigrationsDefined = errors.New("no migrations defined")
var ErrMigrationIDDoesNotExist = errors.New("tried to migrate to an ID that doesn't exist")
var ErrWrongMigrationFileFormat = errors.New("wrong migration file format")
var ErrWrongMigrationCmdFormat = errors.New("wrong migration cmd format")
