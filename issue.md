---
title: Write list operators, calls and niladic builtins by precedence, with the source's call parens recorded, behaviour-neutral
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a1690b-4745-7ef7-9285-0a9f97d86924
blocks:
- 01a1692f-bf96-703b-b6bc-59a216283227
created: 2026-10-23T14:34:36.382741984Z
updated: 2026-10-23T23:07:18.396128832Z
---

## Context

- paths: `internal/parse/canon.go` (the `Call` case for list operators, user and method calls, niladic builtins, and the `Indirect` branch), `internal/parse/call.go` (`ArgParen` on the list-op and user-call paren paths; the `ShapeNiladic` path at ~177), `internal/parse/indirect.go` (`ArgParen` on the indirect paren path), the compile-token sites that build `__FILE__`, `__LINE__`, `__SUB__`, `__PACKAGE__` and `__CLASS__` Call nodes, `internal/parse/listop_test.go` (new), `internal/parse/testdata/m3probes/listop-canon.txt` (new), the ~96 test expectations that pin a parenthesised list-op or call spelling (pushback's grep, round 8: `print(`, `push(`, `join(`, `ok(`, `is(`, `new(`, `bless(` and so on across the ~100 files that call `Canon`), `internal/parse/testdata/t1canon.ratchet`, `internal/parse/testdata/t1fidelity.ratchet`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^(TestCanonWritesListOpByPrecedence|TestCanonWritesNiladicBare|TestCanonListOpWrapsLooserOperand)$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (`emit`, the `Call` case: "Always parenthesised, and that is the whole case chapter 7 names"), `internal/parse/precedence.go` (`bpListOp`)

**perigrin's ruling (2026-10-23, chain-review round 7): canon is precedence-aware, and the call-paren forgiveness rule is removed.** This is the middle of the three issues that follow the named-unary one (P29 split, round 8). It is BEHAVIOUR-NEUTRAL, like the named-unary issue: the call-paren rule stays on, so no file can drop, and the ratchets prove it. It writes list operators, user, method and indirect calls, and niladic builtins by precedence, with the source's call parens recorded in `ArgParen`. The measurement issue before it has already shown the design's counts (fresh5's baseline: 321/619 and 169/347 with the rule off, before the recoveries below); the rule-removal issue after it takes the rule out and asserts the verdicts.

**The design, extending the named-unary issue's.**
- `ArgParen` is set on every Call whose source wrote `NAME(`: list-op builtins (call.go's `ShapeList` path), user and method calls, and indirect-object calls (indirect.go). A Call without it is written `NAME LIST`, the list emitted at `bpListOp+1` (40 at 054e75e6: `or` 10, `not` 30, `,` and `=` above it), with the follow power deciding whether an operator after the call would bind into its list. A looser infix operand (`or`, `and`, `xor`) inside the list gets wrapped, so a wrong extent is a paren the source lacks. A prefix `not` operand is not wrapped (`print not $x` is `print((!$x))`; the measurement issue holds it). A list op or call with no arguments writes the bare name.
- **Niladic builtins and compile-time tokens (K13/M2, round 8).** The 32 `ShapeNiladic` names in `coreShapes()` (`time`, `fork`, `wait`, `wantarray`, `getppid`, `times`, `__FILE__`, `__LINE__`, `__SUB__`, `__PACKAGE__`, `__CLASS__`, ...) are Call nodes built at call.go:~177 and at the compile-token sites, not through `parseWordTerm`'s paren path, and canon writes each `NAME()`. Measured at 054e75e6: `time;` is a bare Call written `time()`; `time();` is an Index node over the Call. So a niladic Call is written as its bare name, and the only pair ever written is the source's own, which the tree already records (the Index node; or `ArgParen`, if the executor moves that pair onto the Call). That recovers fresh5's 17 files (io/fflush.t, io/pipe.t, io/socket.t, op/sleep.t, op/sselect.t, op/getppid.t, op/sigsystem.t, op/context.t, op/current_sub.t, op/decl-refs.t, op/numify_chkflags.t, cmd/subval.t, class/class.t, class/field.t, class/inherit.t, class/phasers.t, win32/signal.t). `__PACKAGE__->new` keeps its arrow form.
- The block forms stay as the source wrote them: `print {$fh} "x"`, `sort { $a <=> $b } @a`, `map { ... } @a`, `grep { ... } @a` and the resolved `WORD BLOCK LIST` forms canon already handles bare for an unresolved word. The `eval`/`do` BLOCK forms are unchanged.
- The `Indirect` branch writes `new Foo "a"`, `find PerlIO::Layer 'perlio'` and `import Config` as the source did; its comment, which leans on the forgiven parens, is rewritten. `require MODULE` keeps its existing bare path.
- **A bare argument list that starts with `(` (M4, round 8).** perl reads `NAME (` as call parens, so a wrong tree whose first argument's emission begins with a `(` that does not enclose the whole list would re-emit the source's tokens exactly: source `print ($a), $b` (perl: `print($a), $b`, the parens end the list) with the wrong tree `print` over `($a), $b` would emit `print ($a), $b`. Canon therefore writes the call parens whenever the list's emission begins with `(` and that pair does not close at the list's end: the wrong tree emits `print(($a), $b)`, a paren the source lacks.
- Tests that pin the old parenthesised spelling are updated to the precedence spelling, each with the perl evidence; the commit message lists them (pushback counted about 96 expectation lines; the executor counts them exactly, as the named-unary issue listed its 11).
- **No test here asserts "unfaithful" (M1).** While the rule exists, `push (@a , 1 or die)` from a wrong tree is still forgiven. The wrong-tree tests assert canon's TEXT; the rule-removal issue asserts the verdicts. "Emits exactly" means the significant tokens of canon's output (through `significantTokens`, the comparison Faithful makes) equal the expected string's, not the raw bytes: canon writes ` , ` between arguments (`push(@a , 1)` today), and the tests must not pin that spacing (fresh7's note, round 8).
- A user sub with a `($)` prototype the file cannot see is parsed and written as a list op here; that is M4 territory (round-7 ruling). The file-visible case is a named unary by the named-unary issue.

