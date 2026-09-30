// ABOUTME: use/no/require, the phasers, and the 5.38 class syntax — the rest of what T2 contains.
// ABOUTME: Syntax only: what `use feature` enables is M2's, and BEGIN's compile-time effect is M4's.

package parse

import (
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

// phasers are the compile-and-run-phase blocks.
//
// `ADJUST` is 5.38's class phaser. It takes the same shape -- a keyword, a
// block, and no `;` required -- so it takes the same path. That it is only
// legal inside a `class` body is a semantic check and not a syntactic one:
// perl parses it anywhere and then rejects it at compile time with
// "Cannot 'ADJUST' outside of a 'class'", which is M2's to say, not ours.
var phasers = map[string]bool{
	"BEGIN": true, "END": true, "CHECK": true,
	"INIT": true, "UNITCHECK": true, "ADJUST": true,
}

// parseTheRest handles the statement forms this issue owns, or returns nil.
func (p *parser) parseTheRest(word lexer.Token) *Node {
	text := keywordName(p.text(word))
	switch {
	case text == "use" || text == "no" || text == "require":
		return p.parseUse(word)
	case phasers[text]:
		return p.parsePhaser(word)
	case text == "DESTROY" || text == "AUTOLOAD":
		return p.parseSpecialSub(word)
	case text == "class":
		return p.parseClass(word)
	}
	return nil
}

// parseUse: `use MODULE LIST;`, `use VERSION;`, and the `no` and `require`
// spellings.
//
// The import list is an ordinary expression, which covers every shape the T2
// corpus actually contains -- surveyed rather than guessed:
//
//	12  use v5.36
//	12  use feature 'class'
//	12  no warnings 'experimental::class'
//	 2  no warnings qw(syntax deprecated)
//	 1  use test_use { () }
//
// The last one is why the list is parsed as an expression rather than as a
// comma-separated list of literals: `{ () }` is a term, and a narrower parser
// would decline it.
func (p *parser) parseUse(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Use, Text: p.text(word), Start: word.Start}
	var module string

	// The module name is a BAREWORD, not an expression.
	//
	// Parsing it as one makes `use feature 'class';` leave the string
	// orphaned: `feature` becomes an unresolved Call, which by design
	// consumes no arguments, so `'class'` is never read and the statement
	// falls to Unknown. Measured on five t/class files before this split.
	//
	// So the name is taken directly and the import list is parsed after it.
	if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
		p.advanceTo(name)
		term := &Node{
			Kind: Term, Text: p.text(name),
			Start: name.Start, End: name.End,
		}

		// A version bundle USED TO BE three tokens: `use v5.36;` lexed as
		// Word("v5") Operator(".") Number(36), because `v5` is a valid
		// identifier and the lexer had no reason to know better. Taking only
		// the Word left `.36` behind, which is not an expression, so the
		// statement fell to Unknown -- measured on seven t/class files. The
		// whole run was absorbed into one term instead.
		//
		// `scanVString` now emits ONE Quote for any `v`-prefixed run with a
		// dot, so a version never reaches this branch: the guard below is
		// `name.Kind == lexer.Word` and a Quote misses it. This code and
		// `isVersionPrefix` have no reachable caller -- tracked by 01a0dc74,
		// which also covers the rest of the stranded reassembly surface.
		if isVersionPrefix(term.Text) {
			for {
				dot, ok := p.peekSignificant()
				if !ok || p.text(dot) != "." {
					break
				}
				num, ok := p.peekAfter(dot)
				if !ok || num.Kind != lexer.Number {
					break
				}
				p.advanceTo(num)
				term.End = num.End
			}
		}
		module = term.Text
		n.Children = append(n.Children, term)
	}

	// The import list, or a version. A version is now ONE Quote token
	// (`scanVString`), so it misses the Word-shaped name path above and
	// reads as an expression here -- which round-trips even though the shape
	// is not what perl's grammar calls a version.
	//
	// It formerly lexed as Word("v5") Operator(".") Number, with the Word
	// half taken as the name above and the remainder reassembled. The Term's
	// text was the truncated "v5" then and is the whole "v5.36" now; the
	// span was already correct either way.
	//
	// A `}` ends the statement as surely as a `;` does -- the final semicolon
	// of a block's last statement is optional in perl, and `eval { require
	// Errno }` as a feature probe is how perl's own suite writes it. Asking
	// only `!= Semicolon` read the closer as the start of an import list and
	// the operand hunt then CONSUMED it, taking the `}` out of the enclosing
	// block: 10 files and 77 nodes of perl.git t/, and a wrong tree rather
	// than only a count, because canon then emitted a spurious `};`.
	//
	// endsStatement is the predicate that already answers this, and answers it
	// for the same reason -- "a `}` belongs to an enclosing construct".
	var list *Node
	if next, ok := p.peekSignificant(); ok && !endsStatement(next, p.src) {
		if arg := p.parseExpr(0); arg != nil {
			list = arg
			n.Children = append(n.Children, arg)
		}
	}

	// Resolve the module, if this parse has a loader and the name is a module
	// rather than a version bundle.
	//
	// Enrichment only: whether the source is found changes what a later call
	// KNOWS, never whether this statement parses. `use` is keyword, bareword,
	// optional list, semicolon, and that is true whether or not the file
	// exists. perl dies here; this parser must not, because parsing text is
	// not running it.
	//
	// `no` is deliberately not resolved, because it UNIMPORTS.
	//
	// `require MODULE` is not resolved either, and for the reason that used to
	// cover `require` entirely: it is a runtime load with NO IMPORT AT ALL, so
	// there is no export list to apply and nothing enters the caller's scope
	// by name.
	//
	// `require './FILE.pl'` is the exception, and what makes it one is that
	// perl's own suite writes its helpers that way: a required `.pl` file
	// declares subs into the CALLER'S package -- it has no `package`
	// statement of its own -- so every `sub NAME` in it is a name the
	// requiring file can call. There is no import list to consult because
	// there is no import; the declarations themselves are the interface.
	// See resolveRequiredFile for the boundary that keeps this safe.
	p.noteFeatures(n.Text, module, list)
	if module != "" && !isVersionPrefix(module) && n.Text != "no" {
		p.notePackage(module)
	}

	switch {
	case module != "" && !isVersionPrefix(module) && n.Text == "use":
		p.resolveImports(module, list)
	case module == "" && n.Text == "require":
		p.resolveRequiredFile(list)
	}

	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
	}
	n.End = p.prevEnd()
	return n
}

