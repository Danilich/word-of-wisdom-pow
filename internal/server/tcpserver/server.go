package tcpserver

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"word-of-wisdom-pow/internal/server/config"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

// Server represents a TCP server
type Server struct {
	handler   Handler
	wg        sync.WaitGroup
	pool      *Pool
	workerNum int
	maxTasks  int
	config    config.Config
}

func NewServer(handler Handler, cfg config.Config) *Server {
	server := &Server{
		handler:   handler,
		workerNum: cfg.WorkerCount,
		maxTasks:  cfg.MaxTasks,
		config:    cfg,
	}

	return server
}

// AcceptConnections handles accepting new TCP connections
func (s *Server) AcceptConnections(ctx context.Context) error {
	// Initialize the worker pool
	s.pool = New(ctx, s.workerNum, s.maxTasks)
	s.pool.Start(s.processConnection)

	addr := s.config.GetServerAddr()
	listener, err := net.Listen("tcp", addr)

	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	log.Info().
		Str("address", addr).
		Msg("TCP Server started")

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		defer func(listener net.Listener) {
			if err := listener.Close(); err != nil {
				log.Error().Err(err).Msg("Error closing listener")
			}
		}(listener)

		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				conn, err := listener.Accept()

				if err != nil {
					if ctx.Err() != nil {
						return nil
					}
					return fmt.Errorf("error accepting connection: %w", err)
				}

				if err := s.HandleConnection(conn); err != nil {
					return fmt.Errorf("error handling connection: %w", err)
				}
			}
		}
	})

	// Wait for any errors from the goroutine
	return g.Wait()
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
