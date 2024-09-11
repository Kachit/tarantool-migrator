package tarantool_migrator

import "errors"

// ErrMissingID is returned when the ID od migration is equal to ""
var ErrMissingID = errors.New("tarantool-migrator: Missing ID in migration")
var ErrMissingMigrateFunc = errors.New("tarantool-migrator: Missing migrate function in migration")
var ErrMissingRollbackFunc = errors.New("tarantool-migrator: Missing rollback function in migration")

// ErrNoMigrationsDefined is returned when no migrations are defined.
var ErrNoMigrationsDefined = errors.New("tarantool-migrator: No migrations defined")
var ErrMigrationIDDoesNotExist = errors.New("tarantool-migrator: Tried to migrate to an ID that doesn't exist")
