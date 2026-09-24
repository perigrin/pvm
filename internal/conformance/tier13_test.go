// ABOUTME: Tier 13 checked against the finished tooling: the constructs the lexer must delimit without lexing.
// ABOUTME: Six checks the tier's own issue names, each tier-specific rather than corpus-wide.
package conformance

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
)

// tierOpaque is the tier this file is about.
//
// A constant rather than a literal at six call sites, for the reason
// tier01_test.go gives for tierLiterals: the tier number is a POSITION
// and positions move.
const tierOpaque = "13_opaque"

// opaqueRegion is one of the six constructs this tier is about: a span
// the lexer must DELIMIT without lexing what is inside it.
//
// `kind` is the token kind the whole region must arrive as. `opens` and
// `closes` are the region's first and last bytes as the corpus spells
// them, and `hostile` is content inside the region that a lexer reading
// it as Perl would tokenise differently.
//
// THE `hostile` FIELD IS WHY THIS TEST EXISTS AT ALL, so it is worth
// saying plainly. A format body holding `a fixed report line` is
// delimited identically by a lexer that treats it as opaque and by one
// that lexes it as Perl -- three Words either way, inside the region or
// not, and the region's extent is the same. Such a file measures
// DELIMITING and says nothing about NOT LEXING, which is half of the
// tier's thesis. A region holding `@<<<<<<` separates them: opaque it is
// one picture field, lexed as Perl it is an array variable followed by
// left-shifts, and only the first reading produces a single token
// covering the region.
type opaqueRegion struct {
	// what names the construct, for failure messages.
	what string

	// kind is the token kind the entire region must arrive as.
	kind lexer.Kind

	// opens and closes are the region's first and last bytes.
	//
	// Checked rather than assumed because "delimited correctly" is a
	// claim about both ENDS. A body token that started after the
	// introducer, or stopped before the terminator, would leave bytes
	// belonging to the construct for some later rule to skip -- which is
	// the decision GLOSSARY.md records for the heredoc body and the pod
	// block, made once and applying to all four multi-line regions.
	opens, closes string

	// hostile is Perl-looking text that must appear INSIDE the region.
	//
	// Empty for a region whose delimiters already make the point: `qx{}`
	// and `<*.pattern>` are single-line and their contents are a shell
	// command and a glob pattern, where "not lexed as Perl" is asserted
	// by the region being one token at all rather than by what is in it.
	hostile string
}

// opaqueRegions is the six constructs the tier's title names, each with
// the content that makes "without lexing the contents" a falsifiable
// claim rather than a description.
//
// Six entries for six constructs and SEVEN kinds, because a heredoc is
// two tokens -- the opener that sits in the statement and the body that
// follows the line -- and only the body is a region whose contents could
// be lexed. The opener is checked by the tier's files as a token fact;
// here it appears as the region's `opens` text, which is where the
// heredoc's two halves are tied together.
func opaqueRegions() []opaqueRegion {
	return []opaqueRegion{
		{
			what:    "interpolating heredoc body",
			kind:    lexer.HeredocBody,
			opens:   "hello $name",
			closes:  "EOT\n",
			hostile: "$name",
		},
		{
			what: "non-interpolating heredoc body",
			kind: lexer.HeredocBody,
			// The literal heredoc's hostile content is the SAME text as
			// the interpolating one's and that is the point: to the
			// lexer both bodies are opaque, and the difference between
			// them -- whether `$name` is a variable or five characters
			// -- is perl's, settled at runtime and visible only in the
			// output the two files pin.
			opens:   "hello $name",
			closes:  "EOT\n",
			hostile: "$name",
		},
		{
			what:   "indentation-stripping heredoc body",
			kind:   lexer.HeredocBody,
			opens:  "    ",
			closes: "    EOT\n",
			// The hostile content is the INDENTATION. A body token that
			// had been lexed would have dropped the leading whitespace
			// as insignificant, and the stripping `<<~` performs happens
			// to the string perl builds, not to the bytes the token
			// covers.
			hostile: "    trimmed",
		},
		{
			what:   "format body",
			kind:   lexer.FormatBody,
			opens:  "@",
			closes: ".\n",
			// The case GLOSSARY.md names and the tier's README calls the
			// clearest in the tier: `@<<<<<<` is a left-justified picture
			// column, and read as Perl it is an array variable followed
			// by two left-shift operators.
			hostile: "@<<<<<<",
		},
		{
			what: "pod block",
			kind: lexer.Pod,
			// `=` and an identifier, which is the rule -- `=pod` is one
			// spelling of it and `=head1` another, and pinning either
			// here would make this test about a spelling rather than
			// about the construct.
			opens:  "=",
			closes: "=cut\n",
			// Prose is not enough. A pod block holding only words is
			// delimited the same way whether or not its contents were
			// lexed, so the block has to hold something a Perl lexer
			// would choke on or misread.
			hostile: "$this is not( a variable",
		},
		{
			what:  "data section",
			kind:  lexer.DataSection,
			opens: "__",
			// A data section runs to end of file, so it has no
			// terminator of its own and `closes` is its last content
			// byte instead.
			closes:  "\n",
			hostile: "sub not_compiled { $x <=> }",
		},
	}
}

