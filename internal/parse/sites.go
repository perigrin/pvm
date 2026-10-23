// ABOUTME: The subject side of the fidelity harness: what this parser reports about each statement.
// ABOUTME: An Unknown hedges every marker, because saying nothing is scored as claiming nothing is there.

package parse

import (
	"bytes"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parseoracle"
)

// markerKinds is every site kind the oracle scores. A statement this parser
// could not settle must hedge all of them, not just the one it happened to
// be looking for.
var markerKinds = []string{
	parseoracle.SiteKindReference,
	parseoracle.SiteKindHash,
	parseoracle.SiteKindMatch,
	parseoracle.SiteKindReadline,
	parseoracle.SiteKindAnonhash,
}

// Sites reports what this parser concluded about each statement, in the
// vocabulary the subject contract asks for.
//
// The rule that matters here is what an Unknown reports. CompareFacts
// attributes each of perl's sites to the innermost statement covering it and
// then decides (compare_facts.go:218-232):
//
//	case owner != nil && owner.decided[m] > 0:  // exact
//	case owner != nil && owner.hedged[m] > 0:   // wider
//	default:                                    // WRONG
//
// An Unknown node emits no sites, so owner is nil for every perl site inside
// it and the verdict is WRONG -- not no-answer. No-answer is a FILE-level
// bucket: Declined, !OK, or no site of a marker's kind anywhere in the file.
// Once the subject reports one hash site, every rv2hv in that file is scored
// per statement.
//
// So "emit Unknown and stay quiet" is not a refusal. It is the claim that
// nothing is there, and it is wrong wherever perl found something. The
// refusal has to be spoken: an Unresolved site of each kind, spanning the
// statement, which lands in hedged and scores wider.
//
// That rule is why Unknown exists. Without it Unknown is a silent wrong
// answer with a reassuring name.
func Sites(root *Node, src []byte) []parseoracle.SubjectCallSite {
	// Non-nil even when empty: the contract distinguishes "I looked and
	// found none" from "I did not look", and a nil slice marshals to JSON
	// null, which decodes as absent. See SubjectFacts.CallSites.
	sites := []parseoracle.SubjectCallSite{}

	lines := lineIndex(src)
	for _, n := range root.Children {
		start, end := lines.at(n.Start), lines.at(n.End-1)

		if n.Kind == Unknown {
			for _, kind := range markerKinds {
				sites = append(sites, parseoracle.SubjectCallSite{
					Kind:       kind,
					Line:       start,
					EndLine:    end,
					Unresolved: true,
				})
			}
			continue
		}

		// A statement this parser DID read reports what it decided. Without
		// this the subject only ever hedges, and a subject that never commits
		// cannot be wrong -- nor right. WRONG=0 with exact=0 is the vacuous
		// gate the M1 issue's "exact floor" row exists to prevent.
		//
		// Each site is reported once per occurrence, spanning its
		// statement. perl's probe walks the CVs, so a backslash in a sub
		// body is reported at the body's line; CompareFacts spends a decided
		// site only against the innermost subject statement covering that
		// line (`compare_facts.go:349`), and a statement span covers it.
		for _, d := range decidedMarkers(n, src) {
			// The span is the whole STATEMENT, because perl's nextstate
			// line can be either end of it.
			//
			// perl attributes a site to its enclosing nextstate, which
			// usually records the statement's FIRST line. Measured on 5.42.0:
			//
			//	my $r = f(       # line 1
			//	    "a",
			//	    { k => 1 },  # line 3
			//	);
			//	-> nextstate(main 1 ml.pl:1), anonhash
			//
			// A site reported as the point 3-3 is never found by `innermost`,
			// which wants a range CONTAINING perl's line
			// (`compare_facts.go:349-359`); starting at the statement fixes
			// that. t/comp/parser_run.t:17 is the case.
			//
			// But perl's copline moves past an anonymous sub, a `when`
			// block or a bareword subscript, and then the nextstate records
			// the statement's LAST line:
			//
			//	my @r = (\@a, sub {  # line 2
			//	 1;
			//	});                  # line 4
			//	-> nextstate(main 2 rep.pl:4), srefgen
			//
			// A site ending at its own node, line 2, never contains line 4.
			// comp/proto.t:540, op/bop.t:706 and op/switch.t:209 are this
			// shape.
			//
			// A site inside a sub body, which perl's CV walk reports at the
			// body's own line (t/class/destruct.t:22), lies inside the
			// statement either way.
			sites = append(sites, parseoracle.SubjectCallSite{
				Kind:       d.kind,
				Line:       start,
				EndLine:    end,
				Unresolved: d.unresolved,

				// The srefgen marker is decided by this field, not by the
				// kind (`compare_facts.go:336`): a reference site without it
				// is "an answer, but not one that accounts for a reference
				// perl took", which lands in neither decided nor hedged. A
				// `\` this parser read IS a reference taken. The other four
				// markers are decided by their kind and ignore the field.
				//
				// A hedge takes nothing -- it is the refusal to say -- and
				// groupByStatement reads Unresolved first either way.
				TookReference: d.kind == parseoracle.SiteKindReference &&
					!d.unresolved,
			})
		}
	}
	return sites
}

// decided is one marker this parser settled and the node that settled it.
// The node travels with the kind because a site's line is its own, not the
// enclosing statement's -- see Sites.
//
// unresolved makes it a hedge rather than a commitment. A nested Unknown
// settles every marker the only way it can: by refusing, out loud.
type decided struct {
	kind       string
	node       *Node
	unresolved bool
}

