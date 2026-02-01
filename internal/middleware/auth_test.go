package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestBasicAuth_Disabled(t *testing.T) {
	config := &AuthConfig{
		Enabled: false,
	}

	handler := BasicAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestBasicAuth_WithValidCredentials(t *testing.T) {
	config := &AuthConfig{
		Enabled:  true,
		Username: "admin",
		Password: "secret",
	}

	handler := BasicAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.SetBasicAuth("admin", "secret")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestBasicAuth_WithInvalidCredentials(t *testing.T) {
	config := &AuthConfig{
		Enabled:  true,
		Username: "admin",
		Password: "secret",
	}

	handler := BasicAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.SetBasicAuth("admin", "wrong")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestBasicAuth_WithNoCredentials(t *testing.T) {
	config := &AuthConfig{
		Enabled:  true,
		Username: "admin",
		Password: "secret",
	}

	handler := BasicAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestBasicAuth_WithValidAPIToken(t *testing.T) {
	config := &AuthConfig{
		Enabled:  true,
		Username: "admin",
		Password: "secret",
		APIToken: "test-token-12345",
	}

	handler := BasicAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer test-token-12345")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestBasicAuth_WithInvalidAPIToken(t *testing.T) {
	config := &AuthConfig{
		Enabled:  true,
		Username: "admin",
		Password: "secret",
		APIToken: "test-token-12345",
	}

	handler := BasicAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestLoadAuthConfig(t *testing.T) {
	// Save original env vars
	originalEnabled := os.Getenv("AUTH_ENABLED")
	originalUsername := os.Getenv("AUTH_USERNAME")
	originalPassword := os.Getenv("AUTH_PASSWORD")
	originalToken := os.Getenv("AUTH_API_TOKEN")

	defer func() {
		os.Setenv("AUTH_ENABLED", originalEnabled)
		os.Setenv("AUTH_USERNAME", originalUsername)
		os.Setenv("AUTH_PASSWORD", originalPassword)
		os.Setenv("AUTH_API_TOKEN", originalToken)
	}()

	// Test default (disabled)
	os.Setenv("AUTH_ENABLED", "")
	config := LoadAuthConfig()
	if config.Enabled {
		t.Error("Auth should be disabled by default")
	}

	// Test enabled
	os.Setenv("AUTH_ENABLED", "true")
	os.Setenv("AUTH_USERNAME", "testuser")
	os.Setenv("AUTH_PASSWORD", "testpass")
	os.Setenv("AUTH_API_TOKEN", "testtoken")

	config = LoadAuthConfig()
	if !config.Enabled {
		t.Error("Auth should be enabled")
	}
	if config.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", config.Username)
	}
	if config.Password != "testpass" {
		t.Errorf("Expected password 'testpass', got '%s'", config.Password)
	}
	if config.APIToken != "testtoken" {
		t.Errorf("Expected token 'testtoken', got '%s'", config.APIToken)
	}
}

func TestBasicAuth_ConstantTimeComparison(t *testing.T) {
	// This test ensures credentials are compared in constant time
	// We can't directly test timing, but we can verify functionality
	config := &AuthConfig{
		Enabled:  true,
		Username: "admin",
		Password: "verylongpasswordthatshouldbeconstanttime",
	}

	handler := BasicAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test with correct password
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.SetBasicAuth("admin", "verylongpasswordthatshouldbeconstanttime")
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Error("Valid credentials should authenticate")
	}

	// Test with almost correct password (different length)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.SetBasicAuth("admin", "verylongpasswordthatshouldbeconstanttim")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Error("Invalid credentials should not authenticate")
	}
}

func TestBasicAuth_ExcludedPaths(t *testing.T) {
	config := &AuthConfig{
		Enabled:      true,
		Username:     "admin",
		Password:     "secret",
		ExcludePaths: []string{"/api/health"},
	}

	handler := BasicAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	tests := []struct {
		name           string
		path           string
		withAuth       bool
		expectedStatus int
	}{
		{
			name:           "excluded path without auth",
			path:           "/api/health",
			withAuth:       false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "excluded path with auth",
			path:           "/api/health",
			withAuth:       true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-excluded path without auth",
			path:           "/api/keys",
			withAuth:       false,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "non-excluded path with auth",
			path:           "/api/keys",
			withAuth:       true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.withAuth {
				req.SetBasicAuth("admin", "secret")
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", tt.expectedStatus, tt.path, w.Code)
			}
		})
	}
}

func TestBasicAuth_ExcludedPathsEmpty(t *testing.T) {
	// Verify that with no excluded paths, all paths require auth
	config := &AuthConfig{
		Enabled:      true,
		Username:     "admin",
		Password:     "secret",
		ExcludePaths: nil,
	}

	handler := BasicAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d when ExcludePaths is nil, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestBasicAuth_ExcludedPathsMultiple(t *testing.T) {
	config := &AuthConfig{
		Enabled:      true,
		Username:     "admin",
		Password:     "secret",
		ExcludePaths: []string{"/api/health", "/api/readyz"},
	}

	handler := BasicAuth(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		path           string
		expectedStatus int
	}{
		{"/api/health", http.StatusOK},
		{"/api/readyz", http.StatusOK},
		{"/api/keys", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", tt.expectedStatus, tt.path, w.Code)
			}
		})
	}
}
