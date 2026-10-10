// ABOUTME: Tests CORE.pmt's grep, map and sort: the multis RFC 0001 declares for the builtins that keep their own parse.
// ABOUTME: Their candidates, measured rows and selection are held here, each measured on perl 5.42.
package parse_test

import (
	"fmt"
	"os"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// grepMapSort are the three builtins RFC 0001 "Builtins that keep their
// own parse" declares as multis. perl 5.42 reports undef for each one's
// prototype.
var grepMapSort = []string{"grep", "map", "sort"}

// candidate is a measuredRow beside what a row does not say: each
// parameter's sigil and the invocant's type.
type candidate struct {
	Row      measuredRow
	Sigils   string
	Invocant types.Type
}

// candidateOf reads a declaration in candidate's terms.
func candidateOf(sig types.Signature) candidate {
	c := candidate{Row: rowOf(sig)}
	for _, p := range sig.Params {
		c.Sigils += string(p.Sigil)
	}
	if sig.Invocant != nil {
		c.Invocant = sig.Invocant.Type
	}
	return c
}

// TestCoreGrepMapSortAreMultis: RFC 0001 "Builtins that keep their own
// parse". grep and map are each a block candidate, `(Code &block, List
// @list = die)` as perlsub's `sub mygrep (&@)` but with its list written,
// and an expression candidate, `(Scalar $expr, List @list)`, the `($@)`
// shape B::Deparse gives `grep /a/ || 1, @l`. sort is a block candidate,
// a plain one and an invocant candidate for `sort byname @x` and `sort $n
// @x`, each with its list written, and each again returning Undef, the
// scalar-context sort (TestCoreSortScalarIsUndef).
func TestCoreGrepMapSortAreMultis(t *testing.T) {
	block := candidate{measuredRow{2, []types.Type{types.Code, types.List}, types.List}, "&@", types.Unknown}
	expr := candidate{measuredRow{1, []types.Type{types.Scalar, types.List}, types.List}, "$@", types.Unknown}
	core := parse.CoreSignatures()
	for name, want := range map[string][]candidate{
		"grep": {block, expr},
		"map":  {block, expr},
		"sort": {
			{measuredRow{2, []types.Type{types.Code, types.List}, types.List}, "&@", types.Unknown},
			{measuredRow{1, []types.Type{types.List}, types.List}, "@", types.Unknown},
			{measuredRow{1, []types.Type{types.List}, types.List}, "@", types.Code | types.Str},
			{measuredRow{2, []types.Type{types.Code, types.List}, types.Undef}, "&@", types.Unknown},
			{measuredRow{1, []types.Type{types.List}, types.Undef}, "@", types.Unknown},
			{measuredRow{1, []types.Type{types.List}, types.Undef}, "@", types.Code | types.Str},
		},
	} {
		var got []candidate
		for _, s := range core[name] {
			got = append(got, candidateOf(s))
		}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Errorf("%s: CORE.pmt has %v; want %v", name, got, want)
		}
	}
}

// TestCoreGrepMapSortOutOfCoreTable: perl reports no prototype for grep,
// map or sort, so CORE.pmt declares them and none enters the prototype
// table: their candidates derive none, grep's and map's `&@` and `$@`
// being inexpressible as one, and sort's invocant deriving none.
func TestCoreGrepMapSortOutOfCoreTable(t *testing.T) {
	perl := perlPrototypes(t)
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	derived, err := parse.DerivedPrototypes(src)
	if err != nil {
		t.Fatal(err)
	}
	core, table := parse.CoreSignatures(), parse.CoreTable()
	for _, name := range grepMapSort {
		if proto, ok := perl[name]; ok {
			t.Errorf("%s: perl reports (%s); measured none", name, proto)
		}
		if len(core[name]) < 2 {
			t.Errorf("%s: CORE.pmt declares %d candidates; want a multi", name, len(core[name]))
		}
		if proto, ok := derived[name]; ok {
			t.Errorf("%s: derives (%s); want none", name, proto)
		}
		if proto, ok := table[name]; ok {
			t.Errorf("%s: in the prototype table as (%s); perl reports none", name, proto)
		}
	}
}

