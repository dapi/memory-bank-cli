//go:build windows

package safepath

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const fileCreated = 2

type fileNameInformation struct {
	ReplaceIfExists uint32
	RootDirectory   windows.Handle
	FileNameLength  uint32
	FileName        [1]uint16
}

func validateRelative(relative string) ([]string, error) {
	relative = filepath.FromSlash(relative)
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

func openRoot(root string) (windows.Handle, error) {
	name, err := windows.UTF16PtrFromString(root)
	if err != nil {
		return 0, err
	}
	handle, err := windows.CreateFile(name, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_FLAG_BACKUP_SEMANTICS|windows.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	if err != nil {
		return 0, err
	}
	file := os.NewFile(uintptr(handle), root)
	if file == nil {
		_ = windows.CloseHandle(handle)
		return 0, errors.New("open safe root")
	}
	info, statErr := file.Stat()
	if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		_ = file.Close()
		if statErr != nil {
			return 0, statErr
		}
		return 0, fmt.Errorf("safe root is not a real directory: %s", root)
	}
	runtime.SetFinalizer(file, nil)
	return handle, nil
}

func ntOpenRelative(parent windows.Handle, name string, access, disposition, options uint32) (windows.Handle, uintptr, error) {
	objectName, err := windows.NewNTUnicodeString(name)
	if err != nil {
		return 0, 0, err
	}
	attributes := windows.OBJECT_ATTRIBUTES{
		Length:        uint32(unsafe.Sizeof(windows.OBJECT_ATTRIBUTES{})),
		RootDirectory: parent,
		ObjectName:    objectName,
		Attributes:    windows.OBJ_CASE_INSENSITIVE | windows.OBJ_DONT_REPARSE,
	}
	var status windows.IO_STATUS_BLOCK
	var handle windows.Handle
	err = windows.NtCreateFile(&handle, access, &attributes, &status, nil, 0, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, disposition, options|windows.FILE_OPEN_REPARSE_POINT, 0, 0)
	return handle, status.Information, err
}

func openParent(root, relative string, create bool) (windows.Handle, string, error) {
	parts, err := validateRelative(relative)
	if err != nil {
		return 0, "", err
	}
	handle, err := openRoot(root)
	if err != nil {
		return 0, "", err
	}
	for _, part := range parts[:len(parts)-1] {
		disposition := uint32(windows.FILE_OPEN)
		if create {
			disposition = windows.FILE_OPEN_IF
		}
		next, _, openErr := ntOpenRelative(handle, part, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE, disposition, windows.FILE_DIRECTORY_FILE)
		_ = windows.CloseHandle(handle)
		if openErr != nil {
			return 0, "", fmt.Errorf("open safe parent: %w", openErr)
		}
		handle = next
	}
	return handle, parts[len(parts)-1], nil
}

func readRegular(root, relative string) ([]byte, fs.FileMode, bool, error) {
	parent, leaf, err := openParent(root, relative, false)
	if err != nil {
		if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, windows.ERROR_PATH_NOT_FOUND) {
			return nil, 0, false, nil
		}
		return nil, 0, false, err
	}
	defer windows.CloseHandle(parent)
	return readRegularAt(parent, leaf, relative)
}

