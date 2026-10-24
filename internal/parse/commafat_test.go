// ABOUTME: A comma directly before a fat comma is an empty slot and is dropped, as
// ABOUTME: two commas in a row are: `tie $@, => 'main', 1` ties with two arguments.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCommaBeforeFatComma: perl drops a separator with no element after it,
// and a fat comma is a separator. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'tie $@, => "main", 1; my @a = (1, => 2);'
//	tie $@, 'main', 1;
//	my(@a) = (1, 2);
//	$ perl -MO=Deparse -e 'sub foo {7} my @x = (foo, => 1);'
//	my(@x) = (foo(), 1);
//
// The fat comma does not reach back past the dropped comma to quote a word.
// perl.git t/op/tie_fetch_count.t:331-335.
func TestCommaBeforeFatComma(t *testing.T) {
	for _, src := range []string{
		`tie $@, => 'main', 1;`,
		`my @a = (1, => 2);`,
		`sub foo {7} my @x = (foo, => 1);`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		canon := parse.Canon(root, []byte(src))
		if again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon)); again != canon {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
	root := parse.Parse([]byte(`sub foo {7} my @x = (foo, => 1);`))
	if call := findCall(root, "foo"); call == nil || call.Fat {
		t.Errorf("`foo, => 1`: foo is a call, not a quoted word; got %s", shape(root))
	}
}

// TestFatCommaQuotesAnOperandsWord is the parser fuzzer's `!y=>0`. A fat
// comma quotes the word directly to its left even when that word is an
// operator's operand: measured on 5.42.0, `sub foo {1} my @a = (!foo=>0)`
// holds "" and 0, `(-foo=>0)` holds "-foo" and 0, and `sub foo {7} (1+foo=>0)`
// holds 1 and 0 -- foo is never called. Read as a call, canon wrote
// `!y() => 0`, and `y()` re-read as a transliteration.
//
// A declaration's initialiser is such an operand too, the fuzzer's
// `my$0=y=>0`: measured on 5.42.0, `sub foo {7} my @a = (my $x = foo=>0)`
// leaves $x "foo" and @a ("foo", 0), and `((my $x = foo)=>0)` calls it.
// So does a word written with its call's parens, which canon writes for the
// grouping it drops: `(my $x = foo()=>0)` leaves $x 7, and
// `(!foo()=>0, 1+foo()=>0)` holds "", 0, 8 and 0.
//
// A comparison's right operand is one, the fuzzer's `y0>y=>0`: measured on
// 5.42.0, Deparse writes `1 > x=>0` as `1 > 'x', '???'` and `1<2<x=>0` as
// `1 < 2 < 'x', '???'`, a chain's last operand quoted as a binary's is.
func TestFatCommaQuotesAnOperandsWord(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"!y=>0", "!y => 0;"},
		{"sub foo {1} my @a = (!foo=>0);", "sub foo {1;} my @a = (!foo => 0);"},
		{"sub foo {7} my @a = (1+foo=>0);", "sub foo {7;} my @a = (1 + foo => 0);"},
		{"sub foo {7} my @a = ((foo)=>0);", "sub foo {7;} my @a = ((foo) => 0);"},
		{"my$x=y=>0", "my $x = y => 0;"},
		{"sub foo {7} my @a = (my $x = foo=>0);", "sub foo {7;} my @a = (my $x = foo => 0);"},
		{"sub foo {7} my @a = ((my $x = foo)=>0);", "sub foo {7;} my @a = ((my $x = foo) => 0);"},
		{"sub foo {7} my @a = (my $x = foo()=>0);", "sub foo {7;} my @a = (my $x = foo() => 0);"},
		{"sub foo {7} my @a = (!foo()=>0, 1+foo()=>0);", "sub foo {7;} my @a = (!foo() => 0, 1 + foo() => 0);"},
		{"y0>y=>0", "y0 > y => 0;"},
		{"1<2<y=>0", "1 < 2 < y => 0;"},
	} {
		got := strings.TrimSpace(parse.Canon(parse.Parse([]byte(tc.src)), []byte(tc.src)))
		if got != tc.want {
			t.Errorf("Canon(parse(%q)):\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
		if again != got {
			t.Errorf("canon is not a fixpoint on %q:\n  once  %q\n  twice %q", tc.src, got, again)
		}
	}
}
