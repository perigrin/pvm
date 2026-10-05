// ABOUTME: Refusal codes: two refusals are distinguishable, and the inventory cannot drift from the sites.
// ABOUTME: A code is a stable identifier, not a message — a reworded message must not move a corpus file.

package parse_test

import (
	"os"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestRefusalCodesDistinct: two refusals with unrelated causes must not
// look the same.
//
// This is the whole point of the issue. Before codes, every refusal said
// "1 Unknown node(s)" and nothing else, so a file waiting on a missing
// operand and a file waiting on an unimplemented statement form were
// indistinguishable, and a refusal that CHANGED CAUSE while staying a
// refusal was invisible.
func TestRefusalCodesDistinct(t *testing.T) {
	// Two sources that refuse for genuinely different reasons: a ternary
	// missing its colon, and an operator whose operand ran out.
	const stmt = "my $x = $a ? $b;\n"
	const operand = "$a +\n"

	a := firstRefusalCode(t, stmt)
	b := firstRefusalCode(t, operand)

	if a == b {
		t.Errorf("two unrelated refusals share code %q; "+
			"a refusal that changes cause must be visible", a)
	}
}

// TestEveryUnknownSiteHasACode: no Unknown may be built without one.
//
// A site that forgets its code reintroduces the anonymous refusal for
// exactly the construct nobody was watching, which is the one that
// matters. The check is on the PRODUCT rather than on the source text:
// every Unknown the parser emits must carry a code the inventory knows,
// and the fixtures must between them reach a DIFFERENT code each, or the
// list could collapse onto one site while still passing.
func TestEveryUnknownSiteHasACode(t *testing.T) {
	srcs := refusingSources()
	reached := map[parse.RefusalCode]string{}

	for _, src := range srcs {
		found := collectUnknown(parse.Parse([]byte(src)))
		if len(found) == 0 {
			t.Errorf("%q no longer refuses; this fixture has outlived its "+
				"purpose and must be replaced, not deleted", src)
			continue
		}
		for _, n := range found {
			if n.Refusal == "" {
				t.Errorf("%q: an Unknown carries no refusal code", src)
				continue
			}
			if _, ok := parse.RefusalSites[n.Refusal]; !ok {
				t.Errorf("%q: Unknown carries code %q, which the inventory "+
					"does not list", src, n.Refusal)
			}
		}
		code := found[0].Refusal
		if other, dup := reached[code]; dup {
			t.Errorf("%q and %q both refuse with %q; one of them is no "+
				"longer reaching the site it was written for", src, other, code)
		}
		reached[code] = src
	}

	if len(reached) != len(srcs) {
		t.Errorf("%d fixtures reached %d distinct codes", len(srcs), len(reached))
	}
}

// TestRefusalCodeInventory: the eleven sites are enumerated where a reader
// finds them, and the enumeration is what the code READS.
//
// A test that merely counted would be a second list free to disagree with
// the first, which this package has been bitten by repeatedly. So
// parse.RefusalSites is not documentation about the sites -- it is the
// table the constants are declared against, and this test holds it to
// three things: every declared code is described, every description is
// non-empty prose a reader can act on, and the count matches the sites
// the parser actually has.
func TestRefusalCodeInventory(t *testing.T) {
	// The site count is COUNTED FROM THE SOURCE, not written down here. A
	// constant would be the second list this test exists to prevent: add
	// a twelfth `&Node{Kind: Unknown, ...}` and a hardcoded eleven would go
	// on passing while the new site refused anonymously.
	sites := countUnknownSites(t)

	// Ten codes for eleven sites, and the eleventh is accounted for rather
	// than forgotten: parse.go's `expr.Kind == Unknown` widens an existing
	// Unknown from the expression to the statement and carries the inner
	// code outward. It is a change of span, not of cause, so it declares
	// none of its own -- see the note in RefusalSites.
	const propagating = 1

	if len(parse.RefusalSites)+propagating != sites {
		t.Errorf("inventory lists %d refusal codes and %d site propagates, "+
			"but the parser has %d Unknown construction sites; the two "+
			"must agree", len(parse.RefusalSites), propagating, sites)
	}

	seen := map[string]bool{}
	for code, desc := range parse.RefusalSites {
		if code == "" {
			t.Error("the inventory holds an empty code")
		}
		if strings.TrimSpace(desc.What) == "" {
			t.Errorf("%s: no description; a code with no prose behind it "+
				"tells a corpus author nothing", code)
		}
		if strings.TrimSpace(desc.Where) == "" {
			t.Errorf("%s: no source location; a reader must find the site "+
				"without grepping", code)
		}
		if seen[desc.Where] {
			t.Errorf("%s: two codes claim source location %q", code, desc.Where)
		}
		seen[desc.Where] = true
	}
}

// countUnknownSites counts `&Node{Kind: Unknown, ...}` constructions in
// the parser's own source.
//
// Reading the source is the point. The inventory's size has to be checked
// against something that CHANGES WHEN A SITE IS ADDED, and the only such
// thing is the sites themselves; any number typed into this file is a
// second list free to go stale, which is the failure this package has hit
// four times.
//
// A string scan rather than go/ast: the construction is a composite
// literal with a constant field, which the text finds exactly and which a
// syntax tree would find by the same shape with more machinery.
//
// The needle includes `Refusal:` on purpose. `Kind: Unknown,` alone also
// matched the PROSE in refusal.go -- two comments saying "one per
// `&Node{Kind: Unknown, ...}`" pushed the count from ten to twelve, so
// editing a comment would have failed this test. Every real site assigns
// a Refusal, which is the property being counted anyway.
func countUnknownSites(t *testing.T) int {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}

	total := 0
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		total += strings.Count(string(src), "&Node{Kind: Unknown, Refusal:")
	}
	if total == 0 {
		t.Fatal("found no Unknown construction sites; the scan is broken, " +
			"not the parser")
	}
	return total
}

