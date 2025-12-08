package crypto

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// PromptPassword prompts the user for a password with hidden input.
// Returns the password string (newline trimmed).
func PromptPassword() (string, error) {
	fmt.Fprint(os.Stderr, "Password: ")

	// Read password with hidden input
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr) // Print newline after password input

	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	return strings.TrimSpace(string(password)), nil
}

// PromptPasswordWithConfirmation prompts twice for password confirmation.
// Used when setting/changing passwords.
func PromptPasswordWithConfirmation() (string, error) {
	fmt.Fprint(os.Stderr, "Enter password: ")
	password1, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	fmt.Fprint(os.Stderr, "Confirm password: ")
	password2, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("failed to read password confirmation: %w", err)
	}

	pwd1 := strings.TrimSpace(string(password1))
	pwd2 := strings.TrimSpace(string(password2))

	if pwd1 != pwd2 {
		return "", fmt.Errorf("passwords do not match")
	}

	if pwd1 == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	return pwd1, nil
}
