package cmd

import (
	"os"

	"github.com/obsy-cli/obsy/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	var total bool

	cmd := &cobra.Command{
		Use:   "deadends",
		Short: "List files with no outgoing links",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, idx, err := openVault()
			if err != nil {
				return err
			}
			deadends := idx.Deadends()
			if len(deadends) == 0 {
				return noResults()
			}
			if total {
				return totalOnly(len(deadends))
			}
			var rows []output.Row
			for _, f := range deadends {
				rows = append(rows, output.NewRow("path", f))
			}
			return output.Print(os.Stdout, cfg.Format, rows)
		},
	}

	cmd.Flags().BoolVar(&total, "total", false, "print count only")
	rootCmd.AddCommand(cmd)
}
