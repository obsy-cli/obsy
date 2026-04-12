package parser

import (
	"testing"
)

func TestParseLinks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Link
	}{
		{
			name:  "basic wikilink",
			input: "See [[note]].",
			want:  []Link{{Raw: "note", IsEmbed: false, Line: 1}},
		},
		{
			name:  "embed",
			input: "![[image.png]]",
			want:  []Link{{Raw: "image.png", IsEmbed: true, Line: 1}},
		},
		{
			name:  "link with display text",
			input: "[[note|display text]]",
			want:  []Link{{Raw: "note|display text", IsEmbed: false, Line: 1}},
		},
		{
			name:  "link with heading anchor",
			input: "[[note#heading]]",
			want:  []Link{{Raw: "note#heading", IsEmbed: false, Line: 1}},
		},
		{
			name:  "link with anchor and display",
			input: "[[note#heading|shown]]",
			want:  []Link{{Raw: "note#heading|shown", IsEmbed: false, Line: 1}},
		},
		{
			name:  "path-qualified link",
			input: "[[folder/note]]",
			want:  []Link{{Raw: "folder/note", IsEmbed: false, Line: 1}},
		},
		{
			name:  "multiple links on one line",
			input: "[[a]] and [[b]]",
			want: []Link{
				{Raw: "a", Line: 1},
				{Raw: "b", Line: 1},
			},
		},
		{
			name:  "links on different lines",
			input: "[[a]]\n\n[[b]]",
			want: []Link{
				{Raw: "a", Line: 1},
				{Raw: "b", Line: 3},
			},
		},
		{
			name:  "ignored inside fenced code block",
			input: "```\n[[inside-fence]]\n```\n[[outside]]",
			want:  []Link{{Raw: "outside", Line: 4}},
		},
		{
			name:  "ignored inside inline code",
			input: "text `[[inline-code]]` and [[real]]",
			want:  []Link{{Raw: "real", Line: 1}},
		},
		{
			name:  "embed with anchor",
			input: "![[note#section]]",
			want:  []Link{{Raw: "note#section", IsEmbed: true, Line: 1}},
		},
		{
			name:  "block id link",
			input: "[[note#^block-id]]",
			want:  []Link{{Raw: "note#^block-id", Line: 1}},
		},
		{
			name:  "no links",
			input: "Just plain text with no links.",
			want:  nil,
		},
		{
			name:  "unclosed link bracket ignored",
			input: "[[unclosed",
			want:  nil,
		},
		{
			name:  "empty link ignored",
			input: "[[]]",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseLinks([]byte(tt.input))
			if len(got) != len(tt.want) {
				t.Fatalf("got %d links, want %d\ngot:  %+v\nwant: %+v", len(got), len(tt.want), got, tt.want)
			}
			for i, g := range got {
				w := tt.want[i]
				if g.Raw != w.Raw {
					t.Errorf("[%d] Raw = %q, want %q", i, g.Raw, w.Raw)
				}
				if g.IsEmbed != w.IsEmbed {
					t.Errorf("[%d] IsEmbed = %v, want %v", i, g.IsEmbed, w.IsEmbed)
				}
				if w.Line != 0 && g.Line != w.Line {
					t.Errorf("[%d] Line = %d, want %d", i, g.Line, w.Line)
				}
			}
		})
	}
}
