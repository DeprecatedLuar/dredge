package selfheal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DeprecatedLuar/dredge-cargo/internal/storage"
)

const migratedVaultName = "dredge-vault-migrated"

// DetectLegacyVault returns true if ~/.local/share/dredge/items/ exists,
// indicating vault data is stored in the registry directory (legacy layout).
func DetectLegacyVault() bool {
	registryDir, err := storage.GetRegistryDir()
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(registryDir, "items"))
	return err == nil
}

// RunMigration copies the legacy vault to ~/dredge-vault-migrated, verifies
// integrity, then removes the old location. Prints result to stdout.
func RunMigration() {
	registryDir, err := storage.GetRegistryDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Migration error: could not locate legacy vault: %v\n", err)
		return
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Migration error: could not get home directory: %v\n", err)
		return
	}

	dst := filepath.Join(homeDir, migratedVaultName)

	if err := os.MkdirAll(dst, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Migration error: could not create destination: %v\n", err)
		return
	}

	if err := os.CopyFS(dst, os.DirFS(registryDir)); err != nil {
		fmt.Fprintf(os.Stderr, "Migration error: copy failed: %v\n%s\n", err, manualInstructions(registryDir, dst))
		return
	}

	if err := checkIntegrity(registryDir, dst); err != nil {
		fmt.Fprintf(os.Stderr, "Migration error: %v\n%s\n", err, manualInstructions(registryDir, dst))
		return
	}

	readme := `Hi, Luar here.

I made a breaking update to dredge and to ensure everything would still work I moved the folder here for now
This vault is now active. The old location (~/.local/share/dredge/) has been deprecated.

Basically just put it wherever you want on the system and run 'dredge init' inside of it
`
	_ = os.WriteFile(filepath.Join(dst, "migration-note.md"), []byte(readme), 0644)

	if err := os.RemoveAll(registryDir); err != nil {
		fmt.Fprintf(os.Stderr, "Migration warning: could not remove old vault at %s: %v\n", registryDir, err)
	}

	if err := storage.SetActivePath(dst); err != nil {
		fmt.Fprintf(os.Stderr, "Migration warning: could not activate new vault: %v\n  Run: dredge init %s\n", err, dst)
	}

	fmt.Printf("Your vault has been moved to %s\n", dst)
	fmt.Printf("Old location at %s has been deprecated.\n", registryDir)
}

func checkIntegrity(src, dst string) error {
	if _, err := os.Stat(filepath.Join(dst, ".dredge-key")); os.IsNotExist(err) {
		return fmt.Errorf("integrity check failed: .dredge-key missing from destination")
	}
	if _, err := os.Stat(filepath.Join(dst, "items")); os.IsNotExist(err) {
		return fmt.Errorf("integrity check failed: items/ directory missing from destination")
	}

	srcItems, err := os.ReadDir(filepath.Join(src, "items"))
	if err != nil {
		return fmt.Errorf("integrity check failed: could not read source items: %w", err)
	}
	dstItems, err := os.ReadDir(filepath.Join(dst, "items"))
	if err != nil {
		return fmt.Errorf("integrity check failed: could not read destination items: %w", err)
	}
	if len(srcItems) != len(dstItems) {
		return fmt.Errorf("integrity check failed: item count mismatch (src=%d dst=%d)", len(srcItems), len(dstItems))
	}

	return nil
}

func manualInstructions(src, dst string) string {
	return fmt.Sprintf(
		"Manual migration instructions:\n  1. Copy %s to a new location\n  2. Run: dredge init <new location>\n  Note: %s may contain a partial copy and should be removed.",
		src, dst,
	)
}
