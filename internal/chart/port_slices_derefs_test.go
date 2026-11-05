// ABOUTME: Tests the slices, glob elements, postfix dereference and sigil casts.
// ABOUTME: Snippets under testdata/port/slices-derefs are recognized; slices-derefs-refused ones refuse.

package chart

import (
	"os"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// slicesDerefsLines are the perly.y alternatives this family ports or fits. None is left for a completing issue.
// The sigil casts add fitted rules citing the variable heads (1846-1868), whose marks LEX set.
var slicesDerefsLines = []int{1349, 1594, 1604, 1614, 1624, 1650, 1652, 1654, 1656, 1659, 1864, 1872, 1873, 1877, 1878, 1882, 1883}

func TestPortSlicesAndDerefs(t *testing.T) {
	if err := grammar.CheckFamilyMarked(slicesDerefsLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "slices-derefs") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

func TestPortSlicesAndDerefsRefusals(t *testing.T) {
	for _, f := range snippets(t, "slices-derefs-refused") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); v.accepted {
			t.Errorf("%s is accepted; perl refuses it", f)
		}
	}
	// The bare statement forms beside the refusals, which have no refused prefix of their own.
	for _, src := range []string{"@a[1, 2];", "@h{qw(a b)};", "%h{'a'};", "$#a;", "*g{CODE};", "${ $x };", "$$x;", "$r->@*;", "$r->$#*;",
		"@$x;", "@{$x};", "&$$x;", "%$x;", "$r->%*;", "$r->&*;", "$r->**;", "$r->$*;"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d, expecting %v", src, v.at, v.expected)
		}
	}
}

// A cast's sigil (pvm's DerefSigil) and a star after -> map to the fitted cast terminals, and a whole variable never does:
// sharing PERLY_DOLLAR with `$x` would let `$x $z` read as a cast (perigrin's ruling "Fitted cast terminal").
func TestPortSlicesAndDerefsCastTerminals(t *testing.T) {
	for _, tc := range []struct {
		src, at, want string
	}{
		{"$y = $$x;", "$$", "CAST_DOLLAR"},
		{"$y = @$x;", "@$", "CAST_SNAIL"},
		{"$y = %$x;", "%$", "CAST_PERCENT_SIGN"},
		{"$y = $r->$*;", "$*", "CAST_STAR"},
		{"$y = $r->@*;", "@*", "CAST_SNAIL"},
		{"$y = $r->%*;", "%*", "CAST_PERCENT_SIGN"},
		{"$y = $r->**;", "**", "CAST_STAR"},
		{"$y = $r->&*;", "&*", "PERLY_AMPERSAND"},
		{"$y = $r->$#*;", "$#*", "DOLSHARP"},
		{"$y = $r->@[0];", "@[", "PERLY_SNAIL"},
		{"$y = $x;", "$x", "PERLY_DOLLAR"},
		{"$y = @*;", "@*", "PERLY_SNAIL"},
	} {
		if got := terminalsAt(t, tc.src, strings.Index(tc.src, tc.at)); !slices.Equal(got, []string{tc.want}) {
			t.Errorf("%q at %q admits %v, want %s alone", tc.src, tc.at, got, tc.want)
		}
	}
}

// Known limits, pinned so they are noticed when they end. toke.c's scan_ident reads `${ {` (any sigil's, after -> too)
// as the punctuation variable named `{`, so `${ {} }` leaves an unmatched `}` (perl: "Unmatched right curly bracket"); internal/lexer emits a
// DerefSigil and two `{` instead, and the chart reads a block holding an anonymous hash. This fails when the lexer
// emits toke.c's token; drop the README line then.
func TestPortSlicesAndDerefsKnownLimits(t *testing.T) {
	for _, src := range []string{"$y = ${ {} };", "$y = ${{}};", "$y = @{ {} };", "$y = %{ {} };", "$y = &{ {} };", "$y = *{ {} };",
		"$y = $x->${ {} };", "$y = ${ {} }{a};", "print ${ {} };"} {
		if !recognizeSource(t, []byte(src)).accepted {
			t.Errorf("%q is refused; the known limit is gone, so drop the README line", src)
		}
	}
}
