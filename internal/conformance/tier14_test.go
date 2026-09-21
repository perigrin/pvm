// ABOUTME: Tier 14 checked against the finished tooling: where the lexer must re-enter Perl inside a delimiter.
// ABOUTME: Five checks the tier's own issue names, each tier-specific rather than corpus-wide.
package conformance

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
)

// tierRecursive is the tier this file is about.
//
// A constant rather than a literal at five call sites, for the reason
// tier01_test.go gives for tierLiterals: the tier number is a POSITION
// and positions move.
const tierRecursive = "14_recursive"

// embeddedRegion is one of this tier's constructs: a span the lexer must
// delimit and then HAND BACK, so Perl runs inside its own output.
//
// `opens` and `closes` are the region's introducer and terminator as the
// corpus spells them, and `hostile` is content inside the region that
// separates the two strategies a lexer can use to find that terminator.
//
// THE `hostile` FIELD IS WHY THIS TEST EXISTS, and the reason is the
// inverse of tier 13's. There, the danger was a region being LEXED when
// it should have been opaque, and the hostile content was Perl-looking
// text that a wrongly-lexing reader would tokenise. Here the danger runs
// the other way: the region must actually RUN as Perl, so the content
// that separates a correct reader from a broken one is content whose
// DELIMITER CHARACTER appears inside a nested quote.
//
// Measured under 5.42.0, and the two constructs genuinely differ:
//
//   - `(?{ ... })`. perl parses the block as Perl to find its end.
//     `/b(?{ $k = length("}}}") })/` compiles and prints 3 -- the three
//     braces inside the string do NOT close the block. A lexer that
//     brace-counted would stop at the first one and hand the parser a
//     truncated region.
//   - `s{a}{ ... }e`. perl COUNTS DELIMITERS here rather than parsing.
//     `s{a}{ $n + length("}}") }e` is a syntax error -- "Unmatched right
//     curly bracket" -- because the brace in the string does close the
//     replacement. So the hostile content for a substitution must use a
//     character that is NOT the delimiter: `s{a}{ $n + length("))") }e`
//     compiles and prints 3bc, and a lexer that scanned for a balanced
//     `)` would truncate it.
//
// A region holding `$n+1` or `$k = 5` -- which is every file this tier
// shipped with -- is delimited identically by a brace counter and by a
// Perl parser, so it measures DELIMITING and says nothing about the
// re-entry that is half this tier's thesis.
type embeddedRegion struct {
	// what names the construct, for failure messages.
	what string

	// prefix is what the whole token must begin with, which is how the
	// construct is IDENTIFIED rather than merely found.
	//
	// Without it `(?{` and its hostile content appear in a `m//` and in
	// a `qr//` alike, and two entries differing only in their name would
	// both be satisfied by whichever token came first. Mutation-tested:
	// removing the hostile content from the match's block failed BOTH
	// the match and the qr// subtests before this field existed, which
	// is the alias the field closes.
	prefix string

	// opens and closes are the region's introducer and terminator as
	// they appear inside the one token.
	//
	// Checked rather than assumed because "delimited correctly" is a
	// claim about both ENDS, exactly as at tier 13. A token that stopped
	// at the hostile character would still start correctly.
	opens, closes string

	// hostile is content that must appear INSIDE the region, chosen so
	// that a reader using the wrong end-finding strategy stops early.
	hostile string
}

// embeddedRegions is the constructs whose contents are handed back to
// Perl, each with the content that makes "re-entered rather than merely
// delimited" a falsifiable claim.
//
// `(??{ })` is not a separate entry. Measured, it is delimited exactly as
// `(?{ })` is -- the extra `?` changes WHEN the engine runs the block,
// which is inside the regex engine and invisible to the lexer -- so a
// second entry would assert the same lexical fact twice. The adjacency
// file carries it, which is where the tier's README puts it.
func embeddedRegions() []embeddedRegion {
	return []embeddedRegion{
		{
			what:   "code block in a match",
			prefix: "/",
			opens:  "(?{",
			closes: "})",
			// Braces inside a string inside the block. perl parses to
			// find the end, so these do not close it; a brace counter
			// stops at the first.
			hostile: `length("}}}")`,
		},
		{
			what:   "code block in a qr//",
			prefix: "qr",
			opens:  "(?{",
			closes: "})",
			// A DIFFERENT count of braces from the match's, so the two
			// entries cannot be satisfied by the same token even if the
			// prefix check were removed. Belt and braces, and cheap: the
			// alias this guards against was real until it was measured.
			hostile: `length("}}")`,
		},
		{
			what:   "expression replacement of s///e",
			prefix: "s{",
			opens:  "s{a}{",
			closes: "}e",
			// Parens rather than braces, because measured, perl's own
			// scan for the replacement's end COUNTS the delimiter: a
			// `}` in a string there is a syntax error in perl itself.
			// The falsifiable claim available is the weaker one that a
			// reader must not balance some OTHER bracket.
			hostile: `length("))")`,
		},
	}
}

