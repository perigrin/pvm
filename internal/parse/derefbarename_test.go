// ABOUTME: A bareword then a subscript inside deref braces names a variable, as toke.c's scan_ident reads it:
// ABOUTME: `@{ foo { "a" } }` is the slice `@foo{"a"}`, whatever sub foo is declared, and canon keeps the braces.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/perltest"
)

// TestDerefBareNameBraceReadsAsPerlDoes: uni/variables.t:402 is
//
//	sub foo (&) { [1] }
//	my %foo = (a=>2);
//	my $ret = @{ foo { "a" } };
//
// and psc read `foo { "a" }` as a call with a block argument, so canon wrote
// `@{foo {"a";}}`, which perl rejects. toke.c's scan_ident reads an
// identifier then `[` or `{` inside `${`, `@{`, `%{` or `*{` as the
// variable's name (the "${foo[0]} and ${foo{bar}} notation") and drops the
// closing brace, with or without the sub, its prototype or `use strict`.
// Measured on perl 5.44.0, -MO=Concise of the line:
//
//	e     <1> padsv_store[$ret:4,5] vKS/LVINTRO ->f
//	d        <@> hslice sK ->e
//	b           <$> const[PV "a"] s ->c
//
// the same ops as `@foo{"a"}`, and Deparse,-p writes `@foo{'a'}`. A word
// with no subscript after it is not this rule, and `sub` is excluded by
// name: `@{ sub {1} }` deparses as a deref of an anonymous sub.
func TestDerefBareNameBraceReadsAsPerlDoes(t *testing.T) {
	cases := []struct{ src, plain, canon string }{
		{`sub foo (&) { [1] } my %foo; my $ret = @{ foo { "a" } };`,
			`sub foo (&) { [1] } my %foo; my $ret = @foo{"a"};`,
			`sub foo (&) {[1];} my %foo;my $ret = @{ foo{"a"}};`},
		{"my %foo; my $ret = @{\n  foo { \"a\" }\n };",
			"my %foo; my $ret = @foo{\n  \"a\"\n };",
			"my %foo;my $ret = @{\n  foo{\"a\"}};"},
		{`my @r = ${ foo [0] };`, `my @r = $foo[0];`, `my @r = ${ foo[0]};`},
		{`my @r = ${ foo {a} {b} };`, `my @r = $foo{a}{b};`, `my @r = ${ foo{a}{b}};`},
		{`my @r = ${ foo {a} }{b};`, `my @r = $foo{a}{b};`, `my @r = ${ foo{a}}{b};`},
		{`my @r = %{ Foo::bar {a} };`, `my @r = %Foo::bar{a};`, `my @r = %{ Foo::bar{a}};`},
		{`my @r = *{ foo {a} };`, `my @r = *foo{a};`, `my @r = *{ foo{a}};`},
	}
	for _, tc := range cases {
		root := parse.Parse([]byte(tc.src))
		plain := parse.Parse([]byte(tc.plain))
		if got, want := shape(root), shape(plain); got != want {
			t.Errorf("%q: got %s, want %s, the tree of %q", tc.src, got, want, tc.plain)
		}
		if got := parse.Canon(root, []byte(tc.src)); got != tc.canon {
			t.Errorf("%q: canon %q, want %q", tc.src, got, tc.canon)
		}
		if ok, why := parse.Faithful(root, []byte(tc.src)); !ok {
			t.Errorf("%q: not faithful: %s", tc.src, why)
		}
	}

	// `sub` is never a name here, nor a word with a leading `::`. Measured
	// on 5.44.0, `@{ ::foo {"a"} }` is rv2av over entersub, with no rv2hv,
	// `%{ ::foo {a} }` and `@{ ::foo::bar {a} }` deparse as derefs, and
	// `@{ ::foo [1] }` is a syntax error.
	for _, tc := range [][2]string{
		{`my @r = @{ sub {1} };`, `my @r = @sub{1};`},
		{`my @r = @{ ::foo {"a"} };`, `my @r = @::foo{"a"};`},
		{`my @r = %{ ::foo {a} };`, `my @r = %::foo{a};`},
		{`my @r = @{ ::foo::bar {a} };`, `my @r = @::foo::bar{a};`},
	} {
		if got := shape(parse.Parse([]byte(tc[0]))); got == shape(parse.Parse([]byte(tc[1]))) {
			t.Errorf("%q read as the slice %q: %s", tc[0], tc[1], got)
		}
	}
	// Sites, against perl 5.44.0's rv2hv count: only %h's.
	checkDecided(t, []arrowCase{
		{`my @r = @{ ::foo {"a"} };`, 0},
		{`my %h; my @r = @{ ::foo { %h } };`, 0},
		{`our %h; my @r = @{ ::foo { $h{b} } };`, 1},
	})

	t.Run("perl", func(t *testing.T) {
		perl := perltest.PinnedPerl(t)
		concise := func(src string) string {
			out, err := perl.Command("-MO=Concise", "-e", src).CombinedOutput()
			if err != nil {
				t.Errorf("perl rejects %q: %s", src, out)
			}
			return string(out)
		}
		// Each source against its plain spelling and its canon, which must
		// compile to the same ops. Each pair shares its declarations, so
		// Concise numbers both programs' pads and ops alike.
		for _, tc := range cases {
			src := concise(tc.src)
			for _, other := range []string{tc.plain, tc.canon} {
				if got := concise(other); got != src {
					t.Errorf("perl compiles %q and %q to different ops:\n%s\n%s", tc.src, other, src, got)
				}
			}
		}
		// The sub's prototype and strict play no part.
		const dollar, strict = `sub foo ($) { [1] } my %foo; `, `use strict; sub foo (&) { [1] } my %foo; `
		for _, pre := range []string{dollar, strict} {
			if a, b := concise(pre+`my $ret = @{ foo { "a" } };`), concise(pre+`my $ret = @foo{"a"};`); a != b {
				t.Errorf("after %q, perl compiles the braced name and the slice to different ops:\n%s\n%s", pre, a, b)
			}
		}
	})
}
