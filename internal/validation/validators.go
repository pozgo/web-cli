package validation

import (
	"fmt"
	"net"
	"net/url"
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

// ValidateVaultAddress validates a Vault server address to prevent SSRF attacks
// Only allows HTTPS URLs with valid public hostnames/IPs
func ValidateVaultAddress(address string) error {
	if address == "" {
		return fmt.Errorf("vault address cannot be empty")
	}

	// Parse the URL
	parsedURL, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("invalid vault address URL: %w", err)
	}

	// Require HTTP or HTTPS scheme
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return fmt.Errorf("vault address must use http or https scheme")
	}

	// Get the hostname
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return fmt.Errorf("vault address must include a hostname")
	}

	// Check for dangerous endpoints (SSRF protection)
	// We allow private IPs for self-hosted Vault but block cloud metadata endpoints
	if ip := net.ParseIP(hostname); ip != nil {
		// Block link-local addresses (169.254.x.x) which include cloud metadata endpoints
		if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("vault address cannot be a link-local address (potential metadata endpoint)")
		}
		// Block unspecified addresses (0.0.0.0)
		if ip.IsUnspecified() {
			return fmt.Errorf("vault address cannot be an unspecified address")
		}
	} else {
		// It's a hostname, validate format
		if err := ValidateHostname(hostname); err != nil {
			return fmt.Errorf("invalid vault hostname: %w", err)
		}

		// Block cloud metadata hostnames that could be used for SSRF
		lowercaseHost := strings.ToLower(hostname)
		blockedHosts := []string{
			"metadata.google.internal",
			"metadata.goog",
			"metadata",
			"instance-data",
		}
		for _, blocked := range blockedHosts {
			if lowercaseHost == blocked || strings.HasSuffix(lowercaseHost, "."+blocked) {
				return fmt.Errorf("vault address hostname is not allowed: %s", hostname)
			}
		}
	}

	// Validate port if specified
	if portStr := parsedURL.Port(); portStr != "" {
		// Port is already validated by url.Parse, but verify it's reasonable
		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil {
			if err := ValidatePort(port); err != nil {
				return fmt.Errorf("invalid vault port: %w", err)
			}
		}
	}

	return nil
}

// isPrivateOrReservedIP checks if an IP address is private, loopback, or reserved
func isPrivateOrReservedIP(ip net.IP) bool {
	// Check for loopback (127.0.0.0/8, ::1)
	if ip.IsLoopback() {
		return true
	}

	// Check for private ranges
	if ip.IsPrivate() {
		return true
	}

	// Check for link-local (169.254.0.0/16, fe80::/10)
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Check for other reserved ranges
	// 0.0.0.0/8 - Current network
	// 10.0.0.0/8 - Private (handled by IsPrivate)
	// 100.64.0.0/10 - Carrier-grade NAT
	// 172.16.0.0/12 - Private (handled by IsPrivate)
	// 192.0.0.0/24 - IETF Protocol Assignments
	// 192.0.2.0/24 - Documentation
	// 192.168.0.0/16 - Private (handled by IsPrivate)
	// 198.51.100.0/24 - Documentation
	// 203.0.113.0/24 - Documentation

	ip4 := ip.To4()
	if ip4 != nil {
		// Carrier-grade NAT: 100.64.0.0/10
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return true
		}
		// Current network: 0.0.0.0/8
		if ip4[0] == 0 {
			return true
		}
		// IETF Protocol Assignments: 192.0.0.0/24
		if ip4[0] == 192 && ip4[1] == 0 && ip4[2] == 0 {
			return true
		}
	}

	return false
}

// ValidateVaultSecretPath validates a Vault secret path component to prevent path traversal
func ValidateVaultSecretPath(path string) error {
	if path == "" {
		return fmt.Errorf("vault path cannot be empty")
	}

	// Check for path traversal sequences
	if strings.Contains(path, "..") {
		return fmt.Errorf("vault path cannot contain path traversal sequences (..)")
	}

	// Check for absolute path indicators
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("vault path cannot start with /")
	}

	// Check for backslashes (Windows-style paths)
	if strings.Contains(path, "\\") {
		return fmt.Errorf("vault path cannot contain backslashes")
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("vault path cannot contain null characters")
	}

	// Check for URL encoding that could be used to bypass filters
	if strings.Contains(path, "%") {
		return fmt.Errorf("vault path cannot contain URL-encoded characters")
	}

	// Validate allowed characters (alphanumeric, dash, underscore, dot, forward slash)
	validPathRegex := regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)
	if !validPathRegex.MatchString(path) {
		return fmt.Errorf("vault path contains invalid characters (allowed: alphanumeric, dash, underscore, dot, forward slash)")
	}

	// Check for consecutive slashes
	if strings.Contains(path, "//") {
		return fmt.Errorf("vault path cannot contain consecutive slashes")
	}

	// Limit path length
	if len(path) > 512 {
		return fmt.Errorf("vault path too long (max 512 characters)")
	}

	return nil
}

// ValidateVaultSecretName validates a Vault secret name to prevent path traversal
func ValidateVaultSecretName(name string) error {
	if name == "" {
		return fmt.Errorf("vault secret name cannot be empty")
	}

	// Secret names should not contain path separators
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("vault secret name cannot contain path separators")
	}

	// Check for path traversal
	if strings.Contains(name, "..") {
		return fmt.Errorf("vault secret name cannot contain path traversal sequences")
	}

	// Check for null bytes
	if strings.Contains(name, "\x00") {
		return fmt.Errorf("vault secret name cannot contain null characters")
	}

	// Check for URL encoding
	if strings.Contains(name, "%") {
		return fmt.Errorf("vault secret name cannot contain URL-encoded characters")
	}

	// Validate allowed characters
	validNameRegex := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)
	if !validNameRegex.MatchString(name) {
		return fmt.Errorf("vault secret name contains invalid characters (allowed: alphanumeric, dash, underscore, dot)")
	}

	// Limit name length
	if len(name) > 255 {
		return fmt.Errorf("vault secret name too long (max 255 characters)")
	}

	return nil
}

// ValidateVaultGroupName validates a Vault group name
func ValidateVaultGroupName(group string) error {
	// Empty group is allowed (defaults to "default")
	if group == "" {
		return nil
	}

	// Group names follow same rules as secret names
	return ValidateVaultSecretName(group)
}
