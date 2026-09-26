// ABOUTME: T2 core: 56 files across five perl5 t/ directories, ratcheted per file.
// ABOUTME: The T2 rate is no longer a gate; this names the per-file shortfall and ratchets it.

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

// TestT2CoreParses ratchets the per-file shortfall over the 56 T2 files.
//
// It NO LONGER CARRIES A RATE TARGET. The M1 gate's was 100% of 56 files
// against a measurement of 14, and the gate's own text said what to do
// about that: "Do not carry an unreachable 100% into a gate: that is how a
// gate becomes advisory." It was already advisory here -- prose and a log
// line, with nothing asserting it -- and the graded corpus is what replaces
// it. "Parses cleanly through tier N" is a claim; "100% of 56 files drawn
// from perl's own suite" names a finish line without saying what crossing
// it would mean. TestCorpusRatchet in internal/conformance is the gate now,
// per file and failing in both directions, over a corpus whose files are
// ordered by what they depend on.
//
// What survives here is the half that was always load-bearing: the
// shortfall map below. T2 is perl's own test suite, which the corpus does
// not replace -- the corpus says what the parser can read, and these files
// say what it still cannot.
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
		// FIVE files moved on one entry, `undef` joining `namedUnary`
		// (issue 01a0dd43): class/destruct.t 7 -> 5, comp/package.t 2 -> 1,
		// comp/parser.t 31 -> 29, comp/uproto.t 2 -> 1 and
		// opbasic/arith.t 21 -> 11. Each delta equals that file's count of
		// `undef` followed by an argument, counted per file rather than
		// inferred from the total: arith.t alone writes `undef $a;` and
		// `undef $s;` five times each, which is the ten. Recorded here
		// rather than beside each entry because it is one cause.
		//
		// `base/lex.t` 37 -> 36 and `base/num.t` 48 -> 44 when the lexer
		// learned that a `.` before a digit starts a NUMBER where a term
		// is expected (issue 01a0c13f). Measured before the fix: 5 of
		// those 85 Unknowns held a leading decimal, and all 5 are gone.
		// base/lex.t 36 -> 35, base/num.t 44 -> 6, comp/hints.t 36 -> 35
		// and comp/package.t 6 -> 2 when `startsTerm` stopped calling
		// every Word a term-starter (issue 01a0cf3e). `eq`, `ne`, `cmp`,
		// `lt`, `x` and `and` are Words to the lexer and OPERATORS to
		// perl, so both filehandle branches read `print FOO eq "x"` as a
		// print to FOO and stranded the comparison. num.t is dense with
		// word-spelled comparisons, which is why it moves furthest.
		//
		// base/lex.t 35 -> 26 when a heredoc body became a child of the
		// statement its opener sits in (issue 01a0c13f). Nine nodes, the
		// largest single move in this map, and the survey predicted it:
		// `base/lex.t` was named there as heredocs plus leading decimals,
		// and the decimals had already gone.
		//
		// Then 26 -> 27 when the lexer began forming `-l` as one operator
		// (issue 01a0cf64), and the RISE IS AN IMPROVEMENT of the same kind
		// class/field.t's was. `0-5x-l{0};` -- perl #123711's crash case at
		// byte 12557 -- used to build a tree that round-tripped and was
		// WRONG:
		//
		//	canon "0 - 5x - l(){0};"     a subscript on a call to sub `l`
		//	perl  "0 - 5 x (-l {0});"    repetition applied to a filetest
		//
		// It now refuses with trailing_tokens instead, which is the honest
		// answer. The blocker is NOT the filetest: `scanNumber` takes the
		// `x` into `Number(5x)`, so the `-l` never reaches an operator
		// position at all. That is the `(1)x3` gap 01a0ce57 measured and
		// left unfixed, and this is a second file where it costs a node.
		//
		// base/num.t 6 -> 3 and base/rs.t 3 -> 1 when a brace after ANY word
		// got intuit_curly's lookahead instead of only map, grep and sort
		// (issue 01a0d087). rs.t's movers are its seven `if (eval {$/ = ...;
		// 1})` guards, whose brace was classified from XTerm inside the paren.
		// What REMAINS in both is `$^O eq '...'`, which is a different gap.
		"base/lex.t": 27, "base/num.t": 3, "base/rs.t": 1,
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
		// at HEAD, 5 after, and every new one is an ADJUST block.
		//
		// class/gh22169.t 5 -> 4, class/inherit.t 7 -> 6 and
		// class/method.t 9 -> 8 with the heredoc body (issue 01a0c13f). One
		// node each: each file opens exactly one heredoc. `class/gh23511.t`
		// holds a `<<` too and did not move -- measured, its Unknown is
		// elsewhere.
		//
		// `ADJUST` joining both phaser tables took four of these, and only
		// files that contain one moved:
		//
		//	class/phasers.t   5 -> 0, left the map   13 ADJUST blocks
		//	class/destruct.t  7 -> 4                  2
		//	class/gh22169.t   4 -> 2                  5
		//	class/inherit.t   6 -> 4                  4
		//
		// The four class files that did NOT move -- construct.t, field.t,
		// gh23511.t and method.t -- hold no `ADJUST` at all, which is what
		// separates a ratchet that improved from one that merely moved.
		//
		// The unary `undef` then took `class/destruct.t` further, 4 -> 2.
		// The two fixes are independent and compose: ADJUST closed the
		// phaser shapes, `undef` the container spellings, and destruct.t
		// holds both.
		//
		// class/method.t 8 -> 6 when `no feature "signatures"` began turning
		// the feature off (issue 01a0dd6f). Both nodes are on line 33,
		// `method retnamed ( :$named = 456 )`, inside the
		// `no feature 'signatures'` block at line 28: the parens are a
		// Prototype token now rather than a mis-lexed signature.
		//
		// That is NOT a tree improvement. A `method` is signatured in perl
		// regardless -- which is what line 28 exists to test -- so the tree
		// is still wrong there; it has two fewer refusal SITES. Recorded
		// because a count falling for the wrong reason is the defect this
		// map keeps finding.
		//
		// Three fixes compose across this block and none of them overlaps
		// another's files: ADJUST took the phaser shapes, `undef` the
		// container spellings, and the signatures pragma one method.
		//
		// class/method.t 6 -> 3 when a brace after ANY word became a block
		// (issue 01a0d087). The three are `method priv { ... }` forms whose
		// NAME is followed by a brace: the lookahead was gated on
		// map/grep/sort, so the body read as a subscript and everything after
		// it fell to trailing_tokens. The three that remain are the same
		// declarations reached through a signature.
		// class/field.t 7 -> 5 with the labelled bare block (issue 01a0de8b),
		// in two steps, and NEITHER of its labels is on a block:
		//
		//	field $forwards  = do { goto HERE; HERE: 1 };
		//	field $backwards = do { my $x; HERE: ; goto HERE if !$x++; 2 };
		//
		// 7 -> 6 was `HERE: 1`, a label on a plain expression statement, which
		// that site was building by hand without its labels. 6 -> 5 was
		// `HERE: ;`, a label on an EMPTY statement, whose check ran before the
		// labels were read and so could not see one. perl accepts both and
		// Deparse emits `HERE: ;` back verbatim.
		"class/destruct.t": 2, "class/field.t": 5, "class/gh22169.t": 2,
		"class/gh23511.t": 1, "class/inherit.t": 4, "class/method.t": 3,
		// cmd/subval.t and comp/package_block.t left the map entirely with
		// the `startsTerm` fix -- they parse cleanly now, which is what
		// removal from this map means.
		"cmd/mod.t": 1, "cmd/switch.t": 2,
		// comp/colon.t 25 -> 0 and opbasic/magic_phase.t 7 -> 0 with the
		// in-file `sub NAME` declaration (issue 01a0c13f). Both left the map
		// entirely. Each declares its own `sub ok` and calls it without
		// parens throughout, which is what the measurement predicted: 25 and
		// 7 of the two files' refusals had a callee declared in the same
		// file, and those are the whole of each count.
		// comp/decl.t 3 -> 2, comp/hints.t 34 -> 31, comp/proto.t 4 -> 2 and
		// comp/require.t 7 -> 6 when a brace after ANY word got
		// intuit_curly's lookahead (issue 01a0d087). The shared mover is
		// `eval { ... }` inside an expression -- `eval {require 5.005}` and
		// `eval { prototype(...) }` -- whose brace was classified from XTerm
		// and read as an anonymous hash, so its `}` closed a subscript.
		"comp/decl.t": 2, "comp/filter_exception.t": 5,
		// comp/form_scope.t 17 -> 7 when parseFormatDecl landed. The file
		// holds nine format declarations, each of which had been swallowing
		// the statement that followed it: an unimplemented keyword runs
		// `skipToStatementEnd`, which takes the NEXT statement's `;` because
		// a format declaration has none of its own. Measured, the seven left
		// are all `trailing_tokens` on constructs this issue does not touch
		// -- `&$clo1(0)`, `make_closure 6` and a bare block in expression
		// position. Untracked: nothing in the chain owns those.
		// comp/fold.t 14 -> 6, comp/form_scope.t 7 -> 5,
		// comp/line_debug.t 4 -> 1, comp/parser.t 55 -> 31,
		// comp/proto.t 41 -> 4, comp/redef.t 21 -> 1, comp/require.t 10 -> 7,
		// comp/uproto.t 3 -> 2, comp/use.t 11 -> 1,
		// opbasic/arith.t 179 -> 21 and opbasic/concat.t 6 -> 5 with the
		// in-file `sub NAME` declaration (issue 01a0c13f).
		//
		// `opbasic/arith.t` is the largest move in this map's history, 158
		// nodes, and the measurement named it before the fix: 169 of its 179
		// refusals opened with a parenless call whose callee this file
		// declares, and `sub tryeq ($$$$)` accounts for 147 of them alone.
		// `comp/proto.t` 41 -> 4 is the same story with prototypes as its
		// SUBJECT rather than incidentally: the file exists to test them.
		//
		// Measured before the fix over all 56 files: 354 refusals open with a
		// bareword + argument, 316 of them with a callee declared by a
		// `sub NAME` in the same file and 0 with one reachable by import --
		// no T2 file uses Test::More, they `require './test.pl'`.
		"comp/final_line_num.t": 1, "comp/fold.t": 6, "comp/form_scope.t": 5,
		// comp/hints.t 35 -> 34 and comp/parser.t 65 -> 64 when a
		// DataSection became trivia (issue 01a0dc84). Each file ends in a
		// trailing `__END__` and each was spending exactly one Unknown on
		// it: the lexer already read the marker and everything after it as
		// one token, and the parser was handing that token to the
		// expression parser, which refused it as `not_a_term`. One node per
		// file is the whole delta -- nothing else in either file moved.
		"comp/hints.t": 31, "comp/line_debug.t": 1, "comp/multiline.t": 2,
		// comp/package.t 7 -> 2 across two fixes, both issue 01a0cf3e.
		// First the filehandle slot learned that a following COMMA
		// denies it, which is perl's own rule -- `print FOO, 1` is "No
		// comma allowed after filehandle" and `print __PACKAGE__, 1` is
		// legal. Then `startsTerm` learned that a word OPERATOR is not a
		// term, which reached `print __PACKAGE__ eq '...' ? ... : ...`.
		// comp/opsubs.t 11 -> 9 with the labelled bare block (issue
		// 01a0de8b). Its one label is `SILENCE_WARNING: {` at line 117;
		// before the fix that brace lexed as an anonymous hash and the two
		// statements after it fell to trailing_tokens.
		"comp/opsubs.t": 9, "comp/package.t": 1,
		// comp/parser.t 64 -> 55 and comp/parser_run.t 12 -> 5 with the
		// heredoc body (issue 01a0c13f) -- nine and seven nodes. parser_run.t
		// is the densest heredoc user in the map: more than half its
		// refusals were bodies.
		// parser_run.t 5 -> 2 with the labelled bare block (issue 01a0de8b).
		// Its `SKIP:` at line 73 and the `{` at line 74 are on separate
		// lines, which is why the fix carries the label across trivia rather
		// than requiring the colon to touch its brace.
		"comp/parser.t": 29, "comp/parser_run.t": 2, "comp/proto.t": 2,
		// comp/require.t went 11 -> 12 when phaser braces became blocks. The
		// `BEGIN { ... }` body is now read as statements rather than as one
		// hashref, and reaching inside it exposed a heredoc the parser did
		// not yet read -- which issue 01a0c13f then closed, taking the file
		// 12 -> 10. Measured by toggling the phaser table alone: 20
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
		// comp/retainedlines.t 19 -> 17 when the three word operators below
		// assignment became reachable (issue 01a0dbe8). ONE site earned
		// both nodes -- line 65's `(!$seen{$_} and /eval (\d+)/)`, the
		// file's only parenthesised word operator. The paren's element loop
		// stopped there because it parsed at the COMMA's power and levels 4
		// and 5 are below it. Back to the 17 the symbol-table deref note
		// above records, by a different route.
		"comp/redef.t": 1, "comp/require.t": 6, "comp/retainedlines.t": 17,
		"comp/uproto.t": 1, "comp/use.t": 1, "comp/utf.t": 3,
		"opbasic/arith.t": 11, "opbasic/cmp.t": 6, "opbasic/concat.t": 5,
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

	t.Logf("T2 core: %d of %d files clean (%.1f%%); %d in the shortfall map. "+
		"Reported, not targeted -- the tier corpus gates what the parser reads.",
		clean, len(files),
		100*float64(clean)/float64(len(files)), len(files)-clean)
}