// decidedMarkers reports which of the five markers this statement contains,
// in the subject vocabulary. Each is a conclusion the parser committed to.
//
// Only what the tree SHOWS, never what the bytes suggest. `$a % $b` is a
// Binary whose Text is "%", not a hash; `$a / $b` is division, not a
// pattern; `$a < $c` is a comparison, not a readline. The lexer settled all
// three by position (§3.2) and the tree records the answer, so reading the
// tree cannot reach a different one. Scanning source text could.
func decidedMarkers(stmt *Node, src []byte) []decided {
	var found []decided

	// One entry per OCCURRENCE, not per kind. CompareFacts consumes a decided
	// site for each of perl's (`compare_facts.go:224`, `owner.decided[m]--`),
	// so a statement holding two anonymous hashes and reporting one site
	// scores the first exact and the second WRONG.
	//
	// Deduplicating here was the cause of 11 of the 20 WRONG files in T2's
	// first subject measurement -- `eq_hash({$o->h}, {qw( the hash )})` in
	// t/class/accessor.t:31 is two anonhash ops to perl and was one site here.
	add := func(kind string, n *Node) {
		found = append(found, decided{kind: kind, node: n})
	}

	// hedge reports a site this parser SEES but cannot settle. It scores
	// wider rather than WRONG, which is the honest bucket for "something
	// here may or may not emit the op, and the answer is not in the text".
	hedge := func(kind string, n *Node) {
		found = append(found, decided{kind: kind, node: n, unresolved: true})
	}

	// arrowChain reports the rv2hv perl builds for the subscript chain
	// topped by n, in context ctx: a decided site for each, or one hedge
	// where the count is not in the text. Each chain is reported once, by
	// the dereference around it or, failing one, at its top.
	reported := map[*Node]bool{}
	arrowChain := func(n *Node, ctx chainContext) {
		if reported[n] {
			return
		}
		reported[n] = true
		count, ok := chainRv2hv(n, ctx)
		if !ok {
			hedge(parseoracle.SiteKindHash, n)
			return
		}
		for ; count > 0; count-- {
			add(parseoracle.SiteKindHash, n)
		}
	}

	// up is the walk's ancestors, outermost first: the context a
	// dereference sits in is its parents, and op_lvalue reads them.
	var up []*Node

	var walk func(*Node)
	walk = func(n *Node) {
		// perl: srefgen, for an argument its callee's prototype takes by
		// reference (markReferenceSlots). The argument's own markers are
		// still walked below: `f(%$r)` through `\%` is a srefgen and a rv2hv.
		if n.RefSlot {
			add(parseoracle.SiteKindReference, n)
		}
		switch n.Kind {
		case Unknown:
			// A nested Unknown speaks its own refusal, at its own span. The
			// statement around it was READ -- `BEGIN { ... }` is a Phaser
			// holding a Block -- so nothing else hedges these bytes, and
			// silence here is the same silent wrong answer the top-level
			// rule exists to prevent: perl found a marker inside, innermost
			// found a statement offering neither a site nor a hedge, and
			// the verdict was WRONG rather than wider.
			//
			// Do not descend: nothing inside a refusal is decided.
			for _, kind := range markerKinds {
				found = append(found, decided{kind: kind, node: n, unresolved: true})
			}
			return

		case AnonHash:
			// perl: anonhash. The lexer's brace stack already chose term
			// over block (§4.9.2), so this node IS the decision.
			add(parseoracle.SiteKindAnonhash, n)

		case Index:
			// A BRACE subscript reads a hash, and whether perl emits rv2hv
			// for it turns on the same undecidable question a named `%h`
			// does -- plus the optimiser, which folds `$h{a}` to multideref
			// and emits nothing at all:
			//
			//	my %h;  @h{"a","b"}       0 rv2hv
			//	our %h; @h{"a","b"}       1 rv2hv
			//	my %h;  delete $h{a}      0
			//	our %h; delete local $h{a} 1
			//
			// A SLICE through a reference is rv2hv, decided: `@$r{"a","b"}`
			// is this node over a Unary `@`, and perl emits the op because a
			// reference is being read. Measured 1, where the same slice of a
			// named lexical hash is 0. `@$aref[0,1]` is an array slice --
			// rv2av -- so the BRACE is what makes this a hash, not the
			// sigil. `@{$r->{x}}` is an array dereference of an element and
			// never reaches here: its Index is inside the Unary.
			//
			// The Unary is not walked as a node of its own, which would
			// report nothing for `@` anyway; its operand and the subscript
			// are.
			if n.Text == "{" && !n.Arrow && len(n.Children) > 0 &&
				n.Children[0].Kind == Unary && n.Children[0].Text == "@" {
				add(parseoracle.SiteKindHash, n)
				for _, c := range n.Children[0].Children {
					arrowChain(c, chainModAggregate)
					walk(c)
				}
				for _, c := range n.Children[1:] {
					walk(c)
				}
				return
			}

			// The same slice spelled postfix, `$r->@{"a","b"}` or
			// `$r->%{"a","b"}`, is the right operand of an arrow: this node
			// with a bare `@` or `%` Term where the base would be. Measured,
			// perl emits rv2hv for both and for `{1..4}->%{1,3}`; the BRACKET
			// forms `->@[...]` and `->%[...]` are rv2av and never get here.
			// op/postfixderef.t:148,316 are this shape.
			if n.Text == "{" && len(n.Children) > 0 &&
				n.Children[0].Kind == Term &&
				(n.Children[0].Text == "@" || n.Children[0].Text == "%") {
				add(parseoracle.SiteKindHash, n)
			}

			// Hedged for a NAMED base and left alone for a dereference,
			// which is decided above.
			if n.Text == "{" && len(n.Children) > 0 &&
				n.Children[0].Kind == Term && !n.Arrow &&
				isNamedVariable(n.Children[0].Text) {
				hedge(parseoracle.SiteKindHash, n)
			}

			// A chain read in place: nothing dereferences its last element,
			// so it is decided as a read. Each level below the top is part of
			// the top's chain and reported with it.
			if len(up) == 0 || up[len(up)-1].Kind != Index ||
				up[len(up)-1].Children[0] != n {
				arrowChain(n, chainRead)
			}

		case Unary:
			// perl: srefgen. prefixName maps `\` to "ref", which keeps the
			// prefix spelling distinct from infix `-`.
			if n.Text == "ref" {
				add(parseoracle.SiteKindReference, n)
			}
			// perl: rv2hv. A DEREFERENCE is a hash site exactly as `%hash`
			// is -- `%$href` and `%{$h}` emit the same op. isHashTerm below
			// tests a LEAF's text, so it catches the plain form and misses
			// every deref, whose `$href` leaf says nothing about the `%`
			// that wraps it.
			//
			// Measured: cmd/subval.t:182,184 are both `%$href`, and the
			// subject committed to those statements with no hash site and no
			// hedge. That is WRONG, the one bucket that fails a build.
			//
			// `@$aref` is rv2av, a different op, so only `%` counts here.
			if n.Text == "%" {
				add(parseoracle.SiteKindHash, n)
			}
			// A braced deref of a brace subscript, `@{$::{$keys[0]}}`, is
			// decided here as one hash site, and its inner Index is not
			// walked, so the subscript does not hedge the same access. perl
			// emits one rv2hv for that statement, measured.
			//
			// The bare slice `@$r{"a","b"}` does not reach here: it is a
			// subscript on the deref, decided in the Index case.
			//
			// NOT `@{$r->{x}}`, which is an ARRAY dereference of a hash
			// element -- rv2av, a different op. Its Index carries Arrow, and
			// a chained `@{$r->{x}{$k}}` or `@{$$r{x}}` reads through a
			// reference the same way; both are counted below, by the
			// context around the dereference.
			if n.Text == "@" && len(n.Children) == 1 &&
				n.Children[0].Kind == Index && n.Children[0].Text == "{" &&
				!throughReference(n.Children[0]) {
				add(parseoracle.SiteKindHash, n)
				for _, c := range n.Children[0].Children {
					walk(c)
				}
				return
			}
			if (n.Text == "@" || n.Text == "%" || n.Text == "$") && len(n.Children) == 1 {
				arrowChain(n.Children[0], derefContext(n, n.Text, up))
			}

		case Call:
			// perl: readline. The BUILTIN and the `<...>` syntax are one op:
			//
			//	$ perl -MO=Concise,-exec -e 'my $l = readline $fh;'
			//	c  <1> readline[t5] sKS/1
			//
			// isReadlineTerm below reads a LEAF's angle-delimited text, which
			// is the lexer's own decision and correct as far as it goes. It
			// cannot see a call, so `readline $fh` was reported nowhere.
			//
			// Resolved only: an unresolved Call is a name this parser has not
			// seen, and claiming the builtin for it would be committing to a
			// reading on no evidence.
			//
			// A METHOD is not the builtin -- perl emits entersub for
			// `$obj->readline`. The arrow's right operand is skipped in the
			// walk below rather than tested here, because a name's meaning
			// is decided by what precedes it and a node cannot see that.
			if n.Text == "readline" && n.Resolved {
				add(parseoracle.SiteKindReadline, n)
			}

		case Binary:
			// The postfix spelling of a dereference, `$r->{k}->@*`, is the
			// same question as `@{$r->{k}}`. `->$#*` reads where `$#{...}`
			// does not, and only the read is measured, so it has no context.
			if n.Text == "->" && len(n.Children) == 2 && n.Children[1].Kind == Term {
				switch t := n.Children[1].Text; t {
				case "@*", "%*", "$*":
					arrowChain(n.Children[0], derefContext(n, t[:1], up))
				case "$#*":
					arrowChain(n.Children[0], chainUnknown)
				}
			}
			// perl: match. `=~` and `!~` bind a pattern; s/// and tr/// are
			// substitution and transliteration, which perl reports with
			// different ops, so they are excluded by their own spelling.
			if n.Text == "=~" || n.Text == "!~" {
				if len(n.Children) > 1 && isMatchOperand(n.Children[1]) {
					add(parseoracle.SiteKindMatch, n)
				}
			}
			// perl: srefgen. Smartmatch takes each array or hash operand by
			// reference (ref_array_or_hash), so `1 ~~ @l` compiles as
			// `1 ~~ \@l`. Measured: one srefgen per aggregate operand, none
			// for a scalar or a slice. op/switch.t's `when` is the same rule,
			// in the Conditional case.
			if n.Text == "~~" {
				for _, c := range n.Children {
					if isAggregate(c) {
						add(parseoracle.SiteKindReference, c)
					}
				}
			}

		case Conditional:
			// perl: srefgen. `when (@l)` smartmatches `$_` against `\@l`.
			// The condition is the first child before a block and the last
			// after a statement, `print 1 when @l;`. A condition that looks
			// boolean, `!@l` or `@l > 1`, is not an aggregate node and takes
			// nothing, as measured.
			if n.Text == "when" && len(n.Children) > 1 {
				cond := n.Children[0]
				if n.Children[len(n.Children)-1].Kind != Block {
					cond = n.Children[len(n.Children)-1]
				}
				if isAggregate(cond) {
					add(parseoracle.SiteKindReference, cond)
				}
			}

		case Term:
			// A deref block the lexer swallowed hides whatever is inside it,
			// and silence about hidden bytes is the same silent wrong answer
			// a quiet Unknown would be. Hedge every marker and do not look
			// further: a walk over nodes cannot reach text.
			if hidesStructure(n.Text) {
				for _, kind := range markerKinds {
					found = append(found, decided{kind: kind, node: n, unresolved: true})
				}
				return
			}
			// Code inside an interpolating string, regex or heredoc --
			// `${\ f()}`, `@{[ ... ]}`, a subscript, a postfix dereference --
			// is compiled by perl into ordinary ops and never parsed here:
			// the literal is one leaf. The plan's M1 rule reports what is
			// not understood as unknown, so every marker hedges until the
			// embedded code is parsed (01a12059-3011).
			//
			// No return: the leaf is still what it is, and a pattern's own
			// match stays decided below. perl emits refgen AND match for
			// `"ba" =~ /b${\do { qr|a| }}/` (re/reg_eval_scope.t:99).
			//
			// ponytail: a heredoc body is tested whatever its opener, so a
			// `<<'EOF'` body holding `@{[` hedges too. perl reports nothing
			// there, so the hedge is never spent; pairing each body with its
			// opener is the upgrade if the precision ever matters.
			if (n.HeredocBody && interpolatesCode(string(src[n.Start:n.End]))) ||
				(isInterpolatingQuote(n.Text) && interpolatesCode(n.Text)) {
				for _, kind := range markerKinds {
					hedge(kind, n)
				}
			}
			switch {
			case isNamedHash(n.Text):
				// perl: rv2hv for a PACKAGE hash, padhv for a lexical --
				// from identical source text. Undecidable here, so hedged.
				hedge(parseoracle.SiteKindHash, n)
			case isHashTerm(n.Text):
				// perl: rv2hv. A dereference reads through a reference and
				// emits the op however anything was declared.
				add(parseoracle.SiteKindHash, n)
			case isReadlineTerm(n.Text):
				// perl: readline.
				add(parseoracle.SiteKindReadline, n)
			case isBarePattern(n.Text):
				add(parseoracle.SiteKindMatch, n)
			}
		}
		up = append(up, n)
		defer func() { up = up[:len(up)-1] }()
		for i, c := range n.Children {
			// A method NAME is not a call to the builtin of that name.
			// `$obj->readline` parses as Binary "->" whose right operand is
			// a resolved Call, and perl emits entersub there rather than
			// readline. The NAME is skipped because a name's meaning is
			// decided by what precedes it, which the node cannot see.
			//
			// Its ARGUMENTS are still walked. `threads->create(\&f, $i)`
			// holds a reference that perl reports, and skipping the whole
			// subtree lost it -- measured, op/threads.t went from clean to
			// three WRONG srefgen sites and re/pat.t gained two.
			if n.Kind == Binary && n.Text == "->" && i == 1 && c.Kind == Call {
				up = append(up, c)
				for _, arg := range c.Children {
					walk(arg)
				}
				up = up[:len(up)-1]
				continue
			}
			walk(c)
		}
		// A heredoc body inside its statement's span is held beside the
		// children rather than among them, and it interpolates as any other.
		for _, b := range n.InnerHeredocBodies {
			walk(b)
		}
	}
	walk(stmt)
	return found
}

