package commands

import (
	"fmt"
	"strings"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/storage"
)

const (
	// Colors (matching search.go)
	viewIdColor    = "\033[38;2;136;136;136m"  // #888888 - medium gray for IDs
	viewTitleColor = "\033[38;2;159;212;159m"  // #9fd49f - bright green for titles
	viewTagColor   = "\033[38;2;97;97;97m"     // #616161 - neutral dark gray for tags
	viewResetColor = "\033[0m"                 // Reset color
)

func HandleView(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dredge view <id>")
	}

	id := args[0]

	// Get password with verification
	password, err := crypto.GetPasswordWithVerification()
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	// Read and decrypt item
	item, err := storage.ReadItem(id, password)
	if err != nil {
		return fmt.Errorf("failed to read item: %w", err)
	}

	// Format title with tags (same as search results)
	var tagParts []string
	for _, tag := range item.Tags {
		tagParts = append(tagParts, "#"+tag)
	}
	tagStr := strings.Join(tagParts, " ")

	// Print [ID] Title #tags (same format as search)
	if len(tagStr) > 0 {
		fmt.Printf("%s[%s]%s %s%s%s %s%s%s\n", viewIdColor, id, viewResetColor, viewTitleColor, item.Title, viewResetColor, viewTagColor, tagStr, viewResetColor)
	} else {
		fmt.Printf("%s[%s]%s %s%s%s\n", viewIdColor, id, viewResetColor, viewTitleColor, item.Title, viewResetColor)
	}

	// Print content (green)
	if item.Content.Text != "" {
		fmt.Println()
		fmt.Printf("%s%s%s\n", viewTitleColor, item.Content.Text, viewResetColor)
	}

	return nil
}
