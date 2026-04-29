// ABOUTME: Tests for the doctor summary tip that points users to auto-fix
// ABOUTME: Verifies the suggestion only appears when the auto-fix engine actually has applicable fixes

package pvm

import (
	"strings"
	"testing"
)

func TestDoctorSummaryTipNamesRealCommand(t *testing.T) {
	// Whatever wording the tip uses, it must reference 'pvm self doctor',
	// not the dead top-level 'pvm doctor'.
	tip := doctorSummaryTip(true)
	if strings.Contains(tip, "pvm doctor") && !strings.Contains(tip, "pvm self doctor") {
		t.Errorf("tip references nonexistent 'pvm doctor': %q", tip)
	}
}

func TestDoctorSummaryTipMentionsFixWhenFixable(t *testing.T) {
	tip := doctorSummaryTip(true)
	if !strings.Contains(tip, "pvm self doctor --fix") {
		t.Errorf("expected fixable tip to mention 'pvm self doctor --fix', got %q", tip)
	}
}

func TestDoctorSummaryTipDoesNotPromiseFixWhenNothingFixable(t *testing.T) {
	// When the auto-fix engine has nothing applicable, the user should not
	// be told to run --fix as if it would resolve their warnings. The tip
	// is reworded but still shown so users discover --fix exists for the
	// general case.
	tip := doctorSummaryTip(false)
	if strings.Contains(tip, "automatically resolve") {
		t.Errorf("tip should not promise auto-resolution when nothing is fixable; got %q", tip)
	}
	if !strings.Contains(tip, "pvm self doctor") {
		t.Errorf("tip should still mention 'pvm self doctor' for discoverability; got %q", tip)
	}
}
