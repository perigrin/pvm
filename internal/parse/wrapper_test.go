// ABOUTME: Tests for the wrapper type names Maybe[T] and Optional[T] in a .pmt declaration.
// ABOUTME: Each names a union, Undef|T and Void|T; neither is a type of its own.
package parse

import (
	"testing"

	"tamarou.com/pvm/internal/types"
)

// TestWrapperTypesAreJoins: the paper's "Absent and undefined". `Maybe[T]`
// is `Undef|T`, present and possibly undef; `Optional[T]` is `Void|T`,
// possibly absent. Measured on 5.42, `srand(undef)` warns "Use of
// uninitialized value" and seeds 18446744073709551615, an Int slot holding
// undef, `Maybe[Int]`; `srand()` warns nothing and seeds itself, a slot
// holding nothing, `Optional[Int]`. They name unions, so each reads as
// exactly its union, prints as it, and carries a container's element as
// the union does. One type parameter, no more and no fewer.
func TestWrapperTypesAreJoins(t *testing.T) {
	sig := typedSignature(t, "sub f (Maybe[Int] $x, Maybe[Int]|Str $y, Optional[List[Str]] @args) Optional[Str];\n")
	if got := sig.Params[0].Type; got != types.Undef|types.Int || got.String() != "Undef|Int" {
		t.Errorf("Maybe[Int] = %v, want Undef|Int", got)
	}
	if got := sig.Params[0].Element; got != types.Unknown {
		t.Errorf("Maybe[Int] has element %v; it is no container", got)
	}
	if got := sig.Params[1].Type; got != types.Undef|types.Int|types.Str {
		t.Errorf("Maybe[Int]|Str = %v, want Undef|Int|Str", got)
	}
	if got := sig.Params[2]; got.Type != types.List || got.Element != types.Str {
		t.Errorf("Optional[List[Str]] = %v of %v, want List of Str", got.Type, got.Element)
	}
	if got := sig.Returns; got != types.Void|types.Str || got.String() != (types.Void|types.Str).String() {
		t.Errorf("Optional[Str] = %v, want Void|Str", got)
	}

	pmtRefuses(t, map[string]string{
		"sub f (Maybe $x);\n":           "sub f: Maybe takes one type, as in Maybe[Int]",
		"sub f (Optional[] $x);\n":      "sub f: Optional takes one type, as in Optional[Int]",
		"sub f (Maybe[Int, Str] $x);\n": "sub f: Maybe takes one type, as in Maybe[Int]",
		"sub f (Str $x) Optional;\n":    "sub f: Optional takes one type, as in Optional[Int]",
		"sub f (Optional[Strng] $x);\n": `sub f: unknown type name "Strng"`,
	})
}
