// ABOUTME: A token fact that cannot fail asserts nothing; two mechanisms make one, and both are checked here.
// ABOUTME: Seven such facts shipped before these checks existed, each found by a reader rather than by the suite.
package conformance

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// reStringLiteralFact matches a token fact naming the `string literal`
// category, capturing the text it claims.
//
// The category is spelled out rather than matched loosely because this
// check is only sound for categories whose token text INCLUDES
// delimiters. `word`, `operator` and `variable` do not, and a fact about
// those is perfectly falsifiable with a bare spelling.
var reStringLiteralFact = regexp.MustCompile(
	`^(?:one|no) string literal whose text is "(.*)"$`)

// TestStringLiteralFactsSpellTheirDelimiters rejects a fact that cannot
// fail.
//
// A Quote token's text includes its delimiters: `internal/lexer/quote.go`
// captures `start` BEFORE consuming the opening quote, so the source
// `"PI"` produces one token whose text is the four-byte `"PI"` and never
// the two-byte `PI`. A fact written as
//
//	no string literal whose text is "PI"
//
// therefore compares a delimiter-inclusive token against a
// delimiter-free string and can never match ANY input. It is
// unfalsifiable, which in a corpus is worse than absent: it reads as a
// claim and holds nothing up.
//
// Five such facts shipped before this check existed, in two commits that
// each described themselves as FIXING a vacuous fact. That is the
// argument for a mechanical check rather than a convention: the failure
// survived two rounds of a reviewer looking directly at it, because
// noticing it requires knowing where `emit` is called in the lexer.
//
// The check is deliberately narrow. It asks only that the claimed text
// begin with a delimiter the lexer could have produced -- a quote, or a
// quote-operator keyword. It does not try to validate the whole spelling,
// because that would duplicate `checkTokenFact` and drift from it.
func TestStringLiteralFactsSpellTheirDelimiters(t *testing.T) {
	tiers, err := os.ReadDir(corpusDir)
	if err != nil {
		t.Fatalf("reading the corpus: %v", err)
	}

	for _, tier := range tiers {
		if !tier.IsDir() {
			continue
		}
		paths, err := filepath.Glob(filepath.Join(corpusDir, tier.Name(), "*.t"))
		if err != nil {
			t.Fatalf("globbing %s: %v", tier.Name(), err)
		}
		for _, path := range paths {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading %s: %v", path, err)
			}
			f, err := ParseFile(string(raw))
			if err != nil {
				t.Fatalf("%s: %v", path, err)
			}
			for _, fact := range f.TokenFacts {
				fact = strings.TrimSpace(fact)
				m := reStringLiteralFact.FindStringSubmatch(fact)
				if m == nil {
					continue
				}
				if !opensAsQuote(m[1]) {
					t.Errorf("%s: %q can never match.\n"+
						"\tA Quote token's text includes its delimiters, so a "+
						"`string literal` fact naming a bare spelling compares "+
						"against something the lexer never emits.\n"+
						"\tSpell the delimiters, as every other such fact in "+
						"the corpus does: `\\\"%s\\\"`.",
						filepath.Join(tier.Name(), filepath.Base(path)),
						fact, m[1])
				}
			}
		}
	}
}

// reNegativeFact matches any `no` token fact, whatever its category,
// capturing the category and the text it claims.
//
// Unlike `reStringLiteralFact` this is category-agnostic, because the
// check below is sound for every category: it asks about the SOURCE
// rather than about how a particular token's text is built.
var reNegativeFact = regexp.MustCompile(
	`^no ([a-z -]+?) whose text is (".*")$`)

