//go:build !windows

package config

import "os"

func isOtherWritable(mode os.FileMode) bool {
	return mode.Perm()&0o002 != 0
}
