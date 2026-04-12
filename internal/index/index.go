package index

import (
	"time"

	"github.com/obsy-cli/obsy/internal/parser"
)

// FileEntry holds all indexed data for a single .md file.
type FileEntry struct {
	Path     string         // relative to vault root
	Mtime    int64          // Unix nanoseconds
	Links    []parser.Link  // all wikilinks (raw)
	Tags     []string       // frontmatter tags + inline tags
	Aliases  []string       // from frontmatter aliases:
	Props    map[string]any // all frontmatter fields
	Tasks    []parser.Task
	Headings []parser.Heading
}

// Index is the in-memory vault index.
type Index struct {
	VaultRoot string
	Files     map[string]*FileEntry // key: path relative to vault root
	ScannedAt time.Time             // last full scan
	UpdatedAt time.Time             // last incremental update (== ScannedAt if no incremental yet)
}

func New(vaultRoot string) *Index {
	return &Index{
		VaultRoot: vaultRoot,
		Files:     make(map[string]*FileEntry),
	}
}
