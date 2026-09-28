// ABOUTME: `try BLOCK catch (VAR) BLOCK [finally BLOCK]` is one statement under feature 'try',
// ABOUTME: and Try::Tiny's `try {...} catch {...};` stays the two calls it is.
package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestTryCatchIsOneStatement holds the try/catch statement of feature 'try'.
//
// Measured on perl 5.42.0:
//
//	$ perl -e 'use feature "try"; try { die "x\n" } catch ($e) { print "c $e" }
//	      finally { print "f\n" } print "after\n"'
//	c x
//	f
//	after
//
// No `;` after the last block, and the next statement follows directly -- a
// statement form, not an expression. op/try.t carried 59 Unknowns, one per
// clause: `try {` read as a call taking a block, `catch ($e)` as a call with
// a list, and the catch block then stranded.
//
// The keyword reading is taken only when `catch` is followed by `(`, because
// the same two words spell Try::Tiny, where they ARE calls:
//
//	$ perl -e 'use feature "try"; try { 1 } catch { 2 }'
//	catch block requires a (VAR)
//
// So under the feature `catch {` is an error, and without it `catch (` is.
// The paren settles which one the source is, with no feature tracking needed.
func TestTryCatchIsOneStatement(t *testing.T) {
	cases := []struct {
		name  string
		src   string
		canon string
	}{
		{"try catch", `try { f() } catch ($e) { g($e) }`,
			`try {f();} catch ($e) {g($e);}`},
		{"with finally", `try { f() } catch ($e) { 1 } finally { h() }`,
			`try {f();} catch ($e) {1;} finally {h();}`},
		{"statement follows without a semicolon", `try { 1 } catch ($e) { 2 } print "a";`,
			`try {1;} catch ($e) {2;} print("a");`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := []byte(c.src)
			n := parse.Parse(src)
			if got := countUnknown(n); got != 0 {
				t.Errorf("Parse(%q) has %d Unknown, want 0", c.src, got)
			}
			got := strings.TrimSpace(parse.Canon(n, src))
			if got != c.canon {
				t.Errorf("Canon(%q):\n  got  %s\n  want %s", c.src, got, c.canon)
			}
			again := strings.TrimSpace(parse.Canon(parse.Parse([]byte(got)), []byte(got)))
			if again != got {
				t.Errorf("canon is not a fixpoint:\n  once  %s\n  twice %s", got, again)
			}
		})
	}
}

// TestTryTinyStaysACall guards the other reading. Try::Tiny's `try` and
// `catch` are subs taking a block, joined by nothing but juxtaposition, and a
// parser that took `try {` as the keyword would read a statement where the
// source has an expression ending in `;`.
func TestTryTinyStaysACall(t *testing.T) {
	src := []byte(`try { f() } catch { g() };`)
	n := parse.Parse(src)
	got := strings.TrimSpace(parse.Canon(n, src))
	if strings.HasPrefix(got, "try {f();} catch (") {
		t.Errorf("Canon(%q) = %s: Try::Tiny read as the try keyword", src, got)
	}
	var sawTryStatement bool
	var walk func(*parse.Node)
	walk = func(n *parse.Node) {
		if n.Kind == parse.Conditional && n.Text == "try" {
			sawTryStatement = true
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(n)
	if sawTryStatement {
		t.Errorf("Parse(%q) built a try statement; `catch {` has no (VAR), so these are calls", src)
	}
}