// TestTierRecursiveReEntersPerl is the tier's thesis, stated as a test.
//
// Each construct must arrive as ONE token covering the whole region,
// with the region's own contents inside that token rather than beside
// it, and with content that separates a reader who re-enters Perl from
// one who counts characters.
//
// WHY THE OPTREE CANNOT ADJUDICATE ANY OF THIS, which is this tier's
// defining measurement and its README's central finding. `(?{ })` and
// `(??{ })` ADD NO OP AT ALL: `$s =~ /a(?{ $n = 1 })b/` compiles to one
// `match` op whose PATTERN STRING carries the block. The block's optree
// exists in the full tree but every op in it is marked `-`, optimised
// out of the exec path -- it lives in a separate CV the regex engine
// calls. `qr//` with embedded code is the same: one `qr` op, the block
// inside its pattern text. So the construct that most plainly re-enters
// Perl is INVISIBLE to a measurement of ops, and only two ops are
// claimed for three constructs. The token stream is not the convenient
// place to assert this; it is the only one.
func TestTierRecursiveReEntersPerl(t *testing.T) {
	files := tierFiles(t, tierRecursive)

	var sources [][]byte
	for _, f := range files {
		sources = append(sources, []byte(f.Source))
	}

	for _, r := range embeddedRegions() {
		t.Run(r.what, func(t *testing.T) {
			found := false
			for _, src := range sources {
				for _, tk := range lexer.Tokenize(src) {
					if tk.Kind != lexer.Quote {
						continue
					}
					text := string(src[tk.Start:tk.End])
					if !strings.HasPrefix(text, r.prefix) {
						continue
					}
					if !strings.Contains(text, r.hostile) {
						continue
					}
					if !strings.Contains(text, r.opens) {
						continue
					}
					found = true
					checkEmbedded(t, r, src, tk)
				}
			}
			if !found {
				t.Errorf("no %s in %s holds %s.\n"+
					"\tThe tier's thesis is that the region is HANDED BACK "+
					"to Perl, and a region whose contents a character "+
					"counter and a Perl parser delimit identically cannot "+
					"distinguish the two.",
					r.what, tierRecursive, r.hostile)
			}
		})
	}
}

// checkEmbedded checks one occurrence of an embedded region: that the
// token covers the construct end to end with the hostile content inside
// it, and that nothing was lexed beside it.
func checkEmbedded(t *testing.T, r embeddedRegion, src []byte, tk lexer.Token) {
	t.Helper()

	text := string(src[tk.Start:tk.End])

	// The region's terminator is INSIDE the one token, after the hostile
	// content. A reader that stopped at the hostile character would have
	// produced a token ending before this, so the check is that the
	// closer follows the hostile content rather than merely appearing.
	h := strings.Index(text, r.hostile)
	if c := strings.Index(text[h:], r.closes); c < 0 {
		t.Errorf("%s token %q holds %s but no %q after it.\n"+
			"\tThe region was cut short: whatever closed it was not its "+
			"own terminator.", r.what, text, r.hostile, r.closes)
	}

	// Nothing else lexed beside it. An overlapping token is the region
	// having ended early, with its tail spilling out as separate tokens
	// -- which is precisely what a delimiter counter produces.
	//
	// Overlap rather than containment, for tier 13's reason: the failure
	// this catches is a region delimited WRONG, whose stray tokens start
	// inside the token's bytes or run past its end.
	for _, other := range lexer.Tokenize(src) {
		if other.Start == tk.Start && other.End == tk.End && other.Kind == tk.Kind {
			continue
		}
		if other.Kind == lexer.Whitespace {
			continue
		}
		if other.Start < tk.End && other.End > tk.Start {
			t.Errorf("%s spans [%d,%d), and %s(%q) at [%d,%d) overlaps it.\n"+
				"\tThe region ended early and its tail was lexed as "+
				"ordinary code.",
				r.what, tk.Start, tk.End,
				other.Kind, string(src[other.Start:other.End]), other.Start, other.End)
		}
	}
}

