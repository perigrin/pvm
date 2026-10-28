// ABOUTME: Edit, Repair and ReparseIncremental, the IncrementalRepair flag, edit triage and widening (spec §6.5, §7.7).
// ABOUTME: Repair re-parses the one top-level statement widening anchors at and splices it between copied neighbours.
package parse

import (
	"bytes"
	"crypto/sha256"
	"slices"
	"sort"
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

// IncrementalRepair switches ReparseIncremental from a full re-parse on
// every change to Repair (spec §6.5.5). Tests may set it locally.
var IncrementalRepair = false

// Edit replaces the old source's bytes [Start, OldEnd) with NewText.
type Edit struct {
	Start, OldEnd int
	NewText       string
}

// ReparseIncremental returns the tree of src, the old source with e applied,
// reusing old where it can. It equals Parse(src) under TreeDiff (§7.7.1).
// With IncrementalRepair set it is Repair, or where Repair returns ok=false
// a full parse carrying what the next Repair resumes from; with it clear,
// always a full parse.
func ReparseIncremental(old *Node, e Edit, src []byte) *Node {
	if IncrementalRepair {
		if tree, ok := Repair(old, e, src); ok {
			return tree
		}
		return parseTracked(src, nil, "")
	}
	return Parse(src)
}

// ReparseIncremental is the resolver form: its result equals a fresh
// NewSession().ParseSourceFrom(path, src, roots...). It parses in a fresh
// Session, never s's module facts, because a Session's second parse reports
// none of the LoadedModules, DeclarationErrors or DeclarationWarnings its
// first already read, and a module edited on disk must be read again.
func (s *Session) ReparseIncremental(old *Node, e Edit, path string, src []byte, roots ...string) *Node {
	if IncrementalRepair {
		if tree, ok := s.Repair(old, e, path, src, roots...); ok {
			return tree
		}
		// What ParseSourceFrom below would parse with.
		dirs := searchDirs(path, roots)
		return parseTracked(src, NewSessionWithCache(s.cache).resolverFor(dirs, false), resolverKey(dirs, false))
	}
	return NewSessionWithCache(s.cache).ParseSourceFrom(path, src, roots...)
}

// Repair is §6.5.3: the tree of src, the bytes old was parsed from with e
// applied, built by re-parsing only the top-level statement widening
// anchors at. Its tree equals Parse(src) under TreeDiff, root fields
// included, or ok is false and the tree nil, never a partial splice. It
// leaves old unchanged.
func Repair(old *Node, e Edit, src []byte) (tree *Node, ok bool) {
	tree, _, ok = repair(old, e, src, "", func() *resolver { return nil })
	return tree, ok
}

// Repair is the resolver form: its tree equals a fresh
// NewSession().ParseSourceFrom(path, src, roots...), or ok is false. Like
// ReparseIncremental it resolves as a fresh Session would, never with s's
// module facts.
func (s *Session) Repair(old *Node, e Edit, path string, src []byte, roots ...string) (tree *Node, ok bool) {
	tree, _, ok = sessionRepair(old, e, path, src, roots, s.cache)
	return tree, ok
}

// sessionRepair is Session.Repair, also returning the statement it
// re-parsed. Its resolvers share cache, which may be nil.
func sessionRepair(old *Node, e Edit, path string, src []byte, roots []string, cache *ModuleCache) (*Node, *Node, bool) {
	dirs := searchDirs(path, roots)
	return repair(old, e, src, resolverKey(dirs, false), func() *resolver { return newResolver(dirs, false).withCache(cache) })
}

// incState is what Repair resumes from, kept on a parse's root.
type incState struct {
	src []byte // the bytes the tree was parsed from
	key string // resolverKey of the parse, "" for none
	// tops is the parser before each top-level statement, while the parse
	// took snapshots, then at EOF before finish. A parse's tree has none:
	// Repair takes them by parsing src again, and a tree it builds carries
	// its own.
	tops  []parserSnapshot
	reads map[string]moduleRead // every Loader call the parse made
	// toks is src's token cache. A parse's tree has none: Repair lexes src
	// for one, and a tree it builds carries the cache the edit updated.
	toks *lexer.TokenCache
	// relexed is how many tokens the edit Repair built this tree for
	// re-lexed.
	relexed int
}

// moduleRead is what one Loader call answered.
type moduleRead struct {
	ok  bool
	sum [sha256.Size]byte
}

// maxResyncTokens bounds the tokens an edit may change: past it, the
// change is not local and a full parse is cheaper (§6.5.3).
const maxResyncTokens = 4000

// repair is Repair, also returning the statement it re-parsed. It resolves
// as the parse whose resolverKey is key does, each mkRes call returning a
// fresh resolver for it: nil for none.
//
// It re-parses the anchor's statement from the parser snapshot before it
// and keeps every other top-level statement, copied and, past the edit,
// shifted, when §6.5.2's checks hold: a top-level snapshot starts at the
// anchor (0); one statement read from it over src ends at the anchor's
// shifted end and leaves the parser as the old parse's snapshot after the
// anchor stood, its token index moved (1, 2); the modules the old parse
// read read the same bytes (3); and the statement consulted no resolver,
// so the root's loaded modules and declaration errors are the old parse's.
// The EOF passes then run over the spliced tree (4). Under ruling R-f
// nothing holding an error is reused, and a statement that gains or loses
// one returns ok=false (§6.8.5).
//
// shortcut: copies and shifts every reused node, the simplest splice; ruling
// R-e chooses an in-place shift or shared subtrees once 01a17a40-4915
// measures them.
func repair(old *Node, e Edit, src []byte, key string, mkRes func() *resolver) (*Node, *Node, bool) {
	if old == nil || old.inc == nil || !isEditOf(old.inc.src, e, src) {
		return nil, nil, false
	}
	strategy, d := triageDiff(old, old.inc.src, e)
	var a *Node
	switch strategy {
	case StrategyLocal:
		a = anchorOf(old, e)
	case StrategyReuseAll:
		a = childHolding(old, e)
	}
	if a == nil || a == old {
		return nil, nil, false
	}
	for _, c := range old.Children {
		if c != a && c.HasError {
			return nil, nil, false
		}
	}
	tops, reads := old.inc.tops, old.inc.reads
	if tops == nil || old.inc.key != key {
		if tops, reads = snapshotsOf(old, mkRes()); tops == nil {
			return nil, nil, false
		}
	} else if !sameReads(reads, mkRes()) {
		return nil, nil, false
	}
	k := sort.Search(len(tops), func(i int) bool { return tops[i].at >= a.Start })
	if k+1 >= len(tops) || tops[k].at != a.Start || tops[k+1].at != a.End {
		return nil, nil, false
	}

	delta, dtok := len(src)-len(old.inc.src), len(d.newToks)-len(d.oldToks)
	res, bodies := mkRes(), heredocBodyIndices(d.newToks)
	p := tops[k].restoreOn(src, d.newToks, bodies, res)
	if p.pos >= len(p.toks) {
		return nil, nil, false // the edit deleted the anchor, trailing trivia, whole
	}
	n := p.statement()
	// Check 1: the first unconsumed token starts at the anchor's shifted
	// end, or the parse is at EOF. The token index in check 2 implies it
	// only while a snapshot follows the anchor. tops ends with an EOF
	// snapshot so one always does, but check 1 is kept outright rather
	// than resting on that (gate review of #512).
	end := len(src)
	if p.pos < len(p.toks) {
		end = p.toks[p.pos].Start
	}
	want := tops[k+1].p.clone()
	want.pos += dtok
	if n == nil || end != a.End+delta || !sameState(p, &want) || consulted(res) {
		return nil, nil, false
	}

	root := &Node{Kind: SourceFile, Start: 0, End: len(src)}
	memo := map[*Node]*Node{}
	for _, c := range old.Children {
		switch {
		case c == a:
			root.Children = append(root.Children, n)
		case c.Start < a.Start:
			root.Children = append(root.Children, copyShifted(c, 0, memo))
		default:
			root.Children = append(root.Children, copyShifted(c, delta, memo))
		}
	}
	fin := tops[len(tops)-1].restoreOn(src, d.newToks, bodies, res)
	fin.pos += dtok
	fin.finish(root)
	if n.HasError != a.HasError {
		return nil, nil, false
	}
	if res != nil {
		root.loaded = slices.Clone(old.loaded)
		root.declErrs = slices.Clone(old.declErrs)
		root.declWarns = slices.Clone(old.declWarns)
	}

	moved := slices.Clone(tops)
	for j := k + 1; j < len(moved); j++ {
		moved[j].at += delta
		moved[j].p.pos += dtok
	}
	root.inc = &incState{src: bytes.Clone(src), key: key, tops: moved, reads: reads, toks: d.cache, relexed: d.relexed}
	return root, n, true
}

// isEditOf reports that edited is old with e applied.
func isEditOf(old []byte, e Edit, edited []byte) bool {
	return 0 <= e.Start && e.Start <= e.OldEnd && e.OldEnd <= len(old) &&
		len(edited) == len(old)-(e.OldEnd-e.Start)+len(e.NewText) &&
		bytes.Equal(old[:e.Start], edited[:e.Start]) &&
		string(edited[e.Start:e.Start+len(e.NewText)]) == e.NewText &&
		bytes.Equal(old[e.OldEnd:], edited[e.Start+len(e.NewText):])
}

// snapshotsOf parses old's source again with res and returns the top-level
// snapshots and Loader calls that parse recorded, when it equals old: old is
// this source's tree, and what the modules it read say has not changed it.
// Otherwise nil.
//
// shortcut: the first Repair of a parse's tree parses the old source again
// for its snapshots; take them in the document's first parse once the LSP
// owns one (§6.7.1). A tree ReparseIncremental built carries them already.
func snapshotsOf(old *Node, res *resolver) ([]parserSnapshot, map[string]moduleRead) {
	root := parseTracked(old.inc.src, res, old.inc.key)
	if TreeDiff(old, root) != "" {
		return nil, nil
	}
	return root.inc.tops, root.inc.reads
}

// parseTracked is the tree parsing src with res builds, which resolves as
// the parse whose resolverKey is key does, carrying what Repair resumes
// from: the top-level snapshots, every Loader call and src's token cache.
func parseTracked(src []byte, res *resolver, key string) *Node {
	reads := map[string]moduleRead{}
	if res != nil {
		load := res.load
		res.load = func(name string) ([]byte, bool) {
			b, ok := load(name)
			reads[name] = moduleRead{ok, sha256.Sum256(b)}
			return b, ok
		}
	}
	src = bytes.Clone(src)
	c := tokensOf(nil, src)
	p := newParserOn(src, c.Tokens, res, false)
	p.snaps = &[]parserSnapshot{}
	root := p.parseFile()
	tops := slices.DeleteFunc(*p.snaps, func(s parserSnapshot) bool { return s.nested })
	tops = append(tops, parserSnapshot{at: len(p.src), p: p.clone()})
	p.finish(root)
	for i := range tops {
		// The resuming parse supplies these; keeping them would hold this
		// parse's tokens and resolver as long as the tree.
		tops[i].p.src, tops[i].p.toks, tops[i].p.bodies, tops[i].p.res = nil, nil, nil, nil
	}
	root.inc = &incState{src: src, key: key, tops: tops, reads: reads, toks: c}
	return root
}

// sameReads reports that each Loader call in reads answers as it did,
// through res's loader: check 3 for a tree that carries its snapshots.
// It reads every module again on each Repair: a ModuleCache saves parsing,
// never reading, so a module edited on disk is always seen.
func sameReads(reads map[string]moduleRead, res *resolver) bool {
	for name, r := range reads {
		b, ok := res.load(name)
		if ok != r.ok || sha256.Sum256(b) != r.sum {
			return false
		}
	}
	return true
}

// consulted reports that a parse used res. A fresh parse would have met the
// old parse's modules already read, so its loaded list and declaration
// errors would no longer be the old parse's.
func consulted(res *resolver) bool {
	return res != nil && (len(res.seen) > 0 || len(res.loaded) > 0 || len(res.declErrs) > 0 || len(res.declWarns) > 0)
}

// copyShifted is a deep copy of n with every byte offset moved by delta, as
// TreeDiffShifted compares them. A node reached twice is copied once. Only
// nodes past the edit's end move, so no offset can underflow (§6.4.2).
func copyShifted(n *Node, delta int, memo map[*Node]*Node) *Node {
	if c, ok := memo[n]; ok {
		return c
	}
	c := *n
	memo[n] = &c
	c.Start += delta
	c.End += delta
	if c.InnerEnd != 0 {
		c.InnerStart += delta
		c.InnerEnd += delta
	}
	c.Children = copyAll(n.Children, delta, memo)
	c.InnerHeredocBodies = copyAll(n.InnerHeredocBodies, delta, memo)
	return &c
}

func copyAll(ns []*Node, delta int, memo map[*Node]*Node) []*Node {
	if ns == nil {
		return nil
	}
	out := make([]*Node, len(ns))
	for i, n := range ns {
		out[i] = copyShifted(n, delta, memo)
	}
	return out
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
// text: src's token cache, old's when old carries one, is updated for e,
// and only the tokens that differ are examined.
func triage(old *Node, src []byte, e Edit) Strategy {
	s, _ := triageDiff(old, src, e)
	return s
}

// triageDiff is triage, also returning the token streams it compared,
// empty when it did not lex.
func triageDiff(old *Node, src []byte, e Edit) (Strategy, tokenDiff) {
	size := max(e.OldEnd-e.Start, len(e.NewText))
	if size > maxEditBytes {
		return StrategyFull, tokenDiff{}
	}
	d := diffTokens(tokensOf(old, src), src, e)
	return triageTokens(old, e, d, size), d
}

// triageTokens is triage once src is lexed before and after e.
func triageTokens(old *Node, e Edit, d tokenDiff, size int) Strategy {
	oc, nc := d.oldChanged(), d.newChanged()
	if max(len(oc), len(nc)) > maxResyncTokens {
		return StrategyFull
	}
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
	cache            *lexer.TokenCache // newSrc's, the old cache edited
	relexed          int               // tokens the edit re-lexed
}

// diffTokens applies e to c, src's token cache, and finds where the streams
// before and after agree. A prefix token ends at or before the edit; a
// suffix token starts at or after it on both sides. Only the tokens from
// the checkpoint before e to where the stream resyncs are lexed, and c is
// left as it was.
func diffTokens(c *lexer.TokenCache, src []byte, e Edit) tokenDiff {
	newSrc := append(append(append([]byte{}, src[:e.Start]...), e.NewText...), src[e.OldEnd:]...)
	nc := *c
	relexed := nc.Edit(newSrc, e.Start, e.OldEnd, e.Start+len(e.NewText))
	d := tokenDiff{oldSrc: src, newSrc: newSrc, oldToks: c.Tokens, newToks: nc.Tokens, cache: &nc, relexed: relexed}
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

// tokensOf is the token cache of src, lexed as Parse lexes a program: the
// one n's tree carries when it is of src, else src lexed whole.
func tokensOf(n *Node, src []byte) *lexer.TokenCache {
	if n != nil && n.inc != nil && n.inc.toks != nil && bytes.Equal(n.inc.src, src) {
		return n.inc.toks
	}
	return lexer.NewTokenCache(src, lexer.StartStateBarewords(coreBarewordOperators()))
}

// widen is §6.5.2: the lexer checkpoint to re-lex from and the node to
// re-parse whole for a StrategyLocal edit of src, whose tree is old.
//
// The anchor is the top-level statement holding the edit, and old itself
// when none holds it or the edit is in a `use`, a BEGIN block or a sub,
// package or class declaration's head: those are never anchors.
// A sub inside a package block, bare block or `if` anchors at that whole
// statement, past the block, because a nested resume stops at the block's
// `}` with no post-block state, and a top-level one runs to end of file.
// The checkpoint is the last at or before the anchor's start.
func widen(old *Node, src []byte, e Edit) (lexer.Checkpoint, *Node) {
	anchor := anchorOf(old, e)
	cps := tokensOf(old, src).Checkpoints
	i := sort.Search(len(cps), func(i int) bool { return cps[i].Byte > anchor.Start })
	return cps[i-1], anchor
}

// anchorOf is widen's anchor for e.
func anchorOf(old *Node, e Edit) *Node {
	for _, top := range old.Children {
		if top.Start <= e.Start && e.OldEnd <= top.End && reparsable(top, e) {
			return top
		}
	}
	return old
}

// reparsable reports that the path from n down to e passes through no
// `use`, BEGIN block, or sub, package or class declaration but strictly
// inside its block: what touchesDeclaration refuses. Anything else a
// statement changes for the statements after it, Repair's check 2 sees.
func reparsable(n *Node, e Edit) bool {
	for ; n != nil; n = childHolding(n, e) {
		switch {
		case n.Kind == Use, n.Kind == Phaser && n.Text == "BEGIN":
			return false
		case n.Kind == Declaration && (n.Text == "sub" || n.Text == "package" || n.Text == "class") && !inBody(n, e):
			return false
		}
	}
	return true
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
