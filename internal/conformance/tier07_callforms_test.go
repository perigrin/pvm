// ABOUTME: The tier 07 call-form slice, and the smallest-useful-corpus gate it completes.
// ABOUTME: Parenless argument extent is MEASURED against perl here, not assumed.
package conformance

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// callFormFiles are the corpus files this slice owns, inside
// 07_subroutines.
//
// NAMED RATHER THAN GLOBBED, and that is the whole reason this list
// exists. The tier directory holds two issues' files: declaration, `@_`,
// `return` and signatures belong to 01a0c362-073a and are checked by
// tier07_test.go, and these five are the call-form slice. A glob over
// the directory would make each issue's checks silently answer for the
// other's files -- so when one of them adds a file, the other's
// acceptance would change meaning without anybody editing it.
//
// `03_call_forms.t` predates this issue and is listed anyway: it is the
// four-name-forms file, it IS a call form, and the slice's gate would be
// lying if it excluded the one call-form file that already existed.
// THE UNDECLARED CALLEE IS NOT IN THIS LIST, and the reason changed
// under issue 01a0c605-eff0. perl REFUSES that source -- the refusal is
// the whole measurement -- so it carries `--- expect parsent`, and
// measured at dc1bea2c the corpus had no such file and could not have
// one: `TestCorpusLints` and every tier's `*Lint` called `opsOf`
// unconditionally, and `perl -MO=Concise` exits 255 on a syntax error
// exactly as `perl -c` does. A `parsent` file anywhere failed them.
//
// It can now. `lintFile` declines the lint for a file whose header says
// perl refuses it, and `conformance/07_subroutines/10_undeclared_callee.t`
// is that file -- the corpus's first, holding the fact fact 1 below
// measures. It stays out of THIS list because the list is what keeps the
// slice's checks from answering for the rest of the tier, and the gate
// below admits only files that PASS or refuse with a citation; a file
// our parser is never asked to read is neither.
//
// The fact remains measured in TestTierCallFormsParenlessExtent as well.
// That is not a duplicate: the corpus file pins the refusal and its
// tokens, and the Go check pins it beside the two extent measurements it
// is the precondition of, where a reader meets all three at once.
// callFormTopics is the tier-07 slice the milestone's target names:
// "tiers 01-04 plus a call-form slice from tier 07".
//
// TOPICS rather than file names. The slice used to be four `.t` files;
// those cases are now sections of two topics, and naming the topics is
// both findable and closer to what the target means -- a slice of the
// tier, not four specific paths.
var callFormTopics = map[string]bool{
	"subroutines.md":     true,
	"argument-extent.md": true,
}

// callFormSlice returns the slice's parsed files, keyed by base name.
//
// Fails rather than skipping when one is missing. A slice whose gate
// quietly ran over four files instead of five would still say PASS, and
// the one thing this issue delivers is a gate that means something.
func callFormSlice(t *testing.T) map[string]*File {
	t.Helper()

	out := map[string]*File{}
	for key, f := range tierFiles(t, tierSubroutines) {
		topic, _, _ := strings.Cut(key, "/")
		if callFormTopics[topic] {
			out[key] = f
		}
	}
	if len(out) == 0 {
		t.Fatalf("the call-form slice is empty; %v name no topic in %s",
			callFormTopics, tierSubroutines)
	}
	return out
}

// parenlessProbe is one source spelled once and compiled twice, under a
// callee with a prototype and a callee without.
//
// ONE PROGRAM TEXT, TWO DECLARATIONS. The call site below -- `print f 1,
// 2;` -- is byte-identical in both, and only the `sub f` line above it
// differs. That is what makes this a measurement of EXTENT rather than
// of two unrelated programs: if the two runs print the same thing, the
// probe has discriminated nothing and says so.
//
// `%s` is where the prototype goes, empty or `($)`.
const parenlessProbe = `sub f %s{ return "f[" . join(",", @_) . "]" }
print f 1, 2;
print "\n";
`

// callFormSpellings are the spellings that stand for each call form, keyed
// by the file that introduces it.
//
// A table rather than a derivation, for the reason subroutineConstructs
// gives: the forms have no common syntactic frame, and a regex loose
// enough to catch `&f`, `f 1, 2` and `$c->()` alike would match so much
// that containment in the adjacency file would stop meaning anything.
//
// A file with no entry here must have one in `adjacencyExempt`, or the
// loop below fails it. An exemption is a measured reason the construct
// CANNOT sit in the shared body; silence is how a corpus stops covering
// a construct without anybody deciding to.
var callFormSpellings = map[string]string{
	"03_call_forms.t":       "&answer,",
	"08_parenless_extent.t": "print f 1, 2;",
	"09_prototype_extent.t": "print g 1, 2;",
	"11_code_ref_call.t":    "->(",
}

