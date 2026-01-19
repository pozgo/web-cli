package server

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

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
	sshKeyID := r.URL.Query().Get("sshKeyId")
	sshKeySource := r.URL.Query().Get("sshKeySource")

	if sshKeyID != "" {
		if sshKeySource == "vault" {
			// Fetch SSH key from Vault by name
			client, err := s.getVaultClient()
			if err == nil {
				ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
				defer cancel()
				// Try to get key from default group first, then try without group
				key, err := client.GetSSHKey(ctx, "default", sshKeyID)
				if err != nil || key == nil {
					// Try listing all groups and searching
					groups, _ := client.ListSSHKeyGroups(ctx)
					for _, group := range groups {
						key, err = client.GetSSHKey(ctx, group, sshKeyID)
						if err == nil && key != nil {
							break
						}
					}
				}
				if key != nil {
					sshPrivateKey = key.PrivateKey
					log.Printf("Loaded SSH key '%s' from Vault", sshKeyID)
				} else {
					log.Printf("Failed to find SSH key '%s' in Vault", sshKeyID)
				}
			} else {
				log.Printf("Failed to get Vault client for SSH key: %v", err)
			}
		} else {
			// Fetch SSH key from local database by ID
			keyID, err := strconv.ParseInt(sshKeyID, 10, 64)
			if err == nil {
				repo := repository.NewSSHKeyRepository(s.db)
				key, err := repo.GetByID(keyID)
				if err == nil {
					sshPrivateKey = key.PrivateKey
					log.Printf("Loaded SSH key ID %d from local database", keyID)
				}
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
