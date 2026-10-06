// ABOUTME: Tests CORE.pmt's typing of perl's array, list and hash builtins.
// ABOUTME: Each type is the one measured on perl 5.42, and keys and values fork on context.
package parse_test

import (
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