## Steps

- [ ] Measure with the pinned perl 5.44.0: `perl -MO=Deparse,-p` on `push @a, 1 or die` (`(push(@a, 1) or die)`), `print $x, 1 and f`, `sub foo; sub bar; foo 1, bar 2, 3` (`foo(1, bar(2, 3))`), `new Foo "a"` with `package Foo` declared, `print {$fh} "x"`, `sort { $a <=> $b } @a`, `print ($a), $b`, `time`, `time()`, `print __FILE__`, `fork;` and `__PACKAGE__->new`. Quote the output in the tests' comments.
- [ ] Create `internal/parse/listop_test.go` with two `// ABOUTME:` lines.
- [ ] RED: write TestCanonWritesListOpByPrecedence: `print "x"`, `push @a, 1 or die`, `foo 1, bar 2, 3` (declared), `print {$fh} "x"`, `sort { $a <=> $b } @a`, `new Foo "a"`, `find PerlIO::Layer 'perlio'`, `shift @a` and `print ($a), $b` each re-emit their source spelling unchanged, and `print("x")` keeps its parens. Run it and watch it fail.
- [ ] RED: write TestCanonWritesNiladicBare: `time;`, `fork;`, `wantarray;`, `print __FILE__;`, `print __PACKAGE__, "x";` and `__PACKAGE__->new` re-emit unchanged, and `time();` keeps its `()`. Run it and watch it fail (`time()`).
- [ ] RED: write TestCanonListOpWrapsLooserOperand: hand-built wrong trees with `push` over `or`, `foo(1, bar(2), 3)` and `print` over `($a), $b` emit exactly `push (@a, 1 or die)`, `foo 1, bar (2), 3` and `print(($a), $b)`. Canon text only (M1). Run it and watch it fail.
- [ ] GREEN: set `ArgParen` on the list-op, user-call, method-call and indirect paren paths; write a Call without it as `NAME LIST` at `bpListOp+1` with the follow power and the leading-`(` check; write a niladic Call bare; keep the block forms bare; rewrite the `Indirect` branch and its comment. Run the three tests and watch them pass, with the rule still in place.
- [ ] Update every test that pins the old parenthesised spelling, with the perl evidence in each, and list them in the commit message. Run `go test ./internal/parse/` and watch it pass.
- [ ] Run TestCanonRatchet, TestCanonFidelityRatchet, TestParseRatchet, TestPerlGitTRatchet and TestPerlGitTFaithfulRatchet. With the rule on, no file can drop; a file that does is a defect to fix before committing, never a re-baseline. A file that became faithful is a gain: re-baseline in the same commit and explain each moved file (the block-form rows and the indirect calls are the expected gains).
- [ ] Add the measured cases, with perl's facts and a `faithful` row each, as probe file `internal/parse/testdata/m3probes/listop-canon.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus (a list op as the last argument of another, `print $fh @list`, `return wantarray ? @a : $a[0]`, `sort foo @a`, `sort keys %h`, a user call before `or` and before `,`, `time + 1`, `__LINE__ . "x"`), and adds them to this issue's probe file so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] Run `make test`, then commit.

**Measured by NL-a (exec/nla bc947a68, 2026-10-23); the measure step quotes this.** These numbers were measured in a scratch worktree at f09a07a0. The design writes list ops, user, method and indirect calls, and niladics bare when there is no ArgParen; writes `not` by follow; and turns the call-paren rule off. ArgParen must also be set on `->name(` and on the indirect `(`.

| config | perl.git t/ | M3 | statements (73,670) |
|---|---|---|---|
| baseline | 342/619 | 178/347 | 67,012 |
| rule off alone | 75 | 49 | 50,935 |
| design | 348 (1 lost, 7 gained) | 183 | 67,183 (3 lost, 174 gained) |
| design + trailing-comma variant | 349 (0 lost) | 184 | 67,186 (0 lost) |

