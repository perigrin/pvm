//go:build linux

// ABOUTME: Tests that psc check opens the file it checks once, counting its opens with inotify.
// ABOUTME: Linux only: inotify is how a test sees each open of a path without standing between psc and the file.

package psc_test

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/psc"
)

// TestCheckReadsFileOnce: psc check opens the file it checks once, and
// takes each diagnostic's line and column from the bytes it parsed. Two
// reads could see two versions of a file being saved, and place one's
// spans in the other's text. perl -w warns about the mismatch at line 2,
// "Use of uninitialized value $u in addition (+) at t.pl line 2", and the
// control is that psc check reports it there.
func TestCheckReadsFileOnce(t *testing.T) {
	isolatePerl(t)
	file := filepath.Join(t.TempDir(), "t.pl")
	require.NoError(t, os.WriteFile(file, []byte("my $u = undef;\nmy $y = $u + 1;\n"), 0o644))

	fd, err := syscall.InotifyInit1(syscall.IN_NONBLOCK | syscall.IN_CLOEXEC)
	require.NoError(t, err)
	defer syscall.Close(fd)
	// Closes are watched as well as opens so that two opens are never
	// adjacent in the queue, where inotify would merge them into one.
	_, err = syscall.InotifyAddWatch(fd, file, syscall.IN_OPEN|syscall.IN_CLOSE)
	require.NoError(t, err)

	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"check", file})
	var stderr strings.Builder
	cmd.SetOut(&strings.Builder{})
	cmd.SetErr(&stderr)
	// Diagnostics make Execute return an error; the output is the subject.
	_ = cmd.Execute()

	assert.Equal(t, 1, inotifyOpens(t, fd), "psc check opens t.pl once")
	assert.Contains(t, stderr.String(), "t.pl:2:9: error:", "control: the mismatch is reported on line 2")
}

// inotifyOpens counts the open events queued on the non-blocking inotify
// fd.
func inotifyOpens(t *testing.T, fd int) int {
	t.Helper()
	buf := make([]byte, 64*(syscall.SizeofInotifyEvent+syscall.NAME_MAX+1))
	n, err := syscall.Read(fd, buf)
	if err == syscall.EAGAIN {
		return 0
	}
	require.NoError(t, err)
	opens := 0
	for off := 0; off < n; {
		ev := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[off]))
		if ev.Mask&syscall.IN_OPEN != 0 {
			opens++
		}
		off += syscall.SizeofInotifyEvent + int(ev.Len)
	}
	return opens
}
