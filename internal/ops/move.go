package ops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/obsy-cli/obsy/internal/index"
	"github.com/obsy-cli/obsy/internal/vault"
)

// MoveResult describes the outcome of a move operation.
type MoveResult struct {
	OldPath    string
	NewPath    string
	Updated    []string // files whose wikilinks were rewritten
	NotUpdated []string // files that needed updating but couldn't be written
}

// Move moves src to dst within the vault and updates all wikilinks.
// dst may be a folder (file moves into it) or a full relative path.
// Validates all writes before touching any file.
func Move(v *vault.Vault, idx *index.Index, src, dst string) (*MoveResult, error) {
	// Resolve src.
	srcRel, ok := idx.ResolveFileArg(src)
	if !ok {
		return nil, fmt.Errorf("file not found: %s", src)
	}

	// Resolve dst: if it's a directory or ends in /, move file into it.
	dstRel := resolveDest(v, srcRel, dst)

	if srcRel == dstRel {
		return nil, fmt.Errorf("source and destination are the same: %s", srcRel)
	}
	if _, exists := idx.Files[dstRel]; exists {
		return nil, fmt.Errorf("destination already exists: %s", dstRel)
	}

	// Find all files that reference src.
	refs := idx.BacklinksTo(srcRel)

	// Validate: can we write to all referencing files?
	if err := validateWritable(v, refs); err != nil {
		return nil, err
	}
	// Validate: can we create the destination.
	dstAbs := filepath.Join(v.Root, dstRel)
	if err := os.MkdirAll(filepath.Dir(dstAbs), 0o755); err != nil {
		return nil, fmt.Errorf("cannot create destination directory: %w", err)
	}

	result := &MoveResult{OldPath: srcRel, NewPath: dstRel}

	// Apply: move the file.
	if err := os.Rename(filepath.Join(v.Root, srcRel), dstAbs); err != nil {
		return nil, fmt.Errorf("move failed: %w", err)
	}

	// Apply: rewrite wikilinks in referencing files.
	oldBase := strings.TrimSuffix(filepath.Base(srcRel), ".md")
	newBase := strings.TrimSuffix(filepath.Base(dstRel), ".md")

	for _, ref := range refs {
		if err := rewriteLinks(v, ref, oldBase, newBase, srcRel, dstRel); err != nil {
			result.NotUpdated = append(result.NotUpdated, ref)
		} else {
			result.Updated = append(result.Updated, ref)
		}
	}

	// Update index.
	entry := idx.Files[srcRel]
	entry.Path = dstRel
	idx.Files[dstRel] = entry
	delete(idx.Files, srcRel)

	return result, nil
}

// resolveDest determines the destination relative path.
func resolveDest(v *vault.Vault, srcRel, dst string) string {
	// Absolute or relative path provided.
	// If dst ends with / or is an existing directory, move into it.
	if strings.HasSuffix(dst, "/") || strings.HasSuffix(dst, string(filepath.Separator)) {
		return filepath.Join(dst, filepath.Base(srcRel))
	}
	absDir := filepath.Join(v.Root, dst)
	if fi, err := os.Stat(absDir); err == nil && fi.IsDir() {
		return filepath.Join(dst, filepath.Base(srcRel))
	}
	// dst is a full path.
	if !strings.HasSuffix(dst, ".md") {
		dst += ".md"
	}
	return dst
}

func validateWritable(v *vault.Vault, paths []string) error {
	for _, rel := range paths {
		abs := filepath.Join(v.Root, rel)
		f, err := os.OpenFile(abs, os.O_WRONLY, 0)
		if err != nil {
			return fmt.Errorf("cannot write to %s: %w", rel, err)
		}
		f.Close()
	}
	return nil
}

// rewriteLinks rewrites wikilinks in a file from old name to new name.
func rewriteLinks(v *vault.Vault, rel, oldBase, newBase, oldRel, newRel string) error {
	abs := filepath.Join(v.Root, rel)
	fi, err := os.Stat(abs)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(abs)
	if err != nil {
		return err
	}

	updated := rewriteWikilinks(string(content), oldBase, newBase, oldRel, newRel)
	if updated == string(content) {
		return nil // nothing changed
	}
	return os.WriteFile(abs, []byte(updated), fi.Mode())
}

// rewriteWikilinks replaces [[old]] with [[new]] in content.
// Handles: [[old]], [[old|alias]], [[old#heading]], [[old#heading|alias]], [[folder/old]].
func rewriteWikilinks(content, oldBase, newBase, oldRel, newRel string) string {
	var b strings.Builder
	i := 0
	for i < len(content) {
		// Find next [[
		j := strings.Index(content[i:], "[[")
		if j < 0 {
			b.WriteString(content[i:])
			break
		}
		b.WriteString(content[i : i+j])
		i += j

		// Find closing ]]
		end := strings.Index(content[i+2:], "]]")
		if end < 0 {
			b.WriteString(content[i:])
			break
		}
		inner := content[i+2 : i+2+end]
		// Determine the path part (before # and |).
		pathPart := inner
		if k := strings.IndexByte(pathPart, '#'); k >= 0 {
			pathPart = pathPart[:k]
		}
		if k := strings.IndexByte(pathPart, '|'); k >= 0 {
			// \| is how Obsidian escapes pipes inside markdown table cells.
			// Strip the backslash so the path is clean; suffix retains \| for round-trip fidelity.
			if k > 0 && pathPart[k-1] == '\\' {
				pathPart = pathPart[:k-1]
			} else {
				pathPart = pathPart[:k]
			}
		}
		// Capture unstripped length before TrimSpace so suffix offset is correct.
		// e.g. inner=" note-a |alias": unstripped len=8, trimmed len=6; inner[6:]="a |alias" (wrong).
		pathPartRawLen := len(pathPart)
		pathPart = strings.TrimSpace(pathPart)

		suffix := inner[pathPartRawLen:]

		// Match by basename or full relative path.
		base := strings.TrimSuffix(filepath.Base(pathPart), ".md")
		relNoExt := strings.TrimSuffix(oldRel, ".md")

		if base == oldBase || pathPart == relNoExt || pathPart == oldRel {
			// Rewrite to newBase (preserve path qualification if it was path-qualified).
			var newPath string
			if pathPart == relNoExt || pathPart == oldRel {
				newPath = strings.TrimSuffix(newRel, ".md")
			} else {
				newPath = newBase
			}
			b.WriteString("[[")
			b.WriteString(newPath)
			b.WriteString(suffix)
			b.WriteString("]]")
		} else {
			b.WriteString("[[")
			b.WriteString(inner)
			b.WriteString("]]")
		}
		i += 2 + end + 2
	}
	return b.String()
}
