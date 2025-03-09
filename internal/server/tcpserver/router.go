package tcpserver

import (
	"context"
	"fmt"
	"net"
)

type HandlerFunc func(ctx context.Context, conn net.Conn) error

// Router manages command routing for the TCP server
type Router struct {
	routes map[int]HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[int]HandlerFunc),
	}
}

// AddRoute registers a handler for a specific command
func (r *Router) AddRoute(cmdID int, handler HandlerFunc) *Router {
	r.routes[cmdID] = handler
	return r
}

// HandleCommand processes by ID
func (r *Router) HandleCommand(ctx context.Context, cmdID int, conn net.Conn) error {
	handler, exists := r.routes[cmdID]
	if !exists {
		_, err := conn.Write([]byte(ResponseUnknownCommand))
		if err != nil {
			return fmt.Errorf("failed to write unknown command response: %w", err)
		}
		return fmt.Errorf("unknown command: %d", cmdID)
	}

	return handler(ctx, conn)
}
