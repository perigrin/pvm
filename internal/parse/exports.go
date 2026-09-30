// ABOUTME: A module's exports and prototypes, read out of its own parsed text.
// ABOUTME: Three tiers — prototype known, name known but shape unknown (XS), name unknown.

package parse

import (
	"slices"
	"strings"
)

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

	// defaults are the names in @EXPORT alone: what a bare `use` imports,
	// and what `:DEFAULT` names in an import list.
	defaults []string

	// protos maps a declared sub's name to its prototype string.
	protos map[string]string

	// opaque marks a module whose export list is not statically readable --
	// a computed list, an import override. Half an answer is worse than
	// none: a name missing from a partial list reads as "not exported" when
	// the truth is "not known".
	opaque bool

	// dynamic marks a module that can put subs into its caller that no list
	// here shows: its own `import`, an `import` inherited from a base other
	// than Exporter, or a glob assigned under a computed name. What it
	// defines is known to perl and not to this parser. See symbolsOpen.
	dynamic bool

	// packages are the packages the module's XS declares beyond its own --
	// Compress::Raw::Bzip2's Bzip2.xs declares Compress::Raw::Bunzip2 -- which
	// exist once it loads.
	packages []string

	// tags are %EXPORT_TAGS, where its lists are literal: tag to names.
	tags map[string][]string

	// builder marks a Test::Builder::Module subclass, whose import list is
	// plan arguments plus `import => [...]`; see builderImportList.
	builder bool

	// loadsXS marks a module that loads C code -- XSLoader::load or a
	// DynaLoader bootstrap. Its subs may be C with prototypes this parser
	// has not read; see importsFrom.
	loadsXS bool

	// globs are the names a literal glob assignment defines, `*run_perl =
	// \&runperl` in t/test.pl: subs as real as a `sub NAME`, with no
	// prototype this parser can read.
	globs []string
}

// readModule pulls the export list and prototypes out of a parsed module.
func readModule(root *Node) moduleFacts {
	facts := moduleFacts{protos: map[string]string{}}
	for _, stmt := range root.Children {
		for _, n := range stmt.Children {
			switch {
			case n.Kind == Declaration && (n.Text == "our" || n.Text == "my" || n.Text == "local"):
				readExportAssignment(n, &facts)
			case n.Kind == Binary && n.Text == "=":
				// `@EXPORT = qw(...)` with no declarator, as Devel::Peek
				// spells it: the same package array.
				readExportAssignment(n, &facts)
			}
		}
	}
	readSubs(root, &facts)
	if _, ok := facts.protos["import"]; ok {
		facts.dynamic = true
	}
	readDynamic(root, &facts)
	return facts
}

// isVariableRef reports whether n is a reference to a scalar, array or hash:
// `\$x`, `\@x`, `\our $TODO`.
func isVariableRef(n *Node) bool {
	if n.Kind != Unary || n.Text != "ref" || len(n.Children) != 1 {
		return false
	}
	target := n.Children[0]
	if target.Kind == Declaration && len(target.Children) == 1 {
		target = target.Children[0]
	}
	return target.Kind == Term && target.Text != "" && strings.ContainsRune("$@%", rune(target.Text[0]))
}

// xsLoaders are the calls that load a module's C code.
var xsLoaders = map[string]bool{
	"XSLoader::load": true, "bootstrap": true,
	"DynaLoader::bootstrap": true, "bootstrap_inherit": true,
}

// noteBase records a base class. Exporter's import is the one the export
// lists describe. Test::Builder::Module's is modelled -- see
// builderImportList. Any other base's import may define subs unseen.
func noteBase(name string, facts *moduleFacts) {
	switch name {
	case "Exporter", "-norequire":
	case "Test::Builder::Module":
		facts.builder = true
	default:
		facts.dynamic = true
	}
}

