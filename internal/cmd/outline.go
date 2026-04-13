package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/obsy-cli/obsy/internal/output"
	"github.com/obsy-cli/obsy/internal/parser"
	"github.com/spf13/cobra"
)

func init() {
	var total bool

	cmd := &cobra.Command{
		Use:   "outline <file>",
		Short: "Show heading structure of a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, idx, err := openVault()
			if err != nil {
				return err
			}
			file, ok := idx.ResolveFileArg(args[0])
			if !ok {
				return fmt.Errorf("file not found: %s", args[0])
			}
			abs := filepath.Join(idx.VaultRoot, file)
			content, err := os.ReadFile(abs)
			if err != nil {
				return err
			}
			_, body, _ := parser.ParseFrontmatter(content)
			headings := parser.ParseHeadings(body)
			if len(headings) == 0 {
				return noResults()
			}
			if total {
				return totalOnly(len(headings))
			}

			if cfg.Format == "text" {
				for _, h := range headings {
					indent := strings.Repeat("  ", h.Level-1)
					fmt.Fprintf(os.Stdout, "%s%s %s\n", indent, strings.Repeat("#", h.Level), h.Text)
				}
				return nil
			}

			var rows []output.Row
			for _, h := range headings {
				rows = append(rows, output.NewRow("level", h.Level, "text", h.Text))
			}
			return output.Print(os.Stdout, cfg.Format, rows)
		},
	}

	cmd.Flags().BoolVar(&total, "total", false, "print heading count only")
	rootCmd.AddCommand(cmd)
}
