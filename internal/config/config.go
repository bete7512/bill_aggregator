// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config represents the application configuration
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	ProviderAPI ProviderAPIConfig
	Logger      LoggerConfig
	JWT         JWTConfig
}
type LoggerConfig struct {
	Format string
	Level  string
	Output string
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Address string
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	ConnectionString    string
	MaxOpenConns        int
	MaxIdleConns        int
	ConnMaxLifetimeSecs int
}

// ProviderAPIConfig represents the provider API configuration
type ProviderAPIConfig struct {
	BaseURL string
}


// Load loads the configuration from environment variables
func Load() (*Config, error) {
	// Default values
	serverAddress := getEnv("SERVER_ADDRESS", ":8080")

	dbConnectionString := getEnv("DB_CONNECTION_STRING", "postgres://postgres:password@localhost:5432/billed_aggregator")

	dbMaxOpenConns, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "10"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_OPEN_CONNS: %w", err)
	}

	dbMaxIdleConns, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_MAX_IDLE_CONNS: %w", err)
	}

	dbConnMaxLifetimeSecs, err := strconv.Atoi(getEnv("DB_CONN_MAX_LIFETIME_SECS", "300"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME_SECS: %w", err)
	}

	providerAPIBaseURL := getEnv("PROVIDER_API_BASE_URL", "http://utility-providers-api:8081")

	loggerFormat := getEnv("LOGGER_FORMAT", "json")
	loggerLevel := getEnv("LOGGER_LEVEL", "info")
	loggerOutput := getEnv("LOGGER_OUTPUT", "stdout")

	jwtSecret := getEnv("JWT_SECRET", "secret")
	jwtExpiration, err := strconv.Atoi(getEnv("JWT_EXPIRATION", "3600"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION: %w", err)
	}

	// Create config
	config := &Config{
		Server: ServerConfig{
			Address: serverAddress,
		},
		Database: DatabaseConfig{
			ConnectionString:    dbConnectionString,
			MaxOpenConns:        dbMaxOpenConns,
			MaxIdleConns:        dbMaxIdleConns,
			ConnMaxLifetimeSecs: dbConnMaxLifetimeSecs,
		},
		ProviderAPI: ProviderAPIConfig{
			BaseURL: providerAPIBaseURL,
		},
		Logger: LoggerConfig{
			Format: loggerFormat,
			Level:  loggerLevel,
			Output: loggerOutput,
		},
		JWT: JWTConfig{
			Secret:     jwtSecret,
			Expiration: time.Duration(jwtExpiration) * time.Second,
		},
	}

	return config, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
