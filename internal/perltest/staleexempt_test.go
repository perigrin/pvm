// ABOUTME: Tests that StaleExemptions catches a platform exemption whose claim no longer holds, and fails closed.
// ABOUTME: Runs entirely on Linux, against synthetic exemptions and fixture modules, not real platform data.

package perltest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeCases writes module's oracle call-case file, the perl list
// cases, into a fresh directory and returns that directory.
func writeCases(t *testing.T, module, cases string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, strings.ReplaceAll(module, "::", "-")+".pl")
	require.NoError(t, os.WriteFile(path, []byte(cases), 0o644))
	return dir
}

// sub is an Exemption for a plain sub called in scalar context.
func sub(kind ExemptKind, typ string) Exemption {
	return Exemption{Kind: kind, Type: typ, Calls: []string{"scalar:sub"}}
}

// TestStalePlatformExemptionFails: an exemption claiming My::Fresh::works
// dies is false -- works just returns 42 -- so StaleExemptions must
// report it, proving the mechanism a per-OS exemption map relies on
// actually catches a claim perl no longer bears out. This runs on Linux,
// against a synthetic exemption built for this test, not against the
// real darwin/windows maps this repository has no such perl to check.
func TestStalePlatformExemptionFails(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Fresh", "package My::Fresh;\nsub works { 42 }\n1;\n")
	cases := writeCases(t, "My::Fresh", "( works => [ [] ] );\n")

	stale := StaleExemptions(t, p, lib, "", cases, map[string]Exemption{"My::Fresh::works": sub(Dies, "")})

	require.Len(t, stale, 1)
	assert.Contains(t, stale[0], "My::Fresh::works")
	assert.Contains(t, stale[0], "no call case did")
}

// TestFreshPlatformExemptionDoesNotFail: an exemption claiming a sub
// that genuinely dies does -- StaleExemptions must not flag it. One
// dying call case is enough, as one is enough to fail the oracle.
func TestFreshPlatformExemptionDoesNotFail(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Broken", "package My::Broken;\nsub fails { $_[0] and die \"nope\\n\"; 1 }\n1;\n")
	cases := writeCases(t, "My::Broken", "( fails => [ [0], [1] ] );\n")

	stale := StaleExemptions(t, p, lib, "", cases, map[string]Exemption{"My::Broken::fails": sub(Dies, "")})

	assert.Empty(t, stale)
}

// TestStaleCheckPerlErrorFails: an exemption naming a module perl
// cannot even load (require fails outright, before the exemption's own
// claim is ever exercised) must be reported, not silently skipped as
// inconclusive -- a perl error during the check is a failure of the
// check, not evidence the exemption is still good.
func TestStaleCheckPerlErrorFails(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	cases := writeCases(t, "My::DoesNotExist", "( works => [ [] ] );\n")

	stale := StaleExemptions(t, p, lib, "", cases, map[string]Exemption{"My::DoesNotExist::works": sub(Dies, "")})

	require.Len(t, stale, 1)
	assert.Contains(t, stale[0], "My::DoesNotExist::works")
	assert.Contains(t, stale[0], "could not check")
}

// TestStaleCheckMissingCaseFails: an exemption whose module has no
// call case for its name -- or no cases file at all -- cannot be
// checked, so it must be reported, not counted as still dying.
func TestStaleCheckMissingCaseFails(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Fresh", "package My::Fresh;\nsub works { 42 }\nsub other { 1 }\n1;\n")
	WriteModule(t, lib, "My::NoCases", "package My::NoCases;\nsub works { 42 }\n1;\n")
	cases := writeCases(t, "My::Fresh", "( other => [ [] ] );\n")

	for _, key := range []string{"My::Fresh::works", "My::NoCases::works"} {
		stale := StaleExemptions(t, p, lib, "", cases, map[string]Exemption{key: sub(Dies, "")})
		if assert.Len(t, stale, 1, key) {
			assert.Contains(t, stale[0], key)
			assert.Contains(t, stale[0], "could not check")
		}
	}
}

// TestStaleCheckCallsMethods: a method is called the way the
// return-type oracle calls it -- on the instance its cases file's
// INSTANCE names -- so a method that works on that instance is
// reported stale, not counted as dying because it was called as a
// plain function with no invocant.
func TestStaleCheckCallsMethods(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Obj", "package My::Obj;\nsub new { bless {}, shift }\nsub works { ref $_[0] or die \"not an instance\\n\"; 42 }\n1;\n")
	cases := writeCases(t, "My::Obj", "( works => [ [] ], INSTANCE => { ctor => 'new', args => [], subs => ['works'] } );\n")

	stale := StaleExemptions(t, p, lib, "", cases, map[string]Exemption{
		"My::Obj::works": {Kind: Dies, Calls: []string{"scalar:method"}},
	})

	require.Len(t, stale, 1)
	assert.Contains(t, stale[0], "My::Obj::works")
	assert.Contains(t, stale[0], "no call case did")
}

// TestStaleCheckIgnoresWhatASubPrints: a sub, or its module as it
// loads, printing to STDOUT is not a death -- the check reads the
// oracle's own report, which nothing a sub prints can stand in for.
func TestStaleCheckIgnoresWhatASubPrints(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Chatty", "package My::Chatty;\nprint \"loaded\\n\";\nsub works { print \"hello\\n\"; 42 }\n1;\n")
	cases := writeCases(t, "My::Chatty", "( works => [ [] ] );\n")

	stale := StaleExemptions(t, p, lib, "", cases, map[string]Exemption{"My::Chatty::works": sub(Dies, "")})

	require.Len(t, stale, 1)
	assert.Contains(t, stale[0], "My::Chatty::works")
	assert.Contains(t, stale[0], "no call case did")
}

// TestStaleCheckHoldsReturnClaims: an exemption claiming a sub returns
// a type is checked too, whatever its reason says -- reported when no
// call case returns that type, kept when one does.
func TestStaleCheckHoldsReturnClaims(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Ret", "package My::Ret;\nsub int { 42 }\n1;\n")
	cases := writeCases(t, "My::Ret", "( int => [ [] ] );\n")

	stale := StaleExemptions(t, p, lib, "", cases, map[string]Exemption{
		"My::Ret::int": sub(Returns, "Str"),
	})
	require.Len(t, stale, 1)
	assert.Contains(t, stale[0], "My::Ret::int")
	assert.Contains(t, stale[0], "no call case did")

	assert.Empty(t, StaleExemptions(t, p, lib, "", cases, map[string]Exemption{"My::Ret::int": sub(Returns, "Int")}))
}

// TestStaleCheckNeverCallsUnsafe: an Unsafe exemption -- one whose call
// can kill perl outright -- is never called, and an entry with no kind
// is reported rather than guessed at.
func TestStaleCheckNeverCallsUnsafe(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	marker := filepath.Join(t.TempDir(), "called")
	WriteModule(t, lib, "My::Unsafe", "package My::Unsafe;\nsub boom { open my $f, '>', $ENV{STALE_MARKER} or die; 1 }\n1;\n")
	cases := writeCases(t, "My::Unsafe", "( boom => [ [] ] );\n")
	t.Setenv("STALE_MARKER", marker)

	assert.Empty(t, StaleExemptions(t, p, lib, "", cases, map[string]Exemption{"My::Unsafe::boom": {Kind: Unsafe}}))
	assert.NoFileExists(t, marker)

	stale := StaleExemptions(t, p, lib, "", cases, map[string]Exemption{"My::Unsafe::boom": {Calls: []string{"scalar:sub"}}})
	require.Len(t, stale, 1)
	assert.Contains(t, stale[0], "kind")
}
