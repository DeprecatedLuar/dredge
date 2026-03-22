//go:build darwin

package session

import (
	"fmt"
	"os"
	"path/filepath"
)

// runtimeDir returns the base runtime directory on macOS.
// Uses XDG_RUNTIME_DIR if set, otherwise falls back to $TMPDIR/dredge-user-$UID
// (macOS has no /run/user equivalent; $TMPDIR is a per-user temporary directory).
func runtimeDir() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return dir
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("dredge-user-%d", os.Getuid()))
}
