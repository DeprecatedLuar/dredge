package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/dredge/internal/storage"
)

func HandleRemove(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: dredge rm <id>")
	}

	id := args[0]

	exists, err := storage.ItemExists(id)
	if err != nil {
		return fmt.Errorf("failed to check item existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("item '%s' not found", id)
	}

	if err := storage.DeleteItem(id); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	fmt.Printf("Deleted item '%s'\n", id)
	return nil
}
