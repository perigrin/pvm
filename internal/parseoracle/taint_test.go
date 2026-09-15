// ABOUTME: Tests that a shebang's taint request is read as a flag cluster, not as a stray letter.
// ABOUTME: Perl accepts -T and -t, combined (-wt) or separated (-w -t), and a path is not a switch.

package parseoracle

import "testing"

// TestWantsTaintFlagForms covers every spelling of a taint request perl
// accepts, and the two ways of getting it wrong: missing a form, and reading a
// letter that is part of a path rather than part of a switch.
func TestWantsTaintFlagForms(t *testing.T) {
	cases := []struct {
		name string
		line string
		want bool
	}{
		{"uppercase", "#!./perl -T", true},
		{"lowercase", "#!./perl -t", true},
		{"combined_wt", "#!./perl -wt", true},
		{"combined_Tw", "#!./perl -Tw", true},
		{"separated", "#!./perl -w -t", true},
		{"no_taint_flag", "#!./perl -w", false},
		{"bare", "#!/usr/bin/perl", false},
		{"t_in_interpreter_path", "#!/usr/bin/perl-testing", false},
		{"t_in_flag_argument", "#!./perl -Mstrict", false},
		{"no_shebang", "use strict;", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := []byte(c.line + "\nprint 1;\n")
			if got := wantsTaint(src); got != c.want {
				t.Errorf("wantsTaint(%q) = %t, want %t", c.line, got, c.want)
			}
		})
	}
}
