package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/search"
	"github.com/DeprecatedLuar/dredge/internal/storage"
	"golang.org/x/term"
)

const (
	resultsCache       = "/tmp/dredge-results-%d" // %d = $PPID
	smartThreshold     = 2.5                       // Top score must be 2.5x higher than second to auto-view
	defaultTermWidth   = 80                        // Default if terminal width can't be detected
	ellipsisLen        = 3                         // Length of "..." for truncation
	minTruncateLen     = 10                        // Minimum useful length before truncation
	tagSpacing         = 5                         // Reserve space between title and tags

	// Colors (24-bit RGB)
	idColor            = "\033[38;2;136;136;136m"  // #888888 - medium gray for IDs
	titleColor         = "\033[38;2;159;212;159m"  // #9fd49f - bright green for titles
	tagColor           = "\033[38;2;97;97;97m"     // #616161 - neutral dark gray for tags
	resetColor         = "\033[0m"                 // Reset color
)

func HandleSearch(query string, luck bool, forceSearch bool) error {
	// Get password (with verification and caching)
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
	items := make(map[string]*storage.Item)
	for _, id := range ids {
		item, err := storage.ReadItem(id, password)
		if err != nil {
			// Skip items that fail to decrypt (corrupted/wrong format)
			continue
		}
		items[id] = item
	}

	// Perform search
	results := search.Search(items, query)

	// Display results
	if len(results) == 0 {
		fmt.Printf("No results found for: %s\n", query)
		return nil
	}

	// Determine viewing mode:
	// 1. -l flag: always view top result
	// 2. -s flag: always show list
	// 3. Smart default: auto-view if clear winner, else list
	shouldAutoView := false

	if luck {
		// Force view top result
		shouldAutoView = true
	} else if !forceSearch {
		// Smart threshold: auto-view if clear winner
		if len(results) == 1 {
			// Only one result, definitely view it
			shouldAutoView = true
		} else if len(results) > 1 {
			// Check if top result is significantly better than second
			topScore := float64(results[0].Score)
			secondScore := float64(results[1].Score)
			if secondScore > 0 && topScore/secondScore >= smartThreshold {
				shouldAutoView = true
			}
		}
	}

	// Auto-view top result if conditions met
	if shouldAutoView {
		return HandleView([]string{results[0].ID})
	}

	// Show list with tags
	termWidth := getTerminalWidth()
	for _, result := range results {
		printResult(result, termWidth)
	}

	// Cache results for numbered access
	cacheResults(results)

	return nil
}

// getTerminalWidth returns the current terminal width, or default if unavailable
func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width <= 0 {
		return defaultTermWidth
	}
	return width
}

// printResult prints a search result with ID, title, and dimmed tags
// Format: [ID] Title #tag1 #tag2 #tag3
// Truncates if needed to fit terminal width
func printResult(result search.Result, termWidth int) {
	id := result.ID
	title := result.Item.Title
	tags := result.Item.Tags

	// Build tag string with # prefix
	var tagParts []string
	for _, tag := range tags {
		tagParts = append(tagParts, "#"+tag)
	}
	tagStr := strings.Join(tagParts, " ")

	// Format: [ID] Title #tags (with colors)
	plainPrefix := fmt.Sprintf("[%s] ", id)
	availableWidth := termWidth - len(plainPrefix)

	// Build full line
	var line string
	if len(tagStr) > 0 {
		fullText := fmt.Sprintf("%s%s%s %s%s%s", titleColor, title, resetColor, tagColor, tagStr, resetColor)
		plainText := fmt.Sprintf("%s %s", title, tagStr)

		// Check if truncation needed (use plain text length for calculation)
		if len(plainText) > availableWidth {
			// Truncate, leaving room for "..."
			maxLen := availableWidth - ellipsisLen
			if maxLen < minTruncateLen {
				maxLen = minTruncateLen
			}

			// Try to keep title + some tags
			if len(title) <= maxLen-tagSpacing {
				// Title fits, truncate tags
				remaining := maxLen - len(title) - 1
				truncatedTags := truncateString(tagStr, remaining)
				line = fmt.Sprintf("%s%s%s %s%s...%s", titleColor, title, resetColor, tagColor, truncatedTags, resetColor)
			} else {
				// Truncate title only
				truncatedTitle := truncateString(title, maxLen)
				line = fmt.Sprintf("%s%s...%s", titleColor, truncatedTitle, resetColor)
			}
		} else {
			line = fullText
		}
	} else {
		// No tags
		if len(title) > availableWidth {
			truncatedTitle := truncateString(title, availableWidth-ellipsisLen)
			line = fmt.Sprintf("%s%s...%s", titleColor, truncatedTitle, resetColor)
		} else {
			line = fmt.Sprintf("%s%s%s", titleColor, title, resetColor)
		}
	}

	fmt.Printf("%s[%s]%s %s\n", idColor, id, resetColor, line)
}

// truncateString truncates a string to maxLen runes (Unicode-safe)
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}

// cacheResults saves search results to /tmp for numbered access
func cacheResults(results []search.Result) {
	ids := make([]string, len(results))
	for i, r := range results {
		ids[i] = r.ID
	}

	data, err := json.Marshal(ids)
	if err != nil {
		return // Non-fatal, just skip caching
	}

	cachePath := fmt.Sprintf(resultsCache, os.Getppid())
	_ = os.WriteFile(cachePath, data, 0600) // Ignore errors
}

// GetCachedResult retrieves the ID for a numbered search result
func GetCachedResult(num int) (string, error) {
	cachePath := fmt.Sprintf(resultsCache, os.Getppid())

	data, err := os.ReadFile(cachePath)
	if err != nil {
		return "", fmt.Errorf("no recent search results")
	}

	var ids []string
	if err := json.Unmarshal(data, &ids); err != nil {
		return "", fmt.Errorf("invalid search cache")
	}

	if num < 1 || num > len(ids) {
		return "", fmt.Errorf("result number out of range (1-%d)", len(ids))
	}

	return ids[num-1], nil
}
