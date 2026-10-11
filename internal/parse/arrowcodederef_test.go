// ABOUTME: An arrow code-deref call keeps its parens: `$r->($v->[0])` calls $r with one argument.
// ABOUTME: Each case is held against the tree, canon's fixpoint, and perl's own Deparse of canon.
package parse_test

import (
	"os/exec"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/conformance"
	"tamarou.com/pvm/internal/parse"
)

// TestArrowCodeDerefCallKeepsItsParens holds perl's reading of `->(`: the
// parens are the call's argument list, whatever the argument is. Measured on
// 5.42.0 with -MO=Deparse:
//
//	$r->($v->[0])               &$r($v->[0])
//	$r->($v->{k})               &$r($v->{'k'})
//	$r->($a[0])                 &$r($a[0])
//	$r->(f(1))                  &$r(f(1))
//	$r->($x)                    &$r($x)
//	$r->(1, 2)                  &$r(1, 2)
//	$r->()                      &$r()
//	__SUB__->($value->{$_})     __SUB__->($value->{$_})
//	$r->[0]->($x->{k})          $r->[0]->($x->{'k'})
//	$r->($x)->($y->[0])         &$r($x)->($y->[0])
//	$r->($v->[0])->[1]          &$r($v->[0])->[1]
//	$r->([1])                   &$r([1])
//	$r->($$x)                   &$r($$x)
//	$r->(\$x)                   &$r(\$x)
//	$r->($v->[0],)              &$r($v->[0])
//
// An argument that was itself subscripted lost the parens: `$r->($v->[0])`
// was a Binary `->` whose right operand was the Index `$v->[0]` with no
// Paren, and canon wrote `$r -> $v->[0]`, a subscript chain on $r.
func TestArrowCodeDerefCallKeepsItsParens(t *testing.T) {
	perl, perlErr := conformance.PerlPath()
	deparse := func(src string) string {
		out, err := exec.Command(perl, "-MO=Deparse", "-e", src).CombinedOutput()
		if err != nil {
			t.Errorf("perl refuses %q: %v\n%s", src, err, out)
		}
		return string(out)
	}

	for _, tc := range []struct {
		src  string
		args []string // each `->(...)` call's parenthesised argument list
	}{
		{"my $z = $r->($v->[0]);", []string{"($v->[0])"}},
		{"my $z = $r->($v->{k});", []string{"($v->{k})"}},
		{"my $z = $r->($a[0]);", []string{"($a[0])"}},
		{"my $z = $r->(f(1));", []string{"(f(1))"}},
		{"my $z = $r->($x);", []string{"($x)"}},
		{"my $z = $r->(1, 2);", []string{"(1, 2)"}},
		{"my $z = $r->();", []string{"()"}},
		{"use feature 'current_sub'; sub g { my $z = __SUB__->($value->{$_}) }", []string{"($value->{$_})"}},
		{"my $z = $r->[0]->($x->{k});", []string{"($x->{k})"}},
		{"my $z = $r->($x)->($y->[0]);", []string{"($x)", "($y->[0])"}},
		{"my $z = $r->($v->[0])->[1];", []string{"($v->[0])"}},
		{"my $z = $r->([1]);", []string{"([1])"}},
		{"my $z = $r->($$x);", []string{"($$x)"}},
		{"my $z = $r->(\\$x);", []string{"(\\$x)"}},
		{"my $z = $r->($v->[0],);", []string{"($v->[0],)"}},
	} {
		root := parse.Parse([]byte(tc.src))
		if u := firstUnknown(root); u != nil {
			t.Errorf("Parse(%q) has an Unknown: %q", tc.src, u.SourceText([]byte(tc.src)))
			continue
		}
		got := arrowCallArgs(root, []byte(tc.src))
		if strings.Join(got, " ") != strings.Join(tc.args, " ") {
			t.Errorf("Parse(%q): `->(` call arguments = %q, want %q\n  shape %s",
				tc.src, got, tc.args, shape(skipWrappers(root)))
		}
		if ok, why := canonMatches([]byte(tc.src)); !ok {
			t.Errorf("%q: canon is not a fixpoint: %s", tc.src, why)
		}
		if perlErr != nil {
			continue
		}
		canon := strings.TrimSpace(parse.Canon(root, []byte(tc.src)))
		if want, have := deparse(tc.src), deparse(canon); have != want {
			t.Errorf("perl reads canon differently from the source:\n  src   %s\n  canon %s\n  perl(src)   %s\n  perl(canon) %s",
				tc.src, canon, want, have)
		}
	}
}

// arrowCallArgs lists, in source order, the source text of the argument list
// of every `->(` call in n: a Binary `->` whose right operand is the
// parenthesised list.
func arrowCallArgs(n *parse.Node, src []byte) []string {
	var out []string
	for _, c := range n.Children {
		out = append(out, arrowCallArgs(c, src)...)
	}
	if n.Kind == parse.Binary && n.Text == "->" && len(n.Children) == 2 {
		if r := n.Children[1]; r.Paren && src[r.Start] == '(' {
			out = append(out, r.SourceText(src))
		}
	}
	return out
}
