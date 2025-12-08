package parser

import (
	"reflect"
	"testing"
)

func TestExtractTags(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedTags  []string
		expectedRemaining []string
	}{
		{
			name:          "tags at end",
			args:          []string{"some", "text", "#tag1", "#tag2"},
			expectedTags:  []string{"tag1", "tag2"},
			expectedRemaining: []string{"some", "text"},
		},
		{
			name:          "tags mixed",
			args:          []string{"#tag1", "some", "#tag2", "text"},
			expectedTags:  []string{"tag1", "tag2"},
			expectedRemaining: []string{"some", "text"},
		},
		{
			name:          "no tags",
			args:          []string{"just", "text"},
			expectedTags:  nil,
			expectedRemaining: []string{"just", "text"},
		},
		{
			name:          "only tags",
			args:          []string{"#tag1", "#tag2"},
			expectedTags:  []string{"tag1", "tag2"},
			expectedRemaining: nil,
		},
		{
			name:          "empty tag ignored",
			args:          []string{"text", "#"},
			expectedTags:  nil,
			expectedRemaining: []string{"text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags, remaining := ExtractTags(tt.args)
			if !reflect.DeepEqual(tags, tt.expectedTags) {
				t.Errorf("tags = %v, want %v", tags, tt.expectedTags)
			}
			if !reflect.DeepEqual(remaining, tt.expectedRemaining) {
				t.Errorf("remaining = %v, want %v", remaining, tt.expectedRemaining)
			}
		})
	}
}

func TestParseTagModifications(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedAdd      []string
		expectedRemove   []string
		expectedRemaining []string
	}{
		{
			name:             "add and remove",
			args:             []string{"+new", "-old", "other"},
			expectedAdd:      []string{"new"},
			expectedRemove:   []string{"old"},
			expectedRemaining: []string{"other"},
		},
		{
			name:             "only additions",
			args:             []string{"+tag1", "+tag2"},
			expectedAdd:      []string{"tag1", "tag2"},
			expectedRemove:   nil,
			expectedRemaining: nil,
		},
		{
			name:             "only removals",
			args:             []string{"-tag1", "-tag2"},
			expectedAdd:      nil,
			expectedRemove:   []string{"tag1", "tag2"},
			expectedRemaining: nil,
		},
		{
			name:             "mixed with text",
			args:             []string{"id", "+production", "-staging"},
			expectedAdd:      []string{"production"},
			expectedRemove:   []string{"staging"},
			expectedRemaining: []string{"id"},
		},
		{
			name:             "ignore single symbols",
			args:             []string{"+", "-", "text"},
			expectedAdd:      nil,
			expectedRemove:   nil,
			expectedRemaining: []string{"+", "-", "text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			add, remove, remaining := ParseTagModifications(tt.args)
			if !reflect.DeepEqual(add, tt.expectedAdd) {
				t.Errorf("add = %v, want %v", add, tt.expectedAdd)
			}
			if !reflect.DeepEqual(remove, tt.expectedRemove) {
				t.Errorf("remove = %v, want %v", remove, tt.expectedRemove)
			}
			if !reflect.DeepEqual(remaining, tt.expectedRemaining) {
				t.Errorf("remaining = %v, want %v", remaining, tt.expectedRemaining)
			}
		})
	}
}

func TestSplitByFlag(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		flag           string
		expectedBefore []string
		expectedAfter  []string
		expectedFound  bool
	}{
		{
			name:           "flag present",
			args:           []string{"title", "here", "-c", "content", "here"},
			flag:           "-c",
			expectedBefore: []string{"title", "here"},
			expectedAfter:  []string{"content", "here"},
			expectedFound:  true,
		},
		{
			name:           "flag at start",
			args:           []string{"-c", "content"},
			flag:           "-c",
			expectedBefore: []string{},
			expectedAfter:  []string{"content"},
			expectedFound:  true,
		},
		{
			name:           "flag at end",
			args:           []string{"title", "-c"},
			flag:           "-c",
			expectedBefore: []string{"title"},
			expectedAfter:  []string{},
			expectedFound:  true,
		},
		{
			name:           "flag not present",
			args:           []string{"just", "text"},
			flag:           "-c",
			expectedBefore: []string{"just", "text"},
			expectedAfter:  nil,
			expectedFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before, after, found := SplitByFlag(tt.args, tt.flag)
			if !reflect.DeepEqual(before, tt.expectedBefore) {
				t.Errorf("before = %v, want %v", before, tt.expectedBefore)
			}
			if !reflect.DeepEqual(after, tt.expectedAfter) {
				t.Errorf("after = %v, want %v", after, tt.expectedAfter)
			}
			if found != tt.expectedFound {
				t.Errorf("found = %v, want %v", found, tt.expectedFound)
			}
		})
	}
}

