// ABOUTME: The generator's input types: symbols, rules and the grammar BuildTables reads.
// ABOUTME: Nothing here knows SLIF; slif.go fills them in from Guacamole's grammar, and a test can state them as Go data.

package grammar

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
}

// Grammar is a read SLIF grammar. G1 and G0 rules are never merged, even
// when they share a left-hand side.
type Grammar struct {
	G1, G0   []Rule
	Priority map[string]int // :lexeme ~ X priority => n
	Discard  []string       // :discard ~ X
	LATM     bool           // lexeme default = latm => 1
}
