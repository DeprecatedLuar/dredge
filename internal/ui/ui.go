package ui

// #132b21 #234133 #131e22 #82543a #623d34 #31201c #240c16 #fdffdf

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Color constants
const (
	ColorTag   = "\033[38;2;128;128;128m" // Muted gray for tags
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

// FormatItem formats item components based on what parts are requested.
// parts: "i" = id, "t" = title, "#" = tags
// Examples: "it#" = [id] title #tags, "it" = [id] title, "t#" = title #tags
func FormatItem(id, title string, tags []string, parts string) string {
	var result strings.Builder

	for _, char := range parts {
		switch char {
		case 'i':
			if result.Len() > 0 {
				result.WriteString(" ")
			}
			result.WriteString("[")
			result.WriteString(id)
			result.WriteString("]")
		case 't':
			if result.Len() > 0 {
				result.WriteString(" ")
			}
			result.WriteString(title)
		case '#':
			tagStr := FormatTags(tags)
			if tagStr != "" {
				if result.Len() > 0 {
					result.WriteString(" ")
				}
				result.WriteString(ColorTag)
				result.WriteString(tagStr)
				result.WriteString(ColorReset)
			}
		}
	}

	return result.String()
}
