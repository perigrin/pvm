// ABOUTME: Tests that the PR #486 review fixes add no false positive to psc check: each repro of perl 5.42's lib is silent, each control still reported.
// ABOUTME: The repros and controls are testdata/review_regressions; perl 5.42 -w is silent on each repro and warns on each control.

package psc_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// diagLine matches a diagnostic psc check prints: file:line:col: severity.
var diagLine = regexp.MustCompile(`(?m)^\S+:\d+:\d+: (error|warning): .*$`)

// TestReviewFixesAddNoFalsePositives: psc check over perl 5.42's lib drew
// 75 lines after the review fixes that it did not before them. Each
// repro here is one root cause of them, in the shape the lib has it, and
// perl 5.42 -w is silent on each, so psc check must be too. Each control
// is the same rule where perl 5.42 -w warns, and psc check must still
// report it at the line perl names.
func TestReviewFixesAddNoFalsePositives(t *testing.T) {
	isolatePerl(t)
	cases := []struct {
		file string
		want string // the line:col a control is reported at; "" for a repro
	}{
		{"truthy.pl", ""},
		{"truthy_control.pl", "t.pl:10:11: error:"},
		{"assign_guard.pl", ""},
		{"assign_guard_control.pl", "t.pl:8:11: error:"},
		{"loop_exit.pl", ""},
		{"loop_exit_control.pl", "t.pl:11:7: error:"},
		{"ternary_guard.pl", ""},
		{"ternary_guard_control.pl", "t.pl:6:16: error:"},
		{"modifier.pl", ""},
		{"modifier_control.pl", "t.pl:8:7: error:"},
		{"aggregate.pl", ""},
		{"aggregate_control.pl", "t.pl:8:14: error:"},
		{"first_element.pl", ""},
		{"first_element_control.pl", "t.pl:8:7: error:"},
		{"glob_value.pl", ""},
		{"glob_value_control.pl", "t.pl:7:11: error:"},
		{"ref_eq.pl", ""},
		{"ref_eq_control.pl", "t.pl:6:14: error:"},
		{"closure_scope.pl", ""},
		{"closure_scope_control.pl", "t.pl:7:46: error:"},
		{"list_assign.pl", ""},
		{"list_assign_control.pl", "t.pl:8:10: error:"},
		{"handle_name.pl", ""},
		{"handle_name_control.pl", "t.pl:5:6: error:"},
		{"vstring.pl", ""},
		{"vstring_control.pl", "t.pl:6:7: warning:"},
	}
	for _, c := range cases {
		t.Run(c.file, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("testdata", "review_regressions", c.file))
			require.NoError(t, err)
			diags := diagLine.FindAllString(runCheckOn(t, string(source)), -1)
			if c.want == "" {
				assert.Empty(t, diags, "perl 5.42 -w is silent on %s", c.file)
				return
			}
			found := false
			for _, d := range diags {
				if strings.Contains(d, string(filepath.Separator)+c.want) {
					found = true
				}
			}
			assert.True(t, found, "control: %s is reported at %s, as perl 5.42 -w warns there; got %q", c.file, c.want, diags)
		})
	}
}