// adjacencyExempt names the slice files whose construct cannot appear in
// the adjacency file, and why.
//
// An EXPLICIT exemption rather than silence. A form left out of the
// adjacency check because nobody noticed is indistinguishable from one
// left out because it cannot be there, and the first is how a corpus
// stops covering a construct.
// EMPTY TODAY, AND THAT IS THE FINDING. The prototype form looked like
// it had to be exempt: `sub g ($)` is a PROTOTYPE only where the
// signatures feature is off, the adjacency file says `use v5.36`, and
// under it perl reads the same three characters as a SIGNATURE and
// enforces arity --
//
//	$ perl -e 'use v5.36; sub g ($) { "g" } print g 1, 2;'
//	Too many arguments for subroutine 'main::g' (got 2; expected 1)
//
// -- so the extent-cutting behaviour is simply absent there. Measured
// 5.42.0, `no feature "signatures"` in a block restores it inside a file
// that otherwise enables them, so the form joins the shared body after
// all and the pair that matters -- `print f 1, 2` printing `f[1-2]`
// beside `print g 1, 2` printing `g[1]2` -- is reachable.
//
// The map stays because the check needs somewhere to put a form that
// genuinely cannot share the body, and an exemption must be a written
// measurement rather than an omission nobody noticed.
var adjacencyExempt = map[string]string{}

// smallestUsefulCorpus is the milestone's stated delivery target, spelled
// as the spec spells it: "tiers 01-04 plus a call-form slice from tier 07".
var smallestUsefulCorpus = []string{
	"01_literals",
	"02_variables",
	"03_context",
	"04_operators",
}

// TestSmallestUsefulCorpus is the gate: tiers 01-04 plus the call-form
// slice are all green.
//
// WHAT "GREEN" HAS TO MEAN HERE, because the obvious two readings are
// both worthless and the gate is the issue's one delivery target.
//
// A gate demanding ZERO REFUSALS fails today and would fail forever. Tier
// 01 documents five refusals and tier 04 five; they are the corpus doing
// its job, naming constructs the parser does not handle, and a milestone
// gate that cannot be met until the parser is finished is a gate nobody
// can act on. It would also invert the corpus's design, which SKIPS a
// documented refusal precisely so the suite stays pristine while still
// recording the gap.
//
// A gate IGNORING refusals asserts nothing. Every file could refuse and
// it would still say PASS, which is the failure eight of the fourteen
// tiers in this milestone shipped in one form or another.
//
// So the gate demands what `verdict` in run.go actually distinguishes.
// Its five outcomes divide cleanly into three that are the corpus working
// and two that are it broken:
//
//   - `passed` -- our parser handled the file. Green.
//   - `knownRefusal` -- our parser did not, and the FILE SAYS SO, citing
//     an issue and, where the refusal is in the parser, the site. Green:
//     this is a measurement that succeeded, recording a gap accurately.
//   - `corpusBug` -- perl disagrees with the file's own expectation. NOT
//     green: nothing was measured about our parser at all.
//   - `staleMarker` -- the file claims a refusal it no longer has, or
//     names a site that no longer declines. NOT green: the file is still
//     being skipped on a claim its header no longer describes, which is
//     exactly how a corpus silently stops measuring.
//   - `refused` -- our parser failed and the file does not admit it. NOT
//     green: an undocumented regression.
//
// The gate is therefore: every file in tiers 01-04 and in the slice
// reaches `passed` or `knownRefusal`, and the count of each is REPORTED
// so a refusal budget that grows is visible in the log rather than
// silently absorbed. That is the strongest statement that is true today
// and that stays meaningful as the parser improves: a gap may exist, but
// it must be documented, cited, and still accurate.
func TestSmallestUsefulCorpus(t *testing.T) {
	type outcome struct {
		file string
		v    verdictResult
	}

	var bad []outcome
	green, refusing := 0, 0
	var refusals []string

	check := func(label string, f *File) {
		v := verdict(t, f)
		switch v.kind {
		case passed:
			green++
		case knownRefusal:
			refusing++
			refusals = append(refusals, fmt.Sprintf("%s (%s)", label, f.Refuses))
		default:
			bad = append(bad, outcome{label, v})
		}
	}

	for _, tier := range smallestUsefulCorpus {
		for name, f := range tierFiles(t, tier) {
			check(tier+"/"+name, f)
		}
	}
	for name, f := range callFormSlice(t) {
		check(tierSubroutines+"/"+name+" (slice)", f)
	}

	sort.Strings(refusals)
	t.Logf("smallest useful corpus: %d files pass, %d refuse with a cited gap",
		green, refusing)
	for _, r := range refusals {
		t.Logf("  documented refusal: %s", r)
	}

	// A gate over no cases would pass vacuously, which is the one
	// failure this test cannot afford. Tiers 01-04 alone hold over a
	// hundred cases, so the floor is only here to make a glob that
	// returns nothing FAIL rather than report success.
	const floor = 50
	if total := green + refusing + len(bad); total < floor {
		t.Fatalf("the gate ran over %d cases, which cannot be tiers 01-04 "+
			"plus a call-form slice.\n\tA gate over nothing passes vacuously.",
			total)
	}

	sort.Slice(bad, func(i, j int) bool { return bad[i].file < bad[j].file })
	for _, b := range bad {
		t.Errorf("%s is %s, not green:\n\t%s\n"+
			"\tThe gate admits a file that PASSES and a file that refuses "+
			"with a documented, still-accurate citation. It admits neither "+
			"a corpus bug, nor a marker that has gone stale, nor an "+
			"undocumented refusal.",
			b.file, b.v.kind, strings.Join(b.v.msgs, "\n\t"))
	}
}

