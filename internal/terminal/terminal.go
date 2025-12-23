package terminal

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
)

// ResizeMessage represents a terminal resize request from the client
type ResizeMessage struct {
	Type string `json:"type"`
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}

// ServerConfig holds server information for SSH config generation
type ServerConfig struct {
	Name      string
	IPAddress string
	Port      int
	Username  string
}


// Validation constants
const (
	maxTerminalRows = 500
	maxTerminalCols = 500
	maxHostnameLen  = 253
	maxUsernameLen  = 32
)

// validHostnamePattern matches valid hostname characters (alphanumeric, hyphen, underscore, dot)
var validHostnamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

// validUsernamePattern matches valid Unix username characters
var validUsernamePattern = regexp.MustCompile(`^[a-z_][a-z0-9_-]*$`)

// ValidateTerminalDimensions checks if terminal dimensions are within acceptable range
func ValidateTerminalDimensions(rows, cols uint16) error {
	if rows == 0 || cols == 0 {
		return fmt.Errorf("rows and cols must be greater than zero")
	}
	if rows > maxTerminalRows {
		return fmt.Errorf("rows %d exceeds maximum %d", rows, maxTerminalRows)
	}
	if cols > maxTerminalCols {
		return fmt.Errorf("cols %d exceeds maximum %d", cols, maxTerminalCols)
	}
	return nil
}

// ValidateServerConfig validates a ServerConfig for safe use in SSH config generation
func ValidateServerConfig(server ServerConfig) error {
	// Validate Name (required, used as Host alias)
	if server.Name != "" {
		if len(server.Name) > maxHostnameLen {
			return fmt.Errorf("server name exceeds maximum length of %d", maxHostnameLen)
		}
		if !validHostnamePattern.MatchString(server.Name) {
			return fmt.Errorf("server name contains invalid characters")
		}
	}

	// Validate IPAddress/Hostname
	if server.IPAddress != "" {
		if len(server.IPAddress) > maxHostnameLen {
			return fmt.Errorf("IP address/hostname exceeds maximum length of %d", maxHostnameLen)
		}
		if !validHostnamePattern.MatchString(server.IPAddress) {
			return fmt.Errorf("IP address/hostname contains invalid characters")
		}
	}

	// Validate Username
	if server.Username != "" {
		if len(server.Username) > maxUsernameLen {
			return fmt.Errorf("username exceeds maximum length of %d", maxUsernameLen)
		}
		if !validUsernamePattern.MatchString(server.Username) {
			return fmt.Errorf("username contains invalid characters")
		}
	}

	// Validate Port
	if server.Port < 0 || server.Port > 65535 {
		return fmt.Errorf("port %d is out of valid range (0-65535)", server.Port)
	}

	return nil
}

// Session manages a PTY session connected to a WebSocket
type Session struct {
	ptmx       *os.File
	cmd        *exec.Cmd
	ws         *websocket.Conn
	done       chan struct{}
	closeOnce  sync.Once
	sshKeyPath string // Path to temporary SSH key file (if any)
	tmpDir     string // Path to temporary directory for session files
}

// NewSession creates a new terminal session with the specified shell
// sshPrivateKey: if provided, will be written to a temp file and used for SSH connections
// servers: list of servers from admin panel to generate SSH config aliases
func NewSession(ws *websocket.Conn, shell string, sshPrivateKey string, servers []ServerConfig) (*Session, error) {
	cmd := exec.Command(shell)
	// Set environment with proper TERM for full terminal support
	env := append(os.Environ(), "TERM=xterm-256color")

	var sshKeyPath string
	var tmpDir string

	// Create temp directory for session files (SSH config, keys, wrapper)
	// We always create this if we have servers or SSH key
	if len(servers) > 0 || sshPrivateKey != "" {
		var err error
		tmpDir, err = os.MkdirTemp("", "webcli-ssh-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp dir: %w", err)
		}

		// Generate SSH config file with server aliases
		if len(servers) > 0 {
			sshConfigPath := filepath.Join(tmpDir, "config")
			if err := generateSSHConfig(sshConfigPath, servers); err != nil {
				os.RemoveAll(tmpDir)
				return nil, fmt.Errorf("failed to generate SSH config: %w", err)
			}
		}

		// Write SSH key if provided
		if sshPrivateKey != "" {
			// Ensure the key has a trailing newline (required by OpenSSH)
			keyContent := sshPrivateKey
			if len(keyContent) > 0 && keyContent[len(keyContent)-1] != '\n' {
				keyContent += "\n"
			}

			sshKeyPath = filepath.Join(tmpDir, "id_rsa")
			if err := os.WriteFile(sshKeyPath, []byte(keyContent), 0600); err != nil {
				os.RemoveAll(tmpDir)
				return nil, fmt.Errorf("failed to write SSH key: %w", err)
			}

			// Add environment variable pointing to the SSH key
			env = append(env, "SSH_KEY_PATH="+sshKeyPath)
		}

		// Create SSH wrapper script that uses our custom config and optionally the key
		wrapperPath := filepath.Join(tmpDir, "ssh")
		wrapperContent := generateSSHWrapper(tmpDir, sshKeyPath)
		if err := os.WriteFile(wrapperPath, []byte(wrapperContent), 0755); err != nil {
			os.RemoveAll(tmpDir)
			return nil, fmt.Errorf("failed to write SSH wrapper: %w", err)
		}

		// Prepend the temp dir to PATH so our ssh wrapper is used
		for i, e := range env {
			if len(e) > 5 && e[:5] == "PATH=" {
				env[i] = "PATH=" + tmpDir + ":" + e[5:]
				break
			}
		}
	}

	cmd.Env = env

	ptmx, err := pty.Start(cmd)
	if err != nil {
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
		return nil, err
	}

	// Set initial size (80x24 is standard)
	if err := pty.Setsize(ptmx, &pty.Winsize{Rows: 24, Cols: 80}); err != nil {
		ptmx.Close()
		cmd.Process.Kill()
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
		return nil, err
	}

	return &Session{
		ptmx:       ptmx,
		cmd:        cmd,
		ws:         ws,
		done:       make(chan struct{}),
		sshKeyPath: sshKeyPath,
		tmpDir:     tmpDir,
	}, nil
}

