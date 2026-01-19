package audit

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// EventType represents the type of audited event
type EventType string

const (
	EventTypeCommandExecution EventType = "COMMAND_EXECUTION"
	EventTypeScriptExecution  EventType = "SCRIPT_EXECUTION"
	EventTypeSSHConnection    EventType = "SSH_CONNECTION"
	EventTypeTerminalSession  EventType = "TERMINAL_SESSION"
	EventTypeConfigChange     EventType = "CONFIG_CHANGE"
	EventTypeAuthAttempt      EventType = "AUTH_ATTEMPT"
)

// EventOutcome represents the result of an audited event
type EventOutcome string

const (
	OutcomeSuccess EventOutcome = "SUCCESS"
	OutcomeFailure EventOutcome = "FAILURE"
	OutcomeDenied  EventOutcome = "DENIED"
)

// AuditEvent represents a single audit log entry
type AuditEvent struct {
	Timestamp   time.Time         `json:"timestamp"`
	EventType   EventType         `json:"event_type"`
	Outcome     EventOutcome      `json:"outcome"`
	Actor       string            `json:"actor"`       // Username or "anonymous"
	SourceIP    string            `json:"source_ip"`   // Client IP address
	Target      string            `json:"target"`      // What was acted upon (e.g., server name, script name)
	Command     string            `json:"command,omitempty"`
	User        string            `json:"user,omitempty"`         // Execution user (e.g., root, ubuntu)
	Server      string            `json:"server,omitempty"`       // Target server for remote commands
	ExitCode    *int              `json:"exit_code,omitempty"`    // Command exit code
	Duration    int64             `json:"duration_ms,omitempty"`  // Execution duration in milliseconds
	ErrorMsg    string            `json:"error,omitempty"`        // Error message if failed
	Metadata    map[string]string `json:"metadata,omitempty"`     // Additional context
}

// Logger handles audit logging
type Logger struct {
	mu       sync.Mutex
	enabled  bool
	file     *os.File
	filePath string
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Initialize creates or returns the singleton audit logger
func Initialize(filePath string) (*Logger, error) {
	var initErr error
	once.Do(func() {
		logger := &Logger{
			enabled:  filePath != "",
			filePath: filePath,
		}

		if filePath != "" {
			file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				initErr = err
				log.Printf("Warning: Failed to open audit log file %s: %v", filePath, err)
				logger.enabled = false
			} else {
				logger.file = file
			}
		}

		defaultLogger = logger
	})

	return defaultLogger, initErr
}

// GetLogger returns the default audit logger
func GetLogger() *Logger {
	if defaultLogger == nil {
		// Return a disabled logger if not initialized
		return &Logger{enabled: false}
	}
	return defaultLogger
}

// Close closes the audit log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// Log writes an audit event
func (l *Logger) Log(event *AuditEvent) {
	if !l.enabled || l.file == nil {
		return
	}

	// Ensure timestamp is set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Warning: Failed to marshal audit event: %v", err)
		return
	}

	// Append newline for JSONL format
	data = append(data, '\n')

	if _, err := l.file.Write(data); err != nil {
		log.Printf("Warning: Failed to write audit event: %v", err)
	}
}

// LogCommandExecution logs a command execution event
func (l *Logger) LogCommandExecution(r *http.Request, command, user, server string, exitCode int, durationMs int64, err error) {
	event := &AuditEvent{
		EventType: EventTypeCommandExecution,
		Actor:     getActorFromRequest(r),
		SourceIP:  getClientIP(r),
		Target:    server,
		Command:   sanitizeCommand(command),
		User:      user,
		Server:    server,
		ExitCode:  &exitCode,
		Duration:  durationMs,
	}

	if err != nil {
		event.Outcome = OutcomeFailure
		event.ErrorMsg = err.Error()
	} else if exitCode != 0 {
		event.Outcome = OutcomeFailure
	} else {
		event.Outcome = OutcomeSuccess
	}

	l.Log(event)
}

// LogScriptExecution logs a script execution event
func (l *Logger) LogScriptExecution(r *http.Request, scriptName, user, server string, exitCode int, durationMs int64, err error) {
	event := &AuditEvent{
		EventType: EventTypeScriptExecution,
		Actor:     getActorFromRequest(r),
		SourceIP:  getClientIP(r),
		Target:    scriptName,
		User:      user,
		Server:    server,
		ExitCode:  &exitCode,
		Duration:  durationMs,
	}

	if err != nil {
		event.Outcome = OutcomeFailure
		event.ErrorMsg = err.Error()
	} else if exitCode != 0 {
		event.Outcome = OutcomeFailure
	} else {
		event.Outcome = OutcomeSuccess
	}

	l.Log(event)
}

// LogTerminalSession logs a terminal session start/end
func (l *Logger) LogTerminalSession(r *http.Request, server, user string, outcome EventOutcome, metadata map[string]string) {
	event := &AuditEvent{
		EventType: EventTypeTerminalSession,
		Outcome:   outcome,
		Actor:     getActorFromRequest(r),
		SourceIP:  getClientIP(r),
		Target:    server,
		User:      user,
		Server:    server,
		Metadata:  metadata,
	}

	l.Log(event)
}

// LogAuthAttempt logs an authentication attempt
func (l *Logger) LogAuthAttempt(r *http.Request, outcome EventOutcome, method string) {
	event := &AuditEvent{
		EventType: EventTypeAuthAttempt,
		Outcome:   outcome,
		Actor:     getActorFromRequest(r),
		SourceIP:  getClientIP(r),
		Target:    r.URL.Path,
		Metadata: map[string]string{
			"method": method,
		},
	}

	l.Log(event)
}

// LogConfigChange logs a configuration change
func (l *Logger) LogConfigChange(r *http.Request, configType, action string, outcome EventOutcome) {
	event := &AuditEvent{
		EventType: EventTypeConfigChange,
		Outcome:   outcome,
		Actor:     getActorFromRequest(r),
		SourceIP:  getClientIP(r),
		Target:    configType,
		Metadata: map[string]string{
			"action": action,
		},
	}

	l.Log(event)
}

// getActorFromRequest extracts the actor (username) from the request
func getActorFromRequest(r *http.Request) string {
	if r == nil {
		return "system"
	}

	// Check for basic auth username
	if username, _, ok := r.BasicAuth(); ok && username != "" {
		return username
	}

	// Check for custom header (if using token-based auth)
	if actor := r.Header.Get("X-Auth-User"); actor != "" {
		return actor
	}

	return "anonymous"
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	// Check X-Forwarded-For header first (for proxied requests)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	remoteAddr := r.RemoteAddr
	// Strip port if present
	for i := len(remoteAddr) - 1; i >= 0; i-- {
		if remoteAddr[i] == ':' {
			return remoteAddr[:i]
		}
		if remoteAddr[i] == ']' {
			// IPv6 address with brackets
			break
		}
	}

	return remoteAddr
}

// sanitizeCommand removes potentially sensitive data from commands
// This is a basic implementation - extend as needed
func sanitizeCommand(cmd string) string {
	// For audit purposes, we log the command but could add sanitization
	// to remove obvious passwords or tokens if needed
	if len(cmd) > 1000 {
		return cmd[:1000] + "...[truncated]"
	}
	return cmd
}
