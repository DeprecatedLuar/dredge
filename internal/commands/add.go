package commands

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/storage"
)

const (
	idLength      = 3
	maxRetries    = 10
)

func generateID() (string, error) {
	bytes := make([]byte, idLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	id := base64.RawURLEncoding.EncodeToString(bytes)
	return id[:idLength], nil
}

func HandleAdd(id, title, content string, tags []string) error {
	debugf("HandleAdd called: id=%q title=%q content=%q tags=%v", id, title, content, tags)

	if title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	if id == "" {
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
	} else {
		exists, err := storage.ItemExists(id)
		if err != nil {
			return fmt.Errorf("failed to check item existence: %w", err)
		}
		if exists {
			return fmt.Errorf("item with ID '%s' already exists", id)
		}
	}

	item := storage.NewTextItem(title, content, tags)

	// Get password with verification (checks/creates .dredge-key)
	password, err := crypto.GetPasswordWithVerification()
	if err != nil {
		return fmt.Errorf("failed to get password: %w", err)
	}

	if err := storage.CreateItem(id, item, password); err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	fmt.Printf("Created item '%s'\n", id)
	return nil
}
