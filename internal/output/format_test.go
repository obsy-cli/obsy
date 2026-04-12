package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestNewRow(t *testing.T) {
	r := NewRow("path", "foo.md", "count", 3)
	if len(r.Fields) != 2 {
		t.Fatalf("fields len = %d, want 2", len(r.Fields))
	}
	if r.Fields[0] != "path" || r.Fields[1] != "count" {
		t.Errorf("unexpected field order: %v", r.Fields)
	}
	if r.Values["path"] != "foo.md" {
		t.Errorf("path = %v, want foo.md", r.Values["path"])
	}
	if r.Values["count"] != 3 {
		t.Errorf("count = %v, want 3", r.Values["count"])
	}
}

func TestNewRow_OddArgs(t *testing.T) {
	// Trailing key without value is silently dropped (i+1 < len check).
	r := NewRow("key")
	if len(r.Fields) != 0 {
		t.Errorf("expected no fields for odd args, got %v", r.Fields)
	}
}

func TestPrint_Empty(t *testing.T) {
	var buf bytes.Buffer
	for _, fmt := range []string{"text", "json", "tsv", "csv"} {
		buf.Reset()
		if err := Print(&buf, fmt, nil); err != nil {
			t.Errorf("[%s] unexpected error: %v", fmt, err)
		}
		if buf.Len() != 0 {
			t.Errorf("[%s] expected empty output for nil rows", fmt)
		}
	}
}

func TestPrint_Text(t *testing.T) {
	rows := []Row{
		NewRow("path", "a.md", "extra", "ignored"),
		NewRow("path", "b.md", "extra", "also ignored"),
	}
	var buf bytes.Buffer
	if err := Print(&buf, "text", rows); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 || lines[0] != "a.md" || lines[1] != "b.md" {
		t.Errorf("text output = %q", buf.String())
	}
}

func TestPrint_JSON(t *testing.T) {
	rows := []Row{
		NewRow("path", "a.md", "count", 2),
	}
	var buf bytes.Buffer
	if err := Print(&buf, "json", rows); err != nil {
		t.Fatal(err)
	}
	var out []map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if len(out) != 1 {
		t.Fatalf("json array len = %d, want 1", len(out))
	}
	if out[0]["path"] != "a.md" {
		t.Errorf("path = %v, want a.md", out[0]["path"])
	}
	if out[0]["count"].(float64) != 2 {
		t.Errorf("count = %v, want 2", out[0]["count"])
	}
}

func TestPrint_TSV(t *testing.T) {
	rows := []Row{
		NewRow("path", "a.md", "count", 2),
		NewRow("path", "b.md", "count", 5),
	}
	var buf bytes.Buffer
	if err := Print(&buf, "tsv", rows); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	// header + 2 data rows
	if len(lines) != 3 {
		t.Fatalf("tsv lines = %d, want 3:\n%s", len(lines), buf.String())
	}
	if lines[0] != "path\tcount" {
		t.Errorf("header = %q, want %q", lines[0], "path\tcount")
	}
	if lines[1] != "a.md\t2" {
		t.Errorf("row 1 = %q, want %q", lines[1], "a.md\t2")
	}
}

func TestPrint_CSV(t *testing.T) {
	rows := []Row{
		NewRow("path", "my,file.md", "count", 1),
	}
	var buf bytes.Buffer
	if err := Print(&buf, "csv", rows); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("csv lines = %d, want 2:\n%s", len(lines), buf.String())
	}
	if lines[0] != "path,count" {
		t.Errorf("header = %q, want %q", lines[0], "path,count")
	}
	// CSV encodes comma-containing values with quotes.
	if lines[1] != `"my,file.md",1` {
		t.Errorf("row = %q, want %q", lines[1], `"my,file.md",1`)
	}
}

func TestPrint_UnknownFormat_FallsBackToText(t *testing.T) {
	rows := []Row{NewRow("path", "x.md")}
	var buf bytes.Buffer
	if err := Print(&buf, "UNKNOWN", rows); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "x.md" {
		t.Errorf("expected text fallback, got %q", buf.String())
	}
}

func TestMsg_Quiet(t *testing.T) {
	var buf bytes.Buffer
	Msg(&buf, true, "hello %s", "world")
	if buf.Len() != 0 {
		t.Errorf("expected no output in quiet mode, got %q", buf.String())
	}
}

func TestMsg_NotQuiet(t *testing.T) {
	var buf bytes.Buffer
	Msg(&buf, false, "hello %s", "world")
	if buf.String() != "hello world\n" {
		t.Errorf("got %q, want %q", buf.String(), "hello world\n")
	}
}
