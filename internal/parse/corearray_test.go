// ABOUTME: Tests CORE.pmt's typing of perl's array, list and hash builtins.
// ABOUTME: Each type is the one measured on perl 5.42, and keys and values fork on context.
package parse_test

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// arrayListHashBuiltins is perlfunc's "Functions for real @ARRAYs",
// "Functions for list data" and "Functions for real %HASHes" (Pod::Functions
// kinds ARRAY, LIST and HASH on 5.42), less those with no prototype (grep,
// map, sort, qw, delete, exists) and less each, which the no-prototype
// builtins issue declares as a multi.
var arrayListHashBuiltins = []string{
	"all", "any", "join", "keys", "pop", "push", "reverse",
	"shift", "splice", "unpack", "unshift", "values",
}

// TestCoreArrayListHashBuiltinsTyped: every builtin of the batch has a
// typed CORE.pmt line -- one that states a return type, which a signature
// derived from a prototype never has. all and any take `Code \&block`;
// keys and values are four candidates, two for scalar context and two for
// list context.
func TestCoreArrayListHashBuiltinsTyped(t *testing.T) {
	core := parse.CoreSignatures()
	for _, name := range arrayListHashBuiltins {
		sigs := core[name]
		if len(sigs) == 0 {
			t.Errorf("%s: CORE.pmt declares no signature", name)
		}
		for _, s := range sigs {
			if s.Returns == types.Unknown {
				t.Errorf("%s: CORE.pmt has an untyped line", name)
			}
		}
	}
	for _, name := range []string{"all", "any"} {
		sigs := core[name]
		if len(sigs) != 1 || len(sigs[0].Params) == 0 || !sigs[0].Params[0].Alias ||
			sigs[0].Params[0].Sigil != '&' || sigs[0].Params[0].Type != types.Code {
			t.Errorf("%s: CORE.pmt has %+v; want one line taking Code \\&block first", name, sigs)
		}
	}
	scalar, list := types.ContextSet(types.ScalarCtx), types.ContextSet(types.ListCtx)
	for _, name := range []string{"keys", "values"} {
		counts := map[types.Contexts]int{}
		for _, s := range core[name] {
			counts[s.Context]++
		}
		if len(core[name]) != 4 || counts[scalar] != 2 || counts[list] != 2 {
			t.Errorf("%s: CORE.pmt has %d candidates by context %v; want two :context($) and two :context(@)", name, len(core[name]), counts)
		}
	}
}

// TestCoreKeysValuesMatchMeasuredSignatures: keys' and values' list-context
// candidates together carry the rows internal/types/signatures.go measured,
// copied here so the check outlives that file -- one argument, a hash or an
// array, giving a List. Their scalar-context candidates return the count,
// an Int, and each name's candidates derive perl's `\[%@]`.
func TestCoreKeysValuesMatchMeasuredSignatures(t *testing.T) {
	golden := map[string]measuredRow{
		"keys":   {1, []types.Type{types.Hash | types.Array}, types.List},
		"values": {1, []types.Type{types.Hash | types.Array}, types.List},
	}
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	derived, err := parse.DerivedPrototypes(src)
	if err != nil {
		t.Fatal(err)
	}
	core := parse.CoreSignatures()
	for _, name := range []string{"keys", "values"} {
		// The list-context candidates, one per container, read as the
		// one row signatures.go wrote: their argument types joined.
		var listRow *measuredRow
		for _, s := range core[name] {
			row := rowOf(s)
			switch s.Context {
			case types.ContextSet(types.ScalarCtx):
				if row.Returns != types.Int {
					t.Errorf("%s: a :context($) candidate returns %v; measured Int", name, row.Returns)
				}
			case types.ContextSet(types.ListCtx):
				if listRow == nil {
					listRow = &row
					continue
				}
				if len(row.Args) != len(listRow.Args) || row.MinArity != listRow.MinArity || row.Returns != listRow.Returns {
					t.Errorf("%s: :context(@) candidates %v and %v differ beyond their argument types", name, *listRow, row)
					continue
				}
				for i := range row.Args {
					listRow.Args[i] |= row.Args[i]
				}
			}
		}
		if listRow == nil || fmt.Sprint(*listRow) != fmt.Sprint(golden[name]) {
			t.Errorf("%s: CORE.pmt's :context(@) candidates have %v; measured %v", name, listRow, golden[name])
		}
		if derived[name] != `\[%@]` {
			t.Errorf("%s: candidates derive (%s); perl says (\\[%%@])", name, derived[name])
		}
	}
}

