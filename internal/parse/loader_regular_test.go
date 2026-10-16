//go:build !windows

// ABOUTME: Tests that the module loader reads only regular files of bounded size, so a FIFO or /dev/zero is not found.
// ABOUTME: A cloned repository's module path must neither hang psc nor exhaust its memory.

package parse_test

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parse"
)

// TestModuleLoaderReadsRegularFilesOnly checks that a module or required
// file that is a FIFO, a symlink to /dev/zero, or larger than
// parse.MaxSourceBytes is not found, promptly, while a regular file
// beside them is.
func TestModuleLoaderReadsRegularFilesOnly(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, syscall.Mkfifo(filepath.Join(dir, "fifo"), 0o644))
	require.NoError(t, os.Symlink("fifo", filepath.Join(dir, "Hang.pm")))
	require.NoError(t, os.Symlink("fifo", filepath.Join(dir, "hang.pl")))
	require.NoError(t, os.Symlink("/dev/zero", filepath.Join(dir, "Zero.pm")))
	big, err := os.Create(filepath.Join(dir, "Big.pm"))
	require.NoError(t, err)
	require.NoError(t, big.Truncate(parse.MaxSourceBytes+1))
	require.NoError(t, big.Close())
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Ok.pm"), []byte("package Ok;\n1;\n"), 0o644))

	load := parse.DirLoader(dir)
	for _, name := range []string{"Hang", "./hang.pl", "Zero", "Big"} {
		t.Run(name, func(t *testing.T) {
			done := make(chan bool, 1)
			go func() {
				_, found := load(name)
				done <- found
			}()
			select {
			case found := <-done:
				assert.False(t, found)
			case <-time.After(5 * time.Second):
				t.Fatal("the loader hung")
			}
		})
	}

	src, found := load("Ok")
	assert.True(t, found, "a regular module is found")
	assert.Equal(t, "package Ok;\n1;\n", string(src))
}

// TestParseFileReadsRegularFilesOnly checks that ParseFile refuses a FIFO
// rather than blocking on it.
func TestParseFileReadsRegularFilesOnly(t *testing.T) {
	dir := t.TempDir()
	fifo := filepath.Join(dir, "a.pl")
	require.NoError(t, syscall.Mkfifo(fifo, 0o644))
	done := make(chan error, 1)
	go func() {
		_, err := parse.ParseFile(fifo)
		done <- err
	}()
	select {
	case err := <-done:
		assert.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("ParseFile hung")
	}
}
