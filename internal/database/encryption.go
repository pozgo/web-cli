package database

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var encryptionKey []byte

// checkEntropyAvailable verifies sufficient system entropy before key generation
// On Linux, checks /proc/sys/kernel/random/entropy_avail
// Returns error if entropy is critically low (< 128 bits)
func checkEntropyAvailable() error {
	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/proc/sys/kernel/random/entropy_avail")
		if err != nil {
			// If we can't read entropy file, log warning but continue
			log.Printf("Warning: unable to check system entropy: %v", err)
			return nil
		}
		entropy, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			log.Printf("Warning: unable to parse entropy value: %v", err)
			return nil
		}
		if entropy < 128 {
			return fmt.Errorf("insufficient system entropy: %d bits (minimum 128 required)", entropy)
		}
		if entropy < 256 {
			log.Printf("Warning: low system entropy: %d bits (recommend >= 256)", entropy)
		}
	default:
		// macOS uses /dev/urandom backed by Yarrow/Fortuna CSPRNG
		// Windows uses CryptGenRandom backed by system CSPRNG
		// Both are considered cryptographically secure
	}
	return nil
}

// InitializeEncryption initializes the encryption key
// If ENCRYPTION_KEY environment variable is set, it uses that
// Otherwise, it generates a random key and stores it in .encryption_key file
func InitializeEncryption(keyPath string) error {
	// Try to get key from environment
	if envKey := os.Getenv("ENCRYPTION_KEY"); envKey != "" {
		decoded, err := base64.StdEncoding.DecodeString(envKey)
		if err == nil && len(decoded) == 32 {
			encryptionKey = decoded
			return nil
		}
	}

	// Try to load from file
	if data, err := os.ReadFile(keyPath); err == nil {
		decoded, err := base64.StdEncoding.DecodeString(string(data))
		if err == nil && len(decoded) == 32 {
			encryptionKey = decoded
			return nil
		}
	}

	// Check system entropy before generating new key
	if err := checkEntropyAvailable(); err != nil {
		return fmt.Errorf("entropy check failed: %w", err)
	}

	// Generate new key
	log.Println("Generating new encryption key...")
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Save key to file
	encoded := base64.StdEncoding.EncodeToString(key)
	if err := os.WriteFile(keyPath, []byte(encoded), 0600); err != nil {
		return fmt.Errorf("failed to save encryption key: %w", err)
	}

	encryptionKey = key
	return nil
}

// Encrypt encrypts data using AES-256-GCM
func Encrypt(plaintext string) ([]byte, error) {
	if encryptionKey == nil {
		return nil, fmt.Errorf("encryption key not initialized")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-256-GCM
// Includes detailed logging for audit purposes (without exposing sensitive data)
func Decrypt(ciphertext []byte) (string, error) {
	if encryptionKey == nil {
		log.Println("Decryption failed: encryption key not initialized")
		return "", fmt.Errorf("encryption key not initialized")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		log.Printf("Decryption failed: cipher creation error (key length: %d)", len(encryptionKey))
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Println("Decryption failed: GCM mode initialization error")
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		log.Printf("Decryption failed: ciphertext too short (length: %d, required: >= %d)", len(ciphertext), nonceSize)
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		// Log detailed info for auditing without exposing sensitive data
		log.Printf("Decryption failed: authentication/integrity check failed (ciphertext length: %d bytes)", len(ciphertext))
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// HashPassword creates a SHA-256 hash of a password (for future authentication)
// HashPassword creates a bcrypt hash of a password
// Uses bcrypt with cost factor 12 for enhanced security against brute-force attacks
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword compares a password against a bcrypt hash
// Returns nil if the password matches, error otherwise
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
