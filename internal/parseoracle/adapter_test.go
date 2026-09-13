// ABOUTME: Tests that our adapter attributes every call site to the statement perl would record it under.
// ABOUTME: Perl's nextstate names a statement's first line, so a site anywhere inside it must report that line.

package parseoracle

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parser"
)

// TestTreeSitterSitesCarryStatementLines: a reference on line 4 belongs to the
// statement that starts on line 3, because that is the line perl's nextstate
// attributes the srefgen to. Reporting the backslash's own line would put the
// two sides one line apart and manufacture a deficit out of agreement.
func TestTreeSitterSitesCarryStatementLines(t *testing.T) {
	src := "sub f{}\n" + // 1
		"my @a;\n" + // 2
		"f(\n" + // 3
		"  \\@a,\n" + // 4
		");\n" + // 5
		"if (0) { }\n" + // 6
		"elsif (f(@a)) { }\n" + // 7  an elsif condition has a nextstate of its own
		"sub g {\n" + // 8
		"  f(@a);\n" + // 9  the inner statement, not the sub declaration
		"}\n" // 10

	tree, err := parser.New().Parse([]byte(src))
	require.NoError(t, err)

	subject := TreeSitterSubject(tree)
	require.True(t, subject.OK, subject.DeclinedReason)

	// Only the attribution is under test here. Whether the grammar hedges or
	// commits on a given call shape is its own question, with its own tests.
	type site struct {
		Line int
		Took bool
	}
	var got []site
	for _, c := range subject.CallSites {
		got = append(got, site{c.Line, c.TookReference})
	}
	assert.ElementsMatch(t, []site{
		{3, false}, {3, true}, // the call, and the \@a on line 4 inside it
		{7, false},
		{9, false},
	}, got)
}
