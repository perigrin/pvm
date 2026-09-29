// ABOUTME: A call with no parens before `->` takes no argument: `shift->m` is `shift()->m`.
// ABOUTME: `->` cannot start a term, so it ends an argument list the way `.` and `==` do.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestArrowEndsArguments holds perl's reading of a parenless call followed by
// an arrow. Measured on 5.42.0, with and without a prototype:
//
//	$ perl -MO=Deparse -e 'sub t (;$) { 1 } my $x = t->m(1);'
//	my $x = t()->m(1);
//
// endsArgumentList lists the infix operators that cannot start a term; `->`
// was missing, so the call took the arrow as its argument -- `shift->m` canon'd
// as `shift(->);m();`. `shift->method` is the idiom every accessor uses.
func TestArrowEndsArguments(t *testing.T) {
	for _, src := range []string{
		"my $x = shift->m;",
		"sub t { 1 } my $x = t->m(1);",
		"sub t (;$) { 1 } my $x = t->m(1);",
	} {
		n := parse.Parse([]byte(src))
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q) has %d Unknown, want 0", src, got)
		}
		got := strings.TrimSpace(parse.Canon(n, []byte(src)))
		if strings.Contains(got, "(->)") {
			t.Errorf("Canon(%q) = %s: the arrow was taken as an argument", src, got)
		}
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
		if again != got {
			t.Errorf("canon is not a fixpoint:\n  once  %s\n  twice %s", got, again)
		}
	}

	// REGRESSION GUARD: a class name before `->` is still a class name.
	src := []byte("my $x = Foo->m;")
	if got := countUnknown(parse.Parse(src)); got != 0 {
		t.Errorf("Parse(%q) has %d Unknown, want 0", src, got)
	}
}
