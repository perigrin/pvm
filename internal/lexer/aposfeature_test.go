// ABOUTME: The apostrophe package separator follows feature apostrophe_as_package_separator:
// ABOUTME: on by default, off under the 5.41+ bundle and `no feature`, back on with `use feature`.

package lexer

import (
	"reflect"
	"testing"
)

// TestApostropheSeparatorFollowsItsFeature: toke.c reads `'` as `::` only
// while feature apostrophe_as_package_separator is on. It is in the default
// bundle and every bundle through :5.40, and out of :5.41 onward. Measured
// on 5.42.0 with -MO=Deparse:
//
//	$main'a;      $'b;      $'x3;      main'f();
//	  $a            $b        $x3        f()          no `use`, use v5.40
//	  $main, then   $', then  $' x 3     main, then   use v5.41, use v5.42,
//	  a string      Bareword             a string     no feature '...'
//
// and `use feature 'apostrophe_as_package_separator'` after `use v5.42`
// restores the first row. Each case lists the significant tokens of the
// statement after the pragmas.
func TestApostropheSeparatorFollowsItsFeature(t *testing.T) {
	const off = "use v5.42;\n"
	const offNumeric = "use 5.042;\n"
	const offFeature = "no feature 'apostrophe_as_package_separator';\n"
	const back = "use v5.42; use feature 'apostrophe_as_package_separator';\n"
	cases := []struct {
		pragma, body string
		want         []string
	}{
		// The feature is on: the apostrophe joins a package name.
		{"", "$main'a;", []string{"$main'a", ";"}},
		{"", "$'b;", []string{"$'b", ";"}},
		{"", "$'x3;", []string{"$'x3", ";"}},
		{"", "main'f();", []string{"main'f", "(", ")", ";"}},
		{"", "sub isn't {1}", []string{"sub", "isn't", "{", "1", "}"}},
		{"use v5.40;\n", "$'x3;", []string{"$'x3", ";"}},
		{back, "$'x3;", []string{"$'x3", ";"}},
		{back, "$main'a;", []string{"$main'a", ";"}},

		// The feature is off: `$'` is the postmatch and a following
		// apostrophe opens a string.
		{off, "$'x3;", []string{"$'", "x", "3", ";"}},
		{off, "$'b;", []string{"$'", "b", ";"}},
		{off, "$main'a';", []string{"$main", "'a'", ";"}},
		{off, "main'f'();", []string{"main", "'f'", "(", ")", ";"}},
		{"use v5.41;\n", "$'x3;", []string{"$'", "x", "3", ";"}},
		{offNumeric, "$'x3;", []string{"$'", "x", "3", ";"}},
		{offFeature, "$'x3;", []string{"$'", "x", "3", ";"}},
		{offFeature, "$main'a';", []string{"$main", "'a'", ";"}},
	}
	for _, c := range cases {
		src := c.pragma + c.body
		var got []string
		for _, tok := range significant(Tokenize([]byte(src))) {
			if tok.Start >= len(c.pragma) {
				got = append(got, src[tok.Start:tok.End])
			}
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%q: got %q, want %q", src, got, c.want)
		}
	}
}
