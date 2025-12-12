//go:build ignore

package main

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
)

func main() {
	// Show current state
	fmt.Printf("Session cache exists: %v\n", crypto.HasActiveSession())
	cached, _ := crypto.GetCachedPassword()
	if cached != "" {
		fmt.Printf("Cached password: %q (len=%d)\n", cached, len(cached))
	}

	// Prompt for password
	fmt.Print("\nEnter password: ")
	pw, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()

	password := string(pw)
	fmt.Printf("You entered: %q (len=%d, bytes=%v)\n", password, len(password), []byte(password))

	// Try verification directly
	fmt.Println("\nVerifying against .dredge-key...")
	err := crypto.VerifyPassword(password)
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)
	} else {
		fmt.Println("SUCCESS!")
	}
}
