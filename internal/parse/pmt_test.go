// ABOUTME: Tests for typed Perl in .pmt declaration files: typed parameters and their defaults.
// ABOUTME: Typed Perl is read only from a declaration file; ordinary source keeps perl's reading.
package parse

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/types"
)

// TestPmtTypedParameters: RFC 0001 "Typed Perl, in `.pmt` only". A .pmt
// declaration is `sub NAME :attrs (Type $v = default, ...)`, its prototype
// in `:prototype(...)` as perl requires once a signature is present. Each
// parameter carries the lattice type its name resolves to and the default
// as written; a scalar with no default is required.
func TestPmtTypedParameters(t *testing.T) {
	facts := readDeclaration([]byte("sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__);\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	if got := facts.protos["bless"]; got != "($;$)" {
		t.Errorf("prototype: got %q, want %q", got, "($;$)")
	}
	want := types.Signature{Params: []types.Param{
		{Name: "ref", Sigil: '$', Type: types.Ref, Required: true},
		{Name: "class", Sigil: '$', Type: types.Str, Default: "__PACKAGE__"},
	}}
	if got, ok := facts.signatures["bless"]; !ok || !reflect.DeepEqual(got, want) {
		t.Errorf("signature: got %+v (recorded %v), want %+v", got, ok, want)
	}
}
