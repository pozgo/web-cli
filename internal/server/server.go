package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pozgo/web-cli/internal/config"
	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/middleware"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

// EmbeddedFrontend holds the embedded frontend files
var EmbeddedFrontend embed.FS

// Server represents the HTTP server
type Server struct {
	config *config.Config
	router *mux.Router
	db     *database.DB
}

// New creates a new Server instance
// Returns an error if critical configuration validation fails (e.g., auth misconfigured)
func New(cfg *config.Config, db *database.DB) (*Server, error) {
	// Validate authentication configuration at startup
	authConfig := middleware.LoadAuthConfig()
	if err := authConfig.Validate(); err != nil {
		return nil, err
	}

	s := &Server{
		config: cfg,
		router: mux.NewRouter(),
		db:     db,
	}

	s.setupRoutes()

	return s, nil
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// Load auth configuration
	authConfig := middleware.LoadAuthConfig()

	// Apply authentication middleware to ALL routes (frontend + API)
	s.router.Use(middleware.BasicAuth(authConfig))

	// API routes
	api := s.router.PathPrefix("/api").Subrouter()

	// Health endpoint (authenticated)
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// SSH Keys endpoints
	api.HandleFunc("/keys", s.handleListSSHKeys).Methods("GET")
	api.HandleFunc("/keys", s.handleCreateSSHKey).Methods("POST")
	api.HandleFunc("/keys/groups", s.handleListSSHKeyGroups).Methods("GET")
	api.HandleFunc("/keys/{id}", s.handleGetSSHKey).Methods("GET")
	api.HandleFunc("/keys/{id}", s.handleUpdateSSHKey).Methods("PUT")
	api.HandleFunc("/keys/{id}", s.handleDeleteSSHKey).Methods("DELETE")

	// Servers endpoints
	api.HandleFunc("/servers", s.handleListServers).Methods("GET")
	api.HandleFunc("/servers", s.handleCreateServer).Methods("POST")
	api.HandleFunc("/servers/groups", s.handleListServerGroups).Methods("GET")
	api.HandleFunc("/servers/{id}", s.handleGetServer).Methods("GET")
	api.HandleFunc("/servers/{id}", s.handleUpdateServer).Methods("PUT")
	api.HandleFunc("/servers/{id}", s.handleDeleteServer).Methods("DELETE")

	// Command execution endpoint
	api.HandleFunc("/commands/execute", s.handleExecuteCommand).Methods("POST")

	// Saved commands endpoints
	api.HandleFunc("/saved-commands", s.handleListSavedCommands).Methods("GET")
	api.HandleFunc("/saved-commands", s.handleCreateSavedCommand).Methods("POST")
	api.HandleFunc("/saved-commands/{id}", s.handleGetSavedCommand).Methods("GET")
	api.HandleFunc("/saved-commands/{id}", s.handleUpdateSavedCommand).Methods("PUT")
	api.HandleFunc("/saved-commands/{id}", s.handleDeleteSavedCommand).Methods("DELETE")

	// Command history endpoints
	api.HandleFunc("/history", s.handleListCommandHistory).Methods("GET")
	api.HandleFunc("/history/{id}", s.handleGetCommandHistory).Methods("GET")

	// Local users endpoints
	api.HandleFunc("/local-users", s.handleListLocalUsers).Methods("GET")
	api.HandleFunc("/local-users", s.handleCreateLocalUser).Methods("POST")
	api.HandleFunc("/local-users/{id}", s.handleGetLocalUser).Methods("GET")
	api.HandleFunc("/local-users/{id}", s.handleUpdateLocalUser).Methods("PUT")
	api.HandleFunc("/local-users/{id}", s.handleDeleteLocalUser).Methods("DELETE")

	// System info endpoints
	api.HandleFunc("/system/current-user", s.handleGetCurrentUser).Methods("GET")
	api.HandleFunc("/system/shells", s.handleListAvailableShells).Methods("GET")

	// Environment variables endpoints
	api.HandleFunc("/env-variables", s.handleListEnvVariables).Methods("GET")
	api.HandleFunc("/env-variables", s.handleCreateEnvVariable).Methods("POST")
	api.HandleFunc("/env-variables/groups", s.handleListEnvVariableGroups).Methods("GET")
	api.HandleFunc("/env-variables/{id}", s.handleGetEnvVariable).Methods("GET")
	api.HandleFunc("/env-variables/{id}", s.handleUpdateEnvVariable).Methods("PUT")
	api.HandleFunc("/env-variables/{id}", s.handleDeleteEnvVariable).Methods("DELETE")

	// Bash scripts endpoints
	api.HandleFunc("/bash-scripts", s.handleListBashScripts).Methods("GET")
	api.HandleFunc("/bash-scripts", s.handleCreateBashScript).Methods("POST")
	api.HandleFunc("/bash-scripts/groups", s.handleListBashScriptGroups).Methods("GET")
	api.HandleFunc("/bash-scripts/execute", s.handleExecuteScript).Methods("POST")
	api.HandleFunc("/bash-scripts/execute/stream", s.handleExecuteScriptStream).Methods("POST")
	api.HandleFunc("/bash-scripts/{id}", s.handleGetBashScript).Methods("GET")
	api.HandleFunc("/bash-scripts/{id}", s.handleUpdateBashScript).Methods("PUT")
	api.HandleFunc("/bash-scripts/{id}", s.handleDeleteBashScript).Methods("DELETE")
	api.HandleFunc("/bash-scripts/{id}/presets", s.handleGetScriptPresetsByScript).Methods("GET")

		// Script preset endpoints
	api.HandleFunc("/script-presets", s.handleListScriptPresets).Methods("GET")
	api.HandleFunc("/script-presets", s.handleCreateScriptPreset).Methods("POST")
	api.HandleFunc("/script-presets/{id}", s.handleGetScriptPreset).Methods("GET")
	api.HandleFunc("/script-presets/{id}", s.handleUpdateScriptPreset).Methods("PUT")
	api.HandleFunc("/script-presets/{id}", s.handleDeleteScriptPreset).Methods("DELETE")

	// Vault integration endpoints
	api.HandleFunc("/vault/config", s.handleGetVaultConfig).Methods("GET")
	api.HandleFunc("/vault/config", s.handleCreateOrUpdateVaultConfig).Methods("POST")
	api.HandleFunc("/vault/config", s.handleDeleteVaultConfig).Methods("DELETE")
	api.HandleFunc("/vault/test", s.handleTestVaultConnection).Methods("POST")
	api.HandleFunc("/vault/status", s.handleGetVaultStatus).Methods("GET")
	api.HandleFunc("/vault/ssh-keys", s.handleListVaultSSHKeys).Methods("GET")
	api.HandleFunc("/vault/ssh-keys", s.handleCreateVaultSSHKey).Methods("POST")
	api.HandleFunc("/vault/servers", s.handleListVaultServers).Methods("GET")
	api.HandleFunc("/vault/servers", s.handleCreateVaultServer).Methods("POST")
	api.HandleFunc("/vault/env-variables", s.handleListVaultEnvVariables).Methods("GET")
	api.HandleFunc("/vault/env-variables", s.handleCreateVaultEnvVariable).Methods("POST")
	api.HandleFunc("/vault/bash-scripts", s.handleListVaultScripts).Methods("GET")
	api.HandleFunc("/vault/bash-scripts", s.handleCreateVaultScript).Methods("POST")
	api.HandleFunc("/vault/scripts", s.handleListVaultScripts).Methods("GET") // Backward compatibility

	// Terminal WebSocket endpoint (for interactive shell)
	api.HandleFunc("/terminal/ws", s.handleTerminalWebSocket)

	// Swagger documentation endpoint (with redirect from /swagger to /swagger/index.html)
	s.router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	// Log auth status
	if authConfig.Enabled {
		log.Println("Authentication is ENABLED for entire application (frontend + API)")
	} else {
		log.Println("WARNING: Authentication is DISABLED (set AUTH_ENABLED=true for production)")
	}

	// Serve static files from frontend build
	s.serveFrontend()
}

