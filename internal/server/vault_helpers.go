package server

import (
	"context"
	"log"
	"time"

	"github.com/pozgo/web-cli/internal/models"
	"github.com/pozgo/web-cli/internal/repository"
	"github.com/pozgo/web-cli/internal/vault"
)

// getVaultClientIfEnabled returns a Vault client if Vault is configured and enabled
// Returns nil if Vault is not available (no error)
func (s *Server) getVaultClientIfEnabled() *vault.Client {
	repo := repository.NewVaultConfigRepository(s.db)
	cfg, err := repo.Get()
	if err != nil || cfg == nil || !cfg.Enabled {
		return nil
	}

	vaultCfg := &vault.Config{
		Address:   cfg.Address,
		Token:     cfg.Token,
		Namespace: cfg.Namespace,
		MountPath: cfg.MountPath,
	}

	client, err := vault.NewClient(vaultCfg)
	if err != nil {
		log.Printf("Warning: Failed to create Vault client: %v", err)
		return nil
	}

	return client
}

// mergeSSHKeysWithVault combines SQLite SSH keys with Vault SSH keys
func (s *Server) mergeSSHKeysWithVault(ctx context.Context, sqliteKeys []*models.SSHKey) []*models.SSHKey {
	// Mark SQLite keys
	for _, k := range sqliteKeys {
		k.Source = "sqlite"
	}

	// Try to get Vault client
	client := s.getVaultClientIfEnabled()
	if client == nil {
		return sqliteKeys
	}

	// Set timeout for Vault operations
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get keys from Vault
	vaultKeys, err := client.ListSSHKeys(ctx)
	if err != nil {
		log.Printf("Warning: Failed to list Vault SSH keys: %v", err)
		return sqliteKeys
	}

	// Convert Vault keys to models and append
	allKeys := make([]*models.SSHKey, 0, len(sqliteKeys)+len(vaultKeys))
	allKeys = append(allKeys, sqliteKeys...)

	for _, vk := range vaultKeys {
		allKeys = append(allKeys, &models.SSHKey{
			ID:         0, // Vault keys don't have numeric IDs
			Name:       vk.Name,
			PrivateKey: vk.PrivateKey,
			Group:      vk.Group,
			Source:     "vault",
			CreatedAt:  vk.CreatedAt,
			UpdatedAt:  vk.CreatedAt,
		})
	}

	return allKeys
}

