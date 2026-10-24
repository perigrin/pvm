---
title: Write not and list operators by precedence, measured first, and remove the call-paren forgiveness
state: done
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a1690b-1d10-7a84-be7e-0cbc31f36812
- 01a168d5-d519-79b1-ae4f-e988e96cf893
blocks:
- 01a1692f-65de-750b-aac5-f2506623ec7f
created: 2026-10-23T13:55:09.253494301Z
updated: 2026-10-24T00:39:46.32383615Z
sessions:
- start_sha: f09a07a0f6b66e2d02cddebd76850d5db05b1ffd
  end_sha: 5cb7269ab1e1b863f91500caf7c076350c33656d
  commits: 2
  started_at: 2026-10-23T21:54:27.16129015Z
  ended_at: 2026-10-24T00:39:46.32383615Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-23T21:54:27.16129015Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-24T00:39:46.32383615Z
observed_paths:
- docs/measurements/2026-10-23-precedence-aware-canon/README.md
- docs/measurements/2026-10-23-precedence-aware-canon/nl-design.diff
- docs/measurements/2026-10-23-precedence-aware-canon/nl-files.tsv
- docs/measurements/2026-10-23-precedence-aware-canon/nl-stmts.tsv
- docs/measurements/2026-10-23-precedence-aware-canon/nl-zzexport_test.go.txt
- docs/measurements/2026-10-23-precedence-aware-canon/nl-zzmeasure_test.go.txt
- docs/measurements/2026-10-23-precedence-aware-canon/nl-zzstmt_test.go.txt
- docs/specs/perl-parser/00-findings.md
- internal/parse/arrowhashread_test.go
- internal/parse/canon.go
- internal/parse/fidelity.go
- internal/parse/notop_test.go
- internal/parse/sites.go
- internal/parse/sites_decided_test.go
- internal/parse/testdata/m3probes/arrow-hash-read.txt
- internal/parse/testdata/m3probes/not-canon.txt
- internal/parse/testdata/perlgittfaithful.ratchet
- internal/parse/testdata/t1fidelity.ratchet
- internal/parseoracle/rv2hvfell_test.go
- internal/parseoracle/testdata/ratchet/baseline.txt
- internal/parseoracle/testdata/ratchet/m3start.txt
- internal/parseoracle/testdata/ratchet/sites.txt
---

**The title is stale and git-zhi cannot retitle.** Read it as "Measure the precedence-aware design for not, list ops, calls and niladics, and write not by the follow power". Split along its measure-first boundary in chain-review round 8 (P29): this issue measures and lands `not`; the list-op issue lands list ops, calls and niladic builtins with the rule still on; the rule-removal issue takes the rule out.

## Context

