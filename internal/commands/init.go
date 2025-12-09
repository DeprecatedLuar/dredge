package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/dredge/internal/git"
	"github.com/DeprecatedLuar/dredge/internal/storage"
)

func HandleInit(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: dredge init <user/repo>")
	}

	repoSlug := args[0]

	// Get dredge directory
	dredgeDir, err := storage.GetDredgeDir()
	if err != nil {
		return fmt.Errorf("failed to get dredge directory: %w", err)
	}

	// Initialize git repository
	return git.Init(dredgeDir, repoSlug)
}