// noteFeatures records what a `use` or `no` statement does to the gated
// builtins. `use feature LIST` turns each named feature on and `no feature
// LIST` off; `use VERSION` from 5.15 up turns on the bundle, which names both
// `fc` and `evalbytes` -- measured, perl's %feature::feature_bundle holds them
// from "5.15" on and not in "5.14".
func (p *parser) noteFeatures(verb, module string, list *Node) {
	if p.features == nil {
		p.features = map[string]bool{}
	}
	switch {
	case module == "feature" && (verb == "use" || verb == "no") && list != nil:
		names, ok := literalNameList(list)
		if !ok {
			return
		}
		for _, name := range names {
			if gatedUnary[name] || name == "keyword_any" || name == "keyword_all" || name == "class" || name == "defer" || name == "isa" {
				p.features[name] = verb == "use"
			}
			if name == "indirect" {
				p.noIndirect = verb == "no"
			}
		}
	case module == "" && verb == "use" && list != nil:
		major, minor, ok := perlVersion(list.Text)
		if ok && (major > 5 || major == 5 && minor >= 15) {
			for name := range gatedUnary {
				p.features[name] = true
			}
		}
		// The 5.36 bundle drops `indirect` -- measured, `use v5.36; new
		// Foo;` is a syntax error on 5.42.0.
		if ok && (major > 5 || major == 5 && minor >= 35) {
			p.noIndirect = true
			// And takes `isa` in: feature.pm's :5.36 bundle names it.
			p.features["isa"] = true
		}
	}
}

// perlVersion reads a `use VERSION` argument: `v5.16`, `5.16.0` or the
// decimal `5.016`, whose fraction perl reads in groups of three digits.
func perlVersion(text string) (major, minor int, ok bool) {
	text, dotted := strings.CutPrefix(text, "v")
	parts := strings.Split(text, ".")
	if len(parts) < 2 {
		return 0, 0, false
	}
	var err error
	if major, err = strconv.Atoi(parts[0]); err != nil {
		return 0, 0, false
	}
	frac := parts[1]
	if len(parts) == 2 && !dotted {
		// Decimal: 5.016 is 5.16, the first three digits of the fraction.
		for len(frac) < 3 {
			frac += "0"
		}
		frac = frac[:3]
	}
	if minor, err = strconv.Atoi(frac); err != nil {
		return 0, 0, false
	}
	return major, minor, true
}

