// ABOUTME: Canon keeps a dereference's braces unless the operand is a plain scalar
// ABOUTME: variable: `${$h->{k}}` and `$$h->{k}` are different programs.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDerefCanonKeepsBraces: a deref Unary emitted its sigil and then its
// operand, so the braces of `${EXPR}` were dropped whatever EXPR was. The
// bare `$$name` form reads only a scalar variable, and anything more binds
// differently. Measured on 5.42.0:
//
//	$ perl -e 'my $h={k=>\"K"}; my $x = ${$h->{k}}; print "$x\n";
//	      my $y = eval q{$$h->{k}}; print defined $y ? "y=$y\n" : "err: $@";'
//	K
//	err: Not a SCALAR reference at (eval 1) line 1.
//
// and Deparse writes `${*$f;}`, `${$h->{'k'};}`, `${\$x;}` and `$$$r`.
func TestDerefCanonKeepsBraces(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"my $a = ${$h->{k}};", "my $a = ${$h->{k}};"},
		{"my $x = ${*$f};", "my $x = ${*$f};"},
		{"my $z = ${\\$x};", "my $z = ${\\$x};"},
		{"my @v = @{$o->list};", "my @v = @{$o -> list};"},
		{"my @y = @{$r};", "my @y = @$r;"},
		{"my $w = $$$r;", "my $w = $$$r;"},
		// Bare, the operand has to be a name the lexer reads after `$$`:
		// `$$;` is the pid and then a terminator, and `$$ x` the pid and a
		// bareword. Measured on 5.42.0, `$; = \"v"; print ${$;}` prints `v`
		// where `print $$;` prints the pid. The parser fuzzer found it from
		// `${$#}`, whose bare `$$#` re-lexed as the pid and a comment.
		{"my $p = ${$;};", "my $p = ${$;};"},
		{"my $q = ${$ x};", "my $q = ${$ x};"},
		{"my $n = ${$1};", "my $n = $$1;"},
		{"my $m = ${$::x};", "my $m = $$::x;"},
		// Issue 01a0e045's shapes, from T1's package_decl_stash.t. Dropping
		// the braces made `$` take `$pkg` alone. Deparse writes
		// `${$pkg . '::';}` and `!!${$pkg . '::';}{$1 . '::'}`.
		{"my $x = ${$pkg . '::'};", "my $x = ${$pkg . '::'};"},
		{"my $x = !!${$pkg . '::'}{$1 . '::'};", "my $x = !!${$pkg . '::'}{$1 . '::'};"},
		{"f(${$pkg . '::'});", "f(${$pkg . '::'});"},
		// The main stash subscript once canon'd as `$:;:{$pkg}`. Deparse
		// writes `$main::{$pkg}`.
		{"my $x = $::{$pkg};", "my $x = $::{$pkg};"},
		{"my $x = $::{$pkg . '::'};", "my $x = $::{$pkg . '::'};"},
		// A plain scalar operand needs no braces, as `@{$r}` above.
		{"my $v = $$r;", "my $v = $$r;"},
		{"my %v = %{$h};", "my %v = %$h;"},
		{"my $w = ${$h}{k};", "my $w = $$h{k};"},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		got := parse.Canon(root, []byte(tc.src))
		if got != tc.want {
			t.Errorf("%q: canon %q, want %q", tc.src, got, tc.want)
		}
		if again := parse.Canon(parse.Parse([]byte(got)), []byte(got)); again != got {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", tc.src, got, again)
		}
	}
}
