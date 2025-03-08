package main

import (
	"context"
	"fmt"
	"os"
	"time"
	"wisdom-pow/internal/server/tcpserver"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"wisdom-pow/internal/server/config"
	"wisdom-pow/internal/server/repository"
	"wisdom-pow/internal/server/services"
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("Server terminated with error")
	}
}

func run() error {
	// Load configuration from environment
	cfg, err := config.Read()

	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set global log level based on config
	zerolog.SetGlobalLevel(cfg.GetLogLevel())
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create services
	powService := services.CreatePow(cfg.Difficulty)
	quotesRepo := repository.NewInMemoryRepository()
	quotesService := services.NewQuotesService(quotesRepo)

	// Create and configure router
	router := tcpserver.NewRouter(ctx)
	router.AddRoute(tcpserver.HandlerQuote, tcpserver.QuoteHandler(quotesService))

	// Create connection handler
	connectionHandler := tcpserver.NewConnectionHandler(ctx, powService, cfg, router)

	// Create TCP server
	server := tcpserver.NewServer(ctx, connectionHandler, cfg)

	// Start accepting connections
	if err := server.AcceptConnections(); err != nil {
		return fmt.Errorf("failed to accept connections: %w", err)
	}

	return server.GracefulShutdown(cancel)
}