// TestNegativeFactsNameTextTheSourceContains rejects a negative fact
// that cannot fail because its spelling is absent from the file.
//
// THE SECOND VACUITY MECHANISM. `TestStringLiteralFactsSpellTheirDelimiters`
// catches a fact naming text the LEXER can never emit. This catches one
// naming text the lexer can emit but THIS SOURCE cannot produce.
//
// `checkTokenFact` compares `src[tk.Start:tk.End]` against the claimed
// text, so a token whose text is `X` requires the bytes `X` at some
// position in the source. A source not containing those bytes can
// produce no such token under ANY lexing, correct or broken -- so the
// count is always zero, the fact always passes, and it asserts nothing.
//
// The two mechanisms are genuinely separate and conflating them is what
// shipped the seventh instance. `+(` is unreachable the first way: it
// is not an entry in `scan.go`'s operator table, so no input produces
// it. `+=` IS an entry and is unreachable the second way in a source
// with no `=`. A reader checking only the table passes the second case.
//
// WHY ONLY `no` FACTS. A `one` fact fails when the count is not one, and
// an absent spelling makes the count zero, so it cannot be vacuous.
//
// WHY THE WHOLE FILE TEXT AND NOT JUST THE SOURCE BLOCK. A fact is
// checked against `f.Source` by `checkTokenFact`, so `f.Source` is the
// right question, and this asks exactly it.
func TestNegativeFactsNameTextTheSourceContains(t *testing.T) {
	for path, f := range allFilesForVacuity(t) {
		for _, fact := range f.TokenFacts {
			fact = strings.TrimSpace(fact)
			m := reNegativeFact.FindStringSubmatch(fact)
			if m == nil {
				continue
			}
			text, err := strconv.Unquote(m[2])
			if err != nil {
				// A silent skip here is how this check shipped inert:
				// the first draft captured the text WITHOUT its quotes,
				// so every Unquote failed, every fact was skipped, and
				// the test passed over a corpus it never read. A check
				// that cannot parse what it is checking must say so.
				t.Errorf("%s: %q has text this check cannot unquote: %v\n"+
					"\tThe fact may still be valid -- checkTokenFact "+
					"parses it separately -- but this check skipped it, "+
					"and a skip it does not report is a check that is "+
					"not running.", path, fact, err)
				continue
			}
			if !strings.Contains(f.Source, text) {
				t.Errorf("%s: %q can never fail.\n"+
					"\tThe source does not contain %q anywhere, so no "+
					"lexing of it -- correct or broken -- can produce a "+
					"token whose text is that. The count is always zero "+
					"and the fact always passes.\n"+
					"\tEither name a spelling the source writes, or drop "+
					"the fact and say in the header why none is honest.",
					path, fact, text)
			}
		}
	}
}

// allFilesForVacuity reads every corpus file, keyed by its tier-relative
// path.
//
// One tier deep rather than a recursive walk, for `allSources`' reason:
// the lint fixtures below the tiers are not corpus files and parsing
// them as such reports failures about programs no one claims.
func allFilesForVacuity(t *testing.T) map[string]File {
	t.Helper()

	tiers, err := os.ReadDir(corpusDir)
	if err != nil {
		t.Fatalf("reading the corpus: %v", err)
	}

	files := map[string]File{}
	for _, tier := range tiers {
		if !tier.IsDir() {
			continue
		}
		paths, err := filepath.Glob(filepath.Join(corpusDir, tier.Name(), "*.t"))
		if err != nil {
			t.Fatalf("globbing %s: %v", tier.Name(), err)
		}
		for _, path := range paths {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("reading %s: %v", path, err)
			}
			f, err := ParseFile(string(raw))
			if err != nil {
				t.Fatalf("%s: %v", path, err)
			}
			files[filepath.Join(tier.Name(), filepath.Base(path))] = *f
		}
	}
	return files
}

// opensAsQuote reports whether a claimed token text starts the way a
// Quote token actually can.
//
// The escaped `\"` is how a double-quoted string is written inside a
// fact; a single quote and the quote-operator keywords are the other
// openers the lexer produces.
func opensAsQuote(text string) bool {
	switch {
	case strings.HasPrefix(text, `\"`), strings.HasPrefix(text, `'`):
		return true
	}
	for _, op := range []string{"q", "qq", "qw", "qr", "qx", "m", "s", "tr", "y"} {
		if strings.HasPrefix(text, op) {
			return true
		}
	}
	return strings.HasPrefix(text, "`")
}
