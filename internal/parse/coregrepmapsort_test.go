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
// parameter's sigil, the invocant's type, and the contexts answered for.
type candidate struct {
	Row      measuredRow
	Sigils   string
	Invocant types.Type
	Context  types.Contexts
}

// candidateOf reads a declaration in candidate's terms.
func candidateOf(sig types.Signature) candidate {
	c := candidate{Row: rowOf(sig), Context: sig.Context}
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
// @x`, each with its list written; scalar sort is undefined, so each
// answers for list context alone.
func TestCoreGrepMapSortAreMultis(t *testing.T) {
	list := types.ContextSet(types.ListCtx)
	block := candidate{measuredRow{2, []types.Type{types.Code, types.List}, types.List}, "&@", types.Unknown, types.EveryContext}
	expr := candidate{measuredRow{1, []types.Type{types.Scalar, types.List}, types.List}, "$@", types.Unknown, types.EveryContext}
	core := parse.CoreSignatures()
	for name, want := range map[string][]candidate{
		"grep": {block, expr},
		"map":  {block, expr},
		"sort": {
			{measuredRow{2, []types.Type{types.Code, types.List}, types.List}, "&@", types.Unknown, list},
			{measuredRow{1, []types.Type{types.List}, types.List}, "@", types.Unknown, list},
			{measuredRow{1, []types.Type{types.List}, types.List}, "@", types.Code | types.Str, list},
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

// TestCoreSortScalarHasNoType: every sort candidate answers for list
// context alone, so a call in scalar context selects none and has no type.
// Measured on 5.42, `my $n = sort 3,1,2` leaves $n undef. The same call
// in list context selects the plain candidate.
func TestCoreSortScalarHasNoType(t *testing.T) {
	sort := parse.CoreSignatures()["sort"]
	if len(sort) != 3 {
		t.Fatalf("sort: CORE.pmt declares %d candidates; want 3", len(sort))
	}
	args := []types.Type{types.Int, types.Int, types.Int}
	if sel := types.Select(sort, args, types.ScalarCtx); sel.Outcome != types.Failed || sel.Returns != types.Unknown {
		t.Errorf("scalar sort selects %+v; want no candidate and no type", sel)
	}
	if sel := types.Select(sort, args, types.ListCtx); sel.Outcome != types.Selected || sel.Candidate != 1 {
		t.Errorf("list sort selects %+v; want the plain candidate, 1", sel)
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
