// ABOUTME: Feature state follows perl's lexical scope and reads bundle names (`:5.10`, `:all`, `:default`).
// ABOUTME: Covers the features the parser consults for `state`, `field` and `isa`, each case measured with perl -c.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestFeatureScopeIsLexical: a `use feature`, `no feature` or `use VERSION`
// lasts to the end of its enclosing block, a bundle name turns on or off
// every feature in the bundle, and `use VERSION` replaces the features in
// effect rather than adding to them. Each source's outcome was measured with
// `perl -c` on 5.44.0; with its feature on, `state` or `field` alone is
// "syntax error at EOF" and `$x isa Y` without it is "Bareword found where
// operator expected".
func TestFeatureScopeIsLexical(t *testing.T) {
	for _, tc := range []struct {
		src    string
		perlOK bool
	}{
		// The three cases of issue 01a17d84-e992.
		{`use feature ':5.10'; state`, false},
		{`use v5.10; {no feature 'state';} state`, false},
		{`{use v5.10;} state`, true},
		{`use v5.10; no feature ':all'; state`, true},

		{`use v5.10; state`, false},
		{`use feature qw(:5.10); state`, false},
		{`use feature ':5.10.1'; state`, false},
		{`use feature ':5.9.5'; state`, false},
		{`use feature ':default'; state`, true},
		{`use v5.10; no feature; state`, true},
		{`use feature ':5.10'; no feature ':5.10'; state`, true},
		{`use feature 'state'; use v5.8; state`, true},
		{`sub f { use feature 'state'; } state`, true},
		{`use feature 'state'; sub f { no feature 'state'; state } state`, false},

		{`{use feature 'class';} field`, true},
		{`use feature 'class'; class C; field`, false},
		{`use feature ':all'; class C; field`, false},
		{`use feature 'class'; class C { no feature ':all'; field }`, true},
		{`use feature 'class'; class C; { no feature 'class'; field }`, true},

		{`{use feature 'isa';} $x isa Y`, false},
		{`{use v5.36;} $x isa Y`, false},
		{`use v5.36; { no feature 'isa'; $x isa Y }`, false},
		{`use feature 'isa'; $x isa Y`, true},
		{`use feature ':5.36'; $x isa Y`, true},
		{`use feature ':all'; $x isa Y`, true},
		{`use v5.36; no feature ':all'; $x isa Y`, false},
		{`use v5.36; no feature ':5.36'; $x isa Y`, false},
		{`use feature 'isa'; use v5.10; $x isa Y`, false},
		{`use feature ':5.36'; $x isa Y; sub f { no feature ':all'; } $y isa Z`, true},
		{`use experimental 'isa'; $x isa Y`, true},
		{`{use experimental 'isa';} $x isa Y`, false},
		{`use v5.36; no experimental 'isa'; $x isa Y`, false},
		{`use experimental 'state'; state`, false},
	} {
		root := parse.Parse([]byte(tc.src))
		refused := root.HasError || len(collectUnknown(root)) != 0
		if refused == tc.perlOK {
			t.Errorf("%s: perl -c OK is %v, the parser refused it: %v %v", tc.src, tc.perlOK, refused, refusalCodes(collectUnknown(root)))
		}
	}
}
