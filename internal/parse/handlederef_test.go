// ABOUTME: A dereference after `print $fh` starts the printed list, so the scalar
// ABOUTME: before it is the handle: `print $cfh $$code` prints $$code to $cfh.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestHandleBeforeDereference: toke.c takes the scalar after print as a
// handle when a `$` or `@` follows it (yyl_dollar's `memCHRs("$@\"'`q", *s)`),
// and the lexer agrees; the parser's startsTerm did not count the deref
// sigil that `$$code` and `@$lines` begin with. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'print $cfh $$code; print $fh @$lines;
//	      print $fh ${$r}; print $x, $$y;'
//	print $cfh $$code;
//	print $fh @$lines;
//	print $fh ${$r;};
//	print $x, $$y;
//
// perl.git t/porting/cpphdrcheck.t:276 and 340.
func TestHandleBeforeDereference(t *testing.T) {
	for _, src := range []string{
		"print $cfh $$code;",
		"print $fh @$lines;",
		"print $fh ${$r};",
		"print $x, $$y;",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		canon := strings.TrimSpace(parse.Canon(root, []byte(src)))
		again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(canon)), []byte(canon)))
		if canon != again {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
}
