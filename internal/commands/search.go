package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/search"
	"github.com/DeprecatedLuar/dredge/internal/storage"
	"github.com/DeprecatedLuar/dredge/internal/ui"
)

const (
	smartThreshold = 2.5 // Top score must be 2.5x higher than second to auto-view
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

	// Show list
	for _, result := range results {
		fmt.Println(ui.FormatItem(result.ID, result.Item.Title, result.Item.Tags, "it#"))
	}

	// Cache results for numbered access
	resultIDs := make([]string, len(results))
	for i, r := range results {
		resultIDs[i] = r.ID
	}
	storage.CacheResults(resultIDs) // Ignore errors (non-fatal)

	return nil
}

// ResolveArgs converts numbered args to IDs using cached search results
// Non-numeric args are passed through as-is (assumed to be IDs)
func ResolveArgs(args []string) ([]string, error) {
	resolved := make([]string, len(args))

	for i, arg := range args {
		// Try parsing as number
		var num int
		if _, err := fmt.Sscanf(arg, "%d", &num); err == nil && num > 0 {
			// It's a number, resolve from cache
			id, cacheErr := storage.GetCachedResult(num)
			if cacheErr != nil {
				return nil, fmt.Errorf("arg %q: %w", arg, cacheErr)
			}
			resolved[i] = id
		} else {
			// Not a number, assume it's an ID
			resolved[i] = arg
		}
	}

	return resolved, nil
}
