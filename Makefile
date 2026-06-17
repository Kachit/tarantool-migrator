.PHONY: test lint coverage generate tidy

test:
	go test ./...

lint:
	golangci-lint run

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func coverage.out
	go tool cover -html=coverage.out -o coverage.html

generate:
	go generate ./...

tidy:
	go mod tidy
