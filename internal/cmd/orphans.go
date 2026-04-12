package cmd

import (
	"github.com/obsy-cli/obsy/internal/output"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	var (
		ignore string
		total  bool
	)

	cmd := &cobra.Command{
		Use:   "orphans",
		Short: "List files with no incoming links",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, idx, err := openVault()
			if err != nil {
				return err
			}
			orphans := idx.Orphans(ignore)
			if len(orphans) == 0 {
				return noResults()
			}
			if total {
				return totalOnly(len(orphans))
			}
			var rows []output.Row
			for _, f := range orphans {
				rows = append(rows, output.NewRow("path", f))
			}
			return output.Print(os.Stdout, cfg.Format, rows)
		},
	}

	cmd.Flags().StringVar(&ignore, "ignore", "", "exclude files matching glob (e.g. \"*/index.md\")")
	cmd.Flags().BoolVar(&total, "total", false, "print count only")

	rootCmd.AddCommand(cmd)
}
