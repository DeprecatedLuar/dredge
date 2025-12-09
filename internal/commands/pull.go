package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/dredge/internal/git"
	"github.com/DeprecatedLuar/dredge/internal/storage"
)

func HandlePull(args []string) error {
	// Get dredge directory
	dredgeDir, err := storage.GetDredgeDir()
	if err != nil {
		return fmt.Errorf("failed to get dredge directory: %w", err)
	}

	// Pull changes
	return git.Pull(dredgeDir)
}