// hidesStructure reports whether a leaf's text contains a construct the
// parser never turned into a node.
//
// `internal/lexer/scan.go:77-94` brace-matches a braced name to its closer
// and emits ONE Variable token; its own comment says the full rule "belongs
// to a later issue" (01a0ad52). So `${[{a=>214}]}` arrives as a single Term
// whose Text holds an anonymous hash that no walk over nodes can reach.
//
// The test is for a nested opener AFTER the leading `${`, `@{` or `$#{`, because
// that is what distinguishes a block with contents from a plain dereference:
//
//	${$x}          nothing hidden -- the inner text is one variable
//	$$x            nothing hidden -- no brace at all
//	${[{a=>1}]}    an array and a hash constructor, both invisible
//	@{$::{$k[0]}}  a hash element and a subscript, both invisible
//
// Deliberately conservative. A false positive costs a hedge where the
// subject could have been silent and correct, which scores wider; a false
// negative is silence over bytes perl found something in, which scores
// WRONG. §7.2's asymmetry says which way to lean.
//
// The `text[1] != '{'` test is DEFENSIVE rather than load-bearing, and
// measured as such: removing it leaves the whole suite green. `$h{k}` parses
// to an Index whose Term child is `$h`, so the braces belong to the Index and
// the leaf text here is never `"$h{k}"`. Kept because this takes a string,
// and a future caller may not have an Index standing between it and the
// source.
func hidesStructure(text string) bool {
	// `$#{ EXPR }` is the same block behind a two-byte sigil, and the lexer
	// swallows it the same way. Measured on 5.42.0: `$#{ \@a }` is a
	// srefgen, `$#{$r->{a}}` a rv2hv, while `$#{$r}` and `$#$r` report none
	// of the five markers. op/goto.t:668 and op/gmagic.t:227 are this shape.
	//
	// The test is the other way round from the bracket scan below: a `\`
	// carries no bracket, so anything but a plain variable or name hedges.
	if strings.HasPrefix(text, "$#{") && strings.HasSuffix(text, "}") {
		inner := strings.TrimSpace(text[3 : len(text)-1])
		inner = strings.TrimLeft(inner, "$")
		return inner == "" || strings.IndexFunc(inner, func(r rune) bool {
			return r != ':' && !isWordByte(byte(r)) && r < utf8.RuneSelf
		}) >= 0
	}
	if len(text) < 3 || text[1] != '{' {
		return false
	}
	switch text[0] {
	case '$', '@', '%':
	default:
		return false
	}
	// Anything bracket-like inside the block is structure this parser did
	// not build. A bare name or a lone variable has none.
	return strings.ContainsAny(text[2:], "{[(")
}

