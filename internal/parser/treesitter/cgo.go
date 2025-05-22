// ABOUTME: CGO configuration for tree-sitter-perl integration
// ABOUTME: Sets up include paths and library linking for PSC parser

package treesitter

/*
#cgo CFLAGS: -I${SRCDIR}/../../../include
#cgo LDFLAGS: -L${SRCDIR}/../../../lib
*/
import "C"
