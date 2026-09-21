// ABOUTME: Extracts t/lib/croak/ cases and splits them by what perl -c says.
// ABOUTME: The directory is named croak; 40% of its cases compile perfectly well.
package conformance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// croakCorpusEnv names a perl checkout, overriding the default below.
//
// `internal/parseoracle` already faces "an external perl tree at a path
// that varies" and answers it with PERL5_CORPUS; this is the same variable
// rather than a third convention for the same fact.
const croakCorpusEnv = "PERL5_CORPUS"

// defaultPerl5Checkout is the tree the split was measured against.
//
// A default rather than a hard requirement: the env var above is the
// escape hatch, and an absent corpus SKIPS with the path named rather than
// failing, because a missing external checkout is not a bug in this repo.
const defaultPerl5Checkout = "~/dev/perl5"

// The recorded split, measured over the corpus under perl 5.42.0 and
// written into docs/plans/2026-09-21-graded-conformance-corpus.md.
//
// These are a RECORD, not a rule. Nothing fails for disagreeing with them;
// DriftReport says so out loud instead, because a count that silently
// rebalances hides exactly the fact the split exists to expose.
const (
	recordedCompileFail = 201
	recordedRuntimeOnly = 64
	recordedExitsZero   = 75
)

// CroakVerdict is what perl says about one case, which is the only thing
// that decides whether it is a negative.
type CroakVerdict int

const (
	// CroakUnclassified is the zero value: measured by nothing. It exists
	// so an unmeasured case is distinguishable from a compile failure
	// rather than defaulting into the bucket that matters most.
	CroakUnclassified CroakVerdict = iota

	// CroakCompileFail is `perl -c` rejecting the source. These, and only
	// these, are genuine `parsent` negatives.
	CroakCompileFail

	// CroakRuntimeOnly is `perl -c` ACCEPTING and the program then dying.
	// perl parses these, so a parser that parses them is right, and the
	// croak is not a parser fact at all.
	CroakRuntimeOnly

	// CroakExitsZero is neither: the case does not fail under this
	// interpreter. Mostly the `signatures` file, which is blead-5.45
	// syntax against a 5.42.0 interpreter -- a version fact, not a
	// parser fact.
	CroakExitsZero
)

func (v CroakVerdict) String() string {
	switch v {
	case CroakCompileFail:
		return "compile-fail"
	case CroakRuntimeOnly:
		return "runtime-only"
	case CroakExitsZero:
		return "exits-zero"
	}
	return "unclassified"
}

// CroakCase is one `EXPECT` block from one file in t/lib/croak/.
//
// Identity is (File, Ordinal), not Name: 87 of the cases carry no
// `# NAME`, so a name-keyed case set would silently collapse them.
type CroakCase struct {
	File    string // basename within t/lib/croak/, e.g. "pp_ctl"
	Ordinal int    // 1-based position within that file
	Name    string // the `# NAME` line, empty for the 87 without one
	Source  string // the program, with the NAME line removed
	Expect  string // perl's expected diagnostic, NOT asserted by us
	Verdict CroakVerdict
}

// Ref is the case's identity, for a failure message.
func (c CroakCase) Ref() string {
	if c.Name != "" {
		return fmt.Sprintf("%s#%d (%s)", c.File, c.Ordinal, c.Name)
	}
	return fmt.Sprintf("%s#%d", c.File, c.Ordinal)
}

// Citation names where a case came from, precisely enough to go and look.
//
// perl is licensed separately, so the corpus is REFERENCED rather than
// vendored. That only works if the reference resolves: the filename says
// which file, the revision says which version of it, and the ordinal says
// which case -- and without all three the citation points at a moving
// target.
func (c CroakCase) Citation(revision string) string {
	return fmt.Sprintf("perl5@%s t/lib/croak/%s case %d", revision, c.File, c.Ordinal)
}

// IsNegative reports whether a case is a `parsent` must-not-parse case.
//
// This is the whole issue in one method. Trusting the directory name would
// return true for everything and assert must-not-parse on the 139 cases
// perl parses happily -- marking a CORRECT parser wrong on 40% of them.
func (c CroakCase) IsNegative() bool { return c.Verdict == CroakCompileFail }

