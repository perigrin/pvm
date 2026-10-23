// ABOUTME: The main stash spelled with an empty package name -- `$::{n}`, `%::` -- parses
// ABOUTME: as a named package's stash does; held by tree shape and canon, not by a zero alone.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestMainStashEmptyName holds issue 01a0de97-77fe at the parser. The lexer's
// mainstash_test.go holds the token stream; this holds the tree. Before the
// lexer fix, `$::{foo}` lexed as `$:` (the format line-break variable) plus a
// stray colon, and every empty-name spelling refused:
//
//	my $r = $::{foo};   Unknown=1  not_a_term ":"
//	my @k = keys %::;   Unknown=1  not_a_term ":;"
//	my %h = %::;        Unknown=1  not_a_term ":;"
//
// Valid perl 5.42.0:
//
//	$ perl -MO=Deparse -e 'my $r = $::{foo};'   -> my $r = $main::{'foo'};
//	$ perl -MO=Deparse -e 'my @k = keys %::;'   -> my(@k) = keys %main::;
func TestMainStashEmptyName(t *testing.T) {
	for _, tc := range []struct{ src, shape, canon string }{
		// A subscript of the stash `%::`, keyed `foo`.
		{"my $r = $::{foo};", "( ( (my $r ({ $:: foo))))", "my $r = $::{foo};"},
		// keys is a named unary, written bare; perl 5.44.0 Deparse,-p:
		// (my(@k) = keys(%main::)) and (my(@k) = keys(%Foo::)).
		{"my @k = keys %::;", "( ( (my @k (keys %::))))", "my @k = keys %::;"},
		{"my %h = %::;", "( ( (my %h %::)))", "my %h = %::;"},

		// REGRESSION GUARD: a NAMED package's stash, which always parsed.
		{"my $r = $Foo::{bar};", "( ( (my $r ({ $Foo:: bar))))", "my $r = $Foo::{bar};"},
		{"my @k = keys %Foo::;", "( ( (my @k (keys %Foo::))))", "my @k = keys %Foo::;"},
		{"*Foo:: = \\%Bar::;", "( ( (= *Foo:: (ref %Bar::))))", "*Foo:: = \\%Bar::;"},
		{"my $x = \\%Foo::;", "( ( (my $x (ref %Foo::))))", "my $x = \\%Foo::;"},
		{"my $r = $Foo::bar;", "( ( (my $r $Foo::bar)))", "my $r = $Foo::bar;"},
	} {
		n := parse.Parse([]byte(tc.src))
		if got := countUnknown(n); got != 0 {
			t.Errorf("Parse(%q): Unknown=%d, want 0", tc.src, got)
		}
		if got := shape(n); got != tc.shape {
			t.Errorf("Parse(%q): shape %s, want %s", tc.src, got, tc.shape)
		}
		if got := strings.TrimSpace(parse.Canon(n, []byte(tc.src))); got != tc.canon {
			t.Errorf("Canon(parse(%q)) = %q, want %q", tc.src, got, tc.canon)
		}
	}
}
