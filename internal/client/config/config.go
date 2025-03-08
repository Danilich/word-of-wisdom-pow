package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Default client settings
const (
	DefaultServerAddr        = "localhost:8080"
	DefaultReadTimeout       = 30 * time.Second
	DefaultConnectionTimeout = 30 * time.Second
	DefaultLogLevel          = "info"
	DefaultNumClients        = 10
)

// Config for the client
type Config struct {
	ServerAddr        string
	ReadTimeout       time.Duration
	ConnectionTimeout time.Duration
	LogLevel          string
	NumClients        int
}

func Read() (Config, error) {
	config := Config{
		ServerAddr:        getEnv("SERVER_ADDR", DefaultServerAddr),
		ReadTimeout:       DefaultReadTimeout,
		ConnectionTimeout: DefaultConnectionTimeout,
		LogLevel:          getEnv("LOG_LEVEL", DefaultLogLevel),
		NumClients:        DefaultNumClients,
	}

	// Parse durations
	if val := os.Getenv("READ_TIMEOUT"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.ReadTimeout = d
		} else {
			return config, fmt.Errorf("invalid READ_TIMEOUT value: %s: %w", val, err)
		}
	}

	if val := os.Getenv("CONNECTION_TIMEOUT"); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			config.ConnectionTimeout = d
		} else {
			return config, fmt.Errorf("invalid CONNECTION_TIMEOUT value: %s: %w", val, err)
		}
	}

	// Parse number of clients
	if val := os.Getenv("NUM_CLIENTS"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			config.NumClients = n
		}
	}

	return config, nil
}

// getEnv gets an environment variable or returns the default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// GetLogLevel converts the string log level to zerolog.Level
func (c Config) GetLogLevel() zerolog.Level {
	switch strings.ToLower(c.LogLevel) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "trace":
		return zerolog.TraceLevel
	default:
		// Log a warning about invalid log level and default to debug to ensure errors are visible
		fmt.Printf("Warning: Invalid log level '%s', defaulting to debug level\n", c.LogLevel)
		return zerolog.DebugLevel
	}
}
