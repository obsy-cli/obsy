package parser

import (
	"testing"
)

func TestParseTasks(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []Task
	}{
		{
			name:  "incomplete task",
			input: "- [ ] Buy milk",
			want:  []Task{{Text: "Buy milk", Done: false, Line: 1}},
		},
		{
			name:  "complete task lowercase x",
			input: "- [x] Done thing",
			want:  []Task{{Text: "Done thing", Done: true, Line: 1}},
		},
		{
			name:  "complete task uppercase X",
			input: "- [X] Also done",
			want:  []Task{{Text: "Also done", Done: true, Line: 1}},
		},
		{
			name:  "asterisk list marker",
			input: "* [ ] Asterisk task",
			want:  []Task{{Text: "Asterisk task", Done: false, Line: 1}},
		},
		{
			name:  "indented task",
			input: "  - [ ] Nested task",
			want:  []Task{{Text: "Nested task", Done: false, Line: 1}},
		},
		{
			name:  "mixed tasks",
			input: "- [ ] Todo\n- [x] Done\n- [ ] Another",
			want: []Task{
				{Text: "Todo", Done: false, Line: 1},
				{Text: "Done", Done: true, Line: 2},
				{Text: "Another", Done: false, Line: 3},
			},
		},
		{
			name:  "not a task — plain list item",
			input: "- Just a list item",
			want:  nil,
		},
		{
			name:  "not a task — no space after bracket",
			input: "- [ ]No space",
			want:  nil,
		},
		{
			name:  "ignored inside fenced code block",
			input: "```\n- [ ] Inside fence\n```\n- [ ] Outside",
			want:  []Task{{Text: "Outside", Done: false, Line: 4}},
		},
		{
			name:  "correct line numbers",
			input: "\n\n- [ ] Line three",
			want:  []Task{{Text: "Line three", Done: false, Line: 3}},
		},
		{
			name:  "no tasks",
			input: "# Heading\n\nSome text.",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTasks([]byte(tt.input))
			if len(got) != len(tt.want) {
				t.Fatalf("got %d tasks, want %d\ngot:  %+v\nwant: %+v", len(got), len(tt.want), got, tt.want)
			}
			for i, g := range got {
				w := tt.want[i]
				if g.Text != w.Text {
					t.Errorf("[%d] Text = %q, want %q", i, g.Text, w.Text)
				}
				if g.Done != w.Done {
					t.Errorf("[%d] Done = %v, want %v", i, g.Done, w.Done)
				}
				if g.Line != w.Line {
					t.Errorf("[%d] Line = %d, want %d", i, g.Line, w.Line)
				}
			}
		})
	}
}
