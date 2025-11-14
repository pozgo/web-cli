package validation

import (
	"strings"
	"testing"
)

func TestValidateIPAddress(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{"valid IPv4", "192.168.1.1", false},
		{"valid IPv4 localhost", "127.0.0.1", false},
		{"valid IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"valid IPv6 short", "::1", false},
		{"empty string", "", true},
		{"invalid IPv4", "256.1.1.1", true},
		{"invalid format", "not-an-ip", true},
		{"hostname not IP", "example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPAddress(tt.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPAddress() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateHostname(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		wantErr  bool
	}{
		{"valid simple hostname", "server1", false},
		{"valid FQDN", "mail.example.com", false},
		{"valid with dash", "my-server", false},
		{"valid with numbers", "server123", false},
		{"empty string", "", true},
		{"starts with dash", "-server", true},
		{"ends with dash", "server-", true},
		{"double dash", "my--server", false}, // Actually valid in RFC 1123
		{"too long", strings.Repeat("a", 254), true},
		{"invalid chars", "my_server", true}, // Underscore not allowed in hostname
		{"valid subdomain", "api.v1.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostname(tt.hostname)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHostname() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIPOrHostname(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid IP", "192.168.1.1", false},
		{"valid hostname", "example.com", false},
		{"valid IPv6", "::1", false},
		{"empty string", "", true},
		{"invalid both", "not_valid-256", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIPOrHostname(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIPOrHostname() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"valid port 22", 22, false},
		{"valid port 80", 80, false},
		{"valid port 443", 443, false},
		{"valid port 65535", 65535, false},
		{"valid port 1", 1, false},
		{"zero", 0, true},
		{"negative", -1, true},
		{"too large", 65536, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePort() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSSHPrivateKey(t *testing.T) {
	// Note: Testing with actual SSH key parsing is complex
	// We test the basic validation logic here
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{"empty string", "", true},
		{"missing BEGIN header", "some random text", true},
		{"whitespace only", "   \n\t  ", true},
		// Real key validation happens in ssh.ParsePrivateKey
		// which requires a properly formatted key
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSSHPrivateKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSSHPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid simple", "john", false},
		{"valid with underscore", "john_doe", false},
		{"valid with dash", "john-doe", false},
		{"valid with numbers", "user123", false},
		{"special case root", "root", false},
		{"special case current", "current", false},
		{"empty string", "", true},
		{"starts with number", "1john", true},
		{"too long", strings.Repeat("a", 33), true},
		{"uppercase", "John", true}, // Unix usernames are lowercase
		{"with spaces", "john doe", true},
		{"special chars", "john@doe", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCommandName(t *testing.T) {
	tests := []struct {
		name        string
		commandName string
		wantErr     bool
	}{
		{"valid simple", "backup-db", false},
		{"valid with spaces", "My Command", false},
		{"valid special chars", "command-123_test", false},
		{"empty string", "", true},
		{"too long", strings.Repeat("a", 256), true},
		{"with null byte", "cmd\x00name", true},
		{"with newline", "cmd\nname", true},
		{"with carriage return", "cmd\rname", true},
		{"valid unicode", "备份数据库", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommandName(tt.commandName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCommandName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
