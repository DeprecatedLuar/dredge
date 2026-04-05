package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DeprecatedLuar/dredge/internal/crypto"
	"github.com/DeprecatedLuar/dredge/internal/git"
	"github.com/DeprecatedLuar/dredge/internal/storage"
)

// HandleUse activates an existing vault directory and persists it as the active vault.
// Usage: dredge use <path>
func HandleUse(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: dredge use <path>")
	}

	path := strings.TrimSpace(args[0])
	if path == "" {
		return fmt.Errorf("usage: dredge use <path>")
	}

	absPath, err := resolveUserPath(path)
	if err != nil {
		return err
	}

	if !isVaultDir(absPath) {
		return fmt.Errorf("not a vault directory: %s (expected %s)", absPath, filepath.Join(absPath, "items"))
	}

	_ = crypto.ClearSession()
	if err := storage.SetActivePath(absPath); err != nil {
		return fmt.Errorf("failed to set active vault: %w", err)
	}

	if url, ok := git.RemoteURL(absPath); ok {
		fmt.Printf("Activated %s\n", strings.TrimSpace(url))
	} else {
		fmt.Printf("Activated %s\n", absPath)
	}

	return nil
}

func resolveUserPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("empty path")
	}
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to expand ~: %w", err)
		}
		if p == "~" {
			p = home
		} else {
			p = filepath.Join(home, p[2:])
		}
	} else if strings.HasPrefix(p, "~") {
		return "", fmt.Errorf("unsupported ~ expansion in path: %q", p)
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %q: %w", p, err)
	}
	return abs, nil
}