// refusingSources are sources that reach the parser's refusal sites, one
// per code the parser can currently produce.
//
// Seven codes, not ten. Three sites are declared and wired but NOT
// REACHABLE from any source today, and saying so is better than a fixture
// that pretends otherwise:
//
//   - unimplemented_statement, parse.go's `statementKeywords` check. The
//     list is empty: `try`, `catch` and `finally` were its last entries, and
//     parseTry reads the one shape of theirs `WORD BLOCK` could not. Its
//     fixture was `use feature 'try'; try { 1 } catch ($e) { 2 }`, which
//     parses now, and no valid perl reaches the site in its place.
//   - empty_deref, term.go's `inner == nil`. `${}` lexes as
//     DerefSigil($) Operator({) CloseBracket(}), and parseExpr inside the
//     braces returns the `}` as a not_a_term Unknown rather than nil, so
//     the nil branch never fires. The site is correct for the day
//     parseExpr learns to return nil there; it is not dead code, it is
//     unreached code.
//   - not_an_expression, parse.go's `expr == nil`. A statement with no
//     significant tokens is consumed before parseStatement is reached, so
//     parseExpr is never handed nothing.
//
// All three are left wired rather than deleted: an unreachable site that
// becomes reachable must arrive with a code, not without one.
//
// A fixture that stops refusing -- as `goto &other;` did once goto
// landed -- is reported by TestEveryUnknownSiteHasACode rather than
// quietly dropped, because a fixture silently doing nothing is how a test
// stops testing.
func refusingSources() []string {
	return []string{
		"$a +\n",                     // missing_operand
		"my $x = $a ? $b;\n",         // ternary_no_colon
		"my $x = $a .. $b .. $c;\n",  // nonassoc_repeated
		"my $x = $a <=> $b == $c;\n", // chain_class_mismatch
		"my $x = ${};\n",             // not_a_term
		"$a $b;\n",                   // trailing_tokens
		"map { $_",                   // unclosed_brace
	}
}

// collectUnknown gathers every Unknown in a tree.
func collectUnknown(n *parse.Node) []*parse.Node {
	if n == nil {
		return nil
	}
	var out []*parse.Node
	if n.Kind == parse.Unknown {
		out = append(out, n)
	}
	for _, c := range n.Children {
		out = append(out, collectUnknown(c)...)
	}
	return out
}

// firstRefusalCode returns the code of the first Unknown in a parse, and
// fails the test when the source does not refuse at all.
func firstRefusalCode(t *testing.T, src string) parse.RefusalCode {
	t.Helper()
	found := collectUnknown(parse.Parse([]byte(src)))
	if len(found) == 0 {
		t.Fatalf("%q no longer refuses; replace this fixture", src)
	}
	return found[0].Refusal
}
