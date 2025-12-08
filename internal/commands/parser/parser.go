package parser

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// ID validation: alphanumeric + dash/underscore, max 32 chars
	idRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)
)

// ExtractTags extracts hashtags from a slice of arguments
// Returns extracted tags and remaining args without hashtags
// Example: ["some", "text", "#tag1", "#tag2"] -> (["tag1", "tag2"], ["some", "text"])
func ExtractTags(args []string) (tags []string, remaining []string) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "#") {
			tag := strings.TrimPrefix(arg, "#")
			if tag != "" {
				tags = append(tags, tag)
			}
		} else {
			remaining = append(remaining, arg)
		}
	}
	return tags, remaining
}

// ParseTagModifications extracts tag additions (+tag) and removals (-tag)
// Returns added tags, removed tags, and remaining args
// Example: ["+new", "-old", "other"] -> (["new"], ["old"], ["other"])
func ParseTagModifications(args []string) (add []string, remove []string, remaining []string) {
	for _, arg := range args {
		if strings.HasPrefix(arg, "+") && len(arg) > 1 {
			add = append(add, strings.TrimPrefix(arg, "+"))
		} else if strings.HasPrefix(arg, "-") && len(arg) > 1 {
			remove = append(remove, strings.TrimPrefix(arg, "-"))
		} else {
			remaining = append(remaining, arg)
		}
	}
	return add, remove, remaining
}

// SplitByFlag splits arguments at a flag marker
// Returns everything before the flag and everything after it
// Example: SplitByFlag(["title", "here", "-c", "content", "here"], "-c")
//   -> (["title", "here"], ["content", "here"])
func SplitByFlag(args []string, flag string) (before []string, after []string, found bool) {
	for i, arg := range args {
		if arg == flag {
			return args[:i], args[i+1:], true
		}
	}
	return args, nil, false
}

// ParseAddCommand parses arguments for the 'add' command
// Format: [title...] -c [content...] #tag1 #tag2
// Or for files: --file <path> [title...] #tag1 #tag2
func ParseAddCommand(args []string, contentFlag string) (title string, content string, tags []string, err error) {
	// First extract tags
	tags, argsWithoutTags := ExtractTags(args)

	// Split by content flag
	beforeContent, afterContent, hasContent := SplitByFlag(argsWithoutTags, contentFlag)

	if !hasContent {
		// No content flag - everything is title
		title = strings.Join(argsWithoutTags, " ")
		return title, "", tags, nil
	}

	// Before -c is title
	title = strings.Join(beforeContent, " ")

	// After -c is content (stop at first hashtag which was already removed)
	content = strings.Join(afterContent, " ")

	return title, content, tags, nil
}

// ValidateID checks if an ID is valid
func ValidateID(id string) error {
	if !idRegex.MatchString(id) {
		return fmt.Errorf("invalid ID format: must be alphanumeric with dashes/underscores, max 32 chars")
	}
	return nil
}

// IsValidID returns true if the string is a valid ID format
func IsValidID(s string) bool {
	return idRegex.MatchString(s)
}

// JoinArgs joins arguments with spaces, useful for reconstructing parsed content
func JoinArgs(args []string) string {
	return strings.Join(args, " ")
}
