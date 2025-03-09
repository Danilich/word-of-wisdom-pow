package tcpserver

import (
	"bytes"
	"context"
	"net"
	"testing"
	"time"

	"word-of-wisdom-pow/internal/server/config"
	"word-of-wisdom-pow/internal/server/services"
)

type mockConn struct {
	readData  bytes.Buffer
	writeData bytes.Buffer
	closed    bool
	addr      net.Addr
}

type mockPow struct {
	challenge    []byte
	solution     []byte
	shouldVerify bool
}

// convert string to byte slice for testing purposes
func strToBytes(s string) []byte {
	return []byte(s)
}

func (m *mockPow) GenerateChallenge() []byte {
	return m.challenge
}

func (m *mockPow) Verify(_, proof []byte) bool {
	if m.shouldVerify {
		return bytes.Equal(proof, m.solution)
	}
	return false
}

func (m *mockPow) Solve(_ context.Context, _ []byte) ([]byte, error) {
	return m.solution, nil
}

func (m *mockConn) Read(b []byte) (n int, err error)   { return m.readData.Read(b) }
func (m *mockConn) Write(b []byte) (n int, err error)  { return m.writeData.Write(b) }
func (m *mockConn) Close() error                       { m.closed = true; return nil }
func (m *mockConn) LocalAddr() net.Addr                { return m.addr }
func (m *mockConn) RemoteAddr() net.Addr               { return m.addr }
func (m *mockConn) SetDeadline(_ time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

type testAddr struct{}

func (a *testAddr) Network() string { return "tcp" }
func (a *testAddr) String() string  { return "127.0.0.1:12345" }

func newMockConn() *mockConn {
	return &mockConn{
		addr: &testAddr{},
	}
}

func TestPowHandler_SendChallenge(t *testing.T) {
	ctx := context.Background()
	powService := services.CreatePow(8)
	cfg := config.Config{
		ConnectionTimeout: 5 * time.Second,
		Difficulty:        8,
	}

	router := NewRouter()
	handler := NewConnectionHandler(ctx, powService, cfg, router)
	conn := newMockConn()
	challenge := strToBytes("challenge")

	// Test
	err := handler.sendChallenge(conn, challenge)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestPowHandler_ReadCommand(t *testing.T) {
	ctx := context.Background()
	powService := services.CreatePow(8)
	cfg := config.Config{
		ConnectionTimeout: 5 * time.Second,
		Difficulty:        8,
	}
	router := NewRouter()
	handler := NewConnectionHandler(ctx, powService, cfg, router)

	conn := newMockConn()
	expectedCmd := byte(42)
	conn.readData.Write([]byte{expectedCmd})

	// Test
	cmd, err := handler.readCommand(conn)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if cmd != int(expectedCmd) {
		t.Errorf("Expected command %d, got %d", expectedCmd, cmd)
	}
}

func TestPowHandler_ProcessPowChallenge_Success(t *testing.T) {
	ctx := context.Background()
	mockPow := &mockPow{
		challenge:    strToBytes("challenge"),
		solution:     strToBytes("solution"),
		shouldVerify: true,
	}
	powService := services.NewPowService(mockPow)
	cfg := config.Config{
		ConnectionTimeout: 5 * time.Second,
		Difficulty:        8,
	}
	router := NewRouter()
	handler := NewConnectionHandler(ctx, powService, cfg, router)

	conn := newMockConn()
	conn.readData.Write(mockPow.solution)

	// Test
	err := handler.processPowChallenge(ctx, conn)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expected := []byte(ResponsePowVerificationSuccess)
	if !bytes.Contains(conn.writeData.Bytes(), expected) {
		t.Errorf("Expected response to contain %q, got %q",
			expected, conn.writeData.Bytes())
	}
}

func TestPowHandler_ProcessPowChallenge_Failure(t *testing.T) {
	ctx := context.Background()
	mockPow := &mockPow{
		challenge:    strToBytes("challenge"),
		solution:     strToBytes("solution"),
		shouldVerify: false,
	}
	powService := services.NewPowService(mockPow)
	cfg := config.Config{
		ConnectionTimeout: 5 * time.Second,
		Difficulty:        8,
	}
	router := NewRouter()
	handler := NewConnectionHandler(ctx, powService, cfg, router)

	conn := newMockConn()
	conn.readData.Write(strToBytes("wrongsolution"))

	// Test
	err := handler.processPowChallenge(ctx, conn)

	if err == nil {
		t.Fatal("Expected an error, got nil")
	}

	expected := []byte(ResponseInvalidPowSolution)
	if !bytes.Contains(conn.writeData.Bytes(), expected) {
		t.Errorf("Expected response to contain %q, got %q",
			expected, conn.writeData.Bytes())
	}
}

func TestPowHandler_ReadSolutionWithTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	powService := services.CreatePow(8)
	cfg := config.Config{
		ConnectionTimeout: 5 * time.Second,
		Difficulty:        8,
	}
	router := NewRouter()
	handler := NewConnectionHandler(ctx, powService, cfg, router)
	conn := newMockConn()

	cancel()

	// Test
	solution, err := handler.readSolutionWithTimeout(ctx, conn)

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if solution != nil {
		t.Errorf("Expected nil solution on timeout, got %v", solution)
	}
}

func TestPowHandler_HandleClient(t *testing.T) {
	ctx := context.Background()
	mockPow := &mockPow{
		challenge:    strToBytes("challenge"),
		solution:     strToBytes("solution"),
		shouldVerify: true,
	}
	powService := services.NewPowService(mockPow)
	cfg := config.Config{
		ConnectionTimeout: 5 * time.Second,
		Difficulty:        8,
	}

	router := NewRouter()
	cmdID := 42
	cmdHandled := false
	router.AddRoute(cmdID, func(ctx context.Context, conn net.Conn) error {
		cmdHandled = true
		return nil
	})

	handler := NewConnectionHandler(ctx, powService, cfg, router)

	conn := newMockConn()
	conn.readData.Write(mockPow.solution)
	conn.readData.Write([]byte{byte(cmdID)})

	// Test
	handler.HandleClient(conn)

	if !cmdHandled {
		t.Error("Command handler was not called")
	}

	expected := []byte(ResponsePowVerificationSuccess)
	if !bytes.Contains(conn.writeData.Bytes(), expected) {
		t.Errorf("Expected response to contain %q, got %q",
			expected, conn.writeData.Bytes())
	}
}

func TestPowHandler_HandleClient_InvalidPow(t *testing.T) {
	ctx := context.Background()
	mockPow := &mockPow{
		challenge:    strToBytes("challenge"),
		solution:     strToBytes("solution"),
		shouldVerify: false,
	}
	powService := services.NewPowService(mockPow)
	cfg := config.Config{
		ConnectionTimeout: 5 * time.Second,
		Difficulty:        8,
	}

	router := NewRouter()
	cmdHandled := false
	router.AddRoute(1, func(ctx context.Context, conn net.Conn) error {
		cmdHandled = true
		return nil
	})

	handler := NewConnectionHandler(ctx, powService, cfg, router)

	conn := newMockConn()
	conn.readData.Write(strToBytes("wrongsolution"))

	// Test
	handler.HandleClient(conn)

	if cmdHandled {
		t.Error("Command handler should not called with invalid PoW")
	}

	expected := []byte(ResponseInvalidPowSolution)
	if !bytes.Contains(conn.writeData.Bytes(), expected) {
		t.Errorf("Expected response to contain %q, got %q",
			expected, conn.writeData.Bytes())
	}
}
