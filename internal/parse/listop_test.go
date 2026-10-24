// ABOUTME: Canon writes list operators, user, method and indirect calls and niladic builtins by precedence:
// ABOUTME: bare where the source was bare, call parens only where the source wrote them or binding power needs them.
package parse

import (
	"strings"
	"testing"
)

// The perl facts these tests rest on, measured with the pinned perl 5.44.0
// (`perl -MO=Deparse,-p -e SRC`):
//
//	push @a, 1 or die                     (push(@a, 1) or die);
//	print $x, 1 and f                     (print($x, 1) and '???');
//	sub foo; sub bar; foo 1, bar 2, 3     foo(1, bar(2, 3));
//	package Foo; sub new {} package main; new Foo "a"
//	                                      'Foo'->new('a');
//	find PerlIO::Layer "perlio"           'PerlIO::Layer'->find('perlio');
//	print {$fh} "x"                       print({$fh;} 'x');
//	sort { $a <=> $b } @a                 sort {$a <=> $b} @a;
//	print ($a), $b                        (print($a), $b);
//	shift @a                              shift(@a);
//	print("x")                            print('x');
//	time                                  time;
//	time()                                time;
//	fork;                                 fork;
//	wantarray;                            wantarray;
//	print __FILE__                        print(__FILE__);
//	print __PACKAGE__, "x"                print(__PACKAGE__, 'x');
//	__PACKAGE__->new                      __PACKAGE__->new;
//
// A list operator's call parens are not in perl's tree: `push @a, 1` and
// `push(@a, 1)` deparse alike. So canon writes the spelling the source
// wrote, and adds call parens only where a tighter operator after the call
// would otherwise bind into its last argument.
//
// "Emits exactly" compares significant tokens, the comparison Faithful
// makes, so canon's ` , ` spacing is not pinned. A block's last `;` is
// written by canon and forgiven by Faithful, so the block forms' expected
// text carries it.

// TestCanonWritesListOpByPrecedence: a list operator or call the source
// wrote bare is re-emitted bare, and source parens are kept.
func TestCanonWritesListOpByPrecedence(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`print "x";`, `print "x";`},
		{`push @a, 1 or die;`, `push @a, 1 or die;`},
		{`print $x, 1 and f;`, `print $x, 1 and f;`},
		{`sub foo; sub bar; foo 1, bar 2, 3;`, `sub foo; sub bar; foo 1, bar 2, 3;`},
		{`print {$fh} "x";`, `print {$fh;} "x";`},
		{`print $fh "x";`, `print $fh "x";`},
		{`sort { $a <=> $b } @a;`, `sort {$a <=> $b;} @a;`},
		{`map { $_ } @a;`, `map {$_;} @a;`},
		{`package Foo; sub new {1} package main; new Foo "a";`, `package Foo; sub new {1;} package main; new Foo "a";`},
		{`find PerlIO::Layer 'perlio';`, `find PerlIO::Layer 'perlio';`},
		{`shift @a;`, `shift @a;`},
		{`print ($a), $b;`, `print ($a), $b;`},
		{`print("x");`, `print("x");`},
		{`$o->m(1, 2);`, `$o->m(1, 2);`},
		{`$o->m();`, `$o->m();`},
		{`$o->m;`, `$o->m;`},
	} {
		n := Parse([]byte(tc.src))
		requireCanon(t, n, tc.src, tc.want)
		if ok, why := Faithful(n, []byte(tc.src)); !ok {
			t.Errorf("Faithful(%q) = false: %s", tc.src, why)
		}
	}
}

// TestCanonWritesNiladicBare: a niladic builtin or compile-time token is
// written as its bare name, and the only pair written after it is the
// source's own, which the tree holds as an Index over the Call.
func TestCanonWritesNiladicBare(t *testing.T) {
	for _, src := range []string{
		`time;`,
		`fork;`,
		`wantarray;`,
		`print __FILE__;`,
		`print __PACKAGE__, "x";`,
		`__PACKAGE__->new;`,
		`time();`,
	} {
		n := Parse([]byte(src))
		requireCanon(t, n, src, src)
		if ok, why := Faithful(n, []byte(src)); !ok {
			t.Errorf("Faithful(%q) = false: %s", src, why)
		}
	}
}

