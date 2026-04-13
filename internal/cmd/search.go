package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/obsy-cli/obsy/internal/ops"
	"github.com/obsy-cli/obsy/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	var (
		context       bool
		path          string
		limit         int
		total         bool
		caseSensitive bool
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Full-text search across the vault (reads from disk)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := openVaultOnly()
			if err != nil {
				return err
			}
			results, err := ops.Search(v, args[0], path, limit, context, caseSensitive)
			if err != nil {
				return err
			}
			if len(results) == 0 {
				return noResults()
			}
			if total {
				return totalOnly(len(results))
			}

			if context {
				if cfg.Format == "text" {
					for _, r := range results {
						for _, line := range r.Context {
							fmt.Fprintf(os.Stdout, "%s: %s\n", r.Path, line)
						}
					}
					return nil
				}
				var rows []output.Row
				for _, r := range results {
					rows = append(rows, output.NewRow(
						"path", r.Path,
						"lines", strings.Join(r.Context, "\n"),
					))
				}
				return output.Print(os.Stdout, cfg.Format, rows)
			}

			var rows []output.Row
			for _, r := range results {
				rows = append(rows, output.NewRow("path", r.Path))
			}
			return output.Print(os.Stdout, cfg.Format, rows)
		},
	}

	cmd.Flags().BoolVar(&context, "context", false, "include matching lines (grep-style)")
	cmd.Flags().StringVar(&path, "path", "", "limit to folder")
	cmd.Flags().IntVar(&limit, "limit", 0, "max results (0 = unlimited)")
	cmd.Flags().BoolVar(&total, "total", false, "print match count only")
	cmd.Flags().BoolVar(&caseSensitive, "case-sensitive", false, "case-sensitive match")

	rootCmd.AddCommand(cmd)
}
