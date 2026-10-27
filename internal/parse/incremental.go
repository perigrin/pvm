// ABOUTME: Edit, ReparseIncremental, the IncrementalRepair flag, edit triage and widening (spec §6.5, §7.7).
// ABOUTME: ReparseIncremental is still a full re-parse; triage and widen are what Repair will start from.
package parse

import (
	"bytes"
	"sort"
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

// IncrementalRepair switches the LSP from a full re-parse on every change to
// ReparseIncremental's Repair (spec §6.5.5). Triage and widening are reached
// only through Repair, so they too sit behind it. Tests may set it locally.
var IncrementalRepair = false

// Edit replaces the old source's bytes [Start, OldEnd) with NewText.
type Edit struct {
	Start, OldEnd int
	NewText       string
}

// ReparseIncremental returns the tree of src, the old source with e applied,
// reusing old where it can. It equals Parse(src) under TreeDiff (§7.7.1).
// Until Repair lands it reuses nothing.
func ReparseIncremental(old *Node, e Edit, src []byte) *Node {
	return Parse(src)
}

// ReparseIncremental is the resolver form: its result equals a fresh
// NewSession().ParseSourceFrom(path, src, roots...). It parses in a fresh
// Session, never s's module facts, because a Session's second parse reports
// none of the LoadedModules, DeclarationErrors or DeclarationWarnings its
// first already read, and a module edited on disk must be read again.
func (s *Session) ReparseIncremental(old *Node, e Edit, path string, src []byte, roots ...string) *Node {
	return NewSession().ParseSourceFrom(path, src, roots...)
}

// Strategy is how much work an edit needs (spec §6.5.1).
type Strategy int

const (
	StrategyReuseAll Strategy = iota // interior to one comment or whitespace token: shift offsets only
	StrategyLocal                    // re-lex and re-parse the region widen returns
	StrategyFull                     // re-parse the whole file
)

func (s Strategy) String() string {
	return [...]string{"StrategyReuseAll", "StrategyLocal", "StrategyFull"}[s]
}

const (
	maxEditBytes  = 64 * 1024 // deleted or inserted, whichever is larger
	maxLocalEdit  = 1024
	maxLocalLines = 10 // newlines in the inserted text
)

// globalTriggers are byte sequences whose count changing in an edit's line
// window can change the parse of text arbitrarily far away (§6.1.2). A
// count, not presence: deleting one `#` of two on a line is still
// structural. Over-approximate freely; a false positive costs a full parse.
var globalTriggers = [][]byte{
	[]byte("<<"),       // heredoc introducer (A2, A3)
	[]byte("__END__"),  // A8
	[]byte("__DATA__"), // A8
	[]byte("=cut"),     // A7
	[]byte("#"),        // A9
	[]byte("use "),     // B2, B3, B7, B8
	[]byte("no "),      // B2
	[]byte("package"),  // B5
	[]byte("class"),    // B5
	[]byte("sub"),      // B1, B4
	[]byte("format"),   // A10
	[]byte("BEGIN"),    // B6
}

// triage classifies e, an edit of src whose tree is old: the tree the
// document's Session parsed, resolver and all.
func triage(old *Node, src []byte, e Edit) Strategy {
	size := max(e.OldEnd-e.Start, len(e.NewText))
	if size > maxEditBytes ||
		introducesGlobalTrigger(src, e) || touchesDeclaration(old, e) ||
		size > maxLocalEdit || strings.Count(e.NewText, "\n") > maxLocalLines {
		return StrategyFull
	}
	if isInteriorTrivia(src, e) {
		return StrategyReuseAll
	}
	return StrategyLocal
}

// introducesGlobalTrigger compares the whole lines e touches before and
// after it: a trigger's count, a POD directive, quote counts and bracket
// balance (§6.1.2 A1, A4, A7, C1, C2).
func introducesGlobalTrigger(src []byte, e Edit) bool {
	before, after := lineWindows(src, e)
	for _, t := range globalTriggers {
		if bytes.Count(before, t) != bytes.Count(after, t) {
			return true
		}
	}
	if hasPodDirective(before) || hasPodDirective(after) {
		return true
	}
	for _, q := range []byte("\"'`") {
		if bytes.Count(before, []byte{q}) != bytes.Count(after, []byte{q}) {
			return true
		}
	}
	for _, pair := range []string{"{}", "()", "[]"} {
		if balance(before, pair) != balance(after, pair) {
			return true
		}
	}
	return false
}

// lineBounds is the span of the whole lines holding src's bytes [start,
// end): from a line start to just past a '\n', or to a buffer edge.
func lineBounds(src []byte, start, end int) (int, int) {
	ls := bytes.LastIndexByte(src[:start], '\n') + 1
	le := len(src)
	if i := bytes.IndexByte(src[end:], '\n'); i >= 0 {
		le = end + i + 1
	}
	return ls, le
}

// lineWindows is the whole lines e touches in src, before and after e.
func lineWindows(src []byte, e Edit) (before, after []byte) {
	ls, le := lineBounds(src, e.Start, e.OldEnd)
	return src[ls:le], append(append(append([]byte{}, src[ls:e.Start]...), e.NewText...), src[e.OldEnd:le]...)
}

// hasPodDirective reports a line of w that starts `=` and a letter, which
// opens or continues POD wherever a statement may start (§6.1.2 A7). Any
// edit on such a line is full, since `=pod` to `=cutx` changes no count.
func hasPodDirective(w []byte) bool {
	for line := range bytes.SplitSeq(w, []byte{'\n'}) {
		if len(line) > 1 && line[0] == '=' && (line[1]|0x20 >= 'a' && line[1]|0x20 <= 'z') {
			return true
		}
	}
	return false
}

func balance(w []byte, pair string) int {
	return bytes.Count(w, []byte{pair[0]}) - bytes.Count(w, []byte{pair[1]})
}

// touchesDeclaration reports e touching what can change how later text
// parses, which no line window sees (§7.7.3's prototype row): a `use` or
// `no`, a BEGIN block, a prototype, or a sub, package or class declaration
// anywhere but strictly inside its block. Touching includes abutting.
func touchesDeclaration(n *Node, e Edit) bool {
	if n.End < e.Start || n.Start > e.OldEnd {
		return false
	}
	switch {
	case n.Kind == Use, n.Kind == PrototypeNode, n.Kind == Phaser && n.Text == "BEGIN":
		return true
	case n.Kind == Declaration && (n.Text == "sub" || n.Text == "package" || n.Text == "class"):
		if !inBody(n, e) {
			return true
		}
	}
	for _, c := range n.Children {
		if touchesDeclaration(c, e) {
			return true
		}
	}
	return false
}

// inBody reports e strictly inside the braces of declaration d's block.
func inBody(d *Node, e Edit) bool {
	if len(d.Children) == 0 {
		return false
	}
	b := d.Children[len(d.Children)-1]
	return b.Kind == Block && b.Start < e.Start && e.OldEnd < b.End
}

// isInteriorTrivia reports e strictly inside one comment, POD or whitespace
// token, deleting and inserting no newline, and inserting only blanks into
// whitespace: the token stays one token of its kind and no line moves.
// A string's body is excluded although §6.5.1 lists it: its node's Text
// changes, so shifting offsets alone would not give the fresh tree.
func isInteriorTrivia(src []byte, e Edit) bool {
	if bytes.IndexByte(src[e.Start:e.OldEnd], '\n') >= 0 || strings.IndexByte(e.NewText, '\n') >= 0 {
		return false
	}
	toks, _ := lexOld(src)
	i := sort.Search(len(toks), func(i int) bool { return toks[i].End > e.Start })
	if i == len(toks) || !(toks[i].Start < e.Start && e.OldEnd < toks[i].End) {
		return false
	}
	switch toks[i].Kind {
	case lexer.Comment, lexer.Pod:
		return true
	case lexer.Whitespace:
		return strings.Trim(e.NewText, " \t") == ""
	}
	return false
}

// lexOld lexes a program source as Parse does, with its safe checkpoints.
//
// shortcut: re-lexes the old source on each call; read the document's token
// cache (§6.3.2) once it exists.
func lexOld(src []byte) ([]lexer.Token, []lexer.Checkpoint) {
	return lexer.Checkpoints(src, lexer.StartStateBarewords(coreBarewordOperators()))
}

// widen is §6.5.2: the lexer checkpoint to re-lex from and the node to
// re-parse whole for a StrategyLocal edit of src, whose tree is old.
//
// The anchor is the top-level statement holding the edit when the edit is
// strictly inside a sub's body, and old itself otherwise: the file top
// level, a `use`, a BEGIN block and a `package NAME;` are never anchors.
// A sub inside a package block, bare block or `if` anchors at that whole
// statement, past the block, because a nested resume stops at the block's
// `}` with no post-block state, and a top-level one runs to end of file.
// The checkpoint is the last at or before the anchor's start.
func widen(old *Node, src []byte, e Edit) (lexer.Checkpoint, *Node) {
	anchor := old
	for _, top := range old.Children {
		if top.Start <= e.Start && e.OldEnd <= top.End && inSubBody(top, e) {
			anchor = top
			break
		}
	}
	_, cps := lexOld(src)
	i := sort.Search(len(cps), func(i int) bool { return cps[i].Byte > anchor.Start })
	return cps[i-1], anchor
}

// inSubBody reports that the path from n down to e enters some sub's body
// and passes through no `use`, BEGIN block or declaration head.
func inSubBody(n *Node, e Edit) bool {
	found := false
	for n != nil {
		switch {
		case n.Kind == Use, n.Kind == Phaser && n.Text == "BEGIN":
			return false
		case n.Kind == Declaration && n.Text == "sub":
			if !inBody(n, e) {
				return false
			}
			found = true
		}
		n = childHolding(n, e)
	}
	return found
}

// childHolding is n's first child spanning all of e, or nil.
func childHolding(n *Node, e Edit) *Node {
	for _, c := range n.Children {
		if c.Start <= e.Start && e.OldEnd <= c.End {
			return c
		}
	}
	return nil
}