// isInterpolatingQuote reports whether a leaf is a literal perl interpolates:
// `"..."`, a backtick command, a bare `/.../`, or qq, qx, m, qr or s with any
// delimiter but `'`, which turns interpolation off (perlop, "Gory details of
// parsing quoted constructs").
func isInterpolatingQuote(text string) bool {
	if text == "" {
		return false
	}
	switch text[0] {
	case '"', '`', '/':
		return true
	}
	for _, op := range []string{"qq", "qx", "qr", "m", "s"} {
		if quoteOpIs(text, op) {
			delim := strings.TrimLeft(text[len(op):], " \t\n")
			return delim != "" && delim[0] != '\''
		}
	}
	return false
}

// interpolatesCode reports whether an interpolating literal's bytes hold an
// expression rather than a plain variable: a `${ ... }` or `@{ ... }` block
// around anything but a name, or a variable followed by a subscript or a
// dereferencing arrow. Measured on 5.42.0, perl reports no marker for `"$x"`,
// `"@a"`, `"${x}"` or `"$x->method"` (the arrow is text there), and the five
// markers turn up only once there is code.
//
// Deliberately conservative, as hidesStructure is: `"$a[0]"` hedges though
// perl folds it to aelemfast, because a false positive costs a hedge and a
// false negative is silence over bytes perl found something in.
func interpolatesCode(s string) bool {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++ // an escaped byte is literal
		case '$', '@':
			if codeAfterSigil(s[i+1:]) {
				return true
			}
		}
	}
	return false
}

