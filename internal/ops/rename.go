package ops

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/obsy-cli/obsy/internal/index"
	"github.com/obsy-cli/obsy/internal/vault"
)

// Rename renames a file in place, preserving its directory.
// newName may omit the .md extension — it will be added automatically.
func Rename(v *vault.Vault, idx *index.Index, src, newName string) (*MoveResult, error) {
	srcRel, ok := idx.ResolveFileArg(src)
	if !ok {
		return nil, fmt.Errorf("file not found: %s", src)
	}

	// Preserve extension.
	if !strings.HasSuffix(newName, ".md") {
		newName += ".md"
	}

	dstRel := filepath.Join(filepath.Dir(srcRel), newName)
	return Move(v, idx, srcRel, dstRel)
}
