// ABOUTME: The four read-only helpers internal/parse exports for the chart, read here from outside the package.
// ABOUTME: InfixInfo, InterpolatesCode, IsReadlineTerm and HidesStructure, each pinned on a few measured forms.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

func TestExportInfixInfo(t *testing.T) {
	plus, ok := parse.InfixInfo("+")
	if !ok {
		t.Fatal(`InfixInfo("+") is not an operator`)
	}
	times, _ := parse.InfixInfo("*")
	or, _ := parse.InfixInfo("or")
	pow, _ := parse.InfixInfo("**")
	if !(times.BP > plus.BP && plus.BP > or.BP) {
		t.Errorf("binding powers out of order: * %d, + %d, or %d", times.BP, plus.BP, or.BP)
	}
	if pow.Assoc != parse.AssocRight || plus.Assoc != parse.AssocLeft {
		t.Errorf("** is %v and + is %v, want right and left", pow.Assoc, plus.Assoc)
	}
	if dots, _ := parse.InfixInfo(".."); dots.Assoc != parse.AssocNone {
		t.Errorf(".. is %v, want nonassoc", dots.Assoc)
	}
	if _, ok := parse.InfixInfo("frobnicate"); ok {
		t.Error(`InfixInfo("frobnicate") is an operator`)
	}
}

func TestExportInterpolatesCode(t *testing.T) {
	for s, want := range map[string]bool{
		`"$x"`: false, `"@a"`: false, `"${x}"`: false, `"$x->method"`: false,
		`"$h{a}"`: true, `"$r->[0]"`: true, `"@{[ 1 + 2 ]}"`: true, `"${\ f()}"`: true,
	} {
		if got := parse.InterpolatesCode(s); got != want {
			t.Errorf("InterpolatesCode(%s) = %v, want %v", s, got, want)
		}
	}
}

func TestExportIsReadlineTerm(t *testing.T) {
	for s, want := range map[string]bool{
		"<>": true, "<$fh>": true, "<FH>": true, "<STDIN>": true, "<My::Handle>": true, "<hòฟ>": true,
		"<*.c>": false, "<$h{x}>": false, "<a b>": false, "<$>": false,
	} {
		if got := parse.IsReadlineTerm(s); got != want {
			t.Errorf("IsReadlineTerm(%s) = %v, want %v", s, got, want)
		}
	}
}

func TestExportHidesStructure(t *testing.T) {
	for s, want := range map[string]bool{
		`$#{ \@a }`: true, `${ \ f() }`: true, "$x": false, "$#a": false, "$#$r": false,
	} {
		if got := parse.HidesStructure(s); got != want {
			t.Errorf("HidesStructure(%s) = %v, want %v", s, got, want)
		}
	}
}
