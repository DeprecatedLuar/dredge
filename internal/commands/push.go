package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/dredge-cargo/internal/git"
	"github.com/DeprecatedLuar/dredge-cargo/internal/storage"
)

func HandlePush(args []string) error {
	// Get dredge directory
	dredgeDir, err := storage.GetDredgeDir()
	if err != nil {
		return fmt.Errorf("failed to get dredge directory: %w", err)
	}

	// Push changes
	return git.Push(dredgeDir)
}
