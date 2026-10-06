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
