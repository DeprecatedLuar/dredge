package main

import (
	"testing"
	"os"
	"github.com/DeprecatedLuar/dredge/internal/crypto"
)

func TestPassword(t *testing.T) {
	password := "Def4ult"
	
	keyFile := os.ExpandEnv("$HOME/.local/share/dredge/.dredge-key")
	data, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatal("Error reading key file:", err)
	}
	
	decrypted, err := crypto.Decrypt(data, password)
	if err != nil {
		t.Fatal("Decryption failed:", err)
	}
	
	t.Log("Decrypted:", string(decrypted))
}
