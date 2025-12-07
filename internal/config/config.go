package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/spf13/viper"
)

var (
	flagsInitialized bool
	flagsMu          sync.Mutex
)

// Config holds the application configuration
type Config struct {
	Port              int    // Server port (default: 7777)
	Host              string // Server host (default: 0.0.0.0)
	FrontendPath      string // Path to frontend build files
	DatabasePath      string // Path to SQLite database file
	EncryptionKeyPath string // Path to encryption key file
	TLSCertPath       string // Path to TLS certificate file (enables HTTPS)
	TLSKeyPath        string // Path to TLS private key file
	RequireHTTPS      bool   // Require HTTPS when auth is enabled (reject HTTP requests)
}

// Load parses command-line flags and environment variables to load configuration
func Load() *Config {
	v := viper.New()

	// Set default values
	v.SetDefault("port", 7777)
	v.SetDefault("host", "0.0.0.0")
	v.SetDefault("frontend_path", "./frontend/build")
	v.SetDefault("database_path", "./data/web-cli.db")
	v.SetDefault("encryption_key_path", "./.encryption_key")
	v.SetDefault("tls_cert_path", "")
	v.SetDefault("tls_key_path", "")
	v.SetDefault("require_https", false)

	// Enable environment variable support
	v.SetEnvPrefix("WEBCLI") // Environment variables will be WEBCLI_PORT, WEBCLI_HOST, etc.
	v.AutomaticEnv()

	// Also support non-prefixed env vars for backward compatibility
	v.BindEnv("port", "PORT", "WEBCLI_PORT")
	v.BindEnv("host", "HOST", "WEBCLI_HOST")
	v.BindEnv("frontend_path", "FRONTEND_PATH", "WEBCLI_FRONTEND_PATH")
	v.BindEnv("database_path", "DATABASE_PATH", "WEBCLI_DATABASE_PATH")
	v.BindEnv("encryption_key_path", "ENCRYPTION_KEY_PATH", "WEBCLI_ENCRYPTION_KEY_PATH")
	v.BindEnv("tls_cert_path", "TLS_CERT_PATH", "WEBCLI_TLS_CERT_PATH")
	v.BindEnv("tls_key_path", "TLS_KEY_PATH", "WEBCLI_TLS_KEY_PATH")
	v.BindEnv("require_https", "REQUIRE_HTTPS", "WEBCLI_REQUIRE_HTTPS")

	// Config file support (optional)
	v.SetConfigName("config")       // config.yaml, config.json, config.toml
	v.SetConfigType("yaml")         // default to yaml
	v.AddConfigPath(".")            // current directory
	v.AddConfigPath("./config")     // config subdirectory
	v.AddConfigPath("/etc/web-cli") // system config directory
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".config", "web-cli")) // user config directory
	}

	// Read config file if it exists (ignore error if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Warning: error reading config file: %v", err)
		}
	} else {
		log.Printf("Using config file: %s", v.ConfigFileUsed())
	}

	// Command-line flags (highest priority) - only define once
	flagsMu.Lock()
	if !flagsInitialized {
		flag.Int("port", v.GetInt("port"), "Port to listen on")
		flag.String("host", v.GetString("host"), "Host to bind to")
		flag.String("frontend", v.GetString("frontend_path"), "Path to frontend build files")
		flag.String("db", v.GetString("database_path"), "Path to SQLite database file")
		flag.String("encryption-key", v.GetString("encryption_key_path"), "Path to encryption key file")
		flag.String("tls-cert", v.GetString("tls_cert_path"), "Path to TLS certificate file (enables HTTPS)")
		flag.String("tls-key", v.GetString("tls_key_path"), "Path to TLS private key file")
		flag.Bool("require-https", v.GetBool("require_https"), "Require HTTPS when auth is enabled")
		flagsInitialized = true
	}
	flagsMu.Unlock()

	if !flag.Parsed() {
		flag.Parse()
	}

	// Bind flags to viper (so flag values override config/env)
	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "port":
			if val, err := strconv.Atoi(f.Value.String()); err == nil {
				v.Set("port", val)
			}
		case "host":
			v.Set("host", f.Value.String())
		case "frontend":
			v.Set("frontend_path", f.Value.String())
		case "db":
			v.Set("database_path", f.Value.String())
		case "encryption-key":
			v.Set("encryption_key_path", f.Value.String())
		case "tls-cert":
			v.Set("tls_cert_path", f.Value.String())
		case "tls-key":
			v.Set("tls_key_path", f.Value.String())
		case "require-https":
			v.Set("require_https", f.Value.String() == "true")
		}
	})

	return &Config{
		Port:              v.GetInt("port"),
		Host:              v.GetString("host"),
		FrontendPath:      v.GetString("frontend_path"),
		DatabasePath:      v.GetString("database_path"),
		EncryptionKeyPath: v.GetString("encryption_key_path"),
		TLSCertPath:       v.GetString("tls_cert_path"),
		TLSKeyPath:        v.GetString("tls_key_path"),
		RequireHTTPS:      v.GetBool("require_https"),
	}
}

// GetAddress returns the full server address (host:port)
func (c *Config) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// TLSEnabled returns true if TLS certificate and key paths are configured
func (c *Config) TLSEnabled() bool {
	return c.TLSCertPath != "" && c.TLSKeyPath != ""
}
