package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/storage"
)

const (
	tempDirBase    = "/tmp/dredge"
	tempFilePrefix = "edit-"
	tempFileSuffix = ".txt"
	defaultEditor  = "vim"
)

// getSessionDir returns the session-specific directory path
func getSessionDir() string {
	return fmt.Sprintf("%s/%d", tempDirBase, os.Getppid())
}

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

	// Convert item to template format
	templateContent := itemToTemplate(item)

	// Ensure session directory exists
	sessionDir := getSessionDir()
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return fmt.Errorf("failed to create session directory: %w", err)
	}

	// Create temp file in session directory
	tempFile, err := os.CreateTemp(sessionDir, tempFilePrefix+"*"+tempFileSuffix)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	// Write template to temp file
	if _, err := tempFile.WriteString(templateContent); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tempFile.Close()

	// Open editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = defaultEditor
	}

	cmd := exec.Command(editor, tempPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	// Read back edited content
	editedContent, err := os.ReadFile(tempPath)
	if err != nil {
		return fmt.Errorf("failed to read edited content: %w", err)
	}

	// Parse template back to item
	updatedItem, err := templateToItem(string(editedContent), item)
	if err != nil {
		return fmt.Errorf("failed to parse edited content: %w", err)
	}

	// Update modification time
	updatedItem.Modified = time.Now()

	// Save updated item
	if err := storage.UpdateItem(id, updatedItem, password); err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	fmt.Printf("✓ [%s] %s\n", id, updatedItem.Title)
	return nil
}

// itemToTemplate converts an Item to the simple template format:
// Line 1: Title -t tag1 tag2
// Line 2: blank
// Lines 3+: Content
func itemToTemplate(item *storage.Item) string {
	var sb strings.Builder

	// Line 1: Title and tags
	sb.WriteString(item.Title)
	if len(item.Tags) > 0 {
		sb.WriteString(" -t ")
		sb.WriteString(strings.Join(item.Tags, " "))
	}
	sb.WriteString("\n")

	// Line 2: Blank separator
	sb.WriteString("\n")

	// Lines 3+: Content
	sb.WriteString(item.Content.Text)

	return sb.String()
}

// templateToItem parses the simple template format back into an Item.
// Preserves metadata (created, type, id) from the existing item.
func templateToItem(content string, existing *storage.Item) (*storage.Item, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("invalid template format: expected at least 3 lines")
	}

	// Parse line 1: Title -t tag1 tag2
	firstLine := lines[0]
	var title string
	var tags []string

	if idx := strings.Index(firstLine, " -t "); idx != -1 {
		title = strings.TrimSpace(firstLine[:idx])
		tagsPart := strings.TrimSpace(firstLine[idx+4:])
		if tagsPart != "" {
			tags = strings.Fields(tagsPart)
		}
	} else {
		title = strings.TrimSpace(firstLine)
	}

	if title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	}

	// Line 2 should be blank (we skip it)

	// Lines 3+ become content
	contentText := strings.Join(lines[2:], "\n")

	// Create updated item, preserving metadata
	updated := &storage.Item{
		Title:    title,
		Tags:     tags,
		Type:     existing.Type,
		Created:  existing.Created,
		Modified: existing.Modified, // Will be updated by caller
		Filename: existing.Filename,
		Size:     existing.Size,
		Content: storage.ItemContent{
			Text: contentText,
		},
	}

	return updated, nil
}
