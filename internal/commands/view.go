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

	// Resolve numbered arg to ID
	ids, err := ResolveArgs(args[:1])
	if err != nil {
		return err
	}
	id := ids[0]

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

	// Print [ID] Title #tags
	fmt.Println(ui.FormatItem(id, item.Title, item.Tags, "it#"))

	// Print content
	if item.Content.Text != "" {
		fmt.Println()
		fmt.Println(item.Content.Text)
	}

	return nil
}
