// ABOUTME: Tests for deriving a .pmt declaration's prototype from its types and its types from its prototype.
// ABOUTME: Expected values are RFC 0001's table under "A typed signature and a prototype say the same thing".
package parse

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/types"
)

// Each table row's parameter, as a prototype-only declaration derives it.
// A derived parameter has no name: a prototype names none. Measured on
// 5.42.0, `f()` is "Not enough arguments" under `$`, `+`, `&`, `*` and
// `$;$`, and compiles under `_`, `@` and `%`.
var (
	protoScalar   = types.Param{Sigil: '$', Type: types.Scalar, Required: true}
	protoOptional = types.Param{Sigil: '$', Type: types.Scalar}
	protoList     = types.Param{Sigil: '@', Type: types.List}
	protoHash     = types.Param{Sigil: '%', Type: types.List}
	protoPlus     = types.Param{Sigil: '$', Type: types.Array | types.Hash | types.Scalar, Required: true}
	protoCode     = types.Param{Sigil: '&', Type: types.Code, Required: true}
	protoGlob     = types.Param{Sigil: '*', Type: types.Glob, Required: true}
	protoTopic    = types.Param{Sigil: '$', Type: types.Scalar, Default: "$_"}
)

// TestDeriveTypesFromPrototype: RFC 0001, "A typed signature and a
// prototype say the same thing". A declaration with only a prototype gets
// the coarse typed signature the table gives, one case per row that takes
// no backslash: `sub foo :prototype($$)` reads as `(Scalar $, Scalar $)`.
func TestDeriveTypesFromPrototype(t *testing.T) {
	for proto, want := range map[string][]types.Param{
		"$$":  {protoScalar, protoScalar},
		"@":   {protoList},
		"%":   {protoHash},
		"+":   {protoPlus},
		"&@":  {protoCode, protoList},
		"*":   {protoGlob},
		"_":   {protoTopic},
		"$;$": {protoScalar, protoOptional},
	} {
		facts := readDeclaration([]byte("sub f :prototype("+proto+");\n"), nil)
		if len(facts.errs) > 0 {
			t.Errorf("(%s): errors %v", proto, facts.errs)
			continue
		}
		if got := facts.signatures["f"]; !reflect.DeepEqual(got, []types.Signature{{Params: want}}) {
			t.Errorf("(%s): got %+v, want %+v", proto, got, want)
		}
	}
}
