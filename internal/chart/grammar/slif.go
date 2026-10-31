// ABOUTME: Reads the subset of Marpa's SLIF notation that Guacamole's grammar uses.
// ABOUTME: Produces G1 and G0 rules kept apart, sequences as sequences, and lexeme priorities.

package grammar

import (
	"fmt"
	"strconv"
	"strings"
)

// Error is a read failure at one line, counted from 1.
type Error struct {
	File string
	Line int
	Msg  string
}

func (e *Error) Error() string { return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Msg) }

// ReadSLIF reads src, naming file in errors. Anything outside the subset is
// an error naming its line; nothing is skipped.
func ReadSLIF(file string, src []byte) (*Grammar, error) {
	toks, err := scan(file, string(src))
	if err != nil {
		return nil, err
	}
	r := &reader{file: file, toks: toks, g: &Grammar{Priority: map[string]int{}}}
	for r.peek(0).kind != tEOF {
		if err := r.statement(); err != nil {
			return nil, err
		}
	}
	if len(r.g.G1)+len(r.g.G0) == 0 {
		return nil, r.errf(1, "no rules")
	}
	return r.g, nil
}

type tokKind int

const (
	tEOF     tokKind = iota
	tName            // Foo
	tBracket         // <foo bar>, text normalised to "foo bar"
	tLiteral         // 'x', text without quotes
	tClass           // [a-z], text with brackets
	tPseudo          // :lexeme
	tAction          // ::first
	tInt             // -2
	tOp              // ::= ~ | || => = * +
)

type token struct {
	kind tokKind
	text string
	line int
}

func isNameByte(c byte, first bool) bool {
	return c == '_' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || !first && '0' <= c && c <= '9'
}

func isDigit(c byte) bool { return '0' <= c && c <= '9' }

func scan(file, s string) ([]token, error) {
	var toks []token
	line := 1
	fail := func(msg string, args ...any) ([]token, error) {
		return nil, &Error{File: file, Line: line, Msg: fmt.Sprintf(msg, args...)}
	}
	// until returns the index of the first c at or after i on this line, or -1.
	until := func(i int, c byte) int {
		for ; i < len(s) && s[i] != '\n'; i++ {
			if s[i] == c {
				return i
			}
		}
		return -1
	}
	nameEnd := func(i int) int {
		for i < len(s) && isNameByte(s[i], false) {
			i++
		}
		return i
	}
	for i := 0; i < len(s); {
		c, start := s[i], i
		text := ""
		var kind tokKind
		switch {
		case c == '\n':
			line++
			i++
			continue
		case c == ' ' || c == '\t' || c == '\r':
			i++
			continue
		case c == '#':
			for i < len(s) && s[i] != '\n' {
				i++
			}
			continue
		case isNameByte(c, true):
			kind, i = tName, nameEnd(i)
		case c == '<':
			end := until(i, '>')
			if end < 0 {
				return fail("unterminated <name>")
			}
			kind, text, i = tBracket, strings.Join(strings.Fields(s[i+1:end]), " "), end+1
		case c == '\'':
			end := until(i+1, '\'')
			if end < 0 {
				return fail("unterminated quoted string")
			}
			kind, text, i = tLiteral, s[i+1:end], end+1
		case c == '[':
			for i++; i < len(s) && s[i] != ']' && s[i] != '\n'; i++ {
				if s[i] == '\\' && i+1 < len(s) && s[i+1] != '\n' {
					i++
				}
			}
			if i >= len(s) || s[i] != ']' {
				return fail("unterminated character class")
			}
			kind, i = tClass, i+1
		case strings.HasPrefix(s[i:], "::="):
			kind, i = tOp, i+3
		case strings.HasPrefix(s[i:], "::") && i+2 < len(s) && isNameByte(s[i+2], true):
			kind, i = tAction, nameEnd(i+2)
		case c == ':' && i+1 < len(s) && isNameByte(s[i+1], true):
			kind, i = tPseudo, nameEnd(i+1)
		case isDigit(c) || c == '-' && i+1 < len(s) && isDigit(s[i+1]):
			for i++; i < len(s) && isDigit(s[i]); i++ {
			}
			kind = tInt
		case strings.HasPrefix(s[i:], "=>") || strings.HasPrefix(s[i:], "||"):
			kind, i = tOp, i+2
		case strings.IndexByte("~|=*+", c) >= 0:
			kind, i = tOp, i+1
		default:
			return fail("unsupported character %q", c)
		}
		if kind != tBracket && kind != tLiteral {
			text = s[start:i]
		}
		toks = append(toks, token{kind, text, line})
	}
	return toks, nil
}

type reader struct {
	file        string
	toks        []token
	i           int
	g           *Grammar
	seenDefault bool
}

func (r *reader) errf(line int, msg string, args ...any) error {
	return &Error{File: r.file, Line: line, Msg: fmt.Sprintf(msg, args...)}
}

func (r *reader) peek(n int) token {
	if r.i+n < len(r.toks) {
		return r.toks[r.i+n]
	}
	line := 1
	if len(r.toks) > 0 {
		line = r.toks[len(r.toks)-1].line
	}
	return token{tEOF, "end of input", line}
}

func (r *reader) isOp(n int, text string) bool {
	t := r.peek(n)
	return t.kind == tOp && t.text == text
}

func isName(t token) bool { return t.kind == tName || t.kind == tBracket }

