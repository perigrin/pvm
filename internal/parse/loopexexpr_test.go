// ABOUTME: `last LABEL`, `next LABEL` and `redo LABEL` inside an expression take their
// ABOUTME: label as they do at a statement's start: `$p = $b and last BIN if $b;`.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestLoopControlInExpression: measured on 5.42.0 with -MO=Deparse,
//
//	$p = $b and last B if $b;   $p = $b and last B if $b;
//	$q = 1 or next B;           next B unless $q = 1;
//	$r ||= redo B;              $r ||= (redo B);
//
// PerlOnJava unit/loop_label_bareword_constant.t:13.
func TestLoopControlInExpression(t *testing.T) {
	for _, src := range []string{
		`B: for my $b (0, 1) { $p = $b and last B if $b; }`,
		`B: for my $b (0, 1) { $q = 1 or next B; }`,
		`B: for my $b (0, 1) { $r ||= redo B; }`,
		`B: for my $b (0, 1) { $s = 2 and last; }`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		if !containsKind(root, parse.LoopControl) {
			t.Errorf("%q: no LoopControl; got %s", src, shape(root))
		}
		canon := parse.Canon(root, []byte(src))
		if again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon)); again != canon {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
}