- paths: `docs/measurements/2026-10-23-precedence-aware-canon/` (the committed harness: README.md, measure-canon.diff, zz*_test.go.txt, *.tsv, fixtures.out; on branch feat/m3-oracle-at-scale until it merges), `internal/parse/canon.go` (the `Unary` case for `not`), `internal/parse/notop_test.go` (new), `internal/parse/testdata/m3probes/not-canon.txt` (new)
- commands: `go test ./internal/parse/ -run '^(TestCanonWritesNotByPrecedence|TestCanonNotWrapsLooserOperand)$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (`emit`, the `Unary` case: prefix ops use `outer`, not the follow power), `internal/parse/term.go` (~144, the FUNC1 `not (` form at `bpDeref`)

**perigrin's ruling (2026-10-23, chain-review round 7): canon is precedence-aware, and the call-paren forgiveness rule is removed.** The named-unary issue that blocks this one did the measured part. `not`, list ops, calls and niladic builtins were NOT measured by perigrin. **This issue measures the whole remaining design first, in a scratch worktree, before anything lands, and then lands only `not`.** The list-op issue and the rule-removal issue build on its numbers.

**The baseline, measured by fresh5 (round 8) on a scratch export of 054e75e6 with the named-unary diff applied** (K13/M2). Starting from the named-unary design and turning the call-paren rule off:

| configuration | perl.git t/ | M3 dirs |
|---|---|---|
| named unary by precedence, rule off, list ops still parenthesised | 74/619 | 49/347 |
| + list ops and calls written bare (args only) | 266 | 140 |
| + zero-argument calls written bare | 321 (23 lost, 2 gained) | 169 |

The 23 lost files, each by its first difference, and the fix that recovers it:
- **17 niladic builtins and compile-time tokens** written `fork()`, `time()`, `print __FILE__()`, and the same for `__LINE__`, `__SUB__`, `__PACKAGE__`, `__CLASS__`, `wantarray`: io/fflush.t, io/pipe.t, io/socket.t, op/sleep.t, op/sselect.t, op/getppid.t, op/sigsystem.t, op/context.t, op/current_sub.t, op/decl-refs.t, op/numify_chkflags.t, cmd/subval.t (which also needs NU's ternary middle-part rule, K14: its next difference is `eof F ? print "not ok $i\n" : print "ok $i\n"` at line 132), class/class.t, class/field.t, class/inherit.t, class/phasers.t, win32/signal.t. These Call nodes come from call.go:~177 (the `ShapeNiladic` path, 32 names in `coreShapes()`) and from the compile-token sites, not from `parseWordTerm`'s paren path. Measured at 054e75e6: `time;` parses as a bare Call and canon writes `time()`; `time();` parses as an Index node over the Call. The list-op issue writes a niladic Call bare, so the source's own `()` (the Index node) is the only pair ever written.
- **2 user subs with a visible unary prototype:** class/construct.t (`use builtin qw(blessed)` then `is(blessed $obj, ...)`) and mro/package_aliases.t (`frump brumkin` with `sub frump($)`). The named-unary issue sets `NamedUnary` through the `knowsShape` route for a visible `($)`, `(_)` or `(;$)` prototype and for a `builtin::` unary import, so these are recovered before this issue starts. (This corrects the earlier claim here that "the file-visible case is a named unary already": at 054e75e6 it is parsed as unary but canon still writes `one($x)`.) A prototype the file cannot see stays M4.
- **2 indirect-object calls** (`find PerlIO::Layer 'perlio'`, `import Config`): the list-op issue's `Indirect` branch writes them as the source did.
- **2 not triaged by fresh5:** op/utf8cache.t (`substr $x, 0, ...`) and uni/parser.t (a heredoc). This issue's measurement triages them.

**`not` (M5, round 8; fresh7's evidence at 054e75e6).** canon does NOT yet write `not` by precedence: `my $v = not $x or 1` re-emits `my $v = (not $x) or 1` (perl 5.44 Deparse,-p: `((my $v = (!$x)) or 1)`), and `not (1) + 1` re-emits `(not (1)) + 1`, because the `Unary` case reads `outer` and not the follow power. The fix is the follow power, as for named unaries: a prefix `not` (NOTOP, power 30) is written bare, its operand emitted at power 31 so a looser operand (`or`, `and`, `xor`) is wrapped: the wrong tree `not` over `or` emits `not ($x or 1)`. **`not` keeps the grouping guard canon has today (M5 hold, round 8):** when the follow power binds tighter than 30, canon wraps the whole `not` in grouping parens, EXCEPT when its operand is a Paren term, which is the FUNC1 `not (X)` form perl reads tighter. Measured at 5.44: source `not $x == 1` is `not(($x == 1))` to perl, so a wrong tree with `==` over `(not $x)` must emit `(not $x) == 1`, which today's canon writes and which a bare `not` would hide. The leading-`(` check (M4) applies to `not`'s operand too: source `not (1) + 1` is `(!1) + 1` to perl, so the wrong tree `not` over `(1) + 1` must emit `not((1) + 1)`, never `not (1) + 1`. `not` never gets call parens for any other reason and is never wrapped as an operand of a list op: `print not $x` is valid perl (`print((!$x))`) and stays bare; `not $x, $y` under `print` stays bare too, since `not` takes the whole comma list. The correct tree of `not (1) + 1` keeps its source parens through the Term's `Paren`. The wrong-tree cases were faithful only through `isName("not")`; their verdicts are asserted by the rule-removal issue, and here the tests assert canon's text only (M1). "Emits exactly" means the significant tokens of canon's output (through `significantTokens`, the comparison Faithful makes) equal the expected string's, not the raw bytes: canon writes ` , ` between arguments (`push(@a , 1)` today), and the tests must not pin that spacing (fresh7's note, round 8).

**Stop condition (M6).** The measurement scores every file (`zzmeasure_test.go`: `perlGitTFiles`, `perlGitTRoots`, `Session.ParseFileFrom`, skipping `perlRejects`) and every Statement node (`zzstmt_test.go`, 73,670 statements without heredocs at 054e75e6). If, under the full design with the rule off, any FILE or any STATEMENT goes from faithful to unfaithful and is not one of the 23 above recovered by the fixes named, the executor stops and brings perigrin the numbers, the file and statement names, and the first differences. The target is fresh5's 321/619 and 169/347 or better, with the 23 recovered.

## Steps

- [ ] MEASURE FIRST, before any change lands. In a scratch worktree at the commit the named-unary issue closed at, reproduce the committed harness (docs/measurements/2026-10-23-precedence-aware-canon/README.md and its zz*_test.go.txt) with: list ops, user and method calls, indirect calls and niladic builtins written bare without `ArgParen`; `not` by the follow power; and the call-paren rule off. Record baseline, rule-off-alone and design counts per file and per statement, and the first difference of every file or statement that changes verdict. Triage op/utf8cache.t and uni/parser.t. Apply the stop condition above. If it does not fire, quote the table in this issue's commit message and in the list-op issue's measure step.
- [ ] Measure with the pinned perl 5.44.0: `perl -MO=Deparse,-p` on `not $x or 1`, `my $v = not $x or 1` (`((my $v = (!$x)) or 1)`), `not (1) + 1` (`(!1) + 1`), `print not $x` (`print((!$x))`), `print not $x, $y`, and `not $x == 1`. Quote the output in the tests' comments.
- [ ] Create `internal/parse/notop_test.go` with two `// ABOUTME:` lines.
- [ ] RED: write TestCanonWritesNotByPrecedence: `not $x or 1`, `my $v = not $x or 1`, `not (1) + 1`, `print not $x`, `print not $x, $y` and `not $x == 1` each re-emit unchanged. Run it and watch it fail (`(not $x) or 1`).
- [ ] RED: write TestCanonNotWrapsLooserOperand: a hand-built wrong tree with `not` over `or` emits exactly `not ($x or 1)`. Canon text only (M1); the rule-removal issue asserts the verdict. Run it and watch it fail.
- [ ] RED: write TestCanonNotKeepsGroupingUnderTighterFollow (M5): the wrong tree `==` over `(not $x)` against source `not $x == 1` emits exactly `(not $x) == 1` (today's spelling, which must survive), and the wrong tree `not` over `(1) + 1` against source `not (1) + 1` emits exactly `not((1) + 1)`; the correct trees of both sources re-emit unchanged. Canon text only. Run it and watch the second case fail.
- [ ] GREEN: give the `Unary` case the follow power for `not`: bare with its operand at power 31; grouping parens around the whole `not` when the follow power is tighter than 30 unless the operand has Paren set (a parenthesised operand of any kind: the `not (X)` form, which perl reads tighter, e.g. `not ($a == 1) + 1` is `((!($a == 1)) + 1)`); call parens when the operand's emission begins with a `(` that does not close at its end. Run the three tests and watch them pass, with the rule still in place.
- [ ] Run TestCanonRatchet, TestCanonFidelityRatchet, TestParseRatchet, TestPerlGitTRatchet, TestPerlGitTFaithfulRatchet, TestCanonCatchesMisgrouping and TestCanonQuotientIsNarrow. A file that goes from faithful to unfaithful stops the issue; a gain is re-baselined in the same commit with each moved file explained.
- [ ] Add the measured cases, with perl's facts and a `faithful` row each, as probe file `internal/parse/testdata/m3probes/not-canon.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus (`not` in a ternary, `not not $x`, `!not $x`, `not` before `x`, `unless not $x`), and adds them to this issue's probe file so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] canon writes `not` by the follow power: `not $x or 1`, `my $v = not $x or 1`, `not (1) + 1` and `print not $x` each re-emit unchanged (`go test ./internal/parse/ -run '^TestCanonWritesNotByPrecedence$' -count=1 -v | grep -q '^--- PASS: TestCanonWritesNotByPrecedence'`)
- [ ] The measured cases are committed as perl-judged probes, held by TestM3ProbesHeldToPerl (`sh -c 'test -s internal/parse/testdata/m3probes/not-canon.txt && go test ./internal/parse/ -run "^TestM3ProbesHeldToPerl$" -count=1 -v | grep -q "^--- PASS: TestM3ProbesHeldToPerl"'`)
- [ ] No file was lost: the perl.git t/ faithful count is at or above faithful_start_all in m3start.txt (`sh -c 's=$(sed -n "s/^faithful_start_all = \([0-9][0-9]*\)$/\1/p" internal/parseoracle/testdata/ratchet/m3start.txt); n=$(sed -n "s|^# \([0-9][0-9]*\)/619 faithful.*|\1|p" internal/parse/testdata/perlgittfaithful.ratchet); [ -n "$s" ] && [ -n "$n" ] && [ "$n" -ge "$s" ]'`)
- [ ] Regression guard (passes today and after any explained re-baseline): the T1 canon ratchet (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)
- [ ] Regression guard (passes today and after any explained re-baseline): the T1 fidelity ratchet (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
- [ ] Regression guard (passes today and after any explained re-baseline): the perl.git t/ faithful ratchet (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)
- [ ] Regression guard: canon still catches misgrouping (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)

### Negative Scenarios

- [ ] A wrong tree's grouping is a paren the source lacks in canon's text: `not` over `or` emits `not ($x or 1)`; the verdict is asserted once the rule is gone, not here (`go test ./internal/parse/ -run '^TestCanonNotWrapsLooserOperand$' -count=1 -v | grep -q '^--- PASS: TestCanonNotWrapsLooserOperand'`)
- [ ] The grouping guard survives: `==` over `(not $x)` emits `(not $x) == 1` as today, and `not` over `(1) + 1` emits `not((1) + 1)`, never the source's `not (1) + 1` (`go test ./internal/parse/ -run '^TestCanonNotKeepsGroupingUnderTighterFollow$' -count=1 -v | grep -q '^--- PASS: TestCanonNotKeepsGroupingUnderTighterFollow'`)
- [ ] `not` is never wrapped as a list op's operand and never given call parens: `print not $x` and `print not $x, $y` re-emit unchanged (`go test ./internal/parse/ -run '^TestCanonNotStaysBareUnderListOp$' -count=1 -v | grep -q '^--- PASS: TestCanonNotStaysBareUnderListOp'`)
