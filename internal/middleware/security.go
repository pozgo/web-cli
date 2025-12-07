package middleware

import (
	"net/http"
)

// SecurityConfig holds security middleware configuration
type SecurityConfig struct {
	RequireHTTPS bool // If true, reject non-HTTPS requests when auth is enabled
	AuthEnabled  bool // Whether authentication is enabled
}

// RequireHTTPS middleware rejects non-HTTPS requests when configured
// This is critical for protecting credentials when basic auth is used
func RequireHTTPS(config *SecurityConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip check if HTTPS requirement is disabled or auth is disabled
			if !config.RequireHTTPS || !config.AuthEnabled {
				next.ServeHTTP(w, r)
				return
			}

			// Check if request is over HTTPS
			// Check TLS directly, or X-Forwarded-Proto header (for reverse proxy setups)
			isHTTPS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"

			if !isHTTPS {
				http.Error(w, "HTTPS required. This endpoint requires a secure connection.", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecureHeaders adds security-related HTTP headers to responses
func SecureHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent clickjacking
			w.Header().Set("X-Frame-Options", "DENY")
			// Prevent MIME type sniffing
			w.Header().Set("X-Content-Type-Options", "nosniff")
			// Enable XSS protection
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			// Referrer policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			next.ServeHTTP(w, r)
		})
	}
}
