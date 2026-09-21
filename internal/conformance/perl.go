// ABOUTME: Resolves the one perl every corpus measurement runs through.
// ABOUTME: A corpus header saying MEASURED perl 5.42.0 is only true if this is.
package conformance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// measuredVersion is the $] every corpus expectation was taken from.
//
// $] is perl's numeric version: 5.042000 is 5.42.0. Compared as a string
// because that is exactly what the interpreter prints, and parsing it into
// a float to compare would introduce a rounding question where there is
// none.
const measuredVersion = "5.042000"

// pinnedPerlEnv names an explicit interpreter, overriding the search.
//
// The search below is a heuristic about one machine; this is the escape
// hatch for every other. CI, a container, or a contributor with perl
// somewhere unusual sets it and stops guessing.
const pinnedPerlEnv = "PVM_PERL"

// perlCandidates are the paths tried, in order, when PVM_PERL is unset.
//
// Bare `perl` comes first because on a correctly set up machine it IS the
// right one, and hard-coding absolute paths ahead of it would override a
// contributor's deliberate choice. What makes that safe is that every
// candidate is VERSION-CHECKED before it is accepted, so this list is a
// search order and not a statement of trust.
//
// The rest are fallbacks for when PATH resolves to the wrong perl, which
// is not hypothetical: measured on the machine this was written on, PATH
// carried three -- ~/.local/bin/perl at 5.42.0, then /usr/bin/perl and
// /bin/perl both at 5.38.2 -- and the right one won by ordering alone.
// A plenv installation is tried too, since CLAUDE.md names plenv as the
// project's version manager even where it is not currently installed.
var perlCandidates = []string{
	"perl",
	"~/.plenv/shims/perl",
	"~/.local/bin/perl",
	"/usr/local/bin/perl",
	"/usr/bin/perl",
}

var (
	perlOnce  sync.Once
	perlFound string
	perlErr   error
)

// perlPath returns the interpreter every corpus measurement runs through.
//
// One resolver rather than two call sites, because `askPerl` adjudicates
// what a file PRINTS and `opsOf` reads the OPTREE the lint checks. Split
// across two interpreters those could disagree silently -- the op list
// describing one language and the output another -- which is a worse
// failure than a single wrong perl, because nothing in the suite would
// have a reason to report it.
//
// Resolved once per process: the answer cannot change mid-run, and every
// corpus file would otherwise pay for the same version check.
func perlPath() (string, error) {
	perlOnce.Do(func() {
		perlFound, perlErr = findPerl()
	})
	return perlFound, perlErr
}

// findPerl returns the first candidate that IS the measured version.
//
// Checking the version rather than merely finding an executable is the
// point. A perl that exists and is wrong produces no error and no crash;
// it produces a corpus that measures a different language and reports the
// disagreement as CORPUS BUG, blaming the file for the environment.
func findPerl() (string, error) {
	if pinned := strings.TrimSpace(envPerl()); pinned != "" {
		v, err := perlVersion(pinned)
		if err != nil {
			return "", fmt.Errorf("%s=%s: %w", pinnedPerlEnv, pinned, err)
		}
		if v != measuredVersion {
			return "", fmt.Errorf("%s=%s reports $] = %s, want %s",
				pinnedPerlEnv, pinned, v, measuredVersion)
		}
		return pinned, nil
	}

	var tried []string
	for _, c := range perlCandidates {
		v, err := perlVersion(c)
		if err != nil {
			tried = append(tried, fmt.Sprintf("%s (not runnable)", c))
			continue
		}
		if v == measuredVersion {
			// Expanded, because this is what every caller execs.
			return expandHome(c), nil
		}
		tried = append(tried, fmt.Sprintf("%s ($] = %s)", c, v))
	}
	return "", fmt.Errorf(
		"no perl reporting $] = %s; tried %s\n"+
			"\tevery corpus file records MEASURED perl 5.42.0, and every "+
			"expectation in it was taken from that interpreter\n"+
			"\tset %s to name one explicitly",
		measuredVersion, strings.Join(tried, ", "), pinnedPerlEnv)
}

// perlVersion runs one candidate and returns its $].
func perlVersion(path string) (string, error) {
	out, err := exec.Command(expandHome(path), "-e", "print $]").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// envPerl reads the PVM_PERL override.
func envPerl() string { return os.Getenv(pinnedPerlEnv) }

// expandHome resolves a leading `~/` against the user's home directory.
//
// exec.Command does not do this -- `~` is a shell convention, and a path
// beginning with it would be looked up literally and never found. The
// candidate list needs it so a per-user install can be named without
// hard-coding whose machine it is.
func expandHome(path string) string {
	rest, ok := strings.CutPrefix(path, "~/")
	if !ok {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Unresolvable, so leave it alone: the candidate simply fails its
		// version check like any other path that is not runnable.
		return path
	}
	return filepath.Join(home, rest)
}
