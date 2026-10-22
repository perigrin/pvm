// ABOUTME: Tests for Options.WarnPossibleMismatch: a union argument with a member that
// ABOUTME: fits the parameter is silent by default; one with no fitting member still reports.

package infer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tamarou.com/pvm/internal/infer"
	"tamarou.com/pvm/internal/types"
)

// TestPossibleMismatchSilentByDefault: perigrin's ruling (2026-10-21,
// "a union argument with a fitting member is a possible mismatch").
// Symbol::qualify's Ref candidate preserves whatever reference it is
// given -- measured, `ref qualify({})` is HASH, not GlobRef -- so
// qualify($h, ...) with $h a HashRef is typed Ref: ScalarRef, ArrayRef,
// HashRef, CodeRef and Object besides the one member, GlobRef, that
// tell's Glob-only parameter takes. psc knows only that the value is
// some reference, never which kind, the same loss of precision
// possible-undef forgives for a declared call's Undef, and perl 5.42 -w
// is silent here since $h is always a GlobRef in two
// IO/Handle.pm:516's own shape.
func TestPossibleMismatchSilentByDefault(t *testing.T) {
	const src = `use Symbol qw(qualify);
my $h = {};
my $r = tell(qualify($h, "main"));
`
	assert.Empty(t, diagsWithOptions(src, infer.Options{}))
}

// TestPossibleMismatchReportedUnderFlag: the same call, reported as an
// ordinary mismatch under --warn-possible-mismatch, as possible-undef is
// under --warn-possible-undef.
func TestPossibleMismatchReportedUnderFlag(t *testing.T) {
	const src = `use Symbol qw(qualify);
my $h = {};
my $r = tell(qualify($h, "main"));
`
	diags := diagsWithOptions(src, infer.Options{WarnPossibleMismatch: true})
	if assert.Len(t, diags, 1) {
		assert.Equal(t, infer.CodeTypeMismatch, diags[0].Code)
	}
}

// TestNoFittingMemberStillReported: a union with no member that fits the
// parameter is still a coercion-mismatch, flag or no flag -- unlike
// qualify's Ref, neither a ScalarRef nor a CodeRef is ever the handle
// tell's Glob-only parameter takes.
func TestNoFittingMemberStillReported(t *testing.T) {
	const src = `my $y = 1;
my $x = $y ? \$y : sub { 1 };
my $r = tell($x);
`
	assert.Len(t, diagsWithOptions(src, infer.Options{}), 1)
	assert.Len(t, diagsWithOptions(src, infer.Options{WarnPossibleMismatch: true}), 1)
}

// TestFittingMemberNeverHidesUndef: I6, a fitting member must never
// hide an Undef possibleUndef itself would not forgive (perigrin,
// 2026-10-21). $h's type is GlobRef|ArrayRef|Undef: the outer
// ternary's else arm is itself a ternary whose own else arm is a
// literal `undef`, which markPossibleUndef never marks (it marks only
// a declared call's own join), so neither ternary is markedPossible-
// Undef. GlobRef fits tell's FileHandle slot, but that fitting member
// must not silence the Undef beside it, flag or no flag: perl 5.42
// dies "Can't use an undefined value as a symbol reference" on
// `tell(undef)`, and dies differently again when $h is the ArrayRef.
func TestFittingMemberNeverHidesUndef(t *testing.T) {
	const src = `my $h = @ARGV ? \*STDOUT : @ARGV > 1 ? [1] : undef;
my $r = tell($h);
`
	assert.NotEmpty(t, diagsWithOptions(src, infer.Options{}))
	assert.NotEmpty(t, diagsWithOptions(src, infer.Options{WarnPossibleUndef: true}))
}

// TestSymbolQualifyReturnsItsRef: Symbol::qualify's Ref candidate returns
// the reference it was given, unchanged -- measured, `ref qualify({})`
// is HASH, not GlobRef as it read while deferred pending this ruling
// (issue 01a15baf).
func TestSymbolQualifyReturnsItsRef(t *testing.T) {
	const src = `use Symbol qw(qualify);
my $h = {};
my $r = qualify($h, "main");
`
	got := resolvedCallType(src, `qualify($h, "main")`)
	assert.Equal(t, types.Ref, got)
}
