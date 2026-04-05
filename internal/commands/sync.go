package commands

import (
	"fmt"

	"github.com/DeprecatedLuar/dredge-cargo/internal/git"
	"github.com/DeprecatedLuar/dredge-cargo/internal/storage"
)

func HandleSync(args []string) error {
	// Get dredge directory
	dredgeDir, err := storage.GetDredgeDir()
	if err != nil {
		return fmt.Errorf("failed to get dredge directory: %w", err)
	}

	// Sync (pull + push)
	return git.Sync(dredgeDir)
}
