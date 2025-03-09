package tcpserver

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"wisdom-pow/internal/server/config"

	"github.com/rs/zerolog/log"
)

// Server represents a TCP server
type Server struct {
	handler   Handler
	ctx       context.Context
	wg        sync.WaitGroup
	pool      *Pool
	workerNum int
	maxTasks  int
	config    config.Config
}

func NewServer(ctx context.Context, handler Handler, cfg config.Config) *Server {
	pool := New(ctx, cfg.WorkerCount, cfg.MaxTasks)
	server := &Server{
		handler:   handler,
		ctx:       ctx,
		workerNum: cfg.WorkerCount,
		maxTasks:  cfg.MaxTasks,
		pool:      pool,
		config:    cfg,
	}

	pool.Start(server.processConnection)

	return server
}

// AcceptConnections handles accepting new TCP connections
func (s *Server) AcceptConnections() error {
	addr := s.config.GetServerAddr()
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	log.Info().
		Str("address", addr).
		Msg("TCP Server started")

	errChan := make(chan error, 1)

	go func() {
		defer func(listener net.Listener) {
			if err := listener.Close(); err != nil {
				errChan <- fmt.Errorf("error closing listener: %w", err)
			}
		}(listener)

		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				conn, err := listener.Accept()

				if err != nil {
					if s.ctx.Err() != nil {
						return
					}
					errChan <- fmt.Errorf("error accepting connection: %w", err)
					continue
				}

				if err := s.HandleConnection(conn); err != nil {
					errChan <- fmt.Errorf("error handling connection: %w", err)
				}
			}
		}
	}()

	select {
	case err := <-errChan:
		return err
	default:
		return nil
	}
}

// HandleConnection adds the connection to the worker pool
func (s *Server) HandleConnection(conn net.Conn) error {
	if err := s.pool.AddTask(conn); err != nil {
		log.Error().Err(err).Msg("Failed to add connection to worker pool")

		if closeErr := conn.Close(); closeErr != nil {
			log.Error().Err(closeErr).Msg("Failed to close connection after pool error")
		}
		return err
	}
	return nil
}

// GracefulShutdown handles graceful shutdown
func (s *Server) GracefulShutdown(cancel context.CancelFunc) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Info().Msg("Received shutdown signal")

	cancel()
	s.pool.Close()
	s.wg.Wait()

	log.Info().Msg("All connections completed, TCP server shutdown successful")

	return nil
}

// processConnection processes each client connection within a worker
func (s *Server) processConnection(workerId int, conn net.Conn) {
	s.wg.Add(1)
	defer s.wg.Done()
	defer func() {
		if err := conn.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close connection")
		}
	}()

	log.Info().Int("worker_id", workerId).Str("remote_addr", conn.RemoteAddr().String()).Msg("Processing connection")
	s.handler.HandleClient(conn)
}
