// ABOUTME: A call paren after a term follows perl: after a subscript, an arrow or `&` it calls,
// ABOUTME: and after a plain term -- `$x()`, `my $x()`, `f()()` -- perl refuses it and so do we.
package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestPostfixCallParenFollowsPerl holds perly.y's rule for a `(` after a
// term. It calls only a `subscripted` -- an element, a list slice, an
// arrow call `->(...)`, or a call after one of those -- or the name that
// `&` or `->` introduces. Measured on 5.42.0 with `perl -c`:
//
//	refused   $x()  my $x()  $$x()  ${$x}()  @x()  %x()  f()()  (f)()
//	          1()  "a"()  $x++()  $x->m()()  $x->$m()()  &$x()()  &f()()
//	          @x[0]()  @x{a}()  $x->@[0]()  $#x()  -$x()  0()  0(
//	accepted  $x[0]()  $x{a}()  $x->[0]()  $x->{a}()  $x->()()
//	          $x->(1)(2)(3)  $x[0]()()  $$x[0]()  ${$x}[0]()  (1)[0]()
//	          ()[0]()  qw(a b)[0]()  &$x()  &f()  &{$x}()  &{; $x}()
//	          $x->$m()  $x->${\"m"}()  \&$x()  $x->m->()  $x->(1)[0]()
//	          *x{CODE}()  __PACKAGE__()  time ()  print ()  CORE::time()
//
// `$x->$m()` and `$x()` reach the same `(` with the same left node, a
// Term `$m` or `$x`; only the arrow before it differs.
func TestPostfixCallParenFollowsPerl(t *testing.T) {
	for _, src := range []string{
		"$x();", "my $x();", "my$0();", "$$x();", "${$x}();", "@x();", "%x();",
		"f()();", "(f)();", "1();", `"a"();`, "$x++();", "$x->m()();",
		"$x->$m()();", "&$x()();", "&f()();", "@x[0]();", "@x{a}();",
		"$x->@[0]();", "$#x();", "-$x();", "f($x());", "0()", "0(",
	} {
		if firstUnknown(parse.Parse([]byte(src))) == nil {
			t.Errorf("Parse(%q) has no Unknown; perl refuses it", src)
		}
	}
	for _, src := range []string{
		"$x[0]();", "$x{a}();", "$x->[0]();", "$x->{a}();", "$x->()();",
		"$x->(1)(2)(3);", "$x[0]()();", "$$x[0]();", "${$x}[0]();",
		"(1)[0]();", "()[0]();", "qw(a b)[0]();", "&$x();", "&f();",
		`&{$x}();`, "&{; $x}();", "$x->$m();", `$x->${\"m"}();`,
		`\&$x();`, "$x->m->();", "$x->(1)[0]();", "$x->[0](1)(2);",
		"$x{a}{b}();", "sub {1}->()();", "f($x->[0]());", "*x{CODE}();",
		"__PACKAGE__();", "time ();", "print ();", "CORE::time();",
	} {
		if !checkCanonFixpoint(t, src) {
			t.Errorf("Parse(%q) has an Unknown; perl accepts it", src)
		}
	}

	// After a builtin that takes a handle, the scalar before a `(` is the
	// handle, measured on 5.42.0 with -MO=Deparse: `print $x(1), 2` is
	// `print $x 1, 2` and `system $x(sh), 1` is `system $x 'sh', 1`.
	for _, src := range []string{
		"print $x(1), 2;", "print $x ();", `printf $x("%s", 1);`,
		"system $x(sh), 1;", "exec $x(1);",
	} {
		n := parse.Parse([]byte(src))
		if !checkCanonFixpoint(t, src) {
			t.Errorf("Parse(%q) has an Unknown; perl accepts it", src)
		}
		if !hasHandle(n, "$x") {
			t.Errorf("Parse(%q): $x is not the handle", src)
		}
	}
}

// hasHandle reports whether a node in n is the handle text.
func hasHandle(n *parse.Node, text string) bool {
	if n.Handle && n.Text == text {
		return true
	}
	for _, c := range n.Children {
		if hasHandle(c, text) {
			return true
		}
	}
	return false
}
