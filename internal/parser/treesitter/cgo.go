// ABOUTME: CGO configuration for tree-sitter-perl integration
// ABOUTME: Sets up include paths and library linking for PSC parser

package treesitter

/*
#cgo CFLAGS: -I${SRCDIR}/../../../include -I${SRCDIR}/../../../lib
#cgo LDFLAGS: -L${SRCDIR}/../../../lib
// Include the generated parser directly
#include "parser.c"
#include "scanner.c"
*/
import "C"
