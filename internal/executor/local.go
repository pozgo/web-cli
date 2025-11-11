package executor

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

// LocalExecutor handles execution of commands on the local machine
type LocalExecutor struct {
	// defaultTimeout for command execution (can be overridden per command)
	defaultTimeout time.Duration
}

// NewLocalExecutor creates a new local command executor
func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{
		defaultTimeout: 5 * time.Minute, // Default 5 minute timeout
	}
}

// ExecuteResult contains the result of a command execution
type ExecuteResult struct {
	Output        string
	ExitCode      int
	ExecutionTime int64 // in milliseconds
	Error         error
}

// Execute runs a command locally as the specified user
// If user is empty or "root", it runs with current process privileges
// sudoPassword is required when running as a different user (empty string for passwordless sudo)
func (e *LocalExecutor) Execute(ctx context.Context, command string, asUser string, sudoPassword string) *ExecuteResult {
	startTime := time.Now()

	// Default to root if not specified
	if asUser == "" {
		asUser = "root"
	}

	// Create context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, e.defaultTimeout)
	defer cancel()

	var cmd *exec.Cmd
	var stdout, stderr bytes.Buffer

	// Check if we need to run as a different user
	currentUser, err := user.Current()
	if err != nil {
		return &ExecuteResult{
			Output:        "",
			ExitCode:      -1,
			ExecutionTime: time.Since(startTime).Milliseconds(),
			Error:         fmt.Errorf("failed to get current user: %w", err),
		}
	}

	// If requested user is different from current user, use sudo
	if asUser != currentUser.Username && asUser != "current" {
		// Use sudo -S to read password from stdin
		// Note: This requires sudo privileges and proper sudoers configuration
		cmd = exec.CommandContext(cmdCtx, "sudo", "-S", "-u", asUser, "bash", "-c", command)

		// If password provided, set up stdin pipe
		if sudoPassword != "" {
			stdin, err := cmd.StdinPipe()
			if err != nil {
				return &ExecuteResult{
					Output:        "",
					ExitCode:      -1,
					ExecutionTime: time.Since(startTime).Milliseconds(),
					Error:         fmt.Errorf("failed to create stdin pipe: %w", err),
				}
			}

			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			// Start the command
			if err := cmd.Start(); err != nil {
				return &ExecuteResult{
					Output:        "",
					ExitCode:      -1,
					ExecutionTime: time.Since(startTime).Milliseconds(),
					Error:         fmt.Errorf("failed to start command: %w", err),
				}
			}

			// Write password to stdin immediately
			_, err = stdin.Write([]byte(sudoPassword + "\n"))
			stdin.Close()
			if err != nil {
				return &ExecuteResult{
					Output:        "",
					ExitCode:      -1,
					ExecutionTime: time.Since(startTime).Milliseconds(),
					Error:         fmt.Errorf("failed to write password: %w", err),
				}
			}

			// Wait for command to complete
			err = cmd.Wait()
		} else {
			// No password provided, let sudo handle it (will likely fail)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
		}
	} else {
		// Run as current user
		cmd = exec.CommandContext(cmdCtx, "bash", "-c", command)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// Execute the command
		err = cmd.Run()
	}

	// Combine stdout and stderr
	output := stdout.String()
	if stderr.Len() > 0 {
		if len(output) > 0 {
			output += "\n"
		}
		output += stderr.String()
	}

	executionTime := time.Since(startTime).Milliseconds()

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// Command failed to start or other error
			exitCode = -1
			// Include error in output if not already there
			if !strings.Contains(output, err.Error()) {
				if len(output) > 0 {
					output += "\n"
				}
				output += fmt.Sprintf("Error: %v", err)
			}
		}
	}

	return &ExecuteResult{
		Output:        output,
		ExitCode:      exitCode,
		ExecutionTime: executionTime,
		Error:         err,
	}
}

// ExecuteWithTimeout runs a command with a custom timeout
func (e *LocalExecutor) ExecuteWithTimeout(ctx context.Context, command string, asUser string, sudoPassword string, timeout time.Duration) *ExecuteResult {
	oldTimeout := e.defaultTimeout
	e.defaultTimeout = timeout
	defer func() { e.defaultTimeout = oldTimeout }()

	return e.Execute(ctx, command, asUser, sudoPassword)
}

// ValidateUser checks if a user exists on the system
func ValidateUser(username string) error {
	if username == "" || username == "root" || username == "current" {
		return nil
	}

	_, err := user.Lookup(username)
	if err != nil {
		return fmt.Errorf("user '%s' not found: %w", username, err)
	}

	return nil
}
