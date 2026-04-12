package parser

import (
	"testing"
)

func TestParseHeadings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Heading
	}{
		{
			name:  "h1",
			input: "# Title",
			want:  []Heading{{Level: 1, Text: "Title", Line: 1}},
		},
		{
			name:  "h2 through h6",
			input: "## Two\n### Three\n#### Four\n##### Five\n###### Six",
			want: []Heading{
				{Level: 2, Text: "Two", Line: 1},
				{Level: 3, Text: "Three", Line: 2},
				{Level: 4, Text: "Four", Line: 3},
				{Level: 5, Text: "Five", Line: 4},
				{Level: 6, Text: "Six", Line: 5},
			},
		},
		{
			name:  "more than 6 hashes — not a heading",
			input: "####### Too many",
			want:  nil,
		},
		{
			name:  "no space after hash — not a heading",
			input: "#nospace",
			want:  nil,
		},
		{
			name:  "ignored inside fenced code block",
			input: "```\n# Inside fence\n```\n# Outside",
			want:  []Heading{{Level: 1, Text: "Outside", Line: 4}},
		},
		{
			name:  "whitespace trimmed from heading text",
			input: "##  Padded  ",
			want:  []Heading{{Level: 2, Text: "Padded", Line: 1}},
		},
		{
			name:  "correct line number",
			input: "Text\n\n## Heading on line 3",
			want:  []Heading{{Level: 2, Text: "Heading on line 3", Line: 3}},
		},
		{
			name:  "no headings",
			input: "Plain text\n- list item",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseHeadings([]byte(tt.input))
			if len(got) != len(tt.want) {
				t.Fatalf("got %d headings, want %d\ngot:  %+v\nwant: %+v", len(got), len(tt.want), got, tt.want)
			}
			for i, g := range got {
				w := tt.want[i]
				if g.Level != w.Level {
					t.Errorf("[%d] Level = %d, want %d", i, g.Level, w.Level)
				}
				if g.Text != w.Text {
					t.Errorf("[%d] Text = %q, want %q", i, g.Text, w.Text)
				}
				if g.Line != w.Line {
					t.Errorf("[%d] Line = %d, want %d", i, g.Line, w.Line)
				}
			}
		})
	}
}
