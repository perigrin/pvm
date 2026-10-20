// ABOUTME: Module declaration files: what a module defines, stated where its source cannot show it.
// ABOUTME: Embedded, and read in place of the module's own source.
package parse

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"
	"sync"

	"tamarou.com/pvm/internal/types"
)

// declarations holds one file per declared module, `Moose::Role` at
// `declarations/Moose/Role.tpm`. TypeScript's `@types/Foo` is the model: an
// interface ASSERTED over code this parser never reads, for a module whose
// own `import` builds what it defines -- Moose::Exporter, a monkey_patch
// loop -- so that no reading of its source can recover it.
//
// A declaration is written in the module syntax the resolver already reads:
// `sub NAME :prototype(PROTO);` for each sub and its prototype. So it is
// read by the same parse as a module. What a `use` imports is marked on
// each name with `:export` or `:export_ok` (readExportMarks), never in an
// export list (refuseInDeclaration), and a typed `our` states a
// variable's type (declareVariable).
//
//go:embed declarations
var embedded embed.FS

// declarations is where declaration files are read from: the embedded
// ones, or a test's own.
var declarations fs.FS = embedded

// OverlayDeclarations reads declaration files from dir, a tree holding a
// `declarations` directory, before the ones read now, until the restore
// it returns is called. A test in another package states its fixture
// modules this way; nothing else calls it, as a user's own declarations
// are RFC 0001's open question 6.
func OverlayDeclarations(dir fs.FS) (restore func()) {
	saved := declarations
	declarations = layeredFS{top: dir, base: saved}
	return func() { declarations = saved }
}

// layeredFS reads a file from top and falls back to base: the fixture's
// declarations sit beside the shipped ones, so CORE.tpm -- which parsing
// reads for every builtin call -- is still there when a test swaps the
// fixture in first.
type layeredFS struct{ top, base fs.FS }

func (l layeredFS) Open(name string) (fs.File, error) {
	if f, err := l.top.Open(name); err == nil {
		return f, nil
	}
	return l.base.Open(name)
}

// declaration returns the declaration file for module, if there is one.
func declaration(module string) ([]byte, bool) {
	src, err := fs.ReadFile(declarations, path.Join("declarations", strings.ReplaceAll(module, "::", "/")+".tpm"))
	return src, err == nil
}

// readDeclaration reads a declaration file. Its language is typed Perl (RFC
// 0001, "Typed Perl, in `.tpm` only"): a module's own source is never read
// this way, so `sub f (Ref $x) { }` there stays perl's prototype.
func readDeclaration(src []byte, res *resolver) moduleFacts {
	return readDeclarationWith(newParser(src, res, true))
}

