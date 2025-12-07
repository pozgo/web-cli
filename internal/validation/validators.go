package validation

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"golang.org/x/crypto/ssh"
)

// hostnameRegex validates RFC 1123 hostnames
var hostnameRegex = regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`)

// ValidateIPAddress validates an IPv4 or IPv6 address
func ValidateIPAddress(ip string) error {
	if ip == "" {
		return fmt.Errorf("IP address cannot be empty")
	}

	if net.ParseIP(ip) == nil {
		return fmt.Errorf("invalid IP address format: %s", ip)
	}

	return nil
}

// ValidateHostname validates a hostname according to RFC 1123
func ValidateHostname(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	if len(hostname) > 253 {
		return fmt.Errorf("hostname too long (max 253 characters)")
	}

	if !hostnameRegex.MatchString(hostname) {
		return fmt.Errorf("invalid hostname format: %s", hostname)
	}

	return nil
}

// ValidateIPOrHostname validates either an IP address or hostname
func ValidateIPOrHostname(value string) error {
	if value == "" {
		return fmt.Errorf("IP address or hostname cannot be empty")
	}

	// Try as IP first
	if err := ValidateIPAddress(value); err == nil {
		return nil
	}

	// Try as hostname
	if err := ValidateHostname(value); err == nil {
		return nil
	}

	return fmt.Errorf("invalid IP address or hostname: %s", value)
}

// ValidatePort validates a port number
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port number: %d (must be between 1 and 65535)", port)
	}
	return nil
}

// ValidateSSHPrivateKey validates that a string is a valid SSH private key
func ValidateSSHPrivateKey(keyData string) error {
	if keyData == "" {
		return fmt.Errorf("SSH private key cannot be empty")
	}

	// Trim whitespace
	keyData = strings.TrimSpace(keyData)

	// Check for common key formats
	if !strings.HasPrefix(keyData, "-----BEGIN") {
		return fmt.Errorf("SSH private key must be in PEM format (should start with -----BEGIN)")
	}

	// Try to parse the key
	_, err := ssh.ParsePrivateKey([]byte(keyData))
	if err != nil {
		return fmt.Errorf("invalid SSH private key format: %w", err)
	}

	return nil
}

// ValidateUsername validates a Unix username
func ValidateUsername(username string) error {
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// Allow special cases
	if username == "root" || username == "current" {
		return nil
	}

	// Validate Unix username format
	// Must start with letter or underscore, can contain letters, digits, underscores, hyphens
	// Max 32 characters
	usernameRegex := regexp.MustCompile(`^[a-z_][a-z0-9_-]{0,31}$`)
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("invalid username format: %s (must start with letter/underscore, max 32 chars)", username)
	}

	return nil
}

// ValidateCommandName validates a saved command name
func ValidateCommandName(name string) error {
	if name == "" {
		return fmt.Errorf("command name cannot be empty")
	}

	if len(name) > 255 {
		return fmt.Errorf("command name too long (max 255 characters)")
	}

	// Prevent potential issues with special characters
	if strings.ContainsAny(name, "\x00\n\r") {
		return fmt.Errorf("command name contains invalid characters")
	}

	return nil
}

// envVarNameRegex validates Unix environment variable names
// Must start with letter or underscore, can contain letters, digits, underscores
var envVarNameRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ValidateEnvVarName validates an environment variable name
func ValidateEnvVarName(name string) error {
	if name == "" {
		return fmt.Errorf("environment variable name cannot be empty")
	}

	if len(name) > 255 {
		return fmt.Errorf("environment variable name too long (max 255 characters)")
	}

	if !envVarNameRegex.MatchString(name) {
		return fmt.Errorf("invalid environment variable name: %s (must start with letter/underscore, contain only letters, digits, underscores)", name)
	}

	return nil
}

// ValidateEnvVarValue validates an environment variable value
func ValidateEnvVarValue(value string) error {
	if value == "" {
		return fmt.Errorf("environment variable value cannot be empty")
	}

	// Limit value size to prevent abuse (1MB max)
	if len(value) > 1024*1024 {
		return fmt.Errorf("environment variable value too long (max 1MB)")
	}

	// Prevent null bytes which could cause issues
	if strings.Contains(value, "\x00") {
		return fmt.Errorf("environment variable value contains invalid null character")
	}

	return nil
}

// ValidateBashScriptName validates a bash script name
func ValidateBashScriptName(name string) error {
	if name == "" {
		return fmt.Errorf("script name cannot be empty")
	}

	if len(name) > 255 {
		return fmt.Errorf("script name too long (max 255 characters)")
	}

	// Prevent potential issues with special characters
	if strings.ContainsAny(name, "\x00\n\r") {
		return fmt.Errorf("script name contains invalid characters")
	}

	return nil
}

// ValidateBashScriptContent validates bash script content
func ValidateBashScriptContent(content string) error {
	if content == "" {
		return fmt.Errorf("script content cannot be empty")
	}

	// Limit content size to prevent abuse (10MB max)
	if len(content) > 10*1024*1024 {
		return fmt.Errorf("script content too long (max 10MB)")
	}

	// Prevent null bytes which could cause issues
	if strings.Contains(content, "\x00") {
		return fmt.Errorf("script content contains invalid null character")
	}

	return nil
}

// ValidateBashScriptFilename validates a bash script filename
func ValidateBashScriptFilename(filename string) error {
	// Filename is optional
	if filename == "" {
		return nil
	}

	if len(filename) > 255 {
		return fmt.Errorf("filename too long (max 255 characters)")
	}

	// Prevent path traversal and dangerous characters
	if strings.Contains(filename, "..") {
		return fmt.Errorf("filename cannot contain path traversal sequences")
	}

	if strings.ContainsAny(filename, "/\\:*?\"<>|\x00\n\r") {
		return fmt.Errorf("filename contains invalid characters")
	}

	return nil
}

// ValidateCommand validates a command string for execution
// This performs basic sanitization to prevent common attacks
func ValidateCommand(command string) error {
	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Limit command size to prevent abuse (64KB max)
	if len(command) > 64*1024 {
		return fmt.Errorf("command too long (max 64KB)")
	}

	// Prevent null bytes which could cause issues with C-based commands
	if strings.Contains(command, "\x00") {
		return fmt.Errorf("command contains invalid null character")
	}

	return nil
}
