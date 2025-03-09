.PHONY: dc build build-client build-server run run-client run-server test lint install-lint generate test-server-container

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

# Run client and server in Windows
run:
	@echo "Starting server and client in separate windows..."
	@run.bat

run-client:
	go run -race cmd/client/main.go

run-server:
	go run -race cmd/server/main.go

# Run tests
test:
	go test -race ./...

# Run server container test
test-server-container:
	docker build -t wisdom-pow-server:latest -f Dockerfile.server .
	go test -race -v ./internal/server/tcpserver -run TestServerContainer

install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.6

# Run linter
lint:
	golangci-lint run ./...
