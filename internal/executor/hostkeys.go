package executor

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

// HostKeyVerifier manages SSH host key verification
type HostKeyVerifier struct {
	knownHostsPath  string
	knownHosts      map[string]ssh.PublicKey
	mu              sync.RWMutex
	trustOnFirstUse bool
}

// NewHostKeyVerifier creates a new host key verifier
// trustOnFirstUse: if true, automatically trust and save new host keys (less secure)
func NewHostKeyVerifier(knownHostsPath string, trustOnFirstUse bool) (*HostKeyVerifier, error) {
	verifier := &HostKeyVerifier{
		knownHostsPath:  knownHostsPath,
		knownHosts:      make(map[string]ssh.PublicKey),
		trustOnFirstUse: trustOnFirstUse,
	}

	// Load existing known_hosts file if it exists
	if err := verifier.loadKnownHosts(); err != nil {
		return nil, fmt.Errorf("failed to load known hosts: %w", err)
	}

	return verifier, nil
}

// VerifyHostKey verifies the host key against known_hosts
func (v *HostKeyVerifier) VerifyHostKey(hostname string, remote net.Addr, key ssh.PublicKey) error {
	v.mu.RLock()
	knownKey, exists := v.knownHosts[hostname]
	v.mu.RUnlock()

	if !exists {
		// Host key not in known_hosts
		if v.trustOnFirstUse {
			// Trust on first use - add to known_hosts
			if err := v.addHostKey(hostname, key); err != nil {
				return fmt.Errorf("failed to add host key: %w", err)
			}
			return nil
		}
		return fmt.Errorf("host key for %s not found in known_hosts (trust-on-first-use is disabled)", hostname)
	}

	// Verify the key matches
	if !keysEqual(knownKey, key) {
		return fmt.Errorf("host key mismatch for %s - possible man-in-the-middle attack!", hostname)
	}

	return nil
}

// GetHostKeyCallback returns a ssh.HostKeyCallback function
func (v *HostKeyVerifier) GetHostKeyCallback() ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		return v.VerifyHostKey(hostname, remote, key)
	}
}

// loadKnownHosts loads known_hosts file into memory
func (v *HostKeyVerifier) loadKnownHosts() error {
	// Create known_hosts file if it doesn't exist
	if _, err := os.Stat(v.knownHostsPath); os.IsNotExist(err) {
		// Ensure directory exists
		dir := filepath.Dir(v.knownHostsPath)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
		// Create empty file
		if err := os.WriteFile(v.knownHostsPath, []byte{}, 0600); err != nil {
			return err
		}
		return nil
	}

	file, err := os.Open(v.knownHostsPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse known_hosts line: hostname key-type base64-key
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue // Skip invalid lines
		}

		hostname := parts[0]
		keyType := parts[1]
		keyData := parts[2]

		// Parse the public key
		key, err := parsePublicKeyFromKnownHosts(keyType, keyData)
		if err != nil {
			continue // Skip invalid keys
		}

		v.knownHosts[hostname] = key
	}

	return scanner.Err()
}

// addHostKey adds a new host key to known_hosts
func (v *HostKeyVerifier) addHostKey(hostname string, key ssh.PublicKey) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Add to in-memory map
	v.knownHosts[hostname] = key

	// Append to file
	file, err := os.OpenFile(v.knownHostsPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	// Format: hostname key-type base64-key
	line := fmt.Sprintf("%s %s %s\n", hostname, key.Type(), ssh.MarshalAuthorizedKey(key))
	if _, err := file.WriteString(line); err != nil {
		return err
	}

	return nil
}

// parsePublicKeyFromKnownHosts parses a public key from known_hosts format
func parsePublicKeyFromKnownHosts(keyType, keyData string) (ssh.PublicKey, error) {
	// Reconstruct the authorized key format
	authorizedKey := fmt.Sprintf("%s %s", keyType, keyData)
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(authorizedKey))
	if err != nil {
		return nil, err
	}
	return key, nil
}

// keysEqual compares two SSH public keys
func keysEqual(a, b ssh.PublicKey) bool {
	return string(a.Marshal()) == string(b.Marshal())
}