// TestCoreSortScalarIsUndef: measured on 5.42, `my $s = sort @a`, `my $t
// = sort { $a <=> $b } @a`, `my $u = sort by @a` and `my $v = sort $n @a`
// are each undef, with "Useless use of sort in scalar context", and `if
// (sort @a)` is false. So each of sort's list candidates has a scalar
// twin taking the same parameters and returning Undef, which scalar and
// boolean context select (RFC 0001, "Context selects by return type");
// list context still selects the List one.
func TestCoreSortScalarIsUndef(t *testing.T) {
	sort := parse.CoreSignatures()["sort"]
	if len(sort) != 6 {
		t.Fatalf("sort: CORE.pmt declares %d candidates; want 6", len(sort))
	}
	for i, list := range sort[:3] {
		twin := sort[i+3]
		if list.Returns != types.List || twin.Returns != types.Undef ||
			fmt.Sprint(list.Params, list.Invocant) != fmt.Sprint(twin.Params, twin.Invocant) {
			t.Errorf("sort: candidate %d %+v has twin %+v; want the same parameters returning Undef", i, list, twin)
		}
	}
	for _, args := range [][]types.Type{{types.Int, types.Int, types.Int}, {types.Code, types.Int, types.Int}} {
		// BooleanCtx is ScalarCtx, so scalar context covers it.
		for ctx, want := range map[types.Context]types.Type{types.ScalarCtx: types.Undef, types.ListCtx: types.List} {
			if sel := types.Select(sort, args, ctx); sel.Outcome != types.Selected || sel.Returns != want {
				t.Errorf("sort%v in %v selects %+v; want %v", args, ctx, sel, want)
			}
		}
	}
}

// TestCoreGrepMapSortMatchMeasuredSignatures: CORE.pmt carries the
// measured types internal/types/signatures.go gives grep, map and sort,
// each row copied here so the check outlives that file, and compared with
// the block candidate for grep and map and the plain one for sort. The
// block candidate's arity of two is perl's: after a block the list must be
// written, `grep {1};` being a syntax error and `grep {1} ()` compiling.
func TestCoreGrepMapSortMatchMeasuredSignatures(t *testing.T) {
	golden := map[string]struct {
		candidate int
		row       measuredRow
	}{
		"map":  {0, measuredRow{2, []types.Type{types.Code, types.List}, types.List}},
		"grep": {0, measuredRow{2, []types.Type{types.Code, types.List}, types.List}},
		"sort": {1, measuredRow{1, []types.Type{types.List}, types.List}},
	}
	core := parse.CoreSignatures()
	for _, name := range grepMapSort {
		want := golden[name]
		if sigs := core[name]; len(sigs) <= want.candidate {
			t.Errorf("%s: CORE.pmt has %d candidates; want candidate %d", name, len(sigs), want.candidate)
		} else if got := rowOf(sigs[want.candidate]); fmt.Sprint(got) != fmt.Sprint(want.row) {
			t.Errorf("%s: CORE.pmt has %v; measured %v", name, got, want.row)
		}
	}
}

// TestCoreSortRequiresAWrittenArgument: RFC 0001 "A required argument
// defaults to `die`". sort's plain candidate is `(List @list = die)`, its
// list recorded as required rather than as a default, so a call that
// writes no argument selects no sort candidate. Measured on 5.42, `sort()`
// and bare `sort` are "Not enough arguments for sort", while `sort(())`
// and `my @e; sort @e` write one, compile, and select the plain candidate.
func TestCoreSortRequiresAWrittenArgument(t *testing.T) {
	sort := parse.CoreSignatures()["sort"]
	if len(sort) != 6 {
		t.Fatalf("sort: CORE.pmt declares %d candidates; want 6", len(sort))
	}
	want := []types.Param{{Name: "list", Sigil: '@', Type: types.List, Required: true}}
	if got := sort[1].Params; fmt.Sprint(got) != fmt.Sprint(want) || sort[1].Invocant != nil {
		t.Errorf("sort's plain candidate is %+v; want %+v", sort[1], want)
	}
	if sel := types.Select(sort, nil, types.ListCtx); sel.Outcome != types.Failed {
		t.Errorf("sort() selects %+v; want no candidate", sel)
	}
	for src, arg := range map[string]types.Type{
		"my @r = sort(());\n":       types.List,
		"my @e; my @r = sort @e;\n": types.Array,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: parses with an Unknown node", src)
		}
		if sel := types.Select(sort, []types.Type{arg}, types.ListCtx); sel.Outcome != types.Selected || sel.Candidate != 1 {
			t.Errorf("%q: selects %+v; want the plain candidate, 1", src, sel)
		}
	}
}
