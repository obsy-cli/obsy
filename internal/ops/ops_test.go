package ops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/obsy-cli/obsy/internal/index"
	"github.com/obsy-cli/obsy/internal/vault"
)

const testVaultPath = "../../testdata/vault"

// --- rewriteWikilinks (pure) ---

func TestRewriteWikilinks_BasicRename(t *testing.T) {
	cases := []struct {
		name             string
		content          string
		oldBase, newBase string
		oldRel, newRel   string
		want             string
	}{
		{
			name:    "bare link",
			content: "See [[note-a]] for details.",
			oldBase: "note-a", newBase: "note-b",
			oldRel: "note-a.md", newRel: "note-b.md",
			want: "See [[note-b]] for details.",
		},
		{
			name:    "link with display alias",
			content: "See [[note-a|My Note]] for details.",
			oldBase: "note-a", newBase: "note-b",
			oldRel: "note-a.md", newRel: "note-b.md",
			want: "See [[note-b|My Note]] for details.",
		},
		{
			name:    "link with heading",
			content: "See [[note-a#Overview]] here.",
			oldBase: "note-a", newBase: "note-b",
			oldRel: "note-a.md", newRel: "note-b.md",
			want: "See [[note-b#Overview]] here.",
		},
		{
			name:    "link with heading and alias",
			content: "[[note-a#Overview|click here]]",
			oldBase: "note-a", newBase: "note-b",
			oldRel: "note-a.md", newRel: "note-b.md",
			want: "[[note-b#Overview|click here]]",
		},
		{
			name:    "embed",
			content: "![[note-a]]",
			oldBase: "note-a", newBase: "note-b",
			oldRel: "note-a.md", newRel: "note-b.md",
			want: "![[note-b]]",
		},
		{
			name:    "multiple occurrences",
			content: "[[note-a]] and again [[note-a|second]]",
			oldBase: "note-a", newBase: "note-b",
			oldRel: "note-a.md", newRel: "note-b.md",
			want: "[[note-b]] and again [[note-b|second]]",
		},
		{
			name:    "no match — unrelated link unchanged",
			content: "[[other-note]] stays.",
			oldBase: "note-a", newBase: "note-b",
			oldRel: "note-a.md", newRel: "note-b.md",
			want: "[[other-note]] stays.",
		},
		{
			// Renaming sub/note-a.md → sub/note-b.md: path-qualified link updates correctly.
			name:    "path-qualified link — full relative path match",
			content: "[[sub/note-a]] here.",
			oldBase: "note-a", newBase: "note-b",
			oldRel: "sub/note-a.md", newRel: "sub/note-b.md",
			want: "[[sub/note-b]] here.",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := rewriteWikilinks(tc.content, tc.oldBase, tc.newBase, tc.oldRel, tc.newRel)
			if got != tc.want {
				t.Errorf("got  %q\nwant %q", got, tc.want)
			}
		})
	}
}

// TestRewriteWikilinks_BasenameCollision documents that renaming a root-level file
// incorrectly rewrites path-qualified links that happen to share the same basename
// but refer to a different file (e.g. [[sub/note-a]] when renaming note-a.md).
// This is a known limitation of the basename-matching heuristic.
func TestRewriteWikilinks_BasenameCollision(t *testing.T) {
	// [[sub/note-a]] points to sub/note-a.md, not note-a.md (root).
	// Renaming note-a.md → note-b.md should ideally leave [[sub/note-a]] alone,
	// but the basename match rewrites it to [[note-b]] (losing the folder path).
	got := rewriteWikilinks("[[sub/note-a]]", "note-a", "note-b", "note-a.md", "note-b.md")
	// Document actual (imperfect) behaviour so any future fix is visible.
	if got != "[[note-b]]" {
		t.Logf("basename collision behaviour changed: got %q (was [[note-b]])", got)
	}
}

