package tcpserver

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestClientServerContainerInteraction tests the interaction between client and server containers
func TestClientServerContainerInteraction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping container test in short mode")
	}

	ctx := context.Background()

	// Create a network for the containers
	networkName := "wisdom-pow-test-network"
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:           networkName,
			CheckDuplicate: true,
		},
	})
	require.NoError(t, err)
	defer network.Remove(ctx)

	// Start the server container
	serverContainer, err := startServerContainer(ctx, networkName)
	require.NoError(t, err)
	defer serverContainer.Terminate(ctx)

	// Allow some time for the server to fully initialize
	time.Sleep(2 * time.Second)

	// Start the client container and verify it can connect to the server
	clientContainer, err := startClientContainer(ctx, networkName)
	require.NoError(t, err)
	defer clientContainer.Terminate(ctx)

	// Wait for the client to finish (using a timeout)
	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Wait for the client container to complete its task
	// We'll use a polling approach to check if the container has stopped
	err = waitForContainerToExit(waitCtx, clientContainer)
	require.NoError(t, err, "Client container did not exit within the timeout period")

	// Check logs to verify successful interaction
	logReader, err := clientContainer.Logs(ctx)
	require.NoError(t, err)
	defer logReader.Close()

	// Read logs and verify client received a wisdom quote
	logContent, err := io.ReadAll(logReader)
	require.NoError(t, err)
	t.Logf("Client logs: %s", logContent)

	assert.Contains(t, string(logContent), "Received wisdom quote")
}

// waitForContainerToExit polls the container state until it exits or times out
func waitForContainerToExit(ctx context.Context, container testcontainers.Container) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			state, err := container.State(ctx)
			if err != nil {
				return err
			}

			// Check if the container has exited
			if !state.Running {
				if state.ExitCode != 0 {
					return fmt.Errorf("container exited with non-zero code: %d", state.ExitCode)
				}
				return nil
			}
		}
	}
}

// startServerContainer starts a server container
func startServerContainer(ctx context.Context, networkName string) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "wisdom-pow-server:latest",
		ExposedPorts: []string{"8080/tcp"},
		Networks:     []string{networkName},
		NetworkAliases: map[string][]string{
			networkName: {"wisdom-server"},
		},
		Env: map[string]string{
			"TCP_ADDR":           "0.0.0.0",
			"TCP_PORT":           "8080",
			"CONNECTION_TIMEOUT": "30s",
			"WORKER_NUM":         "4",
			"MAX_TASKS":          "100",
			"POW_DIFFICULTY":     "22", // Lower difficulty for faster tests
			"LOG_LEVEL":          "debug",
		},
		WaitingFor: wait.ForListeningPort("8080/tcp"),
	}

	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

// startClientContainer starts a client container that connects to the server
func startClientContainer(ctx context.Context, networkName string) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:    "wisdom-pow-client:latest",
		Networks: []string{networkName},
		Env: map[string]string{
			"SERVER_ADDR":     "wisdom-server",
			"SERVER_PORT":     "8080",
			"REQUEST_TIMEOUT": "30s",
			"LOG_LEVEL":       "debug",
		},
		WaitingFor: wait.ForLog("Connected to server"),
	}

	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}