// readDynamic finds what can define subs out of this parser's sight --
// see moduleFacts.dynamic -- and the names literal glob assignments define.
func readDynamic(n *Node, facts *moduleFacts) {
	if xsLoaders[n.Text] {
		facts.loadsXS = true
	}
	switch {
	case n.Kind == Use && len(n.Children) > 0 &&
		(n.Children[0].Text == "parent" || n.Children[0].Text == "base"):
		for _, c := range n.Children[1:] {
			names, ok := literalNameList(c)
			if !ok {
				facts.dynamic = true
			}
			for _, name := range names {
				noteBase(name, facts)
			}
		}
	case (n.Kind == Declaration || n.Kind == Binary && n.Text == "=") &&
		len(n.Children) == 2 && n.Children[0].Text == "@ISA":
		names, ok := literalNameList(n.Children[1])
		if !ok {
			facts.dynamic = true
		}
		for _, name := range names {
			noteBase(name, facts)
		}
	case n.Kind == Binary && n.Text == "=" && len(n.Children) == 2 &&
		n.Children[0].Kind == Term && strings.HasPrefix(n.Children[0].Text, "*"):
		if glob := n.Children[0]; glob.Text == "*" {
			// A variable's reference defines a variable, not a sub: Test/
			// More.pm:210 exports $TODO this way. Anything else may be code.
			if !isVariableRef(n.Children[1]) {
				facts.dynamic = true
			}
		} else {
			facts.globs = append(facts.globs, strings.TrimPrefix(glob.Text, "*"))
		}
	}
	for _, c := range n.Children {
		readDynamic(c, facts)
	}
}

// readSubs records every NAMED sub declaration, however deeply nested: a
// named sub is package-global wherever it is declared. t/test.pl declares
// `sub watchdog ($;$)` inside a `{ # Closure ... }` block, measured to be
// callable after it on 5.42.0. A lexical sub -- `my sub`, `state sub` -- is
// not, so the walk does not enter those declarations.
func readSubs(n *Node, facts *moduleFacts) {
	if n.Kind == Declaration {
		switch n.Text {
		case "my", "state":
			return
		case "sub":
			if name, proto := declaredSub(n); name != "" {
				facts.protos[name] = proto
			}
		}
	}
	for _, c := range n.Children {
		readSubs(c, facts)
	}
}

// readExportAssignment reads `our @EXPORT = qw(...)` and its EXPORT_OK
// spelling, and marks the module opaque if the list is computed.
func readExportAssignment(n *Node, facts *moduleFacts) {
	if len(n.Children) < 2 {
		return
	}
	array := n.Children[0].Text
	switch array {
	case "@EXPORT", "@EXPORT_OK":
	case "%EXPORT_TAGS":
		readExportTags(n.Children[1], facts)
		return
	default:
		return
	}

	names, ok := literalNameList(n.Children[1])
	if !ok {
		// A list built from the module's own tags is as literal as the
		// tags: `@EXPORT_OK = ( @{ $EXPORT_TAGS{'all'} } )` in perl 5.42.0's
		// Hash/Util/FieldHash.pm.
		names, ok = tagBuiltList(n.Children[1], facts.tags)
	}
	if !ok {
		// Computed, so nothing about this module is trustworthy.
		facts.opaque = true
		return
	}
	facts.exports = append(facts.exports, names...)
	if array == "@EXPORT" {
		facts.defaults = append(facts.defaults, names...)
	}
}

// readExportTags reads `%EXPORT_TAGS = ( tag => [ names ], ... )`, keeping
// each tag whose list is literal.
func readExportTags(n *Node, facts *moduleFacts) {
	items := listItems(n)
	for i := 0; i+1 < len(items); i += 2 {
		key, ok := constantName(items[i])
		if !ok || items[i+1].Kind != AnonArray {
			continue
		}
		var names []string
		literal := true
		for _, c := range items[i+1].Children {
			got, ok := literalNameList(c)
			if !ok {
				literal = false
				break
			}
			names = append(names, got...)
		}
		if literal {
			if facts.tags == nil {
				facts.tags = map[string][]string{}
			}
			facts.tags[key] = names
		}
	}
}