func TestRewriteWikilinks_EscapedPipe(t *testing.T) {
	// Obsidian escapes | as \| inside markdown table cells.
	// The path target must be resolved correctly and \| must be preserved in the output.
	cases := []struct {
		content          string
		oldBase, newBase string
		oldRel, newRel   string
		want             string
	}{
		// Bare name in table cell — target is note-a, display text preserved.
		{`[[note-a\|alias]]`, "note-a", "note-b", "note-a.md", "note-b.md", `[[note-b\|alias]]`},
		// Path-qualified in table cell — oldRel must match the qualified path for correct rewrite.
		{`[[sub/note-a\|alias]]`, "note-a", "note-b", "sub/note-a.md", "sub/note-b.md", `[[sub/note-b\|alias]]`},
		// Non-matching link — must be untouched.
		{`[[other\|alias]]`, "note-a", "note-b", "note-a.md", "note-b.md", `[[other\|alias]]`},
	}
	for _, tc := range cases {
		got := rewriteWikilinks(tc.content, tc.oldBase, tc.newBase, tc.oldRel, tc.newRel)
		if got != tc.want {
			t.Errorf("rewriteWikilinks(%q)\n got  %q\n want %q", tc.content, got, tc.want)
		}
	}
}

func TestRewriteWikilinks_SpacedLink(t *testing.T) {
	// Links with internal spaces (non-standard but possible via manual edits).
	// The suffix offset bug would corrupt text after the path component; verify it doesn't.
	// Leading spaces inside [[ are normalised away; suffix starts where the raw path ended.
	cases := []struct {
		content string
		want    string
	}{
		// Leading/trailing spaces around name → name normalised, suffix preserved.
		{"[[ note-a ]]", "[[note-b]]"},
		{"[[ note-a | alias ]]", "[[note-b| alias ]]"},
		{"[[ note-a #heading | alias ]]", "[[note-b#heading | alias ]]"},
	}
	for _, tc := range cases {
		got := rewriteWikilinks(tc.content, "note-a", "note-b", "note-a.md", "note-b.md")
		if got != tc.want {
			t.Errorf("rewriteWikilinks(%q)\n got  %q\n want %q", tc.content, got, tc.want)
		}
	}
}

func TestRewriteWikilinks_UnclosedBracket(t *testing.T) {
	// Unclosed [[ should be passed through verbatim.
	content := "start [[unclosed"
	got := rewriteWikilinks(content, "unclosed", "new", "unclosed.md", "new.md")
	if got != content {
		t.Errorf("got %q, want unchanged %q", got, content)
	}
}

func TestRewriteWikilinks_NoLinks(t *testing.T) {
	content := "Plain text without any links."
	got := rewriteWikilinks(content, "note-a", "note-b", "note-a.md", "note-b.md")
	if got != content {
		t.Errorf("got %q, want unchanged", got)
	}
}

// --- Search ---

func TestSearch_BasicMatch(t *testing.T) {
	v, err := vault.Discover(testVaultPath)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	results, err := Search(v, "science", "", 0, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for 'science'")
	}
	found := false
	for _, r := range results {
		if r.Path == "note-a.md" {
			found = true
		}
	}
	if !found {
		t.Error("expected note-a.md in results for 'science'")
	}
}

func TestSearch_CaseSensitive(t *testing.T) {
	v, err := vault.Discover(testVaultPath)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	// "SCIENCE" all caps should match nothing in case-sensitive mode.
	results, err := Search(v, "SCIENCE", "", 0, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for all-caps SCIENCE (case-sensitive), got %d", len(results))
	}
}