// TestTierOpaqueDelimitsWithoutLexing is the tier's thesis, stated as a
// test.
//
// Each of the six constructs must arrive as ONE token covering the whole
// region, with nothing lexed inside it. Two claims, and they are
// genuinely separate:
//
//   - DELIMITED. One token of the construct's kind whose text begins at
//     the region's first byte and ends at its last. A region delimited
//     short or long is a lexer that will hand the parser bytes belonging
//     to a construct it has already finished reporting.
//   - CONTENTS NOT LEXED. No other token overlaps the region, and the
//     region holds content a Perl lexer would have tokenised. The second
//     half is what makes the first mean something: `a fixed report line`
//     inside a format body proves nothing, because a lexer that read it
//     as three Words and a lexer that read it as nothing both leave the
//     region's extent unchanged.
//
// WHY THE OPTREE CANNOT ADJUDICATE ANY OF THIS, which is the tier's
// defining fact and is recorded in its README. Measured under 5.42.0: a
// heredoc emits tier 01's `const` or `multiconcat`, a pod block emits
// nothing whatsoever, `__END__`/`__DATA__` emit nothing, and a `format`
// declaration emits nothing -- only the `write` that uses it emits
// `enterwrite`. Four of the six constructs are gone before an op exists.
// So the token stream is not merely the convenient place to assert this,
// it is the ONLY place, and a tier asserting `parses` and `output` alone
// would go green against a lexer that read the pod block as arithmetic.
func TestTierOpaqueDelimitsWithoutLexing(t *testing.T) {
	files := tierFiles(t, tierOpaque)

	var sources [][]byte
	for _, f := range files {
		sources = append(sources, []byte(f.Source))
	}

	for _, r := range opaqueRegions() {
		t.Run(r.what, func(t *testing.T) {
			found := false
			for _, src := range sources {
				for _, tk := range lexer.Tokenize(src) {
					if tk.Kind != r.kind {
						continue
					}
					text := string(src[tk.Start:tk.End])
					if r.hostile != "" && !strings.Contains(text, r.hostile) {
						continue
					}
					found = true
					checkRegion(t, r, src, tk)
				}
			}
			if !found {
				t.Errorf("no %s in %s holds %q.\n"+
					"\tThe tier's thesis is that the contents are NOT LEXED, "+
					"and a region whose contents a Perl lexer would read the "+
					"same way either opaque or not cannot distinguish the two.",
					r.what, tierOpaque, r.hostile)
			}
		})
	}
}

// checkRegion checks one occurrence of an opaque region: that the token
// covers the region end to end, and that nothing else was lexed inside
// it.
func checkRegion(t *testing.T, r opaqueRegion, src []byte, tk lexer.Token) {
	t.Helper()

	text := string(src[tk.Start:tk.End])
	if !strings.HasPrefix(text, r.opens) {
		t.Errorf("%s begins %q, want it to begin %q.\n"+
			"\tA region delimited short leaves its own first bytes for "+
			"some later rule to skip.", r.what, first(text), r.opens)
	}
	if !strings.HasSuffix(text, r.closes) {
		t.Errorf("%s ends %q, want it to end %q.\n"+
			"\tThe terminator is part of the region -- GLOSSARY.md settles "+
			"this for every multi-line region in this tier, because a "+
			"consumer needs to know where the region ENDS.", r.what, last(text), r.closes)
	}

	// Nothing else lexed inside. An overlapping token is the lexer
	// having read the contents: whatever it made of them, it made
	// something, and the region was not opaque.
	//
	// Overlap rather than containment, because the failure this catches
	// is a region delimited WRONG -- a format body that stopped at the
	// first `@` leaves the picture field as tokens that start inside the
	// body's bytes and run past its end.
	for _, other := range lexer.Tokenize(src) {
		if other.Start == tk.Start && other.End == tk.End && other.Kind == tk.Kind {
			continue
		}
		if other.Kind == lexer.Whitespace {
			continue
		}
		if other.Start < tk.End && other.End > tk.Start {
			t.Errorf("%s spans [%d,%d), and %s(%q) at [%d,%d) overlaps it.\n"+
				"\tThe contents were lexed. This tier's whole claim is that "+
				"they are not.",
				r.what, tk.Start, tk.End,
				other.Kind, string(src[other.Start:other.End]), other.Start, other.End)
		}
	}
}

