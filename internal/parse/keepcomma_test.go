// ABOUTME: Canon keeps a trailing comma where the source wrote one, so the emission
// ABOUTME: says what the source says: `f(1, 2,)` stays `f(1, 2,)`.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCanonKeepsTrailingComma: perl drops a trailing separator -- measured on
// 5.42.0, `f(1, 2,)` deparses as `f(1, 2)` -- and canon used to drop it too,
// which Faithful counts as a misread. By perigrin's decision (2026-09-30)
// canon writes it where the source did, so the emission is faithful and a
// fixpoint. Issue 01a0ef41's cases.
func TestCanonKeepsTrailingComma(t *testing.T) {
	for _, src := range []string{
		`f(1, 2,);`,
		`my @a = (1, 2,);`,
		`sub f {} f 1, 2,;`,
		`print 1, if $x;`,
		`sub f {} f 1, || 2;`,
		`sub f {} f 1 => ->m;`,
		`my %h = (a => 1, b => 2,);`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		canon := parse.Canon(root, []byte(src))
		if ok, why := parse.Faithful(root, []byte(src)); !ok {
			t.Errorf("%q: canon %q is not faithful: %s", src, canon, why)
		}
		if again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon)); again != canon {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", src, canon, again)
		}
	}
}