// TestCanonListOpWrapsLooserOperand: a wrong tree's list extent is a paren
// the source lacks. Canon's text only; the verdicts are asserted once the
// call-paren forgiveness is removed.
func TestCanonListOpWrapsLooserOperand(t *testing.T) {
	// push over `or`: the looser operand is wrapped.
	src := `push @a, 1 or die;`
	n := Parse([]byte(src))
	st := lastStmt(t, n)
	st.Children[0] = rotate(t, st.Children[0])
	requireCanon(t, n, src, `push (@a, 1 or die);`)

	// foo(1, bar(2), 3) for `foo 1, bar 2, 3`: bar's comma follows it, so
	// bar takes call parens.
	src = `sub foo; sub bar; foo 1, bar 2, 3;`
	n = Parse([]byte(src))
	st = lastStmt(t, n)
	foo := st.Children[0]
	outer := foo.Children[0] // 1 , bar(2 , 3)
	bar := outer.Children[1]
	inner := bar.Children[0] // 2 , 3
	two, three := inner.Children[0], inner.Children[1]
	wrongBar := *bar
	wrongBar.Children = []*Node{two}
	left := *outer
	left.Children = []*Node{outer.Children[0], &wrongBar}
	whole := *outer
	whole.Children = []*Node{&left, three}
	foo.Children = []*Node{&whole}
	requireCanon(t, n, src, `sub foo; sub bar; foo 1, bar (2), 3;`)

	// print over `($a), $b`: the bare list would begin with a `(` that
	// closes early, which perl reads as the call's parens.
	src = `print ($a), $b;`
	n = Parse([]byte(src))
	st = lastStmt(t, n)
	wrong := rotate(t, st.Children[0])
	wrong.ArgParen = false
	comma := wrong.Children[0]
	a := *comma.Children[0]
	a.Paren = true
	a.Start = strings.Index(src, "(")
	a.End = strings.Index(src, ")") + 1
	comma.Children[0] = &a
	st.Children[0] = wrong
	requireCanon(t, n, src, `print(($a), $b);`)
}

// TestCanonIndirectObjectKeepsSourceSpelling: an indirect-object call is
// written method then class, as the source wrote it, with its argument
// parens only where the source had them.
func TestCanonIndirectObjectKeepsSourceSpelling(t *testing.T) {
	const decl = `package Foo; sub new {1} package main; `
	for _, tc := range []struct{ src, want string }{
		{decl + `new Foo "a";`, `package Foo; sub new {1;} package main; new Foo "a";`},
		{decl + `new Foo("a");`, `package Foo; sub new {1;} package main; new Foo("a");`},
		{decl + `new Foo;`, `package Foo; sub new {1;} package main; new Foo;`},
		{decl + `new Foo();`, `package Foo; sub new {1;} package main; new Foo();`},
	} {
		n := Parse([]byte(tc.src))
		requireCanon(t, n, tc.src, tc.want)
		if ok, why := Faithful(n, []byte(tc.src)); !ok {
			t.Errorf("Faithful(%q) = false: %s", tc.src, why)
		}
	}
}

// TestCanonListOpKeepsSourceArgParen: the source's `NAME(` survives, and a
// tree that lost the ArgParen flag emits the bare spelling, which is a
// difference from the source.
func TestCanonListOpKeepsSourceArgParen(t *testing.T) {
	for _, tc := range []struct{ src, name, bare string }{
		{`print("x");`, "print", `print "x";`},
		{`push(@a, 1);`, "push", `push @a, 1;`},
		{`sub foo; foo(1);`, "foo", `sub foo; foo 1;`},
	} {
		n := Parse([]byte(tc.src))
		requireCanon(t, n, tc.src, tc.src)
		call := findCall(n, tc.name)
		if call == nil || !call.ArgParen {
			t.Fatalf("%s in %q: want a Call with ArgParen, got %+v", tc.name, tc.src, call)
		}
		call.ArgParen = false
		requireCanon(t, n, tc.src, tc.bare)
		if sameToks(canonToks(n, tc.src), significantTokens([]byte(tc.src))) {
			t.Errorf("a tree without ArgParen re-emits %q unchanged", tc.src)
		}
	}
}

// TestCanonListOpTrailingCommaEndsList: a bare call whose last argument
// carries the source's trailing comma stays bare, because perl ends the list
// at that comma before these operators. Measured on perl 5.44.0, each
// `substr $x, 0, 1, OP $y` deparses as `(substr($x, 0, 1) OP $y)`:
//
//	=    substr($x, 0, 1) = $y;
//	.=   (substr($x, 0, 1) .= $y);
//	||=  (substr($x, 0, 1) ||= $y);
//	&&=  (substr($x, 0, 1) &&= $y);
//	==   (substr($x, 0, 1) == $y);
//	||   (substr($x, 0, 1) or $y);
//
// Written with call parens it would say the same, but not what the source
// says; the comma is the source's own boundary.
func TestCanonListOpTrailingCommaEndsList(t *testing.T) {
	for _, op := range []string{"=", ".=", "||=", "&&=", "==", "||"} {
		src := `substr $x, 0, 1, ` + op + ` $y;`
		n := Parse([]byte(src))
		call := findCall(n, "substr")
		if call == nil || len(call.Children) != 1 || call.Children[0].TrailingComma != "," || !call.Children[0].EndsList {
			t.Errorf("%q: want substr's one argument to carry the trailing comma that ends its list, got %+v", src, call)
			continue
		}
		requireCanon(t, n, src, src)
		if ok, why := Faithful(n, []byte(src)); !ok {
			t.Errorf("Faithful(%q) = false: %s", src, why)
		}
	}
}

