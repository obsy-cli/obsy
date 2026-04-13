package cmd

import (
	"fmt"
	"os"

	"github.com/obsy-cli/obsy/internal/ops"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "rename <file> <new-name>",
		Short: "Rename a file in place and update all wikilinks across the vault",
		Long:  `Extension is auto-preserved if omitted (e.g. new-name → new-name.md).`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, idx, err := openVault()
			if err != nil {
				return err
			}
			result, err := ops.Rename(v, idx, args[0], args[1])
			if err != nil {
				return err
			}
			if !cfg.Quiet {
				fmt.Fprintf(os.Stderr, "renamed: %s → %s\n", result.OldPath, result.NewPath)
				if len(result.Updated) > 0 {
					fmt.Fprintf(os.Stderr, "updated %d file(s)\n", len(result.Updated))
				}
				if len(result.NotUpdated) > 0 {
					fmt.Fprintf(os.Stderr, "WARNING: could not update %d file(s):\n", len(result.NotUpdated))
					for _, f := range result.NotUpdated {
						fmt.Fprintf(os.Stderr, "  %s\n", f)
					}
				}
			}
			saveIndex(idx)
			return nil
		},
	}
	rootCmd.AddCommand(cmd)
}