// tagBuiltList reads an export list whose elements are literal names or
// `@{ $EXPORT_TAGS{tag} }` of a tag already read.
func tagBuiltList(n *Node, tags map[string][]string) ([]string, bool) {
	var out []string
	for _, item := range listItems(n) {
		if got, ok := literalNameList(item); ok {
			out = append(out, got...)
			continue
		}
		if item.Kind != Unary || item.Text != "@" || len(item.Children) != 1 {
			return nil, false
		}
		idx := item.Children[0]
		if idx.Kind != Index || len(idx.Children) != 2 || idx.Children[0].Text != "$EXPORT_TAGS" {
			return nil, false
		}
		key, ok := constantName(idx.Children[1])
		names, known := tags[key]
		if !ok || !known {
			return nil, false
		}
		out = append(out, names...)
	}
	return out, true
}

// unknownTag reports whether an import list names a `:tag` the module's
// tags do not hold -- what it imports is then not known.
func unknownTag(facts moduleFacts, list []string) bool {
	for _, name := range list {
		if tag, ok := strings.CutPrefix(name, ":"); ok && tag != "DEFAULT" {
			if _, known := facts.tags[tag]; !known {
				return true
			}
		}
	}
	return false
}

// literalNameList reads a `qw(...)` or a list of quoted strings, and reports
// whether the node was one at all.
func literalNameList(n *Node) ([]string, bool) {
	text := strings.TrimSpace(n.Text)

	// qw(a b c), with any of perl's delimiters, and space may stand before
	// the delimiter: `qw "catfile"` in t/op/coreamp.t:842.
	if body, ok := strings.CutPrefix(text, "qw"); ok {
		// A word byte after `qw` is an identifier (`qwerty`), not a quote.
		if body = strings.TrimLeft(body, " \t\n"); len(body) >= 2 && !isWordByteAt(body, 0) {
			return strings.Fields(body[1 : len(body)-1]), true
		}
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

	// A list of them: every child must itself be literal. With parentheses
	// the list is a List; without, `use feature 'a', 'b'` is a comma Binary.
	if (n.Kind == List || n.Kind == Binary && n.Text == ",") && len(n.Children) > 0 {
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

	// A bare `use` imports @EXPORT only -- measured on 5.42.0, an
	// @EXPORT_OK name is not defined after it. In a list, `:DEFAULT` names
	// @EXPORT and another `:tag` names its %EXPORT_TAGS list; a tag not read
	// (see unknownTag), and a `!` or `/pattern/` entry -- a negation or a
	// match -- contribute nothing rather than a false name.
	wanted := facts.defaults
	if listGiven {
		wanted = nil
		for _, name := range list {
			switch {
			case name == ":DEFAULT":
				wanted = append(wanted, facts.defaults...)
			case strings.HasPrefix(name, ":"):
				wanted = append(wanted, facts.tags[strings.TrimPrefix(name, ":")]...)
			case strings.HasPrefix(name, "!") || strings.HasPrefix(name, "/"):
			default:
				wanted = append(wanted, strings.TrimPrefix(name, "&"))
			}
		}
	}

	out := make([]Import, 0, len(wanted))
	for _, name := range wanted {
		proto, known := facts.protos[name]
		// Exported but never declared: the module builds it out of sight,
		// as File::Spec::Functions assigns `*{$meth}` in a loop. A sub
		// built in Perl that way is taken to have no prototype (perigrin's
		// decision, 2026-09-30) -- catdir has none, measured on 5.42.0. A
		// module that loads C keeps it unknown: its sub may be XS with a
		// prototype, as List::Util's `first` is `&@`.
		//
		// Only a name the module exports: a list also carries arguments
		// that are not subs at all -- `use feature 'defer'`.
		if !known && !facts.loadsXS && slices.Contains(facts.exports, name) {
			proto, known = "", true
		}
		out = append(out, Import{
			Name:           name,
			Prototype:      proto,
			PrototypeKnown: known,
		})
	}
	return out
}

// isWordByteAt reports whether s[i] can continue an identifier.
func isWordByteAt(s string, i int) bool {
	c := s[i]
	return c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
