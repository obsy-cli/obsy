package parser

import (
	"reflect"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantTags    []string
		wantAliases []string
		wantBody    string
		wantErr     bool
	}{
		{
			name:     "no frontmatter",
			input:    "# Hello\n\nBody text.",
			wantBody: "# Hello\n\nBody text.",
		},
		{
			name:     "empty frontmatter",
			input:    "---\n---\nBody.",
			wantBody: "Body.",
		},
		{
			name:        "tags as list",
			input:       "---\ntags: [go, cli]\n---\nBody.",
			wantTags:    []string{"go", "cli"},
			wantAliases: nil,
			wantBody:    "Body.",
		},
		{
			name:     "tags as string",
			input:    "---\ntags: go\n---\nBody.",
			wantTags: []string{"go"},
			wantBody: "Body.",
		},
		{
			name:        "aliases",
			input:       "---\naliases: [myalias, alt]\n---\nBody.",
			wantAliases: []string{"myalias", "alt"},
			wantBody:    "Body.",
		},
		{
			name:        "tags and aliases",
			input:       "---\ntags: [a, b]\naliases: [x]\n---\nContent here.",
			wantTags:    []string{"a", "b"},
			wantAliases: []string{"x"},
			wantBody:    "Content here.",
		},
		{
			name:     "no closing delimiter — not parsed",
			input:    "---\ntags: [a]\nno closing",
			wantBody: "---\ntags: [a]\nno closing",
		},
		{
			name:     "delimiter not at start — not parsed",
			input:    "\n---\ntags: [a]\n---\nBody.",
			wantBody: "\n---\ntags: [a]\n---\nBody.",
		},
		{
			name:     "body is empty after frontmatter",
			input:    "---\ntags: [x]\n---\n",
			wantTags: []string{"x"},
			wantBody: "",
		},
		{
			name:     "CRLF line endings",
			input:    "---\r\ntags: [crlf]\r\n---\r\nBody.",
			wantTags: []string{"crlf"},
			wantBody: "Body.",
		},
		{
			name:     "leading blank line in body preserved",
			input:    "---\ntags: [t]\n---\n\nParagraph.",
			wantTags: []string{"t"},
			wantBody: "\nParagraph.",
		},
		{
			name:     "malformed YAML returns error",
			input:    "---\ntags: [unclosed\n---\nBody.",
			wantBody: "---\ntags: [unclosed\n---\nBody.",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := ParseFrontmatter([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(fm.Tags, tt.wantTags) {
				t.Errorf("Tags = %v, want %v", fm.Tags, tt.wantTags)
			}
			if !reflect.DeepEqual(fm.Aliases, tt.wantAliases) {
				t.Errorf("Aliases = %v, want %v", fm.Aliases, tt.wantAliases)
			}
			if string(body) != tt.wantBody {
				t.Errorf("body = %q, want %q", string(body), tt.wantBody)
			}
		})
	}
}
