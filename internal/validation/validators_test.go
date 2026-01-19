package validation

import (
	"testing"
)

func TestValidateVaultAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
		errMsg  string
	}{
		// Valid addresses
		{
			name:    "valid https public URL",
			address: "https://vault.example.com:8200",
			wantErr: false,
		},
		{
			name:    "valid https with path",
			address: "https://vault.example.com/v1",
			wantErr: false,
		},
		{
			name:    "valid http local development",
			address: "http://localhost:8200",
			wantErr: false,
		},
		{
			name:    "valid https localhost",
			address: "https://127.0.0.1:8200",
			wantErr: false,
		},
		{
			name:    "valid private IP 10.x.x.x",
			address: "https://10.0.0.5:8200",
			wantErr: false,
		},
		{
			name:    "valid private IP 172.16.x.x",
			address: "https://172.16.0.10:8200",
			wantErr: false,
		},
		{
			name:    "valid private IP 192.168.x.x",
			address: "https://192.168.1.100:8200",
			wantErr: false,
		},

		// Invalid addresses - SSRF protection
		{
			name:    "block link-local address 169.254.x.x",
			address: "http://169.254.169.254/latest/meta-data",
			wantErr: true,
			errMsg:  "link-local address",
		},
		{
			name:    "block unspecified address 0.0.0.0",
			address: "http://0.0.0.0:8200",
			wantErr: true,
			errMsg:  "unspecified address",
		},
		{
			name:    "block metadata.google.internal",
			address: "http://metadata.google.internal/computeMetadata/v1/",
			wantErr: true,
			errMsg:  "hostname is not allowed",
		},
		{
			name:    "block metadata hostname",
			address: "http://metadata/computeMetadata/v1/",
			wantErr: true,
			errMsg:  "hostname is not allowed",
		},
		{
			name:    "block instance-data hostname",
			address: "http://instance-data/latest/meta-data",
			wantErr: true,
			errMsg:  "hostname is not allowed",
		},

		// Invalid addresses - other validation
		{
			name:    "empty address",
			address: "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "invalid scheme ftp",
			address: "ftp://vault.example.com",
			wantErr: true,
			errMsg:  "must use http or https",
		},
		{
			name:    "invalid scheme file",
			address: "file:///etc/passwd",
			wantErr: true,
			errMsg:  "must use http or https",
		},
		{
			name:    "missing hostname",
			address: "https:///path",
			wantErr: true,
			errMsg:  "must include a hostname",
		},
		{
			name:    "invalid port",
			address: "https://vault.example.com:99999",
			wantErr: true,
			errMsg:  "invalid vault port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVaultAddress(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVaultAddress(%q) error = %v, wantErr %v", tt.address, err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateVaultAddress(%q) error = %v, want error containing %q", tt.address, err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateVaultSecretPath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		// Valid paths
		{
			name:    "simple path",
			path:    "web-cli/ssh-keys",
			wantErr: false,
		},
		{
			name:    "path with underscores",
			path:    "web_cli/ssh_keys",
			wantErr: false,
		},
		{
			name:    "path with dashes",
			path:    "web-cli/ssh-keys/my-key",
			wantErr: false,
		},
		{
			name:    "path with dots",
			path:    "web-cli/config.json",
			wantErr: false,
		},

		// Invalid paths - path traversal
		{
			name:    "path traversal with ..",
			path:    "../../../etc/passwd",
			wantErr: true,
			errMsg:  "path traversal",
		},
		{
			name:    "path traversal in middle",
			path:    "web-cli/../../../etc/passwd",
			wantErr: true,
			errMsg:  "path traversal",
		},
		{
			name:    "absolute path",
			path:    "/etc/passwd",
			wantErr: true,
			errMsg:  "cannot start with /",
		},
		{
			name:    "backslash path",
			path:    "web-cli\\ssh-keys",
			wantErr: true,
			errMsg:  "cannot contain backslashes",
		},
		{
			name:    "null byte",
			path:    "web-cli\x00/ssh-keys",
			wantErr: true,
			errMsg:  "null characters",
		},
		{
			name:    "URL encoded",
			path:    "web-cli%2Fssh-keys",
			wantErr: true,
			errMsg:  "URL-encoded",
		},
		{
			name:    "consecutive slashes",
			path:    "web-cli//ssh-keys",
			wantErr: true,
			errMsg:  "consecutive slashes",
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVaultSecretPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVaultSecretPath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateVaultSecretPath(%q) error = %v, want error containing %q", tt.path, err, tt.errMsg)
				}
			}
		})
	}
}

func TestValidateVaultSecretName(t *testing.T) {
	tests := []struct {
		name       string
		secretName string
		wantErr    bool
		errMsg     string
	}{
		// Valid names
		{
			name:       "simple name",
			secretName: "my-secret",
			wantErr:    false,
		},
		{
			name:       "name with underscores",
			secretName: "my_secret_key",
			wantErr:    false,
		},
		{
			name:       "name with dots",
			secretName: "config.json",
			wantErr:    false,
		},

		// Invalid names
		{
			name:       "path separator forward slash",
			secretName: "path/to/secret",
			wantErr:    true,
			errMsg:     "path separators",
		},
		{
			name:       "path separator backslash",
			secretName: "path\\to\\secret",
			wantErr:    true,
			errMsg:     "path separators",
		},
		{
			name:       "path traversal",
			secretName: "..",
			wantErr:    true,
			errMsg:     "path traversal",
		},
		{
			name:       "empty name",
			secretName: "",
			wantErr:    true,
			errMsg:     "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVaultSecretName(tt.secretName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVaultSecretName(%q) error = %v, wantErr %v", tt.secretName, err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && err != nil {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateVaultSecretName(%q) error = %v, want error containing %q", tt.secretName, err, tt.errMsg)
				}
			}
		})
	}
}

// contains checks if substr is contained in s
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstr(s, substr)))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
