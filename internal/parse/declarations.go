// ABOUTME: Module declaration files: what a module defines, stated where its source cannot show it.
// ABOUTME: Embedded, and read in place of the module's own source.
package parse

import (
	"embed"
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
		for name, proto := range readModule(Parse(src)).protos {
			coreMap[name] = strings.TrimSuffix(strings.TrimPrefix(proto, "("), ")")
		}
	})
	return coreMap
}

var (
	coreOnce sync.Once
	coreMap  map[string]string
)
