// ABOUTME: Tests that a lexical `no warnings` turns off, to the end of its block, the diagnostics of the categories it names.
// ABOUTME: uninitialized covers an Undef where a value is expected, numeric a string or glob where a number is; perl 5.42 -w is the judge.

package infer_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNoWarningsScopeSilencesItsCategories: perl 5.42 -w warns on this
// source at lines 3, 8, 9, 13, 17 and 23, and dies at 28:
//
//   - `no warnings "uninitialized"` silences "Use of uninitialized value"
//     to the end of its block (line 6), and no further (line 8);
//   - `no warnings "numeric"` silences "isn't numeric" (line 12), and
//     leaves the undef (line 13);
//   - "void" is neither, and silences nothing (line 17, an argument);
//   - a bare `no warnings` silences both, a glob's "isn't numeric"
//     included (line 21), until `use warnings` restores them (line 23);
//   - an argument follows the same rule (line 27), while close({}) dies
//     "Not a GLOB reference" under any pragma (line 28).
func TestNoWarningsScopeSilencesItsCategories(t *testing.T) {
	src := `my $u = undef;
my @x = (1);
my $a1 = "x" . $u;
{
    no warnings "uninitialized";
    my $s = "x" . $u;
}
my $a2 = "y" . $u;
my $n1 = "abc" + 1;
{
    no warnings "numeric";
    my $n2 = "abc" + 1;
    my $n3 = $u + 1;
}
{
    no warnings "void";
    my $v = lc($u);
}
{
    no warnings;
    my $w = "x" . $u . ("abc" + 1) . *STDOUT + 1;
    use warnings;
    my $r = "z" . $u;
}
{
    no warnings qw/uninitialized numeric/;
    waitpid($u, 0);
    close({});
}
`
	got := map[int]bool{}
	for _, d := range loweredDiags(src) {
		got[1+bytes.Count([]byte(src[:d.StartByte]), []byte("\n"))] = true
	}
	want := map[int]bool{3: true, 8: true, 9: true, 13: true, 17: true, 23: true, 28: true}
	assert.Equal(t, want, got, "the lines with a diagnostic")
}
