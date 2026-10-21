// ABOUTME: Tests ConstantDifferences reports a zero-arg constant whose die-or-return behavior
// ABOUTME: on the running perl disagrees with what its wantsValue classification (declared with a value, or None) says.

package perltest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstantDifferencesReportsADyingConstant(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Const", "package My::Const;\nsub GOOD () { 1 }\nsub BAD () { die \"nope\\n\" }\n1;\n")

	diffs := ConstantDifferences(t, p, lib, "My::Const", map[string]bool{"GOOD": true, "BAD": true})

	assert.NotContains(t, diffs, "GOOD", "GOOD is declared with a value and returns one, as declared")
	assert.Contains(t, diffs, "BAD", "BAD is declared with a value but dies, a difference")
}

func TestConstantDifferencesReportsAConstantThatReturnsInsteadOfDying(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Const2", "package My::Const2;\nsub DIES () { die \"nope\\n\" }\nsub RETURNS () { 1 }\n1;\n")

	diffs := ConstantDifferences(t, p, lib, "My::Const2", map[string]bool{"DIES": false, "RETURNS": false})

	assert.NotContains(t, diffs, "DIES", "DIES is declared None and dies, as declared")
	assert.Contains(t, diffs, "RETURNS", "RETURNS is declared None but returns a value, a difference")
}

// TestConstantDifferencesTreatsAMissingSubAsDying: a name wantsValue
// names that the module does not define at all behaves the same as one
// that dies -- there is no third way to call it.
func TestConstantDifferencesTreatsAMissingSubAsDying(t *testing.T) {
	p := Perl542(t)
	lib := t.TempDir()
	WriteModule(t, lib, "My::Const3", "package My::Const3;\n1;\n")

	diffs := ConstantDifferences(t, p, lib, "My::Const3", map[string]bool{"NOSUCHSUB": true})

	assert.Contains(t, diffs, "NOSUCHSUB")
}
