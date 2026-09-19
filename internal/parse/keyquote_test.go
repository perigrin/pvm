// ABOUTME: A bareword hash subscript is the string, not a call -- perl autoquotes it.
// ABOUTME: It autoquotes even when a sub of that name is in scope, so the two differ.

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestBarewordSubscriptIsAString: `$h{k}` holds the string "k".
//
// It held a Call, which says `$h{k()}` -- a different program. Perl
// autoquotes a bareword subscript EVEN WHEN a sub of that name exists,
// measured on 5.42.0:
//
//	$ perl -e 'sub k { "other" } my %h = (k => 1); print $h{k}, "\n";'
//	1
//
// Beyond the wrong tree, every bareword subscript was a false entry in the
// unresolved-call set: `Call{Resolved:false}` means "a call to something I
// have not seen" (§4.8.3), and a hash key is not a call at all. sites.go
// scores that set, so this inflated the hedge the parse oracle reads.
func TestBarewordSubscriptIsAString(t *testing.T) {
	for _, src := range []string{
		`$h{k};`,
		`$h->{k};`,
		`$h{k} = 1;`,
		`$$r{k};`,
	} {
		root := parse.Parse([]byte(src))
		if hasKind(root, parse.Call) {
			t.Errorf("%s\n  the subscript is a Call; perl autoquotes it to a string:\n%s",
				src, dumpTree(root, 2))
		}
	}

	// An EXPLICIT call is not autoquoted and must stay one.
	for _, src := range []string{
		`$h{k()};`,
		`$h->{k()};`,
	} {
		root := parse.Parse([]byte(src))
		if !hasKind(root, parse.Call) {
			t.Errorf("%s\n  an explicit call was autoquoted away:\n%s",
				src, dumpTree(root, 2))
		}
	}

	// An array subscript is an INDEX, never a key. Measured:
	//
	//	$ perl -e 'sub k { 1 } my @a=(9,8); print $a[k], "\n";'
	//	8
	//
	// It called k() and took element 1.
	root := parse.Parse([]byte(`$a[k];`))
	if !hasKind(root, parse.Call) {
		t.Errorf("$a[k];\n  an array subscript is not autoquoted:\n%s", dumpTree(root, 2))
	}

	// A SLICE is not autoquoted either, and that is the boundary of the
	// rule. Only a lone bareword filling the whole subscript is a key:
	//
	//	$ perl -e 'sub a { "z" } my %h=(a=>1,b=>2); my @s=@h{a,b};
	//	           print join(",", map { $_ // "undef" } @s), "\n";'
	//	undef,2
	//
	// `a` called a() and looked up "z"; `b` had no sub and autoquoted. The
	// autoquote is a property of the single-word subscript, not of braces.
	slice := parse.Parse([]byte(`@h{a,b};`))
	if !hasKind(slice, parse.Call) {
		t.Errorf("@h{a,b};\n  a slice subscript is not autoquoted:\n%s", dumpTree(slice, 2))
	}
}

// TestAutoquoteShrinksTheUnresolvedSet: the fix is worth measuring, not just
// asserting.
//
// `Call{Resolved:false}` means "a call to something I have not seen"
// (§4.8.3), and sites.go scores that set as the parser's hedge. A hash key is
// not a call at all, so every bareword subscript was a false entry inflating
// what the parse oracle reads. Measured over T1:
//
//	         Call nodes   unresolved
//	before       29,757       13,784
//	after        28,695       12,759
//
// The bound is deliberately one-sided. A later change that legitimately
// resolves more calls must not fail this; only a regression putting keys back
// into the set will.
func TestAutoquoteShrinksTheUnresolvedSet(t *testing.T) {
	dir, files := t1Files(t)

	unresolved := 0
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			continue
		}
		var walk func(*parse.Node)
		walk = func(n *parse.Node) {
			if n.Kind == parse.Call && !n.Resolved {
				unresolved++
			}
			for _, c := range n.Children {
				walk(c)
			}
		}
		walk(parse.Parse(src))
	}

	const wasBeforeAutoquote = 13784
	if unresolved >= wasBeforeAutoquote {
		t.Errorf("%d unresolved calls in T1, was %d before bareword subscripts "+
			"stopped counting as calls -- the keys are back in the set",
			unresolved, wasBeforeAutoquote)
	}
}

// hasKind reports whether n or any descendant has the given kind.
func hasKind(n *parse.Node, k parse.Kind) bool {
	if n.Kind == k {
		return true
	}
	for _, c := range n.Children {
		if hasKind(c, k) {
			return true
		}
	}
	return false
}

// dumpTree renders a subtree for a failure message.
func dumpTree(n *parse.Node, indent int) string {
	var b []byte
	var walk func(*parse.Node, int)
	walk = func(n *parse.Node, d int) {
		for i := 0; i < d; i++ {
			b = append(b, ' ')
		}
		b = append(b, n.Kind.String()...)
		if n.Text != "" {
			b = append(b, ' ')
			b = append(b, n.Text...)
		}
		b = append(b, '\n')
		for _, c := range n.Children {
			walk(c, d+2)
		}
	}
	walk(n, indent)
	return string(b)
}
