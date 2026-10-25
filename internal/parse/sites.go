// ABOUTME: The subject side of the fidelity harness: what this parser reports about each statement.
// ABOUTME: An Unknown hedges every marker, because saying nothing is scored as claiming nothing is there.

package parse

import (
	"bytes"
	"regexp"
	"slices"
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
	decls := newNamedHashes(root, src)
	for _, n := range root.Children {
		start, end := lines.at(n.Start), lines.at(n.End-1)

		// An Unknown statement reports through decidedMarkers too, which
		// hedges every marker for it.
		//
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
		for _, d := range decidedMarkers(n, src, decls) {
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
// settles every marker the only way it can: by refusing, out loud. A
// hedge's reason says why the text cannot settle it.
type decided struct {
	kind       string
	node       *Node
	unresolved bool
	reason     hedgeReason
}

// hedgeReason is why a site hedges. Every hedge names one, from the closed
// set below (TestEveryRv2hvHedgeNamesItsReason).
type hedgeReason string

const (
	// reasonUnread: a node this parser did not read, an Unknown.
	reasonUnread hedgeReason = "unread"
	// reasonHiddenStructure: a construct the lexer kept inside one token.
	reasonHiddenStructure hedgeReason = "hidden structure"
	// reasonInterpolatedCode: code inside an interpolating string.
	reasonInterpolatedCode hedgeReason = "interpolated code"
	// reasonUnmeasuredContext: a chain whose context Sites has not measured.
	reasonUnmeasuredContext hedgeReason = "unmeasured context"
	// reasonKeyMayFold: a key that may constant-fold, as a call to a sub
	// declared `sub f(){1}` does.
	reasonKeyMayFold hedgeReason = "key may fold"
	// reasonFeatureUnknown: a hash declared by `state` or a signature
	// where a module or BEGIN block may have changed the feature.
	reasonFeatureUnknown hedgeReason = "feature unknown"
	// reasonMayBeImported: a package hash used after a module's import or
	// a BEGIN block, either of which may have declared a lexical %h.
	reasonMayBeImported hedgeReason = "may be imported"
	// reasonDeclaredInUnread: a package hash whose name appears in code
	// this parser did not read, which may declare it.
	reasonDeclaredInUnread hedgeReason = "declared in unread code"
	// reasonConstantCondition: a site under a condition that may be
	// constant, whose dead branch perl folds away.
	reasonConstantCondition hedgeReason = "constant condition"
)

// hedgeReasons is the closed set.
var hedgeReasons = []hedgeReason{
	reasonUnread, reasonHiddenStructure, reasonInterpolatedCode,
	reasonUnmeasuredContext, reasonKeyMayFold, reasonFeatureUnknown,
	reasonMayBeImported, reasonDeclaredInUnread, reasonConstantCondition,
}

// namedHashes is what Sites knows about a file beyond one statement: the
// declaration each named hash use resolves to, the names code it did not
// read mentions with a `%`, the brace subscripts read after an import
// point, and the builtins the file may override.
type namedHashes struct {
	uses        map[*Node]hashUse
	unread      map[string]bool
	afterImport map[*Node]bool
	overridden  map[string]bool
}

func newNamedHashes(root *Node, src []byte) namedHashes {
	d := namedHashes{unread: unreadHashNames(root, src), overridden: map[string]bool{}}
	d.uses, d.afterImport = hashScopes(root)
	for _, m := range overridesBuiltin.FindAllSubmatch(src, -1) {
		d.overridden[string(m[1])+string(m[2])] = true
	}
	return d
}

// overridesBuiltin finds the subs a file defines and the CORE::GLOBAL
// names it mentions, either of which may override a builtin: `use subs
// "length"; sub length {1}` and `*CORE::GLOBAL::length = sub {1}` both
// keep `$r->{length "abc"}`'s rv2hv, measured on 5.44.0.
var overridesBuiltin = regexp.MustCompile(`\bsub\s+(?:CORE::GLOBAL::)?(\w+)|CORE::GLOBAL::(\w+)`)

// rv2hv counts the rv2hv perl builds for the named hash n -- a `%h`, or the
// `$h` or `@h` a brace subscript reads %h through -- standing alone: 1 for
// a package hash, none for a lexical, which is padhv. A non-empty reason
// is a hedge.
func (d namedHashes) rv2hv(n *Node) (int, hedgeReason) {
	u, ok := d.uses[n]
	switch {
	case !ok:
		// The resolver walks every node a named hash can sit in; one it
		// did not reach is code it did not read.
		return 0, reasonUnread
	case u.Unsure,
		u.Declares && u.Decl == "state" && u.State == stateUnknown,
		u.Declares && u.Decl == "signature" && u.Signatures == stateUnknown:
		return 0, reasonFeatureUnknown
	case u.Declares && (u.Decl == "our" || u.Decl == "local"):
		// The declaration's own `rv2hv vK/OURINTR` or `vKM/LVINTRO`.
		return 1, ""
	case u.Declares, u.Decl == "my", u.Decl == "state", u.Decl == "field":
		// A lexical, or `state % h` where the feature is off.
		return 0, ""
	case u.Imported:
		return 0, reasonMayBeImported
	case d.unread[hashName(n.Text)]:
		return 0, reasonDeclaredInUnread
	}
	return 1, ""
}

// unreadHashNames collects the hash names that code this parser did not
// read mentions after a `%`, as `my %h` and `my \%h` do: an Unknown's
// bytes, a token hiding structure, and interpolated code.
func unreadHashNames(root *Node, src []byte) map[string]bool {
	names := map[string]bool{}
	var walk func(*Node)
	walk = func(n *Node) {
		text := ""
		switch {
		case n.Kind == Unknown:
			text = string(src[n.Start:n.End])
		case n.Kind == Term && (hidesStructure(n.Text) ||
			isInterpolatingQuote(n.Text) && interpolatesCode(n.Text)):
			text = n.Text
		case n.HeredocBody && interpolatesCode(string(src[n.Start:n.End])):
			text = string(src[n.Start:n.End])
		}
		for _, m := range unreadHashName.FindAllStringSubmatch(text, -1) {
			names[m[1]] = true
		}
		for _, c := range n.Children {
			walk(c)
		}
		for _, b := range n.InnerHeredocBodies {
			walk(b)
		}
	}
	walk(root)
	return names
}

var unreadHashName = regexp.MustCompile(`%\s*(\w+)`)

// decidedMarkers reports which of the five markers this statement contains,
// in the subject vocabulary. Each is a conclusion the parser committed to.
//
// Only what the tree SHOWS, never what the bytes suggest. `$a % $b` is a
// Binary whose Text is "%", not a hash; `$a / $b` is division, not a
// pattern; `$a < $c` is a comparison, not a readline. The lexer settled all
// three by position (§3.2) and the tree records the answer, so reading the
// tree cannot reach a different one. Scanning source text could.
func decidedMarkers(stmt *Node, src []byte, decls namedHashes) []decided {
	var found []decided

	// One entry per OCCURRENCE, not per kind. CompareFacts consumes a decided
	// site for each of perl's (`compare_facts.go:224`, `owner.decided[m]--`),
	// so a statement holding two anonymous hashes and reporting one site
	// scores the first exact and the second WRONG.
	//
	// Deduplicating here was the cause of 11 of the 20 WRONG files in T2's
	// first subject measurement -- `eq_hash({$o->h}, {qw( the hash )})` in
	// t/class/accessor.t:31 is two anonhash ops to perl and was one site here.
	//
	// Under a condition perl may fold away, a hash site is hedged instead:
	// the ops of a dead branch are never built. Measured on 5.44.0,
	// `our %h; if (0) { keys %h }` and `sub D(){0} if (D) { keys %h }`
	// count only the declaration's rv2hv.
	dead := 0
	add := func(kind string, n *Node) {
		if dead > 0 && kind == parseoracle.SiteKindHash {
			found = append(found, decided{kind: kind, node: n, unresolved: true, reason: reasonConstantCondition})
			return
		}
		found = append(found, decided{kind: kind, node: n})
	}

	// hedge reports a site this parser SEES but cannot settle. It scores
	// wider rather than WRONG, which is the honest bucket for "something
	// here may or may not emit the op, and the answer is not in the text".
	hedge := func(kind string, n *Node, why hedgeReason) {
		found = append(found, decided{kind: kind, node: n, unresolved: true, reason: why})
	}
	hedgeAll := func(n *Node, why hedgeReason) {
		for _, kind := range markerKinds {
			hedge(kind, n, why)
		}
	}
	// named reports a named hash's rv2hv, decided or hedged.
	named := func(n *Node) {
		count, why := decls.rv2hv(n)
		if why != "" {
			hedge(parseoracle.SiteKindHash, n, why)
			return
		}
		for ; count > 0; count-- {
			add(parseoracle.SiteKindHash, n)
		}
	}

	// arrowChain reports the rv2hv perl builds for the subscript chain
	// topped by n, in context ctx: a decided site for each, or one hedge
	// where the count is not in the text. Each chain is reported once, by
	// the dereference around it or, failing one, at its top.
	reported := map[*Node]bool{}
	var arrowChain func(n *Node, ctx chainContext)
	arrowChain = func(n *Node, ctx chainContext) {
		if reported[n] {
			return
		}
		reported[n] = true
		switch {
		case n.Kind == Ternary && len(n.Children) == 3:
			// Each branch takes the context: perl propagates op_lvalue
			// into both (TestArrowHashReadTernaryBranchesTakeTheContext).
			// A condition that may be constant folds one branch away.
			if possiblyConstant(n.Children[0]) {
				hedge(parseoracle.SiteKindHash, n, reasonConstantCondition)
				return
			}
			arrowChain(n.Children[1], ctx)
			arrowChain(n.Children[2], ctx)
			return
		case n.Kind == Call && (n.Text == "do" || n.Text == "eval") &&
			slices.ContainsFunc(n.Children, func(c *Node) bool { return c.Kind == Block }):
			// The block's last statement takes it, and keeps an rv2hv in
			// some contexts and not others; unmeasured beyond that.
			hedge(parseoracle.SiteKindHash, n, reasonUnmeasuredContext)
			return
		case n.Kind == Unary && n.Text == "pos" && len(n.Children) == 1:
			arrowChain(n.Children[0], ctx) // unary plus builds no op
			return
		case ctx != chainRead && !isChainLevel(n) && n.Kind != Index &&
			!yieldsFreshValue(n) && containsBraceIndex(n):
			// A modified dereference of an expression holding a hash
			// element: perl passes the context on through some operators
			// (`scalar`, a list's last item) and not others, measured
			// beside yieldsFreshValue.
			hedge(parseoracle.SiteKindHash, n, reasonUnmeasuredContext)
			return
		}
		count, why := chainRv2hv(n, ctx, decls)
		if why != "" {
			hedge(parseoracle.SiteKindHash, n, why)
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
		if constantGuard(n) {
			dead++
			defer func() { dead-- }()
		}
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
			hedgeAll(n, reasonUnread)
			return

		case AnonHash:
			// perl: anonhash. The lexer's brace stack already chose term
			// over block (§4.9.2), so this node IS the decision.
			add(parseoracle.SiteKindAnonhash, n)

		case Index:
			// A BRACE subscript reads a hash, and whether perl emits rv2hv
			// for it turns on the same declaration a named `%h` does --
			// plus the optimiser, which folds `$h{a}` to multideref and
			// emits nothing at all. Measured on 5.44.0, the use alone:
			//
			//	my %h;  @h{"a","b"}       0 rv2hv
			//	our %h; @h{"a","b"}       1 rv2hv
			//	our %h; delete $h{a}      0
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

			// A slice of a named hash, `@h{...}`, reads the hash once:
			// rv2hv for a package hash, padhv for a lexical. A named
			// element is a chain level, counted with its chain below.
			if n.Text == "{" && len(n.Children) > 0 &&
				n.Children[0].Kind == Term && !n.Arrow &&
				isNamedVariable(n.Children[0].Text) && !isChainLevel(n) {
				named(n.Children[0])
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

			// A chain read in place: nothing dereferences its last element,
			// so it is decided as a read. Each level below the top is part of
			// the top's chain and reported with it.
			//
			// `delete local` keeps the last level out of multideref as a
			// modified `${...}` does. Measured on 5.44.0:
			//
			//	my $r; delete local $r->{a}       rv2hv=1
			//	my $r; delete local $r->{a}{b}    rv2hv=0
			//	our %h; delete local $h{a}        rv2hv=2, the declaration's and the use's
			if len(up) == 0 || up[len(up)-1].Kind != Index ||
				up[len(up)-1].Children[0] != n {
				ctx := chainRead
				if len(up) > 1 && up[len(up)-1].Kind == Declaration &&
					up[len(up)-1].Text == "local" &&
					up[len(up)-2].Kind == Call && up[len(up)-2].Text == "delete" {
					ctx = chainModScalar
				}
				arrowChain(n, ctx)
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
			// A dereference of an element, `@{$r->{x}}` or `@{$h{x}}`, is
			// counted by the context around the dereference: the chain's
			// levels through a reference, and a package hash's own.
			//
			// The bare slice `@$r{"a","b"}` does not reach here: it is a
			// subscript on the deref, decided in the Index case.
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
				hedgeAll(n, reasonHiddenStructure)
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
				hedgeAll(n, reasonInterpolatedCode)
			}
			switch {
			case isNamedHash(n.Text) && len(up) > 0 && up[len(up)-1].Kind == Index &&
				up[len(up)-1].Text == "[" && up[len(up)-1].Children[0] == n:
				// `%h[0]` slices the array @h, keys and values.
			case isNamedHash(n.Text):
				// perl: rv2hv for a PACKAGE hash, padhv for a lexical --
				// from identical source text, so the declaration decides.
				named(n)
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
			if foldedAway(n, i) {
				dead++
				walk(c)
				dead--
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
// The declaration in effect settles it (hashScopes, namedHashes.rv2hv): a
// `my`, a `field`, a signature parameter and a `state` with its feature on
// are padhv, and an `our`, a `local` and no declaration at all are rv2hv.
// With the feature off, `state %h` declares nothing -- perl reads
// `state % h` -- and the later `%h` is the package's. A lexical closed over
// by a named sub stays padhv, which lexical scope gives without closure
// analysis. `->%*` is a dereference, not a name: isHashTerm decides it.
func isNamedHash(text string) bool {
	return text != "" && text[0] == '%' && text != "%*" && hashName(text) != ""
}

// hashName is the name a sigilled leaf names: `h` for `%h`, `% h`, `%{h}`,
// `%{ h }`, `$h` or `@{h}`, and "" for a dereference, `%$r` or
// `%{ $r }`. Measured on 5.44.0, `my %h; keys % h` and `keys %{ h }` are
// padhv, as `keys %h` is.
func hashName(text string) string {
	if len(text) < 2 {
		return ""
	}
	name := strings.TrimSpace(text[1:])
	if inner, ok := strings.CutPrefix(name, "{"); ok {
		inner, ok = strings.CutSuffix(inner, "}")
		if !ok {
			return ""
		}
		name = strings.TrimSpace(inner)
		if strings.IndexFunc(strings.TrimPrefix(name, "^"), func(r rune) bool {
			return r != ':' && r < utf8.RuneSelf && !isWordByte(byte(r))
		}) >= 0 {
			return ""
		}
	}
	if name == "" || name[0] == '$' || name[0] == '{' {
		return ""
	}
	return name
}

// isNamedVariable reports whether a subscript's base names a variable rather
// than holding a reference: `$h` and `@h`, not `$$r` or `${...}`.
//
// Same question isNamedHash asks, one node down. `@h{...}` slices the hash
// %h -- package or lexical, as its declaration says -- while `@$r{...}`
// slices through a reference and is read by the Unary case above.
func isNamedVariable(text string) bool {
	return text != "" && (text[0] == '$' || text[0] == '@') && hashName(text) != ""
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
		case (owner.Text == "for" || owner.Text == "foreach") && !owner.CStyle &&
			(last.Kind == Block && node == owner.Children[len(owner.Children)-2] ||
				last.Kind != Block && node == last):
			// The list of a foreach, or of a statement-modifier `for`.
			// A C-style head's parts are not a list it aliases.
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

// constantGuard reports whether n is a conditional construct whose
// condition may be constant, so that perl may fold a branch of it away:
// `if`, `unless`, `elsif`, `while` and `until`, as blocks or statement
// modifiers, a `?:`, and `&&`, `||`, `//`, `and`, `or` and `xor`, whose
// left operand is the condition.
func constantGuard(n *Node) bool {
	if len(n.Children) == 0 {
		return false
	}
	if n.Kind == Loop && n.CStyle {
		// The tree does not say which head part is the condition, so any
		// part that may be constant, and is not a known true, guards it:
		// `for (;0;) { keys %h }` builds no rv2hv for the use, where
		// `for (;;)` and `for (;1;)` keep it.
		return slices.ContainsFunc(n.Children, func(c *Node) bool {
			truth, known := literalTruth(c)
			return c.Kind != Block && possiblyConstant(c) && !(known && truth)
		})
	}
	cond, _, _ := guardCondition(n)
	if cond == nil || !possiblyConstant(cond) {
		return false
	}
	_, known := literalTruth(cond)
	return !known || n.Kind == Binary && n.Text == "xor"
}

// guardCondition is the condition of a construct constantGuard reads, and
// whether it is a statement modifier, whose condition is its last child.
func guardCondition(n *Node) (cond *Node, modifier, ok bool) {
	switch {
	case n.Kind == Conditional && n.Text != "else", n.Kind == Loop && (n.Text == "while" || n.Text == "until"):
		last := n.Children[len(n.Children)-1]
		switch {
		case len(n.Children) > 1 && n.Children[1].Kind == Block:
			return n.Children[0], false, true
		case last.Kind != Block:
			return last, true, true
		}
	case n.Kind == Ternary:
		return n.Children[0], false, true
	case n.Kind == Binary:
		switch n.Text {
		case "&&", "||", "//", "and", "or", "xor":
			return n.Children[0], false, true
		}
	}
	return nil, false, false
}

// foldedAway reports whether perl folds n's child i away under a literal
// condition, whose truth is known: the block of `if (0)` or `while (0)`,
// the `else` of `if (1)`, the statement of `X if 0`, the unchosen branch
// of `1 ? A : B`, and the right operand of `0 && X`, `1 || X` or
// `1 // X`. Measured on 5.44.0, `our %h; my $n = 1 && %h` builds the
// use's rv2hv (op/each.t:233), `while (1) { keys %$r }` keeps its body,
// and `do { keys %h } until 1` runs its block once.
func foldedAway(n *Node, i int) bool {
	// An `elsif` that may be constant may fold the branches after it:
	// `if ($c) {...} elsif (1) {...} else { keys %h }` builds no rv2hv
	// for the else, measured on 5.44.0.
	if n.Kind == Conditional && slices.ContainsFunc(n.Children[:min(i, len(n.Children))], func(c *Node) bool {
		if c.Kind != Conditional || c.Text != "elsif" {
			return false
		}
		cond, _, ok := guardCondition(c)
		if !ok {
			return false
		}
		truth, known := literalTruth(cond)
		return possiblyConstant(cond) && !(known && !truth)
	}) {
		return true
	}
	cond, modifier, ok := guardCondition(n)
	if !ok || len(n.Children) == 0 {
		return false
	}
	truth, known := literalTruth(cond)
	if !known {
		return false
	}
	if n.Text == "unless" || n.Text == "until" {
		truth = !truth
	}
	switch {
	case n.Kind == Binary:
		switch n.Text {
		case "&&", "and":
			return i == 1 && !truth
		case "||", "or":
			return i == 1 && truth
		case "//":
			return i == 1 // a literal is defined
		}
		return false
	case n.Kind == Ternary:
		return truth && i == 2 || !truth && i == 1
	case modifier:
		body := n.Children[0]
		doBlock := body.Kind == Call && body.Text == "do"
		return !truth && i < len(n.Children)-1 && !(n.Kind == Loop && doBlock)
	case !truth:
		return i == 1
	}
	return n.Kind == Conditional && i >= 2
}

// literalTruth is a plain number or quoted literal's truth: "", "0" and
// a number equal to zero are false. Anything else is not known.
func literalTruth(n *Node) (truth, known bool) {
	if n.Kind != Term || n.Text == "" {
		return false, false
	}
	t := n.Text
	if isNumber(n) {
		f, err := strconv.ParseFloat(strings.ReplaceAll(t, "_", ""), 64)
		if err != nil {
			return false, false
		}
		return f != 0, true
	}
	if len(t) >= 2 && (t[0] == '\'' || t[0] == '"') && t[len(t)-1] == t[0] &&
		!strings.ContainsAny(t[1:len(t)-1], "\\$@") {
		v := t[1 : len(t)-1]
		return v != "" && v != "0", true
	}
	return false, false
}

// possiblyConstant reports whether an expression may be a compile-time
// constant: a literal, a call with no arguments (`D` after `sub D(){0}`, or
// a `use constant`), or an operator over those. Measured on 5.44.0,
// `use constant C => 1; push @{ C ? $r->{a} : $r->{b} }, 1` builds one
// rv2hv, where a variable condition builds two.
func possiblyConstant(n *Node) bool {
	switch n.Kind {
	case Term:
		return isLiteral(n)
	case Call:
		return len(n.Children) == 0
	case Unary:
		return len(n.Children) == 1 && possiblyConstant(n.Children[0])
	case Binary, CmpChain, Ternary, List:
		// An operator, a comparison chain, a `?:` or a parenthesised
		// list folds when every operand does.
		return len(n.Children) > 0 && !slices.ContainsFunc(n.Children, func(c *Node) bool { return !possiblyConstant(c) })
	}
	return false
}

// yieldsFreshValue reports whether an expression's value is not one of
// its operands' lvalues, so a dereference's context stops there. Measured
// on perl 5.44.0, each keeping no rv2hv for `$x->{a}`:
//
//	push @{ $x->{a} // [] }, 1      push @{ $x->{a} //= [] }, 1
//	push @{ $x->{a} ||= [] }, 1     push @{ $r = $x->{a} }, 1
//	push @{ $c && $x->{a} }, 1      push @{ [$x->{a}] }, 1
//	push @{ \$x->{a} }, 1
//
// where `push @{ scalar $x->{a} }, 1`, `push @{ +$x->{a} }, 1` and
// `keys %{ (1, $x->{a}) }` keep it.
func yieldsFreshValue(n *Node) bool {
	switch n.Kind {
	case AnonArray, AnonHash:
		return true
	case Unary:
		return n.Text == "ref"
	case Binary:
		switch n.Text {
		case "//", "||", "&&", "//=", "||=", "&&=", "=":
			return true
		}
	}
	return false
}

// containsBraceIndex reports whether n holds a brace subscript.
func containsBraceIndex(n *Node) bool {
	if n.Kind == Index && n.Text == "{" {
		return true
	}
	return slices.ContainsFunc(n.Children, containsBraceIndex)
}

// chainRv2hv counts the rv2hv perl builds for the brace levels of the
// subscript chain topped by top that read through a reference or a
// package hash, when ctx is what is done with its last element. A reason
// when the count is not in the text: an unmeasured context, a key that
// might constant-fold, or a named hash whose declaration may be elsewhere.
//
// A package hash's element folds as an element through a reference does,
// measured on 5.44.0: `our %h; $h{a}{b}` is multideref with no rv2hv for
// the use, `our %h; $h{"a$k"}` keeps one, and `our %h; push @{$h{a}}, 1`
// keeps one, where `my %h` keeps none in any of them.
//
// The levels fold into multideref in runs. A key multideref cannot hold
// ends its run, and keeps its level's rv2hv unless the run before it
// folded something. The last level, under a modifying ctx, is the
// exception TestArrowHashReadFollowsMultideref measures.
func chainRv2hv(top *Node, ctx chainContext, decls namedHashes) (int, hedgeReason) {
	var levels []*Node
	for l := top; isChainLevel(l); l = l.Children[0] {
		levels = append([]*Node{l}, levels...)
	}
	if len(levels) == 0 {
		// A slice, or a subscript this does not model: through a
		// reference it hedges.
		if top.Kind == Index && top.Text == "{" && throughReference(top) {
			return 0, reasonUnmeasuredContext
		}
		return 0, ""
	}
	// The bottom level reads a named hash, or through a reference.
	base := 0
	if b := levels[0]; b.Text == "{" && !b.Arrow && b.Children[0].Kind == Term &&
		isNamedVariable(b.Children[0].Text) {
		n, why := decls.rv2hv(b.Children[0])
		if why != "" {
			return 0, why
		}
		base = n
	}
	rv2hv := func(l *Node) int {
		switch {
		case l == levels[0] && base > 0:
			return 1
		case l.Text == "{" && throughReference(l):
			return 1
		}
		return 0
	}
	refLevels := 0
	for _, l := range levels {
		refLevels += rv2hv(l)
	}
	if refLevels == 0 {
		return 0, "" // a lexical's elements, or arrays'
	}

	count := 0
	var run []*Node
	for i, l := range levels {
		key := l.Children[1]
		switch {
		case !isSimpleKey(key):
			if !nonConstant(key) {
				return 0, reasonKeyMayFold
			}
			if len(run) == 0 {
				count += rv2hv(l)
			}
			run = nil
		case decls.keyMayNotFold(l, key):
			return 0, reasonKeyMayFold
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
			return 0, reasonUnmeasuredContext
		}
	}
	return count, ""
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
// because it reads a variable or an element, shifts or pops one, or
// passes one to a call. A key
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
		// A sub perl inlines as a constant takes no arguments, and a
		// builtin folds only over constants.
		if k.Resolved && (k.Text == "shift" || k.Text == "pop") {
			return true
		}
		for _, c := range k.Children {
			if nonConstant(c) {
				return true
			}
		}
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
// string that interpolates nothing, a plain scalar variable, or literals
// that constant-fold. Measured on 5.42.0, `$p->{a}`, `$p->{-bar}`,
// `$p->{'a$b'}`, `$p->{1.5}`, `$p->{$k}`, `$p->{$_}`, `$p->{$1}` and
// `$p->{$Foo::x}` all fold; on 5.44.0 so do `$p->{-1}`, `$p->{+$k}`,
// `$p->{${k}}`, `$p->{${^W}}`, `$p->{-x}`, `$p->{1+1}` and `$p->{"a"."b"}`
// (TestArrowHashReadKeyShapesFollowPerl). `$p->{-$k}` does not: negating
// a variable is computed.
//
// Conservative the safe way: a key it does not recognise hedges, which
// scores wider, never WRONG.
func isSimpleKey(k *Node) bool {
	if k.Kind == Unary && len(k.Children) == 1 {
		switch c := k.Children[0]; k.Text {
		case "pos":
			return isSimpleKey(c) // unary plus builds no op
		case "neg":
			// `-bar` is a string to perl and `-1` a number; anything
			// else negated is computed.
			return c.Kind == Call && len(c.Children) == 0 || isLiteral(c)
		}
	}
	if isLiteral(k) {
		return true
	}
	if k.Kind != Term || k.Text == "" {
		return false
	}
	t := k.Text
	switch {
	case t[0] == '$':
		// `$k`, `${k}`, `${ k }` and `${^W}` name one scalar.
		name := t[1:]
		if strings.HasPrefix(name, "{") && strings.HasSuffix(name, "}") {
			name = strings.TrimPrefix(strings.TrimSpace(name[1:len(name)-1]), "^")
		}
		return name != "" && strings.IndexFunc(name, func(r rune) bool {
			return r != ':' && (r >= utf8.RuneSelf || !isWordByte(byte(r)))
		}) < 0
	case t[0] == '-':
		// A filetest's spelling, `-x`, is the string "-x" in a subscript.
		return len(t) > 1 && strings.IndexFunc(t[1:], func(r rune) bool {
			return r >= utf8.RuneSelf || !isWordByte(byte(r))
		}) < 0
	}
	return strings.IndexFunc(t, func(r rune) bool {
		return r != '.' && r != ':' && (r >= utf8.RuneSelf || !isWordByte(byte(r)))
	}) < 0
}

// isLiteral reports whether a key is a constant perl folds at compile
// time: a number, a string that interpolates nothing, or `+ - * . << **`,
// a scalar `x`, `~`, `!`, `length` or `join` over those
// (TestArrowHashReadLiteralFoldsDecide). A bareword is not, since strict
// refuses it; `("a") x 2` is a list repeat, which perl does not fold; and
// `lc` and its kin fold only outside `use locale`.
func isLiteral(k *Node) bool {
	switch k.Kind {
	case Binary:
		if len(k.Children) != 2 {
			return false
		}
		switch k.Text {
		case "x":
			return !k.Children[0].Paren && isLiteral(k.Children[0]) && isNumber(k.Children[1])
		case ".":
			return isLiteral(k.Children[0]) && isLiteral(k.Children[1])
		}
		return isNumber(k)
	case Unary:
		if k.Text == "!" {
			return len(k.Children) == 1 && isLiteral(k.Children[0])
		}
		return isNumber(k)
	case Call:
		// A key's own comma, `$r->{1,2}`, joins with `$;` at run time and
		// keeps its rv2hv; join's commas separate its arguments.
		return k.Resolved && (k.Text == "length" || k.Text == "join") &&
			len(k.Children) == 1 && isLiteralList(k.Children[0])
	case Term:
		t := k.Text
		switch {
		case t == "":
			return false
		case t[0] >= '0' && t[0] <= '9', t[0] == '\'':
			return true
		case t[0] == '"' || strings.HasPrefix(t, "qq"):
			return !strings.ContainsAny(t, "$@")
		case strings.HasPrefix(t, "q") && len(t) > 1 && !isWordByte(t[1]):
			return true
		}
	}
	return false
}

// isNumber reports whether a key is a number literal, or `+ - * << **` or
// `~` over number literals. perl does not fold a numeric op over a string,
// which would warn: measured on 5.44.0, `$r->{"a" << 1}`, `$r->{1+"a"}`
// and `$r->{"0x10" << 1}` each keep their rv2hv.
func isNumber(k *Node) bool {
	switch k.Kind {
	case Binary:
		switch k.Text {
		case "+", "-", "*", "<<", "**":
			return len(k.Children) == 2 && isNumber(k.Children[0]) && isNumber(k.Children[1])
		}
	case Unary:
		return k.Text == "~" && len(k.Children) == 1 && isNumber(k.Children[0])
	case Term:
		return k.Text != "" && k.Text[0] >= '0' && k.Text[0] <= '9'
	}
	return false
}

// keyMayNotFold reports whether a key isSimpleKey accepts may still keep
// its level's rv2hv: a quoted or numeric literal read after an import
// point, which may have installed overload::constant, or a `length` or
// `join` the file may override.
func (d namedHashes) keyMayNotFold(level, key *Node) bool {
	var doubt func(*Node) bool
	doubt = func(n *Node) bool {
		switch {
		case n.Kind == Term && d.afterImport[level] && (isNumber(n) || isLiteral(n)):
			return true
		case n.Kind == Call && d.overridden[n.Text]:
			return true
		}
		return slices.ContainsFunc(n.Children, doubt)
	}
	return doubt(key)
}

// isLiteralList reports whether every item of a comma list is a literal.
func isLiteralList(k *Node) bool {
	if k.Kind == Binary && k.Text == "," && len(k.Children) == 2 {
		return isLiteralList(k.Children[0]) && isLiteralList(k.Children[1])
	}
	return isLiteral(k)
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
