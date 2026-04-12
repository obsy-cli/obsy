package cmd

import (
	"fmt"
	"os"

	"github.com/obsy-cli/obsy/internal/ops"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "move <file> <destination>",
		Short: "Move a file and update all wikilinks across the vault",
		Long: `Destination can be a folder path or a full path (supports rename+move in one step).
All wikilinks referencing the file are updated. Validates before touching any file.`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			v, idx, err := openVault()
			if err != nil {
				return err
			}
			result, err := ops.Move(v, idx, args[0], args[1])
			if err != nil {
				return err
			}
			if !cfg.Quiet {
				fmt.Fprintf(os.Stderr, "moved: %s → %s\n", result.OldPath, result.NewPath)
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
