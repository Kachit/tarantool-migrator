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
```go
package main

import (
    "fmt"
    "time"
	"embed"
	"context"
	"github.com/tarantool/go-tarantool/v2"
	"github.com/tarantool/go-tarantool/v2/pool"
	tarantool_migrator "github.com/kachit/tarantool-migrator"
)

//go:embed migrations
var LuaFs embed.FS

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

## Coverage
```bash
go test --coverprofile=coverage.out ./... ; go tool cover -func coverage.out ; go tool cover --html=coverage.out -o coverage.html
```
