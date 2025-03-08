package tcpserver

import (
	"context"
	"net"
)

// PowService represents the proof of work service interface
type PowService interface {
	GenerateChallenge() []byte
	VerifyProof(seed, proof []byte) bool
}

// QuotesService represents the quotes service interface
type QuotesService interface {
	GetRandomQuote(ctx context.Context) (string, error)
}

// Handler interface for handling client connections
type Handler interface {
	HandleClient(conn net.Conn)
}
