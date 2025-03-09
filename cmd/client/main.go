package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"word-of-wisdom-pow/internal/client"
	"word-of-wisdom-pow/internal/client/config"
)

func main() {
	if err := run(); err != nil {
		if errors.Is(err, context.Canceled) {
			log.Info().Msg("Client operation was canceled")
			return
		}
		log.Fatal().Err(err).Msg("Fatal error")
	}
}

func run() error {
	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Load configuration
	cfg, err := config.Read()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Set log level before creating the logger
	zerolog.SetGlobalLevel(cfg.GetLogLevel())
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Setup context with cancellation
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log.Info().Str("server_addr", cfg.ServerAddr).Int("num_clients", cfg.NumClients).Msg("Starting clients")

	// Create an errgroup with a shared context
	g, ctx := errgroup.WithContext(ctx)

	// Run clients concurrently
	for i := 0; i < cfg.NumClients; i++ {
		g.Go(func(id int) func() error {
			return func() error {
				return runClient(ctx, cfg, id)
			}
		}(i))
	}

	// Wait for all goroutines to complete or for an error
	return g.Wait()
}

func runClient(ctx context.Context, cfg config.Config, clientID int) error {
	log.Info().Int("client_id", clientID).Msg("Connecting to server")

	// Connect to server
	c, err := client.Connect(cfg)
	if err != nil {
		return fmt.Errorf("client %d: connection failed: %w", clientID, err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Error().Err(err).Int("client_id", clientID).Msg("Failed to close connection")
		}
	}()

	// Handle PoW challenge
	if err := c.Start(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		return fmt.Errorf("client %d: PoW challenge failed: %w", clientID, err)
	}

	// Get quote
	quote, err := c.GetQuote()
	if err != nil {
		return fmt.Errorf("client %d: failed to get quote: %w", clientID, err)
	}

	log.Info().Int("client_id", clientID).Str("quote", quote).Msg("Received quote")
	return nil
}
