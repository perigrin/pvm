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
// `declarations/Moose/Role.pmt`. TypeScript's `@types/Foo` is the model: an
// interface ASSERTED over code this parser never reads, for a module whose
// own `import` builds what it defines -- Moose::Exporter, a monkey_patch
// loop -- so that no reading of its source can recover it.
//
// A declaration is written in the module syntax the resolver already reads:
// `our @EXPORT` for what a bare `use` imports, `sub NAME :prototype(PROTO);` for each
// sub and its prototype. So it is read by the same parse as a module.
//
//go:embed declarations
var embedded embed.FS

// declarations is where declaration files are read from: the embedded
// ones, or a test's own.
var declarations fs.FS = embedded

// declaration returns the declaration file for module, if there is one.
func declaration(module string) ([]byte, bool) {
	src, err := fs.ReadFile(declarations, path.Join("declarations", strings.ReplaceAll(module, "::", "/")+".pmt"))
	return src, err == nil
}

// readDeclaration reads a declaration file. Its language is typed Perl (RFC
// 0001, "Typed Perl, in `.pmt` only"): a module's own source is never read
// this way, so `sub f (Ref $x) { }` there stays perl's prototype.
func readDeclaration(src []byte, res *resolver) moduleFacts {
	return readDeclarationWith(newParser(src, res, true))
}

// readDeclarationWith is readDeclaration with the parser it reads by.
func readDeclarationWith(p *parser) moduleFacts {
	root, p := parseWith(p)
	src := p.src
	facts := readModule(root)
	facts.signatures, facts.errs, facts.operators = p.signatures, p.typedErrs, p.operators
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

// readLibraryDeclaration reads a library's declaration file, in CORE.pmt's
// language (RFC 0001, "One language for every `.pmt`") less the operator
// classes coined for CORE.pmt alone. An operator naming one is an error and
// is not recorded.
func readLibraryDeclaration(src []byte, res *resolver) moduleFacts {
	facts := readDeclaration(src, res)
	facts.operators = slices.DeleteFunc(facts.operators, func(op operatorDecl) bool {
		if !coreOnlyClasses[op.class] {
			return false
		}
		facts.errs = append(facts.errs, fmt.Errorf("sub %s: operator class %s is CORE.pmt's; XS::Parse::Infix registers no operator at its level", op.name, op.class))
		return true
	})
	return facts
}

// coreTable is perl's builtins by name, each to its prototype without
// parentheses: `bless` to `$;$`. It is built by parsing declarations/CORE.pmt,
// the declaration file for the language itself -- TypeScript's lib.d.ts to a
// module declaration's @types/Foo.
//
// Built on first use rather than at package init: the parse that builds it is
// the parser that consults it. CORE.pmt aliases nothing, so building it never
// asks for it.
func coreTable() map[string]string {
	readCore()
	return coreMap
}

// coreSignatures is perl's builtins by name, each to the typed signatures
// CORE.pmt states or derives for it -- print's too, which has no prototype
// and so no place in coreTable.
func coreSignatures() map[string][]types.Signature {
	readCore()
	return coreSigs
}

// CoreBuiltin is the typed signatures CORE.pmt states or derives for the
// builtin name, one per candidate, and nil for a name it does not declare.
// internal/infer types builtin calls from it.
func CoreBuiltin(name string) []types.Signature {
	return coreSignatures()[name]
}

// CoreOperator is the typed signatures CORE.pmt declares for the operator
// spelled op of fixity "infix" or "prefix", one per candidate, and nil for
// one it does not declare. internal/infer types operators from it.
func CoreOperator(op, fixity string) []types.Signature {
	readCore()
	return coreOps[[2]string{op, fixity}]
}

// coreShapes is perl's builtins by name, each to the shape a call to it
// parses in: named unary, list operator or niladic (deriveShapes).
func coreShapes() map[string]Shape {
	readCore()
	return coreShp
}

// deriveShapes is the parse shape of each builtin CORE.pmt declares. A
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

// readCore reads CORE.pmt, once, into coreTable's, coreSignatures's,
// CoreOperator's and coreShapes's tables.
func readCore() {
	coreOnce.Do(func() {
		src, ok := declaration("CORE")
		if !ok {
			panic("parse: declarations/CORE.pmt is not embedded")
		}
		var err error
		if coreMap, coreSigs, coreOps, err = coreDeclarations(src); err != nil {
			panic("parse: declarations/CORE.pmt: " + err.Error())
		}
		// Its operators' relations derive one precedence order (RFC 0001,
		// "Precedence is a relation between operators"). A library's
		// operators relate to CORE.pmt's, so the order is CORE.pmt's.
		if _, err := precedenceOrder(coreOperatorDecls(src)); err != nil {
			panic("parse: declarations/CORE.pmt: " + err.Error())
		}
		coreShp = deriveShapes(coreMap, coreSigs)
	})
}

// coreOperatorDecls is the operators a CORE.pmt declares.
func coreOperatorDecls(src []byte) []operatorDecl {
	p := newParser(src, nil, true)
	p.buildingCore = true
	return readDeclarationWith(p).operators
}

// coreProtos reads a CORE.pmt into the tables coreTable and coreSignatures
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
	protos := map[string]string{}
	for name, proto := range facts.protos {
		// A builtin with no prototype, print's "", has no entry: aliasTarget
		// reads presence here as a known prototype.
		if proto == "" {
			continue
		}
		protos[name] = strings.TrimSuffix(strings.TrimPrefix(proto, "("), ")")
	}
	return protos, facts.signatures, ops, nil
}

var (
	coreOnce sync.Once
	coreMap  map[string]string
	coreSigs map[string][]types.Signature
	coreOps  map[[2]string][]types.Signature
	coreShp  map[string]Shape
)
