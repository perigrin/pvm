// ABOUTME: A module's exports and prototypes, read out of its own parsed text.
// ABOUTME: Three tiers — prototype known, name known but shape unknown (XS), name unknown.

package parse

import "strings"

// Import is one name a `use` brought into scope.
//
// The prototype is a STRING, not a shape enum, because the prototype
// DETERMINES the shape and a stored shape would be a second copy of one fact:
//
//	sub np  { }        no prototype   np(1,2,3) swallows the list
//	sub one ($) { }    ($)            `one 1, 2` parses as one(1), 2
//	sub nil () { }     ()             `nil + 1`  parses as nil() + 1
//
// Deriving the shape from it is 01a0b84d-ec59.
type Import struct {
	// Name as it appears at a call site.
	Name string

	// Prototype as written, parens included: "($)", "(;$)", "(&@)". Empty
	// when the sub was declared without one.
	Prototype string

	// PrototypeKnown separates "declared with no prototype" from "shape not
	// known at all", which are different answers:
	//
	//	sub subtest { }     known, and the answer is none
	//	XS-backed `first`   NOT known -- the body is in a .so
	//
	// False means refuse rather than guess. Guessing "list operator" for
	// `first` makes the block form a syntax error, while
	// `first { $_ > 1 } @a` is valid perl.
	PrototypeKnown bool

	// Local marks a name the using file declares itself. A local sub shadows
	// an import, and its own prototype is the one that applies.
	Local bool
}

// Imports names everything a parsed file's `use` statements brought into
// scope, plus the subs it declares itself.
//
// Empty for a tree from Parse, which resolves nothing.
func Imports(root *Node) map[string]Import {
	if root == nil || root.imports == nil {
		return map[string]Import{}
	}
	return root.imports
}

// moduleFacts is what one module's source says about itself.
type moduleFacts struct {
	// exports are the names in @EXPORT and @EXPORT_OK.
	exports []string

	// protos maps a declared sub's name to its prototype string.
	protos map[string]string

	// opaque marks a module whose export list is not statically readable --
	// a computed list, an import override. Half an answer is worse than
	// none: a name missing from a partial list reads as "not exported" when
	// the truth is "not known".
	opaque bool
}

// readModule pulls the export list and prototypes out of a parsed module.
func readModule(root *Node) moduleFacts {
	facts := moduleFacts{protos: map[string]string{}}
	for _, stmt := range root.Children {
		for _, n := range stmt.Children {
			if n.Kind != Declaration {
				continue
			}
			switch n.Text {
			case "our", "my", "local":
				readExportAssignment(n, &facts)
			case "sub":
				name, proto := declaredSub(n)
				if name != "" {
					facts.protos[name] = proto
				}
			}
		}
	}
	return facts
}

// readExportAssignment reads `our @EXPORT = qw(...)` and its EXPORT_OK
// spelling, and marks the module opaque if the list is computed.
func readExportAssignment(n *Node, facts *moduleFacts) {
	if len(n.Children) < 2 {
		return
	}
	switch n.Children[0].Text {
	case "@EXPORT", "@EXPORT_OK":
	default:
		return
	}

	names, ok := literalNameList(n.Children[1])
	if !ok {
		// Computed, so nothing about this module is trustworthy.
		facts.opaque = true
		return
	}
	facts.exports = append(facts.exports, names...)
}

// literalNameList reads a `qw(...)` or a list of quoted strings, and reports
// whether the node was one at all.
func literalNameList(n *Node) ([]string, bool) {
	text := strings.TrimSpace(n.Text)

	// qw(a b c), with any of perl's delimiters.
	if strings.HasPrefix(text, "qw") && len(text) > 3 {
		return strings.Fields(text[3 : len(text)-1]), true
	}

	// A single quoted string.
	if len(text) >= 2 && (text[0] == '\'' || text[0] == '"') &&
		text[len(text)-1] == text[0] {
		return []string{text[1 : len(text)-1]}, true
	}

	// q(...) and qq(...), with any of perl's delimiters. 61 perl.git t/
	// files spell `require q(./test.pl)`. A `qq` body that would interpolate
	// is not a literal -- it names something only at runtime -- which is
	// the rule resolveRequiredFile applies to `"..."`.
	if body, interp, ok := genericQuoteBody(text); ok {
		if interp && strings.ContainsAny(body, "$@") {
			return nil, false
		}
		return []string{body}, true
	}

	// A list of them: every child must itself be literal.
	if n.Kind == List && len(n.Children) > 0 {
		var out []string
		for _, c := range n.Children {
			names, ok := literalNameList(c)
			if !ok {
				return nil, false
			}
			out = append(out, names...)
		}
		return out, true
	}

	return nil, false
}

// genericQuoteBody returns the body of a `q` or `qq` quote as written, and
// whether it is the interpolating `qq`. Whitespace may sit between the
// operator and its delimiter, as `q (./test.pl)` does; a bracketing delimiter
// closes with its partner and any other closes with itself.
func genericQuoteBody(text string) (body string, interp, ok bool) {
	rest, found := strings.CutPrefix(text, "qq")
	interp = found
	if !found {
		if rest, found = strings.CutPrefix(text, "q"); !found {
			return "", false, false
		}
	}
	rest = strings.TrimLeft(rest, " \t\n")
	if len(rest) < 2 {
		return "", false, false
	}
	open := rest[0]
	shut := open
	switch open {
	case '(':
		shut = ')'
	case '[':
		shut = ']'
	case '{':
		shut = '}'
	case '<':
		shut = '>'
	}
	// A word byte after `q` is an identifier (`qux`), not a quote.
	if open == '_' || open >= '0' && open <= '9' ||
		open >= 'a' && open <= 'z' || open >= 'A' && open <= 'Z' ||
		rest[len(rest)-1] != shut {
		return "", false, false
	}
	return rest[1 : len(rest)-1], interp, true
}

// declaredSub returns a sub declaration's name and prototype. The prototype is
// empty when the declaration carries none, which is itself an answer.
func declaredSub(n *Node) (name, proto string) {
	for _, c := range n.Children {
		switch c.Kind {
		case Term:
			if name == "" {
				name = c.Text
			}
		case PrototypeNode:
			proto = c.Text
		}
	}
	return name, proto
}

// importsFrom applies one `use` to the facts its module reported.
//
// list is the import list as written, and listGiven separates `use M ()` --
// load but import nothing -- from `use M`, which imports @EXPORT.
func importsFrom(facts moduleFacts, list []string, listGiven bool) []Import {
	if facts.opaque {
		return nil
	}

	wanted := facts.exports
	if listGiven {
		wanted = list
	}

	out := make([]Import, 0, len(wanted))
	for _, name := range wanted {
		proto, known := facts.protos[name]
		out = append(out, Import{
			Name:           name,
			Prototype:      proto,
			PrototypeKnown: known,
		})
	}
	return out
}
