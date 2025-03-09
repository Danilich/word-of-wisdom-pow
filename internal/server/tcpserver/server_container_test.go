package tcpserver

import (
	"bufio"
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"wisdom-pow/pkg/pow"
)

// testConnection represents a test connection with helper methods
type testConnection struct {
	t          *testing.T
	conn       net.Conn
	reader     *bufio.Reader
	serverAddr string
}

// newTestConnection creates a new test connection to the server
func newTestConnection(t *testing.T, serverAddr string) *testConnection {
	conn, err := net.Dial("tcp", serverAddr)
	require.NoError(t, err)

	return &testConnection{
		t:          t,
		conn:       conn,
		reader:     bufio.NewReader(conn),
		serverAddr: serverAddr,
	}
}

func TestServerContainer(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "wisdom-pow-server:latest",
		ExposedPorts: []string{"8080/tcp"},
		Env: map[string]string{
			"TCP_ADDR":           "0.0.0.0",
			"TCP_PORT":           "8080",
			"CONNECTION_TIMEOUT": "30s",
			"WORKER_NUM":         "4",
			"MAX_TASKS":          "100",
			"POW_DIFFICULTY":     "22",
			"LOG_LEVEL":          "debug",
		},
		WaitingFor: wait.ForListeningPort("8080/tcp"),
	}

	serverContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	defer func() {
		if err := serverContainer.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}()

	host, err := serverContainer.Host(ctx)
	require.NoError(t, err)
	port, err := serverContainer.MappedPort(ctx, "8080")
	require.NoError(t, err)

	serverAddr := net.JoinHostPort(host, port.Port())
	t.Logf("Connecting to server at: %s", serverAddr)
	time.Sleep(2 * time.Second)

	// Run test cases
	t.Run("SuccessfulQuoteRequest", func(t *testing.T) {
		testSuccessfulQuoteRequest(t, serverAddr)
	})

	t.Run("InvalidPowSolution", func(t *testing.T) {
		testInvalidPowSolution(t, serverAddr)
	})

	t.Run("UnknownCommand", func(t *testing.T) {
		testUnknownCommand(t, serverAddr)
	})

	t.Run("MultipleConnections", func(t *testing.T) {
		testMultipleConnections(t, serverAddr, 5)
	})
}

// testSuccessfulQuoteRequest tests a successful quote request flow
func testSuccessfulQuoteRequest(t *testing.T, serverAddr string) {
	tc := newTestConnection(t, serverAddr)
	defer tc.close()

	tc.solvePowChallenge()
	tc.requestQuote()
}

// testInvalidPowSolution tests the server's response to an invalid PoW solution
func testInvalidPowSolution(t *testing.T, serverAddr string) {
	tc := newTestConnection(t, serverAddr)
	defer tc.close()

	tc.sendInvalidPowSolution()
}

// testUnknownCommand tests the server's response to an unknown command
func testUnknownCommand(t *testing.T, serverAddr string) {
	tc := newTestConnection(t, serverAddr)
	defer tc.close()

	tc.solvePowChallenge()
	tc.sendUnknownCommand()
}

// testMultipleConnections tests multiple parallel connections to the server
func testMultipleConnections(t *testing.T, serverAddr string, numConnections int) {
	for i := 0; i < numConnections; i++ {
		t.Run("Connection", func(t *testing.T) {
			t.Parallel()

			tc := newTestConnection(t, serverAddr)
			defer tc.close()

			tc.solvePowChallenge()
			tc.requestQuote()
		})
	}
}

// solvePowChallenge reads and solves the PoW challenge
func (tc *testConnection) solvePowChallenge() {
	challengeData := make([]byte, 9)
	_, err := tc.conn.Read(challengeData)
	require.NoError(tc.t, err)

	difficulty := challengeData[0]
	challenge := challengeData[1:]
	tc.t.Logf("Received challenge with difficulty %d", difficulty)

	solution := solvePoW(challenge, difficulty)
	_, err = tc.conn.Write(solution)
	require.NoError(tc.t, err)

	// Verify the response
	verificationResponse, err := tc.reader.ReadString('\n')
	require.NoError(tc.t, err)
	assert.Equal(tc.t, ResponsePowVerificationSuccess, verificationResponse)
}

// sendInvalidPowSolution sends an invalid PoW solution
func (tc *testConnection) sendInvalidPowSolution() {
	challengeData := make([]byte, 9)
	_, err := tc.conn.Read(challengeData)
	require.NoError(tc.t, err)

	// Send an invalid solution (all zeros)
	invalidSolution := make([]byte, 8)
	_, err = tc.conn.Write(invalidSolution)
	require.NoError(tc.t, err)

	// Read the error response
	response, err := tc.reader.ReadString('\n')
	require.NoError(tc.t, err)
	assert.Equal(tc.t, ResponseInvalidPowSolution, response)
}

// requestQuote sends a quote request and verifies the response
func (tc *testConnection) requestQuote() string {
	_, err := tc.conn.Write([]byte{HandlerQuote})
	require.NoError(tc.t, err)

	quoteResponse, err := tc.reader.ReadString('\n')
	require.NoError(tc.t, err)
	tc.t.Logf("Received quote: %s", quoteResponse)

	// Verify the response format
	assert.Contains(tc.t, quoteResponse, "Generated wisdom: ")

	return quoteResponse
}

// sendUnknownCommand sends an unknown command and verifies the response
func (tc *testConnection) sendUnknownCommand() {
	_, err := tc.conn.Write([]byte{255})
	require.NoError(tc.t, err)

	response, err := tc.reader.ReadString('\n')
	require.NoError(tc.t, err)
	assert.Equal(tc.t, ResponseUnknownCommand, response)
}

func solvePoW(challenge []byte, difficulty uint8) []byte {
	hashcash := pow.NewHashcash(difficulty)
	return hashcash.Solve(challenge)
}

func (tc *testConnection) close() {
	if err := tc.conn.Close(); err != nil {
		tc.t.Logf("Failed to close connection: %v", err)
	}
}
