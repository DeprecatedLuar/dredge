package parser

import (
	"fmt"
	"os"

	"github.com/DeprecatedLuar/dredge/internal/commands"
)

func Route(args []string) {
	debugMode := false
	filteredArgs := make([]string, 0, len(args))

	for _, arg := range args {
		if arg == "--debug" {
			debugMode = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	commands.SetDebugMode(debugMode)

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
		err = commands.HandleSearch(cmdArgs)
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
		// Smart query: if arg looks like an ID, view it; otherwise search
		// This enables: dredge <id> or dredge <search-query>
		query := cmd
		if ValidateID(query) == nil {
			// Valid ID format, try to view it
			err = commands.HandleView([]string{query})
		} else {
			// Not a valid ID, treat as search query
			allArgs := append([]string{query}, cmdArgs...)
			err = commands.HandleSearch(allArgs)
		}
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
