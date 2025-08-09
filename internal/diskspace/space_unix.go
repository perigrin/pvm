//go:build unix

// ABOUTME: Unix-specific disk space implementation using syscall.Statfs
// ABOUTME: Supports Linux and macOS platforms with proper filesystem statistics
package diskspace

import (
	"fmt"
	"syscall"
)

// getSpaceInfo returns disk space information using Unix syscalls
func getSpaceInfo(path string) (*SpaceInfo, error) {
	var stat syscall.Statfs_t

	err := syscall.Statfs(path, &stat)
	if err != nil {
		return nil, fmt.Errorf("statfs failed for path %s: %w", path, err)
	}

	// Calculate space information
	blockSize := int64(stat.Bsize)
	totalBlocks := int64(stat.Blocks)
	freeBlocks := int64(stat.Bfree)
	availableBlocks := int64(stat.Bavail)

	return &SpaceInfo{
		Total:     totalBlocks * blockSize,
		Free:      freeBlocks * blockSize,
		Available: availableBlocks * blockSize,
	}, nil
}