func isSymbol(t token) bool { return isName(t) || t.kind == tLiteral || t.kind == tClass }

// startsRule reports whether the token at n is the left-hand side of a rule.
func (r *reader) startsRule(n int) bool {
	return isName(r.peek(n)) && (r.isOp(n+1, "::=") || r.isOp(n+1, "~"))
}

func (r *reader) statement() error {
	t := r.peek(0)
	switch {
	case t.kind == tName && t.text == "lexeme" && r.peek(1).text == "default" && r.isOp(2, "="):
		r.i += 3
		adv, err := r.adverbs()
		if err != nil {
			return err
		}
		if len(adv) != 1 || adv[0].key != "latm" || (adv[0].val != "0" && adv[0].val != "1") {
			return r.errf(t.line, "lexeme default supports only latm => 0 or 1")
		}
		if r.seenDefault {
			return r.errf(t.line, "second lexeme default")
		}
		r.seenDefault, r.g.LATM = true, adv[0].val == "1"
		return nil
	case t.kind == tPseudo && t.text == ":lexeme":
		sym, err := r.pseudoHead(t)
		if err != nil {
			return err
		}
		adv, err := r.adverbs()
		if err != nil {
			return err
		}
		if len(adv) != 1 || adv[0].key != "priority" || adv[0].kind != tInt {
			return r.errf(t.line, ":lexeme supports only priority => <integer>")
		}
		if _, dup := r.g.Priority[sym]; dup {
			return r.errf(t.line, "second :lexeme for %s", sym)
		}
		n, err := strconv.Atoi(adv[0].val)
		if err != nil {
			return r.errf(t.line, "priority %s is out of range", adv[0].val)
		}
		r.g.Priority[sym] = n
		return nil
	case t.kind == tPseudo && t.text == ":discard":
		sym, err := r.pseudoHead(t)
		if err != nil {
			return err
		}
		r.g.Discard = append(r.g.Discard, sym)
		return nil
	case t.kind == tPseudo && t.text == ":default":
		r.i++
		if !r.isOp(0, "::=") {
			return r.errf(t.line, ":default supports only ::=")
		}
		r.i++
		return r.ignoredAdverbs()
	case r.startsRule(0):
		return r.rule()
	}
	return r.errf(t.line, "unsupported statement starting %q", t.text)
}

// pseudoHead reads "~ <symbol>" after a pseudo-symbol and returns the symbol.
func (r *reader) pseudoHead(t token) (string, error) {
	r.i++
	sym := r.peek(1)
	if !r.isOp(0, "~") || !isName(sym) {
		return "", r.errf(t.line, "%s needs ~ and a symbol name", t.text)
	}
	r.i += 2
	return sym.text, nil
}

type adverb struct {
	key, val string
	kind     tokKind
	line     int
}

func (r *reader) adverbs() ([]adverb, error) {
	var out []adverb
	for r.peek(0).kind == tName && r.isOp(1, "=>") {
		key, arrow, val := r.peek(0), r.peek(1), r.peek(2)
		r.i += 2
		if !isSymbol(val) && val.kind != tAction && val.kind != tInt || r.startsRule(0) {
			return nil, r.errf(arrow.line, "%s => has no value", key.text)
		}
		r.i++
		out = append(out, adverb{key.text, val.text, val.kind, key.line})
	}
	return out, nil
}

// ignoredAdverbs reads adverbs that do not change a rule: action and name.
func (r *reader) ignoredAdverbs() error {
	adv, err := r.adverbs()
	for _, a := range adv {
		if a.key != "action" && a.key != "name" {
			return r.errf(a.line, "unsupported adverb %s", a.key)
		}
	}
	return err
}

func (r *reader) rule() error {
	lhs, op := r.peek(0), r.peek(1)
	r.i += 2
	g1 := op.text == "::="
	emptyLine, emptyMsg := op.line, lhs.text+" has an empty right-hand side"
	for first := true; ; first = false {
		rule := Rule{LHS: lhs.text, Decl: lhs.line}
		for isSymbol(r.peek(0)) && !r.startsRule(0) && !r.isOp(1, "=>") {
			t := r.peek(0)
			r.i++
			if rule.Line == 0 {
				rule.Line = t.line
			}
			kind := Name
			switch t.kind {
			case tLiteral:
				kind = Literal
			case tClass:
				kind = Class
			}
			rule.RHS = append(rule.RHS, Symbol{kind, t.text})
		}
		if len(rule.RHS) == 0 {
			return r.errf(emptyLine, "%s", emptyMsg)
		}
		if q := r.peek(0); r.isOp(0, "*") || r.isOp(0, "+") {
			r.i++
			rule.Seq = Star
			if q.text == "+" {
				rule.Seq = Plus
			}
			if len(rule.RHS) != 1 || !first || isSymbol(r.peek(0)) && !r.startsRule(0) {
				return r.errf(q.line, "a sequence rule has one symbol and no alternatives")
			}
		}
		if g1 {
			if err := r.ignoredAdverbs(); err != nil {
				return err
			}
			r.g.G1 = append(r.g.G1, rule)
		} else {
			r.g.G0 = append(r.g.G0, rule)
		}
		if !r.isOp(0, "|") {
			return nil
		}
		if rule.Seq != NotSeq {
			return r.errf(r.peek(0).line, "a sequence rule has one symbol and no alternatives")
		}
		emptyLine, emptyMsg = r.peek(0).line, "empty alternative after |"
		r.i++
	}
}
