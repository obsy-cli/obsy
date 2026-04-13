package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Row is an ordered map: fields defines column order, values holds the data.
type Row struct {
	Fields []string
	Values map[string]any
}

// NewRow creates a Row from alternating key, value pairs.
// e.g. NewRow("path", "foo.md", "count", 3)
func NewRow(kvs ...any) Row {
	r := Row{Values: make(map[string]any)}
	for i := 0; i+1 < len(kvs); i += 2 {
		key := fmt.Sprint(kvs[i])
		r.Fields = append(r.Fields, key)
		r.Values[key] = kvs[i+1]
	}
	return r
}

// Print writes rows to w in the given format.
// For text: prints the first field of each row.
// For json: prints a JSON array of objects.
// For tsv/csv: prints a header row followed by data rows.
func Print(w io.Writer, format string, rows []Row) error {
	if len(rows) == 0 {
		return nil
	}
	switch strings.ToLower(format) {
	case "text", "":
		return printText(w, rows)
	case "json":
		return printJSON(w, rows)
	case "tsv":
		return printSV(w, rows, '\t')
	case "csv":
		return printSV(w, rows, ',')
	default:
		return fmt.Errorf("unknown format %q: want text|json|tsv|csv", format)
	}
}

func printText(w io.Writer, rows []Row) error {
	for _, row := range rows {
		if len(row.Fields) == 0 {
			continue
		}
		fmt.Fprintln(w, fmt.Sprint(row.Values[row.Fields[0]]))
	}
	return nil
}

func printJSON(w io.Writer, rows []Row) error {
	// Build []map[string]any preserving insertion order isn't possible with
	// standard maps, but JSON consumers use field names not order.
	out := make([]map[string]any, len(rows))
	for i, row := range rows {
		m := make(map[string]any, len(row.Fields))
		for _, f := range row.Fields {
			m[f] = row.Values[f]
		}
		out[i] = m
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func printSV(w io.Writer, rows []Row, sep rune) error {
	cw := csv.NewWriter(w)
	cw.Comma = sep

	// Header row uses field names of the first row.
	headers := rows[0].Fields
	if err := cw.Write(headers); err != nil {
		return err
	}
	for _, row := range rows {
		record := make([]string, len(headers))
		for i, h := range headers {
			record[i] = fmt.Sprint(row.Values[h])
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// Msg prints a non-essential message to w, unless quiet is true.
func Msg(w io.Writer, quiet bool, format string, args ...any) {
	if !quiet {
		fmt.Fprintf(w, format+"\n", args...)
	}
}
