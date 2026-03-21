package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/DeprecatedLuar/dredge/internal/git"
	"github.com/DeprecatedLuar/dredge/internal/storage"
)

func HandleInit(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("usage: dredge init [remote-url]")
	}

	remote := ""
	if len(args) == 1 {
		remote = args[0]
	}

	// Get dredge directory
	dredgeDir, err := storage.GetDredgeDir()
	if err != nil {
		return fmt.Errorf("failed to get dredge directory: %w", err)
	}

	// Initialize git repository (and optionally connect remote)
	return git.Init(dredgeDir, remote)
}

// EnsureInitialized checks if a git repo is connected and prompts for one if not.
// Intended to be called from the app Before hook on every command except init/help.
func EnsureInitialized() error {
	dredgeDir, err := storage.GetDredgeDir()
	if err != nil {
		return fmt.Errorf("failed to get dredge directory: %w", err)
	}

	if git.IsInitialized(dredgeDir) {
		return nil
	}

	fmt.Fprintln(os.Stderr, "No git repository initialized for dredge.")
	fmt.Fprint(os.Stderr, "Enter remote git URL (or leave empty for local-only): ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("no input provided")
	}
	remote := strings.TrimSpace(scanner.Text())

	if err := os.MkdirAll(dredgeDir, 0700); err != nil {
		return fmt.Errorf("failed to create dredge directory: %w", err)
	}

	return git.Init(dredgeDir, remote)
}
