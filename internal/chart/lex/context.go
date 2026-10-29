// ABOUTME: The context facts the lexer consumes, owned by chart/lex: declared subs, constant subs, bareword operators.
// ABOUTME: Derived from the token stream as the production lexer derives them; the chart reads them and never writes them.

package lex

import "tamarou.com/pvm/internal/lexer"

// BarewordOperators are the operators whose left operand is lexed as a word.
// They are the production lexer's, which internal/parse holds equal to
// CORE.tpm's `:bareword` operators.
var BarewordOperators = lexer.BarewordOperators

// Context is the facts a source's tokens declare, each with the byte offset
// of the token that declared it. The chart reads them; nothing feeds them
// back (spec §3 rule 2).
type Context struct {
	declared map[string]int // `sub NAME`
	constant map[string]int // names declared with the empty prototype
}

// Declared reports whether a `sub name` before byte at declares name.
func (c *Context) Declared(name string, at int) bool {
	off, ok := c.declared[name]
	return ok && off < at
}

// Constant reports whether name is declared with the empty prototype before
// byte at, which is what makes the production lexer expect an operator after
// it: `use constant NAME`, each key of `use constant { K => ... }`, and
// `sub NAME ()`.
func (c *Context) Constant(name string, at int) bool {
	off, ok := c.constant[name]
	return ok && off < at
}

// The states of facts' constant-sub machine, the production lexer's
// noteConstSub states.
const (
	constNone      = iota
	constSawUse    // `use`
	constSawPragma // `use constant`
	constInHash    // `use constant {`
	constSawSub    // `sub`
	constSawName   // `sub NAME`
)

// facts derives the context facts from toks, every token in source order,
// as the production lexer's noteConstSub derives constSubs: whitespace and
// comments are skipped, and any other token moves the machine.
func facts(src []byte, toks []lexer.Token) *Context {
	c := &Context{declared: map[string]int{}, constant: map[string]int{}}
	note := func(m map[string]int, name string, at int) {
		if _, ok := m[name]; !ok {
			m[name] = at
		}
	}
	state, depth, key := constNone, 0, ""
	for _, t := range toks {
		k := t.Kind
		if k == lexer.Whitespace || k == lexer.Comment {
			continue
		}
		text := string(src[t.Start:t.End])
		quoted := k == lexer.Quote && len(text) >= 2 && (text[0] == '\'' || text[0] == '"')
		switch state {
		case constSawUse:
			if k == lexer.Word && text == "constant" {
				state = constSawPragma
				continue
			}
		case constSawPragma:
			switch {
			case k == lexer.Word:
				note(c.constant, text, t.Start)
			case quoted:
				note(c.constant, text[1:len(text)-1], t.Start)
			case text == "{":
				state, depth, key = constInHash, 1, ""
				continue
			}
		case constInHash:
			switch {
			case text == "{" || text == "[" || text == "(":
				depth++
			case k == lexer.CloseBracket:
				if depth--; depth == 0 {
					state = constNone
				}
			case depth == 1 && text == "=>" && key != "":
				note(c.constant, key, t.Start)
			case depth == 1 && k == lexer.Word:
				key = text
				continue
			case depth == 1 && quoted:
				key = text[1 : len(text)-1]
				continue
			}
			key = ""
			continue
		case constSawSub:
			if k == lexer.Word {
				note(c.declared, text, t.Start)
				state, key = constSawName, text
				continue
			}
		case constSawName:
			if k == lexer.Prototype && text == "()" {
				note(c.constant, key, t.Start)
			}
		}
		switch {
		case k == lexer.Word && text == "use":
			state = constSawUse
		case k == lexer.Word && text == "sub":
			state = constSawSub
		default:
			state = constNone
		}
	}
	return c
}
