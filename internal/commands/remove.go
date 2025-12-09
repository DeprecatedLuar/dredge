package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/storage"
	"github.com/DeprecatedLuar/dredge/internal/ui"
)

func HandleRemove(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dredge rm <id> [<id>...]")
	}

	// Get password once for all items
	password, err := crypto.GetPasswordWithVerification()
	if err != nil {
		return fmt.Errorf("password error: %w", err)
	}

	// Remove each item
	for _, id := range args {
		// Check if item exists
		exists, err := storage.ItemExists(id)
		if err != nil {
			return fmt.Errorf("failed to check item [%s]: %w", id, err)
		}
		if !exists {
			return fmt.Errorf("item [%s] not found", id)
		}

		// Read item to display title
		item, err := storage.ReadItem(id, password)
		if err != nil {
			return fmt.Errorf("failed to read item [%s]: %w", id, err)
		}

		// Move to trash
		if err := storage.MoveToTrash(id); err != nil {
			return fmt.Errorf("failed to move item [%s] to trash: %w", id, err)
		}

		fmt.Println(ui.FormatItem(id, item.Title, nil, "it"))
	}

	return nil
}
