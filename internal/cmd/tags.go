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
	// tags — list all tags
	{
		var (
			counts     bool
			sortBy     string
			pathFilter string
			total      bool
		)
		cmd := &cobra.Command{
			Use:   "tags",
			Short: "List all tags in the vault",
			RunE: func(cmd *cobra.Command, args []string) error {
				_, idx, err := openVault()
				if err != nil {
					return err
				}
				tagCounts := make(map[string]int)
				for _, entry := range idx.Files {
					if pathFilter != "" && !strings.HasPrefix(entry.Path, pathFilter) {
						continue
					}
					for _, t := range entry.Tags {
						tagCounts[t]++
					}
				}
				if len(tagCounts) == 0 {
					return noResults()
				}
				if total {
					return totalOnly(len(tagCounts))
				}

				type tagEntry struct {
					tag   string
					count int
				}
				entries := make([]tagEntry, 0, len(tagCounts))
				for t, c := range tagCounts {
					entries = append(entries, tagEntry{t, c})
				}
				if sortBy == "count" {
					sort.Slice(entries, func(i, j int) bool {
						if entries[i].count != entries[j].count {
							return entries[i].count > entries[j].count
						}
						return entries[i].tag < entries[j].tag
					})
				} else {
					sort.Slice(entries, func(i, j int) bool { return entries[i].tag < entries[j].tag })
				}

				if cfg.Format == "text" {
					for _, e := range entries {
						if counts {
							fmt.Fprintf(os.Stdout, "%s\t%d\n", e.tag, e.count)
						} else {
							fmt.Fprintln(os.Stdout, e.tag)
						}
					}
					return nil
				}
				var rows []output.Row
				for _, e := range entries {
					if counts {
						rows = append(rows, output.NewRow("tag", e.tag, "count", e.count))
					} else {
						rows = append(rows, output.NewRow("tag", e.tag))
					}
				}
				return output.Print(os.Stdout, cfg.Format, rows)
			},
		}
		cmd.Flags().BoolVar(&counts, "counts", false, "include occurrence counts")
		cmd.Flags().StringVar(&sortBy, "sort", "name", "sort by: name|count")
		cmd.Flags().StringVar(&pathFilter, "path", "", "limit to folder")
		cmd.Flags().BoolVar(&total, "total", false, "print count only")
		rootCmd.AddCommand(cmd)
	}

	// tag <name> — files containing a tag
	{
		var (
			pathFilter string
			total      bool
		)
		cmd := &cobra.Command{
			Use:   "tag <name>",
			Short: "List files containing the given tag",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				_, idx, err := openVault()
				if err != nil {
					return err
				}
				tagName := strings.TrimPrefix(args[0], "#")
				var files []string
				for _, entry := range idx.Files {
					if pathFilter != "" && !strings.HasPrefix(entry.Path, pathFilter) {
						continue
					}
					for _, t := range entry.Tags {
						if t == tagName {
							files = append(files, entry.Path)
							break
						}
					}
				}
				sort.Strings(files)
				if len(files) == 0 {
					return noResults()
				}
				if total {
					return totalOnly(len(files))
				}
				var rows []output.Row
				for _, f := range files {
					rows = append(rows, output.NewRow("path", f))
				}
				return output.Print(os.Stdout, cfg.Format, rows)
			},
		}
		cmd.Flags().StringVar(&pathFilter, "path", "", "limit to folder")
		cmd.Flags().BoolVar(&total, "total", false, "print count only")
		rootCmd.AddCommand(cmd)
	}
}
