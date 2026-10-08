// ABOUTME: A trailing comma before an operator that cannot begin a term ends a list
// ABOUTME: operator's arguments: `substr $x, 0, 1, = "a"` assigns to the call.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestTrailingCommaBeforeAssignment: `=` cannot begin an element, so the
// comma before it separates nothing and the list ends there; the list
// operator is then a term, and the assignment takes it as its left side.
// Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'substr $x, 0, 1, = "\x{100}";'
//	substr($x, 0, 1) = "\x{100}";
//	$ perl -MO=Deparse -e 'substr $x, 0, 1, .= "a";'
//	substr($x, 0, 1) .= 'a';
//
// perl.git t/op/utf8cache.t:70.
func TestTrailingCommaBeforeAssignment(t *testing.T) {
	for src, same := range map[string]string{
		`substr $x, 0, 1, = "\x{100}";`: `substr($x, 0, 1) = "\x{100}";`,
		`substr $x, 0, 1, .= "a";`:      `substr($x, 0, 1) .= "a";`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		// The trees agree; canon keeps the source's trailing comma, so the
		// comparison is of shape (see TestCanonKeepsTrailingComma).
		got := shape(root)
		want := shape(parse.Parse([]byte(same)))
		if got != want {
			t.Errorf("%q: shape %s, want %s as for %q", src, got, want, same)
		}
	}
}

// TestTrailingCommaBeforeOperator: the same holds for every infix operator
// that cannot begin a term where one is expected. Measured on 5.42.0, with
// `sub f {}` declared:
//
//	f 1, || 2;       '???' unless f 1;
//	f 1, && 2;       '???' if f 1;
//	f 1, == 2;       f(1) == 2;
//	f 1, ? 1 : 2;    f(1) ? '???' : '???';
//	f 1, =~ /x/;     f(1) =~ /x/;
//	f 1, .. 3;       f(1) .. 3;
//	f 1, . "x";      f(1) . 'x';
//	f 1, != 2;       f(1) != 2;
//
// So does `->`: `f 1, ->m` is `f(1)->m` and `f 1 => ->m` the same, which
// is how PerlOnJava unit/threads_postfix_create_and_invalid_entry.t chains
// `create threads sub {...}=>->join`.
//
// A compound assignment assigns to f's result, so f is an lvalue sub for
// one: with a plain `sub f {}`, `f 1, ||= 2` is "Can't modify non-lvalue
// subroutine call of &main::f in logical or assignment (||=)", and with
// `sub f :lvalue {}` perl accepts it.
//
// An operator whose first byte DOES begin a term there is read as that
// term, and perl rejects the line: `f 1, += 2` is a syntax error, `f 1, **
// 2` reads a glob and `f 1, // 2` a pattern. Those stay refused.
//
// perl.git t/io/open.t:281, `ok open(...), '...',` then `|| _diag $!`.
func TestTrailingCommaBeforeOperator(t *testing.T) {
	for src, same := range map[string]string{
		`sub f {} f 1, && 2;`:          `sub f {} f(1) && 2;`,
		`sub f {} f 1, || 2;`:          `sub f {} f(1) || 2;`,
		`sub f {} f 1, == 2;`:          `sub f {} f(1) == 2;`,
		`sub f {} f 1, ? 1 : 2;`:       `sub f {} f(1) ? 1 : 2;`,
		`sub f {} f 1, =~ /x/;`:        `sub f {} f(1) =~ /x/;`,
		`sub f {} f 1, .. 3;`:          `sub f {} f(1) .. 3;`,
		`sub f {} f 1, . "x";`:         `sub f {} f(1) . "x";`,
		`sub f {} f 1, != 2;`:          `sub f {} f(1) != 2;`,
		`sub f :lvalue {} f 1, ||= 2;`: `sub f :lvalue {} f(1) ||= 2;`,
		`sub f :lvalue {} f 1, >>= 2;`: `sub f :lvalue {} f(1) >>= 2;`,
		`sub f {} f 1, ->m;`:           `sub f {} f(1)->m;`,
		`sub f {} f 1 => ->m;`:         `sub f {} f(1)->m;`,
		`sub f {} f 1, -> [0];`:        `sub f {} f(1)->[0];`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		// The trees agree; canon keeps the source's trailing comma, so the
		// comparison is of shape (see TestCanonKeepsTrailingComma).
		got := shape(root)
		want := shape(parse.Parse([]byte(same)))
		if got != want {
			t.Errorf("%q: shape %s, want %s as for %q", src, got, want, same)
		}
	}
	for _, src := range []string{
		`sub f {} f 1, += 2;`,
		`sub f {} f 1, -= 2;`,
		`sub f {} f 1, *= 2;`,
		`sub f {} f 1, %= 2;`,
	} {
		if root := parse.Parse([]byte(src)); !containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl rejects this; got %s", src, parse.Canon(root, []byte(src)))
		}
	}
}
