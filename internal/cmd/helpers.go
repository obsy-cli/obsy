package cmd

import (
	"fmt"
	"os"

	"github.com/obsy-cli/obsy/internal/index"
	"github.com/obsy-cli/obsy/internal/vault"
)

// openVault discovers the vault and loads (or builds) the index.
func openVault() (*vault.Vault, *index.Index, error) {
	v, err := vault.Discover(cfg.Vault)
	if err != nil {
		return nil, nil, err
	}
	idx, err := index.LoadOrBuild(v, cfg.NoCache)
	if err != nil {
		return nil, nil, err
	}
	return v, idx, nil
}

// openVaultOnly discovers the vault without loading the index.
func openVaultOnly() (*vault.Vault, error) {
	return vault.Discover(cfg.Vault)
}

// saveIndex saves the index silently (best-effort).
func saveIndex(idx *index.Index) {
	if !cfg.NoCache {
		_ = index.Save(idx)
	}
}

// noResults returns ErrNoResults and prints nothing.
func noResults() error {
	return ErrNoResults
}

// totalOnly prints a count and exits 0 (or 1 if zero).
func totalOnly(n int) error {
	fmt.Fprintln(os.Stdout, n)
	if n == 0 {
		return ErrNoResults
	}
	return nil
}
