package vault

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/pozgo/web-cli/internal/validation"
)

// Client wraps the Vault API client with convenience methods
type Client struct {
	client    *api.Client
	mountPath string
}

// Config holds the configuration for connecting to Vault
type Config struct {
	Address   string `json:"address"`
	Token     string `json:"token"`
	Namespace string `json:"namespace,omitempty"`
	MountPath string `json:"mount_path"`
}

// NewClient creates a new Vault client with the given configuration
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("vault config is nil")
	}

	if cfg.Address == "" {
		return nil, fmt.Errorf("vault address is required")
	}

	// Validate Vault address to prevent SSRF attacks
	if err := validation.ValidateVaultAddress(cfg.Address); err != nil {
		return nil, fmt.Errorf("invalid vault address: %w", err)
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("vault token is required")
	}

	vaultCfg := api.DefaultConfig()
	vaultCfg.Address = cfg.Address

	client, err := api.NewClient(vaultCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	client.SetToken(cfg.Token)

	if cfg.Namespace != "" {
		client.SetNamespace(cfg.Namespace)
	}

	mountPath := cfg.MountPath
	if mountPath == "" {
		mountPath = "secret"
	}

	return &Client{
		client:    client,
		mountPath: mountPath,
	}, nil
}

// TestConnection verifies the Vault connection and token validity
func (c *Client) TestConnection(ctx context.Context) error {
	// Try to look up the token to verify it's valid
	_, err := c.client.Auth().Token().LookupSelfWithContext(ctx)
	if err != nil {
		return fmt.Errorf("vault connection test failed: %w", err)
	}
	return nil
}

// InitializeStructure creates the web-cli directory structure in Vault
// This ensures paths exist for listing operations
func (c *Client) InitializeStructure(ctx context.Context) error {
	// Create a .initialized marker in each path to ensure the paths exist
	// This makes listing work even for empty paths
	paths := []string{"ssh-keys", "servers", "env", "scripts"}

	for _, secretType := range paths {
		// Check if default group already has entries
		secrets, err := c.ListSecrets(ctx, secretType, "default")
		if err != nil {
			// If list fails with actual error, try to create anyway
			continue
		}

		// If path is empty, create .initialized marker
		if len(secrets) == 0 {
			data := map[string]interface{}{
				"initialized": true,
				"created_at":  time.Now().Format(time.RFC3339),
			}
			// Ignore errors - token may not have write permission
			_ = c.WriteSecret(ctx, secretType, "default", ".initialized", data)
		}
	}

	return nil
}

// GetHealth returns the Vault server health status
func (c *Client) GetHealth(ctx context.Context) (*api.HealthResponse, error) {
	return c.client.Sys().HealthWithContext(ctx)
}

// secretPath constructs the full path for a secret
// Path format: {mount}/data/{secretType}/{group}/{name}
// Example: web-cli/data/ssh-keys/production/my-key
func (c *Client) secretPath(secretType, group, name string) (string, error) {
	if group == "" {
		group = "default"
	}

	// Validate path components to prevent path traversal
	if err := validation.ValidateVaultSecretName(secretType); err != nil {
		return "", fmt.Errorf("invalid secret type: %w", err)
	}
	if err := validation.ValidateVaultGroupName(group); err != nil {
		return "", fmt.Errorf("invalid group name: %w", err)
	}
	if err := validation.ValidateVaultSecretName(name); err != nil {
		return "", fmt.Errorf("invalid secret name: %w", err)
	}

	return fmt.Sprintf("%s/data/%s/%s/%s", c.mountPath, secretType, group, name), nil
}

// secretPathFlat constructs the path for backward compatibility (no group)
// Path format: {mount}/data/{secretType}/{name}
// Example: web-cli/data/ssh-keys/my-key
func (c *Client) secretPathFlat(secretType, name string) (string, error) {
	// Validate path components to prevent path traversal
	if err := validation.ValidateVaultSecretName(secretType); err != nil {
		return "", fmt.Errorf("invalid secret type: %w", err)
	}
	if err := validation.ValidateVaultSecretName(name); err != nil {
		return "", fmt.Errorf("invalid secret name: %w", err)
	}

	return fmt.Sprintf("%s/data/%s/%s", c.mountPath, secretType, name), nil
}

