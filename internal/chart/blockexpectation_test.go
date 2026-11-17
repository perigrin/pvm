// ABOUTME: The pipeline's verdicts on the first `{` of an XREF block, which the lexer reads by toke.c's block_expectation.
// ABOUTME: Each row's verdict is perl -c's on the pinned perl (94e50866).

package chart

import "testing"

func TestBlockExpectationBracePipeline(t *testing.T) {
	for _, c := range []struct {
		src      string
		accepted bool
	}{
		{"my @x; my @y = map { { $_->get_columns } } @x;", true},
		{"my @x; my @y = map { { 1; } } @x;", false},
		{"my @x; my @y = map {; {1} } @x;", true},
		{"my @x; my @y = map { sub : lvalue { 1 } } @x;", true},
		{"print { {1} } 1;", true},
	} {
		if got := pipelineSource(t, c.src); got != c.accepted {
			t.Errorf("%q: the chart accepts=%v, perl -c accepts=%v", c.src, got, c.accepted)
		}
	}
}
