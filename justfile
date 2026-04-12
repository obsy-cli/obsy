# obsy — Task Runner
# Run `just --list` to see available commands

set dotenv-load := false

root   := justfile_directory()
bin    := root / "bin" / "obsy"
pkg    := "./cmd/obsy"

# Show available commands
@help:
    just --list --unsorted

# ─── Setup ────────────────────────────────────────────────

# Check that all required tools are installed
[group("setup")]
check:
    #!/usr/bin/env bash
    ok=true
    check() {
        if command -v "$1" &>/dev/null; then
            version=$("$1" version 2>&1 | head -1 || "$1" --version 2>&1 | head -1)
            printf "  %-24s ✓  %s\n" "$1" "$version"
        else
            printf "  %-24s ✗  missing — %s\n" "$1" "$2"
            ok=false
        fi
    }
    check go       "https://go.dev/dl/"
    check golangci-lint "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    $ok || exit 1

# ─── Build ────────────────────────────────────────────────

# Build the binary to ./bin/obsy
[group("build")]
build:
    CGO_ENABLED=0 go build -o {{ bin }} {{ pkg }}

# Install binary to $GOPATH/bin (or ~/go/bin)
[group("build")]
install:
    CGO_ENABLED=0 go install {{ pkg }}

# Cross-compile for all release targets → ./dist/
[group("build")]
build-all:
    #!/usr/bin/env bash
    set -euo pipefail
    dist="{{ root }}/dist"
    mkdir -p "$dist"
    targets=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
    )
    for target in "${targets[@]}"; do
        os="${target%/*}"
        arch="${target#*/}"
        out="$dist/obsy-${os}-${arch}"
        printf "  building %-28s → %s\n" "$target" "${out#{{ root }}/}"
        CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build -o "$out" {{ pkg }}
    done
    echo "done."

# Remove build artifacts
[group("build")]
clean:
    rm -rf {{ root }}/bin {{ root }}/dist {{ root }}/coverage.out {{ root }}/coverage.html

# ─── Test ─────────────────────────────────────────────────

# Run all tests
[group("test")]
test:
    go test ./...

# Run all tests with race detector
[group("test")]
test-race:
    go test -race ./...

# Run tests for a specific package  (e.g. just test-pkg internal/parser)
[group("test")]
test-pkg pkg:
    go test ./{{ pkg }}/...

# Run a single named test  (e.g. just test-run TestWikilinks)
[group("test")]
test-run name:
    go test ./... -run {{ name }} -v

# Run fuzz tests for all parsers  (e.g. just fuzz 30s)
[group("test")]
fuzz duration="10s":
    go test -fuzz=FuzzParseLinks       -fuzztime={{ duration }} ./internal/parser/
    go test -fuzz=FuzzParseInlineTags  -fuzztime={{ duration }} ./internal/parser/
    go test -fuzz=FuzzParseFrontmatter -fuzztime={{ duration }} ./internal/parser/
    go test -fuzz=FuzzParseTasks       -fuzztime={{ duration }} ./internal/parser/
    go test -fuzz=FuzzParseHeadings    -fuzztime={{ duration }} ./internal/parser/

# Generate coverage report and open in browser
[group("test")]
coverage:
    go test -coverprofile={{ root }}/coverage.out ./...
    go tool cover -html={{ root }}/coverage.out -o {{ root }}/coverage.html
    xdg-open {{ root }}/coverage.html 2>/dev/null || open {{ root }}/coverage.html 2>/dev/null || true

# ─── Code quality ─────────────────────────────────────────

# Run go vet + golangci-lint
[group("lint")]
lint:
    go vet ./...
    golangci-lint run ./...

# Format all Go source files
[group("lint")]
fmt:
    gofmt -w .

# Tidy go.mod / go.sum
[group("lint")]
tidy:
    go mod tidy

# ─── Dev ──────────────────────────────────────────────────

# Build and run with arbitrary args  (e.g. just run search "foo" --vault=~/wiki)
[group("dev")]
run *args: build
    {{ bin }} {{ args }}

# ─── Smoke ────────────────────────────────────────────────

vault := root / "testdata" / "vault"

# End-to-end smoke test — builds the binary and runs every command against testdata/vault
[group("test")]
smoke: build
    #!/usr/bin/env bash
    set -euo pipefail
    bin="{{ bin }}"
    vault="{{ vault }}"
    ok=0; fail=0

    check() {
        local desc="$1"; shift
        if "$bin" --vault="$vault" "$@" >/dev/null 2>&1; then
            printf "  ✓  %s\n" "$desc"
            (( ok++ )) || true
        else
            local code=$?
            # exit 1 means "no results" — that's a valid success for queries
            if [ "$code" -eq 1 ]; then
                printf "  ✓  %s  (no results)\n" "$desc"
                (( ok++ )) || true
            else
                printf "  ✗  %s  (exit %d)\n" "$desc" "$code"
                (( fail++ )) || true
            fi
        fi
    }

    check_output() {
        local desc="$1"; local pattern="$2"; shift 2
        local out
        out=$("$bin" --vault="$vault" "$@" 2>&1) || true
        if echo "$out" | grep -q "$pattern"; then
            printf "  ✓  %s\n" "$desc"
            (( ok++ )) || true
        else
            printf "  ✗  %s  (pattern '%s' not found in output)\n" "$desc" "$pattern"
            printf "     output: %s\n" "$(echo "$out" | head -3)"
            (( fail++ )) || true
        fi
    }

    echo "── files ─────────────────────────────────────"
    check         "files"                         files
    check         "files --format=json"           files --format=json
    check         "files --format=tsv"            files --format=tsv
    check_output  "files count=12"      "12"      files --total

    echo "── search ────────────────────────────────────"
    check         "search basic"                  search "science"
    check         "search --context"              search "science" --context
    check         "search --case-sensitive"       search "Science" --case-sensitive
    check         "search --path"                 search "note" --path=sub/
    check         "search no results (exit 1)"    search "zzznomatch"

    echo "── graph ─────────────────────────────────────"
    check         "unresolved"                    unresolved
    check         "orphans"                       orphans
    check         "deadends"                      deadends
    check_output  "backlinks to note-a"  "index"  backlinks note-a
    check         "links from index"              links index.md
    check         "links --resolve"               links index.md --resolve

    echo "── tags / properties / tasks ─────────────────"
    check         "tags"                          tags
    check         "tag science"                   tag science
    check         "properties"                    properties
    check         "tasks"                         tasks
    check         "tasks --todo"                  tasks --todo
    check         "tasks --file=note-a.md"        tasks --file=note-a.md

    echo "── read / outline / status ───────────────────"
    check         "read note-a"                   read note-a
    check         "outline note-a"                outline note-a
    check         "status"                        status

    echo "── formats ───────────────────────────────────"
    check_output  "unresolved json"  "does-not-exist"  unresolved --format=json
    check_output  "tags tsv header"  "tag"             tags --format=tsv
    check_output  "tags csv header"  "tag"             tags --format=csv

    echo
    printf "  %d passed, %d failed\n" "$ok" "$fail"
    [ "$fail" -eq 0 ] || exit 2
