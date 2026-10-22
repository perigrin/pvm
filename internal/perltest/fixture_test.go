// ABOUTME: Tests WriteModule's fixture layout and Command's env stripping, once here
// ABOUTME: rather than duplicated as TestHeldToPerlIgnoresTheCallersPerlLib and TestReturnOracleIgnoresTheCallersPerlLib.

package perltest

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteModuleWritesUnderPackagePath: a module name's "::" becomes the
// path separator, mirroring how perl's own @INC resolution finds it.
func TestWriteModuleWritesUnderPackagePath(t *testing.T) {
	lib := t.TempDir()
	WriteModule(t, lib, "My::Nested::Module", "package My::Nested::Module;\n1;\n")
	assert.FileExists(t, filepath.Join(lib, "My", "Nested", "Module.pm"))
}

// TestCommandIgnoresTheCallersPerlLib: Command's perl finds a module as
// an installed perl does, whatever PERL5LIB, PERLLIB, PERL5OPT or
// PERL_LOCAL_LIB_ROOT the calling test runs with -- a module reachable
// only through one of those is not one a real caller's environment would
// resolve either. internal/parse's held_to_perl.pl driver and
// internal/infer's return-type oracle both run perl through this same
// Command, so proving it here once covers both.
func TestCommandIgnoresTheCallersPerlLib(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Fixture", "package My::Fixture;\n1;\n")

	for _, env := range [][2]string{
		{"PERL5LIB", lib}, {"PERLLIB", lib}, {"PERL5OPT", "-I" + lib}, {"PERL_LOCAL_LIB_ROOT", lib},
	} {
		t.Run(env[0], func(t *testing.T) {
			t.Setenv(env[0], env[1])
			out, err := p.Command("-e", "require My::Fixture").CombinedOutput()
			assert.Error(t, err, "Command required My::Fixture through %s: %s", env[0], out)
		})
	}
}

// TestCommandWithoutAnyOfThoseVariablesStillFindsAModuleThere: a sanity
// check that the fixture itself is reachable at all -- without it, a
// pass above would mean nothing.
func TestCommandWithoutAnyOfThoseVariablesStillFindsAModuleThere(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Fixture", "package My::Fixture;\n1;\n")

	out, err := p.Command("-I"+lib, "-e", "require My::Fixture; print 1").Output()
	require.NoError(t, err)
	assert.Equal(t, "1", string(out))
}
