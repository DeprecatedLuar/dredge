package crypto

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	// Password verification file name
	PasswordVerifyFile = ".dredge-key"

	// Content to encrypt in verification file
	VerificationContent = "dredge-vault-v1"
)

// GetVerifyFilePath returns the full path to the password verification file
func GetVerifyFilePath() (string, error) {
	// Use XDG_DATA_HOME or default to ~/.local/share
	baseDir := os.Getenv("XDG_DATA_HOME")
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, ".local", "share")
	}

	dredgeDir := filepath.Join(baseDir, "dredge")
	return filepath.Join(dredgeDir, PasswordVerifyFile), nil
}

// PasswordVerificationExists checks if the .dredge-key file exists
func PasswordVerificationExists() bool {
	path, err := GetVerifyFilePath()
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}

// CreatePasswordVerification creates the .dredge-key file with the given password
func CreatePasswordVerification(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	path, err := GetVerifyFilePath()
	if err != nil {
		return fmt.Errorf("failed to get verify file path: %w", err)
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Encrypt the verification content
	encrypted, err := Encrypt([]byte(VerificationContent), password)
	if err != nil {
		return fmt.Errorf("failed to encrypt verification data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write verification file: %w", err)
	}

	return nil
}

// VerifyPassword attempts to decrypt .dredge-key with the given password
// Returns nil if password is correct, error otherwise
func VerifyPassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	path, err := GetVerifyFilePath()
	if err != nil {
		return fmt.Errorf("failed to get verify file path: %w", err)
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("password verification file not found (run 'dredge add' to create vault)")
	}

	// Read encrypted data
	encrypted, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read verification file: %w", err)
	}

	// Try to decrypt WITHOUT using session cache
	// We need to verify THIS specific password, not fall back to cache
	// So we temporarily clear cache, decrypt, then restore if needed
	cachedPassword, _ := GetCachedPassword()
	_ = ClearSession() // Clear cache temporarily

	decrypted, err := Decrypt(encrypted, password)

	// Restore cache if there was one (and it's different from test password)
	if cachedPassword != "" && cachedPassword != password {
		_ = CachePassword(cachedPassword)
	}

	if err != nil {
		// Decryption failed = wrong password
		return fmt.Errorf("wrong password")
	}

	// Verify content matches expected value
	if string(decrypted) != VerificationContent {
		return fmt.Errorf("verification file corrupted (expected %q, got %q)", VerificationContent, string(decrypted))
	}

	return nil
}

// GetPasswordWithVerification prompts for password and verifies it against .dredge-key
// If .dredge-key doesn't exist, creates it with the entered password
func GetPasswordWithVerification() (string, error) {
	// Check session cache first
	cached, err := GetCachedPassword()
	if err != nil {
		return "", fmt.Errorf("failed to check password cache: %w", err)
	}

	if cached != "" {
		// Verify cached password is still correct
		if err := VerifyPassword(cached); err != nil {
			// Cache is invalid, clear it
			_ = ClearSession()
		} else {
			return cached, nil
		}
	}

	// No valid cached password, prompt user
	password, err := PromptPassword()
	if err != nil {
		return "", fmt.Errorf("failed to prompt for password: %w", err)
	}

	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	// Check if verification file exists
	if !PasswordVerificationExists() {
		// First time - create verification file
		if err := CreatePasswordVerification(password); err != nil {
			return "", fmt.Errorf("failed to create password verification: %w", err)
		}
		fmt.Fprintln(os.Stderr, "Created password verification file")
	} else {
		// Verify password
		if err := VerifyPassword(password); err != nil {
			return "", err
		}
	}

	// Cache the verified password
	if err := CachePassword(password); err != nil {
		// Non-fatal: just warn
		fmt.Fprintf(os.Stderr, "Warning: failed to cache password: %v\n", err)
	}

	return password, nil
}
