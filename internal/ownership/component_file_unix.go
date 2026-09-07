//go:build darwin || linux

package ownership

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"syscall"
)

func checkOriginalComponentFile(info os.FileInfo) error {
	if !info.Mode().IsRegular() || info.Mode()&(os.ModeSetuid|os.ModeSetgid|os.ModeSticky) != 0 {
		return errors.New("component input must be a regular file without special mode bits")
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok || st.Nlink != 1 {
		return errors.New("component input has unsupported hard links or identity")
	}
	return nil
}

func syncComponentFile(repo pinnedRepo, p string) error {
	parent, leaf, err := openDestinationParent(repo, p, false, nil)
	if err != nil {
		return err
	}
	defer syscall.Close(parent)
	fd, err := unix.Openat(parent, leaf, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)
	return syscall.Fsync(fd)
}
func syncComponentDirectory(repo pinnedRepo, p string) error {
	if p == "." {
		return syncLocalDirectory(repo.root)
	}
	fd, _, err := openDestinationParent(repo, p+"/sync-placeholder", false, nil)
	if err != nil {
		return err
	}
	defer syscall.Close(fd)
	return syscall.Fsync(fd)
}

func chmodComponentDirectory(repo pinnedRepo, p string, mode os.FileMode) error {
	fd, _, err := openDestinationParent(repo, p+"/chmod-placeholder", false, nil)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	return unix.Fchmod(fd, uint32(mode.Perm()))
}
