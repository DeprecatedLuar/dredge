package commands

import (
	"fmt"
	"sort"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/storage"
	"github.com/DeprecatedLuar/dredge/internal/ui"
)

func HandleList(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: dredge list")
	}

	// Get password
	password, err := crypto.GetPasswordWithVerification()
	if err != nil {
		return fmt.Errorf("password error: %w", err)
	}

	// Load all item IDs
	ids, err := storage.ListItemIDs()
	if err != nil {
		return fmt.Errorf("failed to list items: %w", err)
	}

	if len(ids) == 0 {
		fmt.Println("No items found. Use 'dredge add' to create one.")
		return nil
	}

	// Load and decrypt all items
	type itemEntry struct {
		id   string
		item *storage.Item
	}

	entries := make([]itemEntry, 0, len(ids))
	for _, id := range ids {
		item, err := storage.ReadItem(id, password)
		if err != nil {
			// Skip items that fail to decrypt
			continue
		}
		entries = append(entries, itemEntry{id: id, item: item})
	}

	// Sort by modification time (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].item.Modified.After(entries[j].item.Modified)
	})

	// Print all items
	for _, entry := range entries {
		fmt.Println(ui.FormatItem(entry.id, entry.item.Title, entry.item.Tags, "it#"))
	}

	return nil
}
