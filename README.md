# Tarantool Migrator
[![Go Test](https://github.com/Kachit/tarantool-migrator/actions/workflows/tests.yml/badge.svg)](https://github.com/Kachit/tarantool-migrator/actions)
[![Codecov](https://codecov.io/github/Kachit/tarantool-migrator/graph/badge.svg)](https://codecov.io/github/Kachit/tarantool-migrator)
[![License](https://img.shields.io/github/license/Kachit/tarantool-migrator)](https://github.com/Kachit/tarantool-migrator/blob/main/LICENSE)

## Description
Simple Tarantool migrator

## Download
```shell
go get -u github.com/kachit/tarantool-migrator
```

## Usage

### Migrations as lua files
```
|-- migrations
    |-- 202410082345_test_migration_1.down.lua // 202410082345_test_migration_1 Down cmd
    |-- 202410082345_test_migration_1.up.lua // 202410082345_test_migration_1 Up cmd
    |-- 202410091201_test_migration_2.down.lua // 202410091201_test_migration_2 Down cmd
    |-- 202410091201_test_migration_2.up.lua // 202410091201_test_migration_2 Up cmd
    |-- --202410091545_test_migration_3.up.lua //excluded migration
```

### Migrations as go files
**NOTICE**: When migrations loaded from filesystem they sorted by name automatically
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

### Let's connect to tarantool
**NOTICE**: When migrations built as go slice they order can`t change
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

	//migrate
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