// metadataPath constructs the metadata path for listing a group
// Path format: {mount}/metadata/{secretType}/{group}
// Example: web-cli/metadata/ssh-keys/production
func (c *Client) metadataPath(secretType, group string) (string, error) {
	// Validate path components to prevent path traversal
	if err := validation.ValidateVaultSecretName(secretType); err != nil {
		return "", fmt.Errorf("invalid secret type: %w", err)
	}
	if group != "" {
		if err := validation.ValidateVaultGroupName(group); err != nil {
			return "", fmt.Errorf("invalid group name: %w", err)
		}
		return fmt.Sprintf("%s/metadata/%s/%s", c.mountPath, secretType, group), nil
	}
	return fmt.Sprintf("%s/metadata/%s", c.mountPath, secretType), nil
}

// metadataPathFlat constructs the metadata path for listing all groups
// Path format: {mount}/metadata/{secretType}
// Example: web-cli/metadata/ssh-keys
func (c *Client) metadataPathFlat(secretType string) (string, error) {
	// Validate path components to prevent path traversal
	if err := validation.ValidateVaultSecretName(secretType); err != nil {
		return "", fmt.Errorf("invalid secret type: %w", err)
	}
	return fmt.Sprintf("%s/metadata/%s", c.mountPath, secretType), nil
}

// ReadSecret reads a secret from Vault (with group support)
func (c *Client) ReadSecret(ctx context.Context, secretType, group, name string) (map[string]interface{}, error) {
	path, err := c.secretPath(secretType, group, name)
	if err != nil {
		return nil, err
	}

	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret %s/%s/%s: %w", secretType, group, name, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, nil // Secret doesn't exist
	}

	// KV v2 stores data under "data" key
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected secret format for %s/%s/%s", secretType, group, name)
	}

	return data, nil
}

// ReadSecretFlat reads a secret from Vault without group (backward compatibility)
func (c *Client) ReadSecretFlat(ctx context.Context, secretType, name string) (map[string]interface{}, error) {
	path, err := c.secretPathFlat(secretType, name)
	if err != nil {
		return nil, err
	}

	secret, err := c.client.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret %s/%s: %w", secretType, name, err)
	}

	if secret == nil || secret.Data == nil {
		return nil, nil // Secret doesn't exist
	}

	// KV v2 stores data under "data" key
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected secret format for %s/%s", secretType, name)
	}

	return data, nil
}

// WriteSecret writes a secret to Vault (with group support)
func (c *Client) WriteSecret(ctx context.Context, secretType, group, name string, data map[string]interface{}) error {
	path, err := c.secretPath(secretType, group, name)
	if err != nil {
		return err
	}

	// KV v2 requires data to be wrapped
	wrappedData := map[string]interface{}{
		"data": data,
	}

	_, err = c.client.Logical().WriteWithContext(ctx, path, wrappedData)
	if err != nil {
		return fmt.Errorf("failed to write secret %s/%s/%s: %w", secretType, group, name, err)
	}

	return nil
}

// WriteSecretFlat writes a secret to Vault without group (backward compatibility)
func (c *Client) WriteSecretFlat(ctx context.Context, secretType, name string, data map[string]interface{}) error {
	path, err := c.secretPathFlat(secretType, name)
	if err != nil {
		return err
	}

	// KV v2 requires data to be wrapped
	wrappedData := map[string]interface{}{
		"data": data,
	}

	_, err = c.client.Logical().WriteWithContext(ctx, path, wrappedData)
	if err != nil {
		return fmt.Errorf("failed to write secret %s/%s: %w", secretType, name, err)
	}

	return nil
}

