// ABOUTME: Parser-state snapshots at re-parse points, the parser half of what a re-parse resumes from.
// ABOUTME: Taken before each top-level statement and each statement of a file-level package or bare block (spec 06 §6.5.3).
package parse

import (
	"maps"
	"slices"
)

// parserSnapshot is the parser as it stood just before one statement: a
// top-level one, or, when nested is set, one directly inside a file-level
// `package NAME BLOCK`, `class NAME BLOCK` or bare block. Those are the
// re-parse points §6.5.2's widening may return; a BEGIN block, a `use` and
// a `format` have none inside them.
type parserSnapshot struct {
	at     int  // Start of toks[pos]: the next token, trivia or a heredoc body included
	nested bool // a resume stops at the enclosing block's `}`
	p      parser
}

// Field classes for parserFields.
const (
	snapshotted = iota + 1 // copied into the snapshot, maps and slices cloned
	scratch                // zero at every re-parse point; copied as is
	shared                 // the resuming parse's own: its source, tokens and resolver
)

// parserFields classifies every field of parser. A field added to parser
// fails TestParserStateSnapshotCoversEveryField until it is listed here, and
// a map or slice listed as snapshotted fails it until clone copies it.
var parserFields = map[string]int{
	"moduleSubs": snapshotted,
	"bodies":     shared,
	"depth":      snapshotted,
	// A count, not a stack: finish walks for HasError when it is non-zero,
	// so a resumed parse carrying the prefix's count only walks the more.
	"recovered":      snapshotted,
	"attrOpenAtEOF":  snapshotted,
	"typed":          snapshotted,
	"signatures":     snapshotted,
	"variables":      snapshotted,
	"buildingCore":   snapshotted,
	"derivingPowers": snapshotted,
	"typedErrs":      snapshotted,
	"operators":      snapshotted,
	"relations":      snapshotted,
	"inError":        snapshotted,
	"declaredMulti":  snapshotted,
	"declaredPlain":  snapshotted,
	"features":       snapshotted,
	"syntax":         snapshotted,
	"libraryOps":     snapshotted,
	"libraryInfix":   snapshotted,
	"noIndirect":     snapshotted,
	"symbolsOpen":    snapshotted,
	"inBegin":        snapshotted,
	"packages":       snapshotted,
	"listed":         snapshotted,
	"lexical":        snapshotted,
	"pkg":            snapshotted,
	"imports":        snapshotted,
	"bodied":         snapshotted,
	"pos":            snapshotted,
	"inSignature":    scratch,
	"inHashKey":      scratch,
	"derefMemo":      scratch,
	"termBP":         scratch,
	"declOperandBP":  scratch,
	"operatorSig":    scratch,
	"multi":          scratch,
	"anchorBlock":    scratch,
	// Set by the statementsTo the snapshot stands in: true at a nested point.
	"inBlock": snapshotted,
	"src":     shared,
	"toks":    shared,
	// The Session doing the re-parse supplies its own resolver, a fresh one
	// per 01a17969-4629's rule; the snapshot never carries one.
	"res":   shared,
	"snaps": shared,
	// Cloned, though no snapshot follows a tier-1 close (rescued).
	"closedEarly": snapshotted,
	"rescued":     scratch,
}

// clone is a copy of p holding no map or slice of p's, so neither can
// change the other. The values in them are not mutated in place, except
// that signatures' slices are appended to (parseTypedSignature): each is
// clipped, so an append copies rather than writing into shared capacity.
func (p *parser) clone() parser {
	c := *p
	c.snaps = nil
	c.moduleSubs = maps.Clone(p.moduleSubs)
	c.signatures = maps.Clone(p.signatures)
	for name, sigs := range c.signatures {
		c.signatures[name] = slices.Clip(sigs)
	}
	c.variables = maps.Clone(p.variables)
	c.typedErrs = slices.Clone(p.typedErrs)
	c.operators = slices.Clone(p.operators)
	c.relations = slices.Clone(p.relations)
	c.inError = maps.Clone(p.inError)
	c.declaredMulti = maps.Clone(p.declaredMulti)
	c.declaredPlain = maps.Clone(p.declaredPlain)
	c.features = maps.Clone(p.features)
	c.syntax = maps.Clone(p.syntax)
	c.libraryOps = slices.Clone(p.libraryOps)
	c.libraryInfix = maps.Clone(p.libraryInfix)
	c.packages = maps.Clone(p.packages)
	c.listed = maps.Clone(p.listed)
	c.lexical = maps.Clone(p.lexical)
	c.imports = maps.Clone(p.imports)
	c.bodied = maps.Clone(p.bodied)
	c.closedEarly = maps.Clone(p.closedEarly)
	return c
}

// checkpoint records a snapshot before the statement at p.pos, when p is
// collecting them and no brace has closed at a column-0 anchor before it.
func (p *parser) checkpoint(nested bool) {
	if p.snaps == nil || p.rescued {
		return
	}
	*p.snaps = append(*p.snaps, parserSnapshot{at: p.toks[p.pos].Start, nested: nested, p: p.clone()})
}

// resume parses src from s: to end of file, or to the `}` closing s's block
// when s is nested. src must lex as the snapshotted source did up to s.at,
// which holds when every edit since lies after it. It returns the statements
// read and the parser that read them, whose finish runs the EOF passes once
// the caller has spliced them into a root.
func (s *parserSnapshot) resume(src []byte, res *resolver) (*parser, []*Node) {
	p := s.restore(src, res)
	return p, p.statementsTo(s.nested, false, nil)
}

// restore is a parser over src standing where s was taken, not yet run.
func (s *parserSnapshot) restore(src []byte, res *resolver) *parser {
	fresh := newParser(src, res, s.p.typed)
	p := s.p.clone()
	p.src, p.toks, p.bodies, p.res = fresh.src, fresh.toks, fresh.bodies, fresh.res
	return &p
}