// handleHealth returns the health status of the server
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// serveFrontend serves the React frontend
func (s *Server) serveFrontend() {
	// Try to use filesystem path first (for development)
	if _, err := os.Stat(s.config.FrontendPath); err == nil {
		log.Printf("Serving frontend from filesystem: %s", s.config.FrontendPath)
		s.serveFrontendFromFilesystem()
		return
	}

	// Fall back to embedded frontend (for production binaries)
	log.Println("Serving frontend from embedded files")
	s.serveFrontendFromEmbedded()
}

// serveFrontendFromFilesystem serves frontend from filesystem
func (s *Server) serveFrontendFromFilesystem() {
	staticFS := http.FileServer(http.Dir(s.config.FrontendPath))

	s.router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(s.config.FrontendPath, r.URL.Path)

		if _, err := os.Stat(path); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(s.config.FrontendPath, "index.html"))
			return
		}

		staticFS.ServeHTTP(w, r)
	})
}

// serveFrontendFromEmbedded serves frontend from embedded files
func (s *Server) serveFrontendFromEmbedded() {
	// Get the build subdirectory from embedded FS
	buildFS, err := fs.Sub(EmbeddedFrontend, "frontend")
	if err != nil {
		log.Printf("Warning: Could not access embedded frontend: %v", err)
		s.serveErrorPage()
		return
	}

	staticFS := http.FileServer(http.FS(buildFS))

	s.router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to read the requested file from embedded FS
		if _, err := fs.Stat(buildFS, filepath.Join(".", r.URL.Path)); err != nil {
			// File doesn't exist, serve index.html for SPA routing
			indexContent, err := fs.ReadFile(buildFS, "index.html")
			if err != nil {
				http.Error(w, "Frontend not available", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write(indexContent)
			return
		}

		staticFS.ServeHTTP(w, r)
	})
}

