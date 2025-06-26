package tree_sitter_typed_perl

// #cgo CFLAGS: -std=c11 -fPIC -I../../src
// #cgo LDFLAGS: -L../../ -l:libtree-sitter-typed-perl.a
// #include <tree_sitter/parser.h>
// extern const TSLanguage *tree_sitter_typed_perl(void);
import "C"

import "unsafe"

// Get the tree-sitter Language for this grammar.
func Language() unsafe.Pointer {
	return unsafe.Pointer(C.tree_sitter_typed_perl())
}
