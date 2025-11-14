package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/user"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pozgo/web-cli/internal/executor"
	"github.com/pozgo/web-cli/internal/models"
	"github.com/pozgo/web-cli/internal/repository"
	"github.com/pozgo/web-cli/internal/validation"
)

// handleListSSHKeys returns all SSH keys
func (s *Server) handleListSSHKeys(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewSSHKeyRepository(s.db)

	keys, err := repo.GetAll()
	if err != nil {
		log.Printf("Error fetching SSH keys: %v", err)
		http.Error(w, "Failed to fetch SSH keys", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

// handleCreateSSHKey creates a new SSH key
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

// handleGetSSHKey returns a single SSH key by ID
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

// handleUpdateSSHKey updates an existing SSH key by ID
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

// handleDeleteSSHKey deletes an SSH key by ID
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

// handleListServers returns all servers
func (s *Server) handleListServers(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewServerRepository(s.db)

	servers, err := repo.GetAll()
	if err != nil {
		log.Printf("Error fetching servers: %v", err)
		http.Error(w, "Failed to fetch servers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

// handleCreateServer creates a new server
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

// handleGetServer returns a single server by ID
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

// handleUpdateServer updates an existing server by ID
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

// handleDeleteServer deletes a server by ID
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

// handleExecuteCommand executes a command locally
func (s *Server) handleExecuteCommand(w http.ResponseWriter, r *http.Request) {
	var exec models.CommandExecution

	if err := json.NewDecoder(r.Body).Decode(&exec); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if exec.Command == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}

	// Default user to root if not specified
	if exec.User == "" {
		exec.User = "root"
	}

	var result *executor.ExecuteResult
	serverName := "local"

	if exec.IsRemote {
		// Remote execution via SSH
		if exec.ServerID == nil {
			http.Error(w, "Server ID is required for remote execution", http.StatusBadRequest)
			return
		}

		// Get server details
		serverRepo := repository.NewServerRepository(s.db)
		server, err := serverRepo.GetByID(*exec.ServerID)
		if err != nil {
			log.Printf("Error fetching server: %v", err)
			http.Error(w, "Server not found", http.StatusNotFound)
			return
		}

		// Get SSH key if provided
		var privateKey string
		if exec.SSHKeyID != nil {
			keyRepo := repository.NewSSHKeyRepository(s.db)
			key, err := keyRepo.GetByID(*exec.SSHKeyID)
			if err != nil {
				log.Printf("Error fetching SSH key: %v", err)
				http.Error(w, "SSH key not found", http.StatusNotFound)
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
		remoteExec := executor.NewRemoteExecutor()
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

	// Return result
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.CommandResult{
		Command:       exec.Command,
		Output:        result.Output,
		ExitCode:      result.ExitCode,
		User:          exec.User,
		ExecutionTime: result.ExecutionTime,
		ExecutedAt:    "",
	})
}

// handleListSavedCommands returns all saved commands
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

// handleCreateSavedCommand creates a new saved command
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

// handleGetSavedCommand returns a single saved command by ID
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

// handleUpdateSavedCommand updates an existing saved command by ID
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

// handleDeleteSavedCommand deletes a saved command by ID
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

// handleListCommandHistory returns command history
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

// handleGetCommandHistory returns a single command history entry by ID
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

// handleListLocalUsers returns all local users
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

// handleCreateLocalUser creates a new local user
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

// handleGetLocalUser returns a single local user by ID
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

// handleUpdateLocalUser updates an existing local user
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// handleDeleteLocalUser deletes a local user by ID
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

// handleGetCurrentUser returns the current system user
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
