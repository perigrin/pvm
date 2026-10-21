// ABOUTME: Checks that a platform exemption's claim -- a sub dies, or returns a type -- still holds on the perl running now.
// ABOUTME: A claim perl no longer bears out is a stale exemption: the map should have narrowed or dropped it.

package perltest

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// ExemptKind is what a platform exemption claims perl does with the
// name it exempts. Each entry says which, rather than leaving
// StaleExemptions to read it out of the entry's reason text.
type ExemptKind int

const (
	// Dies: at least one of the name's oracle call cases dies.
	Dies ExemptKind = iota + 1
	// Returns: at least one call case returns a value of Exemption.Type.
	Returns
	// Unsafe: calling the name can kill perl outright, so it is never
	// called, and so never checked.
	Unsafe
)

// Exemption is one platform exemption as StaleExemptions checks it.
type Exemption struct {
	Kind ExemptKind
	// Type is, for Returns, the type observe_types.pl names the value
	// the exemption claims ("Boolean", "Int", "Undef", ...).
	Type string
	// Calls are how return_oracle.pl calls the name, "CONTEXT:FORM"
	// ("scalar:sub", "list:method", ...), one per form the declaration
	// answers.
	Calls []string
}

// StaleExemptions reports which of exempt's "Module::name" entries make
// a claim that p's perl, run now, does not bear out. Each name is
// called by internal/infer/testdata/return_oracle.pl itself, with its
// module's call cases from casesDir (one <Module-with-dashes>.pl per
// module, internal/infer/testdata/oracle for the real maps): the same
// arguments, setup, method invocant and INSTANCE the return-type
// oracle uses for that name, so the check sees what the oracle would.
// dir, if not "", is the directory that setup runs in, and HOME and
// TMPDIR point there too, as the oracle's own run does.
//
// Anything that keeps the oracle from finishing -- the module cannot be
// loaded, the cases file or the name's case is missing, perl defines no
// such sub -- is reported as a failure of the check, not skipped as
// inconclusive: an exemption this cannot verify must not be allowed to
// look confirmed. So is an entry with no kind, or no calls.
func StaleExemptions(t testing.TB, p Perl, inc, dir, casesDir string, exempt map[string]Exemption) []string {
	t.Helper()
	driver := sourceRelative(t, "..", "infer", "testdata", "return_oracle.pl")
	observer := sourceRelative(t, "..", "infer", "testdata", "observe_types.pl")
	var stale []string
	for _, key := range slices.Sorted(maps.Keys(exempt)) {
		e := exempt[key]
		if e.Kind == Unsafe {
			continue
		}
		idx := strings.LastIndex(key, "::")
		if idx < 0 {
			stale = append(stale, fmt.Sprintf("%s: not a Module::name", key))
			continue
		}
		module, name := key[:idx], key[idx+2:]
		if (e.Kind != Dies && e.Kind != Returns) || len(e.Calls) == 0 {
			stale = append(stale, fmt.Sprintf("%s: an exemption needs a kind and the calls to check it with", key))
			continue
		}
		casesFile := filepath.Join(casesDir, strings.ReplaceAll(module, "::", "-")+".pl")
		args := []string{driver, observer, casesFile, module}
		for _, call := range e.Calls {
			args = append(args, name+":"+call)
		}
		if inc != "" {
			args = append([]string{"-I", inc}, args...)
		}
		cmd := p.Command(args...)
		if dir != "" {
			cmd.Dir = dir
			cmd.Env = append(cmd.Env, "HOME="+dir, "TMPDIR="+dir)
		}
		var stderr strings.Builder
		cmd.Stderr = &stderr
		out, err := cmd.Output()
		// The driver prints "complete" once every call is made, on the
		// STDOUT it keeps for its report: a sub's own output goes to
		// STDERR, so nothing it prints can stand in for this.
		report, complete := strings.CutPrefix(string(out), "complete\n")
		if err != nil || !complete {
			stale = append(stale, fmt.Sprintf("%s: could not check the exemption: the oracle did not finish (%v): %s", key, err, strings.TrimSpace(stderr.String())))
			continue
		}
		var obs []struct{ Type string }
		if err := json.Unmarshal([]byte(report), &obs); err != nil {
			stale = append(stale, fmt.Sprintf("%s: could not check the exemption: the oracle emitted %q: %v", key, report, err))
			continue
		}
		want := "Died"
		if e.Kind == Returns {
			want = e.Type
		}
		if !slices.ContainsFunc(obs, func(o struct{ Type string }) bool { return o.Type == want }) {
			stale = append(stale, fmt.Sprintf("%s: exempt because a call observes %s, but no call case did on %s", key, want, p.Path))
		}
	}
	return stale
}
