package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

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

	// Timeout configurations (all in seconds)
	ReadTimeout       int // HTTP server read timeout (default: 30)
	WriteTimeout      int // HTTP server write timeout (default: 600 for streaming)
	IdleTimeout       int // HTTP server idle timeout (default: 60)
	VaultTimeout      int // Vault operation timeout (default: 30)
	CommandTimeout    int // Command execution timeout (default: 300)
	SSHConnectTimeout int // SSH connection timeout (default: 30)

	// Audit logging
	AuditLogPath string // Path to audit log file (empty to disable)
}

// GetReadTimeout returns the read timeout as a time.Duration
func (c *Config) GetReadTimeout() time.Duration {
	if c.ReadTimeout <= 0 {
		return 30 * time.Second
	}
	return time.Duration(c.ReadTimeout) * time.Second
}

// GetWriteTimeout returns the write timeout as a time.Duration
func (c *Config) GetWriteTimeout() time.Duration {
	if c.WriteTimeout <= 0 {
		return 10 * time.Minute // Default for streaming
	}
	return time.Duration(c.WriteTimeout) * time.Second
}

// GetIdleTimeout returns the idle timeout as a time.Duration
func (c *Config) GetIdleTimeout() time.Duration {
	if c.IdleTimeout <= 0 {
		return 60 * time.Second
	}
	return time.Duration(c.IdleTimeout) * time.Second
}

// GetVaultTimeout returns the Vault timeout as a time.Duration
func (c *Config) GetVaultTimeout() time.Duration {
	if c.VaultTimeout <= 0 {
		return 30 * time.Second
	}
	return time.Duration(c.VaultTimeout) * time.Second
}

// GetCommandTimeout returns the command execution timeout as a time.Duration
func (c *Config) GetCommandTimeout() time.Duration {
	if c.CommandTimeout <= 0 {
		return 5 * time.Minute
	}
	return time.Duration(c.CommandTimeout) * time.Second
}

// GetSSHConnectTimeout returns the SSH connection timeout as a time.Duration
func (c *Config) GetSSHConnectTimeout() time.Duration {
	if c.SSHConnectTimeout <= 0 {
		return 30 * time.Second
	}
	return time.Duration(c.SSHConnectTimeout) * time.Second
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

	// Timeout defaults (in seconds)
	v.SetDefault("read_timeout", 30)
	v.SetDefault("write_timeout", 600) // 10 minutes for streaming
	v.SetDefault("idle_timeout", 60)
	v.SetDefault("vault_timeout", 30)
	v.SetDefault("command_timeout", 300) // 5 minutes
	v.SetDefault("ssh_connect_timeout", 30)
	v.SetDefault("audit_log_path", "") // Empty to disable audit logging

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

	// Timeout environment variables
	v.BindEnv("read_timeout", "READ_TIMEOUT", "WEBCLI_READ_TIMEOUT")
	v.BindEnv("write_timeout", "WRITE_TIMEOUT", "WEBCLI_WRITE_TIMEOUT")
	v.BindEnv("idle_timeout", "IDLE_TIMEOUT", "WEBCLI_IDLE_TIMEOUT")
	v.BindEnv("vault_timeout", "VAULT_TIMEOUT", "WEBCLI_VAULT_TIMEOUT")
	v.BindEnv("command_timeout", "COMMAND_TIMEOUT", "WEBCLI_COMMAND_TIMEOUT")
	v.BindEnv("ssh_connect_timeout", "SSH_CONNECT_TIMEOUT", "WEBCLI_SSH_CONNECT_TIMEOUT")

	// Audit logging
	v.BindEnv("audit_log_path", "AUDIT_LOG_PATH", "WEBCLI_AUDIT_LOG_PATH")

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

		// Timeout values
		ReadTimeout:       v.GetInt("read_timeout"),
		WriteTimeout:      v.GetInt("write_timeout"),
		IdleTimeout:       v.GetInt("idle_timeout"),
		VaultTimeout:      v.GetInt("vault_timeout"),
		CommandTimeout:    v.GetInt("command_timeout"),
		SSHConnectTimeout: v.GetInt("ssh_connect_timeout"),

		// Audit logging
		AuditLogPath: v.GetString("audit_log_path"),
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
