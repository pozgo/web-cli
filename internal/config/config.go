package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	Port              int    // Server port (default: 7777)
	Host              string // Server host (default: 0.0.0.0)
	FrontendPath      string // Path to frontend build files
	DatabasePath      string // Path to SQLite database file
	EncryptionKeyPath string // Path to encryption key file
}

// Load parses command-line flags and environment variables to load configuration
func Load() *Config {
	cfg := &Config{}

	// Define command-line flags
	flag.IntVar(&cfg.Port, "port", getEnvAsInt("PORT", 7777), "Port to listen on")
	flag.StringVar(&cfg.Host, "host", getEnv("HOST", "0.0.0.0"), "Host to bind to")
	flag.StringVar(&cfg.FrontendPath, "frontend", getEnv("FRONTEND_PATH", "./frontend/build"), "Path to frontend build files")
	flag.StringVar(&cfg.DatabasePath, "db", getEnv("DATABASE_PATH", "./data/web-cli.db"), "Path to SQLite database file")
	flag.StringVar(&cfg.EncryptionKeyPath, "encryption-key", getEnv("ENCRYPTION_KEY_PATH", "./.encryption_key"), "Path to encryption key file")

	flag.Parse()

	return cfg
}

// GetAddress returns the full server address (host:port)
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as int or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	var value int
	_, err := fmt.Sscanf(valueStr, "%d", &value)
	if err != nil {
		return defaultValue
	}

	return value
}
