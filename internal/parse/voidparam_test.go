// ABOUTME: Tests that a .pmt parameter whose type includes Void may be absent, and derives a prototype `;` before it.
// ABOUTME: Any is the permissive top, not a union with Void, so an Any parameter stays required.
package parse

import "testing"

// TestVoidParamIsOptional: RFC 0001's `Optional[T]` is `Void|T`, an
// argument that may be absent with no default. Measured on 5.42,
// `srand()` seeds itself where `srand(undef)` warns and seeds from undef,
// and perl's prototype for srand is `;$`. So a parameter whose type
// includes Void is optional, and the table puts a `;` before it. `List`
// includes Void, so `(List $x)` derives `;+`: under `(;+)` perl's `g()`
// passes nothing. Any is the permissive top, as TypeScript's `any` is
// beside its `void`, and an Any parameter is required: `(Any $x)` is `$`.
// A slurpy takes zero arguments already, so `(Void @x)` stays `@`.
func TestVoidParamIsOptional(t *testing.T) {
	for typed, want := range map[string]string{
		"(Void|Int $seed)":      ";$",
		"(Optional[Int] $seed)": ";$",
		"(Void $x)":             ";$",
		"(Glob *socket, Str $msg, Int $flags, Optional[Str] $to)": "*$$;$",
		"(Void @x)":       "@",
		"(List $x)":       ";+",
		"(Any $x)":        "$",
		"(Int $x)":        "$",
		"(Maybe[Int] $x)": "$",
	} {
		if got := prototypeFromTypes(typedSignature(t, "sub f "+typed+";\n")); got != want {
			t.Errorf("%s: got %q, want %q", typed, got, want)
		}
	}
	// A required parameter after one that may be absent has no prototype,
	// and a Void parameter cannot also be required by `= die`.
	pmtRefuses(t, map[string]string{
		"sub f (Optional[Int] $x, Int $y);\n": "sub f: mandatory parameter $y follows optional parameter $x",
		"sub f (Void|Int $x, Str $y);\n":      "sub f: mandatory parameter $y follows optional parameter $x",
		"sub f (Optional[Int] $x = die);\n":   "sub f: $x has type Optional[Int], which may be absent, and defaults to die, which requires it",
	})
	// `\`'s two candidates, `(List @l)` and `(Any $x)`, each take one
	// operand: the Any one is not optional, so they are not ambiguous.
	if got := len(CoreOperator(`\`, "prefix")); got != 2 {
		t.Errorf(`\ has %d prefix candidates, want 2`, got)
	}
	// srand() is the Void case and srand(1) the Int case; srand(1, 2) is
	// too many arguments (measured with perl -c on 5.42).
	for _, src := range []string{"srand();\n", "srand(1);\n", "my @e; srand(@e);\n"} {
		if got := refusedAs(src); len(got) != 0 {
			t.Errorf("%q: refusals %v, want none", src, got)
		}
	}
	wantArityRefusal(t, "srand(1, 2);\n")
}