// codeAfterSigil reads what follows an interpolating `$` or `@`.
func codeAfterSigil(rest string) bool {
	rest = strings.TrimLeft(rest, "$#") // `$$r`, `$#a`, `$#{...}`
	if strings.HasPrefix(rest, "{") {
		end := matchingBrace(rest)
		if end < 0 {
			return true // unclosed: whatever it holds, it is not a name
		}
		inner := strings.TrimLeft(strings.TrimSpace(rest[1:end]), "$^")
		if inner == "" || strings.IndexFunc(inner, func(r rune) bool {
			return r != ':' && (r >= utf8.RuneSelf || !isWordByte(byte(r)))
		}) >= 0 {
			return true
		}
		rest = rest[end+1:]
	} else {
		name := len(rest) - len(strings.TrimLeft(rest, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_:"))
		if name == 0 {
			return false // `$)`, `/x$/`, `mail@`: no variable here
		}
		rest = rest[name:]
	}
	if rest != "" && (rest[0] == '[' || rest[0] == '{') {
		return true
	}
	if after, ok := strings.CutPrefix(rest, "->"); ok && after != "" &&
		strings.ContainsRune("[{@$%", rune(after[0])) {
		return true
	}
	return false
}

// matchingBrace returns the index of the `}` closing the `{` at s[0], or -1.
func matchingBrace(s string) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			i++
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// isHashTerm reports whether a leaf's text is a hash read: `%h`, `%$r`.
//
// A leading `%` in TERM position is a sigil and nowhere else -- scanVariable
// refuses it when the expect state does not want a term, which is what keeps
// `$a % $b` from lexing as a hash (§3.2). So the byte is sufficient here
// precisely because the lexer already did the hard part.
func isHashTerm(text string) bool {
	return len(text) > 1 && text[0] == '%'
}

// isNamedHash reports whether a hash access names a hash rather than
// dereferencing one: `%h` and not `%$r`.
//
// The distinction decides whether this parser can COMMIT. Perl emits rv2hv
// for a package hash and padhv -- no marker -- for a lexical, from source
// text that is byte for byte the same:
//
//	my %h;  keys %h    0 rv2hv
//	our %h; keys %h    1 rv2hv
//
// A dereference has no such ambiguity: `%$r` reads through a reference and
// emits the op however anything was declared.
//
// Scope tracking would not settle it either, and that is why this hedges
// rather than resolving. `state %h` is a LEXICAL declaration that behaves
// like a package hash here (1 rv2hv, measured), so a rule keyed on "was it
// declared with my" answers it wrong -- and a lexical closed over by a named
// sub stays padhv, so the rule would need closure analysis to get the
// easy case right.
func isNamedHash(text string) bool {
	return len(text) > 1 && text[0] == '%' &&
		text[1] != '$' && text[1] != '{'
}

// isNamedVariable reports whether a subscript's base names a variable rather
// than holding a reference: `$h` and `@h`, not `$$r` or `${...}`.
//
// Same question isNamedHash asks, one node down. `@h{...}` slices the hash
// %h -- package or lexical, undecidable -- while `@$r{...}` slices through a
// reference and is read by the Unary case above.
func isNamedVariable(text string) bool {
	return len(text) > 1 && (text[0] == '$' || text[0] == '@') &&
		text[1] != '$' && text[1] != '{'
}

// throughReference reports whether a brace subscript reads its hash through
// a reference: `$r->{k}`, `$$r{k}`, or any subscript of a subscript,
// `$r->{a}{k}` and `$h{a}{k}`, whose implicit arrow the tree leaves off.
// The postfix slice `$r->@{...}` is decided on its own and excluded.
func throughReference(n *Node) bool {
	if len(n.Children) == 0 {
		return false
	}
	base := n.Children[0]
	switch {
	case base.Kind == Term && (base.Text == "@" || base.Text == "%"):
		return false
	case n.Arrow, base.Kind == Index:
		return true
	}
	return base.Kind == Unary && base.Text == "$"
}

// chainContext is what perl's op_lvalue does to the last element of a
// subscript chain, which decides how much of the chain multideref folds.
type chainContext int

const (
	// chainUnknown is a context Sites has not measured: the chain hedges.
	chainUnknown chainContext = iota
	// chainRead leaves the last element alone, and the chain folds.
	chainRead
	// chainModScalar is `${...}` modified: the last element is marked
	// OPpDEREF_SV, and multideref folds the levels before it.
	chainModScalar
	// chainModAggregate is `@{...}` or `%{...}` modified, or sliced: the
	// last element is marked OPpDEREF_AV or _HV, and the run of levels
	// holding it does not fold at all.
	chainModAggregate
)

// derefContext classifies the dereference d, of sigil `@`, `%` or `$`,
// from its ancestors, outermost first. Only contexts measured on perl 5.44.0
// are classified (TestArrowHashReadFollowsMultideref); any other is
// chainUnknown, which hedges.
func derefContext(d *Node, sigil string, up []*Node) chainContext {
	mod := chainModAggregate
	if sigil == "$" {
		mod = chainModScalar
	}
	if len(up) == 0 {
		return chainUnknown
	}
	switch p := up[len(up)-1]; {
	case p.Kind == Index && p.Children[0] == d && sigil != "%":
		// `@{X}[0,1]`, `@{X}{qw(a b)}` and `${X}[0]` dereference X as an
		// aggregate whatever the statement does with the result.
		return chainModAggregate
	case p.Kind == Postfix, p.Kind == Unary && p.Text == "ref":
		return mod
	case p.Kind == Unary && p.Text == "!":
		return chainRead
	}

	// The list the dereference is an item of, and whether it is the first.
	node, first, i := d, true, len(up)-1
	for ; i >= 0; i-- {
		p := up[i]
		if p.Kind != List && (p.Kind != Binary || p.Text != "," && p.Text != "=>") {
			break
		}
		first = first && p.Children[0] == node
		node = p
	}
	if i < 0 {
		return chainUnknown
	}
	owner := up[i]
	last := owner.Children[len(owner.Children)-1]
	switch owner.Kind {
	case Call:
		method := i > 0 && up[i-1].Kind == Binary && up[i-1].Text == "->" &&
			len(up[i-1].Children) == 2 && up[i-1].Children[1] == owner
		if method || !owner.Resolved {
			return mod // a sub's arguments
		}
		switch owner.Text {
		case "push", "unshift", "splice", "pop", "shift":
			if first {
				return mod
			}
		case "chomp", "chop", "keys", "values", "each":
			return mod
		case "grep", "map":
			if !first || node != owner.Children[0] {
				return mod // the list, not the block or expression
			}
		case "sprintf":
			if !first {
				return mod
			}
		case "print", "join", "reverse", "scalar":
			return chainRead
		}
	case LoopControl:
		if owner.Text == "return" {
			return chainRead
		}
	case Loop:
		switch {
		case owner.Text == "while" && node == owner.Children[0]:
			return chainRead
		case (owner.Text == "for" || owner.Text == "foreach") &&
			last.Kind == Block && len(owner.Children) <= 3 &&
			node == owner.Children[len(owner.Children)-2]:
			// ponytail: a C-style `for (; COND ;)` that parses to two
			// children would read as foreach here; none is in the corpus.
			return mod
		}
	case Conditional:
		if (owner.Text == "if" || owner.Text == "unless") &&
			(last.Kind == Block && node == owner.Children[0] ||
				last.Kind != Block && node == last) {
			return chainRead
		}
	case Declaration:
		switch {
		case node != owner.Children[0] &&
			(owner.Text == "my" || owner.Text == "our" || owner.Text == "state"):
			return chainRead // the initialiser
		case node == owner.Children[0] && owner.Text == "local":
			return mod
		}
	case Binary:
		switch owner.Text {
		case "=":
			if node == owner.Children[0] {
				return mod
			}
			return chainRead
		case "+=":
			if node == owner.Children[1] {
				return chainRead
			}
		case "+", "-", "*", "/", "==", "<", "||":
			return chainRead
		}
	case CmpChain:
		return chainRead
	}
	return chainUnknown
}

// chainRv2hv counts the rv2hv perl builds for the brace levels of the
// subscript chain topped by top that read through a reference, when ctx
// is what is done with its last element. False when the count is not in
// the text: an unmeasured context, or a key that might constant-fold.
//
// The levels fold into multideref in runs. A key multideref cannot hold
// ends its run, and keeps its level's rv2hv unless the run before it
// folded something. The last level, under a modifying ctx, is the
// exception TestArrowHashReadFollowsMultideref measures.
func chainRv2hv(top *Node, ctx chainContext) (int, bool) {
	var levels []*Node
	for l := top; isChainLevel(l); l = l.Children[0] {
		levels = append([]*Node{l}, levels...)
	}
	rv2hv := func(l *Node) int {
		if l.Text == "{" && throughReference(l) {
			return 1
		}
		return 0
	}
	if len(levels) == 0 {
		// A slice, or a subscript this does not model: through a
		// reference it hedges.
		return 0, top.Kind != Index || rv2hv(top) == 0
	}
	refLevels := 0
	for _, l := range levels {
		refLevels += rv2hv(l)
	}
	if refLevels == 0 {
		return 0, true // a named element: the declaration's question
	}

	count := 0
	var run []*Node
	for i, l := range levels {
		key := l.Children[1]
		switch {
		case !isSimpleKey(key):
			if !nonConstant(key) {
				return 0, false
			}
			if len(run) == 0 {
				count += rv2hv(l)
			}
			run = nil
		case i < len(levels)-1 || ctx == chainRead:
			run = append(run, l)
		case ctx == chainModScalar:
			if len(run) == 0 {
				count += rv2hv(l)
			}
		case ctx == chainModAggregate:
			for _, r := range append(run, l) {
				count += rv2hv(r)
			}
		default:
			return 0, false
		}
	}
	return count, true
}

// isChainLevel reports whether n is one `{key}` or `[index]` subscript of
// an element chain, rather than a slice or a postfix slice.
func isChainLevel(n *Node) bool {
	if n.Kind != Index || n.Text != "{" && n.Text != "[" || len(n.Children) != 2 {
		return false
	}
	base := n.Children[0]
	switch base.Kind {
	case Term:
		return base.Text == "" || base.Text[0] != '@' && base.Text[0] != '%'
	case Unary:
		return base.Text != "@" && base.Text != "%"
	}
	return true
}

// nonConstant reports whether a subscript key can never constant-fold,
// because it reads a variable or an element, or shifts or pops one. A key
// it is unsure of -- a call to a sub that may be a constant, `||` over
// constants -- is not, and hedges.
func nonConstant(k *Node) bool {
	switch k.Kind {
	case Index:
		return true
	case Term:
		t := k.Text
		if len(t) > 1 && (t[0] == '$' || t[0] == '@') {
			return true
		}
		return (t != "" && t[0] == '"' || strings.HasPrefix(t, "qq")) && interpolatesVariable(t)
	case Call:
		return k.Resolved && (k.Text == "shift" || k.Text == "pop")
	case Ternary:
		return len(k.Children) > 0 && nonConstant(k.Children[0])
	case Unary:
		return len(k.Children) == 1 && nonConstant(k.Children[0])
	case Binary:
		switch k.Text {
		case "&&", "||", "//", "and", "or", "xor":
			return false // these fold on a constant operand alone
		}
		for _, c := range k.Children {
			if nonConstant(c) {
				return true
			}
		}
	}
	return false
}

// interpolatesVariable reports whether a double-quoted literal's text holds
// an unescaped `$` or `@` followed by a name or a brace.
func interpolatesVariable(t string) bool {
	for i := 0; i+1 < len(t); i++ {
		switch t[i] {
		case '\\':
			i++
		case '$', '@':
			if c := t[i+1]; isWordByte(c) || c == '{' || c == ':' {
				return true
			}
		}
	}
	return false
}

// isSimpleKey reports whether a hash subscript's key is one perl's
// multideref takes as it stands: a number, a bareword, `-bareword`, a
// string that interpolates nothing, or a plain scalar variable. Measured
// on 5.42.0, `$p->{a}`, `$p->{-bar}`, `$p->{'a$b'}`, `$p->{1.5}`,
// `$p->{$k}`, `$p->{$_}`, `$p->{$1}` and `$p->{$Foo::x}` all fold.
//
// Conservative the safe way: a key it does not recognise hedges, which
// scores wider, never WRONG.
func isSimpleKey(k *Node) bool {
	if k.Kind == Unary && k.Text == "neg" && len(k.Children) == 1 {
		k = k.Children[0]
		if k.Kind == Call && len(k.Children) == 0 {
			return true // `-bar`, a string to perl
		}
	}
	if k.Kind != Term || k.Text == "" {
		return false
	}
	t := k.Text
	switch {
	case t[0] == '$':
		return len(t) > 1 && strings.IndexFunc(t[1:], func(r rune) bool {
			return r != ':' && (r >= utf8.RuneSelf || !isWordByte(byte(r)))
		}) < 0
	case t[0] == '\'':
		return true
	case t[0] == '"' || strings.HasPrefix(t, "qq"):
		return !strings.ContainsAny(t, "$@")
	case strings.HasPrefix(t, "q") && len(t) > 1 && !isWordByte(t[1]):
		return true
	}
	return strings.IndexFunc(t, func(r rune) bool {
		return r != '.' && r != ':' && (r >= utf8.RuneSelf || !isWordByte(byte(r)))
	}) < 0
}

// isAggregate reports whether a node is a whole array or hash -- what perl's
// ref_array_or_hash takes a reference to: `@l`, `%h`, `@_`, `@$r`, `%{$r}`,
// `$r->@*`, `$r->%*`. A slice is an Index and an element a scalar, so neither
// reaches the reference, measured.
func isAggregate(n *Node) bool {
	switch n.Kind {
	case Term:
		return len(n.Text) > 1 && (n.Text[0] == '@' || n.Text[0] == '%')
	case Unary:
		return n.Text == "@" || n.Text == "%"
	case Binary:
		if n.Text == "->" && len(n.Children) == 2 && n.Children[1].Kind == Term {
			t := n.Children[1].Text
			return t == "@*" || t == "%*"
		}
	}
	return false
}

// isReadlineTerm reports whether a leaf is `<FH>`, `<$fh>` or `<>`.
//
// Same reasoning as the other leaf tests: scanAngle emits a Readline token
// only in term position, so a leaf whose text is angle-delimited is the
// lexer's own decision. But angle-delimited is not enough -- `<*.c>` is a
// GLOB, a different op, and reporting a readline there is a claim perl
// contradicts. Measured on 5.42.0:
//
//	<>          readline      <*.c>       glob
//	<$fh>       readline      <$h{x}>     glob
//	<FH>        readline      <a b>       glob
//	<STDIN>     readline
//	<My::Handle> readline
//
// The rule perl applies: readline iff the content is empty, a plain
// `$scalar`, or a bareword identifier. Anything else -- a subscript, a
// space, a metacharacter -- is a glob pattern.
func isReadlineTerm(text string) bool {
	if len(text) < 2 || text[0] != '<' || text[len(text)-1] != '>' {
		return false
	}
	inner := text[1 : len(text)-1]
	if inner == "" {
		return true // <>
	}
	if inner[0] == '$' {
		inner = inner[1:]
		if inner == "" {
			return false // `<$>` is not a handle
		}
	}
	// A bareword identifier, `::` allowed for a package-qualified handle.
	//
	// Runes, not bytes. `use utf8` widens the identifier class, and
	// t/uni/readline.t:64 reads `<hòฟ>` -- which perl compiles to readline,
	// measured. An ASCII-only loop calls that a glob and loses the site,
	// the same defect the lexer's quote-delimiter guard had (0db016fd).
	// A single `:` is accepted, not only `::`. Measured: `<a:b>` compiles to
	// readline, so perl's rule here is looser than its rule for a package
	// name and this follows perl rather than the tidier guess.
	for _, r := range inner {
		switch {
		case r == '_' || r == ':':
		case r < utf8.RuneSelf:
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
				return false
			}
		case !unicode.IsLetter(r) && !unicode.IsDigit(r):
			return false
		}
	}
	return true
}

// isBarePattern reports whether a leaf is a match rather than a substitution
// or a transliteration.
//
// perl reports s/// with subst and tr/// with trans, neither of which is the
// match marker, so they must not be counted. `qr//` compiles a pattern
// without matching it and is excluded for the same reason.
func isBarePattern(text string) bool {
	switch {
	case quoteOpIs(text, "s"), quoteOpIs(text, "tr"),
		quoteOpIs(text, "y"), quoteOpIs(text, "qr"):
		return false
	case quoteOpIs(text, "m"):
		return true
	case len(text) >= 2 && text[0] == '/':
		// A trailing FLAG still leaves a match. Requiring the text to end
		// in `/` reported `/abc/` and missed `/abc/g`, which is every
		// pattern in re/pat.t -- 35 of its sites, and 50 across the corpus.
		//
		// The closing delimiter must still be there, so this looks for a
		// second `/` rather than trusting the leading one: the lexer emits
		// a Quote token only where a pattern is possible, but a leaf whose
		// text merely begins with `/` and never closes is not one.
		return strings.IndexByte(text[1:], '/') >= 0
	}
	return false
}

// quoteOpIs reports whether a leaf is the quote-like operator `name` with
// any delimiter.
//
// perl takes the next non-whitespace character as the delimiter with no
// allow-list, so enumerating delimiters is enumerating a set that has no
// end. Measured on 5.42.0:
//
//	$ perl -MO=Concise,-exec -e 'my $a = m(x); my $b = m[x]; my $c = m!x!;'
//	-- three match ops
//
// The delimiter must not be a word character, or `qw(a b)` reads as `q`
// with `w` for a delimiter and `tr` as `t` with `r`. The lexer already
// settled this the same way (quote.go: "a word character immediately after
// the keyword means it is part of a longer name"), so this asks the same
// question of the token it produced.
//
// Whitespace needs no special case. perl's skipspace runs before delimiter
// selection, so `m /x/` is a match -- and a space is not a word character,
// so it passes this test on its way to being the delimiter's stand-in.
// Measured:
//
//	$ perl -MO=Concise,-exec -e 'my $a = m /x/;'
//	3  </> match(/"x"/) s
//
// comp/opsubs.t:119 is `isnt( m('unqualified'), ... )`, which scored WRONG
// while only `m/` and `m{` counted, and `s(a)(b)` decided MATCH -- a false
// positive -- while only `s/` and `s{` were excluded.
func quoteOpIs(text, name string) bool {
	if !strings.HasPrefix(text, name) {
		return false
	}
	i := len(name)
	return i < len(text) && !isWordByte(text[i])
}

// isWordByte is perl's \w for an identifier: a quote operator's delimiter
// may not be one, because the byte would belong to a longer name.
func isWordByte(c byte) bool {
	return c == '_' || ('0' <= c && c <= '9') ||
		('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

// isMatchOperand reports whether the right side of `=~` is a match rather
// than a substitution or transliteration.
//
// A VARIABLE on the right is a match too: `$got =~ $expected` compiles
// $expected as a pattern, and perl reports it with the same match op. That
// is not a guess about the variable's contents -- `=~` against anything that
// is not s/// or tr/// is a match, whatever the pattern turns out to be.
// Measured on perl 5.42.0:
//
//	$ perl -MO=Concise,-exec -e 'my ($g,$e); my $r = $g =~ $e;'
//	7  </> match()[$g:1,3] sK
//
// t/comp/use.t:24 and t/comp/uproto.t:24 are this shape and scored WRONG
// while only literals counted.
// The operand's SHAPE is not the question. Requiring a Term meant a binding
// against anything built -- a ternary of two `qr//`, a parenthesised
// expression, a concatenation -- decided nothing. Measured:
//
//	$ perl -MO=Concise,-exec -e 'my $x; print $x =~ ((1 & 1) ? qr/^$/ : qr/o/);'
//	  one match, one qr, one regcomp
//
// One match for the binding, whatever built the pattern. t/cmd/for.t:79 is
// that shape. Only a LEAF that states it is a substitution or a
// transliteration excludes, because those spellings name a different op.
func isMatchOperand(n *Node) bool {
	if n.Kind == Term && isSubstOrTrans(n.Text) {
		return false
	}
	return true
}

// isSubstOrTrans reports whether a leaf is a substitution or a
// transliteration rather than a match.
func isSubstOrTrans(text string) bool {
	return quoteOpIs(text, "s") || quoteOpIs(text, "tr") || quoteOpIs(text, "y")
}

// lineStarts is the byte offset of each line's first byte, so a span can be
// reported in the lines the oracle speaks.
type lineStarts []int

func lineIndex(src []byte) lineStarts {
	starts := lineStarts{0}
	for i := 0; ; {
		j := bytes.IndexByte(src[i:], '\n')
		if j < 0 {
			break
		}
		i += j + 1
		starts = append(starts, i)
	}
	return starts
}

// at returns the 1-based line number containing byte offset off.
func (ls lineStarts) at(off int) int {
	if off < 0 {
		return 1
	}
	// Binary search would be faster; a linear scan is fine for the sizes
	// involved and is obviously correct.
	line := 1
	for i, s := range ls {
		if s > off {
			break
		}
		line = i + 1
	}
	return line
}

// PerlLines numbers src's lines as perl does. For each physical line i
// (1-based; index 0 is unused), line[i] is the line perl records for it and
// own[i] is whether perl still attributes it to file, the path perl was
// given.
//
// A `#line N "name"` directive renumbers what follows it, and a name other
// than file moves what follows into another file -- which the oracle's
// CvFILE filter then leaves out of this file's sites. perl's lexer looks for
// one (toke.c S_incline) at the start of every line it reaches while
// skipping whitespace and comments, and at every line of POD; it never
// looks inside a string, a heredoc body or a regex. So a directive is a
// Comment token starting its line, or a line inside a Pod token.
//
// The parse tree's own spans stay physical: this is the harness's
// numbering, not the parser's.
func PerlLines(src []byte, file string) (line []int, own []bool) {
	starts := lineIndex(src)
	directives := map[int]string{}
	for _, tok := range lexer.TokenizeBarewords(src, coreBarewordOperators()) {
		switch tok.Kind {
		case lexer.Comment:
			if tok.Start == 0 || src[tok.Start-1] == '\n' {
				directives[starts.at(tok.Start)] = strings.TrimSuffix(string(src[tok.Start:tok.End]), "\n")
			}
		case lexer.Pod:
			at := tok.Start
			for _, text := range strings.SplitAfter(string(src[tok.Start:tok.End]), "\n") {
				if strings.HasPrefix(text, "#") {
					directives[starts.at(at)] = strings.TrimSuffix(text, "\n")
				}
				at += len(text)
			}
		}
	}

	line = make([]int, len(starts)+1)
	own = make([]bool, len(starts)+1)
	n, inFile := 0, true
	for i := 1; i <= len(starts); i++ {
		n++
		line[i], own[i] = n, inFile
		if num, name, ok := lineDirective(directives[i]); ok {
			n = num - 1
			if name != "" {
				inFile = name == file
			}
		}
	}
	return line, own
}

// lineDirective reads one line as toke.c's S_incline does: `#`, optional
// blanks, `line`, at least one blank, the number, then optionally a name --
// quoted, which may hold blanks, or bare up to the next space -- and nothing
// after it but blanks. Anything else is an ordinary comment. An unclosed
// quote is part of a bare name, and `""` names no file.
func lineDirective(text string) (num int, name string, ok bool) {
	s, found := strings.CutPrefix(text, "#")
	if !found {
		return 0, "", false
	}
	s, found = strings.CutPrefix(strings.TrimLeft(s, " \t"), "line")
	if !found || s == "" || (s[0] != ' ' && s[0] != '\t') {
		return 0, "", false
	}
	s = strings.TrimLeft(s, " \t")
	digits := len(s) - len(strings.TrimLeft(s, "0123456789"))
	if digits == 0 {
		return 0, "", false
	}
	num, err := strconv.Atoi(s[:digits])
	s = s[digits:]
	if err != nil || (s != "" && !strings.ContainsRune(" \t\r", rune(s[0]))) {
		return 0, "", false
	}
	s = strings.TrimLeft(s, " \t")
	rest := ""
	if quoted, found := strings.CutPrefix(s, `"`); found && strings.Contains(quoted, `"`) {
		name, rest, _ = strings.Cut(quoted, `"`)
	} else {
		end := strings.IndexAny(s, " \t\n\r\f\v")
		if end < 0 {
			end = len(s)
		}
		name, rest = s[:end], s[end:]
	}
	if strings.TrimLeft(rest, " \t\r\f") != "" {
		return 0, "", false
	}
	return num, name, true
}
