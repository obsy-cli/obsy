package ops

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/obsy-cli/obsy/internal/vault"
)

// SearchResult is one matching file.
type SearchResult struct {
	Path    string
	Context []string // matching lines (only populated with --context)
}

// Search scans all .md files in the vault for query.
// It reads from disk (not the index). Case-insensitive unless caseSensitive is true.
func Search(v *vault.Vault, query, pathFilter string, limit int, context, caseSensitive bool) ([]SearchResult, error) {
	files, err := v.Files()
	if err != nil {
		return nil, err
	}

	matchQuery := query
	if !caseSensitive {
		matchQuery = strings.ToLower(query)
	}

	var results []SearchResult
	for _, rel := range files {
		if pathFilter != "" && !strings.HasPrefix(rel, pathFilter) {
			continue
		}
		abs := filepath.Join(v.Root, rel)
		lines, matched, err := searchFile(abs, matchQuery, caseSensitive, context)
		if err != nil || !matched {
			continue
		}
		results = append(results, SearchResult{Path: rel, Context: lines})
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results, nil
}

func searchFile(path, query string, caseSensitive, context bool) ([]string, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	var matchLines []string
	matched := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		cmp := line
		if !caseSensitive {
			cmp = strings.ToLower(line)
		}
		if strings.Contains(cmp, query) {
			matched = true
			if context {
				matchLines = append(matchLines, line)
			}
		}
	}
	return matchLines, matched, scanner.Err()
}