// TestTierRecursivePerlValidated runs every file in the tier through the
// pinned interpreter before it counts.
//
// Identical in intent to TestTierLiteralsPerlValidated, and it carries a
// particular weight here for a reason the README measures. `s/a/CONST/e`
// FOLDS -- `s/a/uc("z")/e` compiles to `const[PV "Z"] s/FOLD` with a
// plain `subst` and no `substcont` at all, the `/e` erased. Every file
// here therefore uses a deliberately unfoldable replacement, and it is
// the OUTPUT that shows the replacement actually ran. A pinned output
// perl does not produce would leave the tier's two claimed ops
// unwitnessed.
func TestTierRecursivePerlValidated(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("validating %s against %s", tierRecursive, perl)

	for name, f := range tierFiles(t, tierRecursive) {
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

// TestTierRecursiveLint runs the dependency lint over the tier.
//
// `TestCorpusLints` runs it over the whole corpus, which is the gate.
// This is the tier's own, so a file that reaches forward fails HERE,
// named as this tier's problem rather than as one subtest among a
// hundred and eighteen.
//
// The lint has an unusual job in the last tier, because this tier's
// declared prerequisite is `09_regex` and NOT tier N-1. A tier at the
// end of the corpus can reach backwards into anything without the lint
// objecting -- there is no later tier to reach INTO. What it still
// catches is the direction that goes wrong in practice: a file whose
// replacement expression grew a construct no tier claims at all.
func TestTierRecursiveLint(t *testing.T) {
	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}

	for name, f := range tierFiles(t, tierRecursive) {
		t.Run(name, func(t *testing.T) {
			if err := lintFile(t, f, tierRecursive, tiers); err != nil {
				t.Errorf("%s", err)
			}
		})
	}
}

// recursiveForms returns the embedded-code forms a source uses, in a
// stable order.
//
// SPELLINGS rather than token kinds, and this is forced rather than
// chosen. Every construct in this tier is a `lexer.Quote`, the same kind
// tier 01's `"abc"` and tier 09's `m//` arrive as: the lexer has no
// `CodeBlock` kind because the block is INSIDE the quote's text, which
// is the whole finding. A kind-based enumeration, as tier 13 uses, would
// report one category for all five files.
//
// Read off the token TEXT rather than the raw source, so a `(?{` written
// inside a comment is not mistaken for a construct.
func recursiveForms(src []byte) []string {
	forms := []struct{ name, spelling string }{
		{"(??{ })", "(??{"},
		{"(?{ })", "(?{"},
		{"s///ee", "ee"},
		{"s///e", "e"},
		{"qr//", "qr"},
	}

	var out []string
	seen := map[string]bool{}
	for _, tk := range lexer.Tokenize(src) {
		if tk.Kind != lexer.Quote {
			continue
		}
		text := string(src[tk.Start:tk.End])
		for _, f := range forms {
			if !usesForm(text, f.name, f.spelling) || seen[f.name] {
				continue
			}
			seen[f.name] = true
			out = append(out, f.name)
			// `(??{` also contains `(?{`... no it does not, but a
			// `s///ee` DOES end in `e`, so the flag forms stop at the
			// first match to keep `/ee` from also reporting `/e`.
			if f.name == "s///ee" {
				seen["s///e"] = true
			}
		}
	}
	return out
}

// usesForm reports whether one quote token exercises one form.
//
// The block forms are a substring test on the token's text. The flag
// forms are not: `e` appears in every `length(...)` this tier writes, so
// a substring test would find `s///e` in a match that has no flag at
// all. A substitution's flags are what follows its LAST delimiter, which
// is where they are read from.
func usesForm(text, name, spelling string) bool {
	switch name {
	case "(??{ })", "(?{ })":
		return strings.Contains(text, spelling)
	case "qr//":
		return strings.HasPrefix(text, "qr")
	case "s///e", "s///ee":
		return strings.HasSuffix(text, spelling) && strings.HasPrefix(text, "s")
	}
	return false
}

// TestTierRecursiveAdjacency checks the tier's adjacency file holds every
// construct the tier introduces, each next to another, and that it pairs
// with the DECLARED prerequisite rather than with tier N-1.
//
// WHAT OPS CANNOT DO HERE, and this tier is the extreme case in the
// corpus. Tier 01 found `padrange` absorbing `pushmark`, so the declared
// INTRODUCES set was a UNION across the tier's files and never a property
// of one. Here it is worse: measured, `(?{ })`, `(??{ })` and `qr//` with
// embedded code emit NO OP OF THEIR OWN AT ALL -- three of the tier's
// five forms are invisible to `-exec`, their bodies living in separate
// CVs marked `-`. Only `substcont` and `entereval` are claimed. An
// adjacency check reading ops would be reading two forms out of five.
//
// So COVERAGE is checked on the TOKENS, as at tier 13 and for the same
// reason, with the difference that the forms are read from the token
// TEXT rather than from its kind -- see recursiveForms.
//
// THE PAIRING IS WITH `09_regex`, NOT `13_opaque`, and the distinction is
// this tier's README's most emphatic claim rather than a technicality.
// Tier 13 is merely N-1. Nothing here is a heredoc, a format, a `qx` or
// a POD block, and nothing here needs one; pairing with 13 would assert
// a dependency that does not exist. `09_regex` is exact: every construct
// in this tier is a regex plus something. `s///e` is tier 09's `s///`
// with a flag, `(?{ })` is its `m//` with a block in the pattern, and
// `/e` is not a thing that exists apart from `s`. So the pairing is read
// from the README via readTierDeps and this test FAILS if the declaration
// changes -- which is the point of reading it rather than computing N-1.
//
// The pairing is then checked as the prerequisite's OP appearing in the
// adjacency file's optree, which is the one place ops are the evidence:
// `09_regex` emits real ops (`match`, `subst`, `qr`, `regcomp`) where
// this tier's own constructs emit almost none, so the prerequisite is
// exactly the half of the file an op measurement can see.
func TestTierRecursiveAdjacency(t *testing.T) {
	files := tierFiles(t, tierRecursive)

	adj, ok := files[adjacencyFile]
	if !ok {
		t.Fatalf("%s has no %s", tierRecursive, adjacencyFile)
	}
	adjForms := recursiveForms([]byte(adj.Source))

	for name, f := range files {
		if name == adjacencyFile {
			continue
		}
		for _, form := range recursiveForms([]byte(f.Source)) {
			if !containsForm(adjForms, form) {
				t.Errorf("%s uses %s, which %s does not.\n"+
					"\tThe adjacency file must hold every construct the "+
					"tier introduces, or the pairing it exists to reach "+
					"is not reachable for that construct.",
					name, form, adjacencyFile)
			}
		}
	}

	ops, err := opsOf(t, adj.Source)
	if err != nil {
		t.Fatalf("%s: %v", adjacencyFile, err)
	}

	// One body, not several. `nextstate` is one per statement, which is
	// how this counts statements without a parser of its own.
	if n := countOp(ops, "nextstate"); n < 2 {
		t.Errorf("%s compiles to %d statement(s); adjacency needs at least 2",
			adjacencyFile, n)
	}

	// The DECLARED prerequisite, READ from the README rather than
	// computed as tier N-1.
	deps, err := readTierDeps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	dep, ok := deps[tierRecursive]
	if !ok {
		t.Fatalf("%s declares no DEPENDS ON", tierRecursive)
	}
	if dep == "nothing" {
		t.Fatalf("%s declares no prerequisite, but every construct here is "+
			"a regex with something added", tierRecursive)
	}

	tiers, err := readTierOps(corpusDir)
	if err != nil {
		t.Fatalf("reading tier READMEs: %v", err)
	}
	depOps := tiers[dep]
	if len(depOps) == 0 {
		t.Fatalf("%s declares %s, which introduces no op to pair with",
			tierRecursive, dep)
	}

	var paired []string
	for _, op := range depOps {
		if countOp(ops, op) > 0 {
			paired = append(paired, op)
		}
	}
	if len(paired) == 0 {
		t.Errorf("%s declares DEPENDS ON %s, and %s emits none of that "+
			"tier's ops (%s).\n"+
			"\tThe adjacency file is where the two tiers meet in one "+
			"compiled body; without one of the prerequisite's ops the "+
			"declared edge is prose.",
			tierRecursive, dep, adjacencyFile, strings.Join(depOps, " "))
		return
	}
	t.Logf("%s pairs with %s through %s",
		adjacencyFile, dep, strings.Join(paired, ", "))
}

func containsForm(fs []string, want string) bool {
	for _, f := range fs {
		if f == want {
			return true
		}
	}
	return false
}

// TestTierRecursiveRefusalsCited checks that a refusing file names WHICH
// site declines, from the stable inventory, rather than describing it.
//
// The rule is tier 01's and is biconditional, because `parse.RefusalCode`
// names sites in the PARSER and a refusal can be purely lexical:
//
//   - A file whose parser produces Unknowns MUST name one of their codes.
//   - A file whose parser produces none MUST NOT name a code, and must
//     carry the token facts that are its refusal instead.
//
// AS OF TODAY THIS TIER HAS NO REFUSING FILE, which is a measurement and
// a surprising one. Its own issue expected the opposite -- "the one most
// likely to refuse wholesale" -- and the corpus says otherwise: all five
// files parse, including the `(?{ })` whose contents perl itself must
// re-enter Perl to delimit. The check is still written, because the event
// it guards is a file ACQUIRING a refusal, and this tier is where the
// biconditional bites hardest: a `(?{ })` file that refused would almost
// certainly refuse LEXICALLY, with no parser Unknown to name at all, and
// a header citing a code anyway would make `run.go` fail it as a stale
// marker -- correctly, because the claim would be false.
func TestTierRecursiveRefusalsCited(t *testing.T) {
	for name, f := range tierFiles(t, tierRecursive) {
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
