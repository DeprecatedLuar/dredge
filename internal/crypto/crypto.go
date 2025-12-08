package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// Encryption constants
const (
	SaltSize  = 16 // 128 bits
	NonceSize = 12 // 96 bits (standard GCM nonce size)

	// Argon2id parameters (per RFC 9106 recommendations)
	Argon2Time      = 1         // 1 iteration
	Argon2Memory    = 64 * 1024 // 64 MB
	Argon2Threads   = 4         // 4 parallel threads
	Argon2KeyLength = 32        // 32 bytes for AES-256
)

// Encrypt encrypts plaintext using password-derived key (Argon2id + AES-256-GCM).
// Returns binary format: [16B salt][12B nonce][N bytes ciphertext + 16B auth tag]
func Encrypt(plaintext []byte, password string) ([]byte, error) {
	if password == "" {
		return nil, fmt.Errorf("password cannot be empty")
	}

	// Generate random salt
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from password using Argon2id
	key := argon2.IDKey(
		[]byte(password),
		salt,
		Argon2Time,
		Argon2Memory,
		Argon2Threads,
		Argon2KeyLength,
	)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Construct final format: salt || nonce || ciphertext
	result := make([]byte, 0, SaltSize+NonceSize+len(ciphertext))
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// Decrypt decrypts encrypted data using password.
// Checks session cache first, prompts for password if needed.
// Input format: [16B salt][12B nonce][N bytes ciphertext + 16B auth tag]
func Decrypt(encrypted []byte, password string) ([]byte, error) {
	// Validate minimum size: salt + nonce + auth tag
	minSize := SaltSize + NonceSize + 16 // 16 = GCM auth tag size
	if len(encrypted) < minSize {
		return nil, fmt.Errorf("encrypted data too short: got %d bytes, need at least %d", len(encrypted), minSize)
	}

	// Extract salt, nonce, and ciphertext
	salt := encrypted[:SaltSize]
	nonce := encrypted[SaltSize : SaltSize+NonceSize]
	ciphertext := encrypted[SaltSize+NonceSize:]

	// Try to get cached password first
	cachedPassword, err := GetCachedPassword()
	if err != nil {
		return nil, fmt.Errorf("failed to check session cache: %w", err)
	}

	// Use cached password if available, otherwise use provided password
	if cachedPassword != "" {
		password = cachedPassword
	} else if password == "" {
		return nil, fmt.Errorf("no cached password and no password provided")
	} else {
		// Cache the password for this session
		if err := CachePassword(password); err != nil {
			// Non-fatal: continue even if caching fails
			fmt.Fprintf(io.Discard, "warning: failed to cache password: %v\n", err)
		}
	}

	// Derive key from password + file's unique salt
	key := argon2.IDKey(
		[]byte(password),
		salt,
		Argon2Time,
		Argon2Memory,
		Argon2Threads,
		Argon2KeyLength,
	)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt and verify authentication tag
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed (wrong password or tampered data): %w", err)
	}

	return plaintext, nil
}

// DeriveKey derives an encryption key from password and salt using Argon2id.
// Used for manual key derivation when needed.
func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		Argon2Time,
		Argon2Memory,
		Argon2Threads,
		Argon2KeyLength,
	)
}
