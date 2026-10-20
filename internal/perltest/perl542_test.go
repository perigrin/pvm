// ABOUTME: Tests that the perl 5.42.0 locator skips, rather than fails, where perl or its privlib is absent.
// ABOUTME: Also checks that a present perl 5.42.0 reports a privlib that exists.

package perltest

import (
	"errors"
	"os"
	"os/exec"
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
	withNoPvmPerl(t)
	t.Setenv("PATH", t.TempDir()) // also none on PATH, as on a runner with neither
	assert.True(t, skipped(t, func(st testing.TB) { Perl542(st) }),
		"perl 5.42.0 missing from pvm and from PATH skips the test")
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

// perl542StandIn writes a shell script at dir/perl that answers Perl542's
// two questions: its version (any argument not -MConfig) and, for
// -MConfig, privlib and archlib under lib.
func perl542StandIn(t *testing.T, dir, version, lib string) string {
	t.Helper()
	path := filepath.Join(dir, "perl")
	script := "#!/bin/sh\ncase \"$1\" in\n" +
		"-MConfig) printf '%s\\n%s\\n' '" + lib + "' '" + lib + "/arch' ;;\n" +
		"*) printf '%s\\n' '" + version + "' ;;\n" +
		"esac\n"
	require.NoError(t, os.WriteFile(path, []byte(script), 0o755))
	return path
}

// withNoPvmPerl replaces perl.GetVersionInfo so pvm itself reports no
// installed perl 5.42.0, restoring it on cleanup.
func withNoPvmPerl(t *testing.T) {
	t.Helper()
	saved := perl.GetVersionInfo
	t.Cleanup(func() { perl.GetVersionInfo = saved })
	perl.GetVersionInfo = func(string) (*perl.VersionInfo, error) {
		return nil, errors.New("version 5.42.0 is not installed")
	}
}

// TestPerl542FindsPerlOnPath: where pvm has no perl 5.42.0 of its own --
// as on a CI runner that installed one with actions-setup-perl rather than
// through pvm -- Perl542 finds one on PATH that reports $] as 5.042000
// and reads its privlib from it, same as a pvm-installed one.
func TestPerl542FindsPerlOnPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in perl is a shell script")
	}
	withNoPvmPerl(t)
	dir := t.TempDir()
	lib := filepath.Join(t.TempDir(), "privlib")
	require.NoError(t, os.MkdirAll(lib, 0o755))
	path := perl542StandIn(t, dir, "5.042000", lib)
	t.Setenv("PATH", dir)

	p := Perl542(t)
	assert.Equal(t, path, p.Path)
	assert.Equal(t, lib, p.Privlib)
	assert.Equal(t, lib+"/arch", p.Archlib)
}

// TestPerl542SkipsAnotherVersion: a perl on PATH that reports a $] other
// than 5.042000 is not perl 5.42.0 -- Perl542 skips rather than using it,
// the same as when PATH has no perl at all.
func TestPerl542SkipsAnotherVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the stand-in perl is a shell script")
	}
	withNoPvmPerl(t)
	dir := t.TempDir()
	perl542StandIn(t, dir, "5.040002", t.TempDir())
	t.Setenv("PATH", dir)

	assert.True(t, skipped(t, func(st testing.TB) { Perl542(st) }),
		"a PATH perl reporting a version other than 5.042000 skips the test")
}

// perl542RequiredFailureChildEnv tells
// TestPerl542RequiredFailureChild to run for real, simulating no perl
// 5.42.0 anywhere, rather than skip itself -- set only by
// TestPerl542FailsWhenRequired's subprocess, following the pattern
// internal/parse/isolation_test.go uses to test code that must fail the
// test rather than pass it.
const perl542RequiredFailureChildEnv = "PVM_PERLTEST_REQUIRED_FAILURE_CHILD"

// TestPerl542RequiredFailureChild calls Perl542 with PVM_REQUIRE_PERL542
// set and no perl 5.42.0 anywhere it looks, so Perl542 must fail it
// rather than skip. Run directly (not as TestPerl542FailsWhenRequired's
// subprocess) it skips itself, so a normal `go test` of this package does
// not fail from the very behavior being verified.
func TestPerl542RequiredFailureChild(t *testing.T) {
	if os.Getenv(perl542RequiredFailureChildEnv) == "" {
		t.Skip("only runs as TestPerl542FailsWhenRequired's subprocess")
	}
	withNoPvmPerl(t)
	t.Setenv("PATH", t.TempDir())
	t.Setenv("PVM_REQUIRE_PERL542", "1")
	Perl542(t)
}

// TestPerl542FailsWhenRequired: PVM_REQUIRE_PERL542=1 turns Perl542's
// skip into a failure, for a CI job (the Linux ratchets job) that a
// quiet skip would turn into a gate that cannot fail.
func TestPerl542FailsWhenRequired(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=^TestPerl542RequiredFailureChild$", "-test.v")
	cmd.Env = append(os.Environ(), perl542RequiredFailureChildEnv+"=1")
	out, err := cmd.CombinedOutput()
	assert.Error(t, err, "Perl542 under PVM_REQUIRE_PERL542 should fail the test, not pass:\n%s", out)
	assert.Contains(t, string(out), "PVM_REQUIRE_PERL542")
}
