//go:build windows

// ABOUTME: Windows-specific file ownership handling for PVM updates.
// ABOUTME: No-op on Windows since Unix-style ownership does not apply.

package updater

import "os"

// chownFile is a no-op on Windows — Unix-style ownership does not apply.
func (r *BinaryReplacer) chownFile(_ string, _ os.FileInfo) error {
	return nil
}
