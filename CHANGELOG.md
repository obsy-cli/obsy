# Changelog

All notable changes to this project will be documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versions follow [Semantic Versioning](https://semver.org/).

---

## [0.1.0] — 2026-04-13

Initial release. A self-contained CLI for querying Obsidian-compatible markdown vaults — no desktop app, no daemon, no CGO.

### Commands

- **`files`** — list all vault files; filter by folder, sort by name or modification time
- **`search`** — full-text search across vault content with optional grep-style context lines, path filter, and case-sensitive mode
- **`read`** — print file contents with frontmatter stripped; accepts bare name, partial path, or wikilink-style input
- **`outline`** — show the heading hierarchy of a file
- **`links`** — list outgoing wikilinks from a file, with optional path resolution
- **`backlinks`** — list all files that link to a given file; alias-aware
- **`unresolved`** — list broken wikilinks and unresolvable embeds (e.g. `![[image.png]]`) across the vault
- **`orphans`** — list files with no incoming links
- **`deadends`** — list files with no outgoing links
- **`tags`** — list all tags (frontmatter + inline `#tag`); sortable by name or frequency
- **`tag`** — list files carrying a specific tag
- **`properties`** — list all frontmatter property names across the vault
- **`property`** — list per-file values for a frontmatter key
- **`tasks`** — list `- [ ]` / `- [x]` task items; filter by status, file, or folder
- **`move`** — move a file and rewrite all wikilinks that reference it vault-wide
- **`rename`** — rename a file in place and rewrite all wikilinks vault-wide
- **`reindex`** — force a full cache rebuild
- **`status`** — print an index health summary (file count, link count, tags, tasks, cache location)

### Features

- **Vault discovery** — resolves vault from `--vault` flag, current directory, or by walking up to `$HOME` looking for `.obsidian/`
- **Incremental index** — gob-encoded cache at `~/.cache/obsy/<vault-id>/index.gob`; only re-parses files whose mtime has changed
- **Wikilink resolution** — exact path, shortest-depth basename, same-folder tiebreaker, alias lookup; matches Obsidian's resolution semantics
- **Output formats** — `text`, `json`, `tsv`, `csv` on every command via `--format`
- **Exit codes** — `0` success, `1` no results, `2` error; safe to use in shell conditionals and scripts
- **Zero CGO** — pure Go; single static binary, no runtime dependencies

### Pre-built binaries

Linux and macOS binaries for `amd64` and `arm64` are attached to this release. See the [README](README.md) for installation instructions.
