// ABOUTME: Holds the committed m3 probe corpus to the pinned perl: each snippet's marker counts, psc's verdict on
// ABOUTME: each, and its faithful row. Every parser, sites.go or canon fix extends testdata/m3probes with its own file.

package parse_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
	"tamarou.com/pvm/internal/perltest"
)

// m3ProbeDir holds the committed probe files; testdata/m3probes/README.md
// describes their format.
const m3ProbeDir = "testdata/m3probes"

// probe is one snippet and the facts its file records about it.
type probe struct {
	file, name string
	src        []byte
	// counts is perl's site count per marker; a marker absent here is one
	// perl must report no site of.
	counts map[parseoracle.Marker]int
	// verdicts is psc's expected bucket per counted marker.
	verdicts map[parseoracle.Marker]parseoracle.Bucket
	// faithful asks that the snippet parse and re-emit faithfully.
	faithful bool
	// defect is the git-zhi issue that tracks psc's known disagreement,
	// empty when psc must agree.
	defect string
}

func (p probe) String() string { return p.file + ": " + p.name }

// zhiID is a full git-zhi issue id. A short prefix can name a different
// issue once more exist, so only the whole id is accepted.
var zhiID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// loadProbes reads every *.txt probe file in dir. A missing directory, or
// one with no snippet in it, is an error: a corpus that checks nothing
// must not pass.
func loadProbes(dir string) ([]probe, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.txt"))
	if err != nil {
		return nil, err
	}
	var probes []probe
	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		got, err := parseProbeFile(filepath.Base(path), string(src))
		if err != nil {
			return nil, err
		}
		probes = append(probes, got...)
	}
	if len(probes) == 0 {
		return nil, fmt.Errorf("no probe snippets in %s: a corpus that checks nothing cannot pass", dir)
	}
	return probes, nil
}