func TestParseAddCommand(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		expectedTitle   string
		expectedContent string
		expectedTags    []string
		expectError     bool
	}{
		{
			name:            "full format with content and tags",
			args:            []string{"SSH", "Config", "-c", "Host github.com", "#config", "#system"},
			expectedTitle:   "SSH Config",
			expectedContent: "Host github.com",
			expectedTags:    []string{"config", "system"},
		},
		{
			name:            "no content flag",
			args:            []string{"Simple", "Title", "#tag"},
			expectedTitle:   "Simple Title",
			expectedContent: "",
			expectedTags:    []string{"tag"},
		},
		{
			name:            "only title",
			args:            []string{"Just", "A", "Title"},
			expectedTitle:   "Just A Title",
			expectedContent: "",
			expectedTags:    nil,
		},
		{
			name:            "multiword content",
			args:            []string{"Title", "-c", "This", "is", "content", "#tag1", "#tag2"},
			expectedTitle:   "Title",
			expectedContent: "This is content",
			expectedTags:    []string{"tag1", "tag2"},
		},
		{
			name:            "content with no title",
			args:            []string{"-c", "Just content"},
			expectedTitle:   "",
			expectedContent: "Just content",
			expectedTags:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, content, tags, err := ParseAddCommand(tt.args, "-c")

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if title != tt.expectedTitle {
				t.Errorf("title = %q, want %q", title, tt.expectedTitle)
			}
			if content != tt.expectedContent {
				t.Errorf("content = %q, want %q", content, tt.expectedContent)
			}
			if !reflect.DeepEqual(tags, tt.expectedTags) {
				t.Errorf("tags = %v, want %v", tags, tt.expectedTags)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		isValid bool
	}{
		{"simple valid", "ssh", true},
		{"with dash", "github-api", true},
		{"with underscore", "proton_email", true},
		{"alphanumeric", "api2key3", true},
		{"max length", "a123456789012345678901234567890b", true},
		{"too long", "a123456789012345678901234567890bc", false},
		{"empty", "", false},
		{"with space", "ssh config", false},
		{"with special char", "ssh@key", false},
		{"with dot", "ssh.key", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateID(tt.id)
			if tt.isValid && err != nil {
				t.Errorf("expected valid but got error: %v", err)
			}
			if !tt.isValid && err == nil {
				t.Error("expected invalid but got no error")
			}
		})
	}
}

func TestIsValidID(t *testing.T) {
	tests := []struct {
		id       string
		expected bool
	}{
		{"ssh", true},
		{"github-api", true},
		{"proton_email", true},
		{"k3m", true},
		{"ssh config", false},
		{"ssh@key", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			result := IsValidID(tt.id)
			if result != tt.expected {
				t.Errorf("IsValidID(%q) = %v, want %v", tt.id, result, tt.expected)
			}
		})
	}
}

func TestJoinArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{"multiple args", []string{"hello", "world"}, "hello world"},
		{"single arg", []string{"hello"}, "hello"},
		{"empty slice", []string{}, ""},
		{"with spaces preserved", []string{"hello", "beautiful", "world"}, "hello beautiful world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JoinArgs(tt.args)
			if result != tt.expected {
				t.Errorf("JoinArgs() = %q, want %q", result, tt.expected)
			}
		})
	}
}
