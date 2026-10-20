// ABOUTME: Checks that a platform exemption claiming a sub dies still actually does, on the perl running now.
// ABOUTME: A name that no longer dies there is a stale exemption: the map should have narrowed or dropped it.

package perltest

import (
	"fmt"
	"strings"
	"testing"
)

// StaleExemptions reports which of exempt's "Module::name" entries claim
// a death that p's perl, run now with no arguments, does not actually
// produce -- the one claim every entry in this milestone's platform
// exemption maps makes (a returned-value claim would need the sub's
// declared return type alongside it to check, not just its reason text,
// so an entry without "dies" in its reason is left unchecked). This lets
// a per-OS exemption map be checked for staleness without a copy of
// every platform's perl: the real maps have nothing to check on an OS
// they do not key, so a test proves the mechanism with its own
// synthetic exemption and fixture module instead (see
// TestStalePlatformExemptionFails).
func StaleExemptions(t testing.TB, p Perl, inc string, exempt map[string]string) []string {
	t.Helper()
	var stale []string
	for key, reason := range exempt {
		if !strings.Contains(reason, "dies") {
			continue
		}
		idx := strings.LastIndex(key, "::")
		if idx < 0 {
			continue
		}
		module, name := key[:idx], key[idx+2:]
		script := fmt.Sprintf(`(my $f = %q) =~ s{::}{/}g; require "$f.pm"; eval { &{%q}() }; print $@ eq '' ? 'ok' : 'died';`,
			module, module+"::"+name)
		args := []string{"-e", script}
		if inc != "" {
			args = append([]string{"-I", inc}, args...)
		}
		out, err := p.Command(args...).Output()
		if err != nil {
			// The call itself could not even be attempted (module not
			// found, say) -- not evidence either way, so this is not
			// claimed stale.
			continue
		}
		if strings.TrimSpace(string(out)) == "ok" {
			stale = append(stale, fmt.Sprintf("%s: exempt because it dies, but calling it on %s did not die", key, p.Path))
		}
	}
	return stale
}
