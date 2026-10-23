// ABOUTME: Resolves each named hash use to the declaration in effect there -- my, our, state, local or none --
// ABOUTME: and says whether the state feature is on there, walking the tree's lexical scopes as perl does.

package parse

import (
	"slices"
	"strconv"
	"strings"
)

// hashUse is what hashScopes says about one named hash in the tree.
type hashUse struct {
	// Decl is the declarator in effect: "my", "our", "state", "local" or
	// "field", or "" where no declaration reaches and the hash is the
	// package's. A signature's hash parameter is "my"; a field, like a my,
	// is padhv.
	// On a declaring variable it is that variable's own declarator.
	Decl string
	// Declares marks the variable a declaration names, `%h` in `my %h`.
	Declares bool
	// State is the state feature at this node. A
	// `state %h` with it off declares nothing: perl reads `state % h`.
	State featureState
	// Unsure marks a use whose declaration is a `state` made where the
	// feature was unknown: perl may have read it as `state % h`, and then
	// the hash is whatever lies outside it.
	Unsure bool
}

// featureState is whether a feature is on, off, or unknown -- turned on or
// off, perhaps, by a module or BEGIN block this resolver cannot read.
type featureState int

const (
	stateOff featureState = iota
	stateOn
	stateUnknown
)

// hashScope is one lexical scope: a block, or an `if`/`while`/`for` or a
// sub whose head's declarations reach its blocks and stop at its end.
// Each declaration is held as the hashUse a use of it gets: Decl and Unsure.
type hashScope struct {
	decls map[string]hashUse
	// pending are the declarations of the statement being read. perl makes
	// a `my` visible from the NEXT statement, so the statement's own
	// initialiser still sees the outer hash. Measured on 5.44.0:
	//
	//	my %h = keys %h            rv2hv, padhv[%h:1,2]
	//	my %h = do { keys %h }     rv2hv, padhv[%h:3,4]
	pending map[string]hashUse
	state   featureState
}

// hashScopes resolves every named hash in root: each `%h`, and each `$h`
// or `@h` a brace subscript reads %h through. A qualified or punctuation
// name never matches a lexical declaration, so it resolves to "".
func hashScopes(root *Node) map[*Node]hashUse {
	r := &hashResolver{uses: map[*Node]hashUse{}}
	r.push()
	r.walk(root)
	return r.uses
}

type hashResolver struct {
	scopes []*hashScope
	uses   map[*Node]hashUse
}

func (r *hashResolver) top() *hashScope { return r.scopes[len(r.scopes)-1] }

func (r *hashResolver) push() {
	s := &hashScope{decls: map[string]hashUse{}, pending: map[string]hashUse{}}
	if len(r.scopes) > 0 {
		s.state = r.top().state
	}
	r.scopes = append(r.scopes, s)
}

func (r *hashResolver) pop() { r.scopes = r.scopes[:len(r.scopes)-1] }

// introduce makes the top scope's pending declarations visible.
func (r *hashResolver) introduce() {
	s := r.top()
	for name, decl := range s.pending {
		s.decls[name] = decl
	}
	clear(s.pending)
}

// resolve records a use of the hash name, from its innermost declaration.
func (r *hashResolver) resolve(n *Node, name string) {
	use := hashUse{State: r.top().state}
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if decl, ok := r.scopes[i].decls[name]; ok {
			use.Decl, use.Unsure = decl.Decl, decl.Unsure
			break
		}
	}
	r.uses[n] = use
}

