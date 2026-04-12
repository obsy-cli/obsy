package vault

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// StripLinkTarget removes the anchor (#...) and display text (|...) from a
// raw wikilink target, returning just the path component.
//
//	"note#heading"       → "note"
//	"note|display"       → "note"
//	"folder/note#h|d"    → "folder/note"
func StripLinkTarget(raw string) string {
	t := raw
	if i := strings.IndexByte(t, '#'); i >= 0 {
		t = t[:i]
	}
	if i := strings.IndexByte(t, '|'); i >= 0 {
		t = t[:i]
	}
	return strings.TrimSpace(t)
}

// ResolveLink resolves a wikilink target to a path relative to the vault root.
// Returns "" if no match is found.
//
// target: already stripped of anchors and display text (call StripLinkTarget first).
// fromFile: the file containing the link (relative to vault root), for tiebreaking.
// files: all .md files in the vault (relative to vault root).
// aliases: map from alias → file path (relative to vault root).
func ResolveLink(target, fromFile string, files []string, aliases map[string]string) string {
	if target == "" {
		return ""
	}

	// Normalize: add .md extension if missing.
	withExt := target
	if !strings.HasSuffix(withExt, ".md") {
		withExt = target + ".md"
	}

	// Exact vault-root-relative path match (e.g. [[folder/note]]).
	if strings.Contains(target, "/") {
		for _, f := range files {
			if f == withExt {
				return f
			}
		}
		return ""
	}

	// Check aliases first.
	if p, ok := aliases[target]; ok {
		return p
	}

	// Collect all files whose basename matches.
	base := filepath.Base(withExt)
	var candidates []string
	for _, f := range files {
		if filepath.Base(f) == base {
			candidates = append(candidates, f)
		}
	}

	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

	// Tiebreaking: fewest path components → same folder as fromFile → alphabetical.
	fromDir := filepath.Dir(fromFile)

	sort.Slice(candidates, func(i, j int) bool {
		di := strings.Count(candidates[i], string(filepath.Separator))
		dj := strings.Count(candidates[j], string(filepath.Separator))
		if di != dj {
			return di < dj
		}
		// Same depth: prefer the one in the same directory as fromFile.
		sameI := filepath.Dir(candidates[i]) == fromDir
		sameJ := filepath.Dir(candidates[j]) == fromDir
		if sameI != sameJ {
			return sameI
		}
		return candidates[i] < candidates[j]
	})
	return candidates[0]
}

// ResolveNonMD checks whether a non-.md embed target exists on disk.
// target must have a non-.md extension. Returns true if the file exists.
func (v *Vault) ResolveNonMD(target string) bool {
	abs := filepath.Join(v.Root, target)
	_, err := os.Stat(abs)
	return err == nil
}

// BuildAliasMap builds a map from alias → relative file path from the index data.
// aliasesPerFile: map[relPath][]alias
func BuildAliasMap(aliasesPerFile map[string][]string) map[string]string {
	m := make(map[string]string)
	for path, aliases := range aliasesPerFile {
		for _, a := range aliases {
			if a != "" {
				m[a] = path
			}
		}
	}
	return m
}