// CroakSplit is the corpus classified three ways.
type CroakSplit struct {
	Cases       []CroakCase
	CompileFail []CroakCase
	RuntimeOnly []CroakCase
	ExitsZero   []CroakCase
	Skipped     []CroakCase // needs a -switch or --FILE--, excluded explicitly
}

// croakCorpusDir resolves the corpus directory.
func croakCorpusDir() string {
	root := strings.TrimSpace(os.Getenv(croakCorpusEnv))
	if root == "" {
		root = defaultPerl5Checkout
	}
	return filepath.Join(expandHome(root), "t", "lib", "croak")
}

// croakRevision reads the pinned revision of the checkout a corpus dir
// lives in, by reading git's own files.
//
// Reading .git directly rather than shelling out to `git -C`: this repo is
// worked on from git worktrees, where a `git -C` at an unrelated tree is
// refused, and the revision is a two-line file read either way. HEAD names
// a ref, the ref file holds the sha, and packed-refs holds it when the
// loose file has been packed away.
func croakRevision(corpusDir string) (string, error) {
	// t/lib/croak -> the checkout root.
	root := filepath.Dir(filepath.Dir(filepath.Dir(corpusDir)))
	gitDir := filepath.Join(root, ".git")

	head, err := os.ReadFile(filepath.Join(gitDir, "HEAD"))
	if err != nil {
		return "", fmt.Errorf("reading %s/HEAD: %w", gitDir, err)
	}
	h := strings.TrimSpace(string(head))

	// A detached HEAD holds the sha outright.
	ref, isRef := strings.CutPrefix(h, "ref: ")
	if !isRef {
		return h, nil
	}

	if sha, err := os.ReadFile(filepath.Join(gitDir, filepath.FromSlash(ref))); err == nil {
		return strings.TrimSpace(string(sha)), nil
	}

	packed, err := os.ReadFile(filepath.Join(gitDir, "packed-refs"))
	if err != nil {
		return "", fmt.Errorf("%s resolves to %s, which has no loose or packed ref", h, ref)
	}
	for _, line := range strings.Split(string(packed), "\n") {
		sha, name, ok := strings.Cut(line, " ")
		if ok && name == ref {
			return sha, nil
		}
	}
	return "", fmt.Errorf("%s resolves to %s, which is in neither refs/ nor packed-refs", h, ref)
}

// Perl's own harness (t/test.pl, run_multiple_progs) defines the format,
// and these mirror it rather than reinventing a reading of the files.
var (
	// Cases are separated by a row of at least four `#`. The FIRST case
	// in each file has no leading separator, which is why the separator
	// count (327) is not the case count (341).
	reCroakSeparator = regexp.MustCompile(`(?m)^#{4,}[ \t]*\r?\n`)

	// `EXPECT` on a line of its own ends the program and begins the
	// expected diagnostic.
	reCroakExpect = regexp.MustCompile(`(?m)^EXPECT[ \t]*\r?$`)

	// Everything before `__END__` is the file's preamble, not a case.
	reCroakEnd = regexp.MustCompile(`(?m)^__END__[ \t]*\r?\n`)

	reCroakName   = regexp.MustCompile(`(?m)^#[ \t]*NAME[ \t]+(.+)\r?\n`)
	reCroakSwitch = regexp.MustCompile(`^[ \t]*-\w+`)
	reCroakSkip   = regexp.MustCompile(`(?m)^#[ \t]*SKIP\b`)
)

