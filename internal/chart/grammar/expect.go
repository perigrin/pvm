// ABOUTME: Per-state expected-terminal sets, and the conflicting pairs observed by lexing probes in XOperator and XTerm.
// ABOUTME: A pair is the two terminals the cut maps the production lexer's two readings of one probe text to (spec §5).

package grammar

import (
	"fmt"
	"maps"
	"regexp/syntax"
	"slices"
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

// expected sets each state's Expected: the terminals after a dot in its
// items, in symbol order, which are its transitions on terminals.
func (t *Tables) expected() {
	for s := range t.States {
		st := &t.States[s]
		for _, tr := range st.Transitions {
			if t.Symbols[tr.Symbol].Terminal {
				st.Expected = append(st.Expected, tr.Symbol)
			}
		}
	}
}

// Conflict is one conflicting pair: the terminals the cut maps the lexer's
// XOperator and XTerm readings of Probe to, named by Text, where the two
// readings first diverge. A reading the cut maps to several terminals lists
// them joined by commas.
type Conflict struct {
	Text, Operator, Term, Probe string
}

func (c Conflict) String() string {
	return fmt.Sprintf("%s %s %s %q", c.Text, c.Operator, c.Term, c.Probe)
}

// Probe is one text the generator lexes, and the terminal it came from.
type Probe struct {
	Terminal, Text string
	Sample         bool // generated from a pattern predicate, not a literal
}

// The production lexer exposes no constructor for an expect state, so each
// state is reached through the lexer's own state machine: after a variable it
// expects an operator, and after `=` a term.
const (
	operatorPrefix = "$a "
	termPrefix     = "$a = "
)

// Probes is every literal text of a cut terminal and, for each terminal
// whose predicates are patterns, one short sample per printable ASCII
// character its texts can start with, in terminal order, then by sample.
func Probes(c *CutGrammar) []Probe {
	var out []Probe
	for _, name := range slices.Sorted(maps.Keys(c.Terminals)) {
		var samples []string
		for first := rune('!'); first <= '~'; first++ {
			best := ""
			for _, p := range c.Terminals[name] {
				if _, ok := literalText(p.Def); ok {
					continue
				}
				if s, ok := sampleFrom(parsed(p), first); ok && (best == "" || len(s) < len(best)) {
					best = s
				}
			}
			if best != "" {
				samples = append(samples, best)
			}
		}
		for _, p := range c.Terminals[name] {
			if lit, ok := literalText(p.Def); ok {
				out = append(out, Probe{name, lit, false})
			}
		}
		slices.Sort(samples)
		for _, s := range samples {
			out = append(out, Probe{name, s, true})
		}
	}
	return out
}

// FindConflicts lexes every probe once after operatorPrefix and once after
// termPrefix. Where the token streams differ, the first differing tokens are
// mapped to the cut's terminals; if both readings map to some terminal, that
// pair is a conflict, named by the two tokens' common prefix. The pairs come
// sorted, one per text and terminal pair, with the first probe that found it.
func FindConflicts(c *CutGrammar) []Conflict {
	seen := map[[3]string]bool{}
	var out []Conflict
	for _, p := range Probes(c) {
		op := lexAfter(operatorPrefix, p.Text)
		term := lexAfter(termPrefix, p.Text)
		i := 0
		for i < len(op) && i < len(term) && op[i] == term[i] {
			i++
		}
		if i == len(op) || i == len(term) {
			continue // the same, or one reading a prefix of the other: no differing token
		}
		opTerms, termTerms := c.terminalsOf(op[i]), c.terminalsOf(term[i])
		if opTerms == "" || termTerms == "" {
			continue // a reading the cut has no terminal for, such as an unterminated regex
		}
		k := [3]string{commonPrefix(op[i].text, term[i].text), opTerms, termTerms}
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, Conflict{k[0], k[1], k[2], p.Text})
	}
	slices.SortFunc(out, func(a, b Conflict) int { return strings.Compare(a.String(), b.String()) })
	return out
}

type tok struct {
	kind lexer.Kind
	text string
}

// lexAfter is the non-whitespace tokens of text, lexed after prefix.
func lexAfter(prefix, text string) []tok {
	src := prefix + text
	var out []tok
	for _, t := range lexer.Tokenize([]byte(src)) {
		if t.End <= len(prefix) || t.Kind == lexer.Whitespace {
			continue
		}
		out = append(out, tok{t.Kind, src[t.Start:t.End]})
	}
	return out
}

