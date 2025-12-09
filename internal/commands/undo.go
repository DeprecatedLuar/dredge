package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/storage"
	"github.com/DeprecatedLuar/dredge/internal/ui"
)

func HandleUndo(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: dredge undo")
	}

	// Get last deleted ID from cache
	id, err := storage.GetLastDeletedID()
	if err != nil {
		return fmt.Errorf("cannot undo: %w", err)
	}

	// Get password to read item title
	password, err := crypto.GetPasswordWithVerification()
	if err != nil {
		return fmt.Errorf("password error: %w", err)
	}

	// Restore item from trash
	if err := storage.RestoreFromTrash(id); err != nil {
		return fmt.Errorf("failed to restore item: %w", err)
	}

	// Read item to display title in confirmation
	item, err := storage.ReadItem(id, password)
	if err != nil {
		// Non-fatal, item is already restored
		fmt.Printf("+ [%s]\n", id)
		return nil
	}

	fmt.Println("+ " + ui.FormatItem(id, item.Title, item.Tags, "it#"))
	return nil
}
