// ABOUTME: A C-style for head canons as `for (INIT; COND; STEP)`, each of its
// ABOUTME: three slots in place, an empty one included, and settles.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCStyleForCanon: the three clauses of `for (INIT; COND; STEP)` keep their
// slots through canon, and the canon is a fixpoint.
//
// Without slots the clauses were three plain children: the first parenthesised
// as a condition and the rest written bare after it, `for (my $i = 0) $i < 3
// $i++ {}`, and the empty slots of `for (;;)` vanished to `for {}`.
//
// The spelling is Deparse's, measured on 5.42.0 -- clauses joined by `; `,
// an empty slot written as nothing, no space before the `)`:
//
//	for($a;$b;$c){}   for ($a; $b; $c) {
//	for($a;;){}       for ($a; ;) {
//	for(;;$c){}       for (; ; $c) {
//	for($a;$b;){}     for ($a; $b;) {
//	for(;$b;$c){}     for (; $b; $c) {
//	for($a;;$c){}     for ($a; ; $c) {
//
// Deparse rewrites `for(;;)` and `for(;$b;)` as `while` loops; canon keeps
// the `for` the source wrote, in the same spelling.
func TestCStyleForCanon(t *testing.T) {
	for src, want := range map[string]string{
		"for(my $i=0;$i<3;$i++){}": "for (my $i = 0; $i < 3; $i++) {}",
		"for($i=0;$i<3;$i++){1}":   "for ($i = 0; $i < 3; $i++) {1;}",
		"for(;;){}":                "for (; ;) {}",
		"for(0;;){}":               "for (0; ;) {}",
		"for(;0;){}":               "for (; 0;) {}",
		"for(;;0){}":               "for (; ; 0) {}",
		"for(0;0;){}":              "for (0; 0;) {}",
		"for(;0;0){}":              "for (; 0; 0) {}",
		"for(0;;0){}":              "for (0; ; 0) {}",
		"for(0;0;0){}":             "for (0; 0; 0) {}",
		"foreach(;;){}":            "foreach (; ;) {}",
		"L: for(;;){}":             "L: for (; ;) {}",

		// The list forms keep their one parenthesised list, in canon's existing
		// spelling of it.
		"for my $i (1,2){}": "for my $i (1, 2) {}",
		"for (1,2){}":       "for (1, 2) {}",
		"while($x){}":       "while ($x) {}",
	} {
		b := []byte(src)
		root := parse.Parse(b)
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		// A block form's trailing separator is the enclosing file's.
		once := parse.Canon(root, b)
		if strings.TrimSpace(once) != want {
			t.Errorf("%q: canon is %q, want %q", src, once, want)
		}
		twice := parse.Canon(parse.Parse([]byte(once)), []byte(once))
		if once != twice {
			t.Errorf("%q: canon is not a fixpoint:\n once  %q\n twice %q", src, once, twice)
		}
	}
}
