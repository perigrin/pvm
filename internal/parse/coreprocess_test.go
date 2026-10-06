// ABOUTME: Tests that CORE.pmt types its process, time, object, scoping and miscellaneous builtins.
// ABOUTME: localtime and gmtime fork on context; each type is the one measured on perl 5.42.
package parse_test

import (
	"os"
	"regexp"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// TestCoreLocaltimeGmtimeAreContextMultis: localtime and gmtime are RFC
// 0001's `:context` example. Measured on 5.42, `my $t = localtime(0)` is
// the date string "Thu Jan  1 00:00:00 1970" (POK), while in list context
// it is nine numbers whose first is 0 -- the string is none of them. So a
// `:context($)` candidate gives the string, undef for a time out of range
// (`localtime(1e20)` is undef), and a `:context(@)` one `List[Int]`; their
// one `Int $time = time` parameter derives perl's `;$`.
func TestCoreLocaltimeGmtimeAreContextMultis(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	derived, err := parse.DerivedPrototypes(src)
	if err != nil {
		t.Fatal(err)
	}
	param := []types.Param{{Name: "time", Sigil: '$', Type: types.Int, Default: "time"}}
	core := parse.CoreSignatures()
	for _, name := range []string{"localtime", "gmtime"} {
		returns := map[types.Contexts]types.Type{}
		for _, s := range core[name] {
			if len(s.Params) != 1 || s.Params[0] != param[0] {
				t.Errorf("%s: a candidate takes %+v; want (Int $time = time)", name, s.Params)
			}
			returns[s.Context] = s.Returns
		}
		want := map[types.Contexts]types.Type{
			types.ContextSet(types.ScalarCtx): types.Str | types.Undef,
			types.ContextSet(types.ListCtx):   types.List,
		}
		if len(core[name]) != 2 || len(returns) != 2 ||
			returns[types.ContextSet(types.ScalarCtx)] != want[types.ContextSet(types.ScalarCtx)] ||
			returns[types.ContextSet(types.ListCtx)] != want[types.ContextSet(types.ListCtx)] {
			t.Errorf("%s: CORE.pmt has %+v; want a :context($) Str|Undef and a :context(@) List[Int] candidate", name, core[name])
		}
		// The lattice keeps a return type's container, not its element, so
		// List[Int] is asked of the line itself.
		list := regexp.MustCompile(`(?m)^multi sub ` + name + ` :context\(@\) \(Int \$time = time\) List\[Int\];$`)
		if len(list.FindAll(src, -1)) != 1 {
			t.Errorf("%s: CORE.pmt has no one :context(@) line returning List[Int]", name)
		}
		if derived[name] != ";$" {
			t.Errorf("%s: candidates derive (%s); perl says (;$)", name, derived[name])
		}
	}
}

// TestCoreLocaltimeVoidHasNoType: neither of localtime's or gmtime's
// candidates answers for void context, so a call there has no type (RFC
// 0001, "`:context(...)`": a call in a context no declaration answers for
// has no type), while scalar and list context each select their own.
func TestCoreLocaltimeVoidHasNoType(t *testing.T) {
	core := parse.CoreSignatures()
	for _, name := range []string{"localtime", "gmtime"} {
		if sel := types.Select(core[name], nil, types.VoidCtx); sel.Outcome != types.Failed || sel.Returns != types.Unknown {
			t.Errorf("%s in void context: %+v; want no candidate and no type", name, sel)
		}
		for _, ctx := range []types.Context{types.ScalarCtx, types.ListCtx} {
			if sel := types.Select(core[name], []types.Type{types.Int}, ctx); sel.Outcome != types.Selected {
				t.Errorf("%s in context %v: %+v; want a candidate selected", name, ctx, sel)
			}
		}
	}
}
