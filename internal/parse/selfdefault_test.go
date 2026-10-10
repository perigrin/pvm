// ABOUTME: Tests that no CORE.pmt parameter defaults to a call of its own sub.
// ABOUTME: An argument perl computes when it is absent is Optional[T], and still derives perl's prototype.
package parse_test

import (
	"os"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCoreHasNoSelfReferentialDefault: what perl does with an absent
// argument, srand's own seed or send's connected peer, is no expression,
// so a default naming the sub's own call states nothing. Such an
// argument is Optional[T], may be absent with no default (RFC 0001,
// "`Maybe[T]` is `Undef|T` and `Optional[T]` is `Void|T`"). The six that
// once wrote one derive perl's prototypes, asked of perl 5.42.
func TestCoreHasNoSelfReferentialDefault(t *testing.T) {
	for name, cands := range parse.CoreSignatures() {
		for _, s := range cands {
			for _, p := range s.Params {
				if p.Default == name || strings.HasPrefix(p.Default, name+"(") {
					t.Errorf("%s: %s defaults to %s, a call of its own sub", name, p.Variable(), p.Default)
				}
			}
		}
	}
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	derived, err := parse.DerivedPrototypes(src)
	if err != nil {
		t.Fatal(err)
	}
	perl := perlPrototypes(t)
	for _, name := range []string{"caller", "reset", "send", "sleep", "srand", "umask"} {
		if derived[name] != perl[name] {
			t.Errorf("%s: derives (%s), perl's prototype is (%s)", name, derived[name], perl[name])
		}
	}
}
