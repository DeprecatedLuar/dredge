package storage

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/DeprecatedLuar/dredge/internal/crypto"
)

const (
	// Manifest file
	manifestFileName = "links.json"

	// Permissions
	manifestPermissions = 0600 // rw-------
	spawnedPermissions  = 0600 // rw-------
)

// LinkEntry represents a single link in the manifest
type LinkEntry struct {
	Path string `json:"path"` // Target path where symlink points (e.g., /home/user/.ssh/config)
	Hash string `json:"hash"` // SHA256 hash of spawned file content
}

// LinkManifest maps item IDs to link entries
type LinkManifest map[string]LinkEntry

// getManifestPath returns the path to links.json
func getManifestPath() (string, error) {
	dredgeDir, err := GetDredgeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dredgeDir, manifestFileName), nil
}

// LoadManifest reads and parses links.json, returns empty map if file doesn't exist
func LoadManifest() (LinkManifest, error) {
	manifestPath, err := getManifestPath()
	if err != nil {
		return nil, err
	}

	// If manifest doesn't exist, return empty map (not an error)
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return make(LinkManifest), nil
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest LinkManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return manifest, nil
}

// SaveManifest writes the manifest to links.json
func SaveManifest(manifest LinkManifest) error {
	manifestPath, err := getManifestPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, data, manifestPermissions); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// RemoveFromManifest removes an entry from the manifest and saves
func RemoveFromManifest(id string) error {
	manifest, err := LoadManifest()
	if err != nil {
		return err
	}

	delete(manifest, id)
	return SaveManifest(manifest)
}

// GetSpawnedPath returns the path to the spawned file for an item
func GetSpawnedPath(id string) (string, error) {
	dredgeDir, err := GetDredgeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dredgeDir, spawnedDirName, id), nil
}

// CreateSpawnedFile writes plain text content to .spawned/<id>
func CreateSpawnedFile(id, content string) error {
	spawnedPath, err := GetSpawnedPath(id)
	if err != nil {
		return err
	}

	// Ensure .spawned/ directory exists
	spawnedDir := filepath.Dir(spawnedPath)
	if err := os.MkdirAll(spawnedDir, dirPermissions); err != nil {
		return fmt.Errorf("failed to create .spawned directory: %w", err)
	}

	// Write plain text content
	if err := os.WriteFile(spawnedPath, []byte(content), spawnedPermissions); err != nil {
		return fmt.Errorf("failed to write spawned file: %w", err)
	}

	return nil
}

// RemoveSpawnedFile deletes the spawned file for an item
func RemoveSpawnedFile(id string) error {
	spawnedPath, err := GetSpawnedPath(id)
	if err != nil {
		return err
	}

	if err := os.Remove(spawnedPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove spawned file: %w", err)
	}

	return nil
}

// hashFile computes SHA256 hash of a file's content
func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", hash), nil
}

// syncItemIfNeeded checks if spawned file changed and syncs to encrypted item if needed
func syncItemIfNeeded(id, password string) error {
	manifest, err := LoadManifest()
	if err != nil {
		return err
	}

	entry, exists := manifest[id]
	if !exists {
		return nil // Not linked, nothing to sync
	}

	spawnedPath, err := GetSpawnedPath(id)
	if err != nil {
		return err
	}

	// Compute current hash of spawned file
	currentHash, err := hashFile(spawnedPath)
	if err != nil {
		return fmt.Errorf("failed to hash spawned file: %w", err)
	}

	// If hash matches, no changes detected
	if currentHash == entry.Hash {
		return nil
	}

	// Hash mismatch → spawned file was manually edited
	// Read spawned content
	spawnedContent, err := os.ReadFile(spawnedPath)
	if err != nil {
		return fmt.Errorf("failed to read spawned file: %w", err)
	}

	// Load encrypted item (inline to avoid recursion with ReadItem)
	itemPath, err := GetItemPath(id)
	if err != nil {
		return fmt.Errorf("failed to get item path: %w", err)
	}

	encryptedData, err := os.ReadFile(itemPath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted item: %w", err)
	}

	decryptedData, err := crypto.Decrypt(encryptedData, password)
	if err != nil {
		return fmt.Errorf("failed to decrypt item: %w", err)
	}

	var item Item
	if err := toml.Unmarshal(decryptedData, &item); err != nil {
		return fmt.Errorf("failed to decode TOML: %w", err)
	}

	// Update item with spawned content (spawned wins)
	item.Content.Text = string(spawnedContent)
	item.Modified = time.Now()

	// Re-encrypt and save (inline to avoid triggering UpdateItem's linked logic)
	var buf bytes.Buffer
	encoder := toml.NewEncoder(&buf)
	if err := encoder.Encode(&item); err != nil {
		return fmt.Errorf("failed to encode item: %w", err)
	}

	encryptedData, err = crypto.Encrypt(buf.Bytes(), password)
	if err != nil {
		return fmt.Errorf("failed to encrypt item: %w", err)
	}

	if err := os.WriteFile(itemPath, encryptedData, itemFilePermissions); err != nil {
		return fmt.Errorf("failed to write encrypted item: %w", err)
	}

	// Update manifest hash
	entry.Hash = currentHash
	manifest[id] = entry
	if err := SaveManifest(manifest); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	return nil
}

// UpdateSpawnedFile updates the spawned file content (called after edit)
func UpdateSpawnedFile(id, content string) error {
	return CreateSpawnedFile(id, content)
}

