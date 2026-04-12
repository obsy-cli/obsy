package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// ErrNoResults is returned by commands that found no results (exit code 1).
var ErrNoResults = errors.New("no results")

// Config holds the global flags shared by all commands.
type Config struct {
	Vault   string
	Format  string
	NoCache bool
	Quiet   bool
}

var cfg Config

var rootCmd = &cobra.Command{
	Use:           "obsy",
	Short:         "Obsidian vault CLI — search, graph, tags, tasks, file ops",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfg.Vault, "vault", "", "vault root (default: cwd, walks up to $HOME looking for .obsidian/)")
	rootCmd.PersistentFlags().StringVar(&cfg.Format, "format", "text", "output format: text|json|tsv|csv")
	rootCmd.PersistentFlags().BoolVar(&cfg.NoCache, "no-cache", false, "skip cache, force fresh scan (does not update cache)")
	rootCmd.PersistentFlags().BoolVar(&cfg.Quiet, "quiet", false, "suppress non-essential output")
}