// readDeclarationWith is readDeclaration with the parser it reads by.
func readDeclarationWith(p *parser) moduleFacts {
	root, p := parseWith(p)
	src := p.src
	facts := readModule(root)
	facts.signatures, facts.errs, facts.operators, facts.relations = p.signatures, p.typedErrs, p.operators, p.relations
	facts.declared, facts.variables = true, p.variables
	// A .tpm's exports are its marks alone; an export list is refused.
	facts.exports, facts.defaults, facts.tags, facts.opaque = nil, nil, nil, false
	facts.errs = append(facts.errs, refuseInDeclaration(root, src)...)
	readExportMarks(root, &facts)
	// A plain `sub` replaces the one declared before it, unless the name is
	// also a `multi sub`: then it is a multi, keeping every candidate in the
	// order declared, with a warning (RFC 0001, "Multi declarations").
	for _, name := range slices.Sorted(maps.Keys(p.declaredPlain)) {
		if p.declaredMulti[name] {
			facts.warns = append(facts.warns, fmt.Errorf("sub %s: declared both as sub and as multi sub; read as a multi", name))
		} else if sigs := facts.signatures[name]; len(sigs) > 1 {
			facts.signatures[name] = sigs[len(sigs)-1:]
		}
	}
	// A statement the parser could not read declares nothing, so it is an
	// error rather than a silent gap: `sub :lvalue f;` is no declaration.
	//
	// ponytail: a malformed typed signature also leaves its statement
	// Unknown, so unread statements are reported only when the typed reader
	// raised nothing; the file is in error either way. Per-statement error
	// spans would report both when a file has several faults.
	for _, n := range root.Children {
		if n.Kind == Unknown && len(p.typedErrs) == 0 {
			facts.errs = append(facts.errs, fmt.Errorf("not a declaration: %q", strings.TrimSpace(n.SourceText(src))))
		}
	}
	inError := p.inError
	if inError == nil {
		inError = map[string]bool{}
	}
	// A name whose candidates are ambiguous records none, as a signature
	// that cannot be read records nothing (RFC 0001, "Multi declarations").
	for _, name := range slices.Sorted(maps.Keys(facts.signatures)) {
		if err := types.Ambiguity(facts.signatures[name]); err != nil {
			facts.errs = append(facts.errs, fmt.Errorf("sub %s: %w", name, err))
			delete(facts.signatures, name)
			inError[name] = true
		}
	}
	// An operator's `multi sub` candidates (RFC 0001, "Operators that
	// fork") are held to the same rule. A prefix and an infix of one
	// symbol are two operators.
	cands := map[[2]string][]types.Signature{}
	for _, op := range facts.operators {
		if op.multi {
			key := [2]string{op.name, op.fixity}
			cands[key] = append(cands[key], op.sig)
		}
	}
	ambiguous := map[[2]string]bool{}
	for _, op := range facts.operators {
		key := [2]string{op.name, op.fixity}
		sigs, unchecked := cands[key]
		if !unchecked {
			continue
		}
		delete(cands, key)
		if err := types.Ambiguity(sigs); err != nil {
			facts.errs = append(facts.errs, fmt.Errorf("sub %s: %w", op.name, err))
			ambiguous[key] = true
		}
	}
	facts.operators = slices.DeleteFunc(facts.operators, func(op operatorDecl) bool {
		return op.multi && ambiguous[[2]string{op.name, op.fixity}]
	})
	// A declaration with only a prototype gets the typed signature its
	// prototype gives (RFC 0001, "A typed signature and a prototype say
	// the same thing"). With none, it takes the rest of the call; an empty
	// one, `()`, takes nothing.
	for _, name := range slices.Sorted(maps.Keys(facts.protos)) {
		proto := facts.protos[name]
		if _, typed := facts.signatures[name]; typed || inError[name] {
			continue
		}
		sig, derived := types.Signature{Params: []types.Param{{Sigil: '@', Type: types.List}}}, true
		if proto != "" {
			var err error
			sig, derived, err = typesFromPrototype(strings.TrimSuffix(strings.TrimPrefix(proto, "("), ")"))
			if err != nil {
				facts.errs = append(facts.errs, fmt.Errorf("sub %s: %w", name, err))
				continue
			}
		}
		if derived {
			if facts.signatures == nil {
				facts.signatures = map[string][]types.Signature{}
			}
			facts.signatures[name] = []types.Signature{sig}
		}
	}
	// A typed declaration gets its prototype from its types, and needs no
	// `:prototype(...)`; one that states both must have them agree.
	for _, name := range slices.Sorted(maps.Keys(facts.signatures)) {
		sigs := facts.signatures[name]
		// One line refuses `:unary` beside `:listop` (decl.go); a multi's
		// candidates may not split them either, or the shape would follow
		// line order.
		if slices.ContainsFunc(sigs, func(s types.Signature) bool { return s.Unary }) &&
			slices.ContainsFunc(sigs, func(s types.Signature) bool { return s.ListOp }) {
			facts.errs = append(facts.errs, fmt.Errorf("sub %s: :unary and :listop are two parses; a builtin parses as one", name))
			delete(facts.signatures, name)
			continue
		}
		// `:listop` states a builtin perl gives no prototype but whose
		// parameters are typed and positional (RFC 0001, "Builtins with no
		// prototype"), so its types derive none, a multi's candidates as
		// a single line's: split's `(Regex|Str $pattern = ' ', Str $string
		// = $_, Int $limit = 0)` would otherwise claim the `;$_$` perl does
		// not report.
		if slices.ContainsFunc(sigs, func(s types.Signature) bool { return s.ListOp }) {
			if declared := facts.protos[name]; declared != "" {
				facts.errs = append(facts.errs, fmt.Errorf("sub %s: :prototype%s disagrees with :listop, which derives no prototype", name, declared))
				delete(facts.signatures, name)
			}
			continue
		}
		// A multi's candidates derive one prototype. Where they derive
		// none, a stated prototype is an error; without one they have
		// none, as perl reports none for grep and select.
		if len(sigs) > 1 {
			proto, err := prototypeFromCandidates(sigs)
			declared := facts.protos[name]
			switch {
			case declared != "" && err != nil:
				err = fmt.Errorf(":prototype%s disagrees with its types: %w", declared, err)
			case declared != "":
				err = agreement(declared, proto)
			case err == nil && proto != "@":
				facts.protos[name] = "(" + proto + ")"
			}
			if declared != "" && err != nil {
				facts.errs = append(facts.errs, fmt.Errorf("sub %s: %w", name, err))
				delete(facts.signatures, name)
			}
			continue
		}
		// No prototype character takes an argument with no comma after
		// it, so a signature with an invocant derives none, as perl
		// reports none for print (RFC 0001, "Builtins that keep their own
		// parse").
		if sigs[0].Invocant != nil {
			if declared := facts.protos[name]; declared != "" {
				facts.errs = append(facts.errs, fmt.Errorf("sub %s: :prototype%s disagrees with its types: an invocant colon derives no prototype", name, declared))
				delete(facts.signatures, name)
			}
			continue
		}
		// `:unary` states the parse of a builtin perl gives no prototype
		// (RFC 0001, "Builtins with no prototype"), so its types derive
		// none: `defined`'s `(Scalar $thing = $_)` would otherwise claim
		// the `_` perl does not report.
		if sigs[0].Unary {
			if declared := facts.protos[name]; declared != "" {
				facts.errs = append(facts.errs, fmt.Errorf("sub %s: :prototype%s disagrees with :unary, which derives no prototype", name, declared))
				delete(facts.signatures, name)
			}
			continue
		}
		proto := prototypeFromTypes(sigs[0])
		var err error
		if facts.protos[name] != "" {
			err = agreement(facts.protos[name], proto)
		}
		switch {
		case err != nil:
			facts.errs = append(facts.errs, fmt.Errorf("sub %s: %w", name, err))
			delete(facts.signatures, name)
		// A derived `@` is no prototype: perl reports none for a sub
		// without one, and the two parse alike.
		case facts.protos[name] == "" && proto != "@":
			facts.protos[name] = "(" + proto + ")"
		}
	}
	return facts
}

