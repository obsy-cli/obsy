package parser

import (
	"reflect"
	"testing"
)

func TestParseInlineTags(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "basic tag",
			input: "Text with #mytag here.",
			want:  []string{"mytag"},
		},
		{
			name:  "tag at start of line is heading — ignored",
			input: "#notag heading",
			want:  nil,
		},
		{
			name:  "heading with leading whitespace — ignored",
			input: "  # also a heading",
			want:  nil,
		},
		{
			name:  "purely numeric — not a tag",
			input: "Reference #123 here.",
			want:  nil,
		},
		{
			name:  "mixed numeric and alpha — valid tag",
			input: "Label #abc123",
			want:  []string{"abc123"},
		},
		{
			name:  "nested tag with slash",
			input: "Status: #project/active",
			want:  []string{"project/active"},
		},
		{
			name:  "tag with hyphen",
			input: "See #my-tag for details.",
			want:  []string{"my-tag"},
		},
		{
			name:  "tag with underscore",
			input: "#under_score",
			want:  nil, // at line start → heading check applies... wait no
		},
		{
			name:  "multiple tags",
			input: "Filed under #work and #personal.",
			want:  []string{"work", "personal"},
		},
		{
			name:  "tag inside fenced code block — ignored",
			input: "```\n#inside-fence\n```\nText with #outside tag.",
			want:  []string{"outside"},
		},
		{
			name:  "tag inside inline code — ignored",
			input: "Use `#ignored` but #kept",
			want:  []string{"kept"},
		},
		{
			name:  "duplicate tags deduplicated",
			input: "I love #go, really love #go.",
			want:  []string{"go"},
		},
		{
			name:  "no tags",
			input: "Just plain text.",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Fix the underscore test — it IS at line start so it's skipped as heading
			got := ParseInlineTags([]byte(tt.input))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
