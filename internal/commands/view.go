package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/storage"
	"github.com/DeprecatedLuar/dredge/internal/ui"
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

	// Format tags
	tagStr := ui.FormatTags(item.Tags)

	// Print [ID] Title #tags (same format as search)
	if len(tagStr) > 0 {
		fmt.Printf("%s[%s]%s %s%s%s %s%s%s\n", ui.ColorID, id, ui.ColorReset, ui.ColorTitle, item.Title, ui.ColorReset, ui.ColorTag, tagStr, ui.ColorReset)
	} else {
		fmt.Printf("%s[%s]%s %s%s%s\n", ui.ColorID, id, ui.ColorReset, ui.ColorTitle, item.Title, ui.ColorReset)
	}

	// Print content (cream)
	if item.Content.Text != "" {
		fmt.Println()
		fmt.Printf("%s%s%s\n", ui.ColorContent, item.Content.Text, ui.ColorReset)
	}

	return nil
}
