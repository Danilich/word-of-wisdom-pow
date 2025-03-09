package client

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"word-of-wisdom-pow/internal/client/config"
	"word-of-wisdom-pow/pkg/pow"
)

const (
	// ChallengeBufSize is the size of the buffer for reading the Pow challenge
	ChallengeBufSize = 32
	// CmdGetQuote is the command byte sent to request a quote
	CmdGetQuote = byte(1)
	// NewlineDelimiter is used for reading string responses
	NewlineDelimiter = '\n'
	// ErrInvalidSolution is the error for invalid solution
	ErrInvalidSolution = "Invalid"
	// ErrTimeout is the error message for timeout waiting for solution
	ErrTimeout = "Timeout waiting for solution"
)

// Client represents a TCP client
type Client struct {
	conn   net.Conn
	reader *bufio.Reader
	config config.Config
}

func NewClient(conn net.Conn, cfg config.Config) *Client {
	return &Client{
		conn:   conn,
		reader: bufio.NewReader(conn),
		config: cfg,
	}
}

func Connect(cfg config.Config) (*Client, error) {
	conn, err := net.DialTimeout("tcp", cfg.ServerAddr, cfg.ConnectionTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server %s: %w", cfg.ServerAddr, err)
	}
	return NewClient(conn, cfg), nil
}

// Start handles the initial PoW challenge from the server
func (c *Client) Start() error {
	if err := c.setReadDeadline(); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}

	// Read difficulty byte
	difficultyByte := make([]byte, 1)
	if _, err := c.conn.Read(difficultyByte); err != nil {
		return fmt.Errorf("failed to read difficulty: %w", err)
	}
	difficulty := difficultyByte[0]

	// Read challenge
	buffer := make([]byte, ChallengeBufSize)
	n, err := c.conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("failed to read challenge: %w", err)
	}
	challenge := buffer[:n]
	log.Debug().Bytes("challenge", challenge).Msg("Received PoW challenge")

	// Solve and send solution
	solution := c.solvePoW(challenge, difficulty)
	if _, err = c.conn.Write(solution); err != nil {
		return fmt.Errorf("failed to send solution: %w", err)
	}

	// Verify server response
	response, err := c.readResponse()
	if err != nil {
		return err
	}

	// Check for error responses
	switch {
	case strings.Contains(response, ErrInvalidSolution):
		return fmt.Errorf("server rejected solution: %s", response)
	case strings.Contains(response, ErrTimeout):
		return fmt.Errorf("server reported timeout: %s", response)
	}

	return nil
}

// GetQuote requests a quote
func (c *Client) GetQuote() (string, error) {
	if _, err := c.conn.Write([]byte{CmdGetQuote}); err != nil {
		return "", fmt.Errorf("failed to send quote request: %w", err)
	}

	if err := c.setReadDeadline(); err != nil {
		return "", err
	}

	response, err := c.readResponse()
	if err != nil {
		return "", err
	}

	return response, nil
}

// Close closes the client connection
func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// solvePoW solves the proof of work challenge
func (c *Client) solvePoW(seed []byte, difficulty uint8) []byte {
	startTime := time.Now()
	proof := pow.NewHashcash(difficulty).Solve(seed)
	elapsed := time.Since(startTime)

	log.Info().
		Float64("duration_seconds", elapsed.Seconds()).
		Uint8("difficulty", difficulty).
		Msg("Found solution")

	return proof
}

// setReadDeadline sets the read deadline on the connection
func (c *Client) setReadDeadline() error {
	if err := c.conn.SetReadDeadline(time.Now().Add(c.config.ReadTimeout)); err != nil {
		return fmt.Errorf("failed to set read deadline: %w", err)
	}
	return nil
}

// readResponse reads a string response from the server
func (c *Client) readResponse() (string, error) {
	response, err := c.reader.ReadString(NewlineDelimiter)

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return "", fmt.Errorf("timeout waiting for server response: %w", err)
		}
		return "", fmt.Errorf("failed to read server response: %w", err)
	}

	return strings.TrimSpace(response), nil
}