// readExportMarks reads what a declaration's `:export` and `:export_ok`
// state into its export lists (RFC 0001, "Every exported name"): `:export`
// is what a bare `use` imports, `:export_ok` what must be asked for, and
// either's argument names the tags that hold the name.
func readExportMarks(root *Node, facts *moduleFacts) {
	// A multi's first candidate's mark, which every later one must repeat:
	// perl exports a name, not a candidate.
	type mark struct {
		ok, bare, differs bool
		tags              []string
	}
	firstMark := map[string]*mark{}
	for _, stmt := range root.Children {
		for _, n := range stmt.Children {
			multi := n.Kind == Declaration && n.Text == "multi" && len(n.Children) == 1
			if multi {
				n = n.Children[0]
			}
			var name string
			switch {
			case n.Kind == Declaration && n.Text == "sub":
				name, _ = declaredSub(n)
			case n.Kind == Declaration && n.Text == "our":
				// A typed or marked `our` declares one variable, and one
				// whose type is in error (declareVariable) exports nothing.
				typed := slices.ContainsFunc(n.Children, func(c *Node) bool { return c.Kind == TypeName })
				_, _, marked := exportMark(n)
				i := slices.IndexFunc(n.Children, func(c *Node) bool { return c.Kind == Term })
				if i < 0 {
					if typed || marked {
						facts.errs = append(facts.errs, fmt.Errorf("our: declare one variable per statement"))
					}
				} else if _, recorded := facts.variables[n.Children[i].Text]; recorded || !typed {
					name = n.Children[i].Text
				}
			}
			bare, tags, ok := exportMark(n)
			if multi && name != "" {
				m := &mark{ok: ok, bare: bare, tags: slices.Sorted(slices.Values(tags))}
				if first := firstMark[name]; first == nil {
					firstMark[name] = m
				} else if !first.differs && (first.ok != m.ok || first.bare != m.bare || !slices.Equal(first.tags, m.tags)) {
					first.differs = true
					facts.errs = append(facts.errs, fmt.Errorf("multi sub %s: candidates mark %s differently", name, name))
				}
			}
			if name == "" || !ok || slices.Contains(facts.exports, name) {
				continue
			}
			facts.exports = append(facts.exports, name)
			if bare {
				facts.defaults = append(facts.defaults, name)
			}
			for _, tag := range tags {
				if facts.tags == nil {
					facts.tags = map[string][]string{}
				}
				facts.tags[tag] = append(facts.tags[tag], name)
			}
		}
	}
}

