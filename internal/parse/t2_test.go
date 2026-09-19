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
// the work, not an excuse for it: `comp/proto.t` belongs to prototype
// recognition, the `class/` files to class syntax, `base/lex.t` to the
// lexer's remaining quote-like forms.
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
	//
	// The t/class counts moved when the lexer stopped reading a quote-op
	// keyword as a quote operator where perl reads it as a name: `$o->s`
	// and `method y { ... }`. class/class.t left the map entirely.
	shortfall := map[string]int{
		"base/lex.t": 37, "base/num.t": 48, "base/rs.t": 3,
		// class/field.t went 9 -> 10 when quote-op keywords stopped eating
		// their fat comma, then back to 9 when goto, delete and exists
		// landed. The rise was never a regression in the parse: reaching
		// further into the file exposed a pre-existing defect, a TRAILING
		// COMMA before a closing paren.
		//
		//	C->new(alpha => "A");     parses
		//	C->new(alpha => "A",);    one Unknown for the `)`
		//
		// Verified to predate the change by stashing it and re-measuring the
		// same source. field.t:224-228 is a multi-line call with a trailing
		// comma, so the count rose while the parse got better. Untracked:
		// nothing in the chain owns trailing commas in argument lists.
		"class/construct.t": 2,
		// class/gh22169.t went 3 -> 5 when anonymous subs and signatures
		// landed. Not a regression in the parse: reaching further into the
		// file exposed `ADJUST { ... }`, a class phaser in no table, twice
		// more. Verified by stashing the change and counting -- 3 Unknowns
		// at HEAD, 5 after, and every new one is an ADJUST block. Untracked:
		// nothing in the chain owns ADJUST.
		"class/destruct.t": 7, "class/field.t": 7, "class/gh22169.t": 5,
		"class/gh23511.t": 1, "class/inherit.t": 7, "class/method.t": 9,
		"class/phasers.t": 5,
		"cmd/mod.t":       1, "cmd/subval.t": 1, "cmd/switch.t": 2,
		"comp/colon.t": 25, "comp/decl.t": 3, "comp/filter_exception.t": 5,
		"comp/final_line_num.t": 1, "comp/fold.t": 14, "comp/form_scope.t": 17,
		"comp/hints.t": 36, "comp/line_debug.t": 4, "comp/multiline.t": 2,
		"comp/opsubs.t": 11, "comp/package.t": 7, "comp/package_block.t": 4,
		"comp/parser.t": 65, "comp/parser_run.t": 12, "comp/proto.t": 41,
		// comp/require.t went 11 -> 12 when phaser braces became blocks. The
		// `BEGIN { ... }` body is now read as statements rather than as one
		// hashref, and reaching inside it exposes a heredoc the parser does
		// not yet read. Measured by toggling the phaser table alone: 20
		// Unknown nodes over 6,947 bytes without it, 12 nodes over 8,238
		// with. Fewer refusals covering MORE bytes, which is the one
		// direction the node count and the byte count disagree -- recorded
		// rather than argued away. T1-easy is unmoved at 52.7% either way,
		// so the trade is a correct OpensBlock flag at no rate cost.
		// base/lex.t 36 -> 37, comp/parser.t 64 -> 65 and
		// comp/retainedlines.t 17 -> 19 when dereferences became several
		// tokens rather than one. Each holds a SYMBOL-TABLE deref, which is
		// the one shape that change did not reach:
		//
		//	is $::{"_<hash-line-eval"}, ...
		//
		// `$::{` is not a dereference failure: measured, it lexes as
		// `Variable "$:"` then `Operator ":"`, because leadingPackageSeparator
		// wants a word byte after `::` and finds a brace. Splitting the sigil
		// reaches the call around it, so the refusal lands in more pieces
		// without the underlying gap moving at all. Bytes barely
		// move -- base/lex.t 1,111 -> 1,114, retainedlines 425 -> 427 -- and
		// comp/parser.t FELL, 4,869 -> 4,818. TestDerefBlockContentsAreDecided
		// names this shape as still opaque and asserts it still hedges.
		"comp/redef.t": 21, "comp/require.t": 12, "comp/retainedlines.t": 19,
		"comp/uproto.t": 3, "comp/use.t": 11, "comp/utf.t": 3,
		"opbasic/arith.t": 179, "opbasic/cmp.t": 6, "opbasic/concat.t": 6,
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
