package commands

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/storage"
)

func HandleExport(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dredge export <id|number> [output-path]")
	}

	// Resolve ID from first argument (supports numbered access)
	ids, err := ResolveArgs([]string{args[0]})
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return fmt.Errorf("no item found")
	}

	id := ids[0]

	// Get password and read item
	password, err := crypto.GetPasswordWithVerification()
	if err != nil {
		return err
	}

	item, err := storage.ReadItem(id, password)
	if err != nil {
		return fmt.Errorf("failed to read item: %w", err)
	}

	// Validate item type
	if item.Type != storage.TypeFile {
		return fmt.Errorf("item %s is not a file item (type: %s)", id, item.Type)
	}

	// Determine output path
	var outputPath string
	if len(args) >= 2 {
		outputPath = args[1]

		// Make path absolute if relative
		if !filepath.IsAbs(outputPath) {
			absPath, err := filepath.Abs(outputPath)
			if err != nil {
				return fmt.Errorf("failed to resolve output path: %w", err)
			}
			outputPath = absPath
		}

		// If output path is a directory, append original filename
		if stat, err := os.Stat(outputPath); err == nil && stat.IsDir() {
			if item.Filename == "" {
				return fmt.Errorf("item has no filename and output path is a directory")
			}
			outputPath = filepath.Join(outputPath, item.Filename)
		}
	} else {
		// Use original filename in current directory
		if item.Filename == "" {
			return fmt.Errorf("item has no filename and no output path provided")
		}
		outputPath = item.Filename
	}

	// Check if file already exists
	if _, err := os.Stat(outputPath); err == nil {
		return fmt.Errorf("file already exists at %s", outputPath)
	}

	// Decode base64 content
	decodedContent, err := base64.StdEncoding.DecodeString(item.Content.Text)
	if err != nil {
		return fmt.Errorf("failed to decode file content: %w", err)
	}

	// Verify size matches (if size metadata exists)
	if item.Size != nil && int64(len(decodedContent)) != *item.Size {
		return fmt.Errorf("size mismatch: expected %d bytes, got %d bytes", *item.Size, len(decodedContent))
	}

	// Write to output path
	if err := os.WriteFile(outputPath, decodedContent, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Exported [%s] %s -> %s (%d bytes)\n", id, item.Title, outputPath, len(decodedContent))
	return nil
}
