// ABOUTME: Packages perl defines at interpreter start -- version, PerlIO::Layer --
// ABOUTME: are known to the indirect object reading with nothing loaded.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestInterpreterPackagesAreKnown: `new version "1.2"` and
// `find PerlIO::Layer 'perlio'` are indirect method calls on packages whose
// subs perl registers at startup -- perl.c's boot_core_PerlIO and
// boot_core_UNIVERSAL, which takes vxs.inc's `version` methods. Measured on
// 5.42.0, with nothing loaded:
//
//	$ perl -MO=Deparse -e 'my $v = new version "1.2";
//	      my $l = PerlIO::Layer->find("perlio");'
//	my $v = 'version'->new('1.2');
//
// perl.git t/op/bop.t:574 and t/io/binmode.t:17.
func TestInterpreterPackagesAreKnown(t *testing.T) {
	for _, tc := range []struct{ src, method string }{
		{`my $v = new version "1.2";`, "new"},
		{`my $l = find PerlIO::Layer 'perlio';`, "find"},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		if call := findCall(root, tc.method); call == nil || !call.Indirect {
			t.Errorf("%q: want the indirect call; got %s", tc.src, shape(root))
		}
	}
}
