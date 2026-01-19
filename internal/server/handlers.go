package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pozgo/web-cli/internal/audit"
	"github.com/pozgo/web-cli/internal/executor"
	"github.com/pozgo/web-cli/internal/models"
	"github.com/pozgo/web-cli/internal/repository"
	"github.com/pozgo/web-cli/internal/validation"
)

// ErrorResponse represents an error response
// @Description Error response returned by the API
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request body"`
}

// HealthResponse represents the health check response
// @Description Health check response
type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// CurrentUserResponse represents the current user response
// @Description Current system user information
type CurrentUserResponse struct {
	Username string `json:"username" example:"root"`
	UID      string `json:"uid" example:"0"`
	GID      string `json:"gid" example:"0"`
	Name     string `json:"name" example:"root"`
	HomeDir  string `json:"home_dir" example:"/root"`
}

// handleListSSHKeys godoc
// @Summary List all SSH keys
// @Description Get a list of all SSH keys stored in the system
// @Tags SSH Keys
// @Accept json
// @Produce json
// @Success 200 {array} models.SSHKey
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Param group query string false "Filter by group name"
// @Router /keys [get]
func (s *Server) handleListSSHKeys(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewSSHKeyRepository(s.db)
	group := r.URL.Query().Get("group")

	var keys []*models.SSHKey
	var err error

	if group != "" {
		keys, err = repo.GetByGroup(group)
	} else {
		keys, err = repo.GetAll()
	}
	if err != nil {
		log.Printf("Error fetching SSH keys: %v", err)
		http.Error(w, "Failed to fetch SSH keys", http.StatusInternalServerError)
		return
	}

	// Merge with Vault keys if available
	allKeys := s.mergeSSHKeysWithVault(r.Context(), keys)

	// Filter Vault keys by group if specified
	if group != "" {
		filtered := make([]*models.SSHKey, 0)
		for _, k := range allKeys {
			if k.Group == group || (k.Group == "" && group == "default") {
				filtered = append(filtered, k)
			}
		}
		allKeys = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allKeys)
}

