# Word of Wisdom (Tcp server)

A TCP server and client implementation demonstrating Proof of Work (PoW) using HashCash algorithm for DDoS protection.

## Task Description

Design and implement "Word of Wisdom" tcp server.  
 • TCP server should be protected from DDOS attacks with the Prof of Work (https://en.wikipedia.org/wiki/Proof_of_work), the challenge-response protocol should be used.  
 • The choice of the POW algorithm should be explained.  
 • After Prof Of Work verification, server should send one of the quotes from "word of wisdom" book or any other collection of the quotes.  
 • Docker file should be provided both for the server and for the client that solves the POW challenge.

## Overview

This project implements a simple TCP server that provides "wisdom quotes" to clients. To prevent abuse and DDoS attacks, the server requires clients to solve a Proof of Work challenge before they can request quotes.

## Why HashCash?

HashCash was chosen as the Proof of Work algorithm for several key reasons:

1. **Asymmetric Workload**: HashCash creates an asymmetric computational burden - it's hard for clients to solve but easy for the server to verify. This makes it ideal for protecting against DDoS attacks.

2. **Tunable Difficulty**: The difficulty level can be adjusted based on server load or threat level, allowing for dynamic resource protection.

3. **Real-World Proven**: HashCash has been battle-tested in production systems like Bitcoin and email spam prevention systems.

## Implementation

In my implementation, when a client connects:

1. The server sends a challenge (random data) and difficulty level (e.g., 20 bits)
2. The client must find a solution (nonce) where:
   ```
   SHA256(challenge + nonce) has at least 20 leading zero bits
   ```
3. Finding this solution requires significant computational work (~2^20 hash attempts on average)
4. The server can verify the solution with a single hash operation

At difficulty 20, a typical laptop might take 1-2 seconds to solve the challenge, while the server verifies it in microseconds. This creates an effective computational barrier against automated attacks.

## Core Components

### 1. HashCash Implementation (`pkg/pow`)
- Lightweight implementation of the HashCash algorithm
- Configurable difficulty levels
- Bit manipulation for leading zero verification

### 2. TCP Server (`internal/server/tcpserver`)
- Concurrent connection handling using goroutines
- Worker pool pattern for efficient connection management
- Context-based cancellation for graceful shutdown
- Configurable connection timeouts

### 3. Quote Service (`internal/server/services`)
- Clean service layer that abstracts quote retrieval logic
- Repository pattern for data access
- Context-aware operations for proper cancellation

### 4. Repository Layer (`internal/server/repository`)
- In-memory storage of wisdom quotes
- Extensible design for alternative storage backends

### 5. Client Implementation (`internal/client`)
- Connection management with error handling
- PoW challenge solving
- Buffered I/O for efficient network


## Project Structure

```
wisdom-pow/
├── cmd/                
│   ├── client/         # Client executable
│   └── server/         # Server executable
├── internal/           # Private application code
│   ├── client/         # Client implementation
│   └── server/         # Server implementation
│       ├── domain/     # Domain models
│       ├── repository/ # Data access layer
│       ├── services/   # Business logic
│       └── tcpserver/  # TCP server implementation
├── pkg/                # Common libraries
```

## Environment Configuration

The project uses `.env.server` and `.env.client` files for configuration. These files contain settings for network addresses, timeouts, worker counts, and PoW difficulty levels. They're automatically loaded when running via make commands or Docker.

## How to Run

The project includes several make commands for easy execution:

```bash
# Start both server and client using Docker
make dc           # Runs docker-compose with the app containers (server on port 8080)

# Run locally without Docker
make run          # Runs both server and client locally
make run-server   # Runs only the server locally on port 8080
make run-client   # Runs only the client locally

# Development
make test         # Runs all tests with race detection
make install-lint # Install linter
make lint         # Runs the linter
make build        # Builds both client and server binaries
```

## Container Testing

Run server container test:

```bash
make test-server-container
```

This test builds a Docker image of the server and runs integration tests against it using testcontainers. It verifies:
- Successful quote requests
- Handling of invalid PoW solutions
- Response to unknown commands
- Multiple parallel connections

## GitHub Actions

This project uses GitHub Actions for continuous integration and delivery:

### Go Workflow
The Go workflow (`wisdom-job.yml`) runs on every push to main/master and pull requests:
- Builds the application using `make build`
- Builds Docker images for testing
- Runs all tests with `make test`
- Performs linting with `make lint`

### Docker Workflow
The Docker workflow (`docker.yml`) handles container image building and publishing
