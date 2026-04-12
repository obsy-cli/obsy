package index

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/obsy-cli/obsy/internal/vault"
)

// BrokenLink is an unresolved wikilink or missing embed.
type BrokenLink struct {
	SourceFile string
	RawTarget  string
	IsEmbed    bool
}

// OutgoingLink is a link from a file with its resolved target (empty = unresolved).
type OutgoingLink struct {
	Raw      string
	Resolved string // relative path, or "" if unresolved
	IsEmbed  bool
}

// AliasMap builds alias→path from the current index.
func (idx *Index) AliasMap() map[string]string {
	m := make(map[string]string)
	for path, entry := range idx.Files {
		for _, a := range entry.Aliases {
			if a != "" {
				m[a] = path
			}
		}
	}
	return m
}

// AllFiles returns a sorted slice of all relative file paths.
func (idx *Index) AllFiles() []string {
	files := make([]string, 0, len(idx.Files))
	for f := range idx.Files {
		files = append(files, f)
	}
	sort.Strings(files)
	return files
}

// ResolveRaw resolves a raw link target from fromFile using the provided alias map and file list.
func (idx *Index) ResolveRaw(raw, fromFile string, aliases map[string]string, files []string) string {
	return idx.resolve(raw, fromFile, aliases, files)
}

// resolve resolves a raw link target from a given source file.
// Returns the resolved relative path, or "" if unresolved.
func (idx *Index) resolve(raw, fromFile string, aliases map[string]string, files []string) string {
	target := vault.StripLinkTarget(raw)
	if target == "" {
		return ""
	}
	// Non-.md embed: check filesystem existence.
	if hasNonMDExtension(target) {
		v := &vault.Vault{Root: idx.VaultRoot}
		if v.ResolveNonMD(target) {
			return target
		}
		return ""
	}
	return vault.ResolveLink(target, fromFile, files, aliases)
}

// UnresolvedLinks returns all broken links across the vault.
func (idx *Index) UnresolvedLinks(pathFilter string) []BrokenLink {
	aliases := idx.AliasMap()
	files := idx.AllFiles()
	var broken []BrokenLink

	for _, entry := range idx.Files {
		if pathFilter != "" && !strings.HasPrefix(entry.Path, pathFilter) {
			continue
		}
		for _, link := range entry.Links {
			if idx.resolve(link.Raw, entry.Path, aliases, files) == "" {
				broken = append(broken, BrokenLink{
					SourceFile: entry.Path,
					RawTarget:  link.Raw,
					IsEmbed:    link.IsEmbed,
				})
			}
		}
	}
	sort.Slice(broken, func(i, j int) bool {
		if broken[i].RawTarget != broken[j].RawTarget {
			return broken[i].RawTarget < broken[j].RawTarget
		}
		return broken[i].SourceFile < broken[j].SourceFile
	})
	return broken
}

// Orphans returns files with no incoming links, optionally ignoring a glob pattern.
func (idx *Index) Orphans(ignoreGlob string) []string {
	aliases := idx.AliasMap()
	files := idx.AllFiles()

	incoming := make(map[string]bool)
	for _, entry := range idx.Files {
		for _, link := range entry.Links {
			if r := idx.resolve(link.Raw, entry.Path, aliases, files); r != "" {
				incoming[r] = true
			}
		}
	}

	var orphans []string
	for _, f := range files {
		if incoming[f] {
			continue
		}
		if ignoreGlob != "" {
			if matched, _ := filepath.Match(ignoreGlob, f); matched {
				continue
			}
		}
		orphans = append(orphans, f)
	}
	sort.Strings(orphans)
	return orphans
}

// Deadends returns files that have no outgoing links that resolve successfully.
func (idx *Index) Deadends() []string {
	aliases := idx.AliasMap()
	files := idx.AllFiles()

	var deadends []string
	for _, f := range files {
		entry := idx.Files[f]
		hasResolved := false
		for _, link := range entry.Links {
			if idx.resolve(link.Raw, entry.Path, aliases, files) != "" {
				hasResolved = true
				break
			}
		}
		if !hasResolved {
			deadends = append(deadends, f)
		}
	}
	sort.Strings(deadends)
	return deadends
}

// BacklinkCounts returns how many links each source file has pointing to target.
func (idx *Index) BacklinkCounts(target string) map[string]int {
	aliases := idx.AliasMap()
	files := idx.AllFiles()
	counts := make(map[string]int)
	for _, f := range files {
		entry := idx.Files[f]
		for _, link := range entry.Links {
			if idx.resolve(link.Raw, entry.Path, aliases, files) == target {
				counts[f]++
			}
		}
	}
	return counts
}

// BacklinksTo returns all files that link to the given file (relative path).
func (idx *Index) BacklinksTo(target string) []string {
	aliases := idx.AliasMap()
	files := idx.AllFiles()

	var sources []string
	for _, f := range files {
		if f == target {
			continue
		}
		entry := idx.Files[f]
		for _, link := range entry.Links {
			if idx.resolve(link.Raw, entry.Path, aliases, files) == target {
				sources = append(sources, f)
				break
			}
		}
	}
	sort.Strings(sources)
	return sources
}

// LinksFrom returns all outgoing links from a file with their resolved targets.
func (idx *Index) LinksFrom(file string) []OutgoingLink {
	entry, ok := idx.Files[file]
	if !ok {
		return nil
	}
	aliases := idx.AliasMap()
	files := idx.AllFiles()

	var out []OutgoingLink
	for _, link := range entry.Links {
		out = append(out, OutgoingLink{
			Raw:      link.Raw,
			Resolved: idx.resolve(link.Raw, entry.Path, aliases, files),
			IsEmbed:  link.IsEmbed,
		})
	}
	return out
}

// ResolveFileArg resolves a file argument (name or path) to a relative path.
// Mirrors wikilink resolution: accepts basename, path, or exact relative path.
func (idx *Index) ResolveFileArg(arg string) (string, bool) {
	// Exact match first.
	if _, ok := idx.Files[arg]; ok {
		return arg, true
	}
	// Try with .md appended.
	withExt := arg
	if !strings.HasSuffix(withExt, ".md") {
		withExt = arg + ".md"
	}
	if _, ok := idx.Files[withExt]; ok {
		return withExt, true
	}
	// Resolve like a wikilink (by basename).
	aliases := idx.AliasMap()
	files := idx.AllFiles()
	r := vault.ResolveLink(arg, "", files, aliases)
	if r != "" {
		return r, true
	}
	return "", false
}

// hasNonMDExtension returns true if target has a file extension that is not .md.
func hasNonMDExtension(target string) bool {
	ext := strings.ToLower(filepath.Ext(target))
	return ext != "" && ext != ".md"
}
