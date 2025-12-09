package parser

import (
	"fmt"
	"os"
	"strconv"

	"github.com/DeprecatedLuar/dredge/internal/commands"
	"github.com/DeprecatedLuar/dredge/internal/crypto"
)

var debugMode bool

func Debugf(format string, args ...any) {
	if debugMode {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

var (
	luckMode   bool
	searchMode bool
)

func Route(args []string) {
	password := ""
	luckMode = false
	searchMode = false
	filteredArgs := make([]string, 0, len(args))

	// Parse flags
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--debug" {
			debugMode = true
		} else if arg == "--password" || arg == "-p" {
			if i+1 < len(args) {
				password = args[i+1]
				i++ // Skip next arg (the password value)
			}
		} else if arg == "--luck" || arg == "-l" {
			luckMode = true
		} else if arg == "--search" || arg == "-s" {
			searchMode = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	// If password provided via flag, cache it immediately
	if password != "" {
		Debugf("Caching password from --password flag: %s", password)
		if err := crypto.CachePassword(password); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to cache password: %v\n", err)
		} else {
			Debugf("Password cached successfully")
		}
	}

	if len(filteredArgs) == 0 {
		commands.HandleHelp(nil)
		return
	}

	cmd := filteredArgs[0]
	cmdArgs := filteredArgs[1:]

	var err error
	switch cmd {
	case "add", "new", "+", "a":
		err = handleAddCommand(cmdArgs)
	case "search", "s":
		// Join all args into a single search query
		query := JoinArgs(append([]string{}, cmdArgs...))
		err = commands.HandleSearch(query, luckMode, searchMode)
	case "view", "v":
		err = commands.HandleView(cmdArgs)
	case "edit", "e":
		err = commands.HandleEdit(cmdArgs)
	case "rm":
		err = commands.HandleRemove(cmdArgs)
	case "link", "ln":
		err = commands.HandleLink(cmdArgs)
	case "unlink":
		err = commands.HandleUnlink(cmdArgs)
	case "init":
		err = commands.HandleInit(cmdArgs)
	case "push":
		err = commands.HandlePush(cmdArgs)
	case "pull":
		err = commands.HandlePull(cmdArgs)
	case "sync":
		err = commands.HandleSync(cmdArgs)
	case "help", "h", "-h", "--help":
		err = commands.HandleHelp(cmdArgs)
	default:
		// Smart query: numbered result → ID → search
		// This enables: dredge 1, dredge <id>, or dredge <search-query>
		allArgs := append([]string{cmd}, cmdArgs...)

		// Try as numbered result first (if single numeric arg)
		if len(cmdArgs) == 0 {
			if num, parseErr := strconv.Atoi(cmd); parseErr == nil && num > 0 {
				id, cacheErr := commands.GetCachedResult(num)
				if cacheErr == nil {
					err = commands.HandleView([]string{id})
					return
				}
				// If cache miss, fall through to try as ID/search
			}

			// Try as direct ID
			err = commands.HandleView([]string{cmd})
			// If view succeeded, we're done
			if err == nil {
				return
			}
		}

		// Fall back to search
		query := JoinArgs(allArgs)
		err = commands.HandleSearch(query, luckMode, searchMode)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func handleAddCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: dredge add [--id <id>] <title...> [-c <content...>] [-t|-#|--tag|--tags <tag1> <tag2> ...]")
	}

	var id string
	var title string
	var content string
	var tags []string
	remainingArgs := args

	if len(args) >= 2 && args[0] == "--id" {
		id = args[1]
		if err := ValidateID(id); err != nil {
			return fmt.Errorf("invalid ID: %w", err)
		}
		remainingArgs = args[2:]
	}

	if len(remainingArgs) == 0 {
		return fmt.Errorf("title is required")
	}

	titleArgs, contentArgs, tagsArgs := splitAddArgs(remainingArgs)

	title = JoinArgs(titleArgs)
	content = JoinArgs(contentArgs)
	tags = tagsArgs

	if title == "" {
		return fmt.Errorf("title is required")
	}

	return commands.HandleAdd(id, title, content, tags)
}

func splitAddArgs(args []string) (title, content, tags []string) {
	var titleEnd, contentStart, contentEnd, tagsStart int
	titleEnd = len(args)

	for i, arg := range args {
		if arg == "-c" {
			titleEnd = i
			if i+1 < len(args) {
				contentStart = i + 1
			}
			contentEnd = len(args)
		}
		if arg == "-t" || arg == "-#" || arg == "--tag" || arg == "--tags" {
			if contentStart > 0 {
				contentEnd = i
			} else {
				titleEnd = i
			}
			if i+1 < len(args) {
				tagsStart = i + 1
			}
			break
		}
	}

	title = args[:titleEnd]

	if contentStart > 0 {
		content = args[contentStart:contentEnd]
	}

	if tagsStart > 0 {
		tags = args[tagsStart:]
	}

	return title, content, tags
}
