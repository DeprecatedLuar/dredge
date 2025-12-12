//go:build ignore

package main

import (
	"fmt"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/ui"
)

func main() {
	fmt.Println("=== Mimicking exact dredge flow ===")
	fmt.Println()

	// Step 1: Check session (like Before hook)
	fmt.Printf("1. HasActiveSession: %v\n", crypto.HasActiveSession())

	// Step 2: Call GetPasswordWithVerification (like HandleView does)
	fmt.Println("2. Calling GetPasswordWithVerification()...")
	fmt.Println()

	password, err := crypto.GetPasswordWithVerification()
	if err != nil {
		fmt.Printf("FAILED: %v\n", err)

		// Debug: try prompting directly and verifying
		fmt.Println("\n=== Debug: trying direct verification ===")
		fmt.Print("Enter password again: ")
		pw, _ := ui.PromptPassword()
		fmt.Printf("Got password: %q (len=%d)\n", pw, len(pw))

		fmt.Println("Verifying directly...")
		if verr := crypto.VerifyPassword(pw); verr != nil {
			fmt.Printf("Direct verify FAILED: %v\n", verr)
		} else {
			fmt.Println("Direct verify SUCCESS!")
		}
		return
	}

	fmt.Printf("SUCCESS! Password: %q\n", password)
}
