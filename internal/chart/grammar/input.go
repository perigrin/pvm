// ABOUTME: The generator's input types: symbols, rules and the grammar BuildTables reads.
// ABOUTME: Nothing here knows a grammar's source format; a test or the port states them as Go data.

package grammar

import "strings"

// SymbolKind says how a right-hand-side symbol was written.
type SymbolKind int

const (
	// Name is a bare or <bracketed> symbol name.
	Name SymbolKind = iota
	// Literal is a single-quoted string, Text without the quotes.
	Literal
	// Class is a character class, Text including its brackets.
	Class
)

// Symbol is one right-hand-side element.
type Symbol struct {
	Kind SymbolKind
	Text string
}

// Seq marks a sequence rule, whose RHS is its single repeated symbol.
type Seq int

const (
	NotSeq Seq = iota
	Star       // X ::= Y*
	Plus       // X ::= Y+
)

// Rule is one alternative of a statement.
type Rule struct {
	LHS  string
	RHS  []Symbol
	Seq  Seq
	Line int // line of the alternative's first symbol
	Decl int // line of the statement's left-hand side
	// Cite is the perly.y alternative the port's rule answers to: "perly.y:N",
	// or "perly.y:N, fitted: <what the lexer emits instead>". Empty for toys.
	Cite string
}

// Grammar is the generator's input: G1 rules, and G0 rules whose left-hand
// sides are the terminals. G1 and G0 rules are never merged, even when they
// share a left-hand side.
type Grammar struct {
	G1, G0   []Rule
	Priority map[string]int // :lexeme ~ X priority => n
	Discard  []string       // :discard ~ X
	LATM     bool           // lexeme default = latm => 1
}

func ruleText(r Rule, op string) string { return nameText(r.LHS) + " " + op + " " + rhsText(r) }

func rhsText(r Rule) string {
	var parts []string
	for _, s := range r.RHS {
		parts = append(parts, symbolText(s))
	}
	s := strings.Join(parts, " ")
	switch r.Seq {
	case Star:
		s += "*"
	case Plus:
		s += "+"
	}
	return s
}

func symbolText(s Symbol) string {
	switch s.Kind {
	case Literal:
		return "'" + s.Text + "'"
	case Class:
		return s.Text
	}
	return nameText(s.Text)
}

// nameText writes a name bracketed if it has a space.
func nameText(n string) string {
	if strings.Contains(n, " ") {
		return "<" + n + ">"
	}
	return n
}