// Start begins bidirectional communication between WebSocket and PTY
func (s *Session) Start() {
	var wg sync.WaitGroup
	wg.Add(2)

	// PTY -> WebSocket (shell output to browser)
	go func() {
		defer wg.Done()
		buf := make([]byte, 4096)
		for {
			select {
			case <-s.done:
				return
			default:
				n, err := s.ptmx.Read(buf)
				if err != nil {
					if err != io.EOF {
						log.Printf("PTY read error: %v", err)
					}
					s.Close()
					return
				}
				if n > 0 {
					if err := s.ws.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
						log.Printf("WebSocket write error: %v", err)
						s.Close()
						return
					}
				}
			}
		}
	}()

	// WebSocket -> PTY (user input to shell)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-s.done:
				return
			default:
				messageType, message, err := s.ws.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
						log.Printf("WebSocket read error: %v", err)
					}
					s.Close()
					return
				}

				switch messageType {
				case websocket.TextMessage:
					// Check if it's a resize message
					var resizeMsg ResizeMessage
					if err := json.Unmarshal(message, &resizeMsg); err == nil && resizeMsg.Type == "resize" {
						if err := s.Resize(resizeMsg.Rows, resizeMsg.Cols); err != nil {
							log.Printf("Resize error: %v", err)
						}
					} else {
						// Regular text input
						if _, err := s.ptmx.Write(message); err != nil {
							log.Printf("PTY write error: %v", err)
							s.Close()
							return
						}
					}
				case websocket.BinaryMessage:
					// Binary data goes directly to PTY
					if _, err := s.ptmx.Write(message); err != nil {
						log.Printf("PTY write error: %v", err)
						s.Close()
						return
					}
				}
			}
		}
	}()

	// Wait for shell process to exit
	go func() {
		s.cmd.Wait()
		s.Close()
	}()

	wg.Wait()
}

// Resize changes the PTY window size
func (s *Session) Resize(rows, cols uint16) error {
	if err := ValidateTerminalDimensions(rows, cols); err != nil {
		return err
	}
	return pty.Setsize(s.ptmx, &pty.Winsize{Rows: rows, Cols: cols})
}

// Close terminates the session and cleans up resources
func (s *Session) Close() {
	s.closeOnce.Do(func() {
		close(s.done)

		if s.ptmx != nil {
			s.ptmx.Close()
		}

		if s.cmd != nil && s.cmd.Process != nil {
			s.cmd.Process.Kill()
		}

		if s.ws != nil {
			s.ws.Close()
		}

		// Clean up session temp directory
		if s.tmpDir != "" {
			os.RemoveAll(s.tmpDir)
		}
	})
}

// generateSSHConfig creates an SSH config file with server aliases
func generateSSHConfig(configPath string, servers []ServerConfig) error {
	var config strings.Builder

	config.WriteString("# Auto-generated SSH config for web-cli session\n")
	config.WriteString("# Server aliases from admin panel\n\n")

	for _, server := range servers {
		// Skip servers without a name (can't create alias)
		if server.Name == "" {
			continue
		}

		// Validate server config to prevent SSH config injection
		if err := ValidateServerConfig(server); err != nil {
			log.Printf("Skipping invalid server config %q: %v", server.Name, err)
			continue
		}

		// Determine the hostname to connect to
		hostname := server.IPAddress
		if hostname == "" {
			hostname = server.Name
		}

		config.WriteString(fmt.Sprintf("Host %s\n", server.Name))
		config.WriteString(fmt.Sprintf("    HostName %s\n", hostname))

		if server.Port > 0 && server.Port != 22 {
			config.WriteString(fmt.Sprintf("    Port %d\n", server.Port))
		}

		if server.Username != "" {
			config.WriteString(fmt.Sprintf("    User %s\n", server.Username))
		}

		config.WriteString("    StrictHostKeyChecking accept-new\n")
		config.WriteString("\n")
	}

	return os.WriteFile(configPath, []byte(config.String()), 0600)
}

// generateSSHWrapper creates an SSH wrapper script that uses our custom config and optional key
func generateSSHWrapper(tmpDir string, sshKeyPath string) string {
	configPath := filepath.Join(tmpDir, "config")

	var wrapper strings.Builder
	wrapper.WriteString("#!/bin/sh\n")

	// Build SSH command with options
	wrapper.WriteString("exec /usr/bin/ssh")

	// Always use our custom config file
	wrapper.WriteString(fmt.Sprintf(" -F \"%s\"", configPath))

	// Add identity file if provided
	if sshKeyPath != "" {
		wrapper.WriteString(fmt.Sprintf(" -i \"%s\"", sshKeyPath))
	}

	// Add default options
	wrapper.WriteString(" -o StrictHostKeyChecking=accept-new")

	// Pass through all arguments
	wrapper.WriteString(" \"$@\"\n")

	return wrapper.String()
}
