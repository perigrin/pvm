// ABOUTME: The subject side of the fidelity harness: what this parser reports about each statement.
// ABOUTME: An Unknown hedges every marker, because saying nothing is scored as claiming nothing is there.

package parse

import (
	"bytes"
	"strings"

	"tamarou.com/pvm/internal/parseoracle"
)

// markerKinds is every site kind the oracle scores. A statement this parser
// could not settle must hedge all of them, not just the one it happened to
// be looking for.
var markerKinds = []string{
	parseoracle.SiteKindReference,
	parseoracle.SiteKindHash,
	parseoracle.SiteKindMatch,
	parseoracle.SiteKindReadline,
	parseoracle.SiteKindAnonhash,
}

// Sites reports what this parser concluded about each statement, in the
// vocabulary the subject contract asks for.
//
// The rule that matters here is what an Unknown reports. CompareFacts
// attributes each of perl's sites to the innermost statement covering it and
// then decides (compare_facts.go:218-232):
//
//	case owner != nil && owner.decided[m] > 0:  // exact
//	case owner != nil && owner.hedged[m] > 0:   // wider
//	default:                                    // WRONG
//
// An Unknown node emits no sites, so owner is nil for every perl site inside
// it and the verdict is WRONG -- not no-answer. No-answer is a FILE-level
// bucket: Declined, !OK, or no site of a marker's kind anywhere in the file.
// Once the subject reports one hash site, every rv2hv in that file is scored
// per statement.
//
// So "emit Unknown and stay quiet" is not a refusal. It is the claim that
// nothing is there, and it is wrong wherever perl found something. The
// refusal has to be spoken: an Unresolved site of each kind, spanning the
// statement, which lands in hedged and scores wider.
//
// That rule is why Unknown exists. Without it Unknown is a silent wrong
// answer with a reassuring name.
func Sites(root *Node, src []byte) []parseoracle.SubjectCallSite {
	// Non-nil even when empty: the contract distinguishes "I looked and
	// found none" from "I did not look", and a nil slice marshals to JSON
	// null, which decodes as absent. See SubjectFacts.CallSites.
	sites := []parseoracle.SubjectCallSite{}

	lines := lineIndex(src)
	for _, n := range root.Children {
		start, end := lines.at(n.Start), lines.at(n.End-1)

		if n.Kind == Unknown {
			for _, kind := range markerKinds {
				sites = append(sites, parseoracle.SubjectCallSite{
					Kind:       kind,
					Line:       start,
					EndLine:    end,
					Unresolved: true,
				})
			}
			continue
		}

		// A statement this parser DID read reports what it decided. Without
		// this the subject only ever hedges, and a subject that never commits
		// cannot be wrong -- nor right. WRONG=0 with exact=0 is the vacuous
		// gate the M1 issue's "exact floor" row exists to prevent.
		//
		// Each site is reported at ITS OWN node's span, not the statement's.
		// perl's probe walks the CVs, so a backslash in a sub body is
		// reported at the body's line; CompareFacts spends a decided site
		// only against the innermost subject statement covering that line
		// (`compare_facts.go:349`). A site pooled under `sub foo {`'s whole
		// span is never reached from inside the body, and a marker this
		// parser read correctly scores WRONG.
		for _, d := range decidedMarkers(n) {
			sites = append(sites, parseoracle.SubjectCallSite{
				Kind:       d.kind,
				Line:       lines.at(d.node.Start),
				EndLine:    lines.at(d.node.End - 1),
				Unresolved: d.unresolved,

				// The srefgen marker is decided by this field, not by the
				// kind (`compare_facts.go:336`): a reference site without it
				// is "an answer, but not one that accounts for a reference
				// perl took", which lands in neither decided nor hedged. A
				// `\` this parser read IS a reference taken. The other four
				// markers are decided by their kind and ignore the field.
				//
				// A hedge takes nothing -- it is the refusal to say -- and
				// groupByStatement reads Unresolved first either way.
				TookReference: d.kind == parseoracle.SiteKindReference &&
					!d.unresolved,
			})
		}
	}
	return sites
}