// resolveImports reads a module's source and records what it brought in.
//
// A module that cannot be found contributes nothing, and calls to what it
// would have exported stay unresolved -- correct, not a gap.
func (p *parser) resolveImports(module string, list *Node) {
	if module == "builtin" {
		p.importBuiltins(list)
		return
	}
	if module == "constant" {
		p.importConstants(list)
		return
	}
	if p.res == nil {
		return
	}
	// `use if COND, MODULE, LIST`, with COND taken to hold: `use MODULE
	// LIST`. See ParseFileAssumingUseIf.
	if module == "if" && p.res.useIfPasses {
		if target, rest, ok := useIfTarget(list); ok {
			p.notePackage(target)
			p.resolveImports(target, rest)
		}
		return
	}
	facts, ok := p.res.resolve(module)
	if !ok {
		p.noteImportKnowledge(module, list, true)
		return
	}
	p.noteImportKnowledge(module, list, facts.dynamic || facts.opaque)

	// The module's own subs are callable by their qualified names whatever
	// it exports: after `use overload;`, `overload::constant 'integer' =>
	// sub {...}` calls lib/overload.pm's `sub constant`. Recorded before the
	// import list is read, which a computed list abandons: loading is not
	// importing. Keyed under the module's name, which is the package a
	// conventional module declares.
	//
	// Kept apart from the imports: they are known, not imported.
	// A name already qualified -- an XS PACKAGE other than the module's
	// own, or a `sub Other::name` -- is known by that name.
	for name, proto := range facts.protos {
		if p.moduleSubs == nil {
			p.moduleSubs = map[string]Import{}
		}
		q := name
		if !strings.Contains(name, "::") {
			q = module + "::" + name
		}
		p.moduleSubs[q] = Import{Name: q, Prototype: proto, PrototypeKnown: true}
	}

	if facts.builder {
		names, given, ok := builderImportList(list)
		if !ok {
			p.symbolsOpen = true
			return
		}
		if p.imports == nil {
			p.imports = map[string]Import{}
		}
		for _, imp := range importsFrom(facts, names, given) {
			p.imports[subKey(imp.Name)] = imp
		}
		return
	}

	// An import list restricts what is imported, and `use M ()` -- an empty
	// list -- is the explicit "load but import nothing" form, which is NOT
	// the same as omitting the list.
	var names []string
	listGiven := list != nil
	if listGiven {
		if got, ok := literalNameList(list); ok {
			names = got
		} else if !list.Paren || len(list.Children) != 0 {
			// A computed import list is opaque: it is not an empty import.
			// It hides subs only from a module whose import defines them --
			// `use overload '%{}' => sub {...}` defines none.
			if !quietPragmas[module] && module != "if" {
				p.symbolsOpen = true
			}
			return
		}
	}

	if p.imports == nil {
		p.imports = map[string]Import{}
	}
	for _, imp := range importsFrom(facts, names, listGiven) {
		p.imports[subKey(imp.Name)] = imp
	}
}

// listItems flattens a `use` list -- commas, fat commas and parens -- into
// its elements.
func listItems(n *Node) []*Node {
	if n.Kind == Binary && (n.Text == "," || n.Text == "=>") || n.Kind == List {
		var items []*Node
		for _, c := range n.Children {
			items = append(items, listItems(c)...)
		}
		return items
	}
	return []*Node{n}
}

// importConstants is `use constant`. constant.pm defines each name in the
// caller as a sub with the empty prototype -- measured on 5.42.0,
// `prototype "main::X"` is "" for `use constant X => 2`, for a hash of them
// and for a list constant -- so `Y / X` divides. There is no module text to
// read, as there is none for `use builtin`, so this holds with no loader.
// The first element names one constant, or a `{...}` names one per key.
func (p *parser) importConstants(list *Node) {
	if list == nil {
		return
	}
	items := listItems(list)
	var names []string
	if first := items[0]; first.Kind == AnonHash {
		for _, c := range first.Children {
			for i, kv := range listItems(c) {
				if i%2 == 0 {
					if got, ok := constantName(kv); ok {
						names = append(names, got)
					}
				}
			}
		}
	} else if got, ok := constantName(first); ok {
		names = append(names, got)
	}
	if p.imports == nil {
		p.imports = map[string]Import{}
	}
	for _, name := range names {
		p.imports[subKey(name)] = Import{Name: name, Prototype: "()", PrototypeKnown: true}
	}
}