func readRegularAt(parent windows.Handle, leaf, relative string) ([]byte, fs.FileMode, bool, error) {
	handle, _, err := ntOpenRelative(parent, leaf, windows.FILE_GENERIC_READ, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE)
	if errors.Is(err, windows.ERROR_FILE_NOT_FOUND) || errors.Is(err, windows.ERROR_PATH_NOT_FOUND) {
		return nil, 0, false, nil
	}
	if err != nil {
		return nil, 0, false, fmt.Errorf("open safe file %q: %w", relative, err)
	}
	file := os.NewFile(uintptr(handle), relative)
	if file == nil {
		_ = windows.CloseHandle(handle)
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
	defer windows.CloseHandle(parent)
	name, err := createTemporary(parent, data, mode)
	if err != nil {
		return err
	}
	defer removeRelative(parent, name, false)
	handle, _, err := ntOpenRelative(parent, name, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE|windows.DELETE, windows.FILE_OPEN, windows.FILE_NON_DIRECTORY_FILE)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
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
	replace := expected == nil || expected.Exists
	if err := setName(handle, parent, leaf, windows.FileRenameInformation, replace); err != nil {
		if expected != nil && !expected.Exists {
			return fmt.Errorf("destination %q appeared during apply: %w", relative, err)
		}
		return fmt.Errorf("replace safe file %q: %w", relative, err)
	}
	return nil
}

func createTemporary(parent windows.Handle, data []byte, mode fs.FileMode) (string, error) {
	for attempt := 0; attempt < 100; attempt++ {
		name := ".memory-bank-cli-" + strconv.Itoa(os.Getpid()) + "-" + strconv.Itoa(attempt)
		handle, _, err := ntOpenRelative(parent, name, windows.FILE_GENERIC_READ|windows.FILE_GENERIC_WRITE, windows.FILE_CREATE, windows.FILE_NON_DIRECTORY_FILE)
		if errors.Is(err, windows.ERROR_FILE_EXISTS) {
			continue
		}
		if err != nil {
			return "", err
		}
		file := os.NewFile(uintptr(handle), name)
		if file == nil {
			_ = windows.CloseHandle(handle)
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
		if writeErr != nil || closeErr != nil {
			_ = removeRelative(parent, name, false)
			if writeErr != nil {
				return "", writeErr
			}
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
	defer windows.CloseHandle(parent)
	_, _, exists, err := readRegularAt(parent, leaf, relative)
	if err != nil {
		return err
	}
	if !exists {
		return fs.ErrNotExist
	}
	return removeRelative(parent, leaf, false)
}

func removeRelative(parent windows.Handle, leaf string, directory bool) error {
	options := uint32(windows.FILE_NON_DIRECTORY_FILE)
	if directory {
		options = windows.FILE_DIRECTORY_FILE
	}
	handle, _, err := ntOpenRelative(parent, leaf, windows.DELETE, windows.FILE_OPEN, options)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)
	flags := uint32(windows.FILE_DISPOSITION_DELETE | windows.FILE_DISPOSITION_POSIX_SEMANTICS | windows.FILE_DISPOSITION_IGNORE_READONLY_ATTRIBUTE)
	var status windows.IO_STATUS_BLOCK
	return windows.NtSetInformationFile(handle, &status, (*byte)(unsafe.Pointer(&flags)), uint32(unsafe.Sizeof(flags)), windows.FileDispositionInformationEx)
}

func nameInformation(root windows.Handle, leaf string, replace bool) ([]byte, error) {
	name, err := windows.UTF16FromString(leaf)
	if err != nil {
		return nil, err
	}
	name = name[:len(name)-1]
	var header fileNameInformation
	buffer := make([]byte, int(unsafe.Offsetof(header.FileName))+len(name)*2)
	info := (*fileNameInformation)(unsafe.Pointer(&buffer[0]))
	if replace {
		info.ReplaceIfExists = windows.FILE_RENAME_REPLACE_IF_EXISTS | windows.FILE_RENAME_POSIX_SEMANTICS
	}
	info.RootDirectory = root
	info.FileNameLength = uint32(len(name) * 2)
	copy(unsafe.Slice(&info.FileName[0], len(name)), name)
	return buffer, nil
}

func setName(handle, root windows.Handle, leaf string, class uint32, replace bool) error {
	buffer, err := nameInformation(root, leaf, replace)
	if err != nil {
		return err
	}
	var status windows.IO_STATUS_BLOCK
	return windows.NtSetInformationFile(handle, &status, &buffer[0], uint32(len(buffer)), class)
}
