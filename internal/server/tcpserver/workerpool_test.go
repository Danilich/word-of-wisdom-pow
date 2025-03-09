package tcpserver

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

type mockConnForPool struct {
	closed bool
}

func (m *mockConnForPool) Read(b []byte) (n int, err error)   { return 0, nil }
func (m *mockConnForPool) Write(b []byte) (n int, err error)  { return len(b), nil }
func (m *mockConnForPool) Close() error                       { m.closed = true; return nil }
func (m *mockConnForPool) LocalAddr() net.Addr                { return nil }
func (m *mockConnForPool) RemoteAddr() net.Addr               { return nil }
func (m *mockConnForPool) SetDeadline(t time.Time) error      { return nil }
func (m *mockConnForPool) SetReadDeadline(t time.Time) error  { return nil }
func (m *mockConnForPool) SetWriteDeadline(t time.Time) error { return nil }

func newMockConnForPool() *mockConnForPool {
	return &mockConnForPool{}
}

func TestWorkerPool(t *testing.T) {
	t.Run("Basic operation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pool := New(ctx, 2, 5)

		var processed sync.WaitGroup
		processed.Add(3)

		pool.Start(func(workerID int, conn net.Conn) {
			time.Sleep(10 * time.Millisecond)
			if err := conn.Close(); err != nil {
				t.Errorf("Failed to close connection: %v", err)
			}
			processed.Done()
		})

		for i := 0; i < 3; i++ {
			err := pool.AddTask(newMockConnForPool())
			if err != nil {
				t.Errorf("Failed to add task: %v", err)
			}
		}

		processed.Wait()
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		pool := New(ctx, 1, 1)

		var wg sync.WaitGroup
		wg.Add(1)

		var handlerStarted sync.WaitGroup
		handlerStarted.Add(1)

		mockConn := newMockConnForPool()

		pool.Start(func(workerID int, conn net.Conn) {
			handlerStarted.Done()

			if conn == mockConn {
				defer wg.Done()
			}

			time.Sleep(1 * time.Millisecond)
			conn.Close()
		})

		// Add first task
		if err := pool.AddTask(mockConn); err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}

		// Wait for handler to start
		handlerStarted.Wait()

		// Add second task to fill queue
		secondConn := newMockConnForPool()
		if err := pool.AddTask(secondConn); err != nil {
			t.Fatalf("Failed to add queue-filling task: %v", err)
		}

		// Cancel context and verify new tasks are rejected
		cancel()
		if err := pool.AddTask(newMockConnForPool()); err == nil {
			t.Error("Expected error after context cancellation")
		}

		wg.Wait()
		if !mockConn.closed {
			t.Error("Connection was not closed")
		}
	})

	t.Run("Pool closure", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pool := New(ctx, 2, 5)

		var wg sync.WaitGroup
		wg.Add(1)

		mockConn := newMockConnForPool()

		pool.Start(func(workerID int, conn net.Conn) {
			defer wg.Done()
			conn.Close()
		})

		if err := pool.AddTask(mockConn); err != nil {
			t.Fatalf("Failed to add task: %v", err)
		}

		wg.Wait()
		if !mockConn.closed {
			t.Error("Connection was not closed")
		}

		pool.Close()
		if err := pool.AddTask(newMockConnForPool()); err == nil {
			t.Error("Expected error after pool closure")
		}
	})
}
