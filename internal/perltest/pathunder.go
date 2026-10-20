// ABOUTME: Checks whether a path lies under a directory, canonicalizing both on Windows first.
// ABOUTME: An 8.3 short name, a differently-cased drive letter, or backslashes versus forward slashes can otherwise make the same directory compare unequal there.

package perltest

import (
	"runtime"
	"strings"
	"testing"
)

// PathUnder reports whether path is under dir. On Windows, perl's own
// %INC entries and $Config{privlibexp}/$Config{archlibexp} can each
// spell the same directory differently -- an 8.3 short name, a
// differently-cased drive letter, or backslashes versus forward slashes
// (measured: TestHeldToPerlFindsArchlibModules, which compares exactly
// these two) -- so both are canonicalized there (long path name,
// lowercase, forward slashes) through p before comparing. Elsewhere this
// is a plain string prefix check; no perl is run.
func PathUnder(t testing.TB, p Perl, path, dir string) bool {
	t.Helper()
	if runtime.GOOS != "windows" {
		return strings.HasPrefix(path, dir)
	}
	script := `my ($a, $b) = @ARGV; require Win32; ` +
		`for ($a, $b) { $_ = lc(Win32::GetLongPathName($_) =~ tr{\\}{/}r) } ` +
		`print index($a, $b) == 0 ? "yes" : "no";`
	out, err := p.Command("-e", script, path, dir).Output()
	if err != nil {
		t.Fatalf("PathUnder(%s, %s): %v", path, dir, err)
	}
	return strings.TrimSpace(string(out)) == "yes"
}
