package parser

import (
	"bytes"
	"strings"
)

// Link represents a wikilink found in a file.
type Link struct {
	Raw     string // content between [[ and ]], e.g. "note#heading|display"
	IsEmbed bool   // true for ![[...]]
	Line    int    // 1-based line number
}

// ParseLinks extracts all wikilinks from content, skipping code blocks.
// content should be the body after frontmatter is removed.
func ParseLinks(content []byte) []Link {
	var links []Link
	lines := bytes.Split(content, []byte("\n"))
	inFence := false
	lineNum := 0

	for _, line := range lines {
		lineNum++
		trimmed := bytes.TrimSpace(line)

		// Toggle fenced code block.
		if bytes.HasPrefix(trimmed, []byte("```")) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}

		links = append(links, parseLinksInLine(line, lineNum)...)
	}
	return links
}

func parseLinksInLine(line []byte, lineNum int) []Link {
	var links []Link
	s := string(line)
	i := 0

	for i < len(s) {
		// Skip inline code spans.
		if s[i] == '`' {
			end := strings.IndexByte(s[i+1:], '`')
			if end >= 0 {
				i += end + 2
				continue
			}
		}

		isEmbed := false
		start := -1

		if i+1 < len(s) && s[i] == '!' && s[i+1] == '[' && i+2 < len(s) && s[i+2] == '[' {
			isEmbed = true
			start = i + 3
			i += 3
		} else if i+1 < len(s) && s[i] == '[' && s[i+1] == '[' {
			start = i + 2
			i += 2
		} else {
			i++
			continue
		}

		end := strings.Index(s[start:], "]]")
		if end < 0 {
			continue
		}

		raw := s[start : start+end]
		if raw != "" {
			links = append(links, Link{Raw: raw, IsEmbed: isEmbed, Line: lineNum})
		}
		i = start + end + 2
	}
	return links
}
