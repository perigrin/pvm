// ABOUTME: Lowering of declarations (§4.14.2): the declarator to Decl.Kind, its variables, initialiser and attributes.
// ABOUTME: The list form is read off the CST's shape -- one assignment child -- and the Paren bit, never re-parsed.

package lower

import "tamarou.com/pvm/internal/parse"

// declKinds maps a declaration's Text to its kind.
var declKinds = map[string]DeclKind{
	"my":    DeclMy,
	"our":   DeclOur,
	"state": DeclState,
	"local": DeclLocal,
	"field": DeclField,

	"sub":     DeclSub,
	"package": DeclPackage,
	"class":   DeclClass,
	"method":  DeclMethod,
}

// decl lowers a declaration, and declines a declarator it has no kind for.
func (l *lowerer) decl(n *parse.Node) Node {
	kind, ok := declKinds[n.Text]
	if !ok {
		return nil
	}
	d := &Decl{Kind: kind}
	if kind == DeclSub || kind == DeclMethod {
		return l.sub(n, kind)
	}
	kids := n.Children
	// `my ($x) = f()` is one child, the `=`; `my $x = f()` is two, the
	// variable then the initialiser (§4.14.2).
	if len(kids) == 1 && kids[0].Kind == parse.Binary && kids[0].Text == "=" && len(kids[0].Children) == 2 {
		lhs := kids[0].Children[0]
		d.ListAssign = true
		if lhs.Kind == parse.Unary && lhs.Text == "ref" && len(lhs.Children) == 1 {
			// perl: my \$x = f() calls f in scalar context, and
			// my \($x) = f() or my (\$x) = f() in list; the paren decides.
			d.RefAlias = true
			d.ListAssign = lhs.Paren || lhs.Children[0].Paren
			lhs = lhs.Children[0]
		}
		d.Vars = []Node{l.expr(lhs)}
		d.Init = l.expr(kids[0].Children[1])
		return d
	}
	// `my ($x) ||= 1` is a compound assignment to the declaration, and
	// perl reads it as `(my $x ||= 1)`: scalar, whatever the parens say.
	if len(kids) == 1 && kids[0].Kind == parse.Binary && assignOps[kids[0].Text] && len(kids[0].Children) == 2 {
		d.Vars = []Node{l.expr(kids[0].Children[0])}
		return &Assign{Op: kids[0].Text, L: l.derive(n, RoleDeclTarget, d), R: l.expr(kids[0].Children[1])}
	}
	// The head is the variable and its attributes; what follows is the
	// initialiser. An attribute keeps the parser's Text, colon included.
	listTarget := false
	for _, c := range kids {
		switch {
		case c.Kind == parse.Attribute:
			d.Attrs = append(d.Attrs, c.Text)
		case c.Kind == parse.TypeName:
			// ponytail: `my Foo $f` drops its class; no Decl field holds it.
		case d.Vars == nil:
			d.Vars = []Node{l.expr(c)}
			listTarget = c.Kind == parse.List
		default:
			d.Init = l.expr(c)
		}
	}
	if n.AssignOp != "" && d.Init != nil {
		// `my $x -= 1` is `(my $x -= 1)`, as `my ($x) ||= 1` is above.
		a := &Assign{Op: n.AssignOp, R: d.Init}
		d.Init = nil
		a.L = l.derive(n, RoleDeclTarget, d)
		return a
	}
	// `my ($x, $y) :shared = f()` assigns a list. One parenthesised scalar
	// with attributes does not: perl reads `my ($x) :shared` as `my $x
	// :shared`, which is a Var here, not a List.
	if d.Init != nil {
		d.ListAssign = listTarget
	}
	return d
}

// sub lowers a sub or method declaration. A named one is perl's
// compile-time `*f = sub {...}`: a Decl whose one Var is &f and whose Init
// is the AnonSub, nil for `sub f;`, which declares without defining. One
// without a name is the AnonSub itself.
func (l *lowerer) sub(n *parse.Node, kind DeclKind) Node {
	kids := n.Children
	if subName(n) == nil {
		return l.anonSub(kids)
	}
	d := &Decl{Kind: kind, Vars: []Node{l.stamp(kids[0], &Var{Sigil: "&", Name: kids[0].Text})}}
	if last := kids[len(kids)-1]; last.Kind == parse.Block || last.Kind == parse.Unknown {
		// The AnonSub is the same syntax site as its Decl (§4.14.1 rule
		// 6), so its id is the declaration's under RoleSubValue.
		d.Init = l.derive(n, RoleSubValue, l.anonSub(kids[1:]))
	}
	return d
}

// anonSub lowers what follows a sub's name: attributes, a signature, and
// the body. A `prototype` child is not a signature and has no field.
func (l *lowerer) anonSub(kids []*parse.Node) *AnonSub {
	s := &AnonSub{}
	for _, c := range kids {
		switch {
		case c.Kind == parse.Attribute:
			s.Attrs = append(s.Attrs, c.Text)
		case c.Kind == parse.PrototypeNode:
		case c.Kind == parse.List && c.Paren:
			for _, p := range c.Children {
				s.Sig = append(s.Sig, l.expr(p))
			}
		case c.Kind == parse.Term && c.Paren:
			s.Sig = []Node{l.expr(c)}
		default:
			// The body is a Block; an Unknown in its place is a Hole.
			s.Body = l.expr(c)
		}
	}
	return s
}
