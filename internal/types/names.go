// ABOUTME: Resolves a type's printed name, or an A|B union of names, back to a Type.
// ABOUTME: A .pmt declaration names its types by string; this is the table it resolves against.

package types

import (
	"fmt"
	"strings"
)

// typesByName inverts typeNames: every type the lattice prints by name can be
// written by that name. Unknown is left out -- a declaration that knew
// nothing would say nothing.
var typesByName = func() map[string]Type {
	m := make(map[string]Type, len(typeNames))
	for typ, name := range typeNames {
		if typ != Unknown {
			m[name] = typ
		}
	}
	return m
}()

// FromName resolves a type name, or a union spelled A|B (RFC 0001, "Type
// names"), to its Type. Names are case-sensitive and must be ones the lattice
// prints: "Boolean", not "Bool" or "bool". An unknown or empty member is an
// error that names it.
func FromName(name string) (Type, error) {
	t := None
	for _, part := range strings.Split(name, "|") {
		part = strings.TrimSpace(part)
		member, ok := typesByName[part]
		if !ok {
			if part == "" {
				return Unknown, fmt.Errorf("empty type name %q in %q", part, name)
			}
			return Unknown, fmt.Errorf("unknown type name %q in %q", part, name)
		}
		t = Join(t, member)
	}
	return t, nil
}

// Wrapper reports the type a wrapper name joins with its one parameter,
// the paper's "Absent and undefined": `Maybe[T]` is `Undef|T`, present and
// possibly undef, and `Optional[T]` is `Void|T`, possibly absent. Each
// names a union already in the lattice, not a type of its own.
func Wrapper(name string) (Type, bool) {
	switch name {
	case "Maybe":
		return Undef, true
	case "Optional":
		return Void, true
	}
	return Unknown, false
}
