package vault

import (
	"testing"
)

func TestStripLinkTarget(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"note", "note"},
		{"note#heading", "note"},
		{"note|display text", "note"},
		{"note#heading|display", "note"},
		{"folder/note", "folder/note"},
		{"folder/note#h|d", "folder/note"},
		{"  spaced  ", "spaced"},
		{"", ""},
	}
	for _, tt := range tests {
		got := StripLinkTarget(tt.input)
		if got != tt.want {
			t.Errorf("StripLinkTarget(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResolveLink(t *testing.T) {
	files := []string{
		"index.md",
		"note-a.md",
		"note-b.md",
		"sub/note-a.md", // same name as note-a.md, deeper
		"sub/child.md",
		"sub/deep/other.md",
	}
	noAliases := map[string]string{}

	tests := []struct {
		name     string
		target   string
		fromFile string
		aliases  map[string]string
		want     string
	}{
		{
			name:   "exact basename match",
			target: "note-b",
			want:   "note-b.md",
		},
		{
			name:   "match with .md extension",
			target: "note-b.md",
			want:   "note-b.md",
		},
		{
			name:   "path-qualified — exact",
			target: "sub/child",
			want:   "sub/child.md",
		},
		{
			name:   "path-qualified — no match",
			target: "sub/missing",
			want:   "",
		},
		{
			name:   "no match",
			target: "does-not-exist",
			want:   "",
		},
		{
			name:   "empty target",
			target: "",
			want:   "",
		},
		{
			// note-a.md (depth 1) beats sub/note-a.md (depth 2)
			name:   "shallowest depth wins",
			target: "note-a",
			want:   "note-a.md",
		},
		{
			// from sub/child.md: sub/note-a.md is same folder, so it wins over note-a.md on depth tie? No — note-a.md is shallower (depth 1 vs 2), so depth rule fires first.
			// Let's use a case where depths are equal: sub/child.md vs sub/note-a.md from sub/deep/other.md
			// both are depth 2. from sub/deep/other.md, sub/ is the parent. sub/child.md parent = sub, sub/note-a.md parent = sub, from file parent = sub/deep. Neither is same folder.
			// alphabetical: sub/child.md < sub/note-a.md
			name:     "equal depth, neither same folder — alphabetical",
			target:   "child",
			fromFile: "sub/deep/other.md",
			want:     "sub/child.md",
		},
		{
			// from sub/child.md: sub/note-a.md is in same folder (sub/). note-a.md is depth 1.
			// depth 1 < depth 2, so note-a.md still wins.
			name:     "depth beats same-folder preference",
			target:   "note-a",
			fromFile: "sub/child.md",
			want:     "note-a.md",
		},
		{
			name:    "alias resolution",
			target:  "myalias",
			aliases: map[string]string{"myalias": "note-a.md"},
			want:    "note-a.md",
		},
		{
			name:    "alias takes priority over filename match",
			target:  "note-b",
			aliases: map[string]string{"note-b": "index.md"},
			want:    "index.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aliases := tt.aliases
			if aliases == nil {
				aliases = noAliases
			}
			fromFile := tt.fromFile
			got := ResolveLink(tt.target, fromFile, files, aliases)
			if got != tt.want {
				t.Errorf("ResolveLink(%q, %q) = %q, want %q", tt.target, fromFile, got, tt.want)
			}
		})
	}
}

func TestResolveLinkSameFolderTiebreaker(t *testing.T) {
	// Two files at the same depth; one is in the same folder as the linking file.
	files := []string{
		"a/note.md",
		"b/note.md",
	}
	aliases := map[string]string{}

	// From a/other.md: a/note.md is same folder → wins.
	got := ResolveLink("note", "a/other.md", files, aliases)
	if got != "a/note.md" {
		t.Errorf("same-folder tiebreaker: got %q, want %q", got, "a/note.md")
	}

	// From b/other.md: b/note.md is same folder → wins.
	got = ResolveLink("note", "b/other.md", files, aliases)
	if got != "b/note.md" {
		t.Errorf("same-folder tiebreaker: got %q, want %q", got, "b/note.md")
	}

	// From root (no folder match): alphabetical → a/note.md.
	got = ResolveLink("note", "index.md", files, aliases)
	if got != "a/note.md" {
		t.Errorf("alphabetical fallback: got %q, want %q", got, "a/note.md")
	}
}
