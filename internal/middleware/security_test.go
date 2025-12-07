package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireHTTPS_Disabled(t *testing.T) {
	config := &SecurityConfig{
		RequireHTTPS: false,
		AuthEnabled:  true,
	}

	handler := RequireHTTPS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status OK when RequireHTTPS is disabled, got %d", rec.Code)
	}
}

func TestRequireHTTPS_AuthDisabled(t *testing.T) {
	config := &SecurityConfig{
		RequireHTTPS: true,
		AuthEnabled:  false,
	}

	handler := RequireHTTPS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status OK when auth is disabled, got %d", rec.Code)
	}
}

func TestRequireHTTPS_RejectsHTTP(t *testing.T) {
	config := &SecurityConfig{
		RequireHTTPS: true,
		AuthEnabled:  true,
	}

	handler := RequireHTTPS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status Forbidden for HTTP request when HTTPS required, got %d", rec.Code)
	}
}

func TestRequireHTTPS_AllowsXForwardedProto(t *testing.T) {
	config := &SecurityConfig{
		RequireHTTPS: true,
		AuthEnabled:  true,
	}

	handler := RequireHTTPS(config)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status OK with X-Forwarded-Proto: https, got %d", rec.Code)
	}
}

func TestSecureHeaders(t *testing.T) {
	handler := SecureHeaders()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Check security headers are set
	tests := []struct {
		header   string
		expected string
	}{
		{"X-Frame-Options", "DENY"},
		{"X-Content-Type-Options", "nosniff"},
		{"X-XSS-Protection", "1; mode=block"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
	}

	for _, tt := range tests {
		got := rec.Header().Get(tt.header)
		if got != tt.expected {
			t.Errorf("Expected %s: %s, got: %s", tt.header, tt.expected, got)
		}
	}
}
