// ABOUTME: Tests that the perl 5.42.0 locator skips, rather than fails, where perl or its privlib is absent.
// ABOUTME: Also checks that a present perl 5.42.0 reports a privlib that exists.

package perltest

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipped runs locate on path in a subtest and reports whether it skipped.
func skipped(t *testing.T, path string) bool {
	t.Helper()
	var sub *testing.T
	t.Run("locate", func(st *testing.T) {
		sub = st
		locate(st, path)
	})
	return sub.Skipped()
}

func TestPerl542SkipsWhenAbsent(t *testing.T) {
	assert.True(t, skipped(t, filepath.Join(t.TempDir(), "perl")),
		"a missing perl skips the test")
}

func TestPerl542SkipsWhenPrivlibMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in perl is a shell script")
	}
	missing := filepath.Join(t.TempDir(), "privlib")
	fake := filepath.Join(t.TempDir(), "perl")
	require.NoError(t, os.WriteFile(fake,
		[]byte("#!/bin/sh\nprintf '%s\\n%s\\n' '"+missing+"' '"+missing+"/arch'\n"), 0o755))
	assert.True(t, skipped(t, fake), "a perl whose privlib is missing skips the test")
}

func TestPerl542ReportsAnExistingPrivlib(t *testing.T) {
	p := Perl542(t)
	fi, err := os.Stat(p.Privlib)
	require.NoError(t, err)
	assert.True(t, fi.IsDir())
	assert.NotEmpty(t, p.Archlib)
}
