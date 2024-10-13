# Tarantool Migrator
[![Go Test](https://github.com/Kachit/tarantool-migrator/actions/workflows/tests.yml/badge.svg)](https://github.com/Kachit/tarantool-migrator/actions)
[![Codecov](https://codecov.io/github/Kachit/tarantool-migrator/graph/badge.svg?token=LbJ2ALZ0gs)](https://codecov.io/github/Kachit/tarantool-migrator)
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
    migrator "github.com/kachit/tarantool-migrator"
)

func main(){
    
}
```

## Coverage
```bash
go test --coverprofile=coverage.out ./... ; go tool cover -func coverage.out ; go tool cover --html=coverage.out -o coverage.html
```