// constantName reads a constant's name: a bareword or a literal string.
func constantName(n *Node) (string, bool) {
	if got, ok := literalNameList(n); ok && len(got) == 1 {
		return got[0], true
	}
	if len(n.Children) == 0 && isBarewordText(n.Text) {
		return n.Text, true
	}
	return "", false
}

// builderImportList reads a Test::Builder::Module subclass's import list as
// its import does (perl 5.42.0's Test/Builder/Module.pm:75-122): the names
// under each `import => [...]` are imported, and everything else is handed to
// plan(). given is false when there is no `import` pair, which imports
// @EXPORT. ok is false when a pair's names are not literal.
func builderImportList(list *Node) (names []string, given, ok bool) {
	if list == nil {
		return nil, false, true
	}
	items := listItems(list)
	for i := 0; i < len(items); i++ {
		key, lit := literalNameList(items[i])
		if !(lit && len(key) == 1 && key[0] == "import") && items[i].Text != "import" {
			continue
		}
		if i+1 >= len(items) || items[i+1].Kind != AnonArray {
			return nil, false, false
		}
		for _, c := range items[i+1].Children {
			got, lit := literalNameList(c)
			if !lit {
				return nil, false, false
			}
			names = append(names, got...)
		}
		given = true
		i++
	}
	return names, given, true
}

// useIfTarget splits `use if`'s list into the module it loads and that
// module's own import list, which is nil when none is given.
func useIfTarget(list *Node) (module string, rest *Node, ok bool) {
	if list == nil {
		return "", nil, false
	}
	items := listItems(list)
	if len(items) < 2 {
		return "", nil, false
	}
	switch names, lit := literalNameList(items[1]); {
	case lit && len(names) == 1:
		module = names[0]
	case len(items[1].Children) == 0 && isBarewordText(items[1].Text):
		module = items[1].Text
	default:
		return "", nil, false
	}
	for _, n := range items[2:] {
		if rest == nil {
			rest = n
			continue
		}
		rest = &Node{Kind: Binary, Text: ",", Start: rest.Start, End: n.End, Children: []*Node{rest, n}}
	}
	return module, rest, true
}

// quietPragmas are the core pragmas whose `import` defines no sub in the
// caller: they switch features, warnings, strictures, layers, @INC or @ISA.
// Their own modules define `import`, which would otherwise mark them dynamic.
// `constant` and `subs` are not here: both declare subs this parser does not
// record.
var quietPragmas = map[string]bool{
	"strict": true, "warnings": true, "utf8": true, "open": true,
	"vars": true, "feature": true, "lib": true, "integer": true,
	"bytes": true, "less": true, "sort": true, "overload": true,
	"parent": true, "base": true, "mro": true, "locale": true,
}

// noteImportKnowledge records whether a `use` left the sub table complete.
// A module that was not found, whose export list is computed, or that is
// dynamic may have defined subs in this file that the table does not hold,
// and from then on an unknown word is no longer one perl lacks a CV for.
//
// Config is generated when perl is built, so no source tree holds it; its
// bare import is %Config alone, and it exports functions only when named.
//
// `use if COND, MODULE, LIST` imports only when COND holds at compile time,
// which this parser cannot evaluate. The import is assumed to fail (perigrin's
// decision, 2026-09-30): it defines nothing, and the table stays complete.
func (p *parser) noteImportKnowledge(module string, list *Node, unseen bool) {
	switch {
	case quietPragmas[module] || module == "if":
	case module == "Config" && list == nil:
	case unseen || module == "Config":
		p.symbolsOpen = true
	}
}

// importBuiltins is `use builtin LIST`. There is no module text to read: the
// functions are the interpreter's own, in interpreterSubs, and the list names
// which of them come into scope. A name builtin.c does not define -- a
// version bundle such as ':5.39' among them -- imports nothing here; the
// bundles are 01a0ec21-d4dc-7883-9f5b-457350a011e0.
func (p *parser) importBuiltins(list *Node) {
	if list == nil {
		return
	}
	names, ok := literalNameList(list)
	if !ok {
		return
	}
	for _, name := range names {
		proto, ok := interpreterSubs["builtin::"+name]
		if !ok {
			continue
		}
		if p.imports == nil {
			p.imports = map[string]Import{}
		}
		p.imports[subKey(name)] = Import{Name: name, Prototype: proto, PrototypeKnown: true}
	}
}

