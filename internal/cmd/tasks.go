package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/obsy-cli/obsy/internal/output"
	"github.com/obsy-cli/obsy/internal/parser"
	"github.com/spf13/cobra"
)

func init() {
	var (
		todo       bool
		done       bool
		file       string
		pathFilter string
		verbose    bool
		total      bool
	)

	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "List tasks across the vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, idx, err := openVault()
			if err != nil {
				return err
			}

			type taskItem struct {
				file string
				task parser.Task
			}
			var items []taskItem

			files := idx.AllFiles()
			for _, f := range files {
				if pathFilter != "" && !strings.HasPrefix(f, pathFilter) {
					continue
				}
				if file != "" {
					resolved, ok := idx.ResolveFileArg(file)
					if !ok || resolved != f {
						continue
					}
				}
				entry := idx.Files[f]
				for _, t := range entry.Tasks {
					if todo && t.Done {
						continue
					}
					if done && !t.Done {
						continue
					}
					items = append(items, taskItem{f, t})
				}
			}

			sort.Slice(items, func(i, j int) bool {
				if items[i].file != items[j].file {
					return items[i].file < items[j].file
				}
				return items[i].task.Line < items[j].task.Line
			})

			if len(items) == 0 {
				return noResults()
			}
			if total {
				return totalOnly(len(items))
			}

			if cfg.Format == "text" {
				if verbose {
					curFile := ""
					for _, item := range items {
						if item.file != curFile {
							curFile = item.file
							fmt.Fprintf(os.Stdout, "\n%s\n", curFile)
						}
						status := "[ ]"
						if item.task.Done {
							status = "[x]"
						}
						fmt.Fprintf(os.Stdout, "  %s L%-4d %s\n", status, item.task.Line, item.task.Text)
					}
				} else {
					for _, item := range items {
						status := "[ ]"
						if item.task.Done {
							status = "[x]"
						}
						fmt.Fprintf(os.Stdout, "%s %s:%d: %s\n", status, item.file, item.task.Line, item.task.Text)
					}
				}
				return nil
			}

			var rows []output.Row
			for _, item := range items {
				rows = append(rows, output.NewRow(
					"file", item.file,
					"line", item.task.Line,
					"done", item.task.Done,
					"text", item.task.Text,
				))
			}
			return output.Print(os.Stdout, cfg.Format, rows)
		},
	}

	cmd.Flags().BoolVar(&todo, "todo", false, "incomplete tasks only")
	cmd.Flags().BoolVar(&done, "done", false, "completed tasks only")
	cmd.Flags().StringVar(&file, "file", "", "tasks from a specific file")
	cmd.Flags().StringVar(&pathFilter, "path", "", "tasks from a folder")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "group by file with line numbers")
	cmd.Flags().BoolVar(&total, "total", false, "print count only")
	rootCmd.AddCommand(cmd)
}