// DeleteSecret deletes a secret from Vault (with group support)
func (c *Client) DeleteSecret(ctx context.Context, secretType, group, name string) error {
	if group == "" {
		group = "default"
	}

	// Validate path components to prevent path traversal
	if err := validation.ValidateVaultSecretName(secretType); err != nil {
		return fmt.Errorf("invalid secret type: %w", err)
	}
	if err := validation.ValidateVaultGroupName(group); err != nil {
		return fmt.Errorf("invalid group name: %w", err)
	}
	if err := validation.ValidateVaultSecretName(name); err != nil {
		return fmt.Errorf("invalid secret name: %w", err)
	}

	// For KV v2, we delete the metadata to permanently remove
	metaPath := fmt.Sprintf("%s/metadata/%s/%s/%s", c.mountPath, secretType, group, name)

	_, err := c.client.Logical().DeleteWithContext(ctx, metaPath)
	if err != nil {
		return fmt.Errorf("failed to delete secret %s/%s/%s: %w", secretType, group, name, err)
	}

	return nil
}

// DeleteSecretFlat deletes a secret from Vault without group (backward compatibility)
func (c *Client) DeleteSecretFlat(ctx context.Context, secretType, name string) error {
	// Validate path components to prevent path traversal
	if err := validation.ValidateVaultSecretName(secretType); err != nil {
		return fmt.Errorf("invalid secret type: %w", err)
	}
	if err := validation.ValidateVaultSecretName(name); err != nil {
		return fmt.Errorf("invalid secret name: %w", err)
	}

	// For KV v2, we delete the metadata to permanently remove
	metaPath := fmt.Sprintf("%s/metadata/%s/%s", c.mountPath, secretType, name)

	_, err := c.client.Logical().DeleteWithContext(ctx, metaPath)
	if err != nil {
		return fmt.Errorf("failed to delete secret %s/%s: %w", secretType, name, err)
	}

	return nil
}

// ListSecrets lists all secrets of a given type (optionally filtered by group)
func (c *Client) ListSecrets(ctx context.Context, secretType, group string) ([]string, error) {
	path, err := c.metadataPath(secretType, group)
	if err != nil {
		return nil, err
	}

	secret, err := c.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		// Check if it's a 403 (permission denied) or 404 (path doesn't exist)
		// In both cases, return empty list - the path may not exist yet or
		// the token may not have list permissions (but might have read/write)
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "404") ||
			strings.Contains(err.Error(), "permission denied") {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list secrets for %s/%s: %w", secretType, group, err)
	}

	if secret == nil || secret.Data == nil {
		return []string{}, nil // No secrets
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	result := make([]string, 0, len(keys))
	for _, k := range keys {
		if s, ok := k.(string); ok {
			// Skip directory entries (end with /) and internal .initialized marker
			if !strings.HasSuffix(s, "/") && s != ".initialized" {
				result = append(result, s)
			}
		}
	}

	return result, nil
}

