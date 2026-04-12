package cmd

import (
	"fmt"
	"os"

	"github.com/obsy-cli/obsy/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	var (
		counts bool
		total  bool
	)

	cmd := &cobra.Command{
		Use:   "backlinks <file>",
		Short: "List files that link TO the given file",
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
			sources := idx.BacklinksTo(file)
			if len(sources) == 0 {
				return noResults()
			}
			if total {
				return totalOnly(len(sources))
			}

			if counts {
				linkCounts := idx.BacklinkCounts(file)
				var rows []output.Row
				for _, src := range sources {
					rows = append(rows, output.NewRow("path", src, "count", linkCounts[src]))
				}
				return output.Print(os.Stdout, cfg.Format, rows)
			}

			var rows []output.Row
			for _, src := range sources {
				rows = append(rows, output.NewRow("path", src))
			}
			return output.Print(os.Stdout, cfg.Format, rows)
		},
	}

	cmd.Flags().BoolVar(&counts, "counts", false, "include link count per source")
	cmd.Flags().BoolVar(&total, "total", false, "print count only")
	rootCmd.AddCommand(cmd)
}
