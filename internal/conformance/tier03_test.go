// ABOUTME: Tier 03 checked against the finished tooling: scalar versus list context.
// ABOUTME: The tier whose subject is a FLAG on an op, which the op-name lint cannot see.
package conformance

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// The tier constant this file is about is `tierContext`, declared in
// tier10_test.go because 10_io needs to name its own prerequisite. One
// constant for one tier, in the package that holds both: a second
// spelling here would be the two-lists-that-must-agree failure the tier
// READMEs already exist to avoid.

// TestTierContextPerlValidated runs every file in the tier through the
// pinned interpreter before it counts.
//
// The same two claims tier 01 checks -- perl agrees with the parses bit,
// perl prints what the file pins -- and this tier has the strongest
// reason of any to want them checked. Its subject leaves NO op behind:
// `my $n = @a` and `my $n = scalar(@a)` compile to byte-identical
// optrees, so nothing in the op stream distinguishes the two halves of a
// discriminating pair. What distinguishes them is the VALUE, and the
// value is what a pinned output is.
//
// `06_localtime.t` is why this runs against a real interpreter rather
// than against a recorded transcript. It pins SHAPE -- nine elements in
// list context, one in scalar -- precisely because the contents depend on
// the clock and the zone, and a shape claim is only worth anything if the
// interpreter that produces it is the one the file names.
func TestTierContextPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierContext, perl)

	for name, f := range tierFiles(t, tierContext) {
		t.Run(name, func(t *testing.T) {
			compiles, output := askPerl(t, f.Source)

			switch {
			case f.ExpectParses && !compiles:
				t.Errorf("file says `expect parses`, perl -c refuses it")
			case f.ExpectParsent && compiles:
				t.Errorf("file says `expect parsent`, perl -c accepts it")
			}
			if f.ExpectOutput != nil && output != *f.ExpectOutput {
				t.Errorf("pinned output %q, perl prints %q", *f.ExpectOutput, output)
			}
		})
	}
}

// TestTierContextLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate.
// This is the tier's own, so a file that reaches forward fails HERE,
// named as this tier's problem rather than as one subtest among a
// hundred and thirteen.
//
// The lint has a specific thing to catch here, and the README names it:
// boolean context is this tier's subject but `if (@a)` emits `and` and
// `cond_expr`, which are tier 06's. A boolean-context file added to this
// tier would satisfy its own output pin and fail here, which is the lint
// doing exactly what it is for -- and is why no such file is in the tier.
func TestTierContextLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierContext) {
		t.Run(name, func(t *testing.T) {
			if err := lintFile(t, f, tierContext, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// contextConstructs are the spellings this tier's construct files
// introduce, keyed by the file identity that owns each.
//
// A TABLE here rather than a spelling extracted from each file's source,
// which is where this differs from tier 01. Tier 01's files each bind one
// literal and the literal IS the construct, so a regex over the source
// recovers it. This tier's constructs are CONTEXTS, and a context has no
// spelling at all: `my $count = @a` and `my @b = @a` differ in the sigil
// of the target, not in any text a search could name. So the construct is
// named by the file that owns it -- the spec makes a file's name its
// identity -- and the spelling below is what that construct is written
// as when it appears somewhere else.
//
// Each spelling is written with enough surrounding syntax to pin it. A
// bare `reverse` would be satisfied by the word inside a comment; `=
// reverse @` requires the call in an assignment on an array.
var contextConstructs = map[string]string{
	"scalar_of_array":         "= @a;",
	"interpolated_array":      `= "@a";`,
	"comma_in_scalar_context": "= (4, 5, 6);",
	"reverse":                 "reverse @a",
	"sort":                    "sort @a",
	"localtime":               "localtime",
	"wantarray":               "wantarray",

	// `caller` (issue 01a0c730). Spelled `= caller;` and not bare
	// `caller`, because the bare word is satisfied by the word in a
	// comment -- and this tier's own README now names the op in prose.
	// The assignment form is how both halves of the pair are written in
	// `08_caller.t` and in the adjacency file alike.
	"caller": "= caller;",

	// `map` and `grep` (issue 01a0c730), which belong here rather than
	// with the list operators because each one's RESULT is what the tier
	// is about: `my @b = map ...` yields the mapped list and `my $n =
	// map ...` yields its count, from the same call.
	//
	// Spelled with the brace so the word alone does not satisfy them.
	// Both files discuss their own construct in prose, and `grep` in
	// particular is a word this corpus uses constantly to describe
	// SEARCHING -- a bare `grep` would be matched by a comment saying
	// "grep for it", which is not the construct.
	"map":  "map {",
	"grep": "grep {",

	// `..`'s three readings (issue 01a0cc04-3370), which are three
	// operators wearing one spelling and are separated by CONTEXT --
	// this tier's subject.
	//
	// Each spelling carries enough of its expression to identify WHICH
	// reading it is, because a bare `..` would match all three. The
	// flip-flop is the one that needs it most: its spelling is the
	// array-element operand pair that makes its state observable, and a
	// shorter spelling would be satisfied by either list form.
	"range_list":   "1 .. scalar(",
	"range_string": `.. "bb"`,
	"flipflop":     "$on[$_] .. $off[$_]",
}

// contextConstructFromName returns the construct a file's name identifies, or ""
// when its name identifies none.
func contextConstructFromName(name string) string {
	m := reNumbered.FindStringSubmatch(name)
	if m == nil {
		return ""
	}
	return strings.TrimSuffix(m[2], ".t")
}

// tierContextPrerequisite is the tier 03_context declares a dependency
// on.
//
// Named and then CHECKED against the README rather than read from it and
// trusted, for the reason tier 10 gives for the same constant: the
// pairing this test makes is specific to what 02_variables supplies, so
// a declaration that moved would leave the pairing asserting something
// about a tier it no longer names.
const tierContextPrerequisite = "02_variables"

// reFlaggedOp matches one op in `B::Concise,-exec` output and keeps BOTH
// the name and the flags:
//
//	e  <@> reverse[t4] lK/1
//
// `lintOps` reads the same format through `reConciseOp` and deliberately
// throws the flags away -- it lints op NAMES and a flag is not an op. This
// tier is the one where that discard is the whole problem, so it has its
// own reader rather than a widening of the shared one.
//
// The targ suffix `[t4]` is excluded from the name, and the flag field is
// whatever follows the whitespace after it.
var reFlaggedOp = regexp.MustCompile(`^\s*\S+\s+<[^>]*>\s+(\w+)(?:\[[^\]]*\])?\s+(\S+)`)

// contextOf returns the context a B::Concise flag field records: "s" for
// scalar, "l" for list, or "" for neither.
//
// B::Concise prints the context as the FIRST character of the flag field,
// from the set `v` (void), `s` (scalar), `l` (list) and `?` (unknown),
// followed by the other flags. Measured across this tier: `padav[@a] s`,
// `padav[@a] lRM*/LVINTRO`, `reverse[t4] lK/1`, `aassign[t6] sKS`.
//
// Void is deliberately NOT a context this reports. Every `padsv_store`
// and every statement-level `aassign` carries `v`, and void is what a
// discarded result is in rather than one of the two this tier
// discriminates.
//
// Measured, that exclusion is DEFENSIVE and not load-bearing: the callers
// ask only whether an op was seen with `s` AND with `l`, so admitting `v`
// as a third answer changes no verdict today. It is written this way so
// that a later caller asking a different question -- "in how many
// contexts?" -- inherits the right set rather than a count inflated by
// every statement in the file.
func contextOf(flags string) string {
	if flags == "" {
		return ""
	}
	switch flags[0] {
	case 's', 'l':
		return flags[:1]
	default:
		return ""
	}
}

// concise runs the pinned perl over a source and returns B::Concise's
// execution-order lines.
//
// `opsOf` does the same and then discards everything but the op name,
// which is what this tier cannot afford. Rather than widen `opsOf` -- a
// shared helper four other tier tests depend on -- this keeps its own
// reader and its own parse of the same bytes.
func concise(t *testing.T, source string) ([]string, error) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "case.pl")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		return nil, fmt.Errorf("writing the case: %w", err)
	}

	perl, err := perlPath()
	if err != nil {
		return nil, err
	}

	out, err := exec.Command(perl, "-MO=Concise,-exec", path).Output()
	if err != nil {
		return nil, fmt.Errorf("perl -MO=Concise refused the source: %w", err)
	}
	return strings.Split(string(out), "\n"), nil
}

// opContexts returns, for one source, the set of contexts each op was
// compiled in -- restricted to scalar and list.
func opContexts(t *testing.T, source string) (map[string]map[string]bool, error) {
	t.Helper()

	lines, err := concise(t, source)
	if err != nil {
		return nil, err
	}

	out := map[string]map[string]bool{}
	for _, line := range lines {
		m := reFlaggedOp.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		ctx := contextOf(m[2])
		if ctx == "" {
			continue
		}
		if out[m[1]] == nil {
			out[m[1]] = map[string]bool{}
		}
		out[m[1]][ctx] = true
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no flagged ops in B::Concise output, which cannot be right")
	}
	return out, nil
}

// discriminatingPairs are the ops this tier claims to measure in BOTH
// contexts, and the file that is supposed to measure each.
//
// This is the tier's subject stated as something falsifiable. Every other
// tier can be checked by asking which ops a file emits; this one cannot,
// because its two halves emit the SAME op and differ only by a flag --
// `reverse[t6] lK/1` against `reverse[t4] sK/1`, `aassign ... vKS`
// against `aassign ... sKS`. An op-name check sees one `reverse` and is
// satisfied by a file that only ever puts it in one context.
//
// The ops here are not the tier's INTRODUCES set and must not be
// confused with it. `join` and `list` are introduced and appear once each
// in scalar context only, because neither HAS a list-context form:
// interpolation is a join whatever receives it, and in list context the
// comma is not a `list` op at all. `wantarray` likewise answers about the
// enclosing context rather than being placed in one. Pairs are a property
// of the ops that answer two questions, which is four of them.
//
// `padav` carries the tier's baseline pair and is tier 02's op, not this
// tier's -- which is the point `01_scalar_of_array.t` exists to make.
// Using an op is not introducing it, and the pair is this tier's even
// where the op is not.
var discriminatingPairs = map[string]string{
	"padav":     "01_scalar_of_array.t",
	"reverse":   "04_reverse.t",
	"aassign":   "05_sort.t",
	"localtime": "06_localtime.t",

	// `caller` is the tier's sharpest pair and the last to arrive. At
	// file scope the two contexts return different AMOUNTS -- undef
	// against the empty list -- rather than two formattings of one
	// answer, so the behavioural half falsifies a collapsed context on
	// its own, which no other pair here can claim.
	"caller": "08_caller.t",
}

// describeContexts names the contexts an op was seen in, for an error
// message.
func describeContexts(have map[string]bool) string {
	switch {
	case have["s"]:
		return "scalar context only"
	case have["l"]:
		return "list context only"
	default:
		return "neither scalar nor list context"
	}
}

// TestTierContextRefusalsCited checks that a refusing file names WHICH
// site declines, from the stable inventory, rather than describing it.
//
// The reasoning is tier 01's and transfers exactly, so it is stated
// briefly here and in full there. `parse.RefusalCode` names a site in the
// PARSER. A file refusing purely on a token fact has no parser Unknown to
// name, and naming one anyway makes `run.go` fail it as a stale marker --
// correctly, because the claim would be false. So the rule is
// biconditional rather than "every refusing file names a code":
//
//   - A file whose parser produces Unknowns MUST name one of their codes.
//   - A file whose parser produces none MUST NOT name a code, and must
//     carry the token facts that are its refusal instead.
//
// AS OF TODAY THIS TIER HAS NO REFUSING FILE, so the loop runs over
// nothing. That is a measurement rather than an oversight, and this tier
// makes it a slightly surprising one: every context form here parses,
// including `wantarray` at file scope and the `() =` count idiom. The
// check is still written, because the event it guards is a file
// ACQUIRING a refusal -- the next parser change that declines the count
// idiom lands a `# STATUS refuses` line here, and this is what then
// requires it to cite a code rather than a sentence.
func TestTierContextRefusalsCited(t *testing.T) {
	for name, f := range tierFiles(t, tierContext) {
		if f.Refuses == "" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			codes := refusalCodes(parse.Parse([]byte(f.Source)))

			if len(codes) == 0 {
				if f.RefusalCode != "" {
					t.Errorf("names Refusal %s, but the parser returns no "+
						"Unknown at all.\n\tThere is no site to name; the "+
						"refusal is lexical.", f.RefusalCode)
				}
				if len(f.TokenFacts) == 0 {
					t.Errorf("refuses (%s) with no parser Unknown and no "+
						"`--- expect tokens` section.\n\tNothing in this "+
						"file measures the refusal it documents.", f.Refuses)
				}
				return
			}

			if f.RefusalCode == "" {
				t.Errorf("refuses (%s) with codes %s but names none.\n"+
					"\tAdd a `Refusal <code>.` clause to the STATUS line -- "+
					"a code is a stable identifier where a message is prose.",
					f.Refuses, joinCodes(codes))
				return
			}
			if _, ok := parse.RefusalSites[f.RefusalCode]; !ok {
				t.Errorf("names Refusal %s, which is not in parse.RefusalSites.\n"+
					"\tThe inventory is the vocabulary; a code outside it is a "+
					"message wearing a code's spelling.", f.RefusalCode)
			}
			if !hasCode(codes, f.RefusalCode) {
				t.Errorf("names Refusal %s, but the parser refuses with %s",
					f.RefusalCode, joinCodes(codes))
			}
		})
	}
}