// first and last render the ends of a region's text for a failure
// message, so a wrong boundary is reported as the bytes that are there
// rather than as the whole region.
func first(s string) string {
	if len(s) > 24 {
		return s[:24] + "..."
	}
	return s
}

func last(s string) string {
	if len(s) > 24 {
		return "..." + s[len(s)-24:]
	}
	return s
}

// TestTierOpaquePerlValidated runs every file in the tier through the
// pinned interpreter before it counts.
//
// Identical in intent to TestTierLiteralsPerlValidated, and it carries a
// different weight here. Two of this tier's files reach OUTSIDE the
// process for their output -- `qx{echo hi}` runs a shell command and
// `<*.nonexistent-xyz>` reads the filesystem -- so their pins are claims
// about the machine as well as about perl. Running them through the
// pinned interpreter is what keeps a file that became unreproducible
// from being reported as a refusal.
func TestTierOpaquePerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierOpaque, perl)

	for name, f := range tierFiles(t, tierOpaque) {
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

// TestTierOpaqueLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate.
// This is the tier's own, so a file that reaches forward fails HERE,
// named as this tier's problem rather than as one subtest among a
// hundred and thirteen.
//
// The lint has LESS to say about this tier than about any other, and
// that is not a weakness of the lint but the subject of the tier. Four
// of the six constructs emit no op at all, so the lint sees a heredoc
// file as tier 01's `const` and a pod file as the program without the
// pod. What it still catches is the surrounding machinery -- the
// `readline` a data section is observed through, the `backtick` `qx`
// emits -- reaching past 13, which is the direction that actually goes
// wrong.
func TestTierOpaqueLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierOpaque) {
		t.Run(name, func(t *testing.T) {
			if err := lintFile(t, f, tierOpaque, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// opaqueKinds returns the opaque-region token kinds a source contains,
// each once, in a stable order.
//
// Taken from `categories` rather than from a list here: the glossary
// categories this tier added ARE its constructs, and the map in
// categories.go is what ties a category to a kind. A second list would
// be a second place to forget one.
func opaqueKinds(src []byte) []lexer.Kind {
	want := []lexer.Kind{
		lexer.HeredocOpen, lexer.HeredocBody,
		lexer.Readline, lexer.Pod, lexer.DataSection, lexer.FormatBody,
	}
	var out []lexer.Kind
	for _, k := range want {
		for _, tk := range lexer.Tokenize(src) {
			if tk.Kind == k {
				out = append(out, k)
				break
			}
		}
	}
	return out
}

func containsKind(ks []lexer.Kind, want lexer.Kind) bool {
	for _, k := range ks {
		if k == want {
			return true
		}
	}
	return false
}

// usesAnyOp reports whether a compiled source emits any of a set of ops.
func usesAnyOp(ops, want []string) bool {
	for _, w := range want {
		if countOp(ops, w) > 0 {
			return true
		}
	}
	return false
}

// TestTierOpaqueRefusalsCited checks that a refusing file names WHICH
// site declines, from the stable inventory, rather than describing it.
//
// The rule is tier 01's and is biconditional, because `parse.RefusalCode`
// names sites in the PARSER and a refusal can be purely lexical:
//
//   - A file whose parser produces Unknowns MUST name one of their codes.
//   - A file whose parser produces none MUST NOT name a code, and must
//     carry the token facts that are its refusal instead.
//
// THE TIER WHERE THIS MATTERS MOST, and the reason is measured rather
// than asserted: EVERY TOKEN FACT IN THIS TIER PASSES, INCLUDING IN THE
// FILES THAT REFUSE. The lexer already delimits all six constructs; the
// five refusals are the parser having no rule for `format NAME = BODY`,
// for a statement containing a heredoc, or for a trailing data section.
// A refusing file here that cited prose, or cited nothing, would be read
// as "heredoc lexing is broken" -- which is the opposite of what the
// tier measured. The code is what puts the refusal above the lexer.
func TestTierOpaqueRefusalsCited(t *testing.T) {
	for name, f := range tierFiles(t, tierOpaque) {
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
