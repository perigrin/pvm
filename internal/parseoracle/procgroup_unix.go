// ABOUTME: Makes a cancelled subprocess take its children with it, on platforms with process groups.
// ABOUTME: CommandContext kills only the direct child; a shell that forked its real work would leak it.

//go:build unix

package parseoracle

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// killGroup puts cmd in its own process group and kills the whole group on
// cancellation. A `#!/bin/sh` subject that does not exec its last command is a
// parent, and killing the parent alone leaves the work running past the
// sweep; the oracle's inner `sh -c perl` is the same shape.
func killGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
}
