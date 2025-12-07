package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	// Clear environment variables that might affect the test
	envVars := []string{"PORT", "HOST", "FRONTEND_PATH", "DATABASE_PATH",
		"ENCRYPTION_KEY_PATH", "TLS_CERT_PATH", "TLS_KEY_PATH", "REQUIRE_HTTPS",
		"WEBCLI_PORT", "WEBCLI_HOST"}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	cfg := Load()

	// Check default values
	if cfg.Port != 7777 {
		t.Errorf("Expected default port 7777, got %d", cfg.Port)
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("Expected default host 0.0.0.0, got %s", cfg.Host)
	}

	if cfg.FrontendPath != "./frontend/build" {
		t.Errorf("Expected default frontend path ./frontend/build, got %s", cfg.FrontendPath)
	}

	if cfg.DatabasePath != "./data/web-cli.db" {
		t.Errorf("Expected default database path ./data/web-cli.db, got %s", cfg.DatabasePath)
	}

	if cfg.EncryptionKeyPath != "./.encryption_key" {
		t.Errorf("Expected default encryption key path ./.encryption_key, got %s", cfg.EncryptionKeyPath)
	}

	if cfg.TLSCertPath != "" {
		t.Errorf("Expected empty TLS cert path by default, got %s", cfg.TLSCertPath)
	}

	if cfg.TLSKeyPath != "" {
		t.Errorf("Expected empty TLS key path by default, got %s", cfg.TLSKeyPath)
	}

	if cfg.RequireHTTPS != false {
		t.Errorf("Expected RequireHTTPS false by default, got %v", cfg.RequireHTTPS)
	}
}

func TestConfigFromEnvironment(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "8080")
	os.Setenv("HOST", "127.0.0.1")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("HOST")
	}()

	cfg := Load()

	if cfg.Port != 8080 {
		t.Errorf("Expected port 8080 from env, got %d", cfg.Port)
	}

	if cfg.Host != "127.0.0.1" {
		t.Errorf("Expected host 127.0.0.1 from env, got %s", cfg.Host)
	}
}

func TestConfigGetAddress(t *testing.T) {
	cfg := &Config{
		Host: "localhost",
		Port: 9999,
	}

	expected := "localhost:9999"
	if addr := cfg.GetAddress(); addr != expected {
		t.Errorf("Expected address %s, got %s", expected, addr)
	}
}

func TestConfigTLSEnabled(t *testing.T) {
	testCases := []struct {
		name     string
		certPath string
		keyPath  string
		expected bool
	}{
		{
			name:     "both paths set",
			certPath: "/path/to/cert.pem",
			keyPath:  "/path/to/key.pem",
			expected: true,
		},
		{
			name:     "only cert path",
			certPath: "/path/to/cert.pem",
			keyPath:  "",
			expected: false,
		},
		{
			name:     "only key path",
			certPath: "",
			keyPath:  "/path/to/key.pem",
			expected: false,
		},
		{
			name:     "neither path set",
			certPath: "",
			keyPath:  "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &Config{
				TLSCertPath: tc.certPath,
				TLSKeyPath:  tc.keyPath,
			}

			if result := cfg.TLSEnabled(); result != tc.expected {
				t.Errorf("TLSEnabled() = %v, expected %v", result, tc.expected)
			}
		})
	}
}

func TestConfigFromConfigFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
port: 3000
host: "10.0.0.1"
frontend_path: "/custom/frontend"
database_path: "/custom/db.sqlite"
require_https: true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Change to temp directory so viper finds the config
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Clear environment variables
	envVars := []string{"PORT", "HOST", "FRONTEND_PATH", "DATABASE_PATH", "REQUIRE_HTTPS"}
	for _, v := range envVars {
		os.Unsetenv(v)
	}

	cfg := Load()

	if cfg.Port != 3000 {
		t.Errorf("Expected port 3000 from config file, got %d", cfg.Port)
	}

	if cfg.Host != "10.0.0.1" {
		t.Errorf("Expected host 10.0.0.1 from config file, got %s", cfg.Host)
	}

	if cfg.FrontendPath != "/custom/frontend" {
		t.Errorf("Expected frontend path /custom/frontend from config file, got %s", cfg.FrontendPath)
	}

	if cfg.RequireHTTPS != true {
		t.Errorf("Expected RequireHTTPS true from config file, got %v", cfg.RequireHTTPS)
	}
}

func TestConfigPrefixedEnvVars(t *testing.T) {
	// Test WEBCLI_ prefixed environment variables
	os.Setenv("WEBCLI_PORT", "4444")
	os.Setenv("WEBCLI_HOST", "192.168.1.1")
	defer func() {
		os.Unsetenv("WEBCLI_PORT")
		os.Unsetenv("WEBCLI_HOST")
		os.Unsetenv("PORT")
		os.Unsetenv("HOST")
	}()

	cfg := Load()

	if cfg.Port != 4444 {
		t.Errorf("Expected port 4444 from WEBCLI_PORT env, got %d", cfg.Port)
	}

	if cfg.Host != "192.168.1.1" {
		t.Errorf("Expected host 192.168.1.1 from WEBCLI_HOST env, got %s", cfg.Host)
	}
}
