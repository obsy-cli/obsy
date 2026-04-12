package cmd

import (
	"fmt"
	"os"

	"github.com/obsy-cli/obsy/internal/index"
	"github.com/obsy-cli/obsy/internal/vault"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Index health report",
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := vault.Discover(cfg.Vault)
			if err != nil {
				return err
			}
			idx, err := index.LoadOrBuild(v, cfg.NoCache)
			if err != nil {
				return err
			}

			// Count links, tags, properties, tasks.
			var links, tags, props, tasks int
			tagSeen := make(map[string]bool)
			propSeen := make(map[string]bool)
			for _, entry := range idx.Files {
				links += len(entry.Links)
				tasks += len(entry.Tasks)
				for _, t := range entry.Tags {
					tagSeen[t] = true
				}
				for k := range entry.Props {
					propSeen[k] = true
				}
			}
			tags = len(tagSeen)
			props = len(propSeen)

			cachePath := index.CachePath(v.Root)
			var cacheSize int64
			if fi, err := os.Stat(cachePath); err == nil {
				cacheSize = fi.Size()
			}

			fmt.Fprintf(os.Stdout, "vault:      %s\n", v.Root)
			fmt.Fprintf(os.Stdout, "files:      %d\n", len(idx.Files))
			fmt.Fprintf(os.Stdout, "links:      %d\n", links)
			fmt.Fprintf(os.Stdout, "tags:       %d unique\n", tags)
			fmt.Fprintf(os.Stdout, "properties: %d unique\n", props)
			fmt.Fprintf(os.Stdout, "tasks:      %d\n", tasks)
			fmt.Fprintf(os.Stdout, "cache:      %s (%d bytes)\n", cachePath, cacheSize)
			fmt.Fprintf(os.Stdout, "scanned:    %s\n", idx.ScannedAt.Format("2006-01-02 15:04:05"))
			fmt.Fprintf(os.Stdout, "updated:    %s\n", idx.UpdatedAt.Format("2006-01-02 15:04:05"))
			return nil
		},
	}
	rootCmd.AddCommand(cmd)
}
