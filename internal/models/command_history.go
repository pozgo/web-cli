package models

import "time"

// CommandHistory represents a command execution record
type CommandHistory struct {
	ID              int64     `json:"id"`
	Command         string    `json:"command"`          // Decrypted value
	Output          string    `json:"output,omitempty"` // Decrypted value
	ExitCode        *int      `json:"exit_code,omitempty"`
	Server          string    `json:"server"`         // "local" for local commands, or server name/IP
	User            string    `json:"user,omitempty"` // User who executed the command (for local commands)
	ExecutionTimeMs int64     `json:"execution_time_ms,omitempty"`
	ExecutedAt      time.Time `json:"executed_at"`
}

// CommandHistoryCreate represents the data needed to create a command history record
type CommandHistoryCreate struct {
	Command         string `json:"command" validate:"required"`
	Output          string `json:"output,omitempty"`
	ExitCode        *int   `json:"exit_code,omitempty"`
	Server          string `json:"server" validate:"required"` // "local" for local commands
	User            string `json:"user,omitempty"`             // User who executed the command
	ExecutionTimeMs int64  `json:"execution_time_ms,omitempty"`
}
