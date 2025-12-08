package crypto

import (
	"fmt"
	"os"
	"strconv"
)

// Session cache configuration
const (
	SessionCacheDir  = "/tmp"
	SessionCacheFile = ".sk-%d" // %d = $PPID (obscured name for security)
	CachePermissions = 0600     // User-only read/write
)

// GetCachedPassword retrieves the cached password from /tmp/.sk-$PPID.
// Returns empty string if cache doesn't exist.
func GetCachedPassword() (string, error) {
	ppid := os.Getppid()
	cachePath := fmt.Sprintf("%s/%s", SessionCacheDir, fmt.Sprintf(SessionCacheFile, ppid))

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // Cache doesn't exist, not an error
		}
		return "", fmt.Errorf("failed to read session cache: %w", err)
	}

	return string(data), nil
}

// CachePassword stores the password in /tmp/.sk-$PPID with 0600 permissions.
func CachePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	ppid := os.Getppid()
	cachePath := fmt.Sprintf("%s/%s", SessionCacheDir, fmt.Sprintf(SessionCacheFile, ppid))

	if err := os.WriteFile(cachePath, []byte(password), CachePermissions); err != nil {
		return fmt.Errorf("failed to cache password: %w", err)
	}

	return nil
}

// ClearSession removes the session cache file.
func ClearSession() error {
	ppid := os.Getppid()
	cachePath := fmt.Sprintf("%s/%s", SessionCacheDir, fmt.Sprintf(SessionCacheFile, ppid))

	err := os.Remove(cachePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to clear session: %w", err)
	}

	return nil
}

// HasActiveSession checks if a session cache exists for current terminal.
func HasActiveSession() bool {
	password, err := GetCachedPassword()
	return err == nil && password != ""
}

// GetPPID returns the parent process ID (for debugging/testing).
func GetPPID() string {
	return strconv.Itoa(os.Getppid())
}
