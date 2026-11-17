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
		{"sub print { { 1; } }", true},
		{"sub map { { my $y = 2; } } 1;", true},
		{"sub exec { {;} }", true},
		{`use v5.38; use feature "class"; no warnings; class C { method print { { 1; } } }`, true},
		{"package Fmt; sub printf { { local $_ = shift; s/x/y/; } return 1; }", true},
		{"sub grep { sub : lvalue {1} }", true},
		{"my @x; my @y = (<<E, map {\nbody\nE\n{ 1; } } @x);", false},
		{"my @x; my @y = (<<E, map {\nbody\nE\n{ 1 } } @x);", true},
	} {
		if got := pipelineSource(t, c.src); got != c.accepted {
			t.Errorf("%q: the chart accepts=%v, perl -c accepts=%v", c.src, got, c.accepted)
		}
	}
}

// TestBlockExpectationBraceKnownLimits pins what the pipeline accepts and
// perl -c refuses here, each row failing once its limit moves: an unfeatured
// `say` (internal/lexer does not track the say feature), and a newline or
// comment between `sub` and `:` (toke.c loses the `sub` with its line).
func TestBlockExpectationBraceKnownLimits(t *testing.T) {
	for _, src := range []string{
		"use v5.10; say { {1;} } 1;",
		"my @x; my @y = map { sub\n: lvalue {1} } @x;",
		"my @x; my @y = map { sub #c\n: lvalue {1} } @x;",
	} {
		if !pipelineSource(t, src) {
			t.Errorf("%q: the chart refuses, as perl -c does; the limit moved", src)
		}
	}
}
