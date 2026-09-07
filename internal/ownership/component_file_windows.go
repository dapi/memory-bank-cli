//go:build windows

package ownership

import (
	"errors"
	"os"
)

func checkOriginalComponentFile(info os.FileInfo) error {
	return errors.New("component mutations require Linux or macOS")
}
func syncComponentFile(repo pinnedRepo, p string) error {
	return errors.New("component mutations require Linux or macOS")
}
func syncComponentDirectory(repo pinnedRepo, p string) error {
	return errors.New("component mutations require Linux or macOS")
}

func chmodComponentDirectory(repo pinnedRepo, p string, mode os.FileMode) error {
	return errors.New("component mutations require Linux or macOS")
}

func componentLinkCount(os.FileInfo) uint64 { return 0 }
