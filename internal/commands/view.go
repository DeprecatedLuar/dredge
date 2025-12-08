package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/storage"
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

	// Print title, blank line, then content
	fmt.Println(item.Title)
	if item.Content.Text != "" {
		fmt.Println()
		fmt.Println(item.Content.Text)
	}

	return nil
}
