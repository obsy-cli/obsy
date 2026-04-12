package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/obsy-cli/obsy/internal/output"
	"github.com/spf13/cobra"
)

func init() {
	// properties — list all property names
	{
		var (
			counts bool
			total  bool
		)
		cmd := &cobra.Command{
			Use:   "properties",
			Short: "List all frontmatter property names",
			RunE: func(cmd *cobra.Command, args []string) error {
				_, idx, err := openVault()
				if err != nil {
					return err
				}
				propCounts := make(map[string]int)
				for _, entry := range idx.Files {
					for k := range entry.Props {
						propCounts[k]++
					}
				}
				if len(propCounts) == 0 {
					return noResults()
				}
				if total {
					return totalOnly(len(propCounts))
				}

				names := make([]string, 0, len(propCounts))
				for k := range propCounts {
					names = append(names, k)
				}
				sort.Strings(names)

				if cfg.Format == "text" {
					for _, n := range names {
						if counts {
							fmt.Fprintf(os.Stdout, "%s\t%d\n", n, propCounts[n])
						} else {
							fmt.Fprintln(os.Stdout, n)
						}
					}
					return nil
				}
				var rows []output.Row
				for _, n := range names {
					if counts {
						rows = append(rows, output.NewRow("property", n, "count", propCounts[n]))
					} else {
						rows = append(rows, output.NewRow("property", n))
					}
				}
				return output.Print(os.Stdout, cfg.Format, rows)
			},
		}
		cmd.Flags().BoolVar(&counts, "counts", false, "include occurrence counts")
		cmd.Flags().BoolVar(&total, "total", false, "print count only")
		rootCmd.AddCommand(cmd)
	}

	// property <name> — list values
	{
		var (
			pathFilter string
			file       string
		)
		cmd := &cobra.Command{
			Use:   "property <name>",
			Short: "List values for a frontmatter property",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				_, idx, err := openVault()
				if err != nil {
					return err
				}
				propName := args[0]

				// Single file mode.
				if file != "" {
					resolved, ok := idx.ResolveFileArg(file)
					if !ok {
						return fmt.Errorf("file not found: %s", file)
					}
					entry := idx.Files[resolved]
					val, ok := entry.Props[propName]
					if !ok {
						return noResults()
					}
					fmt.Fprintln(os.Stdout, fmt.Sprint(val))
					return nil
				}

				type propEntry struct {
					file  string
					value string
				}
				var entries []propEntry
				for _, entry := range idx.Files {
					if pathFilter != "" && !strings.HasPrefix(entry.Path, pathFilter) {
						continue
					}
					val, ok := entry.Props[propName]
					if !ok {
						continue
					}
					entries = append(entries, propEntry{entry.Path, fmt.Sprint(val)})
				}
				sort.Slice(entries, func(i, j int) bool { return entries[i].file < entries[j].file })
				if len(entries) == 0 {
					return noResults()
				}

				if cfg.Format == "text" {
					for _, e := range entries {
						fmt.Fprintf(os.Stdout, "%s\t%s\n", e.file, e.value)
					}
					return nil
				}
				var rows []output.Row
				for _, e := range entries {
					rows = append(rows, output.NewRow("file", e.file, "value", e.value))
				}
				return output.Print(os.Stdout, cfg.Format, rows)
			},
		}
		cmd.Flags().StringVar(&pathFilter, "path", "", "limit to folder")
		cmd.Flags().StringVar(&file, "file", "", "read property from a specific file")
		rootCmd.AddCommand(cmd)
	}
}