// decided is one marker this parser settled and the node that settled it.
// The node travels with the kind because a site's line is its own, not the
// enclosing statement's -- see Sites.
//
// unresolved makes it a hedge rather than a commitment. A nested Unknown
// settles every marker the only way it can: by refusing, out loud.
type decided struct {
	kind       string
	node       *Node
	unresolved bool
}

// decidedMarkers reports which of the five markers this statement contains,
// in the subject vocabulary. Each is a conclusion the parser committed to.
//
// Only what the tree SHOWS, never what the bytes suggest. `$a % $b` is a
// Binary whose Text is "%", not a hash; `$a / $b` is division, not a
// pattern; `$a < $c` is a comparison, not a readline. The lexer settled all
// three by position (§3.2) and the tree records the answer, so reading the
// tree cannot reach a different one. Scanning source text could.
func decidedMarkers(stmt *Node) []decided {
	var found []decided

	// One entry per OCCURRENCE, not per kind. CompareFacts consumes a decided
	// site for each of perl's (`compare_facts.go:224`, `owner.decided[m]--`),
	// so a statement holding two anonymous hashes and reporting one site
	// scores the first exact and the second WRONG.
	//
	// Deduplicating here was the cause of 11 of the 20 WRONG files in T2's
	// first subject measurement -- `eq_hash({$o->h}, {qw( the hash )})` in
	// t/class/accessor.t:31 is two anonhash ops to perl and was one site here.
	add := func(kind string, n *Node) {
		found = append(found, decided{kind: kind, node: n})
	}

	var walk func(*Node)
	walk = func(n *Node) {
		switch n.Kind {
		case Unknown:
			// A nested Unknown speaks its own refusal, at its own span. The
			// statement around it was READ -- `BEGIN { ... }` is a Phaser
			// holding a Block -- so nothing else hedges these bytes, and
			// silence here is the same silent wrong answer the top-level
			// rule exists to prevent: perl found a marker inside, innermost
			// found a statement offering neither a site nor a hedge, and
			// the verdict was WRONG rather than wider.
			//
			// Do not descend: nothing inside a refusal is decided.
			for _, kind := range markerKinds {
				found = append(found, decided{kind: kind, node: n, unresolved: true})
			}
			return

		case AnonHash:
			// perl: anonhash. The lexer's brace stack already chose term
			// over block (§4.9.2), so this node IS the decision.
			add(parseoracle.SiteKindAnonhash, n)

		case Unary:
			// perl: srefgen. prefixName maps `\` to "ref", which keeps the
			// prefix spelling distinct from infix `-`.
			if n.Text == "ref" {
				add(parseoracle.SiteKindReference, n)
			}

		case Binary:
			// perl: match. `=~` and `!~` bind a pattern; s/// and tr/// are
			// substitution and transliteration, which perl reports with
			// different ops, so they are excluded by their own spelling.
			if n.Text == "=~" || n.Text == "!~" {
				if len(n.Children) > 1 && isMatchOperand(n.Children[1]) {
					add(parseoracle.SiteKindMatch, n)
				}
			}

		case Term:
			switch {
			case isHashTerm(n.Text):
				// perl: rv2hv.
				add(parseoracle.SiteKindHash, n)
			case isReadlineTerm(n.Text):
				// perl: readline.
				add(parseoracle.SiteKindReadline, n)
			case isBarePattern(n.Text):
				add(parseoracle.SiteKindMatch, n)
			}
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(stmt)
	return found
}

// isHashTerm reports whether a leaf's text is a hash read: `%h`, `%$r`.
//
// A leading `%` in TERM position is a sigil and nowhere else -- scanVariable
// refuses it when the expect state does not want a term, which is what keeps
// `$a % $b` from lexing as a hash (§3.2). So the byte is sufficient here
// precisely because the lexer already did the hard part.
func isHashTerm(text string) bool {
	return len(text) > 1 && text[0] == '%'
}

// isReadlineTerm reports whether a leaf is `<FH>`, `<$fh>` or `<>`.
//
// Same reasoning: scanAngle emits a Readline token only in term position,
// so a leaf whose text is angle-delimited is the lexer's own decision.
func isReadlineTerm(text string) bool {
	return len(text) >= 2 && text[0] == '<' && text[len(text)-1] == '>'
}

// isBarePattern reports whether a leaf is a match rather than a substitution
// or a transliteration.
//
// perl reports s/// with subst and tr/// with trans, neither of which is the
// match marker, so they must not be counted. `qr//` compiles a pattern
// without matching it and is excluded for the same reason.
func isBarePattern(text string) bool {
	switch {
	case quoteOpIs(text, "s"), quoteOpIs(text, "tr"),
		quoteOpIs(text, "y"), quoteOpIs(text, "qr"):
		return false
	case quoteOpIs(text, "m"):
		return true
	case len(text) >= 2 && text[0] == '/' && text[len(text)-1] == '/':
		return true
	}
	return false
}

// quoteOpIs reports whether a leaf is the quote-like operator `name` with
// any delimiter.
//
// perl takes the next non-whitespace character as the delimiter with no
// allow-list, so enumerating delimiters is enumerating a set that has no
// end. Measured on 5.42.0:
//
//	$ perl -MO=Concise,-exec -e 'my $a = m(x); my $b = m[x]; my $c = m!x!;'
//	-- three match ops
//
// The delimiter must not be a word character, or `qw(a b)` reads as `q`
// with `w` for a delimiter and `tr` as `t` with `r`. The lexer already
// settled this the same way (quote.go: "a word character immediately after
// the keyword means it is part of a longer name"), so this asks the same
// question of the token it produced.
//
// Whitespace needs no special case. perl's skipspace runs before delimiter
// selection, so `m /x/` is a match -- and a space is not a word character,
// so it passes this test on its way to being the delimiter's stand-in.
// Measured:
//
//	$ perl -MO=Concise,-exec -e 'my $a = m /x/;'
//	3  </> match(/"x"/) s
//
// comp/opsubs.t:119 is `isnt( m('unqualified'), ... )`, which scored WRONG
// while only `m/` and `m{` counted, and `s(a)(b)` decided MATCH -- a false
// positive -- while only `s/` and `s{` were excluded.
func quoteOpIs(text, name string) bool {
	if !strings.HasPrefix(text, name) {
		return false
	}
	i := len(name)
	return i < len(text) && !isWordByte(text[i])
}

// isWordByte is perl's \w for an identifier: a quote operator's delimiter
// may not be one, because the byte would belong to a longer name.
func isWordByte(c byte) bool {
	return c == '_' || ('0' <= c && c <= '9') ||
		('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

// isMatchOperand reports whether the right side of `=~` is a match rather
// than a substitution or transliteration.
//
// A VARIABLE on the right is a match too: `$got =~ $expected` compiles
// $expected as a pattern, and perl reports it with the same match op. That
// is not a guess about the variable's contents -- `=~` against anything that
// is not s/// or tr/// is a match, whatever the pattern turns out to be.
// Measured on perl 5.42.0:
//
//	$ perl -MO=Concise,-exec -e 'my ($g,$e); my $r = $g =~ $e;'
//	7  </> match()[$g:1,3] sK
//
// t/comp/use.t:24 and t/comp/uproto.t:24 are this shape and scored WRONG
// while only literals counted.
func isMatchOperand(n *Node) bool {
	if n.Kind != Term {
		return false
	}
	// A quote-like operand states its own kind: s/// substitutes and tr///
	// transliterates, and perl reports each with a different op.
	if isSubstOrTrans(n.Text) {
		return false
	}
	return true
}

// isSubstOrTrans reports whether a leaf is a substitution or a
// transliteration rather than a match.
func isSubstOrTrans(text string) bool {
	return quoteOpIs(text, "s") || quoteOpIs(text, "tr") || quoteOpIs(text, "y")
}

// lineStarts is the byte offset of each line's first byte, so a span can be
// reported in the lines the oracle speaks.
type lineStarts []int

func lineIndex(src []byte) lineStarts {
	starts := lineStarts{0}
	for i := 0; ; {
		j := bytes.IndexByte(src[i:], '\n')
		if j < 0 {
			break
		}
		i += j + 1
		starts = append(starts, i)
	}
	return starts
}

// at returns the 1-based line number containing byte offset off.
func (ls lineStarts) at(off int) int {
	if off < 0 {
		return 1
	}
	// Binary search would be faster; a linear scan is fine for the sizes
	// involved and is obviously correct.
	line := 1
	for i, s := range ls {
		if s > off {
			break
		}
		line = i + 1
	}
	return line
}
