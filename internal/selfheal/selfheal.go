package selfheal

import "github.com/DeprecatedLuar/dredge/internal/storage"

// Run performs silent health checks and cleanup once per session
func Run() {
	// Clean up orphaned links (manifest entries where item no longer exists)
	for _, id := range storage.GetOrphanedLinkIDs() {
		_ = storage.Unlink(id)
	}

	// Clean up orphaned spawned files (not tracked in manifest)
	for _, id := range storage.GetOrphanedSpawnedFiles() {
		_ = storage.RemoveSpawnedFile(id)
	}
}
