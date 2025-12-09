package search

import (
	"testing"

	"github.com/DeprecatedLuar/dredge/internal/storage"
)

func TestSearch_SimpleQuery(t *testing.T) {
	items := map[string]*storage.Item{
		"proton-email": storage.NewTextItem(
			"ProtonMail Login",
			"user@proton.me\npassword: secure123",
			[]string{"email", "personal"},
		),
		"gmail-api": storage.NewTextItem(
			"Gmail API Key",
			"AIza...secret",
			[]string{"email", "api"},
		),
		"ssh-key": storage.NewTextItem(
			"SSH Private Key",
			"-----BEGIN RSA PRIVATE KEY-----",
			[]string{"ssh", "keys"},
		),
	}

	tests := []struct {
		name          string
		query         string
		wantCount     int
		wantFirstID   string
		wantFirstScore int
	}{
		{
			name:          "single term in title",
			query:         "proton",
			wantCount:     1,
			wantFirstID:   "proton-email",
			wantFirstScore: 101, // title(100) + content(1) - "proton" in user@proton.me
		},
		{
			name:          "multi-word AND logic",
			query:         "email proton",
			wantCount:     1,
			wantFirstID:   "proton-email",
			wantFirstScore: 111, // title(100) + tag(10) + content(1)
		},
		{
			name:          "tag match",
			query:         "api",
			wantCount:     1,
			wantFirstID:   "gmail-api",
			wantFirstScore: 110, // title(100) + tag(10)
		},
		{
			name:      "no results",
			query:     "github token",
			wantCount: 0,
		},
		{
			name:          "content match",
			query:         "rsa",
			wantCount:     1,
			wantFirstID:   "ssh-key",
			wantFirstScore: 1, // content match only
		},
		{
			name:          "case insensitive",
			query:         "PROTON",
			wantCount:     1,
			wantFirstID:   "proton-email",
			wantFirstScore: 101, // title(100) + content(1)
		},
		{
			name:      "empty query",
			query:     "",
			wantCount: 0,
		},
		{
			name:      "whitespace query",
			query:     "   ",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Search(items, tt.query)

			if len(results) != tt.wantCount {
				t.Errorf("Search() returned %d results, want %d", len(results), tt.wantCount)
				return
			}

			if tt.wantCount > 0 {
				if results[0].ID != tt.wantFirstID {
					t.Errorf("First result ID = %s, want %s", results[0].ID, tt.wantFirstID)
				}
				if results[0].Score != tt.wantFirstScore {
					t.Errorf("First result score = %d, want %d", results[0].Score, tt.wantFirstScore)
				}
			}
		})
	}
}

func TestSearch_Ranking(t *testing.T) {
	items := map[string]*storage.Item{
		"title-match": storage.NewTextItem(
			"Email Settings",
			"some content",
			[]string{"config"},
		),
		"tag-match": storage.NewTextItem(
			"Settings",
			"some content",
			[]string{"email"},
		),
		"content-match": storage.NewTextItem(
			"Work Notes",
			"My email is configured properly",
			[]string{"notes"},
		),
	}

	results := Search(items, "email")

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Check ranking: title > tag > content
	if results[0].ID != "title-match" {
		t.Errorf("Expected title-match first, got %s", results[0].ID)
	}
	if results[1].ID != "tag-match" {
		t.Errorf("Expected tag-match second, got %s", results[1].ID)
	}
	if results[2].ID != "content-match" {
		t.Errorf("Expected content-match third, got %s", results[2].ID)
	}

	// Verify scores
	if results[0].Score != 100 {
		t.Errorf("Title match score = %d, want 100", results[0].Score)
	}
	if results[1].Score != 10 {
		t.Errorf("Tag match score = %d, want 10", results[1].Score)
	}
	if results[2].Score != 1 {
		t.Errorf("Content match score = %d, want 1", results[2].Score)
	}
}

func TestSearch_MultipleTagMatches(t *testing.T) {
	items := map[string]*storage.Item{
		"item1": storage.NewTextItem(
			"Title",
			"content",
			[]string{"work", "email", "important"},
		),
	}

	// Query with term that matches one tag
	results := Search(items, "email")
	if len(results) != 1 || results[0].Score != 10 {
		t.Errorf("Expected 1 result with score 10, got %d results with score %d", len(results), results[0].Score)
	}
}

func TestSearch_ANDLogic(t *testing.T) {
	items := map[string]*storage.Item{
		"item1": storage.NewTextItem(
			"Email Settings",
			"proton config",
			[]string{"email"},
		),
		"item2": storage.NewTextItem(
			"ProtonMail",
			"settings here",
			[]string{"personal"},
		),
	}

	// Both terms must match
	results := Search(items, "email proton")

	if len(results) != 1 {
		t.Errorf("Expected 1 result (AND logic), got %d", len(results))
		return
	}

	if results[0].ID != "item1" {
		t.Errorf("Expected item1, got %s", results[0].ID)
	}
}

func TestSearch_FuzzyMatching(t *testing.T) {
	items := map[string]*storage.Item{
		"email-item": storage.NewTextItem(
			"Email Settings",
			"configuration",
			[]string{"email", "config"},
		),
		"proton-item": storage.NewTextItem(
			"ProtonMail",
			"mail service",
			[]string{"proton"},
		),
	}

	tests := []struct {
		name       string
		query      string
		wantCount  int
		wantID     string
		minScore   int // Use min score since fuzzy lib behavior may vary
		wantReason string
	}{
		{
			name:       "fuzzy typo in title",
			query:      "emial", // typo: email
			wantCount:  1,
			wantID:     "email-item",
			minScore:   50, // fuzzy title match
			wantReason: "emial→email typo correction",
		},
		{
			name:       "fuzzy typo in tag",
			query:      "protn", // typo: proton
			wantCount:  1,
			wantID:     "proton-item",
			minScore:   5, // fuzzy tag match
			wantReason: "protn→proton typo correction",
		},
		{
			name:       "short term no fuzzy",
			query:      "emi", // too short
			wantCount:  0,
			wantReason: "term too short for fuzzy (< 4 chars)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := Search(items, tt.query)

			if len(results) != tt.wantCount {
				t.Errorf("%s: got %d results, want %d", tt.wantReason, len(results), tt.wantCount)
				return
			}

			if tt.wantCount > 0 {
				if results[0].ID != tt.wantID {
					t.Errorf("%s: got ID %s, want %s", tt.wantReason, results[0].ID, tt.wantID)
				}
				if results[0].Score < tt.minScore {
					t.Errorf("%s: got score %d, want >= %d", tt.wantReason, results[0].Score, tt.minScore)
				}
			}
		})
	}
}

func TestSearch_ExactBeforeFuzzy(t *testing.T) {
	items := map[string]*storage.Item{
		"exact-match": storage.NewTextItem(
			"Email Settings",
			"content",
			[]string{"email"},
		),
		"fuzzy-match": storage.NewTextItem(
			"Emails Archive",
			"content",
			[]string{"archive"},
		),
	}

	results := Search(items, "email")

	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	// Exact match should rank higher than fuzzy
	if results[0].ID != "exact-match" {
		t.Errorf("Exact match should rank first, got %s", results[0].ID)
	}

	if results[0].Score <= results[1].Score {
		t.Errorf("Exact match score (%d) should be > fuzzy match score (%d)",
			results[0].Score, results[1].Score)
	}
}
