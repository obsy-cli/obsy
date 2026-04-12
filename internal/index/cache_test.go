package index

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/obsy-cli/obsy/internal/parser"
)

func TestVaultID_Deterministic(t *testing.T) {
	a := VaultID("/home/user/vault")
	b := VaultID("/home/user/vault")
	if a != b {
		t.Errorf("VaultID not deterministic: %q vs %q", a, b)
	}
}

func TestVaultID_Distinct(t *testing.T) {
	a := VaultID("/vault/a")
	b := VaultID("/vault/b")
	if a == b {
		t.Errorf("different paths produced same VaultID: %q", a)
	}
}

func TestVaultID_Length(t *testing.T) {
	id := VaultID("/any/path")
	if len(id) != 8 {
		t.Errorf("VaultID length = %d, want 8", len(id))
	}
}

func TestCachePath_XDG(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	path := CachePath("/my/vault")
	if !filepath.IsAbs(path) {
		t.Errorf("CachePath not absolute: %q", path)
	}
	// Must be under the XDG dir.
	rel, err := filepath.Rel(tmp, path)
	if err != nil || len(rel) == 0 || rel[0] == '.' {
		t.Errorf("CachePath %q not under XDG_CACHE_HOME %q", path, tmp)
	}
	// Must end with index.gob.
	if filepath.Base(path) != "index.gob" {
		t.Errorf("CachePath base = %q, want index.gob", filepath.Base(path))
	}
}

func TestLoad_MissingFile(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	if got := Load("/nonexistent/vault"); got != nil {
		t.Error("Load of missing cache should return nil")
	}
}

func TestLoad_CorruptFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	// Write garbage to the cache path.
	path := CachePath("/my/vault")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("not gob data"), 0o644); err != nil {
		t.Fatal(err)
	}

	if got := Load("/my/vault"); got != nil {
		t.Error("Load of corrupt cache should return nil")
	}
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	original := &Index{
		VaultRoot: "/my/vault",
		ScannedAt: time.Now().Truncate(time.Second),
		Files: map[string]*FileEntry{
			"note.md": {
				Path:  "note.md",
				Mtime: 12345678,
				Tags:  []string{"science", "reference"},
				Links: []parser.Link{
					{Raw: "other", IsEmbed: false, Line: 3},
				},
				Tasks: []parser.Task{
					{Text: "Buy milk", Done: false, Line: 5},
				},
				Headings: []parser.Heading{
					{Level: 2, Text: "Overview", Line: 1},
				},
				Aliases: []string{"my-note"},
				Props: map[string]any{
					"title":  "My Note",
					"number": 42,
					"list":   []interface{}{"a", "b"},
				},
			},
		},
	}

	if err := Save(original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded := Load("/my/vault")
	if loaded == nil {
		t.Fatal("Load returned nil after Save")
	}

	if loaded.VaultRoot != original.VaultRoot {
		t.Errorf("VaultRoot = %q, want %q", loaded.VaultRoot, original.VaultRoot)
	}

	entry, ok := loaded.Files["note.md"]
	if !ok {
		t.Fatal("note.md missing from loaded index")
	}
	if entry.Mtime != 12345678 {
		t.Errorf("Mtime = %d, want 12345678", entry.Mtime)
	}
	if len(entry.Tags) != 2 {
		t.Errorf("Tags len = %d, want 2", len(entry.Tags))
	}
	if len(entry.Links) != 1 || entry.Links[0].Raw != "other" {
		t.Errorf("Links = %+v", entry.Links)
	}
	if len(entry.Tasks) != 1 || entry.Tasks[0].Text != "Buy milk" {
		t.Errorf("Tasks = %+v", entry.Tasks)
	}
	if len(entry.Headings) != 1 || entry.Headings[0].Text != "Overview" {
		t.Errorf("Headings = %+v", entry.Headings)
	}
	if entry.Props["title"] != "My Note" {
		t.Errorf("Props[title] = %v, want 'My Note'", entry.Props["title"])
	}
	// []interface{} must survive gob round-trip.
	list, ok := entry.Props["list"].([]interface{})
	if !ok || len(list) != 2 {
		t.Errorf("Props[list] = %T %v, want []interface{} of len 2", entry.Props["list"], entry.Props["list"])
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)

	idx := &Index{VaultRoot: "/my/vault", Files: map[string]*FileEntry{}}
	if err := Save(idx); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// No .tmp file should remain after a successful save.
	pattern := filepath.Join(tmp, "**", "*.tmp")
	matches, _ := filepath.Glob(pattern)
	if len(matches) > 0 {
		t.Errorf("leftover .tmp files after Save: %v", matches)
	}
}
