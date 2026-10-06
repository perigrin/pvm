// ABOUTME: Tests for the invocant colon in .pmt signatures, `sub exec (Str $program: List[Str] @args) Boolean;`.
// ABOUTME: The leading slot perl takes with no comma after it: print's filehandle, exec's program.
package parse

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/types"
)

// TestPmtInvocantColonParses: RFC 0001 "Builtins that keep their own
// parse". A leading slot taken with no comma after it is written with
// Raku's invocant colon, and is recorded as the signature's invocant,
// apart from its positional parameters. The slot may have a default.
func TestPmtInvocantColonParses(t *testing.T) {
	facts := readDeclaration([]byte("sub exec (Str $program: List[Str] @args) Boolean;\n"+
		"sub print (FileHandle $fh = select(): List[Str] @args = ($_)) Boolean;\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	args := types.Param{Name: "args", Sigil: '@', Type: types.List, Element: types.Str}
	for name, want := range map[string]types.Signature{
		"exec": {
			Invocant: &types.Param{Name: "program", Sigil: '$', Type: types.Str, Required: true},
			Params:   []types.Param{args},
			Returns:  types.Boolean,
		},
		"print": {
			Invocant: &types.Param{Name: "fh", Sigil: '$', Type: types.FileHandle, Default: "select()"},
			Params:   []types.Param{{Name: "args", Sigil: '@', Type: types.List, Element: types.Str, Default: "($_)"}},
			Returns:  types.Boolean,
		},
	} {
		if got := facts.signatures[name]; !reflect.DeepEqual(got, []types.Signature{want}) {
			t.Errorf("%s: got %+v, want %+v", name, got, want)
		}
	}
}

// TestInvocantColonDerivesNoPrototype: a signature with an invocant colon
// derives no prototype, whatever follows the colon, matching perl: measured
// on 5.42.0, `prototype("CORE::$_")` is undef for print, printf, say, exec
// and system. No prototype is recorded as perl records a sub with none, the
// empty string.
func TestInvocantColonDerivesNoPrototype(t *testing.T) {
	facts := readDeclaration([]byte("sub exec (Str $program: List[Str] @args) Boolean;\nsub f (Str $p: Str $x);\n"), nil)
	if len(facts.errs) > 0 {
		t.Fatalf("errors: %v", facts.errs)
	}
	for _, name := range []string{"exec", "f"} {
		if proto, ok := facts.protos[name]; !ok || proto != "" {
			t.Errorf("%s: prototype %q (declared %v), want none", name, proto, ok)
		}
		if len(facts.signatures[name]) != 1 {
			t.Errorf("%s: signatures %+v, want one", name, facts.signatures[name])
		}
	}
}
