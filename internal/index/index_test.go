package index_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/obsy-cli/obsy/internal/index"
	"github.com/obsy-cli/obsy/internal/vault"
)

const testVault = "../../testdata/vault"

func openTestVault(t *testing.T) (*vault.Vault, *index.Index) {
	t.Helper()
	v, err := vault.Discover(testVault)
	if err != nil {
		t.Fatalf("discover vault: %v", err)
	}
	idx, err := index.Full(v)
	if err != nil {
		t.Fatalf("full scan: %v", err)
	}
	return v, idx
}

// copyTestVault copies testdata/vault into a fresh t.TempDir() and returns
// a vault and full index for the copy. Use this for tests that need to mutate
// the vault (create/delete files) to avoid races with other packages that read
// the real testdata/vault concurrently during `go test ./...`.
func copyTestVault(t *testing.T) (*vault.Vault, *index.Index) {
	t.Helper()
	src, err := vault.Discover(testVault)
	if err != nil {
		t.Fatalf("discover source vault: %v", err)
	}
	dst := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dst, ".obsidian"), 0o755); err != nil {
		t.Fatal(err)
	}
	files, err := src.Files()
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range files {
		dstAbs := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(dstAbs), 0o755); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(src.Root, rel))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dstAbs, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	v, err := vault.Discover(dst)
	if err != nil {
		t.Fatalf("discover temp vault: %v", err)
	}
	idx, err := index.Full(v)
	if err != nil {
		t.Fatalf("full scan temp vault: %v", err)
	}
	return v, idx
}

func TestFullScan_FileCount(t *testing.T) {
	_, idx := openTestVault(t)
	// testdata/vault has 13 .md files (no hidden, no .obsidian)
	if got := len(idx.Files); got != 13 {
		t.Errorf("file count = %d, want 13", got)
	}
}

func TestFullScan_KnownFilesPresent(t *testing.T) {
	_, idx := openTestVault(t)
	want := []string{
		"index.md",
		"note-a.md",
		"note-b.md",
		"dead-end.md",
		"broken.md",
		"aliases.md",
		"circular-a.md",
		"circular-b.md",
		"ambiguous.md",
		"sub/ambiguous.md",
		"sub/child.md",
		"sub/deep/buried.md",
	}
	for _, f := range want {
		if _, ok := idx.Files[f]; !ok {
			t.Errorf("expected file %q not in index", f)
		}
	}
}

func TestFullScan_Tags(t *testing.T) {
	_, idx := openTestVault(t)

	entry, ok := idx.Files["note-a.md"]
	if !ok {
		t.Fatal("note-a.md not indexed")
	}
	// note-a.md has frontmatter tags [science, reference] and inline #science (deduplicated)
	wantTags := map[string]bool{"science": true, "reference": true}
	for _, tag := range entry.Tags {
		delete(wantTags, tag)
	}
	if len(wantTags) > 0 {
		t.Errorf("missing tags in note-a.md: %v", wantTags)
	}
	// inline #science should NOT produce a duplicate
	count := 0
	for _, tag := range entry.Tags {
		if tag == "science" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("'science' appears %d times in tags, want 1", count)
	}
}

func TestFullScan_Aliases(t *testing.T) {
	_, idx := openTestVault(t)
	entry, ok := idx.Files["note-a.md"]
	if !ok {
		t.Fatal("note-a.md not indexed")
	}
	wantAliases := map[string]bool{"first-note": true, "nota": true}
	for _, a := range entry.Aliases {
		delete(wantAliases, a)
	}
	if len(wantAliases) > 0 {
		t.Errorf("missing aliases in note-a.md: %v", wantAliases)
	}
}

func TestFullScan_Tasks(t *testing.T) {
	_, idx := openTestVault(t)
	entry := idx.Files["note-a.md"]
	var todos, dones int
	for _, task := range entry.Tasks {
		if task.Done {
			dones++
		} else {
			todos++
		}
	}
	if todos != 2 {
		t.Errorf("note-a.md todos = %d, want 2", todos)
	}
	if dones != 1 {
		t.Errorf("note-a.md dones = %d, want 1", dones)
	}
}

