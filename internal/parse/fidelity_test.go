// ABOUTME: Fidelity compares the emission against the SOURCE, quotienting one known difference.
// ABOUTME: Stability compares canon against itself and cannot see a tree that misread its source.

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCanonFidelityCatchesMisgrouping: the fidelity comparison fails on a
// tree that is not a parse of its source.
//
// This is the whole reason the check exists. The stability sweep
// (canonMatches) compares canon's output against itself, so the source is
// gone after the first emission and a consistently wrong tree is still a
// fixpoint -- measured:
//
//	WRONG tree: once="print((1 + 2) * 3);"
//	            twice="print((1 + 2) * 3);"   fixpoint
//
// Fidelity puts the source back on one side. The wrong tree emits a paren
// the source does not contain, and no quotient may forgive that.
func TestCanonFidelityCatchesMisgrouping(t *testing.T) {
	src := []byte("print (1+2)*3;")

	// The wrong tree: print swallowing the whole product. Same leaves, same
	// order -- round-trip and the stability sweep both accept it.
	mul := &parse.Node{Kind: parse.Binary, Text: "*", Start: 6, End: 13, Children: []*parse.Node{
		{Kind: parse.Binary, Text: "+", Start: 7, End: 10, Children: []*parse.Node{
			{Kind: parse.Term, Text: "1", Start: 7, End: 8},
			{Kind: parse.Term, Text: "2", Start: 9, End: 10},
		}},
		{Kind: parse.Term, Text: "3", Start: 12, End: 13},
	}}
	wrong := &parse.Node{Kind: parse.SourceFile, Start: 0, End: len(src), Children: []*parse.Node{
		{Kind: parse.Statement, Start: 0, End: len(src), Children: []*parse.Node{
			{Kind: parse.Call, Text: "print", Resolved: true, Start: 0, End: 13, Children: []*parse.Node{mul}},
		}},
	}}

	if ok, _ := parse.Faithful(wrong, src); ok {
		t.Errorf("Faithful accepted a tree that is not a parse of its source; "+
			"it emitted %q from %q", parse.Canon(wrong, src), string(src))
	}

	// And the correct parse of the same source must be accepted, or the
	// check is just "always false".
	if ok, why := parse.Faithful(parse.Parse(src), src); !ok {
		t.Errorf("Faithful rejected the correct parse of %q: %s", string(src), why)
	}
}

// TestCanonQuotientIsNarrow: only the parenless-call paren is forgiven.
//
// That equivalence is real and unavoidable -- `print "x"` and `print("x")`
// parse to the SAME TREE, correctly, so a tree-faithful emitter must print
// one form for both. 519 of 658 disagreeing T1 files differed by exactly
// this and nothing else.
//
// Every other difference must still fail, or the quotient has swallowed the
// property it was carved out of.
func TestCanonQuotientIsNarrow(t *testing.T) {
	// Forgiven: the call's own parens, present or absent in the source.
	for _, src := range []string{
		`print "x";`,
		`print("x");`,
		`ok $x, 'n';`,
		`ok($x, 'n');`,
	} {
		if ok, why := parse.Faithful(parse.Parse([]byte(src)), []byte(src)); !ok {
			t.Errorf("Faithful rejected %q, but a call's parens are the one "+
				"difference it must forgive: %s", src, why)
		}
	}

	// NOT forgiven: a paren that changes grouping. These sources differ from
	// each other only by parens, and each must be accepted against ITSELF
	// and rejected against the other's tree.
	for _, tc := range []struct{ mine, theirs string }{
		{`(1 + 2) * 3;`, `1 + 2 * 3;`},
		{`1 - (2 - 3);`, `1 - 2 - 3;`},
		{`(2 ** 3) ** 2;`, `2 ** 3 ** 2;`},
	} {
		if ok, why := parse.Faithful(parse.Parse([]byte(tc.mine)), []byte(tc.mine)); !ok {
			t.Errorf("Faithful rejected %q against its own parse: %s", tc.mine, why)
		}
		// The other source's tree against this source: a grouping paren is
		// missing or extra, which must not be forgiven.
		if ok, _ := parse.Faithful(parse.Parse([]byte(tc.theirs)), []byte(tc.mine)); ok {
			t.Errorf("Faithful accepted the tree of %q against the source %q; "+
				"those group differently", tc.theirs, tc.mine)
		}
	}
}

// TestCanonFidelityRatchet: the per-file fidelity result, failing in either
// direction.
//
// This is the criterion 01a0a8d8 was written for and could not carry: its
// corpus sweep compares canon against ITSELF, so a tree that misread its
// source passes. Here the source is on one side, and the only forgiven
// difference is the paren a parenless call must acquire.
//
// The count is 1 for a file whose emission does not say what its source
// says, 0 for one that does -- so "clean" in the header is the fidelity
// rate.
func TestCanonFidelityRatchet(t *testing.T) {
	dir, files := t1Files(t)

	now := make(map[string]int, len(files))
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			continue
		}
		now[rel] = 0
		if ok, _ := parse.Faithful(parse.Parse(src), src); !ok {
			now[rel] = 1
		}
	}

	checkRatchet(t, filepath.Join("testdata", "t1fidelity.ratchet"),
		"Files whose canon emission does NOT say what the source says (1 = differs).\n"+t1CorpusNote, now)
}
