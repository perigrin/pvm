// ABOUTME: The no-process-group fallback, so the package still builds where Setpgid does not exist.
// ABOUTME: CommandContext's default (kill the direct child) is all this platform gets.

//go:build !unix

package parseoracle

import "os/exec"

// killGroup is a no-op here; see procgroup_unix.go for what it does elsewhere.
func killGroup(cmd *exec.Cmd) {}
