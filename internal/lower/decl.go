// ABOUTME: Lowering of declarations (§4.14.2): the declarator to Decl.Kind, its variables, initialiser and attributes.
// ABOUTME: The list form is read off the CST's shape -- one `binary "="` child -- and the Paren bit, never re-parsed.

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
	if kind >= DeclSub {
		// ponytail: a named declaration carries its kind only; where its
		// name, signature and body go is the vocabulary's to say.
		return d
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
	for i, c := range kids {
		if i == 0 {
			d.Vars = []Node{l.expr(c)}
			continue
		}
		d.Init = l.expr(c)
	}
	return d
}
