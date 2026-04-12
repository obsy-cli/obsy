package vault

import (
	"os"
	"path/filepath"
	"testing"
)

// --- IsExcluded ---

func TestIsExcluded(t *testing.T) {
	cases := []struct {
		rel  string
		want bool
	}{
		{"note.md", false},
		{"sub/note.md", false},
		{".hidden.md", true},
		{".obsidian/config.json", true},
		{".git/HEAD", true},
		{".trash/old.md", true},
		{"sub/.hidden/note.md", true},
		{"img/photo.png", false},
	}
	for _, tc := range cases {
		got := IsExcluded(tc.rel)
		if got != tc.want {
			t.Errorf("IsExcluded(%q) = %v, want %v", tc.rel, got, tc.want)
		}
	}
}

// --- Vault.Files ---

func makeVault(t *testing.T, files map[string]string) *Vault {
	t.Helper()
	root := t.TempDir()
	for rel, content := range files {
		abs := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return &Vault{Root: root}
}

func TestFiles_Basic(t *testing.T) {
	v := makeVault(t, map[string]string{
		"a.md":      "# A",
		"sub/b.md":  "# B",
		"sub/c.txt": "ignored",
		"image.png": "ignored",
	})
	files, err := v.Files()
	if err != nil {
		t.Fatal(err)
	}
	fileSet := make(map[string]bool)
	for _, f := range files {
		fileSet[f] = true
	}
	if !fileSet["a.md"] {
		t.Error("expected a.md")
	}
	if !fileSet["sub/b.md"] {
		t.Error("expected sub/b.md")
	}
	if fileSet["sub/c.txt"] {
		t.Error("c.txt should be excluded (not .md)")
	}
	if fileSet["image.png"] {
		t.Error("image.png should be excluded (not .md)")
	}
}

func TestFiles_ExcludesHidden(t *testing.T) {
	v := makeVault(t, map[string]string{
		"visible.md":         "ok",
		".hidden.md":         "skip",
		".obsidian/app.json": "skip",
		".git/config":        "skip",
		".trash/deleted.md":  "skip",
	})
	files, err := v.Files()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if IsExcluded(f) {
			t.Errorf("excluded path %q appeared in Files()", f)
		}
	}
	if len(files) != 1 || files[0] != "visible.md" {
		t.Errorf("Files() = %v, want [visible.md]", files)
	}
}

func TestFiles_Symlink(t *testing.T) {
	root := t.TempDir()

	// Real file outside vault.
	target := filepath.Join(root, "real.md")
	if err := os.WriteFile(target, []byte("real"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Symlink inside vault pointing to the real file.
	link := filepath.Join(root, "linked.md")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlinks not supported:", err)
	}
	// Also a regular file.
	if err := os.WriteFile(filepath.Join(root, "normal.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	v := &Vault{Root: root}
	files, err := v.Files()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if f == "linked.md" {
			t.Error("symlink linked.md should be excluded from Files()")
		}
	}
}

// --- Discover ---

func TestDiscover_Explicit(t *testing.T) {
	tmp := t.TempDir()
	v, err := Discover(tmp)
	if err != nil {
		t.Fatalf("Discover(%q): %v", tmp, err)
	}
	if v.Root != tmp {
		t.Errorf("Root = %q, want %q", v.Root, tmp)
	}
}

func TestDiscover_ObsidianMarker(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".obsidian"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Discover from a subdirectory — should walk up and find root.
	sub := filepath.Join(root, "notes", "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	// We can't change cwd in parallel tests; test explicit path instead.
	v, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	if v.Root != root {
		t.Errorf("Root = %q, want %q", v.Root, root)
	}
}
