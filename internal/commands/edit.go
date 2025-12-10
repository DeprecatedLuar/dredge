package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/editor"
	"github.com/DeprecatedLuar/dredge/internal/storage"
)

func HandleEdit(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: dredge edit <id>")
	}

	// Resolve numbered arg to ID
	ids, err := ResolveArgs(args)
	if err != nil {
		return err
	}
	id := ids[0]

	// Get password
	password, err := crypto.GetPasswordWithVerification()
	if err != nil {
		return fmt.Errorf("password error: %w", err)
	}

	// Read existing item
	item, err := storage.ReadItem(id, password)
	if err != nil {
		return fmt.Errorf("failed to read item [%s]: %w", id, err)
	}

	// Open editor with existing item
	updatedItem, err := editor.OpenForExisting(item)
	if err != nil {
		return fmt.Errorf("failed to edit item: %w", err)
	}

	// Save updated item
	if err := storage.UpdateItem(id, updatedItem, password); err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	fmt.Printf("✓ [%s] %s\n", id, updatedItem.Title)
	return nil
}
