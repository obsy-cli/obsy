package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/obsy-cli/obsy/internal/parser"
	"github.com/spf13/cobra"
)

func init() {
	cmd := &cobra.Command{
		Use:   "read <file>",
		Short: "Read file contents (frontmatter stripped)",
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
			abs := filepath.Join(idx.VaultRoot, file)
			content, err := os.ReadFile(abs)
			if err != nil {
				return err
			}
			_, body, _ := parser.ParseFrontmatter(content)
			fmt.Fprint(os.Stdout, string(body))
			return nil
		},
	}
	rootCmd.AddCommand(cmd)
}