// declarationSubAttrs are the attributes a .tpm's sub may carry: perl's
// own -- measured on 5.42.0, these are the CODE attributes a named sub
// takes -- and the declaration language's.
var declarationSubAttrs = map[string]bool{
	"lvalue": true, "method": true, "prototype": true,
	"export": true, "export_ok": true, "unary": true, "listop": true, "context": true,
	"infix": true, "prefix": true, "postfix": true,
	"tighter": true, "looser": true, "equiv": true, "assoc": true,
	"keyword": true, "statement": true, "sublike": true,
}

// refuseInDeclaration is what a .tpm states that its language does not
// (RFC 0001, "Every exported name"): an export list, which :export and
// :export_ok replace, and an attribute no sub or our takes, as a
// misspelled mark would export nothing without a word. A typed our's
// attributes are refused by declareVariable.
func refuseInDeclaration(root *Node, src []byte) []error {
	var errs []error
	// An export list is refused wherever the file names one: declared,
	// assigned, pushed onto, package-qualified, inside a block, by `use
	// vars`, or by an element of it.
	listed := func(n *Node, v string) {
		errs = append(errs, fmt.Errorf("line %d: %s: a .tpm marks each export :export or :export_ok, not in a list", lineIndex(src).at(n.Start), v))
	}
	var walk func(n *Node, prefix string)
	walk = func(n *Node, prefix string) {
		if n.Kind == Declaration {
			prefix = n.Text + " "
		}
		for i, c := range n.Children {
			// A sub's signature is its parameters: local, exporting nothing.
			if n.Kind == Declaration && n.Text == "sub" && c.Kind == List {
				continue
			}
			// A subscripted variable is the container its bracket names:
			// $EXPORT_TAGS{all} is in %EXPORT_TAGS, $EXPORT{x} in %EXPORT.
			v := c.Text
			if n.Kind == Index && i == 0 && !n.Arrow && v != "" {
				v = map[string]string{"[": "@", "{": "%"}[n.Text] + v[1:]
			}
			switch {
			case c.Kind != Term:
			case exportListVar(v):
				listed(c, prefix+c.Text)
			case n.Kind == Use && n.Children[0].Text == "vars" && strings.HasPrefix(c.Text, "qw") && len(c.Text) > 3:
				for _, w := range strings.Fields(c.Text[3 : len(c.Text)-1]) {
					if exportListVar(w) {
						listed(c, "use vars "+w)
					}
				}
			}
			walk(c, prefix)
		}
	}
	walk(root, "")
	for _, stmt := range root.Children {
		for _, n := range stmt.Children {
			if n.Kind == Declaration && n.Text == "multi" && len(n.Children) == 1 {
				n = n.Children[0]
			}
			if n.Kind != Declaration {
				continue
			}
			typed := slices.ContainsFunc(n.Children, func(c *Node) bool { return c.Kind == TypeName })
			for _, c := range n.Children {
				if c.Kind != Attribute {
					continue
				}
				attr, _, _ := strings.Cut(strings.TrimPrefix(c.Text, ":"), "(")
				switch {
				case n.Text == "sub" && !declarationSubAttrs[attr]:
					name, _ := declaredSub(n)
					errs = append(errs, fmt.Errorf("sub %s: unknown attribute %s", name, c.Text))
				case n.Text == "our" && !typed && attr != "export" && attr != "export_ok":
					errs = append(errs, fmt.Errorf("our %s: unknown attribute %s; an our takes :export or :export_ok", n.Children[0].Text, c.Text))
				}
			}
		}
	}
	return errs
}

// exportListVar reports whether v names an export list, @EXPORT,
// @EXPORT_OK or %EXPORT_TAGS, in any package.
func exportListVar(v string) bool {
	if v == "" {
		return false
	}
	base := v[1:]
	if i := strings.LastIndex(base, "::"); i >= 0 {
		base = base[i+2:]
	}
	switch v[0] {
	case '@':
		return base == "EXPORT" || base == "EXPORT_OK"
	case '%':
		return base == "EXPORT_TAGS"
	}
	return false
}

