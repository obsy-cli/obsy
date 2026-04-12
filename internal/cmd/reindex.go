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
		Use:   "reindex",
		Short: "Force full reindex (rebuild cache from scratch)",
		RunE: func(cmd *cobra.Command, args []string) error {
			v, err := vault.Discover(cfg.Vault)
			if err != nil {
				return err
			}
			if !cfg.Quiet {
				fmt.Fprintf(os.Stderr, "Scanning %s ...\n", v.Root)
			}
			idx, err := index.Full(v)
			if err != nil {
				return err
			}
			if err := index.Save(idx); err != nil {
				return err
			}
			if !cfg.Quiet {
				fmt.Fprintf(os.Stderr, "Indexed %d files → %s\n", len(idx.Files), index.CachePath(v.Root))
			}
			return nil
		},
	}
	rootCmd.AddCommand(cmd)
}
