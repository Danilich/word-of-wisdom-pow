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

		pool.Start(func(id int, conn net.Conn) {
			time.Sleep(10 * time.Millisecond)
			conn.Close()
			processed.Done()
		})

		for i := 0; i < 3; i++ {
			err := pool.AddTask(newMockConnForPool())
			if err != nil {
				t.Fatalf("Failed to add task: %v", err)
			}
		}

		processed.Wait()
		pool.Close()
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		pool := New(ctx, 2, 5)
		pool.Start(func(id int, conn net.Conn) {
			conn.Close()
		})

		cancel()

		// Adding task should fail
		err := pool.AddTask(newMockConnForPool())
		if err == nil || err.Error() != "context done" {
			t.Fatalf("Expected 'context done' error, got: %v", err)
		}

		pool.Close()
	})

	t.Run("Pool closed", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		pool := New(ctx, 2, 5)
		pool.Start(func(id int, conn net.Conn) {
			conn.Close()
		})

		pool.Close()

		// Adding task should fail
		err := pool.AddTask(newMockConnForPool())
		if err == nil || err.Error() != "pool closed" {
			t.Fatalf("Expected 'pool closed' error, got: %v", err)
		}
	})
}
