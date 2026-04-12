package cmd

import (
	"fmt"
	"os"

	"github.com/obsy-cli/obsy/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	var (
		path    string
		counts  bool
		verbose bool
		total   bool
	)

	cmd := &cobra.Command{
		Use:   "unresolved",
		Short: "List broken links (.md wikilinks + non-.md embeds)",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, idx, err := openVault()
			if err != nil {
				return err
			}
			broken := idx.UnresolvedLinks(path)
			if len(broken) == 0 {
				return noResults()
			}
			if total {
				return totalOnly(len(broken))
			}

			// Aggregate counts per raw target.
			type entry struct {
				target  string
				sources []string
			}
			seen := map[string]*entry{}
			var order []string
			for _, b := range broken {
				if _, ok := seen[b.RawTarget]; !ok {
					seen[b.RawTarget] = &entry{target: b.RawTarget}
					order = append(order, b.RawTarget)
				}
				seen[b.RawTarget].sources = append(seen[b.RawTarget].sources, b.SourceFile)
			}

			if cfg.Format == "text" {
				for _, t := range order {
					e := seen[t]
					if verbose {
						for _, src := range e.sources {
							fmt.Fprintf(os.Stdout, "%s\t← %s\n", t, src)
						}
					} else if counts {
						fmt.Fprintf(os.Stdout, "%s\t%d\n", t, len(e.sources))
					} else {
						fmt.Fprintln(os.Stdout, t)
					}
				}
				return nil
			}

			var rows []output.Row
			for _, t := range order {
				e := seen[t]
				if verbose {
					for _, src := range e.sources {
						rows = append(rows, output.NewRow("link", t, "source", src))
					}
				} else if counts {
					rows = append(rows, output.NewRow("link", t, "count", len(e.sources)))
				} else {
					rows = append(rows, output.NewRow("link", t))
				}
			}
			return output.Print(os.Stdout, cfg.Format, rows)
		},
	}

	cmd.Flags().StringVar(&path, "path", "", "limit to folder")
	cmd.Flags().BoolVar(&counts, "counts", false, "include occurrence counts")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "include source file for each broken link")
	cmd.Flags().BoolVar(&total, "total", false, "print count only")

	rootCmd.AddCommand(cmd)
}
