// ABOUTME: Regression tests asserting the version-resolution error suggestions
// ABOUTME: do not reference removed/nonexistent commands like 'pvm resolve'

package current

import (
	stdErrors "errors"
	"strings"
	"testing"
)

func TestEnhanceResolutionErrorDoesNotSuggestPVMResolve(t *testing.T) {
	// 'pvm resolve' (top-level) was referenced in error messages but never
	// implemented. The real command is 'pvm perl resolve'. Telling users
	// to run a nonexistent top-level command is worse than no suggestion.
	cases := []struct {
		name    string
		message string
	}{
		{"generic-fallback", "could not find perl"},
		{"perl-version-missing", "no .perl-version file found"},
		{"version-not-available", "version 5.40.0 not available"},
		{"no-versions", "no versions available"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			enhanced := enhanceResolutionError(stdErrors.New(tc.message))
			if enhanced == nil {
				t.Fatal("expected non-nil enhanced error")
			}
			// Quoted literal "'pvm resolve'" (with single quotes) was the
			// exact phrasing of the original bug — checking the quoted form
			// also avoids false negatives from the legitimate 'pvm perl resolve'.
			if strings.Contains(enhanced.Error(), "'pvm resolve'") {
				t.Errorf("error suggests nonexistent 'pvm resolve' command: %q", enhanced.Error())
			}
		})
	}
}

func TestEnhanceResolutionErrorReturnsNilForNil(t *testing.T) {
	if got := enhanceResolutionError(nil); got != nil {
		t.Errorf("expected nil for nil input, got %v", got)
	}
}
