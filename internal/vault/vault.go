package vault

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Vault represents a discovered Obsidian vault.
type Vault struct {
	Root string // absolute path
}

var ErrNotFound = errors.New("vault not found: no .obsidian/ directory found between cwd and $HOME")

// Discover resolves the vault root.
// Priority: explicit path > cwd > walk up to $HOME looking for .obsidian/.
func Discover(explicit string) (*Vault, error) {
	if explicit != "" {
		abs, err := filepath.Abs(explicit)
		if err != nil {
			return nil, err
		}
		return &Vault{Root: abs}, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Check cwd first, then walk up to $HOME.
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".obsidian")); err == nil {
			return &Vault{Root: dir}, nil
		}
		if dir == home {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // filesystem root
		}
		dir = parent
	}

	// Also accept cwd if it contains .md files (vault without .obsidian/).
	entries, _ := os.ReadDir(cwd)
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			return &Vault{Root: cwd}, nil
		}
	}

	return nil, ErrNotFound
}

// Files returns all .md file paths relative to the vault root,
// skipping hidden paths, .obsidian/, .git/, and .trash/.
func (v *Vault) Files() ([]string, error) {
	var files []string
	err := filepath.WalkDir(v.Root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()

		// Skip hidden files and directories, and known excluded dirs.
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			switch name {
			case "img": // non-excluded, but skip nothing extra
			}
			return nil
		}

		// Skip symlinks.
		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		if strings.HasSuffix(name, ".md") {
			rel, err := filepath.Rel(v.Root, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}

// IsExcluded reports whether a relative path should be skipped during scanning.
func IsExcluded(rel string) bool {
	parts := strings.Split(rel, string(filepath.Separator))
	for _, p := range parts {
		if strings.HasPrefix(p, ".") {
			return true
		}
	}
	return false
}
