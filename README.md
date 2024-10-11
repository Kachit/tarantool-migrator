# Tarantool Migrator

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
