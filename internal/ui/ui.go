package ui

// #132b21 #234133 #131e22 #82543a #623d34 #31201c #240c16 #fdffdf

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Color constants (24-bit RGB ANSI escape codes)
const (
	ColorID    = "\033[38;2;136;136;136m" // #888888 - medium gray for IDs
	ColorTitle = "\033[38;2;159;212;159m" // #9fd49f - bright green for titles
	ColorTag   = "\033[38;2;97;97;97m"    // #616161 - dark gray for tags
	ColorReset = "\033[0m"                // Reset to default
)

// Terminal defaults
const (
	DefaultTermWidth = 80
)

// ============================================================================
// Password Prompting
// ============================================================================

// PromptPassword prompts the user for a password with hidden input.
func PromptPassword() (string, error) {
	fmt.Fprint(os.Stderr, "Password: ")

	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)

	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	return strings.TrimSpace(string(password)), nil
}

// PromptPasswordWithConfirmation prompts twice for password confirmation.
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

// ============================================================================
// Terminal Utilities
// ============================================================================

// GetTerminalWidth returns the current terminal width, or DefaultTermWidth if unavailable.
func GetTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return DefaultTermWidth
	}
	return width
}

// TruncateString truncates a string to maxLen runes (Unicode-safe).
func TruncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}

// ============================================================================
// Formatting Helpers
// ============================================================================

// FormatTags formats a slice of tags as "#tag1 #tag2 #tag3".
func FormatTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}

	var parts []string
	for _, tag := range tags {
		parts = append(parts, "#"+tag)
	}
	return strings.Join(parts, " ")
}

// Colorize wraps text with the given color code and resets after.
func Colorize(text, color string) string {
	return color + text + ColorReset
}
