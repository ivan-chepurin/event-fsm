package event_fsm

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Config[T comparable] struct {
	// Logger is a logger instance, required
	Logger *zap.Logger

	// StateDetector is the state detector instance, required
	StateDetector *StateDetector[T]

	// DBConf is the database connection string, required
	DBConf string

	// RedisConf is the redis connection string, required
	RedisConf *Redis

	// AppLabel is the application name to be used in the database connection, required
	AppLabel string

	// MaxOpenConnections is the maximum number of open connections to the database, required
	MaxOpenConnections int

	// MaxIdleConnections is the maximum number of idle connections in the pool, required
	MaxIdleConnections int

	// ConnectionMaxLifetime is the maximum amount of time a connection may be reused, required
	ConnectionMaxLifetime time.Duration
}

type Redis struct {
	URL      string `env:"URL,default=localhost:6379"`
	Password string `env:"PASSWORD,default=qwerty"`
}

func (cfg *Config[T]) check() error {
	if cfg.Logger == nil {
		return fmt.Errorf("Config.Logger is not set")
	}

	if cfg.StateDetector == nil {
		return fmt.Errorf("Config.StateDetector is not set")
	}

	if _, err := cfg.StateDetector.getMainState(); err != nil {
		return fmt.Errorf("Config.StateDetector.getMainState() failed: %w", err)
	}

	if cfg.DBConf == "" {
		return fmt.Errorf("Config.DBConf is not set")
	}

	if cfg.AppLabel == "" {
		return fmt.Errorf("Config.AppLabel is not set")
	}

	if cfg.MaxOpenConnections == 0 {
		return fmt.Errorf("Config.MaxOpenConnections is not set")
	}

	if cfg.MaxIdleConnections == 0 {
		return fmt.Errorf("Config.MaxIdleConnections is not set")
	}

	if cfg.ConnectionMaxLifetime == 0 {
		return fmt.Errorf("Config.ConnectionMaxLifetime is not set")
	}

	return nil
}
