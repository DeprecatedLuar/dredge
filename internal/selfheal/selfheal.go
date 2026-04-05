package selfheal

import "github.com/DeprecatedLuar/dredge-cargo/internal/storage"

// Run performs silent health checks and cleanup once per session
func Run() {
	if DetectLegacyVault() {
		RunMigration()
		return
	}

	// Clean up orphaned links (manifest entries where item no longer exists)
	for _, id := range storage.GetOrphanedLinkIDs() {
		_ = storage.Unlink(id)
	}

	// Recreate broken symlinks (manifest entries where symlink was deleted)
	storage.RepairBrokenSymlinks()

	// Clean up orphaned spawned files (not tracked in manifest)
	for _, id := range storage.GetOrphanedSpawnedFiles() {
		_ = storage.RemoveSpawnedFile(id)
	}
}
