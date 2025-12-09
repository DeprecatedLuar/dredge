package search

import (
	"strings"

	"github.com/DeprecatedLuar/dredge/internal/storage"
	"github.com/hbollon/go-edlib"
)

const (
	// Scoring weights (exact matches)
	titleMatchScore   = 100
	tagMatchScore     = 10
	contentMatchScore = 1

	// Fuzzy match weights (lower than exact)
	fuzzyTitleMatchScore = 50
	fuzzyTagMatchScore   = 5

	// Fuzzy matching config
	minFuzzyLength      = 4   // Only fuzzy match strings >= 4 chars (avoid "api"→"pi")
	similarityThreshold = 0.8 // 80% similarity for fuzzy matching

	// Matching threshold
	matchThresholdPercent = 0.5 // Require >50% of exponential weight to match
)

// Result represents a search result with score
type Result struct {
	ID    string
	Item  *storage.Item
	Score int
}

// Search performs a simple ranked search across items
// Query is split into terms (space-separated)
// All terms must match (AND logic)
// Scoring: title=100, tags=10, content=1
func Search(items map[string]*storage.Item, query string) []Result {
	query = strings.TrimSpace(query)
	if query == "" {
		return []Result{}
	}

	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return []Result{}
	}

	var results []Result

	for id, item := range items {
		score, matched := scoreItem(item, terms)
		if matched {
			results = append(results, Result{
				ID:    id,
				Item:  item,
				Score: score,
			})
		}
	}

	// Sort by score descending (bubble sort is fine for <1000 items)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// scoreItem scores an item against search terms
// Returns (score, matched) where matched=true if >50% of exponential weight matched
// Exponential weighting: longer words dominate (github²=36 >> key²=9)
func scoreItem(item *storage.Item, terms []string) (int, bool) {
	title := strings.ToLower(item.Title)
	content := strings.ToLower(item.Content.Text)

	// Lowercase all tags once
	tags := make([]string, len(item.Tags))
	for i, tag := range item.Tags {
		tags[i] = strings.ToLower(tag)
	}

	score := 0
	totalWeight := 0
	matchedWeight := 0

	// Check how many terms match (exponential length-weighted threshold)
	for _, term := range terms {
		termScore := 0
		termLen := len(term)
		termWeight := termLen * termLen // Exponential: longer words dominate
		totalWeight += termWeight

		// Try exact matches first
		if strings.Contains(title, term) {
			termScore += titleMatchScore * termWeight
		}

		for _, tag := range tags {
			if strings.Contains(tag, term) {
				termScore += tagMatchScore * termWeight
				break // Count tag match once per term
			}
		}

		if strings.Contains(content, term) {
			termScore += contentMatchScore * termWeight
		}

		// If no exact match and term is long enough, try fuzzy matching
		if termScore == 0 && termLen >= minFuzzyLength {
			// Fuzzy match title: try full title first, then individual words
			if fuzzyMatchWord(term, title) {
				termScore += fuzzyTitleMatchScore * termWeight
			} else {
				// Check each word in title
				titleWords := strings.Fields(title)
				for _, word := range titleWords {
					if fuzzyMatchWord(term, word) {
						termScore += fuzzyTitleMatchScore * termWeight
						break
					}
				}
			}

			// Fuzzy match tags (only if title didn't match)
			if termScore == 0 {
				for _, tag := range tags {
					if fuzzyMatchWord(term, tag) {
						termScore += fuzzyTagMatchScore * termWeight
						break
					}
				}
			}
			// Note: No fuzzy matching for content (too much text, too slow)
		}

		// Count matched weight if term scored
		if termScore > 0 {
			matchedWeight += termWeight
			score += termScore
		}
	}

	// Exponential threshold: >50% of weight matched
	// Long words dominate: "github" (36) >> "key" (9)
	matchThreshold := float64(matchedWeight) / float64(totalWeight)
	if matchThreshold > matchThresholdPercent {
		return score, true
	}

	return 0, false
}

// fuzzyMatchWord checks if pattern matches a single word with similarity >= threshold
// Uses go-edlib for edit distance based similarity (handles typos)
func fuzzyMatchWord(pattern, word string) bool {
	if pattern == word {
		return true // Exact match (shouldn't happen here, but safety)
	}

	// Calculate similarity (0.0 to 1.0)
	// Uses Damerau-Levenshtein (better for transpositions like "emial"→"email")
	similarity, err := edlib.StringsSimilarity(pattern, word, edlib.DamerauLevenshtein)
	if err != nil {
		return false
	}

	// Match if similarity >= threshold (0.8 = 80% similar)
	// Note: similarityThreshold is float64, but StringsSimilarity returns float32
	return similarity >= float32(similarityThreshold)
}
