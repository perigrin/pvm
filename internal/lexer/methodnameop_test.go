// ABOUTME: A method name after `->` leaves an operator expected: `$o->iters / 2`
// ABOUTME: divides, it does not open a pattern.

package lexer

import "testing"

// TestMethodNameThenOperator: toke.c returns METHCALL0 for a parenless
// method name, and what follows is an operator. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my $r = $o->iters / $o->cpu_p; my $s = $o->n <2;'
//	my $r = $o->iters / $o->cpu_p;
//	my $s = $o->n < 2;
//
// perl.git t/benchmark/gh7094-speed-up-keys-on-empty-hash.t:57.
func TestMethodNameThenOperator(t *testing.T) {
	for _, src := range []string{
		"my $r = $o->iters / $o->cpu_p;",
		"my $s = $o->n <2;",
	} {
		for _, tok := range Tokenize([]byte(src)) {
			switch tok.Kind {
			case Quote, UnknownRest, Readline:
				t.Errorf("%q: after a method name the operator is an operator; got %v %q",
					src, tok.Kind, src[tok.Start:tok.End])
			}
		}
	}
}
