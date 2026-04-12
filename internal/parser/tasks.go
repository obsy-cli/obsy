package parser

import (
	"bytes"
	"strings"
)

// Task represents a checkbox task item.
type Task struct {
	Text string
	Done bool
	Line int // 1-based
}

// ParseTasks extracts - [ ] and - [x] task items from content.
func ParseTasks(content []byte) []Task {
	var tasks []Task
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

		text, done, ok := parseTaskLine(string(trimmed))
		if ok {
			tasks = append(tasks, Task{Text: text, Done: done, Line: lineNum})
		}
	}
	return tasks
}

// parseTaskLine parses "- [ ] text" or "- [x] text" (any leading whitespace
// already trimmed). Returns the task text, done status, and whether it matched.
func parseTaskLine(s string) (text string, done bool, ok bool) {
	// Strip leading list marker: "- " or "* ".
	if !strings.HasPrefix(s, "- ") && !strings.HasPrefix(s, "* ") {
		return
	}
	rest := s[2:]
	if strings.HasPrefix(rest, "[ ] ") {
		return strings.TrimSpace(rest[4:]), false, true
	}
	if strings.HasPrefix(rest, "[x] ") || strings.HasPrefix(rest, "[X] ") {
		return strings.TrimSpace(rest[4:]), true, true
	}
	return
}
