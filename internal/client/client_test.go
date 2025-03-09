package client

import (
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"word-of-wisdom-pow/internal/client/config"
)

type mockConn struct {
	readData  []byte
	readIndex int
	writeBuf  []byte
	closed    bool
	readErr   error
	writeErr  error
	response  []byte
}

func (m *mockConn) LocalAddr() net.Addr                { return &net.TCPAddr{} }
func (m *mockConn) RemoteAddr() net.Addr               { return &net.TCPAddr{} }
func (m *mockConn) SetDeadline(t time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(t time.Time) error { return nil }

// TestClientStart tests the Start method
func TestClientStart(t *testing.T) {
	difficulty := byte(5)
	challenge := []byte("TestChallenge123")

	mockConn := &mockConn{
		readData: append([]byte{difficulty}, challenge...),
		response: []byte("OK\n"),
	}

	cfg := config.Config{
		ServerAddr:        "localhost:8080",
		ConnectionTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		LogLevel:          "debug",
	}

	client := NewClient(mockConn, cfg)
	err := client.Start()
	assert.NoError(t, err, "Expected no error from Start")
	assert.Greater(t, len(mockConn.writeBuf), 0, "Expected client to write to connection")
}

// TestGetQuote tests the GetQuote method
func TestGetQuote(t *testing.T) {
	expectedQuote := "Wisdom is the daughter of experience.\n"
	mockConn := &mockConn{
		response: []byte(expectedQuote),
	}

	cfg := config.Config{
		ServerAddr:        "localhost:8080",
		ConnectionTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		LogLevel:          "debug",
	}

	client := NewClient(mockConn, cfg)

	quote, err := client.GetQuote()

	assert.NoError(t, err, "Expected no error from GetQuote")
	assert.Equal(t, strings.TrimSpace(expectedQuote), quote, "Quote does not match expected value")
	assert.Equal(t, []byte{CmdGetQuote}, mockConn.writeBuf, "Client did not send the correct command")
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if m.readErr != nil {
		return 0, m.readErr
	}

	if m.readIndex < len(m.readData) {
		n = copy(b, m.readData[m.readIndex:])
		m.readIndex += n
		return
	}

	if len(m.response) > 0 {
		n = copy(b, m.response)
		m.response = nil
		return
	}

	return 0, io.EOF
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	m.writeBuf = append(m.writeBuf, b...)
	return len(b), nil
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}
