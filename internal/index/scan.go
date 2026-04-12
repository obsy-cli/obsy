package index

import (
	"os"
	"path/filepath"
	"time"

	"github.com/obsy-cli/obsy/internal/parser"
	"github.com/obsy-cli/obsy/internal/vault"
)

// Full scans all .md files in the vault and builds a fresh index.
func Full(v *vault.Vault) (*Index, error) {
	files, err := v.Files()
	if err != nil {
		return nil, err
	}
	idx := New(v.Root)
	for _, rel := range files {
		entry, err := parseFile(v.Root, rel)
		if err != nil {
			continue // skip unreadable files
		}
		idx.Files[rel] = entry
	}
	now := time.Now()
	idx.ScannedAt = now
	idx.UpdatedAt = now
	return idx, nil
}

// Incremental updates an existing index with changed/new files and removes deleted ones.
func Incremental(v *vault.Vault, existing *Index) (*Index, error) {
	files, err := v.Files()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool, len(files))
	for _, rel := range files {
		seen[rel] = true

		absPath := filepath.Join(v.Root, rel)
		fi, err := os.Stat(absPath)
		if err != nil {
			continue
		}
		mtime := fi.ModTime().UnixNano()

		if entry, ok := existing.Files[rel]; ok && entry.Mtime == mtime {
			continue // unchanged
		}

		entry, err := parseFile(v.Root, rel)
		if err != nil {
			continue
		}
		existing.Files[rel] = entry
	}

	// Remove deleted files.
	for rel := range existing.Files {
		if !seen[rel] {
			delete(existing.Files, rel)
		}
	}

	existing.UpdatedAt = time.Now()
	return existing, nil
}

// LoadOrBuild loads the cache and runs an incremental update, or does a full
// scan if noCache is true or no cache exists.
func LoadOrBuild(v *vault.Vault, noCache bool) (*Index, error) {
	if !noCache {
		if cached := Load(v.Root); cached != nil {
			return Incremental(v, cached)
		}
	}
	idx, err := Full(v)
	if err != nil {
		return nil, err
	}
	if !noCache {
		_ = Save(idx) // best-effort; ignore save errors
	}
	return idx, nil
}

// parseFile reads and parses a single .md file into a FileEntry.
func parseFile(vaultRoot, rel string) (*FileEntry, error) {
	absPath := filepath.Join(vaultRoot, rel)
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	fm, body := parser.ParseFrontmatter(content)
	links := parser.ParseLinks(body)
	inlineTags := parser.ParseInlineTags(body)
	tasks := parser.ParseTasks(body)
	headings := parser.ParseHeadings(body)

	// Merge frontmatter tags with inline tags (deduplicated).
	tags := mergeTags(fm.Tags, inlineTags)

	return &FileEntry{
		Path:     rel,
		Mtime:    fi.ModTime().UnixNano(),
		Links:    links,
		Tags:     tags,
		Aliases:  fm.Aliases,
		Props:    fm.Props,
		Tasks:    tasks,
		Headings: headings,
	}, nil
}

func mergeTags(a, b []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, t := range a {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	for _, t := range b {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}
