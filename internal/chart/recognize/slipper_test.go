// ABOUTME: Tests the ruby slipper: the zero-width ; toke.c inserts before } and at end of input, scanned where expected.
// ABOUTME: A toy grammar with ;-terminated statements, blocks and a hash whose } takes no ;; the reference must agree.

package recognize

import (
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// slipperGrammar has statements ending in ;, blocks, and a hash H whose }
// takes no ;. The statement sequence is right-recursive, so Leo applies.
var slipperGrammar = []string{
	"Prog ::= Seq",
	"Seq ::= Stmt Seq |",
	"Stmt ::= X PERLY_SEMICOLON | PERLY_BRACE_OPEN Seq PERLY_BRACE_CLOSE | H PERLY_SEMICOLON",
	"H ::= HASH X PERLY_BRACE_CLOSE",
}

// slip spells the toy's terminals short, for the test tables.
var slip = strings.NewReplacer(";", "PERLY_SEMICOLON", "{", "PERLY_BRACE_OPEN", "}", "PERLY_BRACE_CLOSE")

// TestSlipperToy: a ; is inserted before } and at end of input only where
// the set expects one, Inserted lists those nodes, Expected stays what the
// source may hold, a missing ; before another token is still refused, and a
// } that takes no ; gets no insertion. Under every Options combination.
func TestSlipperToy(t *testing.T) {
	tab := toyTables(t, slipperGrammar...)
	for _, tc := range []struct {
		src      string
		accepted bool
		inserted []int
	}{
		{"X", true, []int{1}},
		{"{ X }", true, []int{2}},
		{"X ; { X }", true, []int{4}},
		{"{ X ; }", true, nil},
		{"HASH X } ;", true, nil},        // the hash's } takes no ;
		{"HASH X }", true, []int{3}},     // only the end of input
		{"{ HASH X } }", true, []int{4}}, // before the block's }, not the hash's
		{"X X", false, nil},              // no ; before another token
		{"{ X X }", false, nil},          // nor inside a block
		{"HASH X ; }", false, nil},       // a written ; where the } takes none
		{"X }", false, []int{1}},         // inserted, but the } closes nothing
		{"", true, nil},                  // nothing expects a ;
		{"{ X } X", true, []int{2, 4}},   // end of input ends the last statement too
		{"{ X } X ;", true, []int{2}},
	} {
		src := slip.Replace(tc.src)
		for _, opts := range optionCombos {
			c := run(t, New(tab, opts), src)
			if c.Accepted() != tc.accepted {
				t.Errorf("%+v: Accepted(%q) = %v, want %v", opts, tc.src, c.Accepted(), tc.accepted)
			}
			if !slices.Equal(c.Inserted(), tc.inserted) {
				t.Errorf("%+v: Inserted(%q) = %v, want %v", opts, tc.src, c.Inserted(), tc.inserted)
			}
		}
	}
	c := run(t, New(tab, Options{}), slip.Replace("{ X }"))
	if exp := expectedOf(t, c, 2); !slices.Equal(exp, []string{"PERLY_SEMICOLON"}) {
		t.Errorf("Expected before the } of `{ X }` = %v, want only the ; the source may hold", exp)
	}
}

// slipperEmptyGrammar has a ; that is a whole statement, and blocks that
// need one, so `{ }` is accepted only by an inserted ; that completes from
// the node it was inserted at.
var slipperEmptyGrammar = []string{
	"Prog ::= Seq |",
	"Seq ::= Stmt Seq | Stmt",
	"Stmt ::= X PERLY_SEMICOLON | PERLY_SEMICOLON | PERLY_BRACE_OPEN Seq PERLY_BRACE_CLOSE",
}

// TestSlipperReferenceAgrees: the recognizer and the reference agree on
// acceptance and Expected at every node, under every Options combination
// (Leo on and off), over every input of up to five of each toy's tokens,
// each combination's recognizer warmed across them.
func TestSlipperReferenceAgrees(t *testing.T) {
	agree(t, toyTables(t, slipperGrammar...), "X", ";", "{", "}", "HASH")
	agree(t, toyTables(t, slipperEmptyGrammar...), "X", ";", "{", "}")
	if !refAccepted(toyTables(t, slipperEmptyGrammar...), slip.Replace("{ }")) {
		t.Error("the reference refuses `{ }`, which only an inserted ; accepts")
	}
}

// agree is TestSlipperReferenceAgrees over one toy grammar and its words.
func agree(t *testing.T, tab *grammar.Tables, words ...string) {
	t.Helper()
	warm := map[Options]*Recognizer{}
	build := func(o Options) *Recognizer {
		if warm[o] == nil {
			warm[o] = New(tab, o)
		}
		return warm[o]
	}
	inputs := [][]string{nil}
	for n := 0; n < len(inputs); n++ {
		if len(inputs[n]) < 5 {
			for _, w := range words {
				inputs = append(inputs, append(slices.Clone(inputs[n]), slip.Replace(w)))
			}
		}
	}
	accepted, inserted := 0, 0
	for _, in := range inputs {
		l := tokens(tab, in...)
		if d := oracleDisagreement(tab, build, l); d != "" {
			t.Fatalf("%q: %s", in, d)
		}
		c := mustRecognize(t, build(Options{}), l)
		if c.Accepted() {
			accepted++
		}
		if len(c.Inserted()) > 0 {
			inserted++
		}
	}
	if accepted == 0 || inserted == 0 {
		t.Errorf("of %d inputs %d were accepted and %d had an insertion, want some of each", len(inputs), accepted, inserted)
	}
}
