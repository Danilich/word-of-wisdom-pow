.PHONY: dc build build-client build-server run test lint install-lint

# Docker compose
dc:
	docker-compose up --remove-orphans --build

# Build both client and server
build: build-client build-server

# Build client
build-client:
	go build -race -o bin/client cmd/client/main.go

# Build server
build-server:
	go build -race -o bin/server cmd/server/main.go

# Run client and server
run: run-server run-client

# Run tests
test:
	go test -race ./...

install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6

# Run linter
lint:
	golangci-lint run ./...
