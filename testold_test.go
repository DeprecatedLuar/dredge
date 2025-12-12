package main

import (
	"testing"
	"os"
	"github.com/DeprecatedLuar/dredge/internal/crypto"
)

func TestOldNrg(t *testing.T) {
	password := "Def4ult?"
	
	data, err := os.ReadFile("/tmp/nrg_original.enc")
	if err != nil {
		t.Fatal("Error reading file:", err)
	}
	
	decrypted, err := crypto.Decrypt(data, password)
	if err != nil {
		t.Fatal("Decryption failed:", err)
	}
	
	t.Log("Original TOML:\n", string(decrypted))
}