// UpdateManifestHash recomputes and updates the hash for a linked item
func UpdateManifestHash(id string) error {
	manifest, err := LoadManifest()
	if err != nil {
		return err
	}

	entry, exists := manifest[id]
	if !exists {
		return nil // Not linked, nothing to update
	}

	spawnedPath, err := GetSpawnedPath(id)
	if err != nil {
		return err
	}
	newHash, err := hashFile(spawnedPath)
	if err != nil {
		return fmt.Errorf("failed to hash spawned file: %w", err)
	}

	entry.Hash = newHash
	manifest[id] = entry
	return SaveManifest(manifest)
}

// IsLinked checks if an item has an active link
func IsLinked(id string) bool {
	manifest, err := LoadManifest()
	if err != nil {
		return false
	}

	_, exists := manifest[id]
	return exists
}

// GetLinkedPath returns the target path from manifest (empty string if not linked)
func GetLinkedPath(id string) (string, bool) {
	manifest, err := LoadManifest()
	if err != nil {
		return "", false
	}

	entry, exists := manifest[id]
	if !exists {
		return "", false
	}

	return entry.Path, true
}

// Link creates a symlink from targetPath to .spawned/<id>
func Link(id, targetPath string, force bool) error {
	// Check if already linked
	if IsLinked(id) {
		existingPath, _ := GetLinkedPath(id)
		return fmt.Errorf("item %s already linked to %s", id, existingPath)
	}

	// Load and validate item
	password, err := crypto.GetPasswordWithVerification()
	if err != nil {
		return err
	}

	item, err := ReadItem(id, password)
	if err != nil {
		return fmt.Errorf("failed to load item: %w", err)
	}

	// Only text items can be linked
	if item.Type != TypeText {
		return fmt.Errorf("cannot link binary items (use text items only)")
	}

	// Validate target path is absolute
	if !filepath.IsAbs(targetPath) {
		return fmt.Errorf("target path must be absolute")
	}

	// Check if target file exists
	if _, err := os.Lstat(targetPath); err == nil {
		if !force {
			return fmt.Errorf("file already exists at %s (use --force to overwrite)", targetPath)
		}
		// Remove existing file if --force
		if err := os.Remove(targetPath); err != nil {
			return fmt.Errorf("failed to remove existing file: %w", err)
		}
	}

	// Check parent directory exists
	parentDir := filepath.Dir(targetPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		return fmt.Errorf("parent directory does not exist: %s", parentDir)
	}

	// Create spawned file with decrypted content
	if err := CreateSpawnedFile(id, item.Content.Text); err != nil {
		return err
	}

	// Compute hash of spawned file
	spawnedPath, err := GetSpawnedPath(id)
	if err != nil {
		return err
	}
	hash, err := hashFile(spawnedPath)
	if err != nil {
		RemoveSpawnedFile(id) // Cleanup on failure
		return fmt.Errorf("failed to hash spawned file: %w", err)
	}

	// Create symlink
	if err := os.Symlink(spawnedPath, targetPath); err != nil {
		RemoveSpawnedFile(id) // Cleanup on failure
		return fmt.Errorf("failed to create symlink: %w", err)
	}

	// Add to manifest
	manifest, err := LoadManifest()
	if err != nil {
		os.Remove(targetPath)    // Cleanup symlink
		RemoveSpawnedFile(id)    // Cleanup spawned file
		return err
	}

	manifest[id] = LinkEntry{
		Path: targetPath,
		Hash: hash,
	}

	if err := SaveManifest(manifest); err != nil {
		os.Remove(targetPath)    // Cleanup symlink
		RemoveSpawnedFile(id)    // Cleanup spawned file
		return err
	}

	return nil
}

// Unlink removes the symlink, spawned file, and manifest entry
// Handles broken links gracefully (item deleted, symlink missing, etc.)
// Only errors if nothing exists to clean up
func Unlink(id string) error {
	// Check if linked
	if !IsLinked(id) {
		return fmt.Errorf("item %s is not linked", id)
	}

	// Get target path from manifest
	targetPath, exists := GetLinkedPath(id)
	if !exists {
		return fmt.Errorf("manifest entry not found for %s", id)
	}

	// Track if we cleaned anything
	cleanedAnything := false

	// Check if encrypted item exists
	itemExists, err := ItemExists(id)
	if err == nil && itemExists {
		// Item exists - sync spawned changes before unlinking
		password, err := crypto.GetPasswordWithVerification()
		if err != nil {
			return err
		}

		// ReadItem will automatically sync if spawned file changed
		if _, err := ReadItem(id, password); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to sync before unlink: %v\n", err)
			// Continue with unlink anyway
		}
	}
	// If item doesn't exist, skip password/sync (nothing to sync to)

	// Remove symlink (silent if missing)
	if err := os.Remove(targetPath); err == nil {
		cleanedAnything = true
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: failed to remove symlink at %s: %v\n", targetPath, err)
	}

	// Remove spawned file (silent if missing)
	spawnedPath, err := GetSpawnedPath(id)
	if err == nil {
		if err := os.Remove(spawnedPath); err == nil {
			cleanedAnything = true
		} else if !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Warning: failed to remove spawned file: %v\n", err)
		}
	}

	// Remove from manifest
	if err := RemoveFromManifest(id); err != nil {
		return fmt.Errorf("failed to update manifest: %w", err)
	}

	// Error only if we cleaned nothing (everything was already gone)
	if !cleanedAnything {
		return fmt.Errorf("nothing to clean up for item %s (symlink and spawned file already removed)", id)
	}

	return nil
}
