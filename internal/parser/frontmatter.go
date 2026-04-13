package parser

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Frontmatter holds the parsed YAML front matter of a markdown file.
type Frontmatter struct {
	Tags    []string
	Aliases []string
	Props   map[string]any // all fields, including tags and aliases
}

// ParseFrontmatter extracts YAML front matter from content.
// Returns the parsed front matter, the body (content after the closing ---), and any error.
// If no front matter is present, body == content, Frontmatter is zero, and err is nil.
// err is non-nil only when a valid front matter block is found but the YAML inside is malformed.
func ParseFrontmatter(content []byte) (Frontmatter, []byte, error) {
	const delim = "---"

	// Must start with "---\n" or "---\r\n".
	if !bytes.HasPrefix(content, []byte("---")) {
		return Frontmatter{}, content, nil
	}
	rest := content[3:]
	if len(rest) == 0 || (rest[0] != '\n' && rest[0] != '\r') {
		return Frontmatter{}, content, nil
	}
	if rest[0] == '\r' {
		rest = rest[1:]
	}
	if len(rest) == 0 || rest[0] != '\n' {
		return Frontmatter{}, content, nil
	}
	rest = rest[1:] // consume \n

	// Find closing ---.
	end := findDelimiter(rest, delim)
	if end < 0 {
		return Frontmatter{}, content, nil
	}

	yamlBytes := rest[:end]
	body := rest[end+len(delim):]
	// Consume the newline after closing ---.
	if len(body) > 0 && body[0] == '\r' {
		body = body[1:]
	}
	if len(body) > 0 && body[0] == '\n' {
		body = body[1:]
	}

	var raw map[string]any
	if err := yaml.Unmarshal(yamlBytes, &raw); err != nil {
		return Frontmatter{}, content, fmt.Errorf("malformed frontmatter YAML: %w", err)
	}

	fm := Frontmatter{Props: raw}
	fm.Tags = stringSlice(raw, "tags")
	fm.Aliases = stringSlice(raw, "aliases")
	return fm, body, nil
}

// findDelimiter finds the position of "---" at the start of a line within b.
func findDelimiter(b []byte, delim string) int {
	d := []byte(delim)
	i := 0
	for i < len(b) {
		if bytes.HasPrefix(b[i:], d) {
			// Check it's at line start (i==0 or previous char is \n).
			if i == 0 || b[i-1] == '\n' {
				// Optionally followed by \r\n or \n or EOF.
				after := i + len(d)
				if after >= len(b) || b[after] == '\n' || b[after] == '\r' {
					return i
				}
			}
		}
		nl := bytes.IndexByte(b[i:], '\n')
		if nl < 0 {
			break
		}
		i += nl + 1
	}
	return -1
}

// stringSlice extracts a string or []string YAML value as a Go []string.
func stringSlice(m map[string]any, key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	switch val := v.(type) {
	case string:
		if val == "" {
			return nil
		}
		return []string{val}
	case []any:
		var out []string
		for _, item := range val {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}