// mergeServersWithVault combines SQLite servers with Vault servers
func (s *Server) mergeServersWithVault(ctx context.Context, sqliteServers []*models.Server) []*models.Server {
	// Mark SQLite servers
	for _, srv := range sqliteServers {
		srv.Source = "sqlite"
	}

	// Try to get Vault client
	client := s.getVaultClientIfEnabled()
	if client == nil {
		return sqliteServers
	}

	// Set timeout for Vault operations
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get servers from Vault
	vaultServers, err := client.ListServers(ctx)
	if err != nil {
		log.Printf("Warning: Failed to list Vault servers: %v", err)
		return sqliteServers
	}

	// Convert Vault servers to models and append
	allServers := make([]*models.Server, 0, len(sqliteServers)+len(vaultServers))
	allServers = append(allServers, sqliteServers...)

	now := time.Now()
	for _, vs := range vaultServers {
		allServers = append(allServers, &models.Server{
			ID:        0, // Vault servers don't have numeric IDs
			Name:      vs.Name,
			IPAddress: vs.IPAddress,
			Port:      vs.Port,
			Username:  vs.Username,
			Group:     vs.Group,
			Source:    "vault",
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	return allServers
}

// mergeEnvVariablesWithVault combines SQLite env variables with Vault env variables
func (s *Server) mergeEnvVariablesWithVault(ctx context.Context, sqliteVars []*models.EnvVariable) []*models.EnvVariable {
	// Mark SQLite vars
	for _, v := range sqliteVars {
		v.Source = "sqlite"
	}

	// Try to get Vault client
	client := s.getVaultClientIfEnabled()
	if client == nil {
		return sqliteVars
	}

	// Set timeout for Vault operations
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get env vars from Vault
	vaultVars, err := client.ListEnvVariables(ctx)
	if err != nil {
		log.Printf("Warning: Failed to list Vault env variables: %v", err)
		return sqliteVars
	}

	// Convert Vault vars to models and append
	allVars := make([]*models.EnvVariable, 0, len(sqliteVars)+len(vaultVars))
	allVars = append(allVars, sqliteVars...)

	now := time.Now()
	for _, vv := range vaultVars {
		allVars = append(allVars, &models.EnvVariable{
			ID:          0, // Vault vars don't have numeric IDs
			Name:        vv.Name,
			Value:       vv.Value,
			Description: vv.Description,
			Group:       vv.Group,
			Source:      "vault",
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	return allVars
}

// getSSHKeyByNameFromVault retrieves an SSH key from Vault by name and group
func (s *Server) getSSHKeyByNameFromVault(ctx context.Context, group, name string) (*models.SSHKey, error) {
	client := s.getVaultClientIfEnabled()
	if client == nil {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	vk, err := client.GetSSHKey(ctx, group, name)
	if err != nil {
		return nil, err
	}
	if vk == nil {
		return nil, nil
	}

	return &models.SSHKey{
		ID:         0,
		Name:       vk.Name,
		PrivateKey: vk.PrivateKey,
		Group:      vk.Group,
		Source:     "vault",
		CreatedAt:  vk.CreatedAt,
		UpdatedAt:  vk.CreatedAt,
	}, nil
}

// getServerByNameFromVault retrieves a server from Vault by name and group
func (s *Server) getServerByNameFromVault(ctx context.Context, group, name string) (*models.Server, error) {
	client := s.getVaultClientIfEnabled()
	if client == nil {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	vs, err := client.GetServer(ctx, group, name)
	if err != nil {
		return nil, err
	}
	if vs == nil {
		return nil, nil
	}

	now := time.Now()
	return &models.Server{
		ID:        0,
		Name:      vs.Name,
		IPAddress: vs.IPAddress,
		Port:      vs.Port,
		Username:  vs.Username,
		Group:     vs.Group,
		Source:    "vault",
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// getEnvVariableByNameFromVault retrieves an env variable from Vault by name and group
func (s *Server) getEnvVariableByNameFromVault(ctx context.Context, group, name string) (*models.EnvVariable, error) {
	client := s.getVaultClientIfEnabled()
	if client == nil {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	vv, err := client.GetEnvVariable(ctx, group, name)
	if err != nil {
		return nil, err
	}
	if vv == nil {
		return nil, nil
	}

	now := time.Now()
	return &models.EnvVariable{
		ID:          0,
		Name:        vv.Name,
		Value:       vv.Value,
		Description: vv.Description,
		Group:       vv.Group,
		Source:      "vault",
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// mergeScriptsWithVault combines SQLite scripts with Vault scripts
func (s *Server) mergeScriptsWithVault(ctx context.Context, sqliteScripts []*models.BashScript) []*models.BashScript {
	// Mark SQLite scripts
	for _, script := range sqliteScripts {
		script.Source = "sqlite"
	}

	// Try to get Vault client
	client := s.getVaultClientIfEnabled()
	if client == nil {
		return sqliteScripts
	}

	// Set timeout for Vault operations
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Get scripts from Vault
	vaultScripts, err := client.ListBashScripts(ctx)
	if err != nil {
		log.Printf("Warning: Failed to list Vault scripts: %v", err)
		return sqliteScripts
	}

	// Convert Vault scripts to models and append
	allScripts := make([]*models.BashScript, 0, len(sqliteScripts)+len(vaultScripts))
	allScripts = append(allScripts, sqliteScripts...)

	now := time.Now()
	for _, vs := range vaultScripts {
		allScripts = append(allScripts, &models.BashScript{
			ID:          0, // Vault scripts don't have numeric IDs
			Name:        vs.Name,
			Description: vs.Description,
			Content:     vs.Content,
			Filename:    vs.Filename,
			Group:       vs.Group,
			Source:      "vault",
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	return allScripts
}

// getScriptByNameFromVault retrieves a script from Vault by name and group
func (s *Server) getScriptByNameFromVault(ctx context.Context, group, name string) (*models.BashScript, error) {
	client := s.getVaultClientIfEnabled()
	if client == nil {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	vs, err := client.GetBashScript(ctx, group, name)
	if err != nil {
		return nil, err
	}
	if vs == nil {
		return nil, nil
	}

	now := time.Now()
	return &models.BashScript{
		ID:          0,
		Name:        vs.Name,
		Description: vs.Description,
		Content:     vs.Content,
		Filename:    vs.Filename,
		Group:       vs.Group,
		Source:      "vault",
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}
