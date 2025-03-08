package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Default server settings
const (
	DefaultTCPAddr           = "0.0.0.0"
	DefaultPort              = "8080"
	DefaultConnectionTimeout = 10 * time.Second
	DefaultLogLevel          = "info"
	DefaultWorkerCount       = 4
	DefaultMaxTasks          = 100
	DefaultDifficulty        = 22
)

// Config for TCP server.
type Config struct {
	TCPAddr           string
	Port              string
	ConnectionTimeout time.Duration
	LogLevel          string
	WorkerCount       int
	MaxTasks          int
	Difficulty        uint8
}

func Read() (Config, error) {
	var config Config

	tcpAddr, exists := os.LookupEnv("TCP_ADDR")
	if exists {
		config.TCPAddr = tcpAddr
	} else {
		config.TCPAddr = DefaultTCPAddr
	}

	port, exists := os.LookupEnv("TCP_PORT")
	if exists {
		config.Port = port
	} else {
		config.Port = DefaultPort
	}

	timeoutStr, exists := os.LookupEnv("CONNECTION_TIMEOUT")
	config.ConnectionTimeout = DefaultConnectionTimeout
	if exists {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return config, fmt.Errorf("invalid CONNECTION_TIMEOUT value: %s: %w", timeoutStr, err)
		}
		config.ConnectionTimeout = timeout
	}

	// Worker pool settings
	workerNum, exists := os.LookupEnv("WORKER_NUM")
	config.WorkerCount = DefaultWorkerCount
	if exists {
		count, err := strconv.Atoi(workerNum)
		if err != nil {
			return config, fmt.Errorf("invalid WORKER_NUM value: %s: %w", workerNum, err)
		}
		config.WorkerCount = count
	}

	maxTasks, exists := os.LookupEnv("MAX_TASKS")
	config.MaxTasks = DefaultMaxTasks
	if exists {
		count, err := strconv.Atoi(maxTasks)
		if err != nil {
			return config, fmt.Errorf("invalid MAX_TASKS value: %s: %w", maxTasks, err)
		}
		config.MaxTasks = count
	}

	difficultyStr, exists := os.LookupEnv("POW_DIFFICULTY")
	config.Difficulty = DefaultDifficulty
	if exists {
		difficulty, err := strconv.ParseUint(difficultyStr, 10, 8)
		if err != nil {
			return config, fmt.Errorf("invalid POW_DIFFICULTY value: %s: %w", difficultyStr, err)
		}
		config.Difficulty = uint8(difficulty)
	}

	logLevel, exists := os.LookupEnv("LOG_LEVEL")
	if exists {
		config.LogLevel = logLevel
	} else {
		config.LogLevel = DefaultLogLevel
	}

	return config, nil
}

func (c Config) GetServerAddr() string {
	return fmt.Sprintf("%s:%s", c.TCPAddr, c.Port)
}

// GetLogLevel log level to zerolog.Level
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
		return zerolog.InfoLevel
	}
}
