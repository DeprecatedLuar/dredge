package commands

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/editor"
	"github.com/DeprecatedLuar/dredge/internal/storage"
	"github.com/DeprecatedLuar/dredge/internal/ui"
)

const (
	idLength   = 3
	maxRetries = 10
)

func generateID() (string, error) {
	bytes := make([]byte, idLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	id := base64.RawURLEncoding.EncodeToString(bytes)
	return id[:idLength], nil
}

// parseAddArgs manually parses args to extract title, content, and tags
// Supports flexible flag ordering: title can come first, -c and -t can be in any order
func parseAddArgs(args []string) (title, content string, tags []string) {
	if len(args) == 0 {
		return "", "", nil
	}

	// Find flag positions
	cPos := -1
	tPos := -1
	for i, arg := range args {
		if arg == "-c" {
			cPos = i
		} else if arg == "-t" {
			tPos = i
		}
	}

	// Extract title (everything before first flag)
	firstFlagPos := len(args)
	if cPos != -1 && tPos != -1 {
		if cPos < tPos {
			firstFlagPos = cPos
		} else {
			firstFlagPos = tPos
		}
	} else if cPos != -1 {
		firstFlagPos = cPos
	} else if tPos != -1 {
		firstFlagPos = tPos
	}
	if firstFlagPos > 0 {
		title = strings.Join(args[:firstFlagPos], " ")
	}

	// Extract content (between -c and next flag or end)
	if cPos != -1 {
		contentStart := cPos + 1
		contentEnd := len(args)
		// If -t comes after -c, content ends at -t
		if tPos != -1 && tPos > cPos {
			contentEnd = tPos
		}
		if contentStart < contentEnd {
			content = strings.Join(args[contentStart:contentEnd], " ")
		}
	}

	// Extract tags (after -t until next flag or end)
	if tPos != -1 {
		tagsStart := tPos + 1
		tagsEnd := len(args)
		// If -c comes after -t, tags end at -c
		if cPos != -1 && cPos > tPos {
			tagsEnd = cPos
		}
		if tagsStart < tagsEnd {
			tags = args[tagsStart:tagsEnd]
		}
	}

	return title, content, tags
}

func HandleAdd(args []string) error {
	// Parse args (empty args returns empty title/content/tags)
	title, content, tags := parseAddArgs(args)

	var item *storage.Item

	// If no content provided, open editor (includes empty args case)
	if content == "" {
		var err error
		item, err = editor.OpenForNewItem(title, tags)
		if err != nil {
			return fmt.Errorf("failed to create item via editor: %w", err)
		}
	} else {
		// Create item directly from CLI args
		if title == "" {
			return fmt.Errorf("title cannot be empty")
		}
		item = storage.NewTextItem(title, content, tags)
	}

	// Generate unique ID
	var id string
	var err error
	for i := 0; i < maxRetries; i++ {
		id, err = generateID()
		if err != nil {
			return fmt.Errorf("failed to generate ID: %w", err)
		}

		exists, err := storage.ItemExists(id)
		if err != nil {
			return fmt.Errorf("failed to check item existence: %w", err)
		}
		if !exists {
			break
		}

		if i == maxRetries-1 {
			return fmt.Errorf("failed to generate unique ID after %d attempts", maxRetries)
		}
	}

	// Get password with verification (checks/creates .dredge-key)
	password, err := crypto.GetPasswordWithVerification()
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	if err := storage.CreateItem(id, item, password); err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	fmt.Println("+ " + ui.FormatItem(id, item.Title, item.Tags, "it#"))
	return nil
}
