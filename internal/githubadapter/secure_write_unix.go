//go:build darwin || linux

package githubadapter

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// secureAtomicWrite rechecks the planned file immediately before replacing it
// and resolves every parent through pinned directory descriptors. A parent
// swapped for a symlink after planning therefore cannot redirect a write.
func secureAtomicWrite(root string, mutation mutation) error {
	parent, leaf, err := openAdapterParent(root, mutation.relative)
	if err != nil {
		return err
	}
	defer unix.Close(parent)
	if err := verifyExpected(parent, leaf, mutation); err != nil {
		return err
	}
	temporary, err := createTemporary(parent, []byte(mutation.data))
	if err != nil {
		return err
	}
	defer unix.Unlinkat(parent, temporary, 0)
	// Recheck after staging: this is the last point before the replace.
	if err := verifyExpected(parent, leaf, mutation); err != nil {
		return err
	}
	if !mutation.existed {
		if err := unix.Linkat(parent, temporary, parent, leaf, 0); err != nil {
			if errors.Is(err, unix.EEXIST) {
				return fmt.Errorf("destination %q appeared during apply", mutation.relative)
			}
			return err
		}
		return nil
	}
	return unix.Renameat(parent, temporary, parent, leaf)
}

func openAdapterParent(root, relative string) (int, string, error) {
	if filepath.IsAbs(relative) || strings.Contains(relative, "..") || strings.Contains(relative, "\\") {
		return -1, "", fmt.Errorf("unsafe adapter path %q", relative)
	}
	parts := strings.Split(filepath.FromSlash(relative), string(filepath.Separator))
	fd, err := unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return -1, "", err
	}
	for _, part := range parts[:len(parts)-1] {
		next, err := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if err != nil {
			unix.Close(fd)
			return -1, "", fmt.Errorf("open adapter parent %q: %w", relative, err)
		}
		unix.Close(fd)
		fd = next
	}
	return fd, parts[len(parts)-1], nil
}

func verifyExpected(parent int, leaf string, mutation mutation) error {
	fd, err := unix.Openat(parent, leaf, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if !mutation.existed {
		if errors.Is(err, unix.ENOENT) {
			return nil
		}
		if err == nil {
			unix.Close(fd)
			return fmt.Errorf("destination %q appeared during apply", mutation.relative)
		}
		return err
	}
	if err != nil {
		return fmt.Errorf("destination %q disappeared during apply: %w", mutation.relative, err)
	}
	file := os.NewFile(uintptr(fd), mutation.relative)
	if file == nil {
		unix.Close(fd)
		return fmt.Errorf("open destination %q", mutation.relative)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		if err != nil {
			return err
		}
		return fmt.Errorf("unsafe destination %q", mutation.relative)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	if !bytes.Equal(data, mutation.original) {
		return fmt.Errorf("destination %q changed during apply", mutation.relative)
	}
	return nil
}

func createTemporary(parent int, data []byte) (string, error) {
	for attempt := 0; attempt < 100; attempt++ {
		name := fmt.Sprintf(".mb-cli-github-%d-%d", os.Getpid(), attempt)
		fd, err := unix.Openat(parent, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC, 0o644)
		if errors.Is(err, unix.EEXIST) {
			continue
		}
		if err != nil {
			return "", err
		}
		file := os.NewFile(uintptr(fd), name)
		if file == nil {
			unix.Close(fd)
			return "", fmt.Errorf("create staging file")
		}
		_, writeErr := file.Write(data)
		if writeErr == nil {
			writeErr = file.Sync()
		}
		closeErr := file.Close()
		if writeErr != nil {
			unix.Unlinkat(parent, name, 0)
			return "", writeErr
		}
		if closeErr != nil {
			unix.Unlinkat(parent, name, 0)
			return "", closeErr
		}
		return name, nil
	}
	return "", fmt.Errorf("create unique staging file")
}