// exportMark reads a declaration's `:export` or `:export_ok`; see
// exportAttr. ok is false for a declaration carrying neither.
func exportMark(n *Node) (bare bool, tags []string, ok bool) {
	for _, c := range n.Children {
		if bare, tags, ok = exportAttr(c); ok {
			return bare, tags, true
		}
	}
	return false, nil, false
}

// exportAttr reads the attribute c if it is `:export` or `:export_ok`:
// whether a bare `use` imports the name it marks, and the tags its
// argument names.
func exportAttr(c *Node) (bare bool, tags []string, ok bool) {
	if c.Kind != Attribute {
		return false, nil, false
	}
	attr, args, _ := strings.Cut(strings.TrimPrefix(c.Text, ":"), "(")
	if attr != "export" && attr != "export_ok" {
		return false, nil, false
	}
	for _, tag := range strings.Split(strings.TrimSuffix(args, ")"), ",") {
		if tag = strings.TrimSpace(tag); tag != "" {
			tags = append(tags, tag)
		}
	}
	return attr == "export", tags, true
}

// readLibraryDeclaration reads a library's declaration file, in CORE.tpm's
// language (RFC 0001, "One language for every `.tpm`"). Its infix
// operators must take a level among the parser's (libraryInfix): a bare
// `:infix` places its operator at none, and relations that derive no order
// place none of them. Either is an error, and the operators it is about
// are not recorded.
func readLibraryDeclaration(src []byte, res *resolver) moduleFacts {
	facts := readDeclaration(src, res)
	var infixOps []operatorDecl
	facts.operators = slices.DeleteFunc(facts.operators, func(op operatorDecl) bool {
		if op.fixity != "infix" {
			return false
		}
		if op.class == "" && op.tighter == nil && op.looser == nil && op.equiv == nil {
			facts.errs = append(facts.errs, fmt.Errorf("sub %s: :infix places it at no level; state :tighter, :looser or :equiv", op.name))
			return true
		}
		infixOps = append(infixOps, op)
		return false
	})
	if len(infixOps) == 0 {
		return facts
	}
	if _, err := libraryInfix(infixOps); err != nil {
		facts.errs = append(facts.errs, err)
		facts.operators = slices.DeleteFunc(facts.operators, func(op operatorDecl) bool { return op.fixity == "infix" })
	}
	return facts
}

// coreTable is perl's builtins by name, each to its prototype without
// parentheses: `bless` to `$;$`. It is built by parsing declarations/CORE.tpm,
// the declaration file for the language itself -- TypeScript's lib.d.ts to a
// module declaration's @types/Foo.
//
// Built on first use rather than at package init: the parse that builds it is
// the parser that consults it. CORE.tpm aliases nothing, so building it never
// asks for it.
func coreTable() map[string]string {
	readCore()
	return coreMap
}

// coreSignatures is perl's builtins by name, each to the typed signatures
// CORE.tpm states or derives for it -- print's too, which has no prototype
// and so no place in coreTable.
func coreSignatures() map[string][]types.Signature {
	readCore()
	return coreSigs
}

// CoreBuiltin is the typed signatures CORE.tpm states or derives for the
// builtin name, one per candidate, and nil for a name it does not declare.
// internal/infer types builtin calls from it.
func CoreBuiltin(name string) []types.Signature {
	return coreSignatures()[name]
}

// DeclaredSignatures reads a library's .tpm source as the resolver reads
// it, and returns its typed signatures by sub name, one per candidate,
// with the errors the file states. Tests that hold a declaration to perl
// read its return types from it.
func DeclaredSignatures(src []byte) (map[string][]types.Signature, []error) {
	facts := readLibraryDeclaration(src, nil)
	return facts.signatures, facts.errs
}

// DeclaredVariables reads a library's .tpm source as the resolver reads
// it, and returns its typed `our` variables by name with sigil, with the
// errors the file states. Tests that hold a declared variable's type to
// perl read it from here.
func DeclaredVariables(src []byte) (map[string]types.Param, []error) {
	facts := readLibraryDeclaration(src, nil)
	return facts.variables, facts.errs
}

