package cmd

import (
	"fmt"
	"os"

	"github.com/obsy-cli/obsy/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	var (
		resolve bool
		total   bool
	)

	cmd := &cobra.Command{
		Use:   "links <file>",
		Short: "List outgoing links FROM the given file",
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
			outgoing := idx.LinksFrom(file)
			if len(outgoing) == 0 {
				return noResults()
			}
			if total {
				return totalOnly(len(outgoing))
			}

			if resolve {
				if cfg.Format == "text" {
					for _, l := range outgoing {
						resolved := l.Resolved
						if resolved == "" {
							resolved = "[unresolved]"
						}
						fmt.Fprintf(os.Stdout, "%s → %s\n", l.Target, resolved)
					}
					return nil
				}
				var rows []output.Row
				for _, l := range outgoing {
					resolved := l.Resolved
					if resolved == "" {
						resolved = "[unresolved]"
					}
					rows = append(rows, output.NewRow("link", l.Target, "display", l.Display, "resolved", resolved))
				}
				return output.Print(os.Stdout, cfg.Format, rows)
			}

			var rows []output.Row
			for _, l := range outgoing {
				rows = append(rows, output.NewRow("link", l.Target, "display", l.Display))
			}
			return output.Print(os.Stdout, cfg.Format, rows)
		},
	}

	cmd.Flags().BoolVar(&resolve, "resolve", false, "show resolved paths (marks unresolved)")
	cmd.Flags().BoolVar(&total, "total", false, "print count only")
	rootCmd.AddCommand(cmd)
}
