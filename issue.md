---
title: Write not and list operators by precedence, measured first, and remove the call-paren forgiveness
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a1690b-1d10-7a84-be7e-0cbc31f36812
- 01a168d5-d519-79b1-ae4f-e988e96cf893
blocks:
- 01a1664d-75d4-76eb-aad2-10b322f366a9
created: 2026-10-23T13:55:09.253494301Z
updated: 2026-10-23T13:55:47.138611757Z
---

## Context

- paths: `internal/parse/canon.go` (the `Call` case for list operators and user calls, the `Indirect` branch, the `Unary` case for `not`), `internal/parse/call.go` and `internal/parse/indirect.go` (`ArgParen` on the list-op and user-call paren paths), `internal/parse/fidelity.go` (the `isName` call-paren rule and `matchForgivingSemicolons`, removed), `internal/parse/fidelity_test.go`, `internal/parse/listop_test.go` (new), `internal/parse/testdata/m3probes/listop-canon.txt` (new), `internal/parse/testdata/t1canon.ratchet`, `internal/parse/testdata/t1fidelity.ratchet`, `internal/parse/testdata/perlgittfaithful.ratchet`, `internal/parseoracle/testdata/ratchet/m3start.txt` (read)
- commands: `go test ./internal/parse/ -run '^(TestCanonWritesListOpByPrecedence|TestCanonWritesNotByPrecedence|TestFaithfulCallParenRuleIsGone|TestFaithfulForgivenessIsClosed)$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (`emit`, the `Call` case: "Always parenthesised, and that is the whole case chapter 7 names"), `internal/parse/fidelity.go` (`faithfulTokens`, `isName`), `internal/parse/precedence.go` (`bpListOp`)

**perigrin's ruling (2026-10-23, chain-review round 7): canon is precedence-aware, and the call-paren forgiveness rule is removed.** The named-unary issue that blocks this one did the measured part. This issue extends canon to `not` and to list operators, which perigrin did NOT measure, and then removes the rule, which is where any loss would show. **The issue measures first, the same way (scratchpad/m3/measure-canon.md), before landing anything. If list ops lose faithful files, the executor stops and brings perigrin the numbers.**

**What the scratch measurement found about these two, with named unary ops done and the rule turned off for them only.**
- `not` is a prefix Unary (NOTOP, power 30), not a Call, and canon already writes it by precedence: the correct tree of `not $x or 1` emits `not $x or 1`, and the wrong tree, `not` over `or`, emits `not ($x or 1)`. That wrong tree was still FORGIVEN, because `isName("not")` is true. Removing the rule closes it. The FUNC1 form `not (1) + 1` (term.go:~144 reads it at `bpDeref`) is measured here, with its Deparse,-p grouping `(!1) + 1`.
- List ops were left unchanged there. canon writes every list op and user call parenthesised (`print "x"` as `print("x")`, `push @a, 1` as `push(@a , 1)`, `new Foo "a"` as `new Foo("a")`), and today each is faithful only through the rule. Two wrong trees were still forgiven: `push @a, 1 or die` with `push` over `or` emitted `push(@a , 1 or die())`, and `foo 1, bar 2, 3` (both subs declared) as `foo(1 , bar(2) , 3)`, which perl groups `foo(1, bar(2, 3))`.

**The design, extending the named-unary issue's.**
- `ArgParen` is set on every Call whose source wrote `NAME(`: list-op builtins (call.go's `ShapeList` path), user and method calls, and indirect-object calls (indirect.go). A Call without it is written `NAME LIST`, the list emitted at `bpListOp+1` (40 at 054e75e6: `or` 10, `not` 30, `,` and `=` above it), with the follow power from the named-unary issue deciding whether an operator after the call would bind into its list. A looser operand (`or`, `and`, `xor`, `not`) inside the list gets wrapped, so a wrong extent is a paren the source lacks. A list op or call with no arguments writes the bare name.
- The block forms stay as the source wrote them: `print {$fh} "x"`, `sort { $a <=> $b } @a`, `map { ... } @a`, `grep { ... } @a` and the resolved `WORD BLOCK LIST` forms canon already handles bare for an unresolved word. The `eval`/`do` BLOCK forms are unchanged.
- The `Indirect` branch writes `new Foo "a"` when the source did; its comment, which leans on the forgiven parens, is rewritten.
- `require MODULE` keeps its existing bare path.
- **The rule goes.** `faithfulTokens` loses the `isName` branch and `matchForgivingSemicolons`. The last-`;` rule in the main loop must then cover a block argument on its own (`print {$fh} "x"` is emitted `print {$fh;} "x"`), which the inner match used to handle one level down. TestFaithfulForgivenessIsClosed's enumeration becomes `last-semicolon` alone at this commit; the Paren-once issue adds `extra-nesting`.
- A user sub with a `($)` prototype the file cannot see is M4 territory (round-7 ruling); the file-visible case is a named unary already, by `knowsShape`'s prototype.
- Tests that pin the old parenthesised spelling (`print("x")`, `push(@a , 1)`, and the indirect `new Foo("a")`) are updated to the precedence spelling, each with the perl evidence. The commit message lists them; the executor counts them when the measurement runs, as the named-unary issue listed its 11.

**Why the rule removal lives here and not in the Paren-once issue.** The measurement's "0 files move" is a property of canon with the rule OFF. The rule can only come out once every Call canon parenthesises without a source paren is gone, and that is this issue's last step. The Paren-once issue then starts from a Faithful with one rule and a canon that writes no paren the tree does not imply.

## Steps

- [ ] MEASURE FIRST, before any change lands. In a scratch worktree at the commit the named-unary issue closed at, reproduce scratchpad/m3/measure-canon.md's method (its `zzmeasure_test.go` walks perl.git t/ with `perlGitTFiles`, `perlGitTRoots` and `Session.ParseFileFrom`, skipping `perlRejects`; `zzstmt_test.go` scores every Statement node) with list ops and `not` written by precedence and the call-paren rule off: baseline, rule off alone, design. Record the per-file and per-statement counts, and the first difference of every file that changes verdict. **If any file goes from faithful to unfaithful under the design, stop and bring perigrin the numbers and the first differences; do not land anything.** If none does, quote the table in this issue's commit message.
- [ ] Measure with the pinned perl 5.44.0: `perl -MO=Deparse,-p` on `push @a, 1 or die` (`(push(@a, 1) or die)`), `print $x, 1 and f`, `sub foo; sub bar; foo 1, bar 2, 3` (`foo(1, bar(2, 3))`), `not $x or 1`, `not (1) + 1` (`(!1) + 1`), `new Foo "a"` with `package Foo` declared, `print {$fh} "x"`, and `sort { $a <=> $b } @a`. Quote the output in the tests' comments.
- [ ] Create `internal/parse/listop_test.go` with two `// ABOUTME:` lines.
- [ ] RED: write TestCanonWritesListOpByPrecedence: `print "x"`, `push @a, 1 or die`, `foo 1, bar 2, 3` (declared), `print {$fh} "x"`, `sort { $a <=> $b } @a`, `new Foo "a"` and `shift @a` each re-emit their source spelling unchanged, and `print("x")` keeps its parens. Run it and watch it fail.
- [ ] RED: write TestCanonWritesNotByPrecedence: `not $x or 1`, `not (1) + 1` and `my $v = not $x or 1` re-emit unchanged, and a hand-built wrong tree with `not` over `or` emits `not ($x or 1)`, unfaithful. Run it; the correct cases may already pass, and the wrong-tree case fails while the rule forgives it.
- [ ] RED: write TestCanonListOpWrapsLooserOperand: hand-built wrong trees with `push` over `or` and `foo(1, bar(2), 3)` emit `push (@a, 1 or die)` and `foo 1, bar (2), 3`, and each is unfaithful against its source. Run it and watch it fail.
- [ ] GREEN: set `ArgParen` on the list-op, user-call, method-call and indirect paren paths; write a Call without it as `NAME LIST` at `bpListOp+1` with the follow power; keep the block forms bare. Run the three tests and watch them pass, with the rule still in place.
- [ ] RED: write TestFaithfulCallParenRuleIsGone in fidelity_test.go: through `parse.FaithfulTokens`, source `print "x"` against canon tokens `print ( "x" )` is unfaithful, and so is `defined $x` against `defined ( $x )`. Run it and watch it fail.
- [ ] GREEN: remove the `isName` branch and `matchForgivingSemicolons` from fidelity.go, make the last-`;` rule cover a block argument in the main loop, and drop `call-parens` from TestFaithfulForgivenessIsClosed's enumeration (its fixture moves to TestFaithfulCallParenRuleIsGone as the negative). Run TestFaithfulCallParenRuleIsGone, TestFaithfulForgivenessIsClosed, TestCanonCatchesMisgrouping and TestCanonQuotientIsNarrow, and watch them pass.
- [ ] Update every test that pins the old parenthesised spelling, with the perl evidence in each, and list them in the commit message. Run `go test ./internal/parse/` and watch it pass.
- [ ] Run TestCanonRatchet, TestCanonFidelityRatchet, TestParseRatchet, TestPerlGitTRatchet and TestPerlGitTFaithfulRatchet. Compare with the measurement: a file that went from faithful to unfaithful is the stop condition above, never a re-baseline. A file that became faithful is a gain: re-baseline in the same commit and explain each moved file (the block-form rows and the indirect calls are the expected gains).
- [ ] Add the measured cases, with perl's facts and a `faithful` row each, as probe file `internal/parse/testdata/m3probes/listop-canon.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus (a list op as the last argument of another, `print $fh @list`, `return wantarray ? @a : $a[0]`, `sort foo @a`, a user call before `or` and before `,`, `not` in a ternary), and adds them to this issue's probe file so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] canon writes a list operator or user call by precedence, re-emitting each bare source spelling unchanged and keeping source parens (`go test ./internal/parse/ -run '^TestCanonWritesListOpByPrecedence$' -count=1 -v | grep -q '^--- PASS: TestCanonWritesListOpByPrecedence'`)
- [ ] canon writes `not` by precedence in both its forms (`go test ./internal/parse/ -run '^TestCanonWritesNotByPrecedence$' -count=1 -v | grep -q '^--- PASS: TestCanonWritesNotByPrecedence'`)
- [ ] The forgiven set is exactly the rules in force, each with a perl-measured fixture, and call-parens is not among them (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosed'`)
- [ ] The measured cases are committed as perl-judged probes, held by TestM3ProbesHeldToPerl (`sh -c 'test -s internal/parse/testdata/m3probes/listop-canon.txt && go test ./internal/parse/ -run "^TestM3ProbesHeldToPerl$" -count=1 -v | grep -q "^--- PASS: TestM3ProbesHeldToPerl"'`)
- [ ] No file was lost: the perl.git t/ faithful count is at or above faithful_start_all in m3start.txt (`sh -c 's=$(sed -n "s/^faithful_start_all = \([0-9][0-9]*\)$/\1/p" internal/parseoracle/testdata/ratchet/m3start.txt); n=$(sed -n "s|^# \([0-9][0-9]*\)/619 faithful.*|\1|p" internal/parse/testdata/perlgittfaithful.ratchet); [ -n "$s" ] && [ -n "$n" ] && [ "$n" -ge "$s" ]'`)
- [ ] No M3-dir file was lost: the M3-dir faithful count is at or above faithful_start_m3 (`sh -c 's=$(sed -n "s/^faithful_start_m3 = \([0-9][0-9]*\)$/\1/p" internal/parseoracle/testdata/ratchet/m3start.txt); n=$(awk "\$1==\"0\" && \$2 ~ /^(op|mro|uni|lib|test_pl)\// {k++} END {print k+0}" internal/parse/testdata/perlgittfaithful.ratchet); [ -n "$s" ] && [ "$n" -ge "$s" ]'`)
- [ ] Regression guard (passes today and after any explained re-baseline): the T1 canon ratchet (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)
- [ ] Regression guard (passes today and after any explained re-baseline): the T1 fidelity ratchet (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
- [ ] Regression guard (passes today and after any explained re-baseline): the perl.git t/ faithful ratchet (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)
- [ ] Regression guard: canon still catches misgrouping (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)
- [ ] Regression guard: the forgiven quotient stays narrow (`go test ./internal/parse/ -run '^TestCanonQuotientIsNarrow$' -count=1 -v | grep -q '^--- PASS: TestCanonQuotientIsNarrow'`)

### Negative Scenarios

- [ ] The call-paren rule is gone: source `print "x"` against canon `print("x")`, and `defined $x` against `defined($x)`, are unfaithful through the comparator (`go test ./internal/parse/ -run '^TestFaithfulCallParenRuleIsGone$' -count=1 -v | grep -q '^--- PASS: TestFaithfulCallParenRuleIsGone'`)
- [ ] A wrong list extent is a paren the source lacks: `push` over `or` emits `push (@a, 1 or die)` and `foo(1, bar(2), 3)` emits `foo 1, bar (2), 3`, each unfaithful (`go test ./internal/parse/ -run '^TestCanonListOpWrapsLooserOperand$' -count=1 -v | grep -q '^--- PASS: TestCanonListOpWrapsLooserOperand'`)
- [ ] The last-`;` rule still covers a block argument without the removed inner match: `print {$fh} "x"` emitted `print {$fh;} "x"` is faithful, and `print {$fh} "x"` against `print {$fh;} ("x")` is not (`go test ./internal/parse/ -run '^TestFaithfulLastSemicolonCoversBlockArgument$' -count=1 -v | grep -q '^--- PASS: TestFaithfulLastSemicolonCoversBlockArgument'`)
- [ ] An indirect-object call keeps its source spelling: `new Foo "a"` re-emits bare and `new Foo("a")` keeps its parens; neither is forgiven against the other (`go test ./internal/parse/ -run '^TestCanonIndirectObjectKeepsSourceSpelling$' -count=1 -v | grep -q '^--- PASS: TestCanonIndirectObjectKeepsSourceSpelling'`)
