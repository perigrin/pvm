// ABOUTME: A `string literal` token fact whose text omits its delimiters can never match, so it asserts nothing.
// ABOUTME: Five such facts shipped before this check existed, each described in its own commit as a fix for vacuity.
package conformance

import (
	"os"
	"path/filepath"
	"regexp"
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
