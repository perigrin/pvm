// ABOUTME: Tests for the lowered tree's printers: each node's kind and set fields, children under their field names.
// ABOUTME: The printed children are Children's, so the printer cannot drift from the walk.

package lower_test

import (
	"reflect"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lower"
	"tamarou.com/pvm/internal/parse"
)

func TestTreePrintsKindsFieldsAndRoles(t *testing.T) {
	file := lower.Lower(parse.Parse([]byte("$h->{k};")))
	want := "File\n" +
		"  Stmts: ExprStmt\n" +
		"    X: Key Arrow\n" +
		"      Base: Var Sigil=\"$\" Name=\"h\"\n" +
		"      Key: Bareword Name=\"k\" Quoted\n"
	if got := lower.Tree(file); got != want {
		t.Errorf("Tree:\n%s\nwant:\n%s", got, want)
	}
	wantS := `(File Stmts:(ExprStmt X:(Key Arrow Base:(Var Sigil="$" Name="h") Key:(Bareword Name="k" Quoted))))`
	if got := lower.Sexpr(file); got != wantS {
		t.Errorf("Sexpr:\n%s\nwant:\n%s", got, wantS)
	}
}

func TestPrintMarksHoleAndParen(t *testing.T) {
	file := lower.Lower(parse.Parse([]byte("my $x = ; ($y);")))
	got := lower.Sexpr(file)
	for _, w := range []string{"(Hole)", "Var Paren"} {
		if !strings.Contains(got, w) {
			t.Errorf("Sexpr %s lacks %q", got, w)
		}
	}
}

// TestPrintedChildrenAreChildren: the printer and Children agree on every
// node of a tree that reaches most kinds.
func TestPrintedChildrenAreChildren(t *testing.T) {
	src := `use strict; my ($x, @y) = (1, "a$b", qw(c d)); if ($x) { print $fh map { $_ } @y } ` +
		`for my $i (1..3) { next } while (1) { last } $o->m(1); f(2); sub g { \@_ } eval { 1 };`
	var walk func(lower.Node)
	walk = func(n lower.Node) {
		if got, want := lower.PrintedChildren(n), lower.Children(n); !reflect.DeepEqual(nodesOf(got), want) {
			t.Errorf("%T: Fields %v, Children %v", n, nodesOf(got), want)
		}
		for _, c := range lower.Children(n) {
			walk(c)
		}
	}
	walk(lower.Lower(parse.Parse([]byte(src))))
}

func nodesOf(fs []lower.PrintedChild) []lower.Node {
	var ns []lower.Node
	for _, f := range fs {
		ns = append(ns, f.Node)
	}
	return ns
}