// parseProbeFile reads one probe file: lines before the first `=== name`
// are commentary, each `=== name` opens a snippet, and its `--- facts`
// line ends the snippet and begins its facts.
func parseProbeFile(file, text string) ([]probe, error) {
	var probes []probe
	var cur *probe
	inFacts := false
	for i, line := range strings.Split(text, "\n") {
		where := fmt.Sprintf("%s:%d", file, i+1)
		if name, ok := strings.CutPrefix(line, "=== "); ok {
			probes = append(probes, probe{file: file, name: strings.TrimSpace(name),
				counts: map[parseoracle.Marker]int{}, verdicts: map[parseoracle.Marker]parseoracle.Bucket{}})
			cur, inFacts = &probes[len(probes)-1], false
			continue
		}
		if cur == nil {
			continue
		}
		if strings.TrimSpace(line) == "--- facts" {
			if inFacts {
				return nil, fmt.Errorf("%s: a second --- facts in %q", where, cur.name)
			}
			inFacts = true
			continue
		}
		if !inFacts {
			cur.src = append(cur.src, line+"\n"...)
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		if err := cur.addFact(fields); err != nil {
			return nil, fmt.Errorf("%s: %q: %w", where, cur.name, err)
		}
	}
	for _, p := range probes {
		if len(strings.TrimSpace(string(p.src))) == 0 {
			return nil, fmt.Errorf("%s: %q has no snippet", file, p.name)
		}
	}
	return probes, nil
}

// addFact records one fact line: `MARKER COUNT exact|wider`, `faithful`,
// or `known-defect ISSUE-ID`.
func (p *probe) addFact(fields []string) error {
	switch fields[0] {
	case "faithful":
		if len(fields) != 1 {
			return fmt.Errorf("faithful takes nothing after it, got %q", fields)
		}
		p.faithful = true
		return nil
	case "known-defect":
		if len(fields) != 2 || !zhiID.MatchString(fields[1]) {
			return fmt.Errorf("known-defect must name one full git-zhi issue id, got %q", fields[1:])
		}
		p.defect = fields[1]
		return nil
	}
	m := parseoracle.Marker(fields[0])
	if !slices.Contains(parseoracle.Markers, m) {
		return fmt.Errorf("unknown fact %q; a marker is one of %v", fields[0], parseoracle.Markers)
	}
	if _, dup := p.counts[m]; dup {
		return fmt.Errorf("%s is recorded twice", m)
	}
	if len(fields) != 3 {
		return fmt.Errorf("a marker fact is `%s COUNT exact|wider`, got %q", m, fields)
	}
	n, err := strconv.Atoi(fields[1])
	if err != nil || n < 1 {
		return fmt.Errorf("%s's count must be a positive integer (leave out a marker perl reports none of), got %q", m, fields[1])
	}
	switch fields[2] {
	case "exact":
		p.verdicts[m] = parseoracle.BucketExact
	case "wider":
		p.verdicts[m] = parseoracle.BucketWider
	default:
		return fmt.Errorf("%s's verdict must be exact or wider, got %q", m, fields[2])
	}
	p.counts[m] = n
	return nil
}

// probeProblems holds each probe to perl and to psc, and returns every
// disagreement as a sentence naming the snippet.
func probeProblems(perl perltest.Perl, probes []probe) []string {
	var problems []string
	for _, p := range probes {
		problems = append(problems, oneProbeProblems(perl, p)...)
	}
	return problems
}

func oneProbeProblems(perl perltest.Perl, p probe) []string {
	facts, err := parseoracle.Ask(context.Background(), p.src, parseoracle.Options{Perl: perl.Path})
	if err != nil {
		return []string{fmt.Sprintf("%s: asking perl: %v", p, err)}
	}
	// parse_facts.pl's compile is perl -c's: it fails exactly where the
	// snippet does not compile.
	if !facts.OK {
		return []string{fmt.Sprintf("%s: perl -c rejects the snippet: %s", p, strings.TrimSpace(facts.Stderr))}
	}
	if !facts.Walked {
		return []string{fmt.Sprintf("%s: perl compiled the snippet but its CV walk did not run", p)}
	}

	// The recorded facts against perl's: always, even on a known-defect row.
	var problems []string
	for _, m := range parseoracle.Markers {
		if got, want := len(facts.Sites[m]), p.counts[m]; got != want {
			problems = append(problems, fmt.Sprintf("%s: %s: the file records %d site(s), perl 5.44 reports %d", p, m, want, got))
		}
	}
	if len(problems) > 0 {
		return problems
	}

	// psc's facts against perl's, one marker at a time, so each marker's
	// bucket is its own rather than the file's worst.
	var disagree []string
	root := parse.Parse(p.src)
	subject := parseoracle.SubjectFacts{OK: true, KnowsCallSites: true, CallSites: parse.Sites(root, p.src)}
	for _, m := range parseoracle.Markers {
		if p.counts[m] == 0 {
			continue
		}
		one := facts
		one.Sites = map[parseoracle.Marker][]int{m: facts.Sites[m]}
		v := parseoracle.CompareFacts(one, subject)
		if want := p.verdicts[m]; v.Bucket != want {
			disagree = append(disagree, fmt.Sprintf("%s: %s: the file expects psc %s, psc is %s (%s)", p, m, want, v.Bucket, v.Detail))
		}
	}
	if p.faithful {
		if ok, why := parse.Faithful(root, p.src); !ok {
			disagree = append(disagree, fmt.Sprintf("%s: faithful: canon does not re-emit the snippet: %s", p, why))
		}
	}

	switch {
	case p.defect == "":
		return disagree
	case len(disagree) == 0:
		return []string{fmt.Sprintf("%s: known-defect %s, but psc now agrees: record what psc does and drop the row", p, p.defect)}
	default:
		return nil
	}
}

// TestM3ProbesHeldToPerl holds every committed probe to the pinned perl.
func TestM3ProbesHeldToPerl(t *testing.T) {
	perl := perltest.PinnedPerl(t)
	probes, err := loadProbes(m3ProbeDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, problem := range probeProblems(perl, probes) {
		t.Error(problem)
	}
	t.Logf("%d probe snippet(s) held to %s", len(probes), perl.Path)
}

// writeProbeDir writes one probe file into a fresh directory.
func writeProbeDir(t *testing.T, text string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "fixture.txt"), []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestM3ProbesRejectEmptyDir: a missing directory, an empty one, and one
// whose only file holds no snippet each fail to load.
func TestM3ProbesRejectEmptyDir(t *testing.T) {
	for name, dir := range map[string]string{
		"missing":    filepath.Join(t.TempDir(), "absent"),
		"empty":      t.TempDir(),
		"no snippet": writeProbeDir(t, "# commentary only\n"),
	} {
		if _, err := loadProbes(dir); err == nil {
			t.Errorf("%s probe directory loaded without error", name)
		}
	}
}

// TestM3ProbesDetectDriftedFact: a recorded count perl disagrees with is
// reported with the snippet, the marker and both counts.
func TestM3ProbesDetectDriftedFact(t *testing.T) {
	perl := perltest.PinnedPerl(t)
	probes, err := loadProbes(writeProbeDir(t, "=== two refs\nmy @a; my $r = \\@a; my $s = \\@a;\n--- facts\nsrefgen 1 exact\n"))
	if err != nil {
		t.Fatal(err)
	}
	problems := probeProblems(perl, probes)
	want := "fixture.txt: two refs: srefgen: the file records 1 site(s), perl 5.44 reports 2"
	if len(problems) != 1 || problems[0] != want {
		t.Errorf("got %q, want exactly %q", problems, want)
	}
}

// TestM3ProbesKnownDefectNamesAnIssue: a known-defect row must name one
// full git-zhi issue id, or the file does not load.
func TestM3ProbesKnownDefectNamesAnIssue(t *testing.T) {
	for _, row := range []string{"known-defect", "known-defect someday", "known-defect 01a1664d"} {
		if _, err := loadProbes(writeProbeDir(t, "=== x\n1;\n--- facts\n"+row+"\n")); err == nil {
			t.Errorf("%q loaded without error", row)
		}
	}
	if _, err := loadProbes(writeProbeDir(t, "=== x\n1;\n--- facts\nknown-defect 01a1664d-81b7-7c96-8f48-636ae47d5cb1\n")); err != nil {
		t.Errorf("a known-defect row naming a full id was rejected: %v", err)
	}
}

// TestM3ProbesCheckFaithfulRows: a faithful row fails when perl -c rejects
// its snippet or canon re-emits it unfaithfully, and passes otherwise.
func TestM3ProbesCheckFaithfulRows(t *testing.T) {
	perl := perltest.PinnedPerl(t)
	// The unfaithful case is a real canon defect, the dropped `;` after an
	// anonymous-sub statement (01a16c1b-a55d-77fa-bbe2-8b95461248f3):
	// `sub { 1 }; 2;` re-emits as `sub {1;} 2;`. Once that issue is fixed
	// this case passes, and wants another snippet canon re-emits
	// unfaithfully.
	for name, tc := range map[string]struct{ snippet, facts, want string }{
		"compiles and re-emits": {"my $x = 1 + 2;\n", "", ""},
		"perl -c rejects":       {"my $x = ;\n", "", "perl -c rejects the snippet"},
		"unfaithful":            {"sub { 1 }; 2;\n", "", "faithful: canon does not re-emit"},
	} {
		t.Run(name, func(t *testing.T) {
			probes, err := loadProbes(writeProbeDir(t, "=== "+name+"\n"+tc.snippet+"--- facts\n"+tc.facts+"faithful\n"))
			if err != nil {
				t.Fatal(err)
			}
			problems := probeProblems(perl, probes)
			if tc.want == "" {
				if len(problems) != 0 {
					t.Errorf("a faithful snippet failed: %q", problems)
				}
				return
			}
			if len(problems) != 1 || !strings.Contains(problems[0], tc.want) {
				t.Errorf("got %q, want one problem containing %q", problems, tc.want)
			}
		})
	}
}