// resolveRequiredFile reads a `require './helper.pl'` and records every
// `sub NAME` the helper declares, so a parenless call to one of them parses.
//
// list is the require's argument as parsed -- a single Term for the literal
// spellings, anything at all otherwise.
//
// THE BOUNDARY IS THE SAFETY PROPERTY, and each rule is enforced at a named
// place rather than assumed:
//
//  1. A LITERAL string only, here. `require $file` and `require "./$n.pl"`
//     name a file only at runtime, so they contribute nothing. Measured over
//     perl.git's `t/`: 503 files hold a literal `require`, against 6 with a
//     computed one and 3 with an interpolated path, and BOTH kinds occur in
//     files that also hold a literal one. So refusing them costs nothing and
//     accepting them would be a guess at a name.
//  2. Resolved against the file's own directory and then the parse root, in
//     DirLoader. THE PARSE ROOT IS THE LOAD-BEARING HALF, and it was measured
//     to be so: 464 of those 503 files run `chdir 't' if -d 't'` before the
//     require, so `./test.pl` in `t/op/select.t` names `t/test.pl` and NOT
//     `t/op/test.pl`. Resolving against the requiring file's own directory
//     alone moves 0 files and 0 nodes; with `t/` as the root it moves 93 files
//     and 3,267 nodes. The caller supplies the root because only the caller
//     knows the working directory the suite is run from.
//  3. Only `sub NAME`, here: facts.protos and not facts.exports. A `.pl`
//     helper has no `package` and no `@EXPORT`, so its declarations ARE its
//     interface -- but a variable it declares is not a call shape and brings
//     nothing.
//  4. ONE LEVEL, in resolver.resolveFile. A helper's own `require` is not
//     followed, and it COSTS NOTHING -- measured rather than assumed, because
//     one helper does nest: `thread_it.pl` requires `./test.pl` at its line
//     10, so the 13 `*_thr.t` files that require `thread_it.pl` see its subs
//     and not `test.pl`'s. All 13 were ALREADY CLEAN before this change and
//     stayed clean after it: they are wrappers that set up a thread and
//     `require` the real test, with no parenless call site of their own. A
//     second level would buy zero files and zero nodes.
//  5. An unreadable or unparseable helper contributes nothing and NEVER fails
//     the parse -- resolveFile returns "not found", which is the answer a
//     missing module already gets. `t/` files are parsed from many working
//     directories, including by a suite with no perl5 checkout at all.
func (p *parser) resolveRequiredFile(list *Node) {
	if p.res == nil || list == nil {
		return
	}

	// Rule 1, and INTERPOLATION is the half that is easy to miss.
	// literalNameList reads a single quoted string, which is what a literal
	// `require` argument is -- but `"./$name.pl"` is ALSO a single quoted
	// string to it, and names a file only at runtime. A sigil inside a
	// double-quoted body is what separates the two, so it is checked on the
	// text AS WRITTEN, before the quotes come off.
	if interpolates(list.Text) {
		p.symbolsOpen = true
		return
	}
	names, ok := literalNameList(list)
	if !ok || len(names) != 1 || !isRequiredPath(names[0]) {
		return
	}

	facts, ok := p.res.resolveFile(names[0])
	if !ok || facts.dynamic {
		p.symbolsOpen = true
	}
	if !ok {
		return
	}

	// Rule 3. Every declared sub, with its prototype, because a `.pl` helper
	// declares into the CALLER'S package and has no export list to narrow
	// this. Not Local: the name came from another file, and `Local` marks a
	// sub this file declares itself.
	if p.imports == nil {
		p.imports = map[string]Import{}
	}
	for name, proto := range facts.protos {
		p.imports[subKey(name)] = Import{
			Name:           name,
			Prototype:      proto,
			PrototypeKnown: true,
		}
	}
	// A glob it assigns is a sub too, with no prototype this parser reads.
	for _, name := range facts.globs {
		if _, ok := p.imports[subKey(name)]; !ok {
			p.imports[subKey(name)] = Import{Name: name}
		}
	}
}

// isVersionPrefix reports whether a word is the `v5` of a `v5.36`.
func isVersionPrefix(text string) bool {
	if len(text) < 2 || (text[0] != 'v' && text[0] != 'V') {
		return false
	}
	for i := 1; i < len(text); i++ {
		if text[i] < '0' || text[i] > '9' {
			return false
		}
	}
	return true
}

