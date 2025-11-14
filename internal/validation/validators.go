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
