// ABOUTME: `for VAR (LIST)` with a bare package variable rather than a `my` declaration.
// ABOUTME: The variable sits outside the parens and the list inside them, in both spellings.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestForLoopBareVariable: the loop variable belongs BEFORE the parens and the
// list INSIDE them, whether the variable is declared or a bare package one.
//
// `for my $i (2)` canoned correctly and `for $i (2)` came back as
// `for ($i) 2 {3;}` -- the variable parenthesised as though it were the list
// and the list emitted bare after it -- at Unknown = 0. Only the canon can see
// that: a swapped emission is not a refusal, so the count stays at zero.
//
// perl keeps the order, measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'our $i; for $i (1,2) { print $i }'
//	foreach $i (1, 2) { print $i; }
//
// Asserted on the SIGNIFICANT TOKEN STREAM rather than on exact canon text,
// because that is the criterion the swap actually violates -- `for ( $i ) 2 {`
// against `for $i ( 2 ) {` -- and it does not pin canon's spacing, which is
// not what this defect is about. A hand-written `want` string on the canon
// would have asserted incidental whitespace alongside the real claim.
func TestForLoopBareVariable(t *testing.T) {
	for _, src := range []string{
		"for $i (2) { 3; }",
		"foreach $i (2) { 3; }",

		// The `my` spelling, which was already right -- a regression guard on
		// the shape this fix has to keep.
		"for my $i (2) { 3; }",

		// The other three declarators reach the same Declaration path, and
		// none of them had been measured before this.
		"for our $i (2) { 3; }",
		"for local $i (2) { 3; }",
		"for state $i (2) { 3; }",

		// A package-qualified name is one Variable token, so it takes the bare
		// path too.
		"for $main::i (2) { 3; }",
		"foreach $Foo::Bar::x (@l) { 3; }",

		// An array or hash element as the loop variable: perl accepts neither
		// as an aliasable target, but the parse must still place what it read.
		"for $i (1, 2, 3) { 3; }",

		// Nested, which is the shape the defect was found in.
		"for my $t (1) { for $i (2) { 3 if 1; } }",

		// No loop variable at all: the list alone is the parenthesised part.
		"for (1, 2) { 3; }",
	} {
		b := []byte(src)
		root := parse.Parse(b)
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		got := parse.Canon(root, b)
		if want, have := significant(b), significant([]byte(got)); !equalTokens(want, have) {
			t.Errorf("%q canon reorders its tokens:\n canon %q\n  from %s\n    to %s",
				src, strings.TrimSpace(got),
				strings.Join(want, " | "), strings.Join(have, " | "))
		}
	}
}

// equalTokens compares two significant-token slices element for element.
func equalTokens(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
