// ABOUTME: Defines Position type for source location tracking within the errors package
// ABOUTME: Duplicated locally to avoid circular imports with the ast package

package errors

// Position represents a position in source code
type Position struct {
	Line   int
	Column int
	Offset int
}
