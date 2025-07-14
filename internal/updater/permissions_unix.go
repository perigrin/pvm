//go:build !windows

package updater

import (
	"os"
	"syscall"
)

// chownFileUnix copies ownership from srcInfo to the destination file (Unix-like systems only)
func (r *BinaryReplacer) chownFileUnix(dst string, srcInfo os.FileInfo) error {
	// Extract ownership information from file info
	stat, ok := srcInfo.Sys().(*syscall.Stat_t)
	if !ok {
		// If we can't get syscall.Stat_t, skip ownership copy
		return nil
	}

	// Attempt to change ownership
	err := os.Chown(dst, int(stat.Uid), int(stat.Gid))
	if err != nil {
		// Don't fail if we can't change ownership - might not have permissions
		// This is common when running as a non-root user
		return nil
	}

	return nil
}
