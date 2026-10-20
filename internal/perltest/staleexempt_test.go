// ABOUTME: Tests that StaleExemptions catches a platform exemption whose "dies" claim no longer holds.
// ABOUTME: Runs entirely on Linux, against a synthetic exemption and fixture module, not real platform data.

package perltest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	exempt := map[string]string{
		"My::Fresh::works": "dies on this platform's perl -- synthetic, for TestStalePlatformExemptionFails",
	}
	stale := StaleExemptions(t, p, lib, exempt)

	require.Len(t, stale, 1)
	assert.Contains(t, stale[0], "My::Fresh::works")
}

// TestFreshPlatformExemptionDoesNotFail: an exemption claiming a sub
// that genuinely dies does -- StaleExemptions must not flag it.
func TestFreshPlatformExemptionDoesNotFail(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Broken", "package My::Broken;\nsub fails { die \"nope\\n\" }\n1;\n")

	exempt := map[string]string{
		"My::Broken::fails": "dies on this platform's perl -- synthetic, for TestFreshPlatformExemptionDoesNotFail",
	}
	stale := StaleExemptions(t, p, lib, exempt)

	assert.Empty(t, stale)
}
