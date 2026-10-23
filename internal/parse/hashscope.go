// ABOUTME: Resolves each named hash use to the declaration in effect there -- my, our, state, local or none --
// ABOUTME: and says whether the state feature is on there, walking the tree's lexical scopes as perl does.

package parse

import (
	"slices"
	"strings"
)

// hashUse is what hashScopes says about one named hash in the tree.
type hashUse struct {
	// Decl is the declarator in effect: "my", "our", "state" or "local",
	// or "" where no declaration reaches and the hash is the package's.
	// On a declaring variable it is that variable's own declarator.
	Decl string
	// Declares marks the variable a declaration names, `%h` in `my %h`.
	Declares bool
	// StateOn reports whether the state feature is on at this node. A
	// `state %h` with it off declares nothing: perl reads `state % h`.
	StateOn bool
}

// hashScope is one lexical scope: a block, or an `if`/`while`/`for` whose
// condition's declarations reach its blocks and stop at its end.
type hashScope struct {
	decls map[string]string
	// pending are the declarations of the statement being read. perl makes
	// a `my` visible from the NEXT statement, so the statement's own
	// initialiser still sees the outer hash. Measured on 5.44.0:
	//
	//	my %h = keys %h            rv2hv, padhv[%h:1,2]
	//	my %h = do { keys %h }     rv2hv, padhv[%h:3,4]
	pending map[string]string
	stateOn bool
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
	s := &hashScope{decls: map[string]string{}, pending: map[string]string{}}
	if len(r.scopes) > 0 {
		s.stateOn = r.top().stateOn
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
	use := hashUse{StateOn: r.top().stateOn}
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if decl, ok := r.scopes[i].decls[name]; ok {
			use.Decl = decl
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
		// A condition's `my` reaches the blocks and ends with the
		// construct: perly.y's `remember mexpr` calls intro_my after the
		// condition. Measured on 5.44.0:
		//
		//	if ((my %g) = ()) { keys %g } keys %g     padhv, padhv, rv2hv
		//
		// A statement modifier has no block and opens no scope:
		// `keys %g if my %g` declares %g for the statements after it.
		if !slices.ContainsFunc(n.Children, func(c *Node) bool { return c.Kind == Block }) {
			break
		}
		r.push()
		for _, c := range n.Children {
			if c.Kind == Block {
				r.introduce()
			}
			r.walk(c)
		}
		r.pop()
		return

	case Use:
		r.noteState(n)
		return

	case Declaration:
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
	case "my", "our", "state", "local":
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
	r.uses[t] = hashUse{Decl: decl, Declares: true, StateOn: s.stateOn}
	if decl == "state" && !s.stateOn {
		return
	}
	s.pending[t.Text[1:]] = decl
}

// noteState applies a `use` or `no` to the top scope's state feature, which
// ends with that scope as every feature does. Measured on 5.44.0:
//
//	{ use feature "state"; } state %h; keys %h               rv2hv
//	use feature "state"; no feature; state %h; keys %h       rv2hv
//	use feature "state"; use v5.8; state %h; keys %h         rv2hv
//	use feature ":5.10"; ...; no feature ":5.10"; state %g   rv2hv
//
// `use VERSION` sets the version's bundle, which holds state from 5.10 on,
// and a bare `no feature` returns to the default bundle, which does not.
func (r *hashResolver) noteState(n *Node) {
	if len(n.Children) == 0 {
		return
	}
	s := r.top()
	module := n.Children[0].Text
	switch {
	case module == "feature" && len(n.Children) == 1:
		if n.Text == "no" {
			s.stateOn = false
		}
	case module == "feature":
		names, ok := literalNameList(n.Children[1])
		if !ok {
			return
		}
		for _, name := range names {
			if bundle, ok := strings.CutPrefix(name, ":"); ok {
				if bundle != "all" && !bundleHasState("v"+bundle) {
					continue
				}
			} else if name != "state" {
				continue
			}
			s.stateOn = n.Text == "use"
		}
	case n.Text == "use" && len(n.Children) == 1:
		if _, _, ok := perlVersion(module); ok {
			s.stateOn = bundleHasState(module)
		}
	}
}

// bundleHasState reports whether the feature bundle of a perl version
// holds state: 5.10 and every version after it.
func bundleHasState(version string) bool {
	major, minor, ok := perlVersion(version)
	return ok && (major > 5 || major == 5 && minor >= 10)
}
