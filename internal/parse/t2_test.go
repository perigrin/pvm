// ABOUTME: T2 core: 56 files across five perl5 t/ directories, ratcheted per file.
// ABOUTME: The gate wants 100%; the measurement is 25%, so the shortfall is named rather than hidden.

package parse_test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// t2Dirs is spec §7.3's tier-2 core, the same five directories
// `internal/lexer/ratchet_test.go:59` uses. Duplicated rather than shared
// for the reason the parse ratchet records: that one is unexported in
// `package lexer` and this is `package parse_test`.
var t2Dirs = []string{"base", "cmd", "comp", "opbasic", "class"}

// t2Files returns the T2 core, or skips when the perl5 corpus is absent.
func t2Files(t *testing.T) (string, []string) {
	t.Helper()

	root := os.Getenv("PERL5_CORPUS")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("no PERL5_CORPUS and no home directory: %v", err)
		}
		root = filepath.Join(home, "dev", "perl5")
	}
	tDir := filepath.Join(root, "t")
	if _, err := os.Stat(tDir); err != nil {
		t.Skipf("perl5 corpus not present at %s: %v\n"+
			"set PERL5_CORPUS to a checkout to run the T2 metric", tDir, err)
	}

	var files []string
	for _, d := range t2Dirs {
		matches, err := filepath.Glob(filepath.Join(tDir, d, "*.t"))
		if err != nil {
			t.Skipf("globbing %s: %v", d, err)
		}
		for _, m := range matches {
			rel, err := filepath.Rel(tDir, m)
			if err != nil {
				continue
			}
			files = append(files, rel)
		}
	}
	sort.Strings(files)
	return tDir, files
}

// TestT2CoreParses is the M1 gate's T2 metric.
//
// The gate's target is 100% of 56 files and the measurement at d83e21a9 is
// 14 (25.0%). The gate's own text says what to do about that: "Do not carry
// an unreachable 100% into a gate: that is how a gate becomes advisory."
// So this NAMES THE SHORTFALL per file and ratchets it, rather than
// asserting a number no commit today can reach.
//
// Every file below is one the parser does not yet read cleanly. The list is
// the work, not an excuse for it: `comp/proto.t` at 52 belongs to prototype
// recognition, the nine `class/` files to class syntax, `base/lex.t` at 54
// to the lexer's remaining quote-like forms.
//
// Fails in EITHER direction, like the T1 ratchet. A file that starts parsing
// must be moved out of the shortfall in the commit that earned it.
func TestT2CoreParses(t *testing.T) {
	tDir, files := t2Files(t)

	if len(files) != 56 {
		t.Errorf("T2 core is %d files, want 56 (spec §7.3)", len(files))
	}

	// Pinned at d83e21a9: file -> Unknown nodes, for every file that is not
	// clean. A file absent from this map must parse cleanly.
	shortfall := map[string]int{
		"base/lex.t": 54, "base/num.t": 48, "base/rs.t": 3,
		"class/accessor.t": 7, "class/class.t": 4, "class/construct.t": 3,
		"class/destruct.t": 7, "class/field.t": 18, "class/gh22169.t": 3,
		"class/gh23511.t": 1, "class/inherit.t": 12, "class/method.t": 14,
		"class/phasers.t": 5,
		"cmd/for.t":       2, "cmd/mod.t": 1, "cmd/subval.t": 3, "cmd/switch.t": 2,
		"comp/colon.t": 25, "comp/decl.t": 3, "comp/filter_exception.t": 7,
		"comp/final_line_num.t": 1, "comp/fold.t": 14, "comp/form_scope.t": 19,
		"comp/hints.t": 50, "comp/line_debug.t": 5, "comp/multiline.t": 2,
		"comp/opsubs.t": 24, "comp/package.t": 7, "comp/package_block.t": 4,
		"comp/parser.t": 64, "comp/parser_run.t": 12, "comp/proto.t": 52,
		"comp/redef.t": 21, "comp/require.t": 30, "comp/retainedlines.t": 19,
		"comp/uproto.t": 3, "comp/use.t": 11, "comp/utf.t": 3,
		"opbasic/arith.t": 179, "opbasic/cmp.t": 7, "opbasic/concat.t": 16,
		"opbasic/magic_phase.t": 7,
	}

	var regressed, improved, nowClean, nowDirty []string
	clean := 0
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(tDir, rel))
		if err != nil {
			continue
		}
		n := countUnknown(parse.Parse(src))
		if n == 0 {
			clean++
		}

		was, listed := shortfall[rel]
		switch {
		case !listed && n > 0:
			nowDirty = append(nowDirty, fmt.Sprintf("%s: 0 -> %d", rel, n))
		case listed && n == 0:
			nowClean = append(nowClean, rel)
		case listed && n > was:
			regressed = append(regressed, fmt.Sprintf("%s: %d -> %d", rel, was, n))
		case listed && n < was:
			improved = append(improved, fmt.Sprintf("%s: %d -> %d", rel, was, n))
		}
	}
	sort.Strings(regressed)
	sort.Strings(improved)
	sort.Strings(nowClean)
	sort.Strings(nowDirty)

	if len(nowDirty) > 0 {
		t.Errorf("%d file(s) parsed cleanly and no longer do:\n  %s",
			len(nowDirty), strings.Join(nowDirty, "\n  "))
	}
	if len(regressed) > 0 {
		t.Errorf("%d file(s) parse WORSE than the shortfall records:\n  %s",
			len(regressed), strings.Join(regressed, "\n  "))
	}
	if len(nowClean) > 0 {
		t.Errorf("%d file(s) now parse cleanly:\n  %s\n\n"+
			"Remove them from the shortfall map in the commit that earned it.",
			len(nowClean), strings.Join(nowClean, "\n  "))
	}
	if len(improved) > 0 {
		t.Errorf("%d file(s) parse BETTER than the shortfall records:\n  %s\n\n"+
			"Update the shortfall map in the commit that earned it.",
			len(improved), strings.Join(improved, "\n  "))
	}

	t.Logf("T2 core: %d of %d files clean (%.1f%%). The gate's target is 100%%; "+
		"%d files short.", clean, len(files),
		100*float64(clean)/float64(len(files)), len(files)-clean)
}
