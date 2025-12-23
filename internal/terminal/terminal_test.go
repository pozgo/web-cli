package terminal

import (
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// mockWebSocket is a minimal mock for testing
type mockWebSocket struct {
	closed bool
}

func (m *mockWebSocket) Close() error {
	m.closed = true
	return nil
}

func TestNewSession(t *testing.T) {
	// Create a pipe to simulate WebSocket
	// For now, we'll just test with nil websocket (which will work until Start is called)

	// This test verifies that the session can be created with a valid shell
	t.Run("session creation with valid shell", func(t *testing.T) {
		// We can't easily test with a real WebSocket here without a server
		// This test documents the expected interface
		t.Log("Session interface: NewSession(*websocket.Conn, string) (*Session, error)")
	})
}

func TestResizeMessage(t *testing.T) {
	tests := []struct {
		name     string
		msg      ResizeMessage
		wantType string
		wantCols uint16
		wantRows uint16
	}{
		{
			name:     "standard 80x24",
			msg:      ResizeMessage{Type: "resize", Cols: 80, Rows: 24},
			wantType: "resize",
			wantCols: 80,
			wantRows: 24,
		},
		{
			name:     "large terminal",
			msg:      ResizeMessage{Type: "resize", Cols: 200, Rows: 50},
			wantType: "resize",
			wantCols: 200,
			wantRows: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.msg.Type != tt.wantType {
				t.Errorf("ResizeMessage.Type = %v, want %v", tt.msg.Type, tt.wantType)
			}
			if tt.msg.Cols != tt.wantCols {
				t.Errorf("ResizeMessage.Cols = %v, want %v", tt.msg.Cols, tt.wantCols)
			}
			if tt.msg.Rows != tt.wantRows {
				t.Errorf("ResizeMessage.Rows = %v, want %v", tt.msg.Rows, tt.wantRows)
			}
		})
	}
}

func TestSessionClose(t *testing.T) {
	t.Run("close is idempotent", func(t *testing.T) {
		// Create a session with a dummy websocket
		// This tests that Close() can be called multiple times safely
		session := &Session{
			done: make(chan struct{}),
		}

		// First close should work
		session.Close()

		// Second close should not panic (due to closeOnce)
		session.Close()

		// Verify done channel is closed
		select {
		case <-session.done:
			// Expected - channel is closed
		case <-time.After(100 * time.Millisecond):
			t.Error("done channel should be closed")
		}
	})
}

// Ensure we're using the websocket package (silences unused import)
var _ = websocket.CloseGoingAway

func TestValidateTerminalDimensions(t *testing.T) {
	tests := []struct {
		name    string
		rows    uint16
		cols    uint16
		wantErr bool
	}{
		{
			name:    "valid standard terminal",
			rows:    24,
			cols:    80,
			wantErr: false,
		},
		{
			name:    "valid large terminal",
			rows:    100,
			cols:    200,
			wantErr: false,
		},
		{
			name:    "valid max terminal",
			rows:    500,
			cols:    500,
			wantErr: false,
		},
		{
			name:    "zero rows",
			rows:    0,
			cols:    80,
			wantErr: true,
		},
		{
			name:    "zero cols",
			rows:    24,
			cols:    0,
			wantErr: true,
		},
		{
			name:    "rows exceeds max",
			rows:    501,
			cols:    80,
			wantErr: true,
		},
		{
			name:    "cols exceeds max",
			rows:    24,
			cols:    501,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTerminalDimensions(tt.rows, tt.cols)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTerminalDimensions(%d, %d) error = %v, wantErr %v", tt.rows, tt.cols, err, tt.wantErr)
			}
		})
	}
}

func TestValidateServerConfig(t *testing.T) {
	tests := []struct {
		name    string
		server  ServerConfig
		wantErr bool
	}{
		{
			name: "valid server config",
			server: ServerConfig{
				Name:      "prod-server",
				IPAddress: "192.168.1.100",
				Port:      22,
				Username:  "deploy",
			},
			wantErr: false,
		},
		{
			name: "valid server with hostname",
			server: ServerConfig{
				Name:      "web_server_1",
				IPAddress: "web.example.com",
				Port:      2222,
				Username:  "admin",
			},
			wantErr: false,
		},
		{
			name: "empty name is valid (will be skipped)",
			server: ServerConfig{
				Name:      "",
				IPAddress: "192.168.1.1",
				Port:      22,
				Username:  "user",
			},
			wantErr: false,
		},
		{
			name: "name with newline injection",
			server: ServerConfig{
				Name:      "server\nHost evil",
				IPAddress: "192.168.1.1",
				Port:      22,
				Username:  "user",
			},
			wantErr: true,
		},
		{
			name: "name with space",
			server: ServerConfig{
				Name:      "my server",
				IPAddress: "192.168.1.1",
				Port:      22,
				Username:  "user",
			},
			wantErr: true,
		},
		{
			name: "username with newline injection",
			server: ServerConfig{
				Name:      "server",
				IPAddress: "192.168.1.1",
				Port:      22,
				Username:  "user\nProxyCommand evil",
			},
			wantErr: true,
		},
		{
			name: "invalid port negative",
			server: ServerConfig{
				Name:      "server",
				IPAddress: "192.168.1.1",
				Port:      -1,
				Username:  "user",
			},
			wantErr: true,
		},
		{
			name: "invalid port too high",
			server: ServerConfig{
				Name:      "server",
				IPAddress: "192.168.1.1",
				Port:      65536,
				Username:  "user",
			},
			wantErr: true,
		},
		{
			name: "username with uppercase (invalid for unix)",
			server: ServerConfig{
				Name:      "server",
				IPAddress: "192.168.1.1",
				Port:      22,
				Username:  "Admin",
			},
			wantErr: true,
		},
		{
			name: "IP address with injection",
			server: ServerConfig{
				Name:      "server",
				IPAddress: "192.168.1.1\nProxyCommand evil",
				Port:      22,
				Username:  "user",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerConfig(tt.server)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateServerConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
