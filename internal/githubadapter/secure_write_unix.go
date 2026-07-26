//go:build darwin || linux

package githubadapter

import (
	"os"

	"github.com/dapi/memory-bank-cli/internal/safepath"
)

func secureAtomicWrite(root string, mutation mutation) error {
	return safepath.ReplaceRegular(root, mutation.relative, []byte(mutation.data), 0o644, &safepath.Expected{
		Exists: mutation.existed,
		Data:   mutation.original,
	})
}

func secureRollback(root string, mutation mutation) error {
	if !mutation.existed {
		return safepath.RemoveRegular(root, mutation.relative)
	}
	return safepath.ReplaceRegular(root, mutation.relative, mutation.original, os.FileMode(0o644), nil)
}
