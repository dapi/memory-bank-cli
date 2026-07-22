//go:build windows

package githubadapter

import (
	"bytes"
	"fmt"
	"os"
)

// Windows keeps the same expected-content precondition. The repository's
// Windows ownership transaction supplies descriptor-relative primitives for
// its managed template payload; this adapter has no cross-package access to
// those internal handles, so this fallback retains the no-follow validation
// performed by safePath before each mutation.
func secureAtomicWrite(root string, mutation mutation) error {
	path, err := safePath(root, mutation.relative)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if mutation.existed {
		if err != nil || !bytes.Equal(data, mutation.original) {
			return fmt.Errorf("destination %q changed during apply", mutation.relative)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("destination %q appeared during apply", mutation.relative)
	}
	return atomicWriteFile(path, []byte(mutation.data))
}

func secureRollback(root string, mutation mutation) error {
	path, err := safePath(root, mutation.relative)
	if err != nil {
		return err
	}
	if !mutation.existed {
		return os.Remove(path)
	}
	return atomicWriteFile(path, mutation.original)
}
