package tcpserver

import (
	"bytes"
	"context"
	"net"
	"testing"
)

func TestRouter_AddRoute(t *testing.T) {
	ctx := context.Background()
	router := NewRouter(ctx)
	cmdID := 1
	handlerCalled := false

	router.AddRoute(cmdID, func(ctx context.Context, conn net.Conn) error {
		handlerCalled = true
		return nil
	})

	if len(router.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(router.routes))
	}

	handler, exists := router.routes[cmdID]
	if !exists {
		t.Fatalf("Expected route for command %d to exist", cmdID)
	}

	conn := newMockConn()
	err := handler(ctx, conn)

	if err != nil {
		t.Errorf("Expected no error from handler, got %v", err)
	}

	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestRouter_HandleCommand_Success(t *testing.T) {
	ctx := context.Background()
	router := NewRouter(ctx)

	cmdID := 42
	handlerCalled := false
	router.AddRoute(cmdID, func(ctx context.Context, conn net.Conn) error {
		handlerCalled = true
		return nil
	})

	conn := newMockConn()
	err := router.HandleCommand(cmdID, conn)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

func TestRouter_HandleCommand_UnknownCommand(t *testing.T) {
	ctx := context.Background()
	router := NewRouter(ctx)

	// Add a handler for command 1
	router.AddRoute(1, func(ctx context.Context, conn net.Conn) error {
		return nil
	})

	conn := newMockConn()
	unknownCmdID := 99
	err := router.HandleCommand(unknownCmdID, conn)

	if err == nil {
		t.Fatal("Expected error for unknown command, got nil")
	}

	expected := []byte(ResponseUnknownCommand)
	if !bytes.Contains(conn.writeData.Bytes(), expected) {
		t.Errorf("Expected response to contain %q, got %q",
			expected, conn.writeData.Bytes())
	}
}
