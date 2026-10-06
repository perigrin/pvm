// ABOUTME: Module declaration files: what a module defines, stated where its source cannot show it.
// ABOUTME: Embedded, and read in place of the module's own source.
package parse

import (
	"embed"
	"fmt"
	"path"
	"strings"
	"sync"
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
var declarations embed.FS

// declaration returns the declaration file for module, if there is one.
func declaration(module string) ([]byte, bool) {
	src, err := declarations.ReadFile(path.Join("declarations", strings.ReplaceAll(module, "::", "/")+".pmt"))
	return src, err == nil
}

// readDeclaration reads a declaration file. Its language is typed Perl (RFC
// 0001, "Typed Perl, in `.pmt` only"): a module's own source is never read
// this way, so `sub f (Ref $x) { }` there stays perl's prototype.
func readDeclaration(src []byte, res *resolver) moduleFacts {
	root, p := parseSource(src, res, true)
	facts := readModule(root)
	facts.signatures, facts.errs = p.signatures, p.typedErrs
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
	coreOnce.Do(func() {
		src, ok := declaration("CORE")
		if !ok {
			panic("parse: declarations/CORE.pmt is not embedded")
		}
		coreMap = map[string]string{}
		for name, proto := range readDeclaration(src, nil).protos {
			coreMap[name] = strings.TrimSuffix(strings.TrimPrefix(proto, "("), ")")
		}
	})
	return coreMap
}

var (
	coreOnce sync.Once
	coreMap  map[string]string
)
