package cmd

import (
	"os"
	"sort"
	"strings"

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

			// FilesWithMtime reuses mtime data already collected during the walk.
			infos, err := v.FilesWithMtime()
			if err != nil {
				return err
			}

			if folder != "" {
				var filtered []vault.FileMeta
				for _, fi := range infos {
					if strings.HasPrefix(fi.Path, folder) {
						filtered = append(filtered, fi)
					}
				}
				infos = filtered
			}

			if len(infos) == 0 {
				return noResults()
			}
			if total {
				return totalOnly(len(infos))
			}

			// Sort.
			if sortBy == "modified" {
				sort.Slice(infos, func(i, j int) bool {
					return infos[i].Mtime.After(infos[j].Mtime)
				})
			} else {
				sort.Slice(infos, func(i, j int) bool {
					return infos[i].Path < infos[j].Path
				})
			}

			files := make([]string, len(infos))
			for i, fi := range infos {
				files[i] = fi.Path
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
