//go:build linux

package session

import (
	"fmt"
	"os"
)

// runtimeDir returns the base runtime directory on Linux.
// Uses XDG_RUNTIME_DIR if set, otherwise falls back to /run/user/$UID.
func runtimeDir() string {
	if dir := os.Getenv("XDG_RUNTIME_DIR"); dir != "" {
		return dir
	}
	return fmt.Sprintf("/run/user/%d", os.Getuid())
}