// ExtractCroakCases reads every case out of a t/lib/croak/ directory.
//
// The files carry NO extension, so a `*.t` glob matches nothing there and
// the directory is read whole. Everything before `__END__` is the file's
// preamble and not a case.
func ExtractCroakCases(dir string) ([]CroakCase, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var cases []CroakCase
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		// Perl's own harness skips these, so we do too.
		if strings.HasSuffix(e.Name(), "~") || strings.HasSuffix(e.Name(), ".orig") ||
			strings.HasSuffix(e.Name(), ",v") {
			continue
		}
		src, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		cases = append(cases, croakCasesIn(e.Name(), string(src))...)
	}
	sort.SliceStable(cases, func(i, j int) bool {
		if cases[i].File != cases[j].File {
			return cases[i].File < cases[j].File
		}
		return cases[i].Ordinal < cases[j].Ordinal
	})
	if len(cases) == 0 {
		return nil, fmt.Errorf("%s holds no EXPECT blocks", dir)
	}
	return cases, nil
}

// croakCasesIn splits one file into its cases.
func croakCasesIn(file, src string) []CroakCase {
	// The tests begin after __END__; before it is the file's preamble.
	loc := reCroakEnd.FindStringIndex(src)
	if loc == nil {
		return nil
	}
	body := src[loc[1]:]

	var out []CroakCase
	for _, chunk := range reCroakSeparator.Split(body, -1) {
		loc := reCroakExpect.FindStringIndex(chunk)
		if loc == nil {
			continue // not a case: trailing whitespace, or a stray comment
		}
		prog := chunk[:loc[0]]
		expect := chunk[loc[1]:]

		name := ""
		if m := reCroakName.FindStringSubmatch(prog); m != nil {
			name = strings.TrimSpace(m[1])
			prog = reCroakName.ReplaceAllString(prog, "")
		}

		out = append(out, CroakCase{
			File:    file,
			Ordinal: len(out) + 1,
			Name:    name,
			Source:  prog,
			Expect:  strings.TrimRight(expect, "\n"),
		})
	}
	return out
}

// needsSwitch reports whether a case cannot be run as a plain program.
//
// A leading `-w`-style switch changes how perl is invoked, and a
// `--FILE--` block writes auxiliary files first. Both are excluded
// EXPLICITLY -- a case silently misclassified because its switch was
// ignored would be worse than one left out and named.
func (c CroakCase) needsSwitch() bool {
	return reCroakSwitch.MatchString(c.Source) ||
		strings.Contains(c.Source, "--FILE--") ||
		reCroakSkip.MatchString(c.Source)
}

// MeasureCroakSplit classifies every case by RUNNING perl.
//
// Measured, never read from a list. A checked-in table of verdicts goes
// stale the moment perl gains a diagnostic or moves one from run time to
// compile time, and it goes stale silently: the numbers keep reporting the
// answer that was true once. The classification costs a few seconds over
// the whole corpus, which is cheaper than being wrong about 40% of it.
func MeasureCroakSplit(cases []CroakCase) (*CroakSplit, error) {
	perl, err := perlPath()
	if err != nil {
		return nil, err
	}

	split := &CroakSplit{}
	for _, c := range cases {
		if c.needsSwitch() {
			split.Skipped = append(split.Skipped, c)
			continue
		}
		v, err := classifyCroakCase(perl, c.Source)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", c.Ref(), err)
		}
		c.Verdict = v
		split.Cases = append(split.Cases, c)
		switch v {
		case CroakCompileFail:
			split.CompileFail = append(split.CompileFail, c)
		case CroakRuntimeOnly:
			split.RuntimeOnly = append(split.RuntimeOnly, c)
		case CroakExitsZero:
			split.ExitsZero = append(split.ExitsZero, c)
		}
	}
	return split, nil
}

// classifyCroakCase asks perl the one question that decides the verdict.
//
// `perl -c` first, because compile failure is what `parsent` means. Only
// when perl ACCEPTS the source does running it tell us anything more, and
// what it tells us is that the croak is a runtime fact.
//
// Source is fed on stdin rather than written to a temp file: `-` is the
// filename perl's own croak EXPECT lines already name ("at - line 2"), so
// this is the invocation the expectations were written against.
func classifyCroakCase(perl, src string) (CroakVerdict, error) {
	compiles, err := runCroakPerl(perl, src, "-c")
	if err != nil {
		return CroakUnclassified, err
	}
	if !compiles {
		return CroakCompileFail, nil
	}
	runs, err := runCroakPerl(perl, src)
	if err != nil {
		return CroakUnclassified, err
	}
	if !runs {
		return CroakRuntimeOnly, nil
	}
	return CroakExitsZero, nil
}

