package index

import (
	"fmt"
	"io"
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
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", rel, err)
			continue
		}
		idx.Files[rel] = entry
	}
	now := time.Now()
	idx.ScannedAt = now
	idx.UpdatedAt = now
	return idx, nil
}

// Incremental updates an existing index with changed/new files and removes deleted ones.
// Uses FilesWithMtime so mtime comparison costs no extra stat calls beyond the walk.
func Incremental(v *vault.Vault, existing *Index) (*Index, error) {
	infos, err := v.FilesWithMtime()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool, len(infos))
	for _, info := range infos {
		seen[info.Path] = true
		mtime := info.Mtime.UnixNano()

		if entry, ok := existing.Files[info.Path]; ok && entry.Mtime == mtime {
			continue // unchanged
		}

		entry, err := parseFile(v.Root, info.Path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", info.Path, err)
			continue
		}
		existing.Files[info.Path] = entry
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
	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	fm, body, fmErr := parser.ParseFrontmatter(content)
	if fmErr != nil {
		fmt.Fprintf(os.Stderr, "warning: %s: %v\n", rel, fmErr)
	}
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
