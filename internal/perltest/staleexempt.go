// ABOUTME: Checks that a platform exemption claiming a sub dies still actually does, on the perl running now.
// ABOUTME: A name that no longer dies there is a stale exemption: the map should have narrowed or dropped it.

package perltest

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// StaleExemptions reports which of exempt's "Module::name" entries claim
// a death that p's perl, run now, does not actually produce -- the one
// claim every entry in this milestone's platform exemption maps makes
// (a returned-value claim would need the sub's declared return type
// alongside it to check, not just its reason text, so an entry without
// "dies" in its reason is left unchecked).
//
// casesDir, if not "", is a directory of oracle call-case files such as
// internal/infer/testdata/oracle (one <Module-with-dashes>.pl per
// module): do()ing a module's file gives %cases, name => [[args], ...],
// the same arguments and setup (opening scratch files, say) the
// return-type oracle itself uses for that name, so a sub whose "dies"
// claim depends on its argument (POSIX::fcntl, sysconf, WEXITSTATUS,
// ...), not merely on being called with no arguments, is checked
// faithfully; dir is the directory that setup runs in. Called with
// casesDir "" (as the fixture tests do), every sub is called with no
// arguments instead -- this lets a per-OS exemption map be checked for
// staleness without a copy of every platform's perl: the real maps have
// nothing to check on an OS they do not key, so a test proves the
// mechanism with its own synthetic exemption and fixture module instead
// (see TestStalePlatformExemptionFails).
//
// A perl error -- the module cannot even be loaded, or (with casesDir)
// the oracle's own call case cannot be read -- is reported as a failure
// of the check, not skipped as inconclusive: an exemption this cannot
// verify must not be allowed to look confirmed.
func StaleExemptions(t testing.TB, p Perl, inc, dir, casesDir string, exempt map[string]string) []string {
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
		argsExpr := "()"
		if casesDir != "" {
			casesFile := filepath.Join(casesDir, strings.ReplaceAll(module, "::", "-")+".pl")
			argsExpr = fmt.Sprintf(
				`do { my %%cases = do %q or die "cannot read call cases: $@$!\n"; my $c = $cases{%q} or die "no call case for %s\n"; @{ $c->[0] } }`,
				casesFile, name, name)
		}
		script := fmt.Sprintf(`(my $f = %q) =~ s{::}{/}g; require "$f.pm"; eval { &{%q}(%s) }; print $@ eq '' ? 'ok' : 'died';`,
			module, module+"::"+name, argsExpr)
		args := []string{"-e", script}
		if inc != "" {
			args = append([]string{"-I", inc}, args...)
		}
		cmd := p.Command(args...)
		if dir != "" {
			cmd.Dir = dir
		}
		out, err := cmd.Output()
		if err != nil {
			stale = append(stale, fmt.Sprintf("%s: could not check whether it still dies: %v", key, err))
			continue
		}
		if strings.TrimSpace(string(out)) == "ok" {
			stale = append(stale, fmt.Sprintf("%s: exempt because it dies, but calling it on %s did not die", key, p.Path))
		}
	}
	return stale
}