func TestFullScan_LinksNotInCodeBlock(t *testing.T) {
	_, idx := openTestVault(t)
	entry := idx.Files["note-a.md"]
	// note-a.md has [[ignored-in-code-block]] and `[[also-ignored]]` — neither should appear
	for _, link := range entry.Links {
		if link.Raw == "ignored-in-code-block" || link.Raw == "also-ignored" {
			t.Errorf("link %q should have been ignored (inside code block)", link.Raw)
		}
	}
}

// --- Graph queries ---

func TestUnresolvedLinks(t *testing.T) {
	_, idx := openTestVault(t)
	broken := idx.UnresolvedLinks("")

	// Collect raw targets from broken.md
	brokenTargets := map[string]bool{}
	for _, b := range broken {
		if b.SourceFile == "broken.md" {
			brokenTargets[b.RawTarget] = true
		}
	}

	wantBroken := []string{"does-not-exist", "also-missing|with alias", "missing#heading", "nonexistent.pdf"}
	for _, w := range wantBroken {
		if !brokenTargets[w] {
			t.Errorf("expected broken link %q from broken.md, not found", w)
		}
	}

	// img/photo.png EXISTS on disk, so it must NOT appear as unresolved
	for _, b := range broken {
		if b.RawTarget == "img/photo.png" {
			t.Errorf("img/photo.png exists on disk but appeared as broken link")
		}
	}
}

func TestOrphans(t *testing.T) {
	_, idx := openTestVault(t)
	orphans := idx.Orphans("")

	// broken.md and table-links.md have no incoming links.
	// note-b.md is no longer an orphan: table-links.md links to it.
	orphanSet := make(map[string]bool)
	for _, o := range orphans {
		orphanSet[o] = true
	}
	for _, want := range []string{"broken.md", "table-links.md"} {
		if !orphanSet[want] {
			t.Errorf("expected %q to be an orphan", want)
		}
	}
	if orphanSet["note-b.md"] {
		t.Error("note-b.md should not be an orphan (table-links.md links to it)")
	}
	// index.md, note-a.md, dead-end.md are NOT orphans
	for _, notOrphan := range []string{"index.md", "note-a.md", "dead-end.md"} {
		if orphanSet[notOrphan] {
			t.Errorf("%q should not be an orphan", notOrphan)
		}
	}
}

func TestOrphans_IgnoreGlob(t *testing.T) {
	_, idx := openTestVault(t)
	orphans := idx.Orphans("note-b.md")
	for _, o := range orphans {
		if o == "note-b.md" {
			t.Error("note-b.md should have been excluded by ignore glob")
		}
	}
}

func TestDeadends(t *testing.T) {
	_, idx := openTestVault(t)
	deadends := idx.Deadends()
	deadSet := make(map[string]bool)
	for _, d := range deadends {
		deadSet[d] = true
	}

	// dead-end.md and note-b.md have no outgoing links
	for _, want := range []string{"dead-end.md", "note-b.md"} {
		if !deadSet[want] {
			t.Errorf("expected %q to be a dead-end", want)
		}
	}
	// index.md has outgoing links — not a dead-end
	if deadSet["index.md"] {
		t.Error("index.md should not be a dead-end")
	}
}

func TestBacklinksTo(t *testing.T) {
	_, idx := openTestVault(t)
	sources := idx.BacklinksTo("note-a.md")
	sourceSet := make(map[string]bool)
	for _, s := range sources {
		sourceSet[s] = true
	}
	// index.md links to note-a via [[note-a]], ![[note-a]], [[note-a|...]], [[note-a#Overview]]
	if !sourceSet["index.md"] {
		t.Error("expected index.md to be a backlink source of note-a.md")
	}
	// aliases.md links via [[nota]] which is an alias of note-a
	if !sourceSet["aliases.md"] {
		t.Error("expected aliases.md to be a backlink source of note-a.md (via alias)")
	}
}

func TestLinksFrom(t *testing.T) {
	_, idx := openTestVault(t)
	links := idx.LinksFrom("circular-a.md")
	if len(links) == 0 {
		t.Fatal("circular-a.md should have outgoing links")
	}
	found := false
	for _, l := range links {
		if l.Resolved == "circular-b.md" {
			found = true
		}
	}
	if !found {
		t.Error("circular-a.md should link to circular-b.md")
	}
}