// ListGroups lists all groups for a given secret type
func (c *Client) ListGroups(ctx context.Context, secretType string) ([]string, error) {
	path, err := c.metadataPathFlat(secretType)
	if err != nil {
		return nil, err
	}

	secret, err := c.client.Logical().ListWithContext(ctx, path)
	if err != nil {
		if strings.Contains(err.Error(), "403") || strings.Contains(err.Error(), "404") ||
			strings.Contains(err.Error(), "permission denied") {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list groups for %s: %w", secretType, err)
	}

	if secret == nil || secret.Data == nil {
		return []string{}, nil
	}

	keys, ok := secret.Data["keys"].([]interface{})
	if !ok {
		return []string{}, nil
	}

	result := make([]string, 0, len(keys))
	for _, k := range keys {
		if s, ok := k.(string); ok {
			// Groups are directories (end with /)
			if strings.HasSuffix(s, "/") {
				result = append(result, strings.TrimSuffix(s, "/"))
			}
		}
	}

	return result, nil
}

// SSHKey represents an SSH key stored in Vault
type SSHKey struct {
	Name       string    `json:"name"`
	PrivateKey string    `json:"private_key"`
	Group      string    `json:"group"`
	CreatedAt  time.Time `json:"created_at"`
}

// ListSSHKeys returns all SSH keys from Vault (across all groups)
func (c *Client) ListSSHKeys(ctx context.Context) ([]SSHKey, error) {
	groups, err := c.ListGroups(ctx, "ssh-keys")
	if err != nil {
		return nil, err
	}

	// If no groups exist, check default group
	if len(groups) == 0 {
		groups = []string{"default"}
	}

	var keys []SSHKey
	for _, group := range groups {
		groupKeys, err := c.ListSSHKeysByGroup(ctx, group)
		if err != nil {
			continue
		}
		keys = append(keys, groupKeys...)
	}

	return keys, nil
}

// ListSSHKeysByGroup returns all SSH keys in a specific group
func (c *Client) ListSSHKeysByGroup(ctx context.Context, group string) ([]SSHKey, error) {
	names, err := c.ListSecrets(ctx, "ssh-keys", group)
	if err != nil {
		return nil, err
	}

	keys := make([]SSHKey, 0, len(names))
	for _, name := range names {
		key, err := c.GetSSHKey(ctx, group, name)
		if err != nil {
			continue // Skip keys that can't be read
		}
		if key != nil {
			keys = append(keys, *key)
		}
	}

	return keys, nil
}

// GetSSHKey retrieves a specific SSH key from Vault
func (c *Client) GetSSHKey(ctx context.Context, group, name string) (*SSHKey, error) {
	if group == "" {
		group = "default"
	}
	data, err := c.ReadSecret(ctx, "ssh-keys", group, name)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	key := &SSHKey{Name: name, Group: group}

	// Try standard field name first, then fall back to checking all string values
	if pk, ok := data["private_key"].(string); ok {
		key.PrivateKey = pk
	} else {
		// Flexible parsing: find any string value that looks like a private key
		for _, v := range data {
			if s, ok := v.(string); ok && strings.Contains(s, "PRIVATE KEY") {
				key.PrivateKey = s
				break
			}
		}
	}

	if ca, ok := data["created_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, ca); err == nil {
			key.CreatedAt = t
		}
	}

	return key, nil
}

// SaveSSHKey saves an SSH key to Vault
func (c *Client) SaveSSHKey(ctx context.Context, key *SSHKey) error {
	if key.Group == "" {
		key.Group = "default"
	}
	data := map[string]interface{}{
		"private_key": key.PrivateKey,
		"created_at":  key.CreatedAt.Format(time.RFC3339),
	}
	return c.WriteSecret(ctx, "ssh-keys", key.Group, key.Name, data)
}

// DeleteSSHKey deletes an SSH key from Vault
func (c *Client) DeleteSSHKey(ctx context.Context, group, name string) error {
	if group == "" {
		group = "default"
	}
	return c.DeleteSecret(ctx, "ssh-keys", group, name)
}

// ListSSHKeyGroups returns all groups for SSH keys
func (c *Client) ListSSHKeyGroups(ctx context.Context) ([]string, error) {
	return c.ListGroups(ctx, "ssh-keys")
}

