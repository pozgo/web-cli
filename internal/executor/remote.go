package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// RemoteExecutor handles execution of commands on remote machines via SSH
type RemoteExecutor struct {
	defaultTimeout time.Duration
}

// NewRemoteExecutor creates a new remote command executor
func NewRemoteExecutor() *RemoteExecutor {
	return &RemoteExecutor{
		defaultTimeout: 5 * time.Minute, // Default 5 minute timeout
	}
}

// SSHConfig holds SSH connection configuration
type SSHConfig struct {
	Host       string // hostname or IP address
	Port       int    // SSH port (default 22)
	Username   string // SSH username
	PrivateKey string // PEM-encoded private key (optional)
	Password   string // SSH password (optional, used if key auth fails)
}

// Execute runs a command on a remote server via SSH
// It tries key-based authentication first, then falls back to password if provided
func (e *RemoteExecutor) Execute(ctx context.Context, command string, config *SSHConfig) *ExecuteResult {
	startTime := time.Now()

	// Create context with timeout
	cmdCtx, cancel := context.WithTimeout(ctx, e.defaultTimeout)
	defer cancel()

	// Prepare SSH client configuration
	sshConfig := &ssh.ClientConfig{
		User:            config.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: In production, verify host keys
		Timeout:         10 * time.Second,
		Auth:            []ssh.AuthMethod{},
	}

	// Try private key authentication first if key is provided
	if config.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
		if err == nil {
			sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
		} else {
			// Log key parse error but continue to try password
			fmt.Printf("Warning: Failed to parse private key: %v\n", err)
		}
	}

	// Add password authentication if provided
	if config.Password != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.Password(config.Password))
	}

	// If no auth methods provided, return error
	if len(sshConfig.Auth) == 0 {
		return &ExecuteResult{
			Output:        "",
			ExitCode:      -1,
			ExecutionTime: time.Since(startTime).Milliseconds(),
			Error:         fmt.Errorf("no authentication method provided (need private key or password)"),
		}
	}

	// Connect to remote server
	address := fmt.Sprintf("%s:%d", config.Host, config.Port)

	// Use context-aware dialing
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}

	conn, err := dialer.DialContext(cmdCtx, "tcp", address)
	if err != nil {
		return &ExecuteResult{
			Output:        "",
			ExitCode:      -1,
			ExecutionTime: time.Since(startTime).Milliseconds(),
			Error:         fmt.Errorf("failed to connect to %s: %w", address, err),
		}
	}
	defer conn.Close()

	// Upgrade connection to SSH
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, address, sshConfig)
	if err != nil {
		return &ExecuteResult{
			Output:        "",
			ExitCode:      -1,
			ExecutionTime: time.Since(startTime).Milliseconds(),
			Error:         fmt.Errorf("SSH authentication failed: %w", err),
		}
	}
	defer sshConn.Close()

	client := ssh.NewClient(sshConn, chans, reqs)
	defer client.Close()

	// Create a session
	session, err := client.NewSession()
	if err != nil {
		return &ExecuteResult{
			Output:        "",
			ExitCode:      -1,
			ExecutionTime: time.Since(startTime).Milliseconds(),
			Error:         fmt.Errorf("failed to create SSH session: %w", err),
		}
	}
	defer session.Close()

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Execute command with context monitoring
	errChan := make(chan error, 1)
	go func() {
		errChan <- session.Run(command)
	}()

	// Wait for command completion or timeout
	var cmdErr error
	select {
	case <-cmdCtx.Done():
		// Timeout or cancellation
		session.Signal(ssh.SIGKILL)
		session.Close()
		cmdErr = fmt.Errorf("command execution timeout or cancelled")
	case cmdErr = <-errChan:
		// Command completed
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
	if cmdErr != nil {
		if exitErr, ok := cmdErr.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else if cmdErr != nil {
			// SSH connection error or other error
			exitCode = -1
			if len(output) > 0 {
				output += "\n"
			}
			output += fmt.Sprintf("Error: %v", cmdErr)
		}
	}

	return &ExecuteResult{
		Output:        output,
		ExitCode:      exitCode,
		ExecutionTime: executionTime,
		Error:         cmdErr,
	}
}

// ExecuteWithTimeout runs a remote command with a custom timeout
func (e *RemoteExecutor) ExecuteWithTimeout(ctx context.Context, command string, config *SSHConfig, timeout time.Duration) *ExecuteResult {
	oldTimeout := e.defaultTimeout
	e.defaultTimeout = timeout
	defer func() { e.defaultTimeout = oldTimeout }()

	return e.Execute(ctx, command, config)
}

