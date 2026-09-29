// ABOUTME: `sort SUBNAME LIST` and `sort $subref LIST`: the comparator is a slot, not sort's only argument.
// ABOUTME: A perl keyword after `sort` is not a comparator -- `sort keys %h` sorts the keys.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestSortComparator holds perl's rule for the word after `sort`, which is
// toke.c's KEY_sort:
//
//	s = force_word(s, BAREWORD, CHECK_KEYWORD | ALLOW_PACKAGE);
//
// A word that is not a perl keyword is forced to a bareword -- the
// comparator's NAME -- and perly.y's `LSTOP indirob listexpr` admits a scalar
// in the same slot. Measured on 5.42.0, both spellings with and without
// parens sort with the comparator:
//
//	sort foo 3,1,2    3,2,1      sort $s 3,1,2     1,2,3
//	sort(foo 3,1,2)   3,2,1      sort($s 3,1,2)    1,2,3
//
// The slot is taken, as the filehandle slot of `print` is, only when a term
// follows with no comma. Without it the comparator became sort's only
// argument and the list a statement of its own, at Unknown=0.
func TestSortComparator(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		canon string
	}{
		{"sub name", "my @b = sort foo 3, 1, 2;", "my @b = sort(foo 3 , 1 , 2);"},
		{"scalar", "my @b = sort $s 3, 1, 2;", "my @b = sort($s 3 , 1 , 2);"},
		{"scalar then array", "my @b = sort $s @h;", "my @b = sort($s @h);"},
		{"parenthesised", "my @b = sort(foo @h);", "my @b = sort(foo @h);"},
		// A comma means the scalar is an ELEMENT: `sort $x, $y` sorts two.
		{"scalar is an element", "my @b = sort $x, $y;", "my @b = sort($x , $y);"},
		// Keywords are not comparators. perl declines to force them, so
		// `sort keys %h` sorts the keys.
		{"keys", "my @b = sort keys %h;", "my @b = sort(keys(%h));"},
		{"map", "my @b = sort map { $_ } @a;", "my @b = sort(map({$_;}@a));"},
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