// parsePhaser: `BEGIN { ... }` and the other four.
func (p *parser) parsePhaser(word lexer.Token) *Node {
	p.advanceTo(word)
	n := &Node{Kind: Phaser, Text: p.text(word), Start: word.Start}

	if blk := p.parseBlockOrDecline(); blk != nil {
		n.Children = append(n.Children, blk)
		if n.Text == "BEGIN" {
			p.noteBeginEffects(blk)
		}
	}
	if tok, ok := p.peekSignificant(); ok && tok.Kind == lexer.Semicolon {
		p.advanceTo(tok)
	}
	n.End = p.prevEnd()
	return n
}

// aliasTarget is the sub a `\&NAME` names, with its prototype when this
// parser knows it: a CORE:: builtin from coreProtos, or a sub in the table.
func (p *parser) aliasTarget(n *Node) (Import, bool) {
	if n.Kind != Unary || n.Text != "ref" || len(n.Children) != 1 {
		return Import{}, false
	}
	name, ok := strings.CutPrefix(n.Children[0].Text, "&")
	if !ok || name == "" {
		return Import{}, false
	}
	if builtin, isCore := strings.CutPrefix(name, "CORE::"); isCore {
		proto, known := coreProtos[builtin]
		return Import{Name: name, Prototype: "(" + proto + ")"}, known
	}
	if imp, known := p.lookupSub(name); known && imp.PrototypeKnown {
		return imp, true
	}
	return Import{}, false
}

// noteBeginEffects reads what a BEGIN block does to the sub table before the
// code after it compiles. A literal glob assignment defines a name: `BEGIN {
// *BB::e = \&C::e }` in t/op/method.t. A computed glob, an `import` call, a
// string eval or a `do FILE` may define anything, so the table is no longer
// complete. See symbolsOpen.
func (p *parser) noteBeginEffects(n *Node) {
	switch {
	case n.Kind == Binary && n.Text == "=" && len(n.Children) == 2 &&
		n.Children[0].Kind == Term && strings.HasPrefix(n.Children[0].Text, "*"):
		if name := strings.TrimPrefix(n.Children[0].Text, "*"); name != "" {
			if p.imports == nil {
				p.imports = map[string]Import{}
			}
			// An alias takes its target's prototype: `*my_push =
			// \&CORE::push` is `\@@`, measured on 5.42.0.
			if imp, ok := p.aliasTarget(n.Children[1]); ok {
				p.imports[subKey(name)] = Import{Name: name, Prototype: imp.Prototype, PrototypeKnown: true}
			} else if _, ok := p.imports[subKey(name)]; !ok {
				p.imports[subKey(name)] = Import{Name: name}
			}
		} else {
			p.symbolsOpen = true
		}
	case n.Text == "import" || n.Text == "unimport":
		p.symbolsOpen = true
	case n.Kind == Call && (n.Text == "eval" || n.Text == "do") &&
		len(n.Children) > 0 && n.Children[0].Kind != Block:
		p.symbolsOpen = true
	}
	for _, c := range n.Children {
		p.noteBeginEffects(c)
	}
}

// parseSpecialSub: `DESTROY { ... }` and `AUTOLOAD { ... }`, which declare the
// sub with no `sub` before them, or nil when no block follows.
//
// toke.c sends both to yyl_sub at a statement boundary, alongside the phasers:
//
//	case KEY_AUTOLOAD: case KEY_DESTROY: case KEY_BEGIN: ... case KEY_END:
//	    if (PL_expect == XSTATE)
//	        return yyl_sub(aTHX_ PL_bufptr, key);
//
// Unlike a phaser the result is an ordinary named sub -- Deparse writes `sub
// DESTROY { ... }` -- so it is built as the Declaration `sub NAME BLOCK`
// parseSubDecl would give, declared for the calls below it. Read only here,
// in statement position; elsewhere the word is a word, and `$obj->DESTROY`
// a method name.
func (p *parser) parseSpecialSub(word lexer.Token) *Node {
	if next, ok := p.peekAfter(word); !ok || p.text(next) != "{" {
		return nil
	}
	p.advanceTo(word)
	n := &Node{Kind: Declaration, Text: "sub", Start: word.Start}
	n.Children = append(n.Children, &Node{
		Kind: Term, Text: p.text(word), Start: word.Start, End: word.End,
	})
	p.declareSub(n)
	p.finishBodyOrSemicolon(n)
	return n
}