func (r *hashResolver) walk(n *Node) {
	switch n.Kind {
	case Block:
		r.push()
		r.walkChildren(n)
		r.pop()
		return

	case Statement:
		r.walkChildren(n)
		r.introduce()
		return

	case Conditional, Loop:
		// A head's `my` reaches the blocks and ends with the construct:
		// perly.y's `remember mexpr` calls intro_my after the condition,
		// and a C-style for's init reaches its condition. Measured on
		// 5.44.0:
		//
		//	if ((my %g) = ()) { keys %g } keys %g     padhv, padhv, rv2hv
		//	for (my %h = (); keys %h; ) { keys %h } keys %h
		//	                                padhv padhv padhv rv2hv
		//
		// A statement modifier has no block and opens no scope:
		// `keys %g if my %g` declares %g for the statements after it.
		if slices.ContainsFunc(n.Children, func(c *Node) bool { return c.Kind == Block }) {
			r.walkHead(n, r.walk)
			return
		}

	case Use:
		r.noteState(n)
		return

	case Phaser:
		// A BEGIN block runs as it is compiled and may turn a feature on
		// in the scope around it: `BEGIN { feature->import("state") }`.
		if n.Text == "BEGIN" {
			r.top().state = stateUnknown
		}

	case Declaration:
		switch keywordName(n.Text) {
		case "sub", "method":
			// A signature's hash parameter is a `my` of the body. Measured
			// on 5.44.0, -MO=Concise,-exec,-main,f:
			//
			//	use feature "signatures"; sub f ($x, %o) { keys %o } keys %o
			//	padhv[%o:4,5] in f, rv2hv in main
			r.walkHead(n, r.declareSignature)
			return
		}
		if r.declare(n) {
			return
		}

	case Index:
		// `$h{k}`, `@h{...}` and `%h{...}` read %h. A bracket subscript
		// reads an array, `%h[0]` included.
		if len(n.Children) > 0 && !n.Arrow {
			base := n.Children[0]
			if base.Kind == Term && (isNamedVariable(base.Text) || isNamedHash(base.Text)) {
				if n.Text == "{" {
					r.resolve(base, base.Text[1:])
				}
				for _, c := range n.Children[1:] {
					r.walk(c)
				}
				return
			}
		}

	case Term:
		if isNamedHash(n.Text) {
			r.resolve(n, n.Text[1:])
		}
		return
	}
	r.walkChildren(n)
}

// walkHead walks a construct with blocks in a scope of its own: each
// non-block child through head, its declarations visible from the next
// child on.
func (r *hashResolver) walkHead(n *Node, head func(*Node)) {
	r.push()
	for _, c := range n.Children {
		if c.Kind == Block {
			r.walk(c)
			continue
		}
		head(c)
		r.introduce()
	}
	r.pop()
}

// declareSignature reads a sub's head: a signature's hash parameters are
// `my` declarations, and anything else is walked as usual.
func (r *hashResolver) declareSignature(c *Node) {
	targets := []*Node{c}
	if c.Kind == List {
		targets = c.Children
	}
	for _, t := range targets {
		if t.Kind == Term && isNamedHash(t.Text) {
			r.declareHash(t, "my")
		} else {
			r.walk(t)
		}
	}
}

func (r *hashResolver) walkChildren(n *Node) {
	for _, c := range n.Children {
		r.walk(c)
	}
}

// declare reads a `my`, `our`, `state` or `local` and reports whether it
// did. Its named hashes become pending in the top scope, and its
// initialiser is walked as ordinary uses. `local $h{a}` declares nothing:
// its subscript is a use of %h.
func (r *hashResolver) declare(n *Node) bool {
	decl := keywordName(n.Text)
	switch decl {
	case "my", "our", "state", "local", "field":
	default:
		return false
	}
	headDone := false
	for _, c := range n.Children {
		if c.Kind == TypeName || c.Kind == Attribute {
			continue
		}
		if headDone {
			r.walk(c)
			continue
		}
		headDone = true
		head, init := c, (*Node)(nil)
		if c.Kind == Binary && c.Text == "=" && len(c.Children) == 2 {
			head, init = c.Children[0], c.Children[1]
		}
		targets := []*Node{head}
		if head.Kind == List {
			targets = head.Children
		}
		for _, t := range targets {
			// declared_refs: `my \%h` declares %h. Measured on 5.44.0,
			// `my \%h = {}; keys %h` is padhv.
			if t.Kind == Unary && t.Text == "ref" && len(t.Children) == 1 {
				t = t.Children[0]
			}
			if t.Kind == Term && isNamedHash(t.Text) {
				r.declareHash(t, decl)
			} else {
				r.walk(t)
			}
		}
		if init != nil {
			r.walk(init)
		}
	}
	return true
}

