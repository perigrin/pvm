// ABOUTME: Tests for the round-3 false positives cutover 1 adds on the lowered path: a guarded capture a sub returns, a Glob in a Str slot.
// ABOUTME: Each silent case is one perl 5.42 -w runs silently, paired with a control perl warns about.

package infer_test

import (
	"testing"

	"tamarou.com/pvm/internal/infer"
)

// A capture returned in the branch its own successful match guards is
// defined, so a sub whose other returns are builtin results hands its
// caller no Undef of its own. perl 5.42 -w is silent on each case below:
// the `if` modifier, the block form, `unless` over `!~`, and a broad
// Scalar or a Hole's Any returned beside the capture.
func TestCaptureReturnedAfterMatchIsDefined(t *testing.T) {
	for _, src := range []string{
		"sub xfileno { return $1 if $_[0] =~ /\\A=?(\\d+)\\z/; return fileno $_[0]; }\nprint \"same\\n\" if xfileno(\\*STDOUT) == fileno \\*STDOUT;\n",
		"sub xf { if ($_[0] =~ /(\\d+)/) { return $1 } return fileno $_[0] }\nprint \"x\\n\" if xf(\"a7\") == 7;\n",
		"sub xf { return $1 unless $_[0] !~ /(\\d+)/; return fileno $_[0] }\nprint \"x\\n\" if xf(\"a7\") == 7;\n",
		"sub xs { return $1 if $_[0] =~ /(\\d+)/; my $v = shift; return $v }\nprint \"x\\n\" if xs(1) == 1;\n",
		"sub xa { return $1 if $_[0] =~ /(\\d+)/; return eval '1' }\nprint \"x\\n\" if xa(1) == 1;\n",
	} {
		wantSilent(t, src)
	}

	// perl 5.42 -w warns "Use of uninitialized value in numeric eq (==)"
	// on each: the capture is not under its own successful match -- no
	// match guards it, the guard is `!~`, or the match has no such group.
	for _, src := range []string{
		"sub xf { return $1 if @_; return fileno $_[0] }\nprint \"x\\n\" if xf(\\*STDOUT) == 1;\n",
		"sub xf { return $1 if $_[0] !~ /(\\d+)/; return fileno $_[0] }\nprint \"x\\n\" if xf(\"ab\") == 1;\n",
		"sub xf { return $2 if $_[0] =~ /(\\d+)/; return fileno $_[0] }\nprint \"x\\n\" if xf(\"a7\") == 1;\n",
	} {
		wantReported(t, src, infer.CodeCoercionMismatch)
	}

	// The `if` modifier's match runs before its body, though the source
	// puts it last, so the capture is that match's group, not an earlier
	// match's: perl 5.42 -w warns "Argument "a" isn't numeric in numeric
	// eq (==)" on each.
	for _, src := range []string{
		"sub xf { return $1 if $_[0] =~ /([a-z])/; return fileno $_[0] }\nprint \"x\\n\" if xf(\"a\") == 1;\n",
		"\"x5\" =~ /(\\d)/;\nsub xf { return $1 if $_[0] =~ /([a-z])/; return fileno $_[0] }\nprint \"x\\n\" if xf(\"a\") == 1;\n",
	} {
		wantReported(t, src, infer.CodeCoercionMismatch)
	}
}