// CoreOperator is the typed signatures CORE.tpm declares for the operator
// spelled op of fixity "infix", "prefix" or "postfix", one per candidate, and nil for
// one it does not declare. internal/infer types operators from it.
func CoreOperator(op, fixity string) []types.Signature {
	readCore()
	return coreOps[[2]string{op, fixity}]
}

// coreBarewordOperators is the infix operators CORE.tpm declares whose
// left operand is `:bareword`, read as a word: `=>`.
func coreBarewordOperators() []string {
	readCore()
	return coreBareword
}

// coreShapes is perl's builtins by name, each to the shape a call to it
// parses in: named unary, list operator or niladic (deriveShapes).
func coreShapes() map[string]Shape {
	readCore()
	return coreShp
}

// deriveShapes is the parse shape of each builtin CORE.tpm declares. A
// builtin with a prototype parses as a sub with that prototype would
// (ShapeOf); one with none is a named unary when `:unary` says so (RFC 0001,
// "Builtins with no prototype"), and otherwise a list operator, as a sub
// with no prototype is: split's `:listop` line, whose types derive no
// prototype, is one.
func deriveShapes(protos map[string]string, sigs map[string][]types.Signature) map[string]Shape {
	shapes := map[string]Shape{}
	for name, cands := range sigs {
		if cands[0].Unary {
			shapes[name] = ShapeUnary
		} else {
			shapes[name] = ShapeList
		}
	}
	for name, proto := range protos {
		shapes[name] = ShapeOf("(" + proto + ")")
	}
	return shapes
}

// readCore reads CORE.tpm, once, into coreTable's, coreSignatures's,
// CoreOperator's and coreShapes's tables.
func readCore() {
	coreOnce.Do(func() {
		// Its tables are read with the parser's powers, as every parse is.
		derivePowers()
		src, ok := declaration("CORE")
		if !ok {
			panic("parse: declarations/CORE.tpm is not embedded")
		}
		var err error
		if coreMap, coreSigs, coreOps, err = coreDeclarations(src); err != nil {
			panic("parse: declarations/CORE.tpm: " + err.Error())
		}
		coreShp = deriveShapes(coreMap, coreSigs)
		for key, sigs := range coreOps {
			if key[1] == "infix" && slices.ContainsFunc(sigs, func(s types.Signature) bool { return s.Params[0].Bareword }) {
				coreBareword = append(coreBareword, key[0])
			}
		}
		slices.Sort(coreBareword)
	})
}

// coreProtos reads a CORE.tpm into the tables coreTable and coreSignatures
// hold. A file in error builds neither: a builtin it misdeclares would parse
// wrong.
func coreProtos(src []byte) (map[string]string, map[string][]types.Signature, error) {
	protos, sigs, _, err := coreDeclarations(src)
	return protos, sigs, err
}

// coreDeclarations is coreProtos with CoreOperator's table: each operator's
// candidates by its symbol and fixity.
func coreDeclarations(src []byte) (map[string]string, map[string][]types.Signature, map[[2]string][]types.Signature, error) {
	p := newParser(src, nil, true)
	p.buildingCore = true
	facts := readDeclarationWith(p)
	if len(facts.errs) > 0 {
		return nil, nil, nil, errors.Join(facts.errs...)
	}
	ops := map[[2]string][]types.Signature{}
	for _, op := range facts.operators {
		key := [2]string{op.name, op.fixity}
		ops[key] = append(ops[key], op.sig)
	}
	return protoTable(facts), facts.signatures, ops, nil
}

// protoTable is a CORE.tpm's builtins by name, each to its prototype
// without the parentheses.
func protoTable(facts moduleFacts) map[string]string {
	protos := map[string]string{}
	for name, proto := range facts.protos {
		// A builtin with no prototype, print's "", has no entry: aliasTarget
		// reads presence here as a known prototype.
		if proto == "" {
			continue
		}
		protos[name] = strings.TrimSuffix(strings.TrimPrefix(proto, "("), ")")
	}
	return protos
}

var (
	coreOnce sync.Once
	coreMap  map[string]string
	coreSigs map[string][]types.Signature
	coreOps  map[[2]string][]types.Signature
	coreShp  map[string]Shape
	// coreBareword is the infix operators whose left operand CORE.tpm
	// declares `:bareword`.
	coreBareword []string
)
