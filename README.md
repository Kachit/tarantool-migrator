# Tarantool Migrator
[![Go Test](https://github.com/Kachit/tarantool-migrator/actions/workflows/tests.yml/badge.svg)](https://github.com/Kachit/tarantool-migrator/actions)
[![Codecov](https://codecov.io/github/Kachit/tarantool-migrator/graph/badge.svg)](https://codecov.io/github/Kachit/tarantool-migrator)
[![Go Report Card](https://goreportcard.com/badge/github.com/kachit/tarantool-migrator)](https://goreportcard.com/report/github.com/kachit/tarantool-migrator)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Kachit/tarantool-migrator)](https://go.dev/doc/go1.20)
[![Release](https://img.shields.io/github/v/release/Kachit/tarantool-migrator.svg)](https://github.com/Kachit/tarantool-migrator/releases)
[![GoDoc](https://img.shields.io/badge/go.dev-reference-brightgreen?logo=go&logoColor=white&style=flat)](https://pkg.go.dev/github.com/kachit/tarantool-migrator)
[![License](https://img.shields.io/github/license/Kachit/tarantool-migrator)](https://github.com/Kachit/tarantool-migrator/blob/main/LICENSE)

## Description
Simple Tarantool migrator written in golang

## Download
```shell
go get -u github.com/kachit/tarantool-migrator
```

## Usage

### Migrations as lua files
Lua migration file format: `{migration-name}.{migration-cmd}.lua`
* `{migration-name}` - migration ID in migrations space
* `{migration-cmd}` - (`up` or `down`) migration command
* `--` in start of migration filename - exclude migration from list

**NOTICE**: When migrations loaded from filesystem they sorted by `{migration-name}` automatically
```
|-- migrations
    |-- 202410082345_test_migration_1.down.lua // 202410082345_test_migration_1 Down cmd
    |-- 202410082345_test_migration_1.up.lua // 202410082345_test_migration_1 Up cmd
    |-- 202410091201_test_migration_2.down.lua // 202410091201_test_migration_2 Down cmd
    |-- 202410091201_test_migration_2.up.lua // 202410091201_test_migration_2 Up cmd
    |-- --202410091545_test_migration_3.up.lua //excluded migration
```

### Migrations as go slice
**NOTICE**: When migrations built as go slice they order will not change
```go
package migrations

import (
	tarantool_migrator "github.com/kachit/tarantool-migrator"
	"github.com/tarantool/go-tarantool/v2/pool"
	"context"
)

var Migrations = tarantool_migrator.MigrationsCollection{
	&tarantool_migrator.Migration{
		ID: "202410082345_test_migration_1",
		Migrate: func(ctx context.Context, pooler pool.Pooler, options *tarantool_migrator.Options) error {
			//your migration Up code here
			return nil
		},
		Rollback: func(ctx context.Context, pooler pool.Pooler, options *tarantool_migrator.Options) error {
			//your migration Down code here
			return nil
		},
	},
}
```

### Migrations in tarantool

By default, tarantool migrations stored in `migrations` space, but you can change it in options

Example:

| **id**                        | **executed_at**                       |
|-------------------------------|---------------------------------------|
| 202410082345_test_migration_1 | 2024-10-11 20:33:28.3860465 +0000 UTC |
| 202410091201_test_migration_2 | 2024-10-11 20:33:28.3860465 +0000 UTC |

### Let's connect to tarantool
```go
package main

import (
    "fmt"
    "time"
	"context"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
)

func main(){
	//Your tarantool config
	config := &TarantoolConfig{
		Host:     "127.0.0.1",
		Port:     3301,
		Login:    "login",
		Password: "password",
	}
	ctx := context.Background()
	dialer := tarantool.NetDialer{
		Address:  fmt.Sprintf("%s:%d", config.Host, config.Port),
		User:     config.Login,
		Password: config.Password,
	}
	opts := tarantool.Opts{
		Timeout:   time.Second,
		Reconnect: 2 * time.Second,
	}
	
	//connect to tarantool pool
	instance1 := pool.Instance{Name: "your-instance-address", Dialer: dialer, Opts: opts}
	tt, err :=  pool.Connect(ctx, []pool.Instance{
		instance1,
	})
	if err != nil {
		panic(err)
	}
}
```

### Let's starting migrate
```go
package main

import (
	"embed"
	"context"
	tarantool_migrator "github.com/kachit/tarantool-migrator"
)

//go:embed migrations
var LuaFs embed.FS

func main(){
	ctx := context.Background()
	//load migrations list from embed
	fsLoader := tarantool_migrator.NewEmbedFsLoader(LuaFs)
	migrations, err := fsLoader.LoadMigrations("migrations")
	if err != nil {
		panic(err)
	}

	//migrate
	migrator := tarantool_migrator.NewMigrator(tt, migrations)
	err = migrator.Migrate(ctx)
	if err != nil {
		panic(err)
	}
}
```

### Let's rollback latest migration
```go
package main

import (
	"embed"
	"context"
	tarantool_migrator "github.com/kachit/tarantool-migrator"
)

//go:embed migrations
var LuaFs embed.FS

func main(){
	ctx := context.Background()
	//load migrations list from embed
	fsLoader := tarantool_migrator.NewEmbedFsLoader(LuaFs)
	migrations, err := fsLoader.LoadMigrations("migrations")
	if err != nil {
		panic(err)
	}

	//rollback last applied migration (202410091201_test_migration_2)
	migrator := tarantool_migrator.NewMigrator(tt, migrations)
	err = migrator.RollbackLast(ctx)
	if err != nil {
		panic(err)
	}
}
```

## Coverage
```bash
go test --coverprofile=coverage.out ./... ; go tool cover -func coverage.out ; go tool cover --html=coverage.out -o coverage.html
```

## Linters ##
```bash
golangci-lint run
```
