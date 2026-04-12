package index

import (
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

func init() {
	// Register concrete types that yaml.v3 produces inside map[string]any.
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
	gob.Register(map[interface{}]interface{}{})
}

// VaultID returns the first 8 hex chars of SHA-256 of the vault's absolute path.
func VaultID(vaultRoot string) string {
	sum := sha256.Sum256([]byte(vaultRoot))
	return fmt.Sprintf("%x", sum[:4])
}

// CachePath returns the path to the cache file for the given vault root.
// Respects $XDG_CACHE_HOME; falls back to ~/.cache.
func CachePath(vaultRoot string) string {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".cache")
	}
	return filepath.Join(cacheDir, "obsy", VaultID(vaultRoot), "index.gob")
}

// Load reads the cache from disk.
// Returns nil (no error) if the cache is missing or corrupt — caller rebuilds.
func Load(vaultRoot string) *Index {
	path := CachePath(vaultRoot)
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var idx Index
	if err := gob.NewDecoder(f).Decode(&idx); err != nil {
		return nil
	}
	return &idx
}

// Save writes the index to the cache file, creating directories as needed.
func Save(idx *Index) error {
	path := CachePath(idx.VaultRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := gob.NewEncoder(f).Encode(idx); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()
	return os.Rename(tmp, path)
}