// handleCreateSSHKey godoc
// @Summary Create a new SSH key
// @Description Store a new SSH private key in the system
// @Tags SSH Keys
// @Accept json
// @Produce json
// @Param key body models.SSHKeyCreate true "SSH Key to create"
// @Success 201 {object} models.SSHKey
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /keys [post]
func (s *Server) handleCreateSSHKey(w http.ResponseWriter, r *http.Request) {
	var keyCreate models.SSHKeyCreate

	if err := json.NewDecoder(r.Body).Decode(&keyCreate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if err := validation.ValidateCommandName(keyCreate.Name); err != nil {
		http.Error(w, fmt.Sprintf("Invalid name: %v", err), http.StatusBadRequest)
		return
	}

	if err := validation.ValidateSSHPrivateKey(keyCreate.PrivateKey); err != nil {
		http.Error(w, fmt.Sprintf("Invalid SSH private key: %v", err), http.StatusBadRequest)
		return
	}

	repo := repository.NewSSHKeyRepository(s.db)

	key, err := repo.Create(&keyCreate)
	if err != nil {
		log.Printf("Error creating SSH key: %v", err)
		http.Error(w, "Failed to create SSH key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(key)
}

// handleGetSSHKey godoc
// @Summary Get an SSH key by ID
// @Description Get a single SSH key by its ID
// @Tags SSH Keys
// @Accept json
// @Produce json
// @Param id path int true "SSH Key ID"
// @Success 200 {object} models.SSHKey
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BasicAuth
// @Router /keys/{id} [get]
func (s *Server) handleGetSSHKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid key ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewSSHKeyRepository(s.db)

	key, err := repo.GetByID(id)
	if err != nil {
		log.Printf("Error fetching SSH key: %v", err)
		http.Error(w, "SSH key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(key)
}

// handleUpdateSSHKey godoc
// @Summary Update an SSH key
// @Description Update an existing SSH key by its ID
// @Tags SSH Keys
// @Accept json
// @Produce json
// @Param id path int true "SSH Key ID"
// @Param key body models.SSHKeyUpdate true "SSH Key update data"
// @Success 200 {object} models.SSHKey
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /keys/{id} [put]
func (s *Server) handleUpdateSSHKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid key ID", http.StatusBadRequest)
		return
	}

	var keyUpdate models.SSHKeyUpdate

	if err := json.NewDecoder(r.Body).Decode(&keyUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input - at least one field must be provided
	if keyUpdate.Name == "" && keyUpdate.PrivateKey == "" {
		http.Error(w, "At least one field (name or private_key) must be provided", http.StatusBadRequest)
		return
	}

	repo := repository.NewSSHKeyRepository(s.db)

	key, err := repo.Update(id, &keyUpdate)
	if err != nil {
		log.Printf("Error updating SSH key: %v", err)
		http.Error(w, "Failed to update SSH key", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(key)
}

// handleDeleteSSHKey godoc
// @Summary Delete an SSH key
// @Description Delete an SSH key by its ID
// @Tags SSH Keys
// @Accept json
// @Produce json
// @Param id path int true "SSH Key ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /keys/{id} [delete]
func (s *Server) handleDeleteSSHKey(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid key ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewSSHKeyRepository(s.db)

	if err := repo.Delete(id); err != nil {
		log.Printf("Error deleting SSH key: %v", err)
		http.Error(w, "Failed to delete SSH key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleListServers godoc
// @Summary List all servers
// @Description Get a list of all remote servers configured in the system
// @Tags Servers
// @Accept json
// @Produce json
// @Success 200 {array} models.Server
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Param group query string false "Filter by group name"
// @Router /servers [get]
func (s *Server) handleListServers(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewServerRepository(s.db)
	group := r.URL.Query().Get("group")

	var servers []*models.Server
	var err error

	if group != "" {
		servers, err = repo.GetByGroup(group)
	} else {
		servers, err = repo.GetAll()
	}
	if err != nil {
		log.Printf("Error fetching servers: %v", err)
		http.Error(w, "Failed to fetch servers", http.StatusInternalServerError)
		return
	}

	// Merge with Vault servers if available
	allServers := s.mergeServersWithVault(r.Context(), servers)

	// Filter Vault servers by group if specified
	if group != "" {
		filtered := make([]*models.Server, 0)
		for _, srv := range allServers {
			if srv.Group == group || (srv.Group == "" && group == "default") {
				filtered = append(filtered, srv)
			}
		}
		allServers = filtered
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allServers)
}

// handleCreateServer godoc
// @Summary Create a new server
// @Description Add a new remote server configuration
// @Tags Servers
// @Accept json
// @Produce json
// @Param server body models.ServerCreate true "Server to create"
// @Success 201 {object} models.Server
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /servers [post]
func (s *Server) handleCreateServer(w http.ResponseWriter, r *http.Request) {
	var serverCreate models.ServerCreate

	if err := json.NewDecoder(r.Body).Decode(&serverCreate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input - at least one field must be provided
	if serverCreate.Name == "" && serverCreate.IPAddress == "" {
		http.Error(w, "At least one of name or ip_address must be provided", http.StatusBadRequest)
		return
	}

	// Validate hostname if provided
	if serverCreate.Name != "" {
		if err := validation.ValidateHostname(serverCreate.Name); err != nil {
			http.Error(w, fmt.Sprintf("Invalid hostname: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Validate IP address if provided
	if serverCreate.IPAddress != "" {
		if err := validation.ValidateIPOrHostname(serverCreate.IPAddress); err != nil {
			http.Error(w, fmt.Sprintf("Invalid IP address or hostname: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Validate port if provided
	if serverCreate.Port > 0 {
		if err := validation.ValidatePort(serverCreate.Port); err != nil {
			http.Error(w, fmt.Sprintf("Invalid port: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Validate username if provided
	if serverCreate.Username != "" {
		if err := validation.ValidateUsername(serverCreate.Username); err != nil {
			http.Error(w, fmt.Sprintf("Invalid username: %v", err), http.StatusBadRequest)
			return
		}
	}

	repo := repository.NewServerRepository(s.db)

	server, err := repo.Create(&serverCreate)
	if err != nil {
		log.Printf("Error creating server: %v", err)
		http.Error(w, "Failed to create server", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(server)
}

// handleGetServer godoc
// @Summary Get a server by ID
// @Description Get a single server configuration by its ID
// @Tags Servers
// @Accept json
// @Produce json
// @Param id path int true "Server ID"
// @Success 200 {object} models.Server
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BasicAuth
// @Router /servers/{id} [get]
func (s *Server) handleGetServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewServerRepository(s.db)

	server, err := repo.GetByID(id)
	if err != nil {
		log.Printf("Error fetching server: %v", err)
		http.Error(w, "Server not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(server)
}

// handleUpdateServer godoc
// @Summary Update a server
// @Description Update an existing server configuration by its ID
// @Tags Servers
// @Accept json
// @Produce json
// @Param id path int true "Server ID"
// @Param server body models.ServerUpdate true "Server update data"
// @Success 200 {object} models.Server
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /servers/{id} [put]
func (s *Server) handleUpdateServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	var serverUpdate models.ServerUpdate

	if err := json.NewDecoder(r.Body).Decode(&serverUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input - at least one field must be provided
	if serverUpdate.Name == "" && serverUpdate.IPAddress == "" {
		http.Error(w, "At least one field (name or ip_address) must be provided", http.StatusBadRequest)
		return
	}

	repo := repository.NewServerRepository(s.db)

	server, err := repo.Update(id, &serverUpdate)
	if err != nil {
		log.Printf("Error updating server: %v", err)
		http.Error(w, "Failed to update server", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(server)
}

// handleDeleteServer godoc
// @Summary Delete a server
// @Description Delete a server configuration by its ID
// @Tags Servers
// @Accept json
// @Produce json
// @Param id path int true "Server ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /servers/{id} [delete]
func (s *Server) handleDeleteServer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid server ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewServerRepository(s.db)

	if err := repo.Delete(id); err != nil {
		log.Printf("Error deleting server: %v", err)
		http.Error(w, "Failed to delete server", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleExecuteCommand godoc
// @Summary Execute a command
// @Description Execute a shell command locally or remotely via SSH
// @Tags Commands
// @Accept json
// @Produce json
// @Param command body models.CommandExecution true "Command execution request"
// @Success 200 {object} models.CommandResult
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /commands/execute [post]
func (s *Server) handleExecuteCommand(w http.ResponseWriter, r *http.Request) {
	var exec models.CommandExecution

	if err := json.NewDecoder(r.Body).Decode(&exec); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate command
	if err := validation.ValidateCommand(exec.Command); err != nil {
		http.Error(w, fmt.Sprintf("Invalid command: %v", err), http.StatusBadRequest)
		return
	}

	// Validate and default user
	if exec.User == "" {
		exec.User = "root"
	} else if err := validation.ValidateUsername(exec.User); err != nil {
		http.Error(w, fmt.Sprintf("Invalid user: %v", err), http.StatusBadRequest)
		return
	}

	var result *executor.ExecuteResult
	serverName := "local"

	if exec.IsRemote {
		// Remote execution via SSH
		var server *models.Server
		var err error

		// Get server details - support both ID (SQLite) and Name (Vault)
		if exec.ServerID != nil && *exec.ServerID > 0 {
			serverRepo := repository.NewServerRepository(s.db)
			server, err = serverRepo.GetByID(*exec.ServerID)
			if err != nil {
				log.Printf("Error fetching server by ID: %v", err)
				http.Error(w, "Server not found", http.StatusNotFound)
				return
			}
		} else if exec.ServerName != "" {
			// Try to find server by name from Vault
			server, err = s.getServerByNameFromVault(r.Context(), exec.ServerGroup, exec.ServerName)
			if err != nil {
				log.Printf("Error fetching server from Vault: %v", err)
				http.Error(w, "Server not found in Vault", http.StatusNotFound)
				return
			}
			if server == nil {
				http.Error(w, "Server not found in Vault", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "Server ID or Server Name is required for remote execution", http.StatusBadRequest)
			return
		}

		// Get SSH key if provided - support both ID (SQLite) and Name (Vault)
		var privateKey string
		if exec.SSHKeyID != nil && *exec.SSHKeyID > 0 {
			keyRepo := repository.NewSSHKeyRepository(s.db)
			key, err := keyRepo.GetByID(*exec.SSHKeyID)
			if err != nil {
				log.Printf("Error fetching SSH key by ID: %v", err)
				http.Error(w, "SSH key not found", http.StatusNotFound)
				return
			}
			privateKey = key.PrivateKey
		} else if exec.SSHKeyName != "" {
			// Try to find SSH key by name from Vault
			key, err := s.getSSHKeyByNameFromVault(r.Context(), exec.SSHKeyGroup, exec.SSHKeyName)
			if err != nil {
				log.Printf("Error fetching SSH key from Vault: %v", err)
				http.Error(w, "SSH key not found in Vault", http.StatusNotFound)
				return
			}
			if key == nil {
				http.Error(w, fmt.Sprintf("SSH key '%s' not found in Vault", exec.SSHKeyName), http.StatusNotFound)
				return
			}
			if key.PrivateKey == "" {
				http.Error(w, fmt.Sprintf("SSH key '%s' has no private key data in Vault", exec.SSHKeyName), http.StatusBadRequest)
				return
			}
			privateKey = key.PrivateKey
		}

		// Set server name for history
		if server.Name != "" {
			serverName = server.Name
		} else if server.IPAddress != "" {
			serverName = server.IPAddress
		}

		// Execute remotely
		remoteExec := executor.NewRemoteExecutorWithHostKeys("", true)
		sshConfig := &executor.SSHConfig{
			Host:       server.IPAddress,
			Port:       server.Port,
			Username:   exec.User,
			PrivateKey: privateKey,
			Password:   exec.SSHPassword, // Fallback to password if key fails
		}
		result = remoteExec.Execute(context.Background(), exec.Command, sshConfig)
	} else {
		// Local execution
		localExec := executor.NewLocalExecutor()
		result = localExec.Execute(context.Background(), exec.Command, exec.User, exec.SudoPassword)
	}

	// Store in command history (NEVER store SSH password)
	exitCode := result.ExitCode
	historyRepo := repository.NewCommandHistoryRepository(s.db)
	_, err := historyRepo.Create(&models.CommandHistoryCreate{
		Command:         exec.Command,
		Output:          result.Output,
		ExitCode:        &exitCode,
		Server:          serverName,
		User:            exec.User,
		ExecutionTimeMs: result.ExecutionTime,
	})
	if err != nil {
		log.Printf("Warning: failed to save command history: %v", err)
		// Don't fail the request, just log the error
	}

	// Audit log the command execution
	audit.GetLogger().LogCommandExecution(r, exec.Command, exec.User, serverName, exitCode, result.ExecutionTime, result.Error)

	// Save as template if requested
	if exec.SaveAs != "" {
		savedCmdRepo := repository.NewSavedCommandRepository(s.db)
		_, err := savedCmdRepo.Create(&models.SavedCommandCreate{
			Name:        exec.SaveAs,
			Command:     exec.Command,
			Description: "Auto-saved from execution",
			User:        exec.User,
			IsRemote:    exec.IsRemote,
			ServerID:    exec.ServerID,
			SSHKeyID:    exec.SSHKeyID,
		})
		if err != nil {
			log.Printf("Warning: failed to save command template: %v", err)
			// Don't fail the request, just log the error
		}
	}

	// Return result - include error in output if present
	output := result.Output
	if result.Error != nil && output == "" {
		output = fmt.Sprintf("Error: %s", result.Error.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.CommandResult{
		Command:       exec.Command,
		Output:        output,
		ExitCode:      result.ExitCode,
		User:          exec.User,
		ExecutionTime: result.ExecutionTime,
		ExecutedAt:    "",
	})
}

// handleListSavedCommands godoc
// @Summary List all saved commands
// @Description Get a list of all saved command templates
// @Tags Saved Commands
// @Accept json
// @Produce json
// @Success 200 {array} models.SavedCommand
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /saved-commands [get]
func (s *Server) handleListSavedCommands(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewSavedCommandRepository(s.db)

	commands, err := repo.GetAll()
	if err != nil {
		log.Printf("Error fetching saved commands: %v", err)
		http.Error(w, "Failed to fetch saved commands", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(commands)
}

// handleCreateSavedCommand godoc
// @Summary Create a saved command
// @Description Create a new saved command template
// @Tags Saved Commands
// @Accept json
// @Produce json
// @Param command body models.SavedCommandCreate true "Saved command to create"
// @Success 201 {object} models.SavedCommand
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /saved-commands [post]
func (s *Server) handleCreateSavedCommand(w http.ResponseWriter, r *http.Request) {
	var cmdCreate models.SavedCommandCreate

	if err := json.NewDecoder(r.Body).Decode(&cmdCreate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if cmdCreate.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if cmdCreate.Command == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}

	repo := repository.NewSavedCommandRepository(s.db)

	cmd, err := repo.Create(&cmdCreate)
	if err != nil {
		log.Printf("Error creating saved command: %v", err)
		http.Error(w, "Failed to create saved command", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cmd)
}

// handleGetSavedCommand godoc
// @Summary Get a saved command by ID
// @Description Get a single saved command template by its ID
// @Tags Saved Commands
// @Accept json
// @Produce json
// @Param id path int true "Saved Command ID"
// @Success 200 {object} models.SavedCommand
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BasicAuth
// @Router /saved-commands/{id} [get]
func (s *Server) handleGetSavedCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid command ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewSavedCommandRepository(s.db)

	cmd, err := repo.GetByID(id)
	if err != nil {
		log.Printf("Error fetching saved command: %v", err)
		http.Error(w, "Saved command not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cmd)
}

// handleUpdateSavedCommand godoc
// @Summary Update a saved command
// @Description Update an existing saved command template by its ID
// @Tags Saved Commands
// @Accept json
// @Produce json
// @Param id path int true "Saved Command ID"
// @Param command body models.SavedCommandUpdate true "Saved command update data"
// @Success 200 {object} models.SavedCommand
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /saved-commands/{id} [put]
func (s *Server) handleUpdateSavedCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid command ID", http.StatusBadRequest)
		return
	}

	var cmdUpdate models.SavedCommandUpdate

	if err := json.NewDecoder(r.Body).Decode(&cmdUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	repo := repository.NewSavedCommandRepository(s.db)

	cmd, err := repo.Update(id, &cmdUpdate)
	if err != nil {
		log.Printf("Error updating saved command: %v", err)
		http.Error(w, "Failed to update saved command", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cmd)
}

// handleDeleteSavedCommand godoc
// @Summary Delete a saved command
// @Description Delete a saved command template by its ID
// @Tags Saved Commands
// @Accept json
// @Produce json
// @Param id path int true "Saved Command ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /saved-commands/{id} [delete]
func (s *Server) handleDeleteSavedCommand(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid command ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewSavedCommandRepository(s.db)

	if err := repo.Delete(id); err != nil {
		log.Printf("Error deleting saved command: %v", err)
		http.Error(w, "Failed to delete saved command", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleListCommandHistory godoc
// @Summary List command history
// @Description Get command execution history with optional filtering
// @Tags Command History
// @Accept json
// @Produce json
// @Param server query string false "Filter by server name"
// @Param limit query int false "Maximum number of records to return" default(100)
// @Success 200 {array} models.CommandHistory
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /history [get]
func (s *Server) handleListCommandHistory(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewCommandHistoryRepository(s.db)

	// Check if filtering by server
	server := r.URL.Query().Get("server")
	limitStr := r.URL.Query().Get("limit")

	limit := 100 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	var history []*models.CommandHistory
	var err error

	if server != "" {
		history, err = repo.GetByServer(server, limit)
	} else {
		history, err = repo.GetAll(limit)
	}

	if err != nil {
		log.Printf("Error fetching command history: %v", err)
		http.Error(w, "Failed to fetch command history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// handleGetCommandHistory godoc
// @Summary Get a command history entry by ID
// @Description Get a single command history entry by its ID
// @Tags Command History
// @Accept json
// @Produce json
// @Param id path int true "Command History ID"
// @Success 200 {object} models.CommandHistory
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BasicAuth
// @Router /history/{id} [get]
func (s *Server) handleGetCommandHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid history ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewCommandHistoryRepository(s.db)

	history, err := repo.GetByID(id)
	if err != nil {
		log.Printf("Error fetching command history: %v", err)
		http.Error(w, "Command history not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// handleListLocalUsers godoc
// @Summary List all local users
// @Description Get a list of all local system users configured for command execution
// @Tags Local Users
// @Accept json
// @Produce json
// @Success 200 {array} models.LocalUser
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /local-users [get]
func (s *Server) handleListLocalUsers(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewLocalUserRepository(s.db)

	users, err := repo.GetAll()
	if err != nil {
		log.Printf("Error fetching local users: %v", err)
		http.Error(w, "Failed to fetch local users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// handleCreateLocalUser godoc
// @Summary Create a local user
// @Description Add a new local system user for command execution
// @Tags Local Users
// @Accept json
// @Produce json
// @Param user body models.LocalUserCreate true "Local user to create"
// @Success 201 {object} models.LocalUser
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /local-users [post]
func (s *Server) handleCreateLocalUser(w http.ResponseWriter, r *http.Request) {
	var userCreate models.LocalUserCreate

	if err := json.NewDecoder(r.Body).Decode(&userCreate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if userCreate.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	repo := repository.NewLocalUserRepository(s.db)

	user, err := repo.Create(&userCreate)
	if err != nil {
		log.Printf("Error creating local user: %v", err)
		http.Error(w, "Failed to create local user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// handleGetLocalUser godoc
// @Summary Get a local user by ID
// @Description Get a single local user by its ID
// @Tags Local Users
// @Accept json
// @Produce json
// @Param id path int true "Local User ID"
// @Success 200 {object} models.LocalUser
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BasicAuth
// @Router /local-users/{id} [get]
func (s *Server) handleGetLocalUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewLocalUserRepository(s.db)

	user, err := repo.GetByID(id)
	if err != nil {
		log.Printf("Error fetching local user: %v", err)
		http.Error(w, "Local user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// handleUpdateLocalUser godoc
// @Summary Update a local user
// @Description Update an existing local user by its ID
// @Tags Local Users
// @Accept json
// @Produce json
// @Param id path int true "Local User ID"
// @Param user body models.LocalUserUpdate true "Local user update data"
// @Success 200 {object} models.LocalUser
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /local-users/{id} [put]
func (s *Server) handleUpdateLocalUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var userUpdate models.LocalUserUpdate

	if err := json.NewDecoder(r.Body).Decode(&userUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	repo := repository.NewLocalUserRepository(s.db)

	user, err := repo.Update(id, &userUpdate)
	if err != nil {
		log.Printf("Error updating local user: %v", err)
		http.Error(w, "Failed to update local user", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// handleDeleteLocalUser godoc
// @Summary Delete a local user
// @Description Delete a local user by its ID
// @Tags Local Users
// @Accept json
// @Produce json
// @Param id path int true "Local User ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /local-users/{id} [delete]
func (s *Server) handleDeleteLocalUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewLocalUserRepository(s.db)

	if err := repo.Delete(id); err != nil {
		log.Printf("Error deleting local user: %v", err)
		http.Error(w, "Failed to delete local user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGetCurrentUser godoc
// @Summary Get current system user
// @Description Get information about the current system user running the server
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} CurrentUserResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /system/current-user [get]
func (s *Server) handleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	currentUser, err := user.Current()
	if err != nil {
		log.Printf("Error getting current user: %v", err)
		http.Error(w, "Failed to get current user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"username": currentUser.Username,
		"uid":      currentUser.Uid,
		"gid":      currentUser.Gid,
		"name":     currentUser.Name,
		"home_dir": currentUser.HomeDir,
	})
}

// ShellInfo represents information about an available shell
// @Description Information about an available shell
type ShellInfo struct {
	Name string `json:"name" example:"bash"`
	Path string `json:"path" example:"/bin/bash"`
}

// handleListAvailableShells godoc
// @Summary List available shells
// @Description Get a list of shells that are installed and available on the system
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {array} ShellInfo
// @Security BasicAuth
// @Router /system/shells [get]
func (s *Server) handleListAvailableShells(w http.ResponseWriter, r *http.Request) {
	// List of common shells to check
	shellsToCheck := []struct {
		name  string
		paths []string
	}{
		{"bash", []string{"/bin/bash", "/usr/bin/bash"}},
		{"sh", []string{"/bin/sh", "/usr/bin/sh"}},
		{"zsh", []string{"/bin/zsh", "/usr/bin/zsh"}},
	}

	var availableShells []ShellInfo

	for _, shell := range shellsToCheck {
		for _, path := range shell.paths {
			if _, err := os.Stat(path); err == nil {
				availableShells = append(availableShells, ShellInfo{
					Name: shell.name,
					Path: path,
				})
				break // Found this shell, move to next
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(availableShells)
}

// handleListEnvVariables godoc
// @Summary List all environment variables
// @Description Get a list of all environment variables (values masked by default)
// @Tags Environment Variables
// @Accept json
// @Produce json
// @Param show_values query bool false "Show actual values instead of masked values"
// @Param group query string false "Filter by group name"
// @Success 200 {array} models.EnvVariableResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /env-variables [get]
func (s *Server) handleListEnvVariables(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewEnvVariableRepository(s.db)
	group := r.URL.Query().Get("group")

	var envVars []*models.EnvVariable
	var err error

	if group != "" {
		envVars, err = repo.GetByGroup(group)
	} else {
		envVars, err = repo.GetAll()
	}
	if err != nil {
		log.Printf("Error fetching environment variables: %v", err)
		http.Error(w, "Failed to fetch environment variables", http.StatusInternalServerError)
		return
	}

	// Merge with Vault env variables if available
	allEnvVars := s.mergeEnvVariablesWithVault(r.Context(), envVars)

	// Filter Vault env vars by group if specified
	if group != "" {
		filtered := make([]*models.EnvVariable, 0)
		for _, ev := range allEnvVars {
			if ev.Group == group || (ev.Group == "" && group == "default") {
				filtered = append(filtered, ev)
			}
		}
		allEnvVars = filtered
	}

	// Check if full values are requested (for internal use)
	showValues := r.URL.Query().Get("show_values") == "true"

	// Convert to response format with masked values
	responses := make([]*models.EnvVariableResponse, len(allEnvVars))
	for i, envVar := range allEnvVars {
		responses[i] = envVar.ToResponse(showValues)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// handleCreateEnvVariable godoc
// @Summary Create an environment variable
// @Description Create a new environment variable (stored encrypted)
// @Tags Environment Variables
// @Accept json
// @Produce json
// @Param envvar body models.EnvVariableCreate true "Environment variable to create"
// @Success 201 {object} models.EnvVariableResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /env-variables [post]
func (s *Server) handleCreateEnvVariable(w http.ResponseWriter, r *http.Request) {
	var envVarCreate models.EnvVariableCreate

	if err := json.NewDecoder(r.Body).Decode(&envVarCreate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if err := validation.ValidateEnvVarName(envVarCreate.Name); err != nil {
		http.Error(w, fmt.Sprintf("Invalid name: %v", err), http.StatusBadRequest)
		return
	}

	if err := validation.ValidateEnvVarValue(envVarCreate.Value); err != nil {
		http.Error(w, fmt.Sprintf("Invalid value: %v", err), http.StatusBadRequest)
		return
	}

	repo := repository.NewEnvVariableRepository(s.db)

	envVar, err := repo.Create(&envVarCreate)
	if err != nil {
		log.Printf("Error creating environment variable: %v", err)
		http.Error(w, "Failed to create environment variable", http.StatusInternalServerError)
		return
	}

	// Return with masked value
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(envVar.ToResponse(false))
}

// handleGetEnvVariable godoc
// @Summary Get an environment variable by ID
// @Description Get a single environment variable by its ID
// @Tags Environment Variables
// @Accept json
// @Produce json
// @Param id path int true "Environment Variable ID"
// @Param show_value query bool false "Show actual value instead of masked value"
// @Success 200 {object} models.EnvVariableResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BasicAuth
// @Router /env-variables/{id} [get]
func (s *Server) handleGetEnvVariable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid environment variable ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewEnvVariableRepository(s.db)

	envVar, err := repo.GetByID(id)
	if err != nil {
		log.Printf("Error fetching environment variable: %v", err)
		http.Error(w, "Environment variable not found", http.StatusNotFound)
		return
	}

	// Check if full value is requested
	showValue := r.URL.Query().Get("show_value") == "true"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(envVar.ToResponse(showValue))
}

// handleUpdateEnvVariable godoc
// @Summary Update an environment variable
// @Description Update an existing environment variable by its ID
// @Tags Environment Variables
// @Accept json
// @Produce json
// @Param id path int true "Environment Variable ID"
// @Param envvar body models.EnvVariableUpdate true "Environment variable update data"
// @Success 200 {object} models.EnvVariableResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /env-variables/{id} [put]
func (s *Server) handleUpdateEnvVariable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid environment variable ID", http.StatusBadRequest)
		return
	}

	var envVarUpdate models.EnvVariableUpdate

	if err := json.NewDecoder(r.Body).Decode(&envVarUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input if provided
	if envVarUpdate.Name != "" {
		if err := validation.ValidateEnvVarName(envVarUpdate.Name); err != nil {
			http.Error(w, fmt.Sprintf("Invalid name: %v", err), http.StatusBadRequest)
			return
		}
	}

	if envVarUpdate.Value != "" {
		if err := validation.ValidateEnvVarValue(envVarUpdate.Value); err != nil {
			http.Error(w, fmt.Sprintf("Invalid value: %v", err), http.StatusBadRequest)
			return
		}
	}

	repo := repository.NewEnvVariableRepository(s.db)

	envVar, err := repo.Update(id, &envVarUpdate)
	if err != nil {
		log.Printf("Error updating environment variable: %v", err)
		http.Error(w, "Failed to update environment variable", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(envVar.ToResponse(false))
}

// handleDeleteEnvVariable godoc
// @Summary Delete an environment variable
// @Description Delete an environment variable by its ID
// @Tags Environment Variables
// @Accept json
// @Produce json
// @Param id path int true "Environment Variable ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /env-variables/{id} [delete]
func (s *Server) handleDeleteEnvVariable(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid environment variable ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewEnvVariableRepository(s.db)

	if err := repo.Delete(id); err != nil {
		log.Printf("Error deleting environment variable: %v", err)
		http.Error(w, "Failed to delete environment variable", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleListBashScripts godoc
// @Summary List all bash scripts
// @Description Get a list of all bash scripts (without content by default)
// @Tags Bash Scripts
// @Accept json
// @Produce json
// @Param group query string false "Filter by group name"
// @Success 200 {array} models.BashScriptResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /bash-scripts [get]
func (s *Server) handleListBashScripts(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewBashScriptRepository(s.db)
	group := r.URL.Query().Get("group")

	var scripts []*models.BashScript
	var err error

	if group != "" {
		scripts, err = repo.GetByGroup(group)
	} else {
		scripts, err = repo.GetAll()
	}
	if err != nil {
		log.Printf("Error fetching bash scripts: %v", err)
		http.Error(w, "Failed to fetch bash scripts", http.StatusInternalServerError)
		return
	}

	// Merge with Vault scripts
	scripts = s.mergeScriptsWithVault(r.Context(), scripts)

	// Filter Vault scripts by group if specified
	if group != "" {
		filtered := make([]*models.BashScript, 0)
		for _, s := range scripts {
			if s.Group == group || (s.Group == "" && group == "default") {
				filtered = append(filtered, s)
			}
		}
		scripts = filtered
	}

	// Convert to response format (without content for listing)
	responses := models.BashScriptsToList(scripts)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// handleCreateBashScript godoc
// @Summary Create a bash script
// @Description Create a new bash script (stored encrypted)
// @Tags Bash Scripts
// @Accept json
// @Produce json
// @Param script body models.BashScriptCreate true "Bash script to create"
// @Success 201 {object} models.BashScriptResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /bash-scripts [post]
func (s *Server) handleCreateBashScript(w http.ResponseWriter, r *http.Request) {
	var scriptCreate models.BashScriptCreate

	if err := json.NewDecoder(r.Body).Decode(&scriptCreate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if err := validation.ValidateBashScriptName(scriptCreate.Name); err != nil {
		http.Error(w, fmt.Sprintf("Invalid name: %v", err), http.StatusBadRequest)
		return
	}

	if err := validation.ValidateBashScriptContent(scriptCreate.Content); err != nil {
		http.Error(w, fmt.Sprintf("Invalid content: %v", err), http.StatusBadRequest)
		return
	}

	if err := validation.ValidateBashScriptFilename(scriptCreate.Filename); err != nil {
		http.Error(w, fmt.Sprintf("Invalid filename: %v", err), http.StatusBadRequest)
		return
	}

	repo := repository.NewBashScriptRepository(s.db)

	script, err := repo.Create(&scriptCreate)
	if err != nil {
		log.Printf("Error creating bash script: %v", err)
		http.Error(w, "Failed to create bash script", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(script.ToResponse(true))
}

// handleGetBashScript godoc
// @Summary Get a bash script by ID
// @Description Get a single bash script by its ID
// @Tags Bash Scripts
// @Accept json
// @Produce json
// @Param id path int true "Bash Script ID"
// @Param include_content query bool false "Include script content in response" default(true)
// @Success 200 {object} models.BashScriptResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BasicAuth
// @Router /bash-scripts/{id} [get]
func (s *Server) handleGetBashScript(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid bash script ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewBashScriptRepository(s.db)

	script, err := repo.GetByID(id)
	if err != nil {
		log.Printf("Error fetching bash script: %v", err)
		http.Error(w, "Bash script not found", http.StatusNotFound)
		return
	}

	// Check if content is requested (default true for single item)
	includeContent := r.URL.Query().Get("include_content") != "false"

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(script.ToResponse(includeContent))
}

// handleUpdateBashScript godoc
// @Summary Update a bash script
// @Description Update an existing bash script by its ID
// @Tags Bash Scripts
// @Accept json
// @Produce json
// @Param id path int true "Bash Script ID"
// @Param script body models.BashScriptUpdate true "Bash script update data"
// @Success 200 {object} models.BashScriptResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /bash-scripts/{id} [put]
func (s *Server) handleUpdateBashScript(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid bash script ID", http.StatusBadRequest)
		return
	}

	var scriptUpdate models.BashScriptUpdate

	if err := json.NewDecoder(r.Body).Decode(&scriptUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input if provided
	if scriptUpdate.Name != "" {
		if err := validation.ValidateBashScriptName(scriptUpdate.Name); err != nil {
			http.Error(w, fmt.Sprintf("Invalid name: %v", err), http.StatusBadRequest)
			return
		}
	}

	if scriptUpdate.Content != "" {
		if err := validation.ValidateBashScriptContent(scriptUpdate.Content); err != nil {
			http.Error(w, fmt.Sprintf("Invalid content: %v", err), http.StatusBadRequest)
			return
		}
	}

	if scriptUpdate.Filename != "" {
		if err := validation.ValidateBashScriptFilename(scriptUpdate.Filename); err != nil {
			http.Error(w, fmt.Sprintf("Invalid filename: %v", err), http.StatusBadRequest)
			return
		}
	}

	repo := repository.NewBashScriptRepository(s.db)

	script, err := repo.Update(id, &scriptUpdate)
	if err != nil {
		log.Printf("Error updating bash script: %v", err)
		http.Error(w, "Failed to update bash script", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(script.ToResponse(true))
}

// handleDeleteBashScript godoc
// @Summary Delete a bash script
// @Description Delete a bash script by its ID
// @Tags Bash Scripts
// @Accept json
// @Produce json
// @Param id path int true "Bash Script ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /bash-scripts/{id} [delete]
func (s *Server) handleDeleteBashScript(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid bash script ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewBashScriptRepository(s.db)

	if err := repo.Delete(id); err != nil {
		log.Printf("Error deleting bash script: %v", err)
		http.Error(w, "Failed to delete bash script", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleExecuteScript godoc
// @Summary Execute a bash script
// @Description Execute a stored bash script locally or remotely
// @Tags Bash Scripts
// @Accept json
// @Produce json
// @Param execution body models.ScriptExecution true "Script execution request"
// @Success 200 {object} models.ScriptResult
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /bash-scripts/execute [post]
func (s *Server) handleExecuteScript(w http.ResponseWriter, r *http.Request) {
	var exec models.ScriptExecution

	if err := json.NewDecoder(r.Body).Decode(&exec); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input - either ScriptID or ScriptName must be provided
	if exec.ScriptID == 0 && exec.ScriptName == "" {
		http.Error(w, "Script ID or Script Name is required", http.StatusBadRequest)
		return
	}

	// Validate and default user
	if exec.User == "" {
		exec.User = "root"
	} else if err := validation.ValidateUsername(exec.User); err != nil {
		http.Error(w, fmt.Sprintf("Invalid user: %v", err), http.StatusBadRequest)
		return
	}

	// Fetch the script - support both ID (SQLite) and Name (Vault)
	var script *models.BashScript
	var err error
	if exec.ScriptID > 0 {
		scriptRepo := repository.NewBashScriptRepository(s.db)
		script, err = scriptRepo.GetByID(exec.ScriptID)
		if err != nil {
			log.Printf("Error fetching script by ID: %v", err)
			http.Error(w, "Script not found", http.StatusNotFound)
			return
		}
	} else if exec.ScriptName != "" {
		script, err = s.getScriptByNameFromVault(r.Context(), exec.ScriptGroup, exec.ScriptName)
		if err != nil {
			log.Printf("Error fetching script from Vault: %v", err)
			http.Error(w, "Script not found in Vault", http.StatusNotFound)
			return
		}
		if script == nil {
			http.Error(w, "Script not found in Vault", http.StatusNotFound)
			return
		}
	}

	// Build the script content with optional env vars
	var scriptContent strings.Builder
	envVarsCount := 0

	// Determine which env vars to include
	// Priority: EnvVarIDs (specific selection) > IncludeEnvVars (all) > none
	envRepo := repository.NewEnvVariableRepository(s.db)

	if len(exec.EnvVarIDs) > 0 || len(exec.EnvVarNames) > 0 {
		// Fetch specific environment variables by ID (SQLite)
		for _, envVarID := range exec.EnvVarIDs {
			envVar, err := envRepo.GetByID(envVarID)
			if err != nil {
				log.Printf("Warning: env variable ID %d not found: %v", envVarID, err)
				continue
			}
			// Escape single quotes in the value for safe shell export
			escapedValue := strings.ReplaceAll(envVar.Value, "'", "'\\''")
			scriptContent.WriteString(fmt.Sprintf("export %s='%s'\n", envVar.Name, escapedValue))
			envVarsCount++
		}
		// Fetch specific environment variables by Name (Vault)
		for i, envVarName := range exec.EnvVarNames {
			// Get group from EnvVarGroups if available, otherwise use default
			envVarGroup := "default"
			if i < len(exec.EnvVarGroups) {
				envVarGroup = exec.EnvVarGroups[i]
			}
			envVar, err := s.getEnvVariableByNameFromVault(r.Context(), envVarGroup, envVarName)
			if err != nil {
				log.Printf("Warning: env variable '%s' not found in Vault: %v", envVarName, err)
				continue
			}
			if envVar == nil {
				log.Printf("Warning: env variable '%s' not found in Vault", envVarName)
				continue
			}
			escapedValue := strings.ReplaceAll(envVar.Value, "'", "'\\''")
			scriptContent.WriteString(fmt.Sprintf("export %s='%s'\n", envVar.Name, escapedValue))
			envVarsCount++
		}
	} else if exec.IncludeEnvVars {
		// Backwards compatibility: fetch all environment variables
		envVars, err := envRepo.GetAll()
		if err != nil {
			log.Printf("Error fetching environment variables: %v", err)
			http.Error(w, "Failed to fetch environment variables", http.StatusInternalServerError)
			return
		}

		// Prepend env vars as export statements
		for _, envVar := range envVars {
			// Escape single quotes in the value for safe shell export
			escapedValue := strings.ReplaceAll(envVar.Value, "'", "'\\''")
			scriptContent.WriteString(fmt.Sprintf("export %s='%s'\n", envVar.Name, escapedValue))
			envVarsCount++
		}
	}

	// Append the actual script content
	scriptContent.WriteString(script.Content)

	finalScript := scriptContent.String()

	var result *executor.ExecuteResult
	serverName := "local"

	if exec.IsRemote {
		// Remote execution via SSH
		var server *models.Server

		// Get server details - support both ID (SQLite) and Name (Vault)
		if exec.ServerID != nil && *exec.ServerID > 0 {
			serverRepo := repository.NewServerRepository(s.db)
			server, err = serverRepo.GetByID(*exec.ServerID)
			if err != nil {
				log.Printf("Error fetching server by ID: %v", err)
				http.Error(w, "Server not found", http.StatusNotFound)
				return
			}
		} else if exec.ServerName != "" {
			server, err = s.getServerByNameFromVault(r.Context(), exec.ServerGroup, exec.ServerName)
			if err != nil {
				log.Printf("Error fetching server from Vault: %v", err)
				http.Error(w, "Server not found in Vault", http.StatusNotFound)
				return
			}
			if server == nil {
				http.Error(w, "Server not found in Vault", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "Server ID or Server Name is required for remote execution", http.StatusBadRequest)
			return
		}

		// Get SSH key if provided - support both ID (SQLite) and Name (Vault)
		var privateKey string
		if exec.SSHKeyID != nil && *exec.SSHKeyID > 0 {
			keyRepo := repository.NewSSHKeyRepository(s.db)
			key, err := keyRepo.GetByID(*exec.SSHKeyID)
			if err != nil {
				log.Printf("Error fetching SSH key by ID: %v", err)
				http.Error(w, "SSH key not found", http.StatusNotFound)
				return
			}
			privateKey = key.PrivateKey
		} else if exec.SSHKeyName != "" {
			key, err := s.getSSHKeyByNameFromVault(r.Context(), exec.SSHKeyGroup, exec.SSHKeyName)
			if err != nil {
				log.Printf("Error fetching SSH key from Vault: %v", err)
				http.Error(w, "SSH key not found in Vault", http.StatusNotFound)
				return
			}
			if key == nil {
				http.Error(w, fmt.Sprintf("SSH key '%s' not found in Vault", exec.SSHKeyName), http.StatusNotFound)
				return
			}
			if key.PrivateKey == "" {
				http.Error(w, fmt.Sprintf("SSH key '%s' has no private key data in Vault", exec.SSHKeyName), http.StatusBadRequest)
				return
			}
			privateKey = key.PrivateKey
		}

		// Set server name for response
		if server.Name != "" {
			serverName = server.Name
		} else if server.IPAddress != "" {
			serverName = server.IPAddress
		}

		// Execute remotely
		remoteExec := executor.NewRemoteExecutorWithHostKeys("", true)
		sshConfig := &executor.SSHConfig{
			Host:       server.IPAddress,
			Port:       server.Port,
			Username:   exec.User,
			PrivateKey: privateKey,
			Password:   exec.SSHPassword,
		}
		result = remoteExec.Execute(context.Background(), finalScript, sshConfig)
	} else {
		// Local execution
		localExec := executor.NewLocalExecutor()
		result = localExec.Execute(context.Background(), finalScript, exec.User, exec.SudoPassword)
	}

	// Store in command history
	exitCode := result.ExitCode
	historyRepo := repository.NewCommandHistoryRepository(s.db)
	_, histErr := historyRepo.Create(&models.CommandHistoryCreate{
		Command:         fmt.Sprintf("[Script: %s] %s", script.Name, script.Content[:min(100, len(script.Content))]),
		Output:          result.Output,
		ExitCode:        &exitCode,
		Server:          serverName,
		User:            exec.User,
		ExecutionTimeMs: result.ExecutionTime,
	})
	if histErr != nil {
		log.Printf("Warning: failed to save command history: %v", histErr)
	}

	// Audit log the script execution
	audit.GetLogger().LogScriptExecution(r, script.Name, exec.User, serverName, exitCode, result.ExecutionTime, result.Error)

	// Return result - include error in output if present
	scriptOutput := result.Output
	if result.Error != nil && scriptOutput == "" {
		scriptOutput = fmt.Sprintf("Error: %s", result.Error.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.ScriptResult{
		ScriptID:      script.ID,
		ScriptName:    script.Name,
		Output:        scriptOutput,
		ExitCode:      result.ExitCode,
		User:          exec.User,
		Server:        serverName,
		ExecutionTime: result.ExecutionTime,
		EnvVarsCount:  envVarsCount,
	})
}

// StreamMessage represents a message sent via SSE
type StreamMessage struct {
	Type   string               `json:"type"`             // "output", "result", "error"
	Data   string               `json:"data"`             // output chunk or error message
	Result *models.ScriptResult `json:"result,omitempty"` // final result
}

// handleExecuteScriptStream godoc
// @Summary Execute a bash script with streaming output
// @Description Execute a stored bash script locally or remotely with real-time output streaming via SSE
// @Tags Bash Scripts
// @Accept json
// @Produce text/event-stream
// @Param execution body models.ScriptExecution true "Script execution request"
// @Success 200 {object} StreamMessage
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /bash-scripts/execute/stream [post]
func (s *Server) handleExecuteScriptStream(w http.ResponseWriter, r *http.Request) {
	var exec models.ScriptExecution

	if err := json.NewDecoder(r.Body).Decode(&exec); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input - either ScriptID or ScriptName must be provided
	if exec.ScriptID == 0 && exec.ScriptName == "" {
		http.Error(w, "Script ID or Script Name is required", http.StatusBadRequest)
		return
	}

	// Validate and default user
	if exec.User == "" {
		exec.User = "root"
	} else if err := validation.ValidateUsername(exec.User); err != nil {
		http.Error(w, fmt.Sprintf("Invalid user: %v", err), http.StatusBadRequest)
		return
	}

	// Fetch the script - support both ID (SQLite) and Name (Vault)
	var script *models.BashScript
	var err error
	if exec.ScriptID > 0 {
		scriptRepo := repository.NewBashScriptRepository(s.db)
		script, err = scriptRepo.GetByID(exec.ScriptID)
		if err != nil {
			log.Printf("Error fetching script by ID: %v", err)
			http.Error(w, "Script not found", http.StatusNotFound)
			return
		}
	} else if exec.ScriptName != "" {
		script, err = s.getScriptByNameFromVault(r.Context(), exec.ScriptGroup, exec.ScriptName)
		if err != nil {
			log.Printf("Error fetching script from Vault: %v", err)
			http.Error(w, "Script not found in Vault", http.StatusNotFound)
			return
		}
		if script == nil {
			http.Error(w, "Script not found in Vault", http.StatusNotFound)
			return
		}
	}

	// Build the script content with optional env vars
	var scriptContent strings.Builder
	envVarsCount := 0

	// Determine which env vars to include
	envRepo := repository.NewEnvVariableRepository(s.db)

	if len(exec.EnvVarIDs) > 0 || len(exec.EnvVarNames) > 0 {
		// Fetch specific environment variables by ID (SQLite)
		for _, envVarID := range exec.EnvVarIDs {
			envVar, err := envRepo.GetByID(envVarID)
			if err != nil {
				log.Printf("Warning: env variable ID %d not found: %v", envVarID, err)
				continue
			}
			escapedValue := strings.ReplaceAll(envVar.Value, "'", "'\\''")
			scriptContent.WriteString(fmt.Sprintf("export %s='%s'\n", envVar.Name, escapedValue))
			envVarsCount++
		}
		// Fetch specific environment variables by Name (Vault)
		for i, envVarName := range exec.EnvVarNames {
			// Get group from EnvVarGroups if available, otherwise use default
			envVarGroup := "default"
			if i < len(exec.EnvVarGroups) {
				envVarGroup = exec.EnvVarGroups[i]
			}
			envVar, err := s.getEnvVariableByNameFromVault(r.Context(), envVarGroup, envVarName)
			if err != nil {
				log.Printf("Warning: env variable '%s' not found in Vault: %v", envVarName, err)
				continue
			}
			if envVar == nil {
				log.Printf("Warning: env variable '%s' not found in Vault", envVarName)
				continue
			}
			escapedValue := strings.ReplaceAll(envVar.Value, "'", "'\\''")
			scriptContent.WriteString(fmt.Sprintf("export %s='%s'\n", envVar.Name, escapedValue))
			envVarsCount++
		}
	} else if exec.IncludeEnvVars {
		envVars, err := envRepo.GetAll()
		if err != nil {
			log.Printf("Error fetching environment variables: %v", err)
			http.Error(w, "Failed to fetch environment variables", http.StatusInternalServerError)
			return
		}
		for _, envVar := range envVars {
			escapedValue := strings.ReplaceAll(envVar.Value, "'", "'\\''")
			scriptContent.WriteString(fmt.Sprintf("export %s='%s'\n", envVar.Name, escapedValue))
			envVarsCount++
		}
	}

	scriptContent.WriteString(script.Content)
	finalScript := scriptContent.String()

	serverName := "local"

	// Set up SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Send initial message
	sendSSE(w, flusher, "status", "Starting script execution...")

	ctx := r.Context()

	if exec.IsRemote {
		// Remote execution via SSH with streaming
		var server *models.Server

		// Get server details - support both ID (SQLite) and Name (Vault)
		if exec.ServerID != nil && *exec.ServerID > 0 {
			serverRepo := repository.NewServerRepository(s.db)
			server, err = serverRepo.GetByID(*exec.ServerID)
			if err != nil {
				log.Printf("Error fetching server by ID: %v", err)
				sendSSE(w, flusher, "error", "Server not found")
				return
			}
		} else if exec.ServerName != "" {
			server, err = s.getServerByNameFromVault(r.Context(), exec.ServerGroup, exec.ServerName)
			if err != nil {
				log.Printf("Error fetching server from Vault: %v", err)
				sendSSE(w, flusher, "error", "Server not found in Vault")
				return
			}
			if server == nil {
				sendSSE(w, flusher, "error", "Server not found in Vault")
				return
			}
		} else {
			sendSSE(w, flusher, "error", "Server ID or Server Name is required for remote execution")
			return
		}

		// Get SSH key if provided - support both ID (SQLite) and Name (Vault)
		var privateKey string
		if exec.SSHKeyID != nil && *exec.SSHKeyID > 0 {
			keyRepo := repository.NewSSHKeyRepository(s.db)
			key, err := keyRepo.GetByID(*exec.SSHKeyID)
			if err != nil {
				log.Printf("Error fetching SSH key by ID: %v", err)
				sendSSE(w, flusher, "error", "SSH key not found")
				return
			}
			privateKey = key.PrivateKey
		} else if exec.SSHKeyName != "" {
			key, err := s.getSSHKeyByNameFromVault(r.Context(), exec.SSHKeyGroup, exec.SSHKeyName)
			if err != nil {
				log.Printf("Error fetching SSH key from Vault: %v", err)
				sendSSE(w, flusher, "error", "SSH key not found in Vault")
				return
			}
			if key == nil {
				sendSSE(w, flusher, "error", fmt.Sprintf("SSH key '%s' not found in Vault", exec.SSHKeyName))
				return
			}
			if key.PrivateKey == "" {
				sendSSE(w, flusher, "error", fmt.Sprintf("SSH key '%s' has no private key data in Vault", exec.SSHKeyName))
				return
			}
			privateKey = key.PrivateKey
		}

		if server.Name != "" {
			serverName = server.Name
		} else if server.IPAddress != "" {
			serverName = server.IPAddress
		}

		sendSSE(w, flusher, "status", fmt.Sprintf("Connecting to %s...", serverName))

		// Execute with streaming
		remoteExec := executor.NewRemoteExecutorWithHostKeys("", true)
		sshConfig := &executor.SSHConfig{
			Host:       server.IPAddress,
			Port:       server.Port,
			Username:   exec.User,
			PrivateKey: privateKey,
			Password:   exec.SSHPassword,
		}

		outputChan, resultChan := remoteExec.ExecuteWithStreaming(ctx, finalScript, sshConfig)

		// Stream output
		var fullOutput strings.Builder
		for chunk := range outputChan {
			fullOutput.WriteString(chunk)
			sendSSE(w, flusher, "output", chunk)
		}

		// Get final result
		result := <-resultChan

		// Save to history
		exitCode := result.ExitCode
		historyRepo := repository.NewCommandHistoryRepository(s.db)
		_, err = historyRepo.Create(&models.CommandHistoryCreate{
			Command:         fmt.Sprintf("[Script: %s] %s", script.Name, script.Content[:min(100, len(script.Content))]),
			Output:          result.Output,
			ExitCode:        &exitCode,
			Server:          serverName,
			User:            exec.User,
			ExecutionTimeMs: result.ExecutionTime,
		})
		if err != nil {
			log.Printf("Warning: failed to save command history: %v", err)
		}

		// Audit log the script execution
		audit.GetLogger().LogScriptExecution(r, script.Name, exec.User, serverName, exitCode, result.ExecutionTime, result.Error)

		// Send final result
		scriptResult := models.ScriptResult{
			ScriptID:      script.ID,
			ScriptName:    script.Name,
			Output:        result.Output,
			ExitCode:      result.ExitCode,
			User:          exec.User,
			Server:        serverName,
			ExecutionTime: result.ExecutionTime,
			EnvVarsCount:  envVarsCount,
		}
		sendSSEResult(w, flusher, &scriptResult)

	} else {
		// Local execution with streaming
		localExec := executor.NewLocalExecutor()
		outputChan, resultChan := localExec.ExecuteWithStreaming(ctx, finalScript, exec.User, exec.SudoPassword)

		// Stream output
		var fullOutput strings.Builder
		for chunk := range outputChan {
			fullOutput.WriteString(chunk)
			sendSSE(w, flusher, "output", chunk)
		}

		// Get final result
		result := <-resultChan

		// Save to history
		exitCode := result.ExitCode
		historyRepo := repository.NewCommandHistoryRepository(s.db)
		_, err = historyRepo.Create(&models.CommandHistoryCreate{
			Command:         fmt.Sprintf("[Script: %s] %s", script.Name, script.Content[:min(100, len(script.Content))]),
			Output:          result.Output,
			ExitCode:        &exitCode,
			Server:          serverName,
			User:            exec.User,
			ExecutionTimeMs: result.ExecutionTime,
		})
		if err != nil {
			log.Printf("Warning: failed to save command history: %v", err)
		}

		// Audit log the script execution
		audit.GetLogger().LogScriptExecution(r, script.Name, exec.User, serverName, exitCode, result.ExecutionTime, result.Error)

		// Send final result
		scriptOutput := result.Output
		if result.Error != nil && scriptOutput == "" {
			scriptOutput = fmt.Sprintf("Error: %s", result.Error.Error())
		}

		scriptResult := models.ScriptResult{
			ScriptID:      script.ID,
			ScriptName:    script.Name,
			Output:        scriptOutput,
			ExitCode:      result.ExitCode,
			User:          exec.User,
			Server:        serverName,
			ExecutionTime: result.ExecutionTime,
			EnvVarsCount:  envVarsCount,
		}
		sendSSEResult(w, flusher, &scriptResult)
	}
}

// sendSSE sends a Server-Sent Event message
func sendSSE(w http.ResponseWriter, flusher http.Flusher, eventType, data string) {
	msg := StreamMessage{
		Type: eventType,
		Data: data,
	}
	jsonData, _ := json.Marshal(msg)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}

// sendSSEResult sends the final result via SSE
func sendSSEResult(w http.ResponseWriter, flusher http.Flusher, result *models.ScriptResult) {
	msg := StreamMessage{
		Type:   "result",
		Result: result,
	}
	jsonData, _ := json.Marshal(msg)
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
}

// ========== Script Preset Handlers ==========

// handleListScriptPresets godoc
// @Summary List all script presets
// @Description Get a list of all script execution presets
// @Tags Script Presets
// @Accept json
// @Produce json
// @Success 200 {array} models.ScriptPresetResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /script-presets [get]
func (s *Server) handleListScriptPresets(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewScriptPresetRepository(s.db)

	presets, err := repo.GetAll()
	if err != nil {
		log.Printf("Error fetching script presets: %v", err)
		http.Error(w, "Failed to fetch script presets", http.StatusInternalServerError)
		return
	}

	responses := models.ScriptPresetsToList(presets)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// handleCreateScriptPreset godoc
// @Summary Create a script preset
// @Description Create a new script execution preset
// @Tags Script Presets
// @Accept json
// @Produce json
// @Param preset body models.ScriptPresetCreate true "Script preset to create"
// @Success 201 {object} models.ScriptPresetResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /script-presets [post]
func (s *Server) handleCreateScriptPreset(w http.ResponseWriter, r *http.Request) {
	var presetCreate models.ScriptPresetCreate

	if err := json.NewDecoder(r.Body).Decode(&presetCreate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if presetCreate.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if presetCreate.ScriptID == 0 {
		http.Error(w, "Script ID is required", http.StatusBadRequest)
		return
	}

	// Verify the script exists
	scriptRepo := repository.NewBashScriptRepository(s.db)
	_, err := scriptRepo.GetByID(presetCreate.ScriptID)
	if err != nil {
		http.Error(w, "Script not found", http.StatusBadRequest)
		return
	}

	// Verify env var IDs exist if provided
	if len(presetCreate.EnvVarIDs) > 0 {
		envRepo := repository.NewEnvVariableRepository(s.db)
		for _, envVarID := range presetCreate.EnvVarIDs {
			_, err := envRepo.GetByID(envVarID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Environment variable with ID %d not found", envVarID), http.StatusBadRequest)
				return
			}
		}
	}

	// Verify server exists if provided
	if presetCreate.ServerID != nil {
		serverRepo := repository.NewServerRepository(s.db)
		_, err := serverRepo.GetByID(*presetCreate.ServerID)
		if err != nil {
			http.Error(w, "Server not found", http.StatusBadRequest)
			return
		}
	}

	// Verify SSH key exists if provided
	if presetCreate.SSHKeyID != nil {
		keyRepo := repository.NewSSHKeyRepository(s.db)
		_, err := keyRepo.GetByID(*presetCreate.SSHKeyID)
		if err != nil {
			http.Error(w, "SSH key not found", http.StatusBadRequest)
			return
		}
	}

	repo := repository.NewScriptPresetRepository(s.db)

	preset, err := repo.Create(&presetCreate)
	if err != nil {
		log.Printf("Error creating script preset: %v", err)
		http.Error(w, "Failed to create script preset", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(preset.ToResponse())
}

// handleGetScriptPreset godoc
// @Summary Get a script preset by ID
// @Description Get a single script preset by its ID
// @Tags Script Presets
// @Accept json
// @Produce json
// @Param id path int true "Script Preset ID"
// @Success 200 {object} models.ScriptPresetResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Security BasicAuth
// @Router /script-presets/{id} [get]
func (s *Server) handleGetScriptPreset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid script preset ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewScriptPresetRepository(s.db)

	preset, err := repo.GetByID(id)
	if err != nil {
		log.Printf("Error fetching script preset: %v", err)
		http.Error(w, "Script preset not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preset.ToResponse())
}

// handleUpdateScriptPreset godoc
// @Summary Update a script preset
// @Description Update an existing script preset by its ID
// @Tags Script Presets
// @Accept json
// @Produce json
// @Param id path int true "Script Preset ID"
// @Param preset body models.ScriptPresetUpdate true "Script preset update data"
// @Success 200 {object} models.ScriptPresetResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /script-presets/{id} [put]
func (s *Server) handleUpdateScriptPreset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid script preset ID", http.StatusBadRequest)
		return
	}

	var presetUpdate models.ScriptPresetUpdate

	if err := json.NewDecoder(r.Body).Decode(&presetUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Verify script exists if being updated
	if presetUpdate.ScriptID != nil {
		scriptRepo := repository.NewBashScriptRepository(s.db)
		_, err := scriptRepo.GetByID(*presetUpdate.ScriptID)
		if err != nil {
			http.Error(w, "Script not found", http.StatusBadRequest)
			return
		}
	}

	// Verify env var IDs exist if being updated
	if len(presetUpdate.EnvVarIDs) > 0 {
		envRepo := repository.NewEnvVariableRepository(s.db)
		for _, envVarID := range presetUpdate.EnvVarIDs {
			_, err := envRepo.GetByID(envVarID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Environment variable with ID %d not found", envVarID), http.StatusBadRequest)
				return
			}
		}
	}

	// Verify server exists if being updated
	if presetUpdate.ServerID != nil {
		serverRepo := repository.NewServerRepository(s.db)
		_, err := serverRepo.GetByID(*presetUpdate.ServerID)
		if err != nil {
			http.Error(w, "Server not found", http.StatusBadRequest)
			return
		}
	}

	// Verify SSH key exists if being updated
	if presetUpdate.SSHKeyID != nil {
		keyRepo := repository.NewSSHKeyRepository(s.db)
		_, err := keyRepo.GetByID(*presetUpdate.SSHKeyID)
		if err != nil {
			http.Error(w, "SSH key not found", http.StatusBadRequest)
			return
		}
	}

	repo := repository.NewScriptPresetRepository(s.db)

	preset, err := repo.Update(id, &presetUpdate)
	if err != nil {
		log.Printf("Error updating script preset: %v", err)
		http.Error(w, "Failed to update script preset", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preset.ToResponse())
}

// handleDeleteScriptPreset godoc
// @Summary Delete a script preset
// @Description Delete a script preset by its ID
// @Tags Script Presets
// @Accept json
// @Produce json
// @Param id path int true "Script Preset ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /script-presets/{id} [delete]
func (s *Server) handleDeleteScriptPreset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid script preset ID", http.StatusBadRequest)
		return
	}

	repo := repository.NewScriptPresetRepository(s.db)

	if err := repo.Delete(id); err != nil {
		log.Printf("Error deleting script preset: %v", err)
		http.Error(w, "Failed to delete script preset", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGetScriptPresetsByScript godoc
// @Summary Get presets for a script
// @Description Get all presets for a specific bash script
// @Tags Script Presets
// @Accept json
// @Produce json
// @Param id path int true "Bash Script ID"
// @Success 200 {array} models.ScriptPresetResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /bash-scripts/{id}/presets [get]
func (s *Server) handleGetScriptPresetsByScript(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	scriptID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid script ID", http.StatusBadRequest)
		return
	}

	// Verify script exists
	scriptRepo := repository.NewBashScriptRepository(s.db)
	_, err = scriptRepo.GetByID(scriptID)
	if err != nil {
		http.Error(w, "Script not found", http.StatusNotFound)
		return
	}

	repo := repository.NewScriptPresetRepository(s.db)

	presets, err := repo.GetByScriptID(scriptID)
	if err != nil {
		log.Printf("Error fetching script presets: %v", err)
		http.Error(w, "Failed to fetch script presets", http.StatusInternalServerError)
		return
	}

	responses := models.ScriptPresetsToList(presets)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responses)
}

// handleListSSHKeyGroups godoc
// @Summary List all SSH key groups
// @Description Get a list of all distinct group names for SSH keys
// @Tags SSH Keys
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /keys/groups [get]
func (s *Server) handleListSSHKeyGroups(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewSSHKeyRepository(s.db)

	groups, err := repo.GetGroups()
	if err != nil {
		log.Printf("Error fetching SSH key groups: %v", err)
		http.Error(w, "Failed to fetch groups", http.StatusInternalServerError)
		return
	}

	// Merge with Vault groups if available
	client := s.getVaultClientIfEnabled()
	if client != nil {
		vaultGroups, err := client.ListSSHKeyGroups(r.Context())
		if err == nil {
			// Add Vault groups that don't exist in SQLite
			groupSet := make(map[string]bool)
			for _, g := range groups {
				groupSet[g] = true
			}
			for _, g := range vaultGroups {
				if !groupSet[g] {
					groups = append(groups, g)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// handleListServerGroups godoc
// @Summary List all server groups
// @Description Get a list of all distinct group names for servers
// @Tags Servers
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /servers/groups [get]
func (s *Server) handleListServerGroups(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewServerRepository(s.db)

	groups, err := repo.GetGroups()
	if err != nil {
		log.Printf("Error fetching server groups: %v", err)
		http.Error(w, "Failed to fetch groups", http.StatusInternalServerError)
		return
	}

	// Merge with Vault groups if available
	client := s.getVaultClientIfEnabled()
	if client != nil {
		vaultGroups, err := client.ListServerGroups(r.Context())
		if err == nil {
			groupSet := make(map[string]bool)
			for _, g := range groups {
				groupSet[g] = true
			}
			for _, g := range vaultGroups {
				if !groupSet[g] {
					groups = append(groups, g)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// handleListEnvVariableGroups godoc
// @Summary List all environment variable groups
// @Description Get a list of all distinct group names for environment variables
// @Tags Environment Variables
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /env-variables/groups [get]
func (s *Server) handleListEnvVariableGroups(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewEnvVariableRepository(s.db)

	groups, err := repo.GetGroups()
	if err != nil {
		log.Printf("Error fetching environment variable groups: %v", err)
		http.Error(w, "Failed to fetch groups", http.StatusInternalServerError)
		return
	}

	// Merge with Vault groups if available
	client := s.getVaultClientIfEnabled()
	if client != nil {
		vaultGroups, err := client.ListEnvVariableGroups(r.Context())
		if err == nil {
			groupSet := make(map[string]bool)
			for _, g := range groups {
				groupSet[g] = true
			}
			for _, g := range vaultGroups {
				if !groupSet[g] {
					groups = append(groups, g)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// handleListBashScriptGroups godoc
// @Summary List all bash script groups
// @Description Get a list of all distinct group names for bash scripts
// @Tags Bash Scripts
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /bash-scripts/groups [get]
func (s *Server) handleListBashScriptGroups(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewBashScriptRepository(s.db)

	groups, err := repo.GetGroups()
	if err != nil {
		log.Printf("Error fetching bash script groups: %v", err)
		http.Error(w, "Failed to fetch groups", http.StatusInternalServerError)
		return
	}

	// Merge with Vault groups if available
	client := s.getVaultClientIfEnabled()
	if client != nil {
		vaultGroups, err := client.ListBashScriptGroups(r.Context())
		if err == nil {
			groupSet := make(map[string]bool)
			for _, g := range groups {
				groupSet[g] = true
			}
			for _, g := range vaultGroups {
				if !groupSet[g] {
					groups = append(groups, g)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}
