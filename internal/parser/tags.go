package parser

import (
	"bytes"
	"regexp"
	"strings"
	"unicode"
)

// tagPattern matches #tag in body text.
// Tag chars: letters, digits, _, /, -
// Must not be purely numeric.
var tagPattern = regexp.MustCompile(`(?:^|[\s(,;])#([a-zA-Z0-9_/\-]+)`)

// ParseInlineTags extracts inline #tags from the body (post-frontmatter).
// Skips tags at line start (headings), inside code blocks, inside code spans.
func ParseInlineTags(body []byte) []string {
	var tags []string
	seen := make(map[string]bool)

	lines := bytes.Split(body, []byte("\n"))
	inFence := false

	for _, line := range lines {
		trimmed := bytes.TrimSpace(line)

		if bytes.HasPrefix(trimmed, []byte("```")) {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}

		// Skip headings (lines starting with #).
		if bytes.HasPrefix(trimmed, []byte("#")) {
			continue
		}

		lineStr := stripInlineCode(string(line))
		matches := tagPattern.FindAllStringSubmatch(lineStr, -1)
		for _, m := range matches {
			tag := m[1]
			if isValidTag(tag) && !seen[tag] {
				seen[tag] = true
				tags = append(tags, tag)
			}
		}
	}
	return tags
}

// isValidTag returns true if the tag has at least one non-numeric character.
func isValidTag(tag string) bool {
	for _, r := range tag {
		if !unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// stripInlineCode removes content inside backtick spans.
func stripInlineCode(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '`' {
			end := strings.IndexByte(s[i+1:], '`')
			if end >= 0 {
				i += end + 2
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}