func TestSearch_CaseInsensitive(t *testing.T) {
	v, err := vault.Discover(testVaultPath)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	lower, err := Search(v, "science", "", 0, false, false)
	if err != nil {
		t.Fatal(err)
	}
	upper, err := Search(v, "SCIENCE", "", 0, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(lower) != len(upper) {
		t.Errorf("case-insensitive: 'science' got %d results, 'SCIENCE' got %d", len(lower), len(upper))
	}
}

func TestSearch_PathFilter(t *testing.T) {
	v, err := vault.Discover(testVaultPath)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	// Filter to sub/ — note-a.md (root) should not appear.
	results, err := Search(v, "note", "sub/", 0, false, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if !strings.HasPrefix(r.Path, "sub/") {
			t.Errorf("path %q is outside sub/ filter", r.Path)
		}
	}
}

func TestSearch_Limit(t *testing.T) {
	v, err := vault.Discover(testVaultPath)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	results, err := Search(v, "note", "", 2, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) > 2 {
		t.Errorf("limit=2 but got %d results", len(results))
	}
}

func TestSearch_ContextLines(t *testing.T) {
	v, err := vault.Discover(testVaultPath)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	results, err := Search(v, "science", "", 0, true, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.Path == "note-a.md" {
			if len(r.Context) == 0 {
				t.Error("expected context lines for note-a.md, got none")
			}
			return
		}
	}
	t.Error("note-a.md not found in search results")
}

func TestSearch_NoMatch(t *testing.T) {
	v, err := vault.Discover(testVaultPath)
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	results, err := Search(v, "zzznomatchzzz", "", 0, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// --- Move ---

func setupTempVault(t *testing.T) (*vault.Vault, *index.Index) {
	t.Helper()
	src := testVaultPath
	dst := t.TempDir()

	// Copy .obsidian marker so Discover works.
	if err := os.MkdirAll(filepath.Join(dst, ".obsidian"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Copy all md files from the real testdata vault.
	v, err := vault.Discover(src)
	if err != nil {
		t.Fatalf("discover source vault: %v", err)
	}
	files, err := v.Files()
	if err != nil {
		t.Fatal(err)
	}
	for _, rel := range files {
		srcAbs := filepath.Join(src, rel)
		dstAbs := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(dstAbs), 0o755); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(srcAbs)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dstAbs, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// Also copy non-md assets (e.g. img/photo.png).
	img := filepath.Join(src, "img")
	if fi, err := os.Stat(img); err == nil && fi.IsDir() {
		dstImg := filepath.Join(dst, "img")
		os.MkdirAll(dstImg, 0o755)
		entries, _ := os.ReadDir(img)
		for _, e := range entries {
			data, _ := os.ReadFile(filepath.Join(img, e.Name()))
			os.WriteFile(filepath.Join(dstImg, e.Name()), data, 0o644)
		}
	}

	tmpV, err := vault.Discover(dst)
	if err != nil {
		t.Fatalf("discover temp vault: %v", err)
	}
	idx, err := index.Full(tmpV)
	if err != nil {
		t.Fatalf("index temp vault: %v", err)
	}
	return tmpV, idx
}

func TestMove_Basic(t *testing.T) {
	v, idx := setupTempVault(t)

	result, err := Move(v, idx, "dead-end.md", "archived/dead-end.md")
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	if result.OldPath != "dead-end.md" {
		t.Errorf("OldPath = %q, want dead-end.md", result.OldPath)
	}
	if result.NewPath != "archived/dead-end.md" {
		t.Errorf("NewPath = %q, want archived/dead-end.md", result.NewPath)
	}

	// File must exist at new location.
	if _, err := os.Stat(filepath.Join(v.Root, "archived/dead-end.md")); err != nil {
		t.Error("new file does not exist after move")
	}
	// File must be gone from old location.
	if _, err := os.Stat(filepath.Join(v.Root, "dead-end.md")); err == nil {
		t.Error("old file still exists after move")
	}
	// Index must be updated.
	if _, ok := idx.Files["archived/dead-end.md"]; !ok {
		t.Error("index not updated with new path")
	}
	if _, ok := idx.Files["dead-end.md"]; ok {
		t.Error("old path still in index")
	}
}

func TestMove_RewritesBacklinks(t *testing.T) {
	v, idx := setupTempVault(t)

	// index.md links to note-a.md. Moving note-a should rewrite those links.
	result, err := Move(v, idx, "note-a.md", "notes/note-a.md")
	if err != nil {
		t.Fatalf("move: %v", err)
	}

	updated := make(map[string]bool)
	for _, u := range result.Updated {
		updated[u] = true
	}
	if !updated["index.md"] {
		t.Error("index.md should have been rewritten (it links to note-a)")
	}
}

func TestMove_SameSourceDest(t *testing.T) {
	v, idx := setupTempVault(t)
	_, err := Move(v, idx, "note-a.md", "note-a.md")
	if err == nil {
		t.Error("expected error when src == dst")
	}
}

func TestMove_DestExists(t *testing.T) {
	v, idx := setupTempVault(t)
	_, err := Move(v, idx, "note-a.md", "note-b.md")
	if err == nil {
		t.Error("expected error when destination already exists in index")
	}
}

func TestMove_FileNotFound(t *testing.T) {
	v, idx := setupTempVault(t)
	_, err := Move(v, idx, "no-such-file.md", "somewhere.md")
	if err == nil {
		t.Error("expected error for non-existent source")
	}
}

// --- Rename ---

func TestRename_Basic(t *testing.T) {
	v, idx := setupTempVault(t)

	result, err := Rename(v, idx, "dead-end.md", "archived-dead-end")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if result.NewPath != "archived-dead-end.md" {
		t.Errorf("NewPath = %q, want archived-dead-end.md", result.NewPath)
	}
}

func TestRename_PreservesExtension(t *testing.T) {
	v, idx := setupTempVault(t)

	// Providing .md explicitly should not double it.
	result, err := Rename(v, idx, "dead-end.md", "new-name.md")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if result.NewPath != "new-name.md" {
		t.Errorf("NewPath = %q, want new-name.md", result.NewPath)
	}
}

func TestRename_StaysInSameDir(t *testing.T) {
	v, idx := setupTempVault(t)

	result, err := Rename(v, idx, "sub/child.md", "child-renamed")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}
	if result.NewPath != "sub/child-renamed.md" {
		t.Errorf("NewPath = %q, want sub/child-renamed.md", result.NewPath)
	}
}

func TestRename_FileNotFound(t *testing.T) {
	v, idx := setupTempVault(t)
	_, err := Rename(v, idx, "no-such-file.md", "new-name")
	if err == nil {
		t.Error("expected error for non-existent source, got nil")
	}
}

func TestMove_EscapedPipeLinksUpdated(t *testing.T) {
	v, idx := setupTempVault(t)

	// table-links.md references note-a as [[note-a\|first note]] (table cell syntax).
	// Moving note-a.md should rewrite it to [[moved-a\|first note]].
	result, err := Move(v, idx, "note-a.md", "moved-a.md")
	if err != nil {
		t.Fatalf("move: %v", err)
	}

	updated := make(map[string]bool)
	for _, u := range result.Updated {
		updated[u] = true
	}
	if !updated["table-links.md"] {
		t.Error("table-links.md should have been updated (it links to note-a via \\|)")
	}

	content, err := os.ReadFile(filepath.Join(v.Root, "table-links.md"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	if strings.Contains(body, `[[note-a\|`) {
		t.Error("table-links.md still contains old [[note-a\\| reference after move")
	}
	if !strings.Contains(body, `[[moved-a\|`) {
		t.Error("table-links.md missing new [[moved-a\\| reference after move")
	}
}

func TestRename_EscapedPipeLinksUpdated(t *testing.T) {
	v, idx := setupTempVault(t)

	// table-links.md references note-a as [[note-a\|first note]].
	// Renaming note-a.md should rewrite it to [[note-a-renamed\|first note]].
	result, err := Rename(v, idx, "note-a.md", "note-a-renamed")
	if err != nil {
		t.Fatalf("rename: %v", err)
	}

	updated := make(map[string]bool)
	for _, u := range result.Updated {
		updated[u] = true
	}
	if !updated["table-links.md"] {
		t.Error("table-links.md should have been updated (it links to note-a via \\|)")
	}

	content, err := os.ReadFile(filepath.Join(v.Root, "table-links.md"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	if strings.Contains(body, `[[note-a\|`) {
		t.Error("table-links.md still contains old [[note-a\\| reference after rename")
	}
	if !strings.Contains(body, `[[note-a-renamed\|`) {
		t.Error("table-links.md missing new [[note-a-renamed\\| reference after rename")
	}
}

func TestMove_PreservesFilePermissions(t *testing.T) {
	v, idx := setupTempVault(t)

	// Make index.md only owner-readable.
	abs := filepath.Join(v.Root, "index.md")
	if err := os.Chmod(abs, 0o600); err != nil {
		t.Fatal(err)
	}

	// Move note-a.md (linked from index.md) — this triggers a link rewrite on index.md.
	_, err := Move(v, idx, "note-a.md", "notes/note-a.md")
	if err != nil {
		t.Fatalf("move: %v", err)
	}

	fi, err := os.Stat(abs)
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode().Perm() != 0o600 {
		t.Errorf("index.md permissions = %o after rewrite, want 0600", fi.Mode().Perm())
	}
}
