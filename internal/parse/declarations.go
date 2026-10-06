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
	root, p := parseSource(src, res, true)
	facts := readModule(root)
	facts.signatures, facts.errs, facts.operators = p.signatures, p.typedErrs, p.operators
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

// readCore reads CORE.pmt, once, into coreTable's and coreSignatures's
// tables.
func readCore() {
	coreOnce.Do(func() {
		src, ok := declaration("CORE")
		if !ok {
			panic("parse: declarations/CORE.pmt is not embedded")
		}
		var err error
		if coreMap, coreSigs, err = coreProtos(src); err != nil {
			panic("parse: declarations/CORE.pmt: " + err.Error())
		}
	})
}

// coreProtos reads a CORE.pmt into the tables coreTable and coreSignatures
// hold. A file in error builds neither: a builtin it misdeclares would parse
// wrong.
func coreProtos(src []byte) (map[string]string, map[string][]types.Signature, error) {
	facts := readDeclaration(src, nil)
	if len(facts.errs) > 0 {
		return nil, nil, errors.Join(facts.errs...)
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
	return protos, facts.signatures, nil
}

var (
	coreOnce sync.Once
	coreMap  map[string]string
	coreSigs map[string][]types.Signature
)
