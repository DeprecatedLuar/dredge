package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	tempDirBase = "/tmp/dredge"

	// Cache file names (no PID needed, stored in session dir)
	resultsCacheFile = "results"
	deletedCacheFile = "deleted"
)

// getSessionDir returns the session-specific directory path
func getSessionDir() string {
	return filepath.Join(tempDirBase, fmt.Sprintf("%d", os.Getppid()))
}

// ensureSessionDir creates /tmp/dredge/$PPID if it doesn't exist
func ensureSessionDir() error {
	return os.MkdirAll(getSessionDir(), 0700)
}

// ============================================================================
// Results Cache (for list/search numbered access)
// ============================================================================

// CacheResults saves item IDs for numbered access
func CacheResults(ids []string) error {
	if err := ensureSessionDir(); err != nil {
		return err
	}

	data, err := json.Marshal(ids)
	if err != nil {
		return fmt.Errorf("failed to marshal IDs: %w", err)
	}

	cachePath := filepath.Join(getSessionDir(), resultsCacheFile)
	return os.WriteFile(cachePath, data, 0600)
}

// GetCachedResult retrieves a single ID by number (1-indexed)
func GetCachedResult(num int) (string, error) {
	cachePath := filepath.Join(getSessionDir(), resultsCacheFile)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no recent search results")
		}
		return "", fmt.Errorf("failed to read cache: %w", err)
	}

	var ids []string
	if err := json.Unmarshal(data, &ids); err != nil {
		return "", fmt.Errorf("invalid cache format")
	}

	if num < 1 || num > len(ids) {
		return "", fmt.Errorf("result number out of range (1-%d)", len(ids))
	}

	return ids[num-1], nil
}

// ============================================================================
// Deleted Cache (for undo)
// ============================================================================

// CacheDeleted saves deleted item IDs for undo
func CacheDeleted(ids []string) error {
	if err := ensureSessionDir(); err != nil {
		return err
	}

	data, err := json.Marshal(ids)
	if err != nil {
		return fmt.Errorf("failed to marshal IDs: %w", err)
	}

	cachePath := filepath.Join(getSessionDir(), deletedCacheFile)
	return os.WriteFile(cachePath, data, 0600)
}

// GetDeleted retrieves deleted IDs for undo (count <= 0 returns all)
func GetDeleted(count int) ([]string, error) {
	cachePath := filepath.Join(getSessionDir(), deletedCacheFile)

	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no recent deletion found")
		}
		return nil, fmt.Errorf("failed to read cache: %w", err)
	}

	var ids []string
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil, fmt.Errorf("invalid cache format")
	}

	if len(ids) == 0 {
		return nil, fmt.Errorf("no items to restore")
	}

	// Return requested count (or all if count <= 0)
	if count <= 0 || count > len(ids) {
		return ids, nil
	}

	return ids[:count], nil
}
