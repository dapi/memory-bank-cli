//go:build darwin || linux

package safepath

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

func validateRelative(relative string) ([]string, error) {
	if relative == "" || filepath.IsAbs(relative) || filepath.Clean(relative) != relative {
		return nil, fmt.Errorf("unsafe path %q", relative)
	}
	parts := strings.Split(relative, string(filepath.Separator))
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return nil, fmt.Errorf("unsafe path %q", relative)
		}
	}
	return parts, nil
}

func openParent(root, relative string, create bool) (int, string, error) {
	parts, err := validateRelative(relative)
	if err != nil {
		return -1, "", err
	}
	fd, err := unix.Open(root, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return -1, "", fmt.Errorf("open safe root: %w", err)
	}
	for _, part := range parts[:len(parts)-1] {
		next, openErr := unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		if errors.Is(openErr, unix.ENOENT) && create {
			if mkdirErr := unix.Mkdirat(fd, part, 0o755); mkdirErr != nil && !errors.Is(mkdirErr, unix.EEXIST) {
				_ = unix.Close(fd)
				return -1, "", fmt.Errorf("create safe parent: %w", mkdirErr)
			}
			next, openErr = unix.Openat(fd, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
		}
		if openErr != nil {
			_ = unix.Close(fd)
			return -1, "", fmt.Errorf("open safe parent: %w", openErr)
		}
		_ = unix.Close(fd)
		fd = next
	}
	return fd, parts[len(parts)-1], nil
}

func readRegular(root, relative string) ([]byte, fs.FileMode, bool, error) {
	parent, leaf, err := openParent(root, relative, false)
	if err != nil {
		if errors.Is(err, unix.ENOENT) {
			return nil, 0, false, nil
		}
		return nil, 0, false, err
	}
	defer unix.Close(parent)
	return readRegularAt(parent, leaf, relative)
}

func readRegularAt(parent int, leaf, relative string) ([]byte, fs.FileMode, bool, error) {
	fd, err := unix.Openat(parent, leaf, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_NONBLOCK, 0)
	if errors.Is(err, unix.ENOENT) {
		return nil, 0, false, nil
	}
	if err != nil {
		return nil, 0, false, fmt.Errorf("open safe file %q: %w", relative, err)
	}
	file := os.NewFile(uintptr(fd), relative)
	if file == nil {
		_ = unix.Close(fd)
		return nil, 0, false, fmt.Errorf("open safe file %q", relative)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, 0, false, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, 0, false, fmt.Errorf("path is not a regular file: %s", relative)
	}
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, 0, false, err
	}
	return data, info.Mode().Perm(), true, nil
}

func replaceRegular(root, relative string, data []byte, mode fs.FileMode, expected *Expected) error {
	parent, leaf, err := openParent(root, relative, true)
	if err != nil {
		return err
	}
	defer unix.Close(parent)
	name, err := createTemporary(parent, data, mode)
	if err != nil {
		return err
	}
	defer unix.Unlinkat(parent, name, 0)
	if expected != nil {
		current, _, exists, readErr := readRegularAt(parent, leaf, relative)
		if readErr != nil || exists != expected.Exists || exists && !bytes.Equal(current, expected.Data) {
			if readErr != nil {
				return readErr
			}
			return fmt.Errorf("destination %q changed during apply", relative)
		}
	}
	staged, _, stagedExists, stagedErr := readRegularAt(parent, name, name)
	if stagedErr != nil {
		return stagedErr
	}
	if !stagedExists || !bytes.Equal(staged, data) {
		return fmt.Errorf("safe staged payload %q changed before replacement", relative)
	}
	if expected != nil && !expected.Exists {
		if err := unix.Linkat(parent, name, parent, leaf, 0); err != nil {
			if errors.Is(err, unix.EEXIST) {
				return fmt.Errorf("destination %q appeared during apply", relative)
			}
			return fmt.Errorf("install safe file %q: %w", relative, err)
		}
		return nil
	}
	if err := unix.Renameat(parent, name, parent, leaf); err != nil {
		return fmt.Errorf("replace safe file %q: %w", relative, err)
	}
	return nil
}

func createTemporary(parent int, data []byte, mode fs.FileMode) (string, error) {
	for attempt := 0; attempt < 100; attempt++ {
		name := ".memory-bank-cli-" + strconv.Itoa(os.Getpid()) + "-" + strconv.Itoa(attempt)
		fd, err := unix.Openat(parent, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_CLOEXEC, uint32(mode.Perm()))
		if errors.Is(err, unix.EEXIST) {
			continue
		}
		if err != nil {
			return "", err
		}
		file := os.NewFile(uintptr(fd), name)
		if file == nil {
			_ = unix.Close(fd)
			return "", errors.New("create safe temporary file")
		}
		_, writeErr := file.Write(data)
		if writeErr == nil {
			writeErr = file.Chmod(mode.Perm())
		}
		if writeErr == nil {
			writeErr = file.Sync()
		}
		closeErr := file.Close()
		if writeErr != nil {
			_ = unix.Unlinkat(parent, name, 0)
			return "", writeErr
		}
		if closeErr != nil {
			_ = unix.Unlinkat(parent, name, 0)
			return "", closeErr
		}
		return name, nil
	}
	return "", errors.New("create unique safe temporary file")
}

func removeRegular(root, relative string) error {
	parent, leaf, err := openParent(root, relative, false)
	if err != nil {
		return err
	}
	defer unix.Close(parent)
	_, _, exists, err := readRegularAt(parent, leaf, relative)
	if err != nil {
		return err
	}
	if !exists {
		return fs.ErrNotExist
	}
	if err := unix.Unlinkat(parent, leaf, 0); err != nil {
		return err
	}
	return nil
}
