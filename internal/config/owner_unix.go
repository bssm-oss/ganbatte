//go:build !windows

package config

import (
	"fmt"
	"os"
	"syscall"
)

func isOwnedByCurrentUser(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("checking owner: %w", err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("checking owner: unsupported file info for %s", path)
	}
	return stat.Uid == uint32(os.Geteuid()), nil
}
