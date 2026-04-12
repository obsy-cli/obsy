package cmd

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/obsy-cli/obsy/internal/output"
	"github.com/obsy-cli/obsy/internal/vault"
	"github.com/spf13/cobra"
)

func init() {
	var (
		folder string
		sortBy string
		limit  int
		total  bool
	)

	cmd := &cobra.Command{
		Use:   "files",
		Short: "List files in the vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := vault.Discover(cfg.Vault)
			if err != nil {
				return err
			}
			files, err := v.Files()
			if err != nil {
				return err
			}

			if folder != "" {
				var filtered []string
				for _, f := range files {
					if strings.HasPrefix(f, folder) {
						filtered = append(filtered, f)
					}
				}
				files = filtered
			}

			if len(files) == 0 {
				return noResults()
			}
			if total {
				return totalOnly(len(files))
			}

			// Sort.
			if sortBy == "modified" {
				type entry struct {
					path  string
					mtime time.Time
				}
				entries := make([]entry, 0, len(files))
				for _, f := range files {
					abs := filepath.Join(v.Root, f)
					fi, err := os.Stat(abs)
					if err != nil {
						continue
					}
					entries = append(entries, entry{f, fi.ModTime()})
				}
				sort.Slice(entries, func(i, j int) bool {
					return entries[i].mtime.After(entries[j].mtime)
				})
				files = make([]string, len(entries))
				for i, e := range entries {
					files[i] = e.path
				}
			} else {
				sort.Strings(files)
			}

			if limit > 0 && len(files) > limit {
				files = files[:limit]
			}

			var rows []output.Row
			for _, f := range files {
				rows = append(rows, output.NewRow("path", f))
			}
			return output.Print(os.Stdout, cfg.Format, rows)
		},
	}

	cmd.Flags().StringVar(&folder, "folder", "", "filter by folder")
	cmd.Flags().StringVar(&sortBy, "sort", "name", "sort by: name|modified")
	cmd.Flags().IntVar(&limit, "limit", 0, "limit results (0 = unlimited)")
	cmd.Flags().BoolVar(&total, "total", false, "print count only")
	rootCmd.AddCommand(cmd)
}
