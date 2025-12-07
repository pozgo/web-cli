package database

import (
	"runtime"
	"testing"
)

func TestCheckEntropyAvailable(t *testing.T) {
	// This test verifies that entropy check runs without error
	// On Linux, it checks /proc/sys/kernel/random/entropy_avail
	// On other systems, it should return nil immediately
	err := checkEntropyAvailable()
	if err != nil {
		// Only fail on Linux where we expect entropy check to work
		if runtime.GOOS == "linux" {
			t.Errorf("checkEntropyAvailable() failed on Linux: %v", err)
		}
	}
}

func TestCheckEntropyAvailable_Platform(t *testing.T) {
	// Test that non-Linux platforms return nil
	if runtime.GOOS != "linux" {
		err := checkEntropyAvailable()
		if err != nil {
			t.Errorf("checkEntropyAvailable() should return nil on %s, got: %v", runtime.GOOS, err)
		}
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	// Setup encryption
	tmpDir := t.TempDir()
	keyPath := tmpDir + "/.encryption_key"

	if err := InitializeEncryption(keyPath); err != nil {
		t.Fatalf("Failed to initialize encryption: %v", err)
	}

	// Test decryption with various invalid inputs
	testCases := []struct {
		name       string
		ciphertext []byte
		shouldFail bool
	}{
		{
			name:       "empty ciphertext",
			ciphertext: []byte{},
			shouldFail: true,
		},
		{
			name:       "too short ciphertext",
			ciphertext: []byte{1, 2, 3},
			shouldFail: true,
		},
		{
			name:       "corrupted ciphertext",
			ciphertext: make([]byte, 50), // Random bytes won't decrypt
			shouldFail: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Decrypt(tc.ciphertext)
			if tc.shouldFail && err == nil {
				t.Errorf("Decrypt() should have failed for %s", tc.name)
			}
		})
	}
}

func TestPasswordHashingAndVerification(t *testing.T) {
	testCases := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "password123",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!#$%^&*()",
		},
		{
			name:     "unicode password",
			password: "пароль密码🔐",
		},
		{
			name:     "empty password",
			password: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := HashPassword(tc.password)
			if err != nil {
				t.Fatalf("HashPassword() failed: %v", err)
			}

			// Hash should be different from password
			if hash == tc.password {
				t.Error("Hash should be different from password")
			}

			// Verify correct password
			err = VerifyPassword(tc.password, hash)
			if err != nil {
				t.Errorf("VerifyPassword() failed for correct password: %v", err)
			}

			// Verify incorrect password fails
			err = VerifyPassword(tc.password+"wrong", hash)
			if err == nil {
				t.Error("VerifyPassword() should fail for incorrect password")
			}
		})
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	password := "testpassword123"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	// Same password should produce different hashes (bcrypt uses random salt)
	if hash1 == hash2 {
		t.Error("bcrypt should produce different hashes for same password due to random salt")
	}

	// Both hashes should verify against the same password
	if err := VerifyPassword(password, hash1); err != nil {
		t.Errorf("VerifyPassword() failed for hash1: %v", err)
	}
	if err := VerifyPassword(password, hash2); err != nil {
		t.Errorf("VerifyPassword() failed for hash2: %v", err)
	}
}
