package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/pozgo/web-cli/internal/models"
	"github.com/pozgo/web-cli/internal/repository"
	"github.com/pozgo/web-cli/internal/vault"
)

// handleGetVaultConfig godoc
// @Summary Get Vault configuration
// @Description Retrieve the current Vault configuration (token is never returned)
// @Tags Vault
// @Produce json
// @Success 200 {object} models.VaultConfigResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /vault/config [get]
func (s *Server) handleGetVaultConfig(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewVaultConfigRepository(s.db)
	cfg, err := repo.Get()
	if err != nil {
		log.Printf("Error getting vault config: %v", err)
		http.Error(w, "Failed to get vault configuration", http.StatusInternalServerError)
		return
	}

	if cfg == nil {
		// Return empty response if not configured
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.VaultConfigResponse{})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg.ToResponse())
}

// handleCreateOrUpdateVaultConfig godoc
// @Summary Create or update Vault configuration
// @Description Configure or update the Vault connection settings
// @Tags Vault
// @Accept json
// @Produce json
// @Param config body models.VaultConfigCreate true "Vault configuration"
// @Success 200 {object} models.VaultConfigResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /vault/config [post]
func (s *Server) handleCreateOrUpdateVaultConfig(w http.ResponseWriter, r *http.Request) {
	var create models.VaultConfigCreate
	if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if create.Address == "" {
		http.Error(w, "Vault address is required", http.StatusBadRequest)
		return
	}

	if create.Token == "" {
		// If no token provided, check if we're updating existing config
		repo := repository.NewVaultConfigRepository(s.db)
		existing, _ := repo.Get()
		if existing == nil || existing.Token == "" {
			http.Error(w, "Vault token is required", http.StatusBadRequest)
			return
		}
		// Use existing token
		create.Token = existing.Token
	}

	// Set default mount path
	if create.MountPath == "" {
		create.MountPath = "secret"
	}

	repo := repository.NewVaultConfigRepository(s.db)
	cfg, err := repo.CreateOrUpdate(&create)
	if err != nil {
		log.Printf("Error saving vault config: %v", err)
		http.Error(w, "Failed to save vault configuration", http.StatusInternalServerError)
		return
	}

	// If Vault is enabled, initialize the structure automatically
	if cfg.Enabled {
		go func() {
			vaultCfg := &vault.Config{
				Address:   cfg.Address,
				Token:     cfg.Token,
				Namespace: cfg.Namespace,
				MountPath: cfg.MountPath,
			}

			client, err := vault.NewClient(vaultCfg)
			if err != nil {
				log.Printf("Warning: Failed to create Vault client for structure initialization: %v", err)
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			if err := client.InitializeStructure(ctx); err != nil {
				log.Printf("Warning: Failed to initialize Vault structure: %v", err)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cfg.ToResponse())
}

// handleDeleteVaultConfig godoc
// @Summary Delete Vault configuration
// @Description Remove the Vault configuration
// @Tags Vault
// @Success 204 "No Content"
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /vault/config [delete]
func (s *Server) handleDeleteVaultConfig(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewVaultConfigRepository(s.db)
	if err := repo.Delete(); err != nil {
		log.Printf("Error deleting vault config: %v", err)
		http.Error(w, "Failed to delete vault configuration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleTestVaultConnection godoc
// @Summary Test Vault connection
// @Description Test the connection to Vault and initialize the secrets structure
// @Tags Vault
// @Produce json
// @Success 200 {object} models.VaultStatus
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /vault/test [post]
func (s *Server) handleTestVaultConnection(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewVaultConfigRepository(s.db)
	cfg, err := repo.Get()
	if err != nil {
		log.Printf("Error getting vault config: %v", err)
		http.Error(w, "Failed to get vault configuration", http.StatusInternalServerError)
		return
	}

	if cfg == nil {
		http.Error(w, "Vault is not configured", http.StatusBadRequest)
		return
	}

	// Create Vault client
	vaultCfg := &vault.Config{
		Address:   cfg.Address,
		Token:     cfg.Token,
		Namespace: cfg.Namespace,
		MountPath: cfg.MountPath,
	}

	client, err := vault.NewClient(vaultCfg)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.VaultStatus{
			Configured: true,
			Enabled:    cfg.Enabled,
			Connected:  false,
			Address:    cfg.Address,
			Error:      err.Error(),
		})
		return
	}

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := client.TestConnection(ctx); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.VaultStatus{
			Configured: true,
			Enabled:    cfg.Enabled,
			Connected:  false,
			Address:    cfg.Address,
			Error:      err.Error(),
		})
		return
	}

	// Connection successful - initialize structure in background
	go func() {
		initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer initCancel()

		if err := client.InitializeStructure(initCtx); err != nil {
			log.Printf("Warning: Failed to initialize Vault structure: %v", err)
		}
	}()

	// Check health status
	var vaultSealed bool
	health, err := client.GetHealth(ctx)
	if err == nil && health != nil {
		vaultSealed = health.Sealed
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.VaultStatus{
		Configured:  true,
		Enabled:     cfg.Enabled,
		Connected:   true,
		Address:     cfg.Address,
		VaultSealed: vaultSealed,
	})
}

// handleGetVaultStatus godoc
// @Summary Get Vault connection status
// @Description Check the current Vault connection status
// @Tags Vault
// @Produce json
// @Success 200 {object} models.VaultStatus
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /vault/status [get]
func (s *Server) handleGetVaultStatus(w http.ResponseWriter, r *http.Request) {
	repo := repository.NewVaultConfigRepository(s.db)
	cfg, err := repo.Get()
	if err != nil {
		log.Printf("Error getting vault config: %v", err)
		http.Error(w, "Failed to get vault configuration", http.StatusInternalServerError)
		return
	}

	if cfg == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(models.VaultStatus{
			Configured: false,
			Enabled:    false,
			Connected:  false,
		})
		return
	}

	status := models.VaultStatus{
		Configured: true,
		Enabled:    cfg.Enabled,
		Connected:  false,
		Address:    cfg.Address,
	}

	// Only test connection if enabled
	if cfg.Enabled {
		vaultCfg := &vault.Config{
			Address:   cfg.Address,
			Token:     cfg.Token,
			Namespace: cfg.Namespace,
			MountPath: cfg.MountPath,
		}

		client, err := vault.NewClient(vaultCfg)
		if err != nil {
			status.Error = err.Error()
		} else {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			if err := client.TestConnection(ctx); err != nil {
				status.Error = err.Error()
			} else {
				status.Connected = true

				// Check health
				health, err := client.GetHealth(ctx)
				if err == nil && health != nil {
					status.VaultSealed = health.Sealed
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleListVaultSSHKeys godoc
// @Summary List SSH keys from Vault
// @Description Retrieve all SSH keys stored in Vault
// @Tags Vault
// @Produce json
// @Success 200 {array} vault.SSHKey
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /vault/ssh-keys [get]
func (s *Server) handleListVaultSSHKeys(w http.ResponseWriter, r *http.Request) {
	client, err := s.getVaultClient()
	if err != nil {
		log.Printf("Error getting vault client: %v", err)
		http.Error(w, sanitizeVaultError(err), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	keys, err := client.ListSSHKeys(ctx)
	if err != nil {
		log.Printf("Error listing vault SSH keys: %v", err)
		http.Error(w, "Failed to list SSH keys from Vault", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

// handleListVaultServers godoc
// @Summary List servers from Vault
// @Description Retrieve all server configurations stored in Vault
// @Tags Vault
// @Produce json
// @Success 200 {array} vault.Server
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /vault/servers [get]
func (s *Server) handleListVaultServers(w http.ResponseWriter, r *http.Request) {
	client, err := s.getVaultClient()
	if err != nil {
		http.Error(w, sanitizeVaultError(err), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	servers, err := client.ListServers(ctx)
	if err != nil {
		log.Printf("Error listing vault servers: %v", err)
		http.Error(w, "Failed to list servers from Vault", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(servers)
}

// handleListVaultEnvVariables godoc
// @Summary List environment variables from Vault
// @Description Retrieve all environment variables stored in Vault
// @Tags Vault
// @Produce json
// @Success 200 {array} vault.EnvVariable
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /vault/env-variables [get]
func (s *Server) handleListVaultEnvVariables(w http.ResponseWriter, r *http.Request) {
	client, err := s.getVaultClient()
	if err != nil {
		http.Error(w, sanitizeVaultError(err), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	vars, err := client.ListEnvVariables(ctx)
	if err != nil {
		log.Printf("Error listing vault env variables: %v", err)
		http.Error(w, "Failed to list environment variables from Vault", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vars)
}

// handleListVaultScripts godoc
// @Summary List bash scripts from Vault
// @Description Retrieve all bash scripts stored in Vault
// @Tags Vault
// @Produce json
// @Success 200 {array} vault.BashScript
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /vault/bash-scripts [get]
func (s *Server) handleListVaultScripts(w http.ResponseWriter, r *http.Request) {
	client, err := s.getVaultClient()
	if err != nil {
		http.Error(w, sanitizeVaultError(err), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	scripts, err := client.ListBashScripts(ctx)
	if err != nil {
		log.Printf("Error listing vault scripts: %v", err)
		http.Error(w, "Failed to list scripts from Vault", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scripts)
}

// getVaultClient creates a Vault client from stored configuration
func (s *Server) getVaultClient() (*vault.Client, error) {
	repo := repository.NewVaultConfigRepository(s.db)
	cfg, err := repo.Get()
	if err != nil {
		return nil, err
	}

	if cfg == nil {
		return nil, &VaultNotConfiguredError{}
	}

	if !cfg.Enabled {
		return nil, &VaultDisabledError{}
	}

	vaultCfg := &vault.Config{
		Address:   cfg.Address,
		Token:     cfg.Token,
		Namespace: cfg.Namespace,
		MountPath: cfg.MountPath,
	}

	return vault.NewClient(vaultCfg)
}

// VaultNotConfiguredError is returned when Vault is not configured
type VaultNotConfiguredError struct{}

func (e *VaultNotConfiguredError) Error() string {
	return "Vault is not configured"
}

// VaultDisabledError is returned when Vault is disabled
type VaultDisabledError struct{}

func (e *VaultDisabledError) Error() string {
	return "Vault integration is disabled"
}

// sanitizeVaultError returns a safe error message for API responses
// This prevents leaking internal details like IP addresses, paths, or stack traces
func sanitizeVaultError(err error) string {
	if err == nil {
		return "Unknown error"
	}

	// Check for known safe error types
	switch err.(type) {
	case *VaultNotConfiguredError:
		return err.Error()
	case *VaultDisabledError:
		return err.Error()
	}

	// Check error message for safe patterns
	errMsg := err.Error()

	// Safe messages that can be returned directly
	safeMessages := []string{
		"Vault is not configured",
		"Vault integration is disabled",
		"vault address is required",
		"vault token is required",
		"vault config is nil",
	}

	for _, safe := range safeMessages {
		if errMsg == safe {
			return errMsg
		}
	}

	// For validation errors, return them (they don't contain sensitive info)
	if containsAny(errMsg, []string{
		"invalid secret type",
		"invalid group name",
		"invalid secret name",
		"invalid vault address",
		"cannot be empty",
		"cannot contain",
		"too long",
	}) {
		return errMsg
	}

	// Generic message for all other errors to prevent information leakage
	return "Vault operation failed"
}

// containsAny checks if the string contains any of the substrings
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// handleCreateVaultSSHKey godoc
// @Summary Create SSH key in Vault
// @Description Store a new SSH key in Vault
// @Tags Vault
// @Accept json
// @Produce json
// @Param key body object{name=string,private_key=string,group=string} true "SSH Key"
// @Success 201 {object} object{name=string,group=string,created_at=string,source=string}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /vault/ssh-keys [post]
func (s *Server) handleCreateVaultSSHKey(w http.ResponseWriter, r *http.Request) {
	client, err := s.getVaultClient()
	if err != nil {
		http.Error(w, sanitizeVaultError(err), http.StatusBadRequest)
		return
	}

	var req struct {
		Name       string `json:"name"`
		PrivateKey string `json:"private_key"`
		Group      string `json:"group"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.PrivateKey == "" {
		http.Error(w, "Name and private_key are required", http.StatusBadRequest)
		return
	}

	if req.Group == "" {
		req.Group = "default"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	key := &vault.SSHKey{
		Name:       req.Name,
		PrivateKey: req.PrivateKey,
		Group:      req.Group,
		CreatedAt:  time.Now(),
	}

	if err := client.SaveSSHKey(ctx, key); err != nil {
		log.Printf("Error saving SSH key to Vault: %v", err)
		http.Error(w, "Failed to save SSH key to Vault", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":       key.Name,
		"group":      key.Group,
		"created_at": key.CreatedAt,
		"source":     "vault",
	})
}

// handleCreateVaultServer godoc
// @Summary Create server in Vault
// @Description Store a new server configuration in Vault
// @Tags Vault
// @Accept json
// @Produce json
// @Param server body object{name=string,ip_address=string,port=int,username=string,group=string} true "Server"
// @Success 201 {object} object{name=string,ip_address=string,port=int,username=string,group=string,source=string}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /vault/servers [post]
func (s *Server) handleCreateVaultServer(w http.ResponseWriter, r *http.Request) {
	client, err := s.getVaultClient()
	if err != nil {
		http.Error(w, sanitizeVaultError(err), http.StatusBadRequest)
		return
	}

	var req struct {
		Name      string `json:"name"`
		IPAddress string `json:"ip_address"`
		Port      int    `json:"port"`
		Username  string `json:"username"`
		Group     string `json:"group"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// At least one of name or ip_address is required
	if req.Name == "" && req.IPAddress == "" {
		http.Error(w, "At least one of name or ip_address is required", http.StatusBadRequest)
		return
	}

	// Use name as identifier, fallback to ip_address
	identifier := req.Name
	if identifier == "" {
		identifier = req.IPAddress
	}

	if req.Group == "" {
		req.Group = "default"
	}

	if req.Port == 0 {
		req.Port = 22
	}

	if req.Username == "" {
		req.Username = "root"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	srv := &vault.Server{
		Name:      identifier,
		IPAddress: req.IPAddress,
		Port:      req.Port,
		Username:  req.Username,
		Group:     req.Group,
	}

	if err := client.SaveServer(ctx, srv); err != nil {
		log.Printf("Error saving server to Vault: %v", err)
		http.Error(w, "Failed to save server to Vault", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":       srv.Name,
		"ip_address": srv.IPAddress,
		"port":       srv.Port,
		"username":   srv.Username,
		"group":      srv.Group,
		"source":     "vault",
	})
}

// handleCreateVaultEnvVariable godoc
// @Summary Create environment variable in Vault
// @Description Store a new environment variable in Vault
// @Tags Vault
// @Accept json
// @Produce json
// @Param envVar body object{name=string,value=string,description=string,group=string} true "Environment Variable"
// @Success 201 {object} object{name=string,description=string,group=string,source=string}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Security BasicAuth
// @Router /vault/env-variables [post]
func (s *Server) handleCreateVaultEnvVariable(w http.ResponseWriter, r *http.Request) {
	client, err := s.getVaultClient()
	if err != nil {
		http.Error(w, sanitizeVaultError(err), http.StatusBadRequest)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Value       string `json:"value"`
		Description string `json:"description"`
		Group       string `json:"group"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Value == "" {
		http.Error(w, "Name and value are required", http.StatusBadRequest)
		return
	}

	if req.Group == "" {
		req.Group = "default"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	envVar := &vault.EnvVariable{
		Name:        req.Name,
		Value:       req.Value,
		Description: req.Description,
		Group:       req.Group,
	}

	if err := client.SaveEnvVariable(ctx, envVar); err != nil {
		log.Printf("Error saving env variable to Vault: %v", err)
		http.Error(w, "Failed to save environment variable to Vault", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":        envVar.Name,
		"description": envVar.Description,
		"group":       envVar.Group,
		"source":      "vault",
	})
}

// handleCreateVaultScript godoc
// @Summary Store bash script in Vault
// @Description Create a new bash script in HashiCorp Vault storage
// @Tags Vault
// @Accept json
// @Produce json
// @Param script body object true "Bash script data" example({"name":"deploy-script","description":"Production deployment script","content":"#!/bin/bash\necho 'Deploying...'","filename":"deploy.sh","group":"production"})
// @Success 201 {object} map[string]interface{} "Created script info with source field"
// @Failure 400 {string} string "Invalid request or Vault not configured"
// @Failure 500 {string} string "Failed to store script"
// @Security BasicAuth
// @Router /vault/bash-scripts [post]
func (s *Server) handleCreateVaultScript(w http.ResponseWriter, r *http.Request) {
	client, err := s.getVaultClient()
	if err != nil {
		http.Error(w, sanitizeVaultError(err), http.StatusBadRequest)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
		Filename    string `json:"filename"`
		Group       string `json:"group"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Content == "" {
		http.Error(w, "Name and content are required", http.StatusBadRequest)
		return
	}

	if req.Group == "" {
		req.Group = "default"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	script := &vault.BashScript{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		Filename:    req.Filename,
		Group:       req.Group,
	}

	if err := client.SaveBashScript(ctx, script); err != nil {
		log.Printf("Error saving script to Vault: %v", err)
		http.Error(w, "Failed to save script to Vault", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":        script.Name,
		"description": script.Description,
		"filename":    script.Filename,
		"group":       script.Group,
		"source":      "vault",
	})
}
