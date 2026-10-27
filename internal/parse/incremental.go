// ABOUTME: Edit, ReparseIncremental, the IncrementalRepair flag, edit triage and widening (spec §6.5, §7.7).
// ABOUTME: ReparseIncremental is still a full re-parse; triage and widen are what Repair will start from.
package parse

import (
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

// triggerWords are the words whose arrival or departure can change how text
// arbitrarily far away parses (§6.1.2 B1-B7, A8, A10), matched as whole
// Word tokens so no spelling of the space after one escapes.
var triggerWords = map[string]bool{
	"use": true, "no": true, "require": true, "BEGIN": true, "sub": true,
	"package": true, "class": true, "format": true, "__END__": true, "__DATA__": true,
}

// triggerKinds are the token kinds that hold lexer state across lines:
// a comment (A9), POD (A7), a heredoc (A2, A3), a format body (A10), the
// data section (A8), a prototype (B1), and a construct left unterminated.
var triggerKinds = map[lexer.Kind]bool{
	lexer.Comment: true, lexer.Pod: true, lexer.HeredocOpen: true, lexer.HeredocBody: true,
	lexer.FormatBody: true, lexer.DataSection: true, lexer.Prototype: true, lexer.UnknownRest: true,
}

// triage classifies e, an edit of src whose tree is old: the tree the
// document's Session parsed, resolver and all. It decides on tokens, not
// text: src is lexed before and after e, and only the tokens that differ
// are examined.
func triage(old *Node, src []byte, e Edit) Strategy {
	size := max(e.OldEnd-e.Start, len(e.NewText))
	if size > maxEditBytes {
		return StrategyFull
	}
	d := diffTokens(src, e)
	oc, nc := d.oldChanged(), d.newChanged()
	// One trivia token changed, strictly inside, into one of its kind: every
	// other token only shifts.
	if len(oc) == 1 && len(nc) == 1 && oc[0].Kind == nc[0].Kind &&
		oc[0].Start < e.Start && e.OldEnd < oc[0].End &&
		(oc[0].Kind == lexer.Whitespace || oc[0].Kind == lexer.Comment || oc[0].Kind == lexer.Pod) {
		return StrategyReuseAll
	}
	if size > maxLocalEdit || strings.Count(e.NewText, "\n") > maxLocalLines ||
		hasTrigger(oc, d.oldSrc) || hasTrigger(nc, d.newSrc) ||
		!nests(oc, d.oldSrc) || !nests(nc, d.newSrc) || touchesDeclaration(old, e) {
		return StrategyFull
	}
	// The change must end inside the region widen re-parses: past it, the
	// old tree's tail is no longer the new source's.
	a, delta := anchorOf(old, e), len(e.NewText)-(e.OldEnd-e.Start)
	if len(oc) > 0 && (oc[0].Start < a.Start || oc[len(oc)-1].End > a.End) ||
		len(nc) > 0 && (nc[0].Start < a.Start || nc[len(nc)-1].End > a.End+delta) {
		return StrategyFull
	}
	return StrategyLocal
}

// hasTrigger reports a trigger word or kind among toks.
func hasTrigger(toks []lexer.Token, src []byte) bool {
	for _, t := range toks {
		if triggerKinds[t.Kind] || t.Kind == lexer.Word && triggerWords[string(src[t.Start:t.End])] {
			return true
		}
	}
	return false
}

// nests reports that toks' brackets balance and never close more than they
// opened, counted as refuseUnclosedOpener counts them, so the change can
// neither end the anchor's block early (`}{`) nor reach past it (C1, C2, C4).
func nests(toks []lexer.Token, src []byte) bool {
	depth := 0
	for _, t := range toks {
		if t.End-t.Start != 1 {
			continue
		}
		switch {
		case t.Kind == lexer.Operator && strings.IndexByte("([{", src[t.Start]) >= 0:
			depth++
		case t.Kind == lexer.CloseBracket:
			if depth--; depth < 0 {
				return false
			}
		}
	}
	return depth == 0
}

// touchesDeclaration reports e touching what can change how later text
// parses (§7.7.3's prototype row): a `use` or `no`, a BEGIN block, a
// prototype, or a sub, package or class declaration anywhere but strictly
// inside its block. Touching includes abutting. A sub renamed changes only
// a Word, which the token rules do not see.
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

// tokenDiff is an edit seen as tokens: old's and new's token streams agree
// on their first pre tokens, and on their last suf once old's are shifted.
// What lies between is what the edit changed.
type tokenDiff struct {
	oldSrc, newSrc   []byte
	oldToks, newToks []lexer.Token
	pre, suf         int
}

// diffTokens lexes src before and after e and finds where they agree. A
// prefix token ends at or before the edit; a suffix token starts at or
// after it on both sides.
//
// shortcut: lexes both whole sources; resume from a checkpoint against the
// token cache (§6.3.2) once it exists.
func diffTokens(src []byte, e Edit) tokenDiff {
	newSrc := append(append(append([]byte{}, src[:e.Start]...), e.NewText...), src[e.OldEnd:]...)
	d := tokenDiff{oldSrc: src, newSrc: newSrc}
	d.oldToks, _ = lexOld(src)
	d.newToks, _ = lexOld(newSrc)
	o, n := d.oldToks, d.newToks
	for d.pre < min(len(o), len(n)) && o[d.pre] == n[d.pre] && o[d.pre].End <= e.Start {
		d.pre++
	}
	delta, newEnd := len(e.NewText)-(e.OldEnd-e.Start), e.Start+len(e.NewText)
	for d.suf < min(len(o), len(n))-d.pre {
		a, b := o[len(o)-1-d.suf], n[len(n)-1-d.suf]
		if a.Start < e.OldEnd || b.Start < newEnd {
			break
		}
		a.Start, a.End = a.Start+delta, a.End+delta
		if a != b {
			break
		}
		d.suf++
	}
	return d
}

func (d tokenDiff) oldChanged() []lexer.Token { return d.oldToks[d.pre : len(d.oldToks)-d.suf] }
func (d tokenDiff) newChanged() []lexer.Token { return d.newToks[d.pre : len(d.newToks)-d.suf] }

// spanOf is the bytes of src toks cover, nil for none.
func spanOf(toks []lexer.Token, src []byte) []byte {
	if len(toks) == 0 {
		return nil
	}
	return src[toks[0].Start:toks[len(toks)-1].End]
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
	anchor := anchorOf(old, e)
	_, cps := lexOld(src)
	i := sort.Search(len(cps), func(i int) bool { return cps[i].Byte > anchor.Start })
	return cps[i-1], anchor
}

// anchorOf is widen's anchor for e.
func anchorOf(old *Node, e Edit) *Node {
	for _, top := range old.Children {
		if top.Start <= e.Start && e.OldEnd <= top.End && inSubBody(top, e) {
			return top
		}
	}
	return old
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
