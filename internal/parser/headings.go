package parser

import (
	"bytes"
	"strings"
)

// Heading represents a markdown heading.
type Heading struct {
	Level int
	Text  string
	Line  int // 1-based
}

// ParseHeadings extracts all ATX headings (# through ######) from content.
func ParseHeadings(content []byte) []Heading {
	var headings []Heading
	lines := bytes.Split(content, []byte("\n"))
	inFence := false

	for i, line := range lines {
		lineNum := i + 1
		trimmed := bytes.TrimSpace(line)

		if bytes.HasPrefix(trimmed, []byte("```")) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}

		level, text, ok := parseHeadingLine(string(trimmed))
		if ok {
			headings = append(headings, Heading{Level: level, Text: text, Line: lineNum})
		}
	}
	return headings
}

func parseHeadingLine(s string) (level int, text string, ok bool) {
	if !strings.HasPrefix(s, "#") {
		return
	}
	i := 0
	for i < len(s) && s[i] == '#' {
		i++
	}
	if i > 6 || i >= len(s) || s[i] != ' ' {
		return
	}
	return i, strings.TrimSpace(s[i+1:]), true
}
