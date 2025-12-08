package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	password := "test-password-123"
	plaintext := []byte("This is a secret message that should be encrypted.")

	// Encrypt
	encrypted, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Verify encrypted data is longer (salt + nonce + ciphertext + auth tag)
	minExpectedSize := SaltSize + NonceSize + len(plaintext) + 16 // 16 = GCM auth tag
	if len(encrypted) < minExpectedSize {
		t.Errorf("Encrypted data too short: got %d bytes, expected at least %d", len(encrypted), minExpectedSize)
	}

	// Verify encrypted data is different from plaintext
	if bytes.Contains(encrypted, plaintext) {
		t.Error("Encrypted data contains plaintext (encryption failed)")
	}

	// Decrypt
	decrypted, err := Decrypt(encrypted, password)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	// Verify decrypted matches original
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted data doesn't match original.\nGot:  %q\nWant: %q", decrypted, plaintext)
	}
}

func TestDecrypt_WrongPassword(t *testing.T) {
	password := "correct-password"
	wrongPassword := "wrong-password"
	plaintext := []byte("Secret data")

	// Encrypt with correct password
	encrypted, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Try to decrypt with wrong password
	_, err = Decrypt(encrypted, wrongPassword)
	if err == nil {
		t.Error("Decrypt should fail with wrong password, but succeeded")
	}
}

func TestEncrypt_EmptyPassword(t *testing.T) {
	plaintext := []byte("Some data")

	_, err := Encrypt(plaintext, "")
	if err == nil {
		t.Error("Encrypt should fail with empty password, but succeeded")
	}
}

func TestDecrypt_TamperedData(t *testing.T) {
	password := "test-password"
	plaintext := []byte("Original secret")

	// Encrypt
	encrypted, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Tamper with the ciphertext (flip a bit in the middle)
	tampered := make([]byte, len(encrypted))
	copy(tampered, encrypted)
	midpoint := len(tampered) / 2
	tampered[midpoint] ^= 0xFF // Flip all bits in one byte

	// Try to decrypt tampered data
	_, err = Decrypt(tampered, password)
	if err == nil {
		t.Error("Decrypt should fail with tampered data (GCM auth should fail), but succeeded")
	}
}

func TestDecrypt_TooShort(t *testing.T) {
	password := "test-password"
	tooShort := []byte("short") // Less than salt + nonce + auth tag

	_, err := Decrypt(tooShort, password)
	if err == nil {
		t.Error("Decrypt should fail with data too short, but succeeded")
	}
}

func TestEncrypt_UniqueSalts(t *testing.T) {
	// Clear session to ensure clean state
	_ = ClearSession()

	password := "same-password"
	plaintext := []byte("Same plaintext")

	// Encrypt the same data twice
	encrypted1, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("First encrypt failed: %v", err)
	}

	encrypted2, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Second encrypt failed: %v", err)
	}

	// Salts should be different (first 16 bytes)
	salt1 := encrypted1[:SaltSize]
	salt2 := encrypted2[:SaltSize]

	if bytes.Equal(salt1, salt2) {
		t.Error("Salts should be unique for each encryption, but they're identical")
	}

	// Ciphertexts should be different
	if bytes.Equal(encrypted1, encrypted2) {
		t.Error("Encrypted outputs should differ (due to unique salts/nonces), but they're identical")
	}

	// Both should decrypt correctly
	decrypted1, err := Decrypt(encrypted1, password)
	if err != nil || !bytes.Equal(decrypted1, plaintext) {
		t.Error("First encrypted data failed to decrypt correctly")
	}

	decrypted2, err := Decrypt(encrypted2, password)
	if err != nil || !bytes.Equal(decrypted2, plaintext) {
		t.Error("Second encrypted data failed to decrypt correctly")
	}
}

func TestEncryptDecrypt_EmptyPlaintext(t *testing.T) {
	// Clear session to ensure clean state
	_ = ClearSession()

	password := "test-password"
	plaintext := []byte{}

	encrypted, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt empty plaintext failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, password)
	if err != nil {
		t.Fatalf("Decrypt empty plaintext failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("Decrypted empty plaintext doesn't match.\nGot:  %q\nWant: %q", decrypted, plaintext)
	}
}

func TestEncryptDecrypt_LargeData(t *testing.T) {
	// Clear session to ensure clean state
	_ = ClearSession()

	password := "test-password"
	// Create 1MB of data
	plaintext := make([]byte, 1024*1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	encrypted, err := Encrypt(plaintext, password)
	if err != nil {
		t.Fatalf("Encrypt large data failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, password)
	if err != nil {
		t.Fatalf("Decrypt large data failed: %v", err)
	}

	if !bytes.Equal(decrypted, plaintext) {
		t.Error("Decrypted large data doesn't match original")
	}
}

func TestDeriveKey(t *testing.T) {
	password := "test-password"
	salt := []byte("16-byte-salt-val") // 16 bytes

	// Derive key twice with same inputs
	key1 := DeriveKey(password, salt)
	key2 := DeriveKey(password, salt)

	// Should be identical
	if !bytes.Equal(key1, key2) {
		t.Error("DeriveKey should produce identical keys for same inputs")
	}

	// Should be 32 bytes (AES-256)
	if len(key1) != Argon2KeyLength {
		t.Errorf("Derived key wrong length: got %d, want %d", len(key1), Argon2KeyLength)
	}

	// Different salt should produce different key
	differentSalt := []byte("different-salt-!")
	key3 := DeriveKey(password, differentSalt)

	if bytes.Equal(key1, key3) {
		t.Error("DeriveKey should produce different keys for different salts")
	}
}