// serveErrorPage serves an error page when frontend is not available
func (s *Server) serveErrorPage() {
	s.router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>Web CLI</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 600px; margin: 0 auto; background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; }
        p { color: #666; line-height: 1.6; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Web CLI Server</h1>
        <p>Server is running, but frontend is not available.</p>
        <p>API endpoints are available at <code>/api/*</code></p>
    </div>
</body>
</html>
		`))
	})
}

// Start begins listening for HTTP requests
func (s *Server) Start() error {
	// Get allowed origins from environment, default to localhost only
	allowedOrigins := []string{"http://localhost:7777", "http://127.0.0.1:7777"}
	if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
		// Split by comma for multiple origins
		allowedOrigins = strings.Split(envOrigins, ",")
		for i := range allowedOrigins {
			allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
		}
	}

	// Setup CORS with restrictive defaults
	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	// Apply security headers middleware
	securedHandler := middleware.SecureHeaders()(c.Handler(s.router))

	// Load auth config for HTTPS enforcement check
	authConfig := middleware.LoadAuthConfig()

	// Apply HTTPS enforcement middleware if configured
	securityConfig := &middleware.SecurityConfig{
		RequireHTTPS: s.config.RequireHTTPS,
		AuthEnabled:  authConfig.Enabled,
	}
	handler := middleware.RequireHTTPS(securityConfig)(securedHandler)

	addr := s.config.GetAddress()
	log.Printf("Starting server on %s", addr)
	log.Printf("Frontend path: %s", s.config.FrontendPath)
	log.Printf("Database path: %s", s.config.DatabasePath)
	log.Printf("CORS allowed origins: %v", allowedOrigins)

	// Create HTTP server with proper timeouts
	// WriteTimeout is set high to support long-running script streaming (SSE)
	// Individual handlers can implement their own timeouts via context
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  s.config.GetReadTimeout(),
		WriteTimeout: s.config.GetWriteTimeout(),
		IdleTimeout:  s.config.GetIdleTimeout(),
	}

	// Start with TLS if configured
	if s.config.TLSEnabled() {
		log.Printf("TLS enabled - using certificate: %s", s.config.TLSCertPath)
		if s.config.RequireHTTPS && authConfig.Enabled {
			log.Println("HTTPS enforcement is ENABLED (non-HTTPS requests will be rejected)")
		}
		return server.ListenAndServeTLS(s.config.TLSCertPath, s.config.TLSKeyPath)
	}

	// Warn if auth is enabled without HTTPS
	if authConfig.Enabled && !s.config.TLSEnabled() {
		log.Println("WARNING: Authentication is enabled but TLS is not configured!")
		log.Println("WARNING: Credentials will be transmitted in plain text!")
		log.Println("WARNING: Set TLS_CERT_PATH and TLS_KEY_PATH environment variables for production use.")
	}

	return server.ListenAndServe()
}
