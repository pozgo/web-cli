package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Enabled  bool
	Username string
	Password string
	APIToken string
}

// LoadAuthConfig loads authentication configuration from environment
func LoadAuthConfig() *AuthConfig {
	// Auth is disabled by default for development
	// Set AUTH_ENABLED=true in production
	enabled := os.Getenv("AUTH_ENABLED") == "true"

	return &AuthConfig{
		Enabled:  enabled,
		Username: os.Getenv("AUTH_USERNAME"),
		Password: os.Getenv("AUTH_PASSWORD"),
		APIToken: os.Getenv("AUTH_API_TOKEN"),
	}
}

// BasicAuth provides HTTP Basic Authentication middleware
func BasicAuth(config *AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth if disabled
			if !config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Check for API token first (Bearer token in Authorization header)
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if config.APIToken != "" && subtle.ConstantTimeCompare([]byte(token), []byte(config.APIToken)) == 1 {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Fall back to Basic Auth
			if config.Username != "" && config.Password != "" {
				username, password, ok := r.BasicAuth()
				if !ok {
					requireAuth(w)
					return
				}

				// Use constant time comparison to prevent timing attacks
				usernameMatch := subtle.ConstantTimeCompare([]byte(username), []byte(config.Username)) == 1
				passwordMatch := subtle.ConstantTimeCompare([]byte(password), []byte(config.Password)) == 1

				if !usernameMatch || !passwordMatch {
					requireAuth(w)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			// If auth is enabled but no credentials configured, deny access
			http.Error(w, "Authentication required but not configured", http.StatusInternalServerError)
		})
	}
}

// requireAuth sends a 401 response requesting authentication
func requireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Web CLI"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
