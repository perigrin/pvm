// ABOUTME: The §6.8.4 anti-cascade rule: one syntax error, at most one diagnostic.
// ABOUTME: parseErrors finds the reported errors, diagFilter cools down and caps them, gateTypes drops type diagnostics inside them.

package psc

import (
	"slices"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/parse"
)

const (
	maxParseDiagnostics = 20 // hard cap per publication
	errorCooldownBytes  = 40 // suppress an error starting this close to the last one's end
)

// CodeMissingToken is the code of a diagnostic for a token the parser
// synthesized because the source did not supply it: a `}`, `)`, `]`, `;` or
// identifier (a Missing node with no refusal code).
const CodeMissingToken = "missing_token"

// diagFilter admits syntax diagnostics in source order (§6.8.4). One is made
// per analysis, so its count and cooldown never carry over between
// publications or documents.
type diagFilter struct {
	lastErrorEnd uint32
	count        int
}

// admit reports whether d is published: not past the cap, and, after the
// first, starting at least errorCooldownBytes after the last admitted end.
func (f *diagFilter) admit(d infer.Diagnostic) bool {
	if f.count >= maxParseDiagnostics {
		return false
	}
	if f.count > 0 && d.StartByte < f.lastErrorEnd+errorCooldownBytes {
		return false
	}
	f.lastErrorEnd = d.EndByte
	f.count++
	return true
}

// span is a byte range [start, end) of the source.
type span struct{ start, end uint32 }

// parseErrors is every reported syntax error in root's tree, in source
// order: each refusal perl shares (perlRefuses), each Missing node for a
// token, and each Error or Missing node whose code is not catch-all,
// missing_operand among them (R-g). A catch-all node is structure only and
// reports nothing (R-a′); any other refusal is a gap in the parser and
// reports nothing either. Nothing inside a refusal is reported again: a
// Missing `}` inside an unclosed_brace Unknown, or a refusal surfaced under
// a catch-all Error, is part of that one error.
//
// gates holds, per error, the span no type diagnostic may come from: the
// innermost statement holding it, so `$u + 1` in `my $y = $u + 1 + ;` is
// gated though the Missing operand is zero-width after it, or the error's
// own span where that reaches further.
func parseErrors(root *parse.Node) (errs []infer.Diagnostic, gates []span) {
	var walk func(n, stmt *parse.Node)
	walk = func(n, stmt *parse.Node) {
		if n.Kind == parse.Statement {
			stmt = n
		}
		var d infer.Diagnostic
		switch {
		case n.Refused():
			if parse.CatchAll(n.Refusal) || n.Kind == parse.Unknown && !perlRefuses[n.Refusal] {
				return
			}
			d = infer.Diagnostic{Code: string(n.Refusal), Message: "perl refuses this: " + parse.RefusalSites[n.Refusal].What}
		case n.Kind == parse.Missing:
			what := "an identifier"
			if n.Text != "" {
				what = "`" + n.Text + "`"
			}
			d = infer.Diagnostic{Code: CodeMissingToken, Message: "missing " + what}
		default:
			for _, c := range n.Children {
				walk(c, stmt)
			}
			return
		}
		d.StartByte, d.EndByte, d.Severity = uint32(n.Start), uint32(n.End), infer.Error
		errs = append(errs, d)
		g := span{d.StartByte, d.EndByte}
		if stmt != nil {
			g = span{min(g.start, uint32(stmt.Start)), max(g.end, uint32(stmt.End))}
		}
		gates = append(gates, g)
	}
	walk(root, nil)
	return errs, gates
}

// stmtErrors is parseErrors of one top-level statement that holds an
// error: its index among the root's children, the statement as the tree
// it was found in holds it, and its errors and gates in that tree.
type stmtErrors struct {
	i     int
	node  *parse.Node
	errs  []infer.Diagnostic
	gates []span
}

// syntaxErrors is parseErrors of root, by top-level statement, in order.
func syntaxErrors(root *parse.Node) []stmtErrors {
	var out []stmtErrors
	for i := range root.Children {
		out = appendErrors(out, root, i)
	}
	return out
}

// resyntax is syntaxErrors(root) for root, the tree of an edit, given prev,
// syntaxErrors of the tree before it, and redo, in order, the indexes of
// the top-level statements the edit changed (lower.Positions.Renumbered).
// Every other statement is one of the old tree's moved whole, so its
// errors are prev's, still in the old tree's offsets (flatten moves them).
func resyntax(prev []stmtErrors, root *parse.Node, redo []int) []stmtErrors {
	var out []stmtErrors
	for len(prev) > 0 || len(redo) > 0 {
		if len(redo) == 0 || len(prev) > 0 && prev[0].i < redo[0] {
			out, prev = append(out, prev[0]), prev[1:]
			continue
		}
		if len(prev) > 0 && prev[0].i == redo[0] {
			prev = prev[1:]
		}
		out, redo = appendErrors(out, root, redo[0]), redo[1:]
	}
	return out
}

// appendErrors is out and parseErrors of root's i-th statement, if it
// holds an error.
func appendErrors(out []stmtErrors, root *parse.Node, i int) []stmtErrors {
	if errs, gates := parseErrors(root.Children[i]); len(errs) != 0 {
		out = append(out, stmtErrors{i, root.Children[i], errs, gates})
	}
	return out
}

// flatten is the errors and gates of syn, root's syntaxErrors, in root's
// offsets: each statement's moved by how far it moved since they were
// found. The move is added mod 2^32, so a statement that moved back
// subtracts.
func flatten(syn []stmtErrors, root *parse.Node) (errs []infer.Diagnostic, gates []span) {
	for _, e := range syn {
		d := uint32(root.Children[e.i].Start - e.node.Start)
		for _, x := range e.errs {
			x.StartByte, x.EndByte = x.StartByte+d, x.EndByte+d
			errs = append(errs, x)
		}
		for _, g := range e.gates {
			gates = append(gates, span{g.start + d, g.end + d})
		}
	}
	return errs, gates
}

// filterParse is what diagFilter admits of errs, which are in source order.
func filterParse(errs []infer.Diagnostic) []infer.Diagnostic {
	var f diagFilter
	var out []infer.Diagnostic
	for _, d := range errs {
		if f.admit(d) {
			out = append(out, d)
		}
	}
	return out
}

// gateTypes drops each type diagnostic overlapping a gate (parseErrors),
// its error suppressed by the filter or not: a type error from a tree the
// parser knows is wrong is noise (§6.8.4, PSC gating). A zero-width gate at
// p overlaps a diagnostic holding the byte at p; a diagnostic ending exactly
// where a gate starts is kept.
func gateTypes(diags []infer.Diagnostic, gates []span) []infer.Diagnostic {
	var out []infer.Diagnostic
	for _, d := range diags {
		if !slices.ContainsFunc(gates, func(g span) bool { return g.overlaps(d) }) {
			out = append(out, d)
		}
	}
	return out
}

// overlaps reports whether d holds a byte of g, or, for a zero-width g at p,
// the byte at p.
func (g span) overlaps(d infer.Diagnostic) bool {
	return g.start < d.EndByte && d.StartByte < max(g.end, g.start+1)
}
