// ABOUTME: `for my ($a, $b) (LIST)` iterates several variables at once, and the
// ABOUTME: parenthesised variable list belongs to the declaration, not the loop head.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestForLoopManyVariables holds the multi-variable foreach of perl 5.36.
//
// perl 5.42.0 takes it with no feature and no pragma:
//
//	$ perl -e 'for my ($k, $v) (a=>1, b=>2) { print "$k=$v\n" }'
//	a=1
//	b=2
//
// parseFor read a declarator and then asked for a Variable. `(` is not one, so
// the declaration ended empty and the variable list was taken as the loop HEAD
// -- after which the real list `(a=>1, b=>2)` had nowhere to go and the body
// fell away. op/for-many.t is nothing but this construct and carried 81
// Unknowns for it.
//
// Every case asserts the canon and that the canon re-parses to itself, because
// the failure this replaces was a wrong SHAPE, not only a count.
func TestForLoopManyVariables(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		canon string
	}{
		{"two", "for my ($k, $v) (@x) { 1 }", "for my ($k, $v) (@x) {1;}"},
		{"one in parens", "for my ($a) (1, 2) { 1 }", "for my ($a) (1 , 2) {1;}"},
		{"foreach spelling", "foreach my ($a, $b) (@x) { 1 }", "foreach my ($a, $b) (@x) {1;}"},
		// perl accepts empty slots between the variables -- measured,
		// `for my ($a,,, $b) (1,2)` prints 12 -- and op/for-many.t's
		// "comma test" cases spell forty of them.
		{"empty slots", "for my ($a,,, $b) (@x) { 1 }", "for my ($a, $b) (@x) {1;}"},
		// A refalias iterator, under `use feature 'refaliasing'`.
		{"refalias", `foreach my ($k, \@a) (@x) { 1 }`, `foreach my ($k, \@a) (@x) {1;}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := []byte(c.src)
			n := parse.Parse(src)
			if got := countUnknown(n); got != 0 {
				t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
			}
			got := strings.TrimSpace(parse.Canon(n, src))
			if got != c.canon {
				t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
			}
			again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
			if again != got {
				t.Errorf("canon is not a fixpoint:\n  once  %s\n  twice %s", got, again)
			}
		})
	}
}

// TestListEmptySlot holds the list rule the variable list above relies on,
// where it applies to every parenthesised list and not only a loop's.
//
// perl drops a comma that has no element before it, but only once an element
// has been read. Measured on 5.42.0:
//
//	$ perl -e 'my @a = (1,,,2); print scalar(@a)'
//	2
//	$ perl -e 'my @a = (,1);'
//	syntax error at -e line 1, near "(,"
func TestListEmptySlot(t *testing.T) {
	src := []byte("my @a = (1,,,2);")
	n := parse.Parse(src)
	if got := countUnknown(n); got != 0 {
		t.Errorf("Parse(%q) has %d Unknown, want 0", src, got)
	}
	if got, want := strings.TrimSpace(parse.Canon(n, src)), "my @a = (1, 2);"; got != want {
		t.Errorf("Canon(%q):\n  got  %s\n  want %s", src, got, want)
	}

	lead := []byte("my @a = (,1);")
	if got := countUnknown(parse.Parse(lead)); got == 0 {
		t.Errorf("Parse(%q) has no Unknown; a leading comma is a syntax error in perl", lead)
	}
}