// parseClass: 5.38's `class NAME { ... }` and `class NAME;`.
//
// Structurally a package with a different keyword, so it shares
// finishBodyOrSemicolon. `field` and `method` inside the body are handled by
// parseDeclaration, which already knows `method` as a sub spelling.
func (p *parser) parseClass(word lexer.Token) *Node {
	p.advanceTo(word)
	// A class declaration means class syntax is in effect -- by the feature
	// or by Object::Pad, which provides the same keywords -- so `method`
	// declares from here on.
	if p.features == nil {
		p.features = map[string]bool{}
	}
	p.features["class"] = true
	n := &Node{Kind: Declaration, Text: p.text(word), Start: word.Start}

	if name, ok := p.peekSignificant(); ok && name.Kind == lexer.Word {
		p.advanceTo(name)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(name),
			Start: name.Start, End: name.End,
		})
		p.notePackage(p.text(name))
	}

	// An optional version: `class Point 1.0 { }`.
	if v, ok := p.peekSignificant(); ok && v.Kind == lexer.Number {
		p.advanceTo(v)
		n.Children = append(n.Children, &Node{
			Kind: Term, Text: p.text(v), Start: v.Start, End: v.End,
		})
	}

	// Attributes: `class Point :isa(Shape) { }`.
	p.parseAttributes(n)

	p.finishBodyOrSemicolon(n)
	return n
}

// parseAttributes reads a run of `:name` or `:name(args)` and appends them.
//
// Attributes hang off classes, fields, methods and subs alike, and each one
// is an Attribute node: canon has to tell a declaration's HEAD from its
// INITIALISER, and an attribute is head. As a Term it was indistinguishable
// and `my $x :shared = 1;` re-emitted as `my $x = :shared = 1;`.
func (p *parser) parseAttributes(n *Node) {
	for {
		colon, ok := p.peekSignificant()
		if !ok || p.text(colon) != ":" {
			return
		}
		name, ok := p.peekAfter(colon)
		if !ok || name.Kind != lexer.Word {
			return
		}
		p.advanceTo(name)
		attr := &Node{
			Kind: Attribute, Text: ":" + p.text(name),
			Start: colon.Start, End: name.End,
		}

		// An optional parenthesised argument, taken as an opaque span: an
		// attribute's argument is not Perl in general (`:isa(Shape)` is, but
		// `:lvalue` takes none and XS attributes take arbitrary text).
		//
		// `:prototype(...)` arrives as ONE Prototype token, because its
		// contents are a prototype and the lexer scans them the way it scans
		// `sub f ($$)`. Balancing brackets cannot read it: `$)` is a real
		// perl variable and would swallow the closing paren.
		if proto, ok := p.peekSignificant(); ok && proto.Kind == lexer.Prototype {
			p.advanceTo(proto)
			attr.Text += p.text(proto)
			attr.End = proto.End
			n.Children = append(n.Children, attr)
			continue
		}
		// Only a paren TOUCHING the name is its argument: toke.c tests
		// `*d == '('` directly after it. A spaced one is a syntax error
		// without signatures and the signature with them -- measured on
		// 5.42.0, `sub g :lvalue ($x, $) { $x }` deparses unchanged.
		if open, ok := p.peekSignificant(); ok && p.text(open) == "(" && open.Start == name.End {
			depth := 0
			for p.pos < len(p.toks) {
				tok := p.toks[p.pos]
				p.pos++
				if p.text(tok) == "(" {
					depth++
				}
				if tok.Kind == lexer.CloseBracket && p.src[tok.Start] == ')' {
					depth--
					if depth == 0 {
						break
					}
				}
			}
			attr.End = p.prevEnd()
			// Text, not just the span. Canon writes Text, so extending only
			// End DROPPED the argument: measured, `class Point :isa(Shape)`
			// re-emitted as `class Point :isa` and lost its parent class.
			// Unknown could not see it -- the tree was right and only the
			// emission was wrong -- so the count stayed at zero.
			attr.Text = string(p.src[attr.Start:attr.End])
		}
		n.Children = append(n.Children, attr)
	}
}

// peekAfter returns the next significant token after the given one.
func (p *parser) peekAfter(tok lexer.Token) (lexer.Token, bool) {
	save := p.pos
	p.advanceTo(tok)
	next, ok := p.peekSignificant()
	p.pos = save
	return next, ok
}
