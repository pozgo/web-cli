package models

import "time"

// SavedCommand represents a command template that can be reused
// Users can save commands with descriptions for easy execution later
type SavedCommand struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`        // Friendly name for the command
	Command     string    `json:"command"`     // The actual command to execute
	Description string    `json:"description"` // Optional description
	User        string    `json:"user"`        // User to run as (default: root)
	IsRemote    bool      `json:"is_remote"`   // True if this is a remote command
	ServerID    *int64    `json:"server_id"`   // Foreign key to servers table (for remote commands)
	SSHKeyID    *int64    `json:"ssh_key_id"`  // Foreign key to ssh_keys table (for remote commands)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SavedCommandCreate represents the data needed to create a new saved command
type SavedCommandCreate struct {
	Name        string `json:"name" validate:"required"`
	Command     string `json:"command" validate:"required"`
	Description string `json:"description,omitempty"`
	User        string `json:"user"`       // Optional, defaults to "root"
	IsRemote    bool   `json:"is_remote"`  // True if this is a remote command
	ServerID    *int64 `json:"server_id"`  // For remote commands
	SSHKeyID    *int64 `json:"ssh_key_id"` // For remote commands
}

// SavedCommandUpdate represents the data that can be updated for a saved command
type SavedCommandUpdate struct {
	Name        string `json:"name,omitempty"`
	Command     string `json:"command,omitempty"`
	Description string `json:"description,omitempty"`
	User        string `json:"user,omitempty"`
	IsRemote    *bool  `json:"is_remote,omitempty"`
	ServerID    *int64 `json:"server_id,omitempty"`
	SSHKeyID    *int64 `json:"ssh_key_id,omitempty"`
}

// CommandExecution represents a request to execute a command
type CommandExecution struct {
	Command      string `json:"command" validate:"required"` // Command to execute
	User         string `json:"user"`                        // User to run as (default: root)
	SudoPassword string `json:"sudo_password,omitempty"`     // Sudo password (required when user != current for local)
	SSHPassword  string `json:"ssh_password,omitempty"`      // SSH password (for remote, if key auth fails)
	SaveAs       string `json:"save_as,omitempty"`           // Optional: save as template with this name
	IsRemote     bool   `json:"is_remote"`                   // True if remote execution
	ServerID     *int64 `json:"server_id,omitempty"`         // Server ID for remote execution (SQLite)
	ServerName   string `json:"server_name,omitempty"`       // Server name for remote execution (Vault)
	ServerGroup  string `json:"server_group,omitempty"`      // Server group for remote execution (Vault)
	SSHKeyID     *int64 `json:"ssh_key_id,omitempty"`        // SSH key ID for remote execution (SQLite)
	SSHKeyName   string `json:"ssh_key_name,omitempty"`      // SSH key name for remote execution (Vault)
	SSHKeyGroup  string `json:"ssh_key_group,omitempty"`     // SSH key group for remote execution (Vault)
}

// CommandResult represents the result of a command execution
type CommandResult struct {
	Command       string `json:"command"`
	Output        string `json:"output"`
	ExitCode      int    `json:"exit_code"`
	User          string `json:"user"`
	ExecutionTime int64  `json:"execution_time_ms"` // Execution time in milliseconds
	ExecutedAt    string `json:"executed_at"`
}

// ScriptExecution represents a request to execute a stored bash script
type ScriptExecution struct {
	ScriptID       int64    `json:"script_id,omitempty"`      // ID of the script to execute (SQLite)
	ScriptName     string   `json:"script_name,omitempty"`    // Name of the script to execute (Vault)
	ScriptGroup    string   `json:"script_group,omitempty"`   // Script group for execution (Vault)
	User           string   `json:"user"`                     // User to run as (default: root)
	SudoPassword   string   `json:"sudo_password,omitempty"`  // Sudo password (required when user != current for local)
	SSHPassword    string   `json:"ssh_password,omitempty"`   // SSH password (for remote, if key auth fails)
	IsRemote       bool     `json:"is_remote"`                // True if remote execution
	ServerID       *int64   `json:"server_id,omitempty"`      // Server ID for remote execution (SQLite)
	ServerName     string   `json:"server_name,omitempty"`    // Server name for remote execution (Vault)
	ServerGroup    string   `json:"server_group,omitempty"`   // Server group for remote execution (Vault)
	SSHKeyID       *int64   `json:"ssh_key_id,omitempty"`     // SSH key ID for remote execution (SQLite)
	SSHKeyName     string   `json:"ssh_key_name,omitempty"`   // SSH key name for remote execution (Vault)
	SSHKeyGroup    string   `json:"ssh_key_group,omitempty"`  // SSH key group for remote execution (Vault)
	IncludeEnvVars bool     `json:"include_env_vars"`         // Deprecated: use EnvVarIDs instead
	EnvVarIDs      []int64  `json:"env_var_ids,omitempty"`    // Specific env var IDs to include (SQLite)
	EnvVarNames    []string `json:"env_var_names,omitempty"`  // Names of env vars to include (Vault)
	EnvVarGroups   []string `json:"env_var_groups,omitempty"` // Groups of env vars to include (Vault, paired with EnvVarNames)
}

// ScriptResult represents the result of a script execution
type ScriptResult struct {
	ScriptID      int64  `json:"script_id"`
	ScriptName    string `json:"script_name"`
	Output        string `json:"output"`
	ExitCode      int    `json:"exit_code"`
	User          string `json:"user"`
	Server        string `json:"server"`            // "local" or server name
	ExecutionTime int64  `json:"execution_time_ms"` // Execution time in milliseconds
	EnvVarsCount  int    `json:"env_vars_injected"` // Number of env vars injected
}