// TestLintSkipsAParsentFile pins the skip that lets this tier carry
// `10_undeclared_callee.t`, the corpus's first `--- expect parsent` file.
//
// Measured 5.42.0, `perl -MO=Concise,-exec` exits 255 on a syntax error
// exactly as `perl -c` does, so `opsOf` returns an error for a source
// perl refuses. Every tier's `*Lint` test and `TestCorpusLints` reach
// `lintOps` through `lintFile`, which declines for a file whose own
// header says perl refuses it.
//
// The claim is checked on the corpus file rather than on a fixture,
// because a fixture would prove the guard works on a source this test
// chose and say nothing about whether the file that needed it is reached
// by the guard.
func TestLintSkipsAParsentFile(t *testing.T) {
	// ANY parsent case, found rather than named. The old version
	// hardcoded `10_undeclared_callee.t` and called it "the corpus's
	// only parsent file", which stopped being true when tier 04 gained
	// two more and stopped being findable when cases replaced files.
	var found *File
	var where, foundTier string
	paths, err := filepath.Glob(filepath.Join(corpusDir, "mdtest", "*.md"))
	if err != nil {
		t.Fatalf("globbing topics: %v", err)
	}
	for _, path := range paths {
		if filepath.Base(path) == "FORMAT.md" {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		tier, err := topicTier(string(raw))
		if err != nil {
			continue
		}
		cases, err := ParseTopic(string(raw))
		if err != nil {
			t.Fatalf("%s: %v", filepath.Base(path), err)
		}
		for _, c := range cases {
			if c.ExpectParsent {
				found = c.asFile()
				where = filepath.Base(path) + "/" + c.Title
				foundTier = tier
				break
			}
		}
		if found != nil {
			break
		}
	}
	if found == nil {
		t.Fatalf("the corpus has no `parses: no` case.\n" +
			"\tWithout one nothing exercises the lint's skip, and the " +
			"skip becomes a branch no test reaches.")
	}

	tiers, err2 := readTierOps(corpusDir)
	if err2 != nil {
		t.Fatalf("reading tier READMEs: %v", err2)
	}

	// The claim: a case perl refuses emits NO OPS -- perl builds no
	// optree for a program it will not compile -- so the op budget
	// cannot bind it and the lint must say nothing.
	if err := lintFile(t, found, foundTier, tiers); err != nil {
		t.Errorf("the lint reports on a `parses: no` case (%s): %v\n"+
			"\tperl builds no optree for a program it refuses, so there "+
			"are no ops to be over budget.", where, err)
	}
}

// TestLintErrorsOnABrokenParsesFile is the other half, and it is the
// reason the skip is keyed on the file's DECLARATION rather than on
// perl's exit status.
//
// The tempting fix for the `parsent` problem is to make `opsOf` return an
// empty op set instead of an error whenever perl refuses. It would be
// wrong: a file claiming `--- expect parses` whose source is genuinely
// broken would then LINT CLEAN, silently, because an empty op set
// violates no tier. The corpus would report nothing about a file that
// compiles nothing.
//
// So this runs the lint over a `parses` file with a real syntax error and
// demands an error. Mutation-tested: rekeying `lintFile`'s guard on
// whether perl compiles the source, rather than on `ExpectParsent`, turns
// this red while every other check in the package stays green.
func TestLintErrorsOnABrokenParsesFile(t *testing.T) {
	// A `parses` claim over the same source the parsent file carries.
	// The only difference is which section it declares, which is exactly
	// the distinction under test.
	broken := &File{
		Source:       "f 1, 2;\nsub f { return \"f[@_]\" }\n",
		ExpectParses: true,
	}

	if compiles, _ := askPerl(t, broken.Source); compiles {
		t.Fatalf("perl compiles %q, so this fixture is not broken and "+
			"measures nothing", broken.Source)
	}

	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	if err := lintFile(t, broken, tierSubroutines, tiers); err == nil {
		t.Errorf("the lint accepted a `parses` file perl REFUSES.\n" +
			"\tAn empty op set violates no tier, so a lint that swallowed " +
			"perl's refusal would pass this file forever while it " +
			"compiled nothing. The skip is keyed on the file's own " +
			"`parsent` declaration for exactly this reason.")
	}
}
