package tcpserver

import (
	"context"
	"fmt"
	"net"
	"sync"
)

// Pool manages a pool of workers
type Pool struct {
	workers   int
	taskQueue chan net.Conn
	ctx       context.Context
	mu        sync.Mutex
	closed    bool
	wg        sync.WaitGroup
}

func New(ctx context.Context, workers, maxTasks int) *Pool {
	return &Pool{
		workers:   workers,
		taskQueue: make(chan net.Conn, maxTasks),
		ctx:       ctx,
	}
}

// Start initializes the worker pool
func (p *Pool) Start(handler func(net.Conn)) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-p.ctx.Done():
					return
				case conn, ok := <-p.taskQueue:
					if !ok || conn == nil {
						return
					}
					func() {
						defer func() { recover() }()
						handler(conn)
					}()
				}
			}
		}()
	}
}

// AddTask adds a new connection to worker pool
func (p *Pool) AddTask(conn net.Conn) error {
	p.mu.Lock()
	closed := p.closed
	p.mu.Unlock()

	if closed {
		return fmt.Errorf("pool closed")
	}

	select {
	case p.taskQueue <- conn:
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("context done")
	default:
		return fmt.Errorf("queue full")
	}
}

// Close shuts down the worker pool and cleans up resources
func (p *Pool) Close() {
	p.mu.Lock()
	if !p.closed {
		p.closed = true
		close(p.taskQueue)
	}
	p.mu.Unlock()
	p.wg.Wait()
}
