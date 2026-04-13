# obsy

A standalone CLI for [Obsidian](https://obsidian.md)-compatible markdown vaults.

## Why this exists

[Andrej Karpathy's model](https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f#file-llm-wiki-md) for a personal LLM knowledge base revolves around a folder of plaintext markdown files (Obsidian vault) linked together with wikilinks (`[[note]]`). The idea is simple — keep your knowledge in a format that both humans and LLMs can read, navigate, and reason over.

Obsidian ships a CLI (`obsidian`) but it only works as a sidecar to the running desktop application. No desktop app, no CLI. That makes it useless in headless environments — servers, CI pipelines, editor terminals, or any agent loop that needs to query the vault without spawning a GUI.

**obsy is a self-contained replacement.** It reads the vault directly from disk, maintains its own index cache, and exposes the full query surface — search, link graph, tags, tasks, properties — as a composable CLI with machine-readable output formats. No Obsidian app required. No daemon. No CGO.

---

## Installation

### go install (recommended)

```bash
go install github.com/obsy-cli/obsy/cmd/obsy@latest
```

Requires Go 1.24+. Installs to `$(go env GOPATH)/bin/obsy`.

### Pre-built binaries (curl)

Download the latest release for your platform from GitHub Releases:

```bash
# Linux amd64
curl -fsSL https://github.com/obsy-cli/obsy/releases/latest/download/obsy-linux-amd64 \
  -o ~/.local/bin/obsy && chmod +x ~/.local/bin/obsy

# Linux arm64
curl -fsSL https://github.com/obsy-cli/obsy/releases/latest/download/obsy-linux-arm64 \
  -o ~/.local/bin/obsy && chmod +x ~/.local/bin/obsy

# macOS amd64 (Intel)
curl -fsSL https://github.com/obsy-cli/obsy/releases/latest/download/obsy-darwin-amd64 \
  -o ~/.local/bin/obsy && chmod +x ~/.local/bin/obsy

# macOS arm64 (Apple Silicon)
curl -fsSL https://github.com/obsy-cli/obsy/releases/latest/download/obsy-darwin-arm64 \
  -o ~/.local/bin/obsy && chmod +x ~/.local/bin/obsy
```

`~/.local/bin` is on `$PATH` by default on most Linux distros and macOS. If it isn't, add `export PATH="$HOME/.local/bin:$PATH"` to your shell profile. Alternatively, use `sudo` to install system-wide to `/usr/local/bin`.

Or download manually from the [Releases page](https://github.com/obsy-cli/obsy/releases).

### Build from source

```bash
git clone https://github.com/obsy-cli/obsy
cd obsy
just build        # → ./bin/obsy
just install      # → $GOPATH/bin/obsy
```

---

## Vault discovery

Every command needs to know which vault to operate on. Resolution order:

1. `--vault=/path/to/vault` flag
2. Current working directory
3. Walk up toward `$HOME` looking for a `.obsidian/` directory

```bash
obsy status                         # vault auto-discovered from cwd
obsy --vault=~/wiki status          # explicit vault
```

Excluded from all scans: `.obsidian/`, `.git/`, `.trash/`, any hidden path (starts with `.`).

---

## Index and cache

obsy maintains a gob-encoded index at `~/.cache/obsy/<vault-id>/index.gob` (respects `$XDG_CACHE_HOME`). The index is updated incrementally on each run — only files whose mtime has changed are re-parsed.

```bash
obsy status          # show index health, file count, cache location
obsy reindex         # force full rebuild
obsy --no-cache ...  # skip cache, scan fresh (does not update cache)
```

The vault-id is the first 16 hex characters of the SHA-256 of the vault's absolute path, so multiple vaults never collide.

---

## Output formats

All commands support `--format`:

| Format | Description |
|--------|-------------|
| `text` | Human-readable, one result per line (default) |
| `json` | Array of objects; field names match TSV/CSV column names |
| `tsv`  | Tab-separated values with header row |
| `csv`  | Comma-separated values with header row |

```bash
obsy tags --format=json
obsy tasks --format=tsv
obsy unresolved --format=csv
```

---

## Global flags

| Flag | Default | Description |
|------|---------|-------------|
| `--vault string` | (auto) | Vault root path |
| `--format string` | `text` | Output format: `text`, `json`, `tsv`, `csv` |
| `--no-cache` | false | Skip index cache, force fresh scan |
| `--quiet` | false | Suppress non-essential stderr output |

---

## Exit codes

| Code | Meaning |
|------|---------|
| `0` | Success — results found or operation completed |
| `1` | No results — query returned nothing |
| `2` | Error — vault not found, invalid arguments, etc. |

This makes obsy composable in shell scripts:

```bash
obsy unresolved --quiet && echo "vault is clean" || echo "broken links found"
```

---

## Commands

### `files` — list vault files

```
obsy files [flags]
```

List all markdown files in the vault.

| Flag | Default | Description |
|------|---------|-------------|
| `--folder string` | | Filter by folder prefix |
| `--sort string` | `name` | Sort by `name` or `modified` (most recent first) |
| `--limit int` | `0` | Max results (0 = unlimited) |
| `--total` | false | Print count only |

```bash
obsy files
obsy files --folder=journal/
obsy files --sort=modified --limit=10
obsy files --total
```

---

### `search` — full-text search

```
obsy search <query> [flags]
```

Searches file content, headings, tags, and frontmatter values. Reads directly from disk (not the index cache).

| Flag | Default | Description |
|------|---------|-------------|
| `--context` | false | Include matching lines (grep-style) |
| `--path string` | | Limit to folder |
| `--case-sensitive` | false | Case-sensitive matching |
| `--limit int` | `0` | Max results |
| `--total` | false | Print count only |

```bash
obsy search "neural networks"
obsy search "todo" --path=inbox/ --context
obsy search "Epictetus" --case-sensitive
obsy search "meeting" --format=json
```

---

### `read` — display file contents

```
obsy read <file>
```

Print file contents with frontmatter stripped. Accepts any form that would work as a wikilink: bare name, partial path, or full relative path.

```bash
obsy read stoicism          # finds philosophy/stoicism.md
obsy read "meeting notes"   # fuzzy basename match
obsy read journal/2024-01-15.md
```

---

### `outline` — heading structure

```
obsy outline <file> [flags]
```

Show the heading hierarchy of a file.

| Flag | Default | Description |
|------|---------|-------------|
| `--total` | false | Print heading count only |

```bash
obsy outline stoicism
obsy outline index.md --format=json
```

**Text output:**
```
# Stoicism
  ## Core Principles
    ### The Dichotomy of Control
  ## Practitioners
```

**JSON output:**
```json
[
  {"level": 1, "text": "Stoicism"},
  {"level": 2, "text": "Core Principles"},
  {"level": 3, "text": "The Dichotomy of Control"},
  {"level": 2, "text": "Practitioners"}
]
```

---

### `links` — outgoing links from a file

```
obsy links <file> [flags]
```

List all wikilinks and embeds that a file points to.

| Flag | Default | Description |
|------|---------|-------------|
| `--resolve` | false | Show resolved file paths; marks broken links as `[unresolved]` |
| `--total` | false | Print count only |

```bash
obsy links index.md
obsy links index.md --resolve
obsy links index.md --format=json
```

**Without `--resolve`:**
```
note-a
dead-end
sub/child
```

**With `--resolve`:**
```
note-a → note-a.md
dead-end → dead-end.md
sub/child → sub/child.md
broken-ref → [unresolved]
```

---

### `backlinks` — incoming links to a file

```
obsy backlinks <file> [flags]
```

List all files that contain a wikilink pointing to the given file. Alias resolution is applied — a link `[[nota]]` targeting an alias of `note-a.md` counts as a backlink to `note-a.md`.

| Flag | Default | Description |
|------|---------|-------------|
| `--counts` | false | Include link count per source file |
| `--total` | false | Print count only |

```bash
obsy backlinks stoicism
obsy backlinks index.md --counts
```

---

### `unresolved` — broken links

```
obsy unresolved [flags]
```

List all wikilinks that cannot be resolved to a file in the vault. Both `.md` wikilinks (resolved via index) and non-`.md` embeds (e.g., `![[image.png]]`, resolved via `stat`) are checked.

| Flag | Default | Description |
|------|---------|-------------|
| `--path string` | | Limit to folder |
| `--counts` | false | Include occurrence count per broken link |
| `--verbose` | false | Show source file for each occurrence |
| `--total` | false | Print count only |

```bash
obsy unresolved
obsy unresolved --verbose
obsy unresolved --path=drafts/ --format=json
```

---

### `orphans` — files with no incoming links

```
obsy orphans [flags]
```

List files that no other file links to. Useful for finding notes that have drifted out of the graph.

| Flag | Default | Description |
|------|---------|-------------|
| `--ignore string` | | Exclude files matching glob pattern (e.g., `*/index.md`) |
| `--total` | false | Print count only |

```bash
obsy orphans
obsy orphans --ignore="*/index.md"
```

---

### `deadends` — files with no outgoing links

```
obsy deadends [flags]
```

List files that contain no wikilinks or embeds. These are leaf nodes in the knowledge graph.

| Flag | Default | Description |
|------|---------|-------------|
| `--total` | false | Print count only |

```bash
obsy deadends
obsy deadends --format=json
```

---

### `tags` — list all tags

```
obsy tags [flags]
```

List every tag across the vault. Captures both frontmatter tags (`tags: [science, reference]`) and inline tags (`#science`).

| Flag | Default | Description |
|------|---------|-------------|
| `--counts` | false | Include occurrence count per tag |
| `--sort string` | `name` | Sort by `name` or `count` |
| `--path string` | | Limit to folder |
| `--total` | false | Print unique tag count only |

```bash
obsy tags
obsy tags --counts --sort=count
obsy tags --path=journal/ --format=json
```

---

### `tag` — files with a specific tag

```
obsy tag <name> [flags]
```

List all files that carry a given tag. The `#` prefix is optional.

| Flag | Default | Description |
|------|---------|-------------|
| `--path string` | | Limit to folder |
| `--total` | false | Print count only |

```bash
obsy tag science
obsy tag "#reference" --path=papers/
```

---

### `properties` — frontmatter property names

```
obsy properties [flags]
```

List all YAML frontmatter keys that appear across the vault.

| Flag | Default | Description |
|------|---------|-------------|
| `--counts` | false | Include occurrence count per key |
| `--total` | false | Print unique property count only |

```bash
obsy properties
obsy properties --counts
```

---

### `property` — values for a property

```
obsy property <name> [flags]
```

List every file and its value for the given frontmatter key.

| Flag | Default | Description |
|------|---------|-------------|
| `--path string` | | Limit to folder |
| `--file string` | | Read value from a single file |

```bash
obsy property author
obsy property created --format=tsv
obsy property status --file=index.md
```

---

### `tasks` — list tasks across the vault

```
obsy tasks [flags]
```

List all Obsidian-style task items (`- [ ] ...` / `- [x] ...`) across the vault.

| Flag | Default | Description |
|------|---------|-------------|
| `--todo` | false | Incomplete tasks only |
| `--done` | false | Completed tasks only |
| `--file string` | | Tasks from a specific file |
| `--path string` | | Tasks from a folder |
| `--verbose` | false | Group output by file with line numbers |
| `--total` | false | Print task count only |

```bash
obsy tasks
obsy tasks --todo
obsy tasks --done --format=json
obsy tasks --path=projects/ --todo --verbose
obsy tasks --file=inbox.md
obsy tasks --total
```

**Default text output:**
```
[ ] index.md:12: Write introduction
[x] index.md:17: Review outline
[ ] notes/inbox.md:3: Process fleeting notes
```

**With `--verbose`:**
```
index.md
  [ ] L12     Write introduction
  [x] L17     Review outline

notes/inbox.md
  [ ] L3      Process fleeting notes
```

**JSON output:**
```json
[
  {"file": "index.md", "line": 12, "done": false, "text": "Write introduction"},
  {"file": "index.md", "line": 17, "done": true,  "text": "Review outline"},
  {"file": "notes/inbox.md", "line": 3, "done": false, "text": "Process fleeting notes"}
]
```

---

### `move` — move a file and update links

```
obsy move <source> <destination>
```

Move a file to a new location and rewrite every wikilink that referenced it across the entire vault. All writes are validated before any file is touched.

- Destination can be a folder (`archives/`) or a full path (`archives/old-note.md`)
- Trailing `/` or an existing directory path → file moves into that folder keeping its name
- Renames and cross-folder moves can be combined in a single command

```bash
obsy move inbox.md journal/
obsy move drafts/essay.md published/final-essay.md
```

---

### `rename` — rename a file and update links

```
obsy rename <source> <new-name>
```

Rename a file in place and update all wikilinks vault-wide. The `.md` extension is preserved if omitted from `<new-name>`.

```bash
obsy rename old-name new-name
obsy rename "meeting notes" "2024-01-15-meeting"
```

---

### `reindex` — rebuild the index

```
obsy reindex
```

Force a full re-scan of the vault and write a fresh cache. Useful after bulk external edits that may have bypassed incremental detection.

```bash
obsy reindex
obsy reindex --quiet
```

---

### `status` — index health report

```
obsy status
```

Print a summary of the vault and cache state.

```
vault:      /home/user/wiki
files:      427
links:      1243
tags:       84 unique
properties: 12 unique
tasks:      156
cache:      /home/user/.cache/obsy/96ebdb84ab637a3c/index.gob (3.2 MB)
scanned:    2024-04-13 14:23:45
updated:    2024-04-13 14:23:45
```

---

## Wikilink resolution

obsy resolves wikilinks the same way Obsidian does:

1. **Exact path match** — `[[sub/note]]` matches `sub/note.md`
2. **Basename match** — `[[note]]` matches any file named `note.md`, choosing the shallowest path on collision
3. **Alias match** — `[[nota]]` resolves to the file whose frontmatter `aliases` list contains `nota`
4. **Case-insensitive** — matching is case-insensitive for basename and alias lookup

Ambiguous basename collisions (same filename at different depths) resolve to the shallower path, mirroring Obsidian's behavior. Alias collisions (same alias claimed by two files) produce a warning to stderr and resolve deterministically to the lexicographically first file path.

---

## LLM usage patterns

obsy is designed to feed structured vault data to LLM workflows:

```bash
# Give an LLM the full topic graph
obsy links "topic/machine-learning.md" --resolve --format=json

# Feed all tasks into a planning prompt
obsy tasks --todo --format=json | llm "organize these into a weekly plan"

# Check knowledge coverage before generating
obsy tag "physics" --format=json | jq length

# Find gaps in the knowledge graph
obsy orphans --format=json
obsy unresolved --format=json

# Build a context window from a topic and its linked notes
obsy read "machine-learning"
obsy backlinks "machine-learning" --format=json | jq -r '.[].file' | xargs -I{} obsy read {}
```

---

## Technical notes

- **Zero CGO.** Pure Go. All binaries compile with `CGO_ENABLED=0`.
- **No runtime dependencies.** No external programs, no daemons, no Obsidian app.
- **Parser is minimal by design.** Frontmatter, links, tags, tasks, and headings are extracted with targeted parsers — no full markdown AST. This keeps parsing fast and the binary small.
- **Cache is vault-local.** Each vault gets its own isolated cache keyed by its absolute path. Safe to use across multiple vaults simultaneously.
- **Links inside code blocks and inline code are ignored**, matching Obsidian's behavior.
