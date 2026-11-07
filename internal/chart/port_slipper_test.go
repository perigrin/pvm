// ABOUTME: Tests the committed grammar with the recognizer's ruby slipper: a statement's last ; left off before } or at end of file.
// ABOUTME: Each row's verdict and its nearby perl-accepted form were checked with the pinned perl's -c (94e50866).

package chart

import "testing"

// TestPortSlipper: toke.c inserts the ; before } and at end of file, so
// these are accepted with no rule fitted for it; a missing ; before any
// other token, or a } that closes nothing, is still refused. perl -c on
// 94e50866: every accepted row and every nearby form is "syntax OK"; both
// refused rows fail to compile. `sub f { 1 }` is a sub's block ending without
// its ;, and `{ { 1 } }` a block ending in a block.
// Nor is `use strict $x` a refused row: perl's grammar accepts it as a use
// with an import list (`use lib $x` is "syntax OK"), and only strict's
// import croaks, in BEGIN.
func TestPortSlipper(t *testing.T) {
	for _, src := range []string{"use strict", "{ use strict }", "no warnings", "{ $x = 1 }", "{ { 1 } }", "sub f { 1 }"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d, expecting %v", src, v.at, v.expected)
		}
	}
	for _, tc := range []struct{ refused, nearby string }{
		{"use strict use warnings", "use strict; use warnings"}, // no ; is inserted before a non-} token
		{"use strict }", "{ use strict }"},                      // the } has no matching {
	} {
		if v := recognizeSource(t, []byte(tc.refused)); v.accepted {
			t.Errorf("%q is accepted", tc.refused)
		}
		if v := recognizeSource(t, []byte(tc.nearby)); !v.accepted {
			t.Errorf("%q, near %q, is refused at %d", tc.nearby, tc.refused, v.at)
		}
	}
}
