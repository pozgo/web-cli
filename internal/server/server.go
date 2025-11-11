package server

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/pozgo/web-cli/internal/config"
	"github.com/pozgo/web-cli/internal/database"
	"github.com/rs/cors"
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
func New(cfg *config.Config, db *database.DB) *Server {
	s := &Server{
		config: cfg,
		router: mux.NewRouter(),
		db:     db,
	}

	s.setupRoutes()

	return s
}

// setupRoutes configures all HTTP routes
func (s *Server) setupRoutes() {
	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", s.handleHealth).Methods("GET")

	// SSH Keys endpoints
	api.HandleFunc("/keys", s.handleListSSHKeys).Methods("GET")
	api.HandleFunc("/keys", s.handleCreateSSHKey).Methods("POST")
	api.HandleFunc("/keys/{id}", s.handleGetSSHKey).Methods("GET")
	api.HandleFunc("/keys/{id}", s.handleUpdateSSHKey).Methods("PUT")
	api.HandleFunc("/keys/{id}", s.handleDeleteSSHKey).Methods("DELETE")

	// Servers endpoints
	api.HandleFunc("/servers", s.handleListServers).Methods("GET")
	api.HandleFunc("/servers", s.handleCreateServer).Methods("POST")
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
	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	handler := c.Handler(s.router)

	addr := s.config.GetAddress()
	log.Printf("Starting server on %s", addr)
	log.Printf("Frontend path: %s", s.config.FrontendPath)
	log.Printf("Database path: %s", s.config.DatabasePath)

	return http.ListenAndServe(addr, handler)
}