// declareHash records the declaring variable t and makes it pending.
func (r *hashResolver) declareHash(t *Node, decl string) {
	s := r.top()
	r.uses[t] = hashUse{Decl: decl, Declares: true, State: s.state}
	if decl == "state" && s.state == stateOff {
		return
	}
	s.pending[t.Text[1:]] = hashUse{Decl: decl, Unsure: decl == "state" && s.state == stateUnknown}
}

// noteState applies a `use` or `no` to the top scope's state feature, which
// ends with that scope as every feature does. Measured on 5.44.0:
//
//	{ use feature "state"; } state %h; keys %h               rv2hv
//	use feature "state"; no feature; state %h; keys %h       rv2hv
//	use feature "state"; use v5.8; state %h; keys %h         rv2hv
//	use feature ":5.10"; ...; no feature ":5.10"; state %g   rv2hv
//	use if 1, feature => "state"; state %h; keys %h          padhv
//	use warnings; state %h; keys %h                          rv2hv
//
// `use VERSION` sets the version's bundle, and a bare `no feature` returns
// to the default bundle, which has no state. Any other module's import may
// turn state on or off -- Modern::Perl and Mojo::Base do -- so after one the
// feature is unknown, unless it is a core pragma known to leave features
// alone. A `require` runs when the program does, too late to change one.
func (r *hashResolver) noteState(n *Node) {
	if n.Text == "require" || len(n.Children) == 0 {
		return
	}
	s := r.top()
	module := n.Children[0].Text
	switch {
	case module == "feature" && len(n.Children) == 1:
		if n.Text == "no" {
			s.state = stateOff
		}
	case module == "feature":
		names, ok := literalNameList(n.Children[1])
		if !ok {
			s.state = stateUnknown
			return
		}
		for _, name := range names {
			if bundle, ok := strings.CutPrefix(name, ":"); ok {
				if has, _ := versionHasState("v" + bundle); bundle != "all" && !has {
					continue
				}
			} else if name != "state" {
				continue
			}
			s.state = stateOff
			if n.Text == "use" {
				s.state = stateOn
			}
		}
	case len(n.Children) == 1 && isVersionText(module):
		if has, ok := versionHasState(module); ok && n.Text == "use" {
			s.state = stateOff
			if has {
				s.state = stateOn
			}
		}
	case !quietPragmas[module]:
		s.state = stateUnknown
	}
}

// isVersionText reports whether a `use` names a perl version: `v5.10`,
// `5.010`, `5.9.5`.
func isVersionText(text string) bool {
	_, ok := versionHasState(text)
	return ok
}

// versionHasState reports whether the feature bundle of a perl version holds
// state, which arrived in development release 5.9.5. Measured on 5.44.0,
// each before `state %h; keys %h`:
//
//	use 5.9.5    padhv     use 5.009005   padhv     use 5.009_005   padhv
//	use 5.9.4    rv2hv     use 5.009004   rv2hv     use 5.010_001   padhv
//
// A decimal version's fraction is read in groups of three digits, and an
// underscore is not a digit.
func versionHasState(text string) (has, ok bool) {
	text, dotted := strings.CutPrefix(text, "v")
	parts := strings.Split(strings.ReplaceAll(text, "_", ""), ".")
	if len(parts) < 2 || len(parts) > 3 {
		return false, false
	}
	if len(parts) == 2 && !dotted {
		frac := (parts[1] + "000000")[:6]
		parts = []string{parts[0], frac[:3], frac[3:]}
	}
	var v [3]int
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil {
			return false, false
		}
		v[i] = n
	}
	return v[0] > 5 || v[0] == 5 && (v[1] > 9 || v[1] == 9 && v[2] >= 5), true
}
