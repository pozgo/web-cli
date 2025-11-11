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
	ServerID     *int64 `json:"server_id,omitempty"`         // Server ID for remote execution
	SSHKeyID     *int64 `json:"ssh_key_id,omitempty"`        // SSH key ID for remote execution
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
