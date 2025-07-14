//go:build windows

package updater

import (
	"os"
)

// chownFileUnix is a no-op on Windows since Windows doesn't use Unix-style ownership
func (r *BinaryReplacer) chownFileUnix(dst string, srcInfo os.FileInfo) error {
	// Windows doesn't use Unix-style ownership
	return nil
}