// runCroakPerl runs one case and reports whether perl exited 0.
//
// An ExitError is the ANSWER here, not a failure: a nonzero exit is what
// "this case croaks" means. Only a perl that could not be started at all
// is an error worth returning.
func runCroakPerl(perl, src string, args ...string) (bool, error) {
	cmd := exec.Command(perl, append(args, "-")...)
	cmd.Stdin = strings.NewReader(src)
	cmd.Stdout = nil
	cmd.Stderr = nil
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if _, ok := err.(*exec.ExitError); ok {
		return false, nil
	}
	return false, err
}

// DriftReport states the measured split against the recorded one.
//
// The recorded figures are a measurement someone took once, not a
// contract, so this REPORTS rather than fails. What it must never do is
// stay quiet: a runner that rebalanced its buckets without saying so would
// conceal the one fact -- that the split moved -- which tells a reader the
// corpus is no longer describing the perl in front of them.
func (s *CroakSplit) DriftReport() string {
	rev := "unknown"
	if r, err := croakRevision(croakCorpusDir()); err == nil {
		rev = r
	}

	cf, rt, ez := len(s.CompileFail), len(s.RuntimeOnly), len(s.ExitsZero)

	var b strings.Builder
	fmt.Fprintf(&b, "croak split, measured over %d cases from perl5@%s\n",
		len(s.Cases), rev)
	fmt.Fprintf(&b, "  %-14s %14s %14s\n", "", "recorded", "measured")
	fmt.Fprintf(&b, "  %-14s %14d %14d%s\n", "compile-fail", recordedCompileFail, cf, croakDrift(recordedCompileFail, cf))
	fmt.Fprintf(&b, "  %-14s %14d %14d%s\n", "runtime-only", recordedRuntimeOnly, rt, croakDrift(recordedRuntimeOnly, rt))
	fmt.Fprintf(&b, "  %-14s %14d %14d%s\n", "exits-zero", recordedExitsZero, ez, croakDrift(recordedExitsZero, ez))
	fmt.Fprintf(&b, "  %-14s %14d %14d%s\n", "total",
		recordedCompileFail+recordedRuntimeOnly+recordedExitsZero, cf+rt+ez,
		croakDrift(recordedCompileFail+recordedRuntimeOnly+recordedExitsZero, cf+rt+ez))

	if len(s.Skipped) > 0 {
		fmt.Fprintf(&b, "  %d case(s) excluded, needing a -switch or --FILE--:\n", len(s.Skipped))
		for _, c := range s.Skipped {
			fmt.Fprintf(&b, "    %s\n", c.Ref())
		}
	}

	// Per-file detail, because a total that moved says nothing about
	// WHERE, and where is what a reader needs to go and look.
	b.WriteString("  per file:\n")
	type row struct{ cf, rt, ez int }
	byFile := map[string]*row{}
	var order []string
	for _, c := range s.Cases {
		r, ok := byFile[c.File]
		if !ok {
			r = &row{}
			byFile[c.File] = r
			order = append(order, c.File)
		}
		switch c.Verdict {
		case CroakCompileFail:
			r.cf++
		case CroakRuntimeOnly:
			r.rt++
		case CroakExitsZero:
			r.ez++
		}
	}
	sort.Strings(order)
	for _, f := range order {
		r := byFile[f]
		fmt.Fprintf(&b, "    %-12s compile-fail=%-4d runtime-only=%-4d exits-zero=%-4d\n",
			f, r.cf, r.rt, r.ez)
	}
	return b.String()
}

// croakDrift annotates a measured figure that no longer matches the record.
func croakDrift(recorded, measured int) string {
	if recorded == measured {
		return ""
	}
	return fmt.Sprintf("   DRIFT %+d", measured-recorded)
}

// itoa is strconv.Itoa under a shorter name, for the assertions that look
// for a number inside a report.
func itoa(n int) string { return strconv.Itoa(n) }
