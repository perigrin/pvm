// ABOUTME: Tier 09 checked against the finished tooling, with delimiters as the tier's substance.
// ABOUTME: Five checks the tier's own issue names, each tier-specific rather than corpus-wide.
package conformance

import (
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// tierRegex is the tier this file is about.
//
// A constant rather than a literal at five call sites, for the reason
// tier01_test.go gives: the tier number is a POSITION and positions move.
const tierRegex = "09_regex"

// tierRegexDependsOn is the prerequisite 09_regex/README.md DECLARES.
//
// Written down here so the adjacency test can check that the README still
// says it, and fail loudly rather than silently re-targeting if it
// changes.
//
// The spec names 09_regex as one of the two tiers whose DEPENDS ON is
// deliberately NOT N-1. Nothing in a regex needs a reference: a pattern
// is not a reference, a captured group is not a reference, and `qr//` is
// claimed here only in its non-recursive form. Pairing this tier's
// constructs with 08_references in the adjacency file would assert a
// dependency that does not exist, which is the failure mode the spec
// warns about -- so the pairing target is read from the declaration and
// never computed as N-1.
const tierRegexDependsOn = "01_literals"

// TestTierRegexPerlValidated runs every file in the tier through the
// pinned interpreter before it counts.
//
// The same two claims tier 01 checks -- perl AGREES with the file's
// parses/parsent bit, and perl PRINTS what the file pins -- but what a
// wrong interpreter would cost is different here. Tier 01 pins literals,
// where perl's answers are stable across decades. This tier pins the
// STRINGIFICATION of a `qr//` object, which is `(?^:abc)` and not
// `(?^u:abc)`: the `u` appears under a unicode_strings-like feature
// bundle, so the pinned bytes are a property of 5.42.0 run without one
// rather than a property of Perl. Measured against whatever perl PATH
// happened to offer, that pin would be wrong silently.
func TestTierRegexPerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierRegex, perl)

	for name, f := range tierFiles(t, tierRegex) {
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

// TestTierRegexLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate.
// This is the tier's own, and it exists because the tier issue's
// acceptance is checkable on its own: a file reaching forward must fail
// HERE, named as this tier's problem, rather than as one subtest among a
// hundred and thirteen.
//
// What the lint proves is ONE DIRECTION only: no file uses an op that no
// tier at or before 09 introduces. It cannot prove the tier's declared
// set is earned, and in this tier that gap is not academic -- see
// TestTierRegexOpsAreEarned.
func TestTierRegexLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierRegex) {
		t.Run(name, func(t *testing.T) {
			if err := lintFile(t, f, tierRegex, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// TestTierRegexOpsAreEarned checks that every op the tier CLAIMS is one
// some file in it actually emits.
//
// The lint runs one way: it stops a file using an op no tier at or before
// this one introduces. Nothing stops the reverse -- a README claiming an
// op no file produces. That direction is invisible by construction,
// because a wider allowance never fails a lint; it only lets a later
// tier's file pass a check it should have failed.
//
// This tier is where it bites. `regcomp` is claimed on the strength of
// ONE file, `05_interpolated.t`, whose source differs from the baseline
// by a single character -- `/abc/` becomes `/$p/` -- and the whole op
// appears. Delete that file and the claim survives with nothing behind
// it, quietly granting `regcomp` to every tier from 09 onward.
func TestTierRegexOpsAreEarned(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	claimed := tiers[tierRegex]
	if len(claimed) == 0 {
		t.Fatalf("%s claims no ops at all", tierRegex)
	}

	emitted := map[string]bool{}
	for name, f := range tierFiles(t, tierRegex) {
		ops, err := opsOf(t, f.Source)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, op := range ops {
			emitted[op] = true
		}
	}

	for _, op := range claimed {
		if !emitted[op] {
			t.Errorf("%s/README.md claims %s, which no file in the tier emits.\n"+
				"\tAn unearned op widens every later tier's allowance, and no "+
				"lint can see it: widening never fails.", tierRegex, op)
		}
	}
}

// TestTierRegexCoversItsDelimiters checks the tier covers the delimiter
// forms it is scoped to, rather than the ones it happened to think of.
//
// DELIMITERS ARE THE TIER'S SUBSTANCE and its issue says so: "`m//`,
// `m{}`, `m##`, and the rule that a paired delimiter nests while an
// unpaired one does not". A tier whose whole distinguishing claim is
// about delimiters, covering whichever delimiters got written, is a tier
// whose coverage is an accident -- and nothing else in the suite can
// notice, because the optree cannot see a delimiter and the lint reads
// ops.
//
// `m##` is the entry that makes this worth a test rather than trust. `#`
// is PERL'S COMMENT CHARACTER, so `m#abc#` is the one delimiter form
// where a lexer reading left to right has already made a decision --
// "this is a comment, discard to end of line" -- before it can know it
// was wrong. Measured under 5.42.0 the op stream is byte-identical to
// `/abc/`'s, so neither the optree nor the printed output would report
// the mistake; a token fact is the only place the claim can live.
//
// The forms below are the issue's own list plus the non-bracketing case
// the nesting rule needs, not a fresh opinion about what matters. Each
// entry names a DELIMITER FORM, in `delimiterForms`' spelling, and a form
// nowhere in the tier is a delimiter the corpus makes no claim about.
func TestTierRegexCoversItsDelimiters(t *testing.T) {
	// Why each form is here, since a list is otherwise unfalsifiable:
	//
	//   m//    the baseline, and the match-vs-divide ambiguity
	//   m{}    a bracketing delimiter, which nests
	//   m!!    a non-bracketing delimiter, which does not
	//   m##    the comment character as a delimiter
	//   s///   substitution's baseline, two pairs sharing one character
	//   s{}{}  two bracketing pairs, where the second may differ
	//   qr//   the non-recursive compiled pattern this tier claims
	required := []struct{ form, why string }{
		{"m//", "the baseline match, and the match-vs-divide ambiguity"},
		{"m{}", "a bracketing delimiter, which nests"},
		{"m!!", "a non-bracketing delimiter, which does not nest"},
		{"m##", "the comment character used as a delimiter"},
		{"s///", "substitution's baseline, both pairs sharing one character"},
		{"s{}{}", "two bracketing pairs, where the second's opener may differ"},
		{"qr//", "the non-recursive compiled pattern this tier claims"},
	}

	// The union across the tier, adjacency file included: the adjacency
	// file is where several forms sit together, and a form it alone
	// carries is still a form the tier covers.
	var covered []string
	for _, f := range tierFiles(t, tierRegex) {
		for _, form := range delimiterForms(f) {
			if !slices.Contains(covered, form) {
				covered = append(covered, form)
			}
		}
	}
	// `m//` has no positive token fact anywhere and cannot: GLOSSARY.md
	// defines `quote-like operator` as a quote spelled with an OPERATOR
	// NAME, and a bare `/abc/` is spelled with delimiters alone, so there
	// is no category to assert it under. Its coverage is therefore read
	// from the source, where it is `=~ /` -- the binding followed by the
	// slash, which distinguishes a pattern from a division.
	for _, f := range tierFiles(t, tierRegex) {
		if strings.Contains(f.Source, "=~ /") {
			covered = append(covered, "m//")
			break
		}
	}
	sort.Strings(covered)

	for _, r := range required {
		if !slices.Contains(covered, r.form) {
			t.Errorf("no file in %s covers the delimiter form %s (%s)\n"+
				"\tThe tier is scoped to delimiters, so a form it does not "+
				"spell is a form it asserts nothing about -- and the optree "+
				"cannot notice, because delimiters erase themselves.\n"+
				"\tcovered: %s",
				tierRegex, r.form, r.why, strings.Join(covered, " "))
		}
	}

	// THE NESTING RULE, which is the second half of what the issue scopes
	// this tier to and is not a delimiter FORM at all. `m{}` appearing
	// somewhere says a bracketing delimiter was used; it does not say the
	// brackets NEST, and every `m{abc}` and `s{a}{z}` in the tier has a
	// body with no bracket in it, so all of them are satisfied by a lexer
	// that stops at the first `}`.
	//
	// The rule is asymmetric and both halves are measured. `m{a{b}c}`
	// matches the five-character string `a{b}c` and NOT `aXc`, so the
	// inner braces were content rather than a terminator; `m!a!b!c!` is a
	// syntax error, because a non-bracketing delimiter has no depth to
	// count. A pattern containing its own opening delimiter is the only
	// source text that can tell those apart.
	var nests bool
	for _, f := range tierFiles(t, tierRegex) {
		if reNestedDelimiter.MatchString(f.Source) {
			nests = true
			break
		}
	}
	if !nests {
		t.Errorf("no file in %s spells a pattern containing its own opening "+
			"delimiter, such as `m{a{b}c}`.\n"+
			"\tThe tier's issue scopes it to \"the rule that a paired "+
			"delimiter nests while an unpaired one does not\", and every "+
			"bracketing pattern here has a body with no bracket in it -- so "+
			"a lexer that stops at the first `}` passes the whole tier.",
			tierRegex)
	}
}

// reNestedDelimiter matches a bracketing quote-like operator whose
// pattern contains the opening delimiter again:
//
//	m{a{b}c}
//	s{a{b}c}{ok}
//
// Only `{` is matched, because that is the bracket this tier's files use
// and a pattern is where the check has to look. Written against the
// SOURCE rather than the token facts, unlike the forms above: a nested
// delimiter is a property of the pattern CONTENT, which `delimiterForms`
// deliberately discards.
var reNestedDelimiter = regexp.MustCompile(`\b[a-z]*\{[^{}]*\{[^{}]*\}`)

// TestTierRegexAdjacency checks the tier's adjacency file holds every
// construct the tier introduces, each next to another, paired with the
// DECLARED prerequisite.
//
// WHY OPS CANNOT DO THIS HERE, and it is the tier's defining fact.
// `m{abc}` and `/abc/` emit BYTE-IDENTICAL op streams -- measured under
// 5.42.0, identical down to the pattern text perl prints inside the op,
// `match(/"abc"/) sKS` -- as do `s{a}{z}` and `s/a/z/`, and `m#abc#`. The
// optree cannot see a delimiter at all. An adjacency check asking "does
// the adjacency file emit the tier's ops" therefore goes green against a
// file holding five copies of `/abc/`, which makes no adjacency claim
// about delimiters whatsoever.
//
// So the claim is checked on the DELIMITER FORM: the operator name and
// the delimiter characters, with the pattern between them ignored. That
// is the right abstraction rather than a convenience. `m{zzz}` in the
// adjacency file and `m{abc}` in the construct file are the SAME
// construct -- a match through a bracketing delimiter -- and an adjacency
// file forced to reuse the construct file's exact pattern would be
// pinning prose. What must carry over is the delimiter, because the
// delimiter is what a lexer gets wrong.
//
// The forms come from the construct files' own `--- expect tokens`
// claims, not from a list here. That is not laziness, it is the tier's
// own rule turned into a mechanism: a delimiter file MUST carry token
// claims or it measures nothing the plain form does not, so every
// delimiter this tier covers is already written down in a fact line. A
// second enumeration beside them is how this package has drifted four
// times.
//
// What this ESTABLISHES: every delimiter form the tier teaches appears in
// one compiled body, so a lexer that delimits each alone and mis-delimits
// the pair is reachable from this corpus. That is the bug a
// one-construct-per-file corpus cannot see, and in a tier whose operand
// is not Perl it is the bug to expect: a lexer scanning for the next `/`
// is correct on every file here taken alone and wrong the moment a
// `s{a}{z}` sits between two matches.
//
// What it does NOT establish: that the constructs are adjacent in any
// stronger sense than "in the same body". Source position would pin a
// layout the file is free to change.
func TestTierRegexAdjacency(t *testing.T) {
	deps, err := readTierDeps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	if got := deps[tierRegex]; got != tierRegexDependsOn {
		t.Fatalf("%s declares DEPENDS ON %q, not %q.\n"+
			"\tThe pairing is with the DECLARED prerequisite; pairing with "+
			"tier N-1 here would assert a dependency on 08_references that "+
			"nothing in a regex has.", tierRegex, got, tierRegexDependsOn)
	}

	files := tierFiles(t, tierRegex)
	adj, ok := files[adjacencyFile]
	if !ok {
		t.Fatalf("%s has no %s", tierRegex, adjacencyFile)
	}
	adjForms := delimiterForms(adj)

	for name, f := range files {
		if name == adjacencyFile {
			continue
		}
		forms := delimiterForms(f)
		if len(forms) == 0 {
			// Not every construct file has a delimiter form to give:
			// `01_bare_match.t` asserts a bare `/abc/`, which GLOSSARY.md
			// gives no positive category -- a bare slash is not spelled
			// with an operator name -- so its facts are negative and it
			// contributes nothing here. That is a property of the
			// glossary, not a defect in the file, so it is skipped rather
			// than reported.
			continue
		}
		for _, form := range forms {
			if !slices.Contains(adjForms, form) {
				t.Errorf("%s introduces the delimiter form %s, which %s does not use.\n"+
					"\tThe adjacency file must hold every delimiter form the tier "+
					"introduces, or the lexer bug it exists to reach -- one "+
					"delimiter form mis-terminating the next -- is unreachable "+
					"for that form.\n\tadjacency file has: %s",
					name, form, adjacencyFile, strings.Join(adjForms, " "))
			}
		}
	}

	// The pairing with the declared prerequisite, EXERCISED rather than
	// named. 01_literals is what a pattern gets its operand from: every
	// match in this tier runs against a string literal bound to a pad
	// slot. A README naming a prerequisite whose construct appears
	// nowhere in the adjacency file asserts nothing, which is exactly
	// what a pairing with 08_references would have been.
	if !strings.Contains(adj.Source, `= "`) {
		t.Errorf("%s pairs with %s, but binds no string literal for a pattern "+
			"to match against.\n\tA prerequisite that is named and not used "+
			"is the assertion the spec warns about.",
			adjacencyFile, tierRegexDependsOn)
	}

	// One body, not several. Adjacency needs at least two constructs to
	// be adjacent TO, and `nextstate` is one per statement -- which is
	// how this counts statements without a parser of its own.
	ops, err := opsOf(t, adj.Source)
	if err != nil {
		t.Fatalf("%s: %v", adjacencyFile, err)
	}
	if n := countOp(ops, "nextstate"); n < 2 {
		t.Errorf("%s compiles to %d statement(s); adjacency needs at least 2", adjacencyFile, n)
	}
}

// rePositiveQuoteFact matches a corpus file's positive quote-like claim:
//
//	one quote-like operator whose text is "s{a}{z}"
//
// Positive only. A `no ... whose text is` line asserts an ABSENCE, and an
// absence names no construct the file introduces -- `01_bare_match.t`
// asserting `no operator whose text is "/"` is saying what a mis-lex
// would produce, not what the file covers.
var rePositiveQuoteFact = regexp.MustCompile(`^one quote-like operator whose text is (".*")$`)

// reDelimiterForm splits a quote-like operator's text into its operator
// name and everything after it.
//
// `m`, `s`, `qr`, `qq`, `tr` and the rest are letters; whatever follows
// the letters is delimiter and content. Anchored at the start because an
// operator name is a prefix -- `s{a}{z}` is `s` then `{a}{z}` -- and a
// search anywhere would find the `z`.
var reDelimiterForm = regexp.MustCompile(`^([a-z]+)(.*)$`)

// delimiterForms returns the delimiter forms a file's token facts claim,
// sorted and deduplicated so a failure message is the same bytes twice.
//
// A FORM is the operator name plus the delimiter characters with the
// pattern content removed: `m{abc}` and `m{zzz}` are both `m{}`, while
// `m!abc!` is `m!!` and `s{a}{z}` is `s{}{}`. That is the quantity this
// tier is about -- the optree cannot distinguish any of them, and a lexer
// that gets one right may get the next wrong -- and it is deliberately
// insensitive to the pattern, so an adjacency file is free to match
// something other than what a construct file matched.
//
// Content removal is positional rather than parsed: the form keeps every
// occurrence of the delimiter character and of its closing partner, and
// drops everything else. Nesting is not modelled, because no file here
// nests -- and a file that did would show up as an unfamiliar form
// rather than passing quietly.
func delimiterForms(f *File) []string {
	var out []string
	for _, fact := range f.TokenFacts {
		m := rePositiveQuoteFact.FindStringSubmatch(strings.TrimSpace(fact))
		if m == nil {
			continue
		}
		text, err := strconv.Unquote(m[1])
		if err != nil {
			continue
		}
		parts := reDelimiterForm.FindStringSubmatch(text)
		if parts == nil {
			continue
		}
		name, rest := parts[1], parts[2]
		if rest == "" {
			continue
		}

		// The delimiter is the first character after the name; the form
		// keeps every occurrence of it and of its closing partner, and
		// drops everything else. `s{a}{z}` keeps `{`, `}`, `{`, `}`.
		open := rune(rest[0])
		closer := closingDelimiter(open)
		var form strings.Builder
		form.WriteString(name)
		for _, r := range rest {
			if r == open || r == closer {
				form.WriteRune(r)
			}
		}
		if s := form.String(); !slices.Contains(out, s) {
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// closingDelimiter returns the character that closes a bracketing
// delimiter, or the delimiter itself when it does not bracket.
//
// GLOSSARY.md records the asymmetry this encodes: bracketing delimiters
// nest and take a distinct closer, non-bracketing ones do not nest and
// close with themselves. Measured, `q{a{b}c}` is the five-character
// string `a{b}c` while `q!a!b!c!` is a syntax error.
func closingDelimiter(open rune) rune {
	switch open {
	case '{':
		return '}'
	case '(':
		return ')'
	case '[':
		return ']'
	case '<':
		return '>'
	}
	return open
}

// TestTierRegexRefusalsCited checks that a refusing file names WHICH site
// declines, from the stable inventory, rather than describing it.
//
// A message is prose: reword it and a file that quoted it breaks over a
// change that moved nothing. A code changes only when the reason the
// parser declines changes, which is the event a refusing file wants to be
// told about.
//
// The rule is BICONDITIONAL rather than "every refusing file names a
// code", and this tier is full of the case that forces it.
// `parse.RefusalCode` names a site in the PARSER, so a file refusing on a
// purely LEXICAL fact has no Unknown to name -- and naming one anyway
// makes `run.go` fail the file as a stale marker, correctly, because the
// claim would be false. A delimiter file is exactly that shape: its whole
// measurement is its token claims, and a lexer that normalised `m{abc}`
// to `/abc/` would fail it with a perfectly happy parser behind it.
//
//   - A file whose parser produces Unknowns MUST name one of their
//     codes. It has a site to name and prose would be the alternative.
//   - A file whose parser produces none MUST NOT name a code, and must
//     carry the token facts that are its refusal instead. A lexical
//     refusal's citation is the assertion itself.
//
// Both halves are needed. Without the first a parse refusal may stay
// uncited; without the second a lexical refusal may cite a parse site it
// does not have, which is the false claim the codes exist to forbid.
func TestTierRegexRefusalsCited(t *testing.T) {
	for name, f := range tierFiles(t, tierRegex) {
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
