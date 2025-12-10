package commands

import (
	"fmt"
	"strings"

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

	// Print [ID] Title #tags (use <ID> for file items)
	line := ui.FormatItem(id, item.Title, item.Tags, "it#")
	if item.Type == storage.TypeFile {
		// Replace [id] with <id> for file items
		line = strings.Replace(line, "["+id+"]", "<"+id+">", 1)
	}
	fmt.Println(line)
	fmt.Println()

	// For file items, show metadata instead of base64
	if item.Type == storage.TypeFile {
		fmt.Printf("Type: file\n")
		if item.Filename != "" {
			fmt.Printf("Filename: %s\n", item.Filename)
		}
		if item.Size != nil {
			fmt.Printf("Size: %d bytes (%.2f KB)\n", *item.Size, float64(*item.Size)/1024.0)
		}
		fmt.Printf("\nUse 'dredge export %s [path]' to extract this file.\n", id)
	} else {
		// For text items, show content
		if item.Content.Text != "" {
			fmt.Println(item.Content.Text)
		}
	}

	return nil
}