func TestResolveFileArg(t *testing.T) {
	_, idx := openTestVault(t)

	tests := []struct {
		arg  string
		want string
		ok   bool
	}{
		{"note-a.md", "note-a.md", true},
		{"note-a", "note-a.md", true},
		{"nota", "note-a.md", true}, // alias
		{"sub/child", "sub/child.md", true},
		{"missing", "", false},
	}
	for _, tt := range tests {
		got, ok := idx.ResolveFileArg(tt.arg)
		if ok != tt.ok || got != tt.want {
			t.Errorf("ResolveFileArg(%q) = (%q, %v), want (%q, %v)", tt.arg, got, ok, tt.want, tt.ok)
		}
	}
}

// --- Incremental scan ---

func TestIncrementalScan_DetectsNewFile(t *testing.T) {
	// Use a temp copy so we don't race with other packages reading testdata/vault.
	v, idx := copyTestVault(t)

	newFile := filepath.Join(v.Root, "new-incremental.md")
	if err := os.WriteFile(newFile, []byte("# New\n\nContent."), 0o644); err != nil {
		t.Fatalf("create file: %v", err)
	}

	updated, err := index.Incremental(v, idx)
	if err != nil {
		t.Fatalf("incremental scan: %v", err)
	}
	if _, ok := updated.Files["new-incremental.md"]; !ok {
		t.Error("new file not picked up by incremental scan")
	}
}

func TestIncrementalScan_DetectsModifiedFile(t *testing.T) {
	v, idx := openTestVault(t)

	// Force the mtime of note-b.md to look stale in the index.
	entry := idx.Files["note-b.md"]
	originalMtime := entry.Mtime
	entry.Mtime = 0 // force re-parse
	idx.Files["note-b.md"] = entry

	updated, err := index.Incremental(v, idx)
	if err != nil {
		t.Fatalf("incremental scan: %v", err)
	}
	// After re-parse, mtime should be restored to real value.
	if updated.Files["note-b.md"].Mtime == 0 {
		t.Error("mtime not updated after incremental re-parse")
	}
	_ = originalMtime
}

func TestIncrementalScan_RemovesDeletedFile(t *testing.T) {
	v, idx := openTestVault(t)

	// Inject a phantom file into the index.
	idx.Files["phantom.md"] = &index.FileEntry{Path: "phantom.md", Mtime: time.Now().UnixNano()}

	updated, err := index.Incremental(v, idx)
	if err != nil {
		t.Fatalf("incremental scan: %v", err)
	}
	if _, ok := updated.Files["phantom.md"]; ok {
		t.Error("phantom file should have been removed by incremental scan")
	}
}

func TestFullScan_UnreadableFileWarnsToStderr(t *testing.T) {
	// Build a temp vault with one unreadable file.
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".obsidian"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "good.md"), []byte("# Good"), 0o644); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(root, "bad.md")
	if err := os.WriteFile(bad, []byte("# Bad"), 0o000); err != nil {
		t.Fatal(err)
	}
	// Restore permissions after test so t.TempDir cleanup can remove it.
	t.Cleanup(func() { os.Chmod(bad, 0o644) })

	// Running as root bypasses permission checks — skip in that case.
	if os.Getuid() == 0 {
		t.Skip("running as root: file permissions are not enforced")
	}

	v, err := vault.Discover(root)
	if err != nil {
		t.Fatal(err)
	}

	// Capture stderr.
	r, w, _ := os.Pipe()
	orig := os.Stderr
	os.Stderr = w

	idx, err := index.Full(v)

	w.Close()
	os.Stderr = orig
	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	if err != nil {
		t.Fatalf("Full scan returned error: %v", err)
	}
	if _, ok := idx.Files["good.md"]; !ok {
		t.Error("good.md should still be indexed despite bad.md failure")
	}
	if _, ok := idx.Files["bad.md"]; ok {
		t.Error("bad.md should have been skipped")
	}
	if !strings.Contains(buf.String(), "warning") {
		t.Errorf("expected warning on stderr, got: %q", buf.String())
	}
}
