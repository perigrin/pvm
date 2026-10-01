// ABOUTME: Module declaration files: what a module defines, stated where its source cannot show it.
// ABOUTME: Embedded, and read in place of the module's own source.
package parse

import (
	"embed"
	"path"
	"strings"
)

// declarations holds one file per declared module, `Moose::Role` at
// `declarations/Moose/Role.pmd`. TypeScript's `@types/Foo` is the model: an
// interface ASSERTED over code this parser never reads, for a module whose
// own `import` builds what it defines -- Moose::Exporter, a monkey_patch
// loop -- so that no reading of its source can recover it.
//
// A declaration is written in the module syntax the resolver already reads:
// `our @EXPORT` for what a bare `use` imports, `sub NAME (PROTO);` for each
// sub and its prototype. So it is read by the same parse as a module.
//
//go:embed declarations
var declarations embed.FS

// declaration returns the declaration file for module, if there is one.
func declaration(module string) ([]byte, bool) {
	src, err := declarations.ReadFile(path.Join("declarations", strings.ReplaceAll(module, "::", "/")+".pmd"))
	return src, err == nil
}