// terminalsOf is every terminal a predicate of which admits t, joined by
// commas in name order.
func (c *CutGrammar) terminalsOf(t tok) string {
	var names []string
	for name, preds := range c.Terminals {
		if slices.ContainsFunc(preds, func(p Pred) bool { return p.Match(t.kind, t.text) }) {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return strings.Join(names, ",")
}

func commonPrefix(a, b string) string {
	n := 0
	for n < len(a) && n < len(b) && a[n] == b[n] {
		n++
	}
	return a[:n]
}

func parsed(p Pred) *syntax.Regexp {
	re, err := syntax.Parse(p.text.String(), syntax.Perl)
	if err != nil {
		panic(fmt.Sprintf("grammar: reparsing %s: %v", p, err)) // it compiled once already
	}
	return re.Simplify()
}

// sample is a short text of re's language: each alternative's shortest
// sample, one pass through a repetition (none through whitespace, which
// only pads), and x from a class that holds it.
func sample(re *syntax.Regexp) string {
	switch re.Op {
	case syntax.OpLiteral:
		return string(re.Rune)
	case syntax.OpCharClass:
		return string(classRune(re.Rune))
	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		return "x"
	case syntax.OpCapture:
		return sample(re.Sub[0])
	case syntax.OpConcat:
		var b strings.Builder
		for _, s := range re.Sub {
			b.WriteString(sample(s))
		}
		return b.String()
	case syntax.OpAlternate:
		best := sample(re.Sub[0])
		for _, s := range re.Sub[1:] {
			if t := sample(s); len(t) < len(best) {
				best = t
			}
		}
		return best
	case syntax.OpStar, syntax.OpPlus, syntax.OpRepeat:
		s := sample(re.Sub[0])
		return strings.Repeat(s, passes(re, s))
	}
	return "" // empty match, anchors, ?
}

// passes is how many times sample takes a repetition whose one pass is s.
func passes(re *syntax.Regexp, s string) int {
	switch {
	case re.Op == syntax.OpRepeat && (re.Min > 0 || re.Max == 0):
		return re.Min
	case re.Op != syntax.OpPlus && strings.TrimSpace(s) == "":
		return 0
	}
	return 1
}

// sampleFrom is sample restricted to texts that start with first, and
// whether there is one.
func sampleFrom(re *syntax.Regexp, first rune) (string, bool) {
	switch re.Op {
	case syntax.OpLiteral:
		return string(re.Rune), len(re.Rune) > 0 && re.Rune[0] == first
	case syntax.OpCharClass:
		return string(first), inClass(re.Rune, first)
	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		return string(first), true
	case syntax.OpCapture:
		return sampleFrom(re.Sub[0], first)
	case syntax.OpQuest:
		return sampleFrom(re.Sub[0], first)
	case syntax.OpConcat:
		// The first sub that is not left empty starts the text.
		for i, sub := range re.Sub {
			if s, ok := sampleFrom(sub, first); ok {
				var b strings.Builder
				b.WriteString(s)
				for _, rest := range re.Sub[i+1:] {
					b.WriteString(sample(rest))
				}
				return b.String(), true
			}
			if !canBeEmpty(sub) {
				break
			}
		}
		return "", false
	case syntax.OpAlternate:
		best, found := "", false
		for _, sub := range re.Sub {
			if s, ok := sampleFrom(sub, first); ok && (!found || len(s) < len(best)) {
				best, found = s, true
			}
		}
		return best, found
	case syntax.OpStar, syntax.OpPlus, syntax.OpRepeat:
		if re.Op == syntax.OpRepeat && re.Max == 0 {
			return "", false
		}
		s, ok := sampleFrom(re.Sub[0], first)
		if !ok {
			return "", false
		}
		if n := passes(re, sample(re.Sub[0])); n > 1 {
			s += strings.Repeat(sample(re.Sub[0]), n-1)
		}
		return s, true
	}
	return "", false
}

func canBeEmpty(re *syntax.Regexp) bool {
	switch re.Op {
	case syntax.OpEmptyMatch, syntax.OpBeginLine, syntax.OpEndLine, syntax.OpBeginText, syntax.OpEndText,
		syntax.OpWordBoundary, syntax.OpNoWordBoundary, syntax.OpStar, syntax.OpQuest:
		return true
	case syntax.OpRepeat:
		return re.Min == 0 || canBeEmpty(re.Sub[0])
	case syntax.OpCapture, syntax.OpPlus:
		return canBeEmpty(re.Sub[0])
	case syntax.OpConcat:
		for _, s := range re.Sub {
			if !canBeEmpty(s) {
				return false
			}
		}
		return true
	case syntax.OpAlternate:
		return slices.ContainsFunc(re.Sub, canBeEmpty)
	}
	return false
}

func inClass(ranges []rune, r rune) bool {
	for i := 0; i+1 < len(ranges); i += 2 {
		if ranges[i] <= r && r <= ranges[i+1] {
			return true
		}
	}
	return false
}

// classRune picks x if the class holds it, else its first printable,
// non-space ASCII rune, else its first rune.
func classRune(ranges []rune) rune {
	if inClass(ranges, 'x') {
		return 'x'
	}
	for r := rune('!'); r <= '~'; r++ {
		if inClass(ranges, r) {
			return r
		}
	}
	return ranges[0]
}
