// ABOUTME: Pins the two WRONG verdicts the four new markers exposed, and what each one means.
// ABOUTME: Both are grammar defects the srefgen-only metric could not see; neither is a harness bug.

package parseoracle

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parser"
)

// corpusFile reads a file from the shim if one is present, so these tests
// describe the real corpus rather than a paraphrase of it.
func corpusFile(t *testing.T, rel string) []byte {
	t.Helper()
	shim := os.Getenv("PARSEORACLE_SHIM")
	if shim == "" {
		t.Skip("PARSEORACLE_SHIM unset; these pin real corpus files")
	}
	src, err := os.ReadFile(filepath.Join(shim, "t", rel))
	if err != nil {
		t.Skipf("corpus file absent: %v", err)
	}
	return src
}

// TestAnonhashLostInFullFile is op/universal.t, and it is a grammar defect
// rather than a marker defect.
//
// perl reports an anonymous hash at lines 16, 30, 104, 151, 184, 195 and 229.
// The subject reports six of the seven: everything but line 16, `$a = {};`.
//
// That statement is not unusual, and it is not the adapter: `$a = {};` alone,
// after a `plan` call, after a BEGIN block, and in a copy of this file's first
// twenty lines ALL yield an anonhash site. Only the full-file parse loses it,
// and in the full file no site of any kind is reported for lines 14..18 --
// the statement is gone from the tree, with HasError() false.
//
// So this is the silent-shred class the stacked-filetest bug belonged to: the
// grammar drops a statement and reports a clean parse. The srefgen-only metric
// could not see it because line 16 takes no reference. Adding the anonhash
// marker is what made a whole missing statement visible, which is the argument
// for the other four markers stated as a measurement rather than a hope.
func TestAnonhashLostInFullFile(t *testing.T) {
	src := corpusFile(t, "op/universal.t")
	tree, err := parser.New().Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if tree.RootNode().HasError() {
		t.Skip("the grammar now reports an error here; re-measure rather than assume")
	}

	var covering int
	for _, s := range TreeSitterSubject(tree).CallSites {
		end := s.EndLine
		if end < s.Line {
			end = s.Line
		}
		if s.Line <= 16 && 16 <= end {
			covering++
		}
	}
	if covering != 0 {
		t.Logf("line 16 is now covered by %d site(s): the grammar defect is fixed, "+
			"so re-baseline and delete this test", covering)
	}
}

// TestBareMatchInForList is comp/our.t. perl reports a match at lines 32 and
// 33 -- `if (/TIE/)` and `elsif (/calls/)`, bare matches against $_ inside a
// `for` whose list is itself a match:
//
//	for ($AUTOLOAD =~ /TieAll::(.*)/) {
//	    if (/TIE/) { return bless {} }
//	    elsif (/calls/) { ... }
//
// The subject does report match sites covering both lines, but they belong to
// the enclosing `for` (31..41) and `if` (32..40) statements, and those sites
// are spent against perl's match for the `for` list itself. The innermost
// statement owning line 32 has no match site of its own, so the rule finds a
// committed statement with no site and no hedge, and says WRONG.
//
// Whether that is the right verdict is a real question and is NOT settled
// here: the subject did see matches at those lines, so "committed with no
// site" overstates what it got wrong. Pinned as the status quo with the
// mechanism recorded, because a WRONG verdict fails a build and this one
// deserves a decision rather than a silent baseline entry.
func TestBareMatchInForList(t *testing.T) {
	src := corpusFile(t, "comp/our.t")
	tree, err := parser.New().Parse(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	var innermost *SubjectCallSite
	for i, s := range TreeSitterSubject(tree).CallSites {
		end := s.EndLine
		if end < s.Line {
			end = s.Line
		}
		if s.Kind != "match" || s.Line > 32 || 32 > end {
			continue
		}
		if innermost == nil || s.Line > innermost.Line {
			innermost = &TreeSitterSubject(tree).CallSites[i]
		}
	}
	if innermost == nil {
		t.Fatal("no match site covers line 32: the shape this test describes has changed")
	}
	t.Logf("innermost match site covering line 32: lines %d..%d",
		innermost.Line, innermost.EndLine)
}
