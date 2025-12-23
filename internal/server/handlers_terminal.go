package server

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/pozgo/web-cli/internal/repository"
	"github.com/pozgo/web-cli/internal/terminal"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (authentication is handled by middleware)
		return true
	},
}

// handleTerminalWebSocket handles WebSocket connections for interactive terminal sessions
func (s *Server) handleTerminalWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Determine which shell to use
	shell := "/bin/bash"
	if queryShell := r.URL.Query().Get("shell"); queryShell != "" {
		// Only allow specific shells for security
		switch queryShell {
		case "bash", "/bin/bash":
			shell = "/bin/bash"
		case "sh", "/bin/sh":
			shell = "/bin/sh"
		case "zsh", "/bin/zsh":
			shell = "/bin/zsh"
		default:
			log.Printf("Invalid shell requested: %s, using default", queryShell)
		}
	}

	// Check if SSH key is requested
	var sshPrivateKey string
	if sshKeyIDStr := r.URL.Query().Get("sshKeyId"); sshKeyIDStr != "" {
		sshKeyID, err := strconv.ParseInt(sshKeyIDStr, 10, 64)
		if err == nil {
			repo := repository.NewSSHKeyRepository(s.db)
			key, err := repo.GetByID(sshKeyID)
			if err == nil {
				sshPrivateKey = key.PrivateKey
			}
		}
	}

	// Fetch all servers from admin panel for SSH config generation
	var servers []terminal.ServerConfig
	serverRepo := repository.NewServerRepository(s.db)
	serverList, err := serverRepo.GetAll()
	if err == nil {
		for _, srv := range serverList {
			servers = append(servers, terminal.ServerConfig{
				Name:      srv.Name,
				IPAddress: srv.IPAddress,
				Port:      srv.Port,
				Username:  srv.Username,
			})
		}
	}

	// Create new terminal session with optional SSH key and server configs
	session, err := terminal.NewSession(ws, shell, sshPrivateKey, servers)
	if err != nil {
		log.Printf("Failed to create terminal session: %v", err)
		ws.WriteMessage(websocket.TextMessage, []byte("Failed to create terminal session: "+err.Error()))
		ws.Close()
		return
	}

	log.Printf("Terminal session started with shell: %s", shell)

	// Start the session (blocks until session ends)
	session.Start()

	log.Printf("Terminal session ended")
}