// TestCanonListOpTrailingCommaDoesNotEndList: before these operators perl
// does not end the list at a trailing comma. Each begins a term after a
// comma, so it re-lexes or is a syntax error, measured on perl 5.44.0 for
// `substr $x, 0, 1, OP $y`:
//
//	-=   syntax error, near "-="
//	*=   syntax error, near "*= $y"
//	%=   syntax error, near "%= $y"
//	/=   Search pattern not terminated
//	&=   syntax error, near "&= $y"
//	**=  substr($x, 0, 1, (** = $y));     the glob **
//	x=   substr($x, 0, 1, ('x' = $y));    the bareword x
//	//=  substr($x, 0, 1, (// = $y));     an empty match
//	<<=  Use of bare << to mean <<"" is forbidden
//
// So no call here ends at the comma, and the trailing-comma rule never
// keeps one bare before them: the parser refuses the statement or reads
// the operator's first character as the start of a further argument.
func TestCanonListOpTrailingCommaDoesNotEndList(t *testing.T) {
	for _, op := range []string{"-=", "*=", "%=", "/=", "&=", "**=", "x=", "//=", "<<="} {
		src := `substr $x, 0, 1, ` + op + ` $y;`
		n := Parse([]byte(src))
		if call := findCall(n, "substr"); call != nil {
			for _, c := range call.Children {
				if c.TrailingComma != "" || c.EndsList {
					t.Errorf("%q: substr's list ends at the comma, which perl does not do", src)
				}
			}
		}
	}
}

// TestCanonListOpTrailingCommaBeforeCloser: a trailing comma before a closer
// ends no list before an operator, so the call is written by precedence.
// Canon does not write every source paren, and a comma after a trailing one
// does not end the list either. Measured on perl 5.44.0:
//
//	sub foo; my @l = (foo 1, 2,), 3;   ((my(@l) = foo(1, 2)), '???');
//	sub foo; my @l = foo 1, 2 ,, 3;    (my(@l) = foo(1, 2, 3));
//	sub foo; my @l = (1 * foo 1, 2,) + 3;
//	                                   (my(@l) = ((1 * foo(1, 2)) + 3));
//
// So the call takes parens there, and the meaning is kept.
func TestCanonListOpTrailingCommaBeforeCloser(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`sub foo; my @l = (foo 1, 2,), 3;`, `sub foo; my @l = foo(1, 2,), 3;`},
		{`sub foo; my @l = ($y = foo 1, 2,), 3;`, `sub foo; my @l = $y = foo(1, 2,), 3;`},
		{`sub foo; my @l = (1 * foo 1, 2,) + 3;`, `sub foo; my @l = 1 * foo(1, 2,) + 3;`},
	} {
		n := Parse([]byte(tc.src))
		if call := findCall(n, "foo"); call == nil || call.Children[0].TrailingComma != "," || call.Children[0].EndsList {
			t.Fatalf("%q: want foo's argument to carry a trailing comma that ends no list, got %+v", tc.src, call)
		}
		requireCanon(t, n, tc.src, tc.want)
	}
}

// TestCanonCommaFollowsWordUnspaced: perl's checkcomma (toke.c) counts the
// blanks after a word into the word, so a bare word written before ` ,` as
// print's first argument is read as a filehandle. Measured on perl 5.44.0:
//
//	print time, "x"           print(time, 'x');
//	print time , "x"          No comma allowed after filehandle
//	print __PACKAGE__ , "x"   No comma allowed after filehandle
//	print time, or die        (print(time) or die);
//	print time , or die       No comma allowed after filehandle
//
// Canon writes the comma right after its left operand, as Deparse does.
// This pins the bytes, not only the tokens, because the blank is the
// difference.
func TestCanonCommaFollowsWordUnspaced(t *testing.T) {
	for _, src := range []string{
		`print time, "x";`,
		`print __PACKAGE__, "x";`,
		`use constant FOO => 1; print FOO, "x";`,
		`print(time, "x");`,
		`print time, or die;`,
		`print time,;`,
	} {
		if got := Canon(Parse([]byte(src)), []byte(src)); strings.Contains(got, " ,") {
			t.Errorf("Canon(%q) = %q, which writes a blank before a comma", src, got)
		}
	}
}

// TestCanonFatCommaKeepsSourceCallParens: a call the source wrote with
// parens before `=>` is a call, not an autoquoted word, and keeps them.
// Measured on perl 5.44.0, `sub foo {7} print foo() => 1` is
// `print(foo(), 1)` and `print foo => 1` is `print('foo', 1)`; in a list,
// `(foo() => 1)` is `(foo(), 1)`.
func TestCanonFatCommaKeepsSourceCallParens(t *testing.T) {
	for _, src := range []string{
		`sub foo {7} print foo() => 1;`,
		`sub foo {7} print foo => 1;`,
		`sub foo {7} my @a = (foo() => 1);`,
		`sub foo {7} my @a = (foo => 1);`,
	} {
		requireCanon(t, Parse([]byte(src)), src, strings.Replace(src, "{7}", "{7;}", 1))
	}
}
