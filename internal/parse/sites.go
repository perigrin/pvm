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
		for _, kind := range decidedMarkers(n) {
			sites = append(sites, parseoracle.SubjectCallSite{
				Kind:    kind,
				Line:    start,
				EndLine: end,
			})
		}
	}
	return sites
}

// decidedMarkers reports which of the five markers this statement contains,
// in the subject vocabulary. Each is a conclusion the parser committed to.
//
// Only what the tree SHOWS, never what the bytes suggest. `$a % $b` is a
// Binary whose Text is "%", not a hash; `$a / $b` is division, not a
// pattern; `$a < $c` is a comparison, not a readline. The lexer settled all
// three by position (§3.2) and the tree records the answer, so reading the
// tree cannot reach a different one. Scanning source text could.
func decidedMarkers(stmt *Node) []string {
	var found []string
	seen := make(map[string]bool, len(markerKinds))

	add := func(kind string) {
		if !seen[kind] {
			seen[kind] = true
			found = append(found, kind)
		}
	}

	var walk func(*Node)
	walk = func(n *Node) {
		switch n.Kind {
		case Unknown:
			// A nested Unknown is the statement's own refusal, already
			// hedged at statement level. Do not descend into it, and do not
			// let anything inside it count as decided.
			return

		case AnonHash:
			// perl: anonhash. The lexer's brace stack already chose term
			// over block (§4.9.2), so this node IS the decision.
			add(parseoracle.SiteKindAnonhash)

		case Unary:
			// perl: srefgen. prefixName maps `\` to "ref", which keeps the
			// prefix spelling distinct from infix `-`.
			if n.Text == "ref" {
				add(parseoracle.SiteKindReference)
			}

		case Binary:
			// perl: match. `=~` and `!~` bind a pattern; s/// and tr/// are
			// substitution and transliteration, which perl reports with
			// different ops, so they are excluded by their own spelling.
			if n.Text == "=~" || n.Text == "!~" {
				if len(n.Children) > 1 && isMatchOperand(n.Children[1]) {
					add(parseoracle.SiteKindMatch)
				}
			}

		case Term:
			switch {
			case isHashTerm(n.Text):
				// perl: rv2hv.
				add(parseoracle.SiteKindHash)
			case isReadlineTerm(n.Text):
				// perl: readline.
				add(parseoracle.SiteKindReadline)
			case isBarePattern(n.Text):
				add(parseoracle.SiteKindMatch)
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
	case strings.HasPrefix(text, "s/"), strings.HasPrefix(text, "s{"),
		strings.HasPrefix(text, "tr"), strings.HasPrefix(text, "y/"),
		strings.HasPrefix(text, "qr"):
		return false
	case strings.HasPrefix(text, "m/"), strings.HasPrefix(text, "m{"):
		return true
	case len(text) >= 2 && text[0] == '/' && text[len(text)-1] == '/':
		return true
	}
	return false
}

// isMatchOperand reports whether the right side of `=~` is a match rather
// than a substitution or transliteration.
func isMatchOperand(n *Node) bool {
	if n.Kind != Term {
		return false
	}
	return isBarePattern(n.Text)
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
