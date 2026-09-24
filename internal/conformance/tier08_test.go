// ABOUTME: Tier 08 checked against the finished tooling: references, deref, arrow, anonymous constructors.
// ABOUTME: Five checks the tier's issue names, plus the spelling split the op stream cannot see.
package conformance

import (
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// tierReferences is the tier this file is about.
//
// A constant rather than a literal at six call sites, for the reason
// tier01_test.go gives for tierLiterals: the tier number is a POSITION
// and positions move.
const tierReferences = "08_references"

// TestTierReferencesPerlValidated runs every file in the tier through the
// pinned interpreter before it counts.
//
// Identical in intent to TestTierLiteralsPerlValidated, and this tier has
// its own reason to want it. A REFERENCE STRINGIFIES AS `ARRAY(0x...)`,
// whose address changes every run, so no file here may print one: every
// pinned output in this tier is a REFERENT or a `ref` category, reached
// through a dereference. That substitution is only sound if perl actually
// produces the pinned bytes from it, and this is where that is
// established rather than asserted.
//
// The optimiser gives the second reason. `my $r = \1` folds to
// `const[IV \1] s/FOLD` and emits no `srefgen` at all, so several files
// here carry a named-variable fixture whose only job is to defeat the
// fold. A fixture that stopped working would leave the file measuring
// nothing, and perl is the only thing that can say whether it still does.
func TestTierReferencesPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierReferences, perl)

	for name, f := range tierFiles(t, tierReferences) {
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

// TestTierReferencesLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate.
// This is the tier's own, so a file that reaches forward fails HERE,
// named as this tier's problem rather than as one subtest among a
// hundred and thirteen.
//
// The lint has a specific thing to catch in this tier, and the README
// names it: the anonymous HASH constructor. `{}` and `{ %a }` compile to
// `emptyavhv` and `anonhash`, which tier 11 claims where objects are
// built. A file here written as `my $r = { a => 1 }` -- the obvious
// companion to `03_anonymous_array.t`, and the one a reader keeps
// expecting -- would satisfy its own output pin and fail here, which is
// the lint doing exactly what it is for. `07_deref_brace_hash.t` reaches
// the same place through `\%h` instead, which is this tier's own op.
func TestTierReferencesLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierReferences) {
		t.Run(name, func(t *testing.T) {
			if err := lintFile(t, f, tierReferences, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// referenceConstructs are the spellings this tier's construct files
// introduce, keyed by the file identity that owns each.
//
// THE NAME, not a pattern over the source, for tier 05's reason: every
// file in this tier contains a backslash or a `$r`, because a reference
// needs one to exist before it can be dereferenced. Reading each file's
// subject from its source would report `\` nine times and the tier would
// look like one construct.
//
// The value is a SPELLING that must appear in the file named for it and
// in the adjacency file, which is what makes a substring search mean
// "this construct appears there" rather than "these characters do". Each
// is written with enough surrounding syntax to pin it: `->[` rather than
// `->`, since the arrow also spells a method call and a code call.
var referenceConstructs = map[string]string{
	"backslash_scalar": `= \`,
	"backslash_list":   `= \(`,
	"anonymous_array":  `= [`,
	"arrow_deref":      `->[`,
	"brace_deref":      `${$`,
	"deref_at":         `@{$`,
	"deref_at_sigil":   `@$`,
	"deref_postfix":    `->@*`,
	"deref_brace_hash": `%{$`,
	"code_ref":         `\&`,
	"ref_builtin":      `ref(`,

	// `prototype` (issue 01a0c730). Spelled with the operator attached
	// rather than as the bare word, for the reason the comment above
	// gives: the bare word appears in this tier's own README prose and
	// in the file's header, and a substring search would be satisfied by
	// either. `prototype \&` is the construct.
	"prototype_builtin": `prototype \&`,
}

// constructFromName returns the construct spelling a file is named for,
// or "" when its name names none.
func constructFromName(name string) string {
	m := reNumbered.FindStringSubmatch(name)
	if m == nil {
		return ""
	}
	return referenceConstructs[strings.TrimSuffix(m[2], ".t")]
}

// TestTierReferencesRefusalsCited checks that a refusing file names WHICH
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
// AS OF TODAY THIS TIER HAS NO REFUSING FILE, so the check is vacuous.
// That is a measurement rather than an oversight, and a surprising one:
// the tier's README records a KNOWN defect in the tree -- a brace before
// `->` is a hashref and two anonhash cases are misread, M1 issue
// 01a0bae2 -- yet no file here trips it, because the anonymous hash
// constructor belongs to tier 11 and this tier's braces are all
// dereferences. The check is still written, because the event it guards
// is a file ACQUIRING a refusal: the next parser change that declines
// `$r->@*` or `${$r}[0]` lands a `# STATUS refuses` line here, and this
// is what then requires it to cite a code rather than a sentence.
func TestTierReferencesRefusalsCited(t *testing.T) {
	for name, f := range tierFiles(t, tierReferences) {
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

// The arrow form is the one entry with no `no variable` claim of its
// own, because there is nothing for a lexer to swallow: `$r->[2]` is a
// plain variable, an arrow and a subscript, and what a lexer can get
// wrong is the ARROW rather than the grouping. Its claim is the counted
// arrow, which `05_brace_deref.t` mirrors as `no operator whose text is
// "->"` -- the pair is the distinction, and neither half alone is.
var derefSpellings = []struct {
	spelling string // as written in a corpus file's source
	fact     string // the token fact that constitutes the claim about it
	why      string // what a lexer that got it wrong would do
}{
	{`->[`, `one operator whose text is "->"`, "read the arrow as part of a variable name, or drop it"},
	{`${$`, `no variable whose text is "${$r}"`, "swallow ${$r} whole as one variable token"},
	{`@{$`, `no variable whose text is "@{$r}"`, "swallow @{$r} whole as one variable token"},
	{`@$`, `no variable whose text is "@$r"`, "swallow @$r whole as one variable token"},
	{`->@*`, `no variable whose text is "$r->@*"`, "read $r->@* as a variable followed by a glob"},
}

// TestTierReferencesSpellingsAreDistinguished checks that every deref
// spelling the op stream cannot separate is separated by a token fact.
//
// THE HOLE THIS TIER IS SHAPED TO HAVE, and the README diagnoses it
// precisely without checking it. Two independent collapses meet here:
//
//   - `$r->[0]` and `${$r}[0]` differ by op COUNT and a private flag,
//     never by op NAME. A lint comparing name sets sees them as one file.
//   - `@{$r}`, `@$r` and `$r->@*` are byte-identical in the op stream,
//     all three being `rv2av`.
//
// The README's conclusion is that `expect tokens` is "the one check that
// reads the spelling rather than the result". A tier that reached that
// conclusion and then asserted nothing lexical would be the shape tiers
// 09 and 10 were found in: behaviour pinned, spelling unmeasured, and a
// lexer that mis-split the construct passing everything silently.
//
// Measured, that is not hypothetical here either. Our lexer emits
// `DerefSigil("@") Variable("$r")` for `@$r` -- two tokens -- and a lexer
// that emitted `Variable("@$r")` instead would still satisfy every
// behavioural pin in this tier, because perl's output is the same and
// the ops are the same. `no variable whose text is "@$r"` is the one
// assertion that fails.
//
// The check is on the tier's TOKEN FACTS rather than on its sources,
// because a source containing the spelling asserts nothing about it --
// `06_deref_at.t` contains `@{$r}` INSIDE A DOUBLE-QUOTED STRING, where
// our lexer produces a single `Quote` token and the deref is not
// separately tokenised at all. That file's facts are satisfied entirely
// by the `\@a` on the line above it, which is how a claim about the
// construct can be absent from a file the construct appears in.
func TestTierReferencesSpellingsAreDistinguished(t *testing.T) {
	files := tierFiles(t, tierReferences)

	// Every token fact the tier makes, with the file that makes it.
	type claim struct{ file, fact string }
	var claims []claim
	for name, f := range files {
		for _, fact := range f.TokenFacts {
			claims = append(claims, claim{name, fact})
		}
	}

	for _, s := range derefSpellings {
		var made []string
		for _, c := range claims {
			if c.fact != s.fact {
				continue
			}
			// The file must actually CONTAIN the spelling, or the fact
			// is about a construct that file does not have -- a `no
			// variable` claim about an absent spelling is vacuously true
			// and measures nothing. This is the half `06_deref_at.t`
			// failed before the bare `my @c = @{$r};` was added to it:
			// the construct was present only inside a string, where our
			// lexer emits one `Quote` token and never reaches it.
			if strings.Contains(files[c.file].Source, s.spelling) {
				made = append(made, c.file+": "+c.fact)
			}
		}
		if len(made) == 0 {
			t.Errorf("no file in %s asserts a token fact distinguishing %q.\n"+
				"\tThe README measures this spelling as INDISTINGUISHABLE in "+
				"the op stream and says the files must therefore separate it "+
				"in `expect tokens`.\n"+
				"\tA lexer that would %s passes every behavioural pin in this "+
				"tier, because perl's output and perl's ops are unchanged.",
				tierReferences, s.spelling, s.why)
			continue
		}
		slices.Sort(made)
		t.Logf("%q distinguished by:\n\t%s", s.spelling, strings.Join(made, "\n\t"))
	}
}
