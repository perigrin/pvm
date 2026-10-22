// ABOUTME: Tests what the XS utility modules' .tpm files say to the parser: what a bare use imports,
// ABOUTME: and that List::Util's block-taking subs read their braces as a block.

package parse

import (
	"testing"
)

// TestScalarUtilBareUseImportsNothing: Scalar::Util's @EXPORT is empty, so
// a bare `use Scalar::Util` imports nothing and a call to blessed carries
// no declaration; asked for by name, it carries Scalar/Util.tpm's.
// Measured on 5.42.0: `perl -e 'use Scalar::Util; print
// defined(&main::blessed) ? 1 : 0'` prints 0.
func TestScalarUtilBareUseImportsNothing(t *testing.T) {
	call := findCallNamed(ParseWithLoader([]byte("use Scalar::Util; my $x; blessed($x);\n"), noModules), "blessed")
	if call == nil {
		t.Fatal("no call to blessed")
	}
	if call.Signatures != nil {
		t.Errorf("a bare use imported blessed: signatures %+v", call.Signatures)
	}

	call = findCallNamed(ParseWithLoader([]byte("use Scalar::Util qw(blessed); my $x; blessed($x);\n"), noModules), "blessed")
	if call == nil {
		t.Fatal("no call to blessed")
	}
	if call.Signatures == nil {
		t.Error("use Scalar::Util qw(blessed) did not import blessed's declaration")
	}
}

// blockArg walks the leftmost child of call, as mapbrace_test.go's
// firstArgIsAnonHash does for map and grep, to say whether the call's
// argument begins with a block or an anonymous hash.
func blockArg(call *Node) (block, hash bool) {
	n := call
	for len(n.Children) > 0 {
		n = n.Children[0]
		if n.Kind == AnonHash {
			return false, true
		}
		if n.Kind == Block {
			return true, false
		}
	}
	return false, false
}

// TestListUtilFirstParsesBlockArgument: first's prototype is (&@), so
// after `use List::Util qw(first)` the braces after it are a block, not a
// hash constructor. Measured on 5.42.0, `perl -MO=Deparse -e 'use
// List::Util qw(first); first { $_ } @x'` prints `first {$_;} @x;`.
func TestListUtilFirstParsesBlockArgument(t *testing.T) {
	src := "use List::Util qw(first); my @x; my $f = first { $_ } @x;\n"
	root := ParseWithLoader([]byte(src), noModules)
	call := findCallNamed(root, "first")
	if call == nil {
		t.Fatalf("no call to first")
	}
	block, hash := blockArg(call)
	if !block || hash {
		t.Errorf("first's braces are no block: call %+v", call)
	}
}
