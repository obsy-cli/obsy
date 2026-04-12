package parser

// Fuzz tests verify that no input causes a panic or infinite loop.
// Run with: go test -fuzz=FuzzParseLinks -fuzztime=30s ./internal/parser/

import "testing"

func FuzzParseLinks(f *testing.F) {
	// Seed corpus from known interesting inputs.
	seeds := []string{
		"[[note]]",
		"![[embed]]",
		"[[note|alias]]",
		"[[note#heading|alias]]",
		"```\n[[inside-fence]]\n```\n[[outside]]",
		"`[[inline-code]]` [[kept]]",
		"[[unclosed",
		"]]orphan",
		"[[ ]]",
		"[[]]",
		string([]byte{0x00, 0x01, 0xff}),
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		// Must not panic.
		_ = ParseLinks(data)
	})
}

func FuzzParseInlineTags(f *testing.F) {
	seeds := []string{
		"#tag",
		"text #tag here",
		"#123",
		"#abc/def",
		"#my-tag",
		"```\n#inside\n```\n#outside",
		"`#inline` #kept",
		string([]byte{0x00, '#', 0xff}),
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = ParseInlineTags(data)
	})
}

func FuzzParseFrontmatter(f *testing.F) {
	seeds := []string{
		"---\ntags: [a, b]\n---\nbody",
		"---\ntitle: foo\n---",
		"---\n---",
		"no frontmatter",
		"---\nbad yaml: [unclosed",
		"---\nnested:\n  key: val\n---",
		string([]byte{0x00, '-', '-', '-', '\n'}),
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ParseFrontmatter(data)
	})
}

func FuzzParseTasks(f *testing.F) {
	seeds := []string{
		"- [ ] task",
		"- [x] done",
		"* [ ] asterisk",
		"  - [ ] nested",
		"```\n- [ ] fenced\n```\n- [ ] outside",
		"- [ ]no space",
		string([]byte{'-', ' ', '[', 0x00, ']', ' ', 't'}),
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = ParseTasks(data)
	})
}

func FuzzParseHeadings(f *testing.F) {
	seeds := []string{
		"# Heading",
		"## Two",
		"####### Too many",
		"#nospace",
		"```\n# fenced\n```\n# outside",
		string([]byte{'#', ' ', 0x00, 0xff}),
	}
	for _, s := range seeds {
		f.Add([]byte(s))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = ParseHeadings(data)
	})
}
