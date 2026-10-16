// ABOUTME: Tests that the perl 5.42.0 locator skips, rather than fails, where perl or its privlib is absent.
// ABOUTME: Also checks that a present perl 5.42.0 reports a privlib that exists.

package perltest

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/perl"
)

// skipped runs find in a subtest and reports whether it skipped.
func skipped(t *testing.T, find func(testing.TB)) bool {
	t.Helper()
	var sub *testing.T
	t.Run("find", func(st *testing.T) {
		sub = st
		find(st)
	})
	return sub.Skipped()
}

// at is a find that locates the perl at path.
func at(path string) func(testing.TB) {
	return func(t testing.TB) { locate(t, path) }
}

func TestPerl542SkipsWhenAbsent(t *testing.T) {
	assert.True(t, skipped(t, at(filepath.Join(t.TempDir(), "perl"))),
		"a missing perl skips the test")
}

func TestPerl542SkipsWhenNotInstalledInPvm(t *testing.T) {
	saved := perl.GetVersionInfo
	t.Cleanup(func() { perl.GetVersionInfo = saved })
	perl.GetVersionInfo = func(string) (*perl.VersionInfo, error) {
		return nil, errors.New("version 5.42.0 is not installed")
	}
	assert.True(t, skipped(t, func(st testing.TB) { Perl542(st) }),
		"perl 5.42.0 missing from pvm skips the test")
}

func TestPerl542SkipsWhenPrivlibMissing(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in perl is a shell script")
	}
	missing := filepath.Join(t.TempDir(), "privlib")
	fake := filepath.Join(t.TempDir(), "perl")
	require.NoError(t, os.WriteFile(fake,
		[]byte("#!/bin/sh\nprintf '%s\\n%s\\n' '"+missing+"' '"+missing+"/arch'\n"), 0o755))
	assert.True(t, skipped(t, at(fake)), "a perl whose privlib is missing skips the test")
}

func TestPerl542ReportsAnExistingPrivlib(t *testing.T) {
	p := Perl542(t)
	fi, err := os.Stat(p.Privlib)
	require.NoError(t, err)
	assert.True(t, fi.IsDir())
	assert.NotEmpty(t, p.Archlib)
}