- **This issue must adopt the trailing-comma variant:** a bare call whose last argument carries the source's trailing comma stays bare.
  - op/utf8cache.t:70 shows why. The source is `substr $x, 0, 1, = "\x{100}"`, and perl reads it as `substr($x,0,1) = ...`.
  - The trailing comma ends the list before `=`, `.=`, `||=`, `&&=`, `==` and `||`. It does not end it before `-=`, `*=`, `%=`, `/=`, `&=`, `**=`, `x=`, `//=` or `<<=`.
  - The variant also recovers io/open.t:281.
- **The stop condition did not fire.** fresh5's 23 lost files are all faithful under the design except op/utf8cache.t.
- **NL-a pins `print(not $x);` and `print(not $x, $y);`** because list ops are still parenthesised. This issue updates those pins to the bare form.

The data files are in docs/measurements/2026-10-23-precedence-aware-canon/ (nl-*).

## Acceptance criteria

### Positive Scenarios

- [ ] canon writes a list operator, user, method or indirect call by precedence, re-emitting each bare source spelling unchanged and keeping source parens (`go test ./internal/parse/ -run '^TestCanonWritesListOpByPrecedence$' -count=1 -v | grep -q '^--- PASS: TestCanonWritesListOpByPrecedence'`)
- [ ] canon writes a niladic builtin or compile-time token bare, and keeps the source's own `()` (`go test ./internal/parse/ -run '^TestCanonWritesNiladicBare$' -count=1 -v | grep -q '^--- PASS: TestCanonWritesNiladicBare'`)
- [ ] The measured cases are committed as perl-judged probes, held by TestM3ProbesHeldToPerl (`sh -c 'test -s internal/parse/testdata/m3probes/listop-canon.txt && go test ./internal/parse/ -run "^TestM3ProbesHeldToPerl$" -count=1 -v | grep -q "^--- PASS: TestM3ProbesHeldToPerl"'`)
- [ ] No file was lost: the perl.git t/ faithful count is at or above faithful_start_all in m3start.txt (`sh -c 's=$(sed -n "s/^faithful_start_all = \([0-9][0-9]*\)$/\1/p" internal/parseoracle/testdata/ratchet/m3start.txt); n=$(sed -n "s|^# \([0-9][0-9]*\)/619 faithful.*|\1|p" internal/parse/testdata/perlgittfaithful.ratchet); [ -n "$s" ] && [ -n "$n" ] && [ "$n" -ge "$s" ]'`)
- [ ] No M3-dir file was lost: the M3-dir faithful count is at or above faithful_start_m3 (`sh -c 's=$(sed -n "s/^faithful_start_m3 = \([0-9][0-9]*\)$/\1/p" internal/parseoracle/testdata/ratchet/m3start.txt); n=$(awk "\$1==\"0\" && \$2 ~ /^(op|mro|uni|lib|test_pl)\// {k++} END {print k+0}" internal/parse/testdata/perlgittfaithful.ratchet); [ -n "$s" ] && [ "$n" -ge "$s" ]'`)
- [ ] Regression guard (passes today and after any explained re-baseline): the T1 canon ratchet (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)
- [ ] Regression guard (passes today and after any explained re-baseline): the T1 fidelity ratchet (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
- [ ] Regression guard (passes today and after any explained re-baseline): the perl.git t/ faithful ratchet (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)
- [ ] Regression guard: canon still catches misgrouping (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)
- [ ] Regression guard: the forgiven quotient stays narrow (`go test ./internal/parse/ -run '^TestCanonQuotientIsNarrow$' -count=1 -v | grep -q '^--- PASS: TestCanonQuotientIsNarrow'`)

### Negative Scenarios

- [ ] A wrong list extent is a paren the source lacks in canon's text: `push` over `or` emits `push (@a, 1 or die)`, `foo(1, bar(2), 3)` emits `foo 1, bar (2), 3`, and `print` over `($a), $b` emits `print(($a), $b)`; the verdicts are asserted once the rule is gone, not here (`go test ./internal/parse/ -run '^TestCanonListOpWrapsLooserOperand$' -count=1 -v | grep -q '^--- PASS: TestCanonListOpWrapsLooserOperand'`)
- [ ] An indirect-object call keeps its source spelling: `new Foo "a"` re-emits bare and `new Foo("a")` keeps its parens (`go test ./internal/parse/ -run '^TestCanonIndirectObjectKeepsSourceSpelling$' -count=1 -v | grep -q '^--- PASS: TestCanonIndirectObjectKeepsSourceSpelling'`)
- [ ] A source `NAME(` is recorded and kept: `print("x")`, `push(@a, 1)` and `foo(1)` keep their parens, and a tree with `ArgParen` cleared emits the bare form, which is a different token sequence (`go test ./internal/parse/ -run '^TestCanonListOpKeepsSourceArgParen$' -count=1 -v | grep -q '^--- PASS: TestCanonListOpKeepsSourceArgParen'`)
