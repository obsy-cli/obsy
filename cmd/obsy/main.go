package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/obsy-cli/obsy/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if errors.Is(err, cmd.ErrNoResults) {
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