// Server represents a server configuration stored in Vault
type Server struct {
	Name      string `json:"name"`
	IPAddress string `json:"ip_address"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Group     string `json:"group"`
}

// ListServers returns all servers from Vault (across all groups)
func (c *Client) ListServers(ctx context.Context) ([]Server, error) {
	groups, err := c.ListGroups(ctx, "servers")
	if err != nil {
		return nil, err
	}

	// If no groups exist, check default group
	if len(groups) == 0 {
		groups = []string{"default"}
	}

	var servers []Server
	for _, group := range groups {
		groupServers, err := c.ListServersByGroup(ctx, group)
		if err != nil {
			continue
		}
		servers = append(servers, groupServers...)
	}

	return servers, nil
}

// ListServersByGroup returns all servers in a specific group
func (c *Client) ListServersByGroup(ctx context.Context, group string) ([]Server, error) {
	names, err := c.ListSecrets(ctx, "servers", group)
	if err != nil {
		return nil, err
	}

	servers := make([]Server, 0, len(names))
	for _, name := range names {
		srv, err := c.GetServer(ctx, group, name)
		if err != nil {
			continue
		}
		if srv != nil {
			servers = append(servers, *srv)
		}
	}

	return servers, nil
}

// GetServer retrieves a specific server from Vault
func (c *Client) GetServer(ctx context.Context, group, name string) (*Server, error) {
	if group == "" {
		group = "default"
	}
	data, err := c.ReadSecret(ctx, "servers", group, name)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	srv := &Server{Name: name, Group: group}

	// Try standard field names first
	if ip, ok := data["ip_address"].(string); ok {
		srv.IPAddress = ip
	} else if ip, ok := data["ip"].(string); ok {
		srv.IPAddress = ip
	} else if ip, ok := data["host"].(string); ok {
		srv.IPAddress = ip
	} else if ip, ok := data["address"].(string); ok {
		srv.IPAddress = ip
	} else {
		// Flexible parsing: find any string value that looks like an IP or hostname
		for _, v := range data {
			if s, ok := v.(string); ok && s != "" {
				srv.IPAddress = s
				break
			}
		}
	}

	if port, ok := data["port"].(float64); ok {
		srv.Port = int(port)
	} else if port, ok := data["port"].(int); ok {
		srv.Port = port
	} else {
		srv.Port = 22 // Default port
	}

	if user, ok := data["username"].(string); ok {
		srv.Username = user
	} else if user, ok := data["user"].(string); ok {
		srv.Username = user
	}

	return srv, nil
}

// SaveServer saves a server to Vault
func (c *Client) SaveServer(ctx context.Context, srv *Server) error {
	if srv.Group == "" {
		srv.Group = "default"
	}
	data := map[string]interface{}{
		"ip_address": srv.IPAddress,
		"port":       srv.Port,
		"username":   srv.Username,
	}
	return c.WriteSecret(ctx, "servers", srv.Group, srv.Name, data)
}

// DeleteServer deletes a server from Vault
func (c *Client) DeleteServer(ctx context.Context, group, name string) error {
	if group == "" {
		group = "default"
	}
	return c.DeleteSecret(ctx, "servers", group, name)
}

// ListServerGroups returns all groups for servers
func (c *Client) ListServerGroups(ctx context.Context) ([]string, error) {
	return c.ListGroups(ctx, "servers")
}

// EnvVariable represents an environment variable stored in Vault
type EnvVariable struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Group       string `json:"group"`
}

// ListEnvVariables returns all environment variables from Vault (across all groups)
func (c *Client) ListEnvVariables(ctx context.Context) ([]EnvVariable, error) {
	groups, err := c.ListGroups(ctx, "env")
	if err != nil {
		return nil, err
	}

	// If no groups exist, check default group
	if len(groups) == 0 {
		groups = []string{"default"}
	}

	var vars []EnvVariable
	for _, group := range groups {
		groupVars, err := c.ListEnvVariablesByGroup(ctx, group)
		if err != nil {
			continue
		}
		vars = append(vars, groupVars...)
	}

	return vars, nil
}

// ListEnvVariablesByGroup returns all environment variables in a specific group
func (c *Client) ListEnvVariablesByGroup(ctx context.Context, group string) ([]EnvVariable, error) {
	names, err := c.ListSecrets(ctx, "env", group)
	if err != nil {
		return nil, err
	}

	vars := make([]EnvVariable, 0, len(names))
	for _, name := range names {
		v, err := c.GetEnvVariable(ctx, group, name)
		if err != nil {
			continue
		}
		if v != nil {
			vars = append(vars, *v)
		}
	}

	return vars, nil
}

// GetEnvVariable retrieves a specific environment variable from Vault
func (c *Client) GetEnvVariable(ctx context.Context, group, name string) (*EnvVariable, error) {
	if group == "" {
		group = "default"
	}
	data, err := c.ReadSecret(ctx, "env", group, name)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	v := &EnvVariable{Name: name, Group: group}

	if val, ok := data["value"].(string); ok {
		v.Value = val
	}

	if desc, ok := data["description"].(string); ok {
		v.Description = desc
	}

	return v, nil
}

// SaveEnvVariable saves an environment variable to Vault
func (c *Client) SaveEnvVariable(ctx context.Context, v *EnvVariable) error {
	if v.Group == "" {
		v.Group = "default"
	}
	data := map[string]interface{}{
		"value":       v.Value,
		"description": v.Description,
	}
	return c.WriteSecret(ctx, "env", v.Group, v.Name, data)
}

// DeleteEnvVariable deletes an environment variable from Vault
func (c *Client) DeleteEnvVariable(ctx context.Context, group, name string) error {
	if group == "" {
		group = "default"
	}
	return c.DeleteSecret(ctx, "env", group, name)
}

// ListEnvVariableGroups returns all groups for environment variables
func (c *Client) ListEnvVariableGroups(ctx context.Context) ([]string, error) {
	return c.ListGroups(ctx, "env")
}

// BashScript represents a bash script stored in Vault
type BashScript struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Filename    string `json:"filename"`
	Group       string `json:"group"`
}

// ListBashScripts returns all bash scripts from Vault (across all groups)
func (c *Client) ListBashScripts(ctx context.Context) ([]BashScript, error) {
	groups, err := c.ListGroups(ctx, "scripts")
	if err != nil {
		return nil, err
	}

	// If no groups exist, check default group
	if len(groups) == 0 {
		groups = []string{"default"}
	}

	var scripts []BashScript
	for _, group := range groups {
		groupScripts, err := c.ListBashScriptsByGroup(ctx, group)
		if err != nil {
			continue
		}
		scripts = append(scripts, groupScripts...)
	}

	return scripts, nil
}

// ListBashScriptsByGroup returns all bash scripts in a specific group
func (c *Client) ListBashScriptsByGroup(ctx context.Context, group string) ([]BashScript, error) {
	names, err := c.ListSecrets(ctx, "scripts", group)
	if err != nil {
		return nil, err
	}

	scripts := make([]BashScript, 0, len(names))
	for _, name := range names {
		s, err := c.GetBashScript(ctx, group, name)
		if err != nil {
			continue
		}
		if s != nil {
			scripts = append(scripts, *s)
		}
	}

	return scripts, nil
}

// GetBashScript retrieves a specific bash script from Vault
func (c *Client) GetBashScript(ctx context.Context, group, name string) (*BashScript, error) {
	if group == "" {
		group = "default"
	}
	data, err := c.ReadSecret(ctx, "scripts", group, name)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	s := &BashScript{Name: name, Group: group}

	if content, ok := data["content"].(string); ok {
		s.Content = content
	}

	if desc, ok := data["description"].(string); ok {
		s.Description = desc
	}

	if fn, ok := data["filename"].(string); ok {
		s.Filename = fn
	}

	return s, nil
}

// SaveBashScript saves a bash script to Vault
func (c *Client) SaveBashScript(ctx context.Context, s *BashScript) error {
	if s.Group == "" {
		s.Group = "default"
	}
	data := map[string]interface{}{
		"content":     s.Content,
		"description": s.Description,
		"filename":    s.Filename,
	}
	return c.WriteSecret(ctx, "scripts", s.Group, s.Name, data)
}

// DeleteBashScript deletes a bash script from Vault
func (c *Client) DeleteBashScript(ctx context.Context, group, name string) error {
	if group == "" {
		group = "default"
	}
	return c.DeleteSecret(ctx, "scripts", group, name)
}

// ListBashScriptGroups returns all groups for bash scripts
func (c *Client) ListBashScriptGroups(ctx context.Context) ([]string, error) {
	return c.ListGroups(ctx, "scripts")
}