// ExecuteWithStreaming runs a command and streams output in real-time
// Returns a channel that will receive output chunks as they arrive
func (e *RemoteExecutor) ExecuteWithStreaming(ctx context.Context, command string, config *SSHConfig) (<-chan string, <-chan *ExecuteResult) {
	outputChan := make(chan string, 10)
	resultChan := make(chan *ExecuteResult, 1)

	go func() {
		defer close(outputChan)
		defer close(resultChan)

		startTime := time.Now()

		// Prepare SSH client configuration (same as Execute)
		sshConfig := &ssh.ClientConfig{
			User:            config.Username,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         10 * time.Second,
			Auth:            []ssh.AuthMethod{},
		}

		if config.PrivateKey != "" {
			signer, err := ssh.ParsePrivateKey([]byte(config.PrivateKey))
			if err == nil {
				sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
			}
		}

		if config.Password != "" {
			sshConfig.Auth = append(sshConfig.Auth, ssh.Password(config.Password))
		}

		if len(sshConfig.Auth) == 0 {
			resultChan <- &ExecuteResult{
				Output:        "",
				ExitCode:      -1,
				ExecutionTime: time.Since(startTime).Milliseconds(),
				Error:         fmt.Errorf("no authentication method provided"),
			}
			return
		}

		// Connect to remote server
		address := fmt.Sprintf("%s:%d", config.Host, config.Port)
		client, err := ssh.Dial("tcp", address, sshConfig)
		if err != nil {
			resultChan <- &ExecuteResult{
				Output:        "",
				ExitCode:      -1,
				ExecutionTime: time.Since(startTime).Milliseconds(),
				Error:         fmt.Errorf("SSH connection failed: %w", err),
			}
			return
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			resultChan <- &ExecuteResult{
				Output:        "",
				ExitCode:      -1,
				ExecutionTime: time.Since(startTime).Milliseconds(),
				Error:         fmt.Errorf("failed to create session: %w", err),
			}
			return
		}
		defer session.Close()

		// Set up pipes for streaming output
		stdoutPipe, err := session.StdoutPipe()
		if err != nil {
			resultChan <- &ExecuteResult{
				Output:        "",
				ExitCode:      -1,
				ExecutionTime: time.Since(startTime).Milliseconds(),
				Error:         fmt.Errorf("failed to create stdout pipe: %w", err),
			}
			return
		}

		stderrPipe, err := session.StderrPipe()
		if err != nil {
			resultChan <- &ExecuteResult{
				Output:        "",
				ExitCode:      -1,
				ExecutionTime: time.Since(startTime).Milliseconds(),
				Error:         fmt.Errorf("failed to create stderr pipe: %w", err),
			}
			return
		}

		// Start the command
		if err := session.Start(command); err != nil {
			resultChan <- &ExecuteResult{
				Output:        "",
				ExitCode:      -1,
				ExecutionTime: time.Since(startTime).Milliseconds(),
				Error:         fmt.Errorf("failed to start command: %w", err),
			}
			return
		}

		// Read and stream output
		var fullOutput bytes.Buffer
		outputDone := make(chan bool)

		// Stream stdout
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := stdoutPipe.Read(buf)
				if n > 0 {
					chunk := string(buf[:n])
					outputChan <- chunk
					fullOutput.WriteString(chunk)
				}
				if err == io.EOF || err != nil {
					break
				}
			}
			outputDone <- true
		}()

		// Stream stderr
		go func() {
			buf := make([]byte, 1024)
			for {
				n, err := stderrPipe.Read(buf)
				if n > 0 {
					chunk := string(buf[:n])
					outputChan <- chunk
					fullOutput.WriteString(chunk)
				}
				if err == io.EOF || err != nil {
					break
				}
			}
			outputDone <- true
		}()

		// Wait for output streams to complete
		<-outputDone
		<-outputDone

		// Wait for command to complete
		cmdErr := session.Wait()

		executionTime := time.Since(startTime).Milliseconds()

		exitCode := 0
		if cmdErr != nil {
			if exitErr, ok := cmdErr.(*ssh.ExitError); ok {
				exitCode = exitErr.ExitStatus()
			} else {
				exitCode = -1
			}
		}

		resultChan <- &ExecuteResult{
			Output:        fullOutput.String(),
			ExitCode:      exitCode,
			ExecutionTime: executionTime,
			Error:         cmdErr,
		}
	}()

	return outputChan, resultChan
}
