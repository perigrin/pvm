// ABOUTME: Tests that CORE.pmt types its process, time, object, scoping and miscellaneous builtins.
// ABOUTME: localtime and gmtime fork on context; each type is the one measured on perl 5.42.
package parse_test

import (
	"os"
	"regexp"
	"slices"
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

// TestCoreNiladicBuiltinsTypedEmpty: a builtin perl prototypes `()` --
// `time`, `wait`, `fork` -- is typed with no parameters, never with an
// empty list's `List @l`, which would derive `@`. So its derived
// prototype is the empty one, present in the prototype table, and not
// absent as a builtin with no prototype is.
func TestCoreNiladicBuiltinsTypedEmpty(t *testing.T) {
	perl := perlPrototypes(t)
	core := parse.CoreSignatures()
	table := parse.CoreTable()
	for _, name := range []string{"fork", "time", "wait"} {
		if proto, ok := perl[name]; !ok || proto != "" {
			t.Errorf("%s: perl reports (%s), present %v; measured ()", name, proto, ok)
		}
		sigs := core[name]
		if len(sigs) != 1 || len(sigs[0].Params) != 0 || sigs[0].Returns == types.Unknown {
			t.Errorf("%s: CORE.pmt has %+v; want one typed line with no parameters", name, sigs)
		}
		if proto, ok := table[name]; !ok || proto != "" {
			t.Errorf("%s: the prototype table has (%s), present %v; want ()", name, proto, ok)
		}
	}
}

// globSlotBuiltins are the batch's lines that hold a glob slot, `\[$@%*]`
// and kin, which has no spelling yet (RFC 0001, open question 11,
// perigrin 2026-10-03): they stay prototype-only.
var globSlotBuiltins = []string{"lock", "tie", "tied", "undef", "untie"}

// processTimeMiscBuiltins is perlfunc's "Functions for processes and
// process groups", "Time-related functions", "Keywords related to Perl
// modules", "Keywords related to classes and object-orientation",
// "Keywords related to scoping" and "Miscellaneous functions" among
// CORE.pmt's prototyped lines, with catch, isa and not, the prototyped
// keywords Pod::Functions files under no kind, and without the names the
// earlier batches hold (dbmopen and dbmclose are files) or the glob-slot
// lines.
func processTimeMiscBuiltins(t *testing.T) []string {
	t.Helper()
	others := perlFunctionKinds(t, "String", "Regexp", "Math", "ARRAY", "LIST", "HASH", "I/O", "Binary", "File")
	names := slices.DeleteFunc(perlFunctionKinds(t, "Process", "Time", "Modules", "Objects", "Namespace", "Misc"),
		func(name string) bool {
			return slices.Contains(others, name) || slices.Contains(globSlotBuiltins, name)
		})
	names = append(names, "catch", "isa", "not")
	slices.Sort(names)
	// Measured on 5.42.0: 34 lines, less the five glob-slot ones. An empty
	// or short answer is perl not being asked.
	if len(names) != 29 {
		t.Fatalf("Pod::Functions gives %d prototyped process, time, object, scoping and miscellaneous builtins, %v; measured 29", len(names), names)
	}
	return names
}

// TestCoreProcessTimeMiscBuiltinsTyped: every process, time, module,
// object, scoping and miscellaneous builtin has a typed CORE.pmt line, but
// for the keywordLines, which are no builtins; isa is typed by its infix
// declaration.
func TestCoreProcessTimeMiscBuiltinsTyped(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	names := slices.DeleteFunc(processTimeMiscBuiltins(t),
		func(name string) bool { return keywordLines[name] != "" })
	for _, name := range untypedLines(src, names) {
		if !typedByOperator(name) {
			t.Errorf("%s: CORE.pmt declares it untyped", name)
		}
	}
}

// TestCoreGlobSlotLinesStayUntyped: lock, tie, tied, undef and untie are
// not typed by a guess. Each keeps its prototype-only line, with no
// return type, and the prototype table still has perl's.
func TestCoreGlobSlotLinesStayUntyped(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	if bad := untypedLines(src, globSlotBuiltins); !slices.Equal(bad, globSlotBuiltins) {
		t.Errorf("of %v, CORE.pmt leaves %v untyped; want all of them", globSlotBuiltins, bad)
	}
	perl, table := perlPrototypes(t), parse.CoreTable()
	for _, name := range globSlotBuiltins {
		if table[name] != perl[name] {
			t.Errorf("%s: the prototype table has (%s); perl says (%s)", name, table[name], perl[name])
		}
	}
}

// TestCoreUncallableKeywordsSettled: perigrin, 2026-10-07 -- catch and
// method are keywords, like try and sub, not builtins to type; isa is the
// infix operator, typed by its operator declaration. perl reports `()` for
// all three, so each keeps that line; the every-line check holds catch and
// method as keywords and counts isa typed.
func TestCoreUncallableKeywordsSettled(t *testing.T) {
	for _, k := range []string{"catch", "method"} {
		if keywordLines[k] == "" {
			t.Errorf("%s: not held as a keyword", k)
		}
	}
	if _, exempt := prototypeOnlyLines["isa"]; exempt {
		t.Error("isa is exempt, but its infix declaration types it")
	}
	if ops := parse.CoreOperator("isa", "infix"); len(ops) == 0 || ops[0].Returns != types.Boolean {
		t.Errorf("isa: infix declaration %+v; want one returning Boolean", ops)
	}
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range untypedPrototypedLines(src, perlPrototypes(t)) {
		t.Error(b)
	}
}