// TestCoreKeysScalarContextIsCount: context selects between keys'
// candidates, RFC 0001 "`:context(...)`". `my $n = keys %h` is the count,
// an Int, and `my @k = keys %h` the list; values, and an array, alike.
func TestCoreKeysScalarContextIsCount(t *testing.T) {
	core := parse.CoreSignatures()
	for _, name := range []string{"keys", "values"} {
		for _, container := range []types.Type{types.Hash, types.Array} {
			for ctx, want := range map[types.Context]types.Type{types.ScalarCtx: types.Int, types.ListCtx: types.List} {
				sel := types.Select(core[name], []types.Type{container}, ctx)
				if sel.Outcome != types.Selected || sel.Returns != want {
					t.Errorf("%s of a %v in context %v: %+v; want the %v candidate selected", name, container, ctx, sel, want)
				}
			}
		}
	}
}

// TestCoreAllAnyDeriveRefAmpAt: all and any are `(Code \&block, List
// @list) Boolean` (perigrin, 2026-10-05; RFC 0001 "The scalar container"),
// whose types derive perl's `\&@`.
func TestCoreAllAnyDeriveRefAmpAt(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	derived, err := parse.DerivedPrototypes(src)
	if err != nil {
		t.Fatal(err)
	}
	want := []types.Param{
		{Name: "block", Sigil: '&', Type: types.Code, Required: true, Alias: true},
		{Name: "list", Sigil: '@', Type: types.List},
	}
	core := parse.CoreSignatures()
	for _, name := range []string{"all", "any"} {
		sigs := core[name]
		if len(sigs) != 1 || !slices.Equal(sigs[0].Params, want) || sigs[0].Returns != types.Boolean {
			t.Errorf("%s: CORE.pmt has %+v; want (Code \\&block, List @list) Boolean", name, sigs)
		}
		if derived[name] != `\&@` {
			t.Errorf("%s: types derive (%s); perl says (\\&@)", name, derived[name])
		}
	}
}

// TestCoreKeysValuesRefuseScalar: perl 5.42 refuses `keys $x` and `values
// $x` ("Experimental keys on scalar is now forbidden"), so a scalar argument
// fits none of their candidates, in either context.
func TestCoreKeysValuesRefuseScalar(t *testing.T) {
	core := parse.CoreSignatures()
	for _, name := range []string{"keys", "values"} {
		for _, ctx := range []types.Context{types.ScalarCtx, types.ListCtx} {
			if sel := types.Select(core[name], []types.Type{types.Scalar}, ctx); sel.Outcome != types.Failed {
				t.Errorf("%s of a Scalar in context %v: %+v; want no candidate", name, ctx, sel)
			}
		}
	}
}

// TestCoreEachDeclaredOnce: this batch leaves each to the no-prototype
// builtins issue, so CORE.pmt declares it once -- one plain line, or a set
// of multi candidates, never both and never two plain lines.
func TestCoreEachDeclaredOnce(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	plain := regexp.MustCompile(`(?m)^sub each\b`).FindAll(src, -1)
	multi := regexp.MustCompile(`(?m)^multi sub each\b`).FindAll(src, -1)
	if len(plain) > 1 || (len(plain) == 1) == (len(multi) > 0) {
		t.Errorf("CORE.pmt declares each in %d plain lines and %d multi candidates; want one declaration set", len(plain), len(multi))
	}
}

// TestCoreReverseIsContextMulti: measured on 5.42, `reverse` in scalar
// context concatenates its list and reverses the string --
// `my $s = reverse "ab", "cd"` is "dcba", POK -- which is not one of the
// list-context values (`("cd", "ab")`). RFC 0001's case for `:context`, so
// a Str candidate for scalar context beside the List one.
func TestCoreReverseIsContextMulti(t *testing.T) {
	core := parse.CoreSignatures()
	for ctx, want := range map[types.Context]types.Type{types.ScalarCtx: types.Str, types.ListCtx: types.List} {
		sel := types.Select(core["reverse"], []types.Type{types.Str, types.Str}, ctx)
		if sel.Outcome != types.Selected || sel.Returns != want {
			t.Errorf("reverse in context %v: %+v; want %v", ctx, sel, want)
		}
	}
}
