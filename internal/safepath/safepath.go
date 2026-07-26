// Package safepath provides no-follow regular-file operations rooted at a
// repository directory. It is used by commands that must not let a path swap
// redirect a write outside their declared root.
package safepath

import (
	"bytes"
	"fmt"
	"io/fs"
)

// Expected describes the destination state required before a replacement.
// Callers use it when preserving a user-owned file is part of their contract.
type Expected struct {
	Exists bool
	Data   []byte
}

// ReadRegular returns one regular file below root without following any
// symlink in its path.
func ReadRegular(root, relative string) ([]byte, fs.FileMode, error) {
	data, mode, exists, err := readRegular(root, relative)
	if err != nil {
		return nil, 0, err
	}
	if !exists {
		return nil, 0, fs.ErrNotExist
	}
	return data, mode, nil
}

// ReplaceRegular atomically replaces a regular file below root. It creates
// absent parents under the same no-follow boundary. When expected is non-nil,
// the destination is checked immediately before the replacement.
func ReplaceRegular(root, relative string, data []byte, mode fs.FileMode, expected *Expected) error {
	current, _, exists, err := readRegular(root, relative)
	if err != nil {
		return err
	}
	if expected != nil {
		if exists != expected.Exists {
			return fmt.Errorf("destination %q changed during apply", relative)
		}
		if exists && !bytes.Equal(current, expected.Data) {
			return fmt.Errorf("destination %q changed during apply", relative)
		}
	}
	return replaceRegular(root, relative, data, mode, expected)
}

// RemoveRegular removes one regular file below root without following
// symlinks. A concurrently introduced symlink is never followed.
func RemoveRegular(root, relative string) error {
	return removeRegular(root, relative)
}
