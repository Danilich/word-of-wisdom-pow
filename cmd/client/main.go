package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"word-of-wisdom-pow/internal/client"
	"word-of-wisdom-pow/internal/client/config"
)

func main() {
	if err := run(); err != nil && !errors.Is(err, context.Canceled) {
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

	// Run clients concurrently
	var wg sync.WaitGroup
	errCh := make(chan error, cfg.NumClients)

	for i := 0; i < cfg.NumClients; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			if err := runClient(ctx, cfg, id); err != nil && !errors.Is(err, context.Canceled) {
				errCh <- fmt.Errorf("client %d: %w", id, err)
			}
		}(i)
	}

	// Wait for all clients and collect errors
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Return first error
	for err := range errCh {
		return err // Return on first error
	}

	return nil
}

func runClient(ctx context.Context, cfg config.Config, clientID int) error {
	log.Info().Int("client_id", clientID).Msg("Connecting to server")

	if err := ctx.Err(); err != nil {
		return err
	}

	// Connect to server
	c, err := client.Connect(cfg)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			log.Error().Err(err).Int("client_id", clientID).Msg("Failed to close connection")
		}
	}()

	// Handle PoW challenge
	if err := c.Start(); err != nil {
		return fmt.Errorf("PoW challenge failed: %w", err)
	}

	// Get quote
	quote, err := c.GetQuote()
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}

	log.Info().Int("client_id", clientID).Str("quote", quote).Msg("Received quote")
	return nil
}
