//go:build windows

// ABOUTME: Windows-specific disk space implementation using Windows APIs
// ABOUTME: Uses GetDiskFreeSpaceEx for accurate filesystem statistics
package diskspace

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	getDiskFreeSpaceEx = kernel32.NewProc("GetDiskFreeSpaceExW")
)

// getSpaceInfo returns disk space information using Windows APIs
func getSpaceInfo(path string) (*SpaceInfo, error) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, fmt.Errorf("converting path to UTF16: %w", err)
	}

	var freeBytesAvailable, totalNumberOfBytes, totalNumberOfFreeBytes uint64

	ret, _, err := getDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		uintptr(unsafe.Pointer(&freeBytesAvailable)),
		uintptr(unsafe.Pointer(&totalNumberOfBytes)),
		uintptr(unsafe.Pointer(&totalNumberOfFreeBytes)),
	)

	if ret == 0 {
		return nil, fmt.Errorf("GetDiskFreeSpaceEx failed for path %s: %w", path, err)
	}

	return &SpaceInfo{
		Total:     int64(totalNumberOfBytes),
		Free:      int64(totalNumberOfFreeBytes),
		Available: int64(freeBytesAvailable),
	}, nil
}
