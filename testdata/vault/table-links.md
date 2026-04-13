---
tags: [table]
---

# Table Links

Wikilinks inside a markdown table require \| to escape the pipe so the table
parser doesn't split on it. This file exercises that syntax.

| Note | Description |
|---|---|
| [[note-a\|first note]] | Link with escaped pipe display text |
| [[note-b\|orphan note]] | Another escaped pipe link |
| [[sub/child\|child note]] | Path-qualified with escaped pipe |
