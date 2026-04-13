package index

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStderr redirects os.Stderr for the duration of f, returning what was written.
func captureStderr(t *testing.T, f func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	orig := os.Stderr
	os.Stderr = w

	f()

	w.Close()
	os.Stderr = orig
	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()
	return buf.String()
}

func TestAliasMap_NoCollision(t *testing.T) {
	idx := New("/vault")
	idx.Files["a.md"] = &FileEntry{Path: "a.md", Aliases: []string{"alpha"}}
	idx.Files["b.md"] = &FileEntry{Path: "b.md", Aliases: []string{"beta"}}

	stderr := captureStderr(t, func() {
		m := idx.AliasMap()
		if m["alpha"] != "a.md" {
			t.Errorf("alpha → %q, want a.md", m["alpha"])
		}
		if m["beta"] != "b.md" {
			t.Errorf("beta → %q, want b.md", m["beta"])
		}
	})
	if stderr != "" {
		t.Errorf("unexpected stderr output: %q", stderr)
	}
}

func TestAliasMap_CollisionWarns(t *testing.T) {
	idx := New("/vault")
	idx.Files["a.md"] = &FileEntry{Path: "a.md", Aliases: []string{"shared"}}
	idx.Files["b.md"] = &FileEntry{Path: "b.md", Aliases: []string{"shared"}}

	var aliasMap map[string]string
	stderr := captureStderr(t, func() {
		aliasMap = idx.AliasMap()
	})

	if !strings.Contains(stderr, "warning") || !strings.Contains(stderr, "shared") {
		t.Errorf("expected collision warning on stderr, got: %q", stderr)
	}

	// Resolves deterministically to lexicographically first path.
	if aliasMap["shared"] != "a.md" {
		t.Errorf("collision: alias resolved to %q, want a.md (lex first)", aliasMap["shared"])
	}
}

func TestAliasMap_EmptyAliasIgnored(t *testing.T) {
	idx := New("/vault")
	idx.Files["a.md"] = &FileEntry{Path: "a.md", Aliases: []string{"", "valid"}}

	stderr := captureStderr(t, func() {
		m := idx.AliasMap()
		if _, ok := m[""]; ok {
			t.Error("empty string should not be added to alias map")
		}
		if m["valid"] != "a.md" {
			t.Errorf("valid → %q, want a.md", m["valid"])
		}
	})
	if stderr != "" {
		t.Errorf("unexpected stderr: %q", stderr)
	}
}
