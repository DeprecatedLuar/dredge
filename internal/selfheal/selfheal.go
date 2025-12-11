package selfheal

import (
	"os"
	"path/filepath"

	"github.com/DeprecatedLuar/dredge/internal/storage"
)

// Run performs silent health checks and cleanup once per session
func Run() {
	cleanupOrphanedLinks()
}

// cleanupOrphanedLinks detects and removes orphaned links and spawned files
func cleanupOrphanedLinks() {
	manifest, err := storage.LoadManifest()
	if err != nil {
		return // Silent failure
	}

	// Call Unlink for each manifest entry (Unlink handles broken links gracefully)
	for id := range manifest {
		_ = storage.Unlink(id) // Ignore errors, Unlink cleans what exists
	}

	// Clean up orphaned spawned files (files in .spawned/ not in manifest)
	dredgeDir, err := storage.GetDredgeDir()
	if err != nil {
		return
	}

	spawnedDir := filepath.Join(dredgeDir, ".spawned")
	entries, err := os.ReadDir(spawnedDir)
	if err != nil {
		return // Directory might not exist, that's OK
	}

	// Reload manifest after Unlink calls (might have changed)
	manifest, err = storage.LoadManifest()
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		id := entry.Name()

		// If not in manifest, it's orphaned - remove it
		if _, exists := manifest[id]; !exists {
			spawnedPath := filepath.Join(spawnedDir, id)
			_ = os.Remove(spawnedPath) // Silent cleanup
		}
	}
}
