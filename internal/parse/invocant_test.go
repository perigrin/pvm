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

// TestPmtInvocantColonMisplacedRefused: the colon marks only a leading
// slot, so it follows the first parameter and nothing else -- not a later
// one, not a second time, and not an empty slot. The slot holds one item,
// so a List before the colon, by its sigil or its type, is refused too.
func TestPmtInvocantColonMisplacedRefused(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f (Str $a, Str $b: List @l);\n": "sub f: invocant colon after $b; only the first parameter is an invocant",
		"sub f (Str $a: Str $b: List @l);\n": "sub f: invocant colon after $b; only the first parameter is an invocant",
		"sub f (: List @l);\n":               "sub f: invocant colon with no parameter before it",
		"sub f (List @a: List @b);\n":        "sub f: invocant @a is a List; an invocant slot holds one item",
		"sub f (%h: List @b);\n":             "sub f: invocant %h is a List; an invocant slot holds one item",
		"sub f (List $x: List @b);\n":        "sub f: invocant $x is a List; an invocant slot holds one item",
	})
}

// TestInvocantColonWithPrototypeIsError: an invocant colon derives no
// prototype, so any `:prototype(...)` beside one disagrees with its types.
func TestInvocantColonWithPrototypeIsError(t *testing.T) {
	pmtRefuses(t, map[string]string{
		"sub f :prototype($@) (Str $p: List @a);\n": "sub f: :prototype($@) disagrees with its types: an invocant colon derives no prototype",
	})
}

// TestCoreDeclaresInvocantBuiltins: CORE.pmt types the builtins whose
// leading slot takes no comma, each with an invocant. None of them has a
// prototype, so none is in coreTable, which holds only builtins that do:
// aliasTarget reads presence there as a known prototype.
//
// The slot sits outside positional binding, so a call's arguments all
// bind to the list after it: print(Str, Str) and print(ArrayRef) select
// print, where a positional FileHandle would refuse both.
func TestCoreDeclaresInvocantBuiltins(t *testing.T) {
	sigs := coreSignatures()
	for _, name := range []string{"print", "printf", "say", "exec", "system"} {
		if got := sigs[name]; len(got) != 1 || got[0].Invocant == nil {
			t.Errorf("%s: CORE.pmt declares %+v, want one signature with an invocant", name, got)
		}
		if proto, ok := coreTable()[name]; ok {
			t.Errorf("%s: in coreTable with prototype (%s)", name, proto)
		}
	}
	for _, args := range [][]types.Type{{types.Str, types.Str}, {types.ArrayRef}} {
		if got := types.Select(sigs["print"], args, types.ListCtx); got.Outcome != types.Selected || got.Returns != types.Boolean|types.Undef {
			t.Errorf("print%v: got %+v, want print selected, returning Boolean|Undef", args, got)
		}
	}
}

// TestCoreInvocantBuiltinsMatchMeasuredSignatures: CORE.pmt's print and say
// carry the signatures measured for them before CORE.pmt typed them: no
// argument required, each taken as a Str, and a Boolean returned. The rows
// are golden values, copied here, so the declarations answer to them alone.
// The handle is the invocant, not an argument.
func TestCoreInvocantBuiltinsMatchMeasuredSignatures(t *testing.T) {
	type row struct {
		minArity int
		args     []types.Type
		returns  types.Type
	}
	// signatures.go's rows said Boolean; measured on 5.42, a print or say
	// to a closed handle returns undef, so the result is Boolean|Undef.
	measured := row{minArity: 0, args: []types.Type{types.Str}, returns: types.Boolean | types.Undef}
	for _, name := range []string{"print", "say"} {
		sigs := coreSignatures()[name]
		if len(sigs) != 1 {
			t.Errorf("%s: CORE.pmt declares %+v, want one signature", name, sigs)
			continue
		}
		got := row{returns: sigs[0].Returns}
		for _, p := range sigs[0].Params {
			if p.Required {
				got.minArity++
			}
			arg := p.Type
			if p.Element != types.Unknown {
				arg = p.Element
			}
			got.args = append(got.args, arg)
		}
		if !reflect.DeepEqual(got, measured) {
			t.Errorf("%s: CORE.pmt declares %d required, %v -> %v; measured %d required, %v -> %v",
				name, got.minArity, got.args, got.returns, measured.minArity, measured.args, measured.returns)
		}
	}
}

// TestCorePrintFamilyReturnsUndefOnFailure: measured on 5.42, print, printf
// and say return a boolean when the write succeeds and undef when it fails
// (a closed handle), as the I/O batch writes binmode and close's kin.
func TestCorePrintFamilyReturnsUndefOnFailure(t *testing.T) {
	for _, name := range []string{"print", "printf", "say"} {
		sigs := coreSignatures()[name]
		if len(sigs) != 1 || sigs[0].Returns != types.Boolean|types.Undef {
			t.Errorf("%s: CORE.pmt declares %+v; want one signature returning Boolean|Undef", name, sigs)
		}
	}
}
