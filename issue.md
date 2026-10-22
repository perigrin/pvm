---
title: Classify dropped source parens by position, then write a Paren node's parens once
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5e5e-73cc-a210-0341ea517914
created: 2026-10-22T23:46:33.790255282Z
updated: 2026-10-22T23:46:59.257136471Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/fidelity.go`, `internal/parse/fidelity_test.go`, `internal/parse/canon_test.go`, the parser files a tree-defect fix touches (e.g. `internal/parse/term.go`), `internal/parse/testdata/t1fidelity.ratchet`, `internal/parse/testdata/t1canon.ratchet`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^(TestDroppedParenPositionsAreClassified|TestCanonWritesParenOnce|TestFaithfulForgivesOnlyExtraNesting|TestCanonCatchesMisgrouping|TestCanonQuotientIsNarrow)$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (~344, the `List`/`Paren` case), `internal/parse/fidelity.go` (`Faithful`, `matchParen`), `internal/parse/term.go` (`finishList`, ~723)

117 of the 169 M3-dir re-emission failures are `source has "(", canon has <something else>`: a source paren canon does not write. Examples are `(shift)->SUPER::foo` (mro/basic_05_c3.t:32), `ok(($y eq '...'), ...)` (mro/c3_with_overload.t:40) and `if (($Config{...} !~ /.../) )` (lib/commonsense.t:12). canon writes a source paren only for a `List`, and `Paren` is a bool.

**perigrin's ruling (2026-10-22, revised after measurement).** Measured on perl 5.42.0 with B::Concise. Where parens change meaning, perl builds different ops: `my ($x) = @_` is aassign but `my $x = @_` is padsv_store, and `($x) x 3` is a list repeat (`lK/DOLIST`) but `$x x 3` is a string repeat. Where they do not change meaning, perl keeps the paren as one boolean bit, OPf_PARENS (`P`), on the op: `my $z = ($x)` is `padsv sP` vs `padsv s`. `(($y eq "x"))` and `($y eq "x")` are identical trees. So **canon writes a `Paren` node's parens ONCE**, and **`Faithful` forgives ONLY extra nesting**: a source paren pair nested directly inside another around the same node (`((x))` vs `(x)`). Before anything is forgiven, the 117 failures are classified by position. A failure in a context-changing position (before `=`, before `x`, after `my`/`our`/`local`/`state`) is a tree defect, where psc read a list as a scalar, and it is fixed, not forgiven. This supersedes the earlier same-day ruling that `Faithful` forgive redundant parens.

## Steps

- [ ] Measure with perl first. Run `perl -MO=Concise` on `my ($x) = @_` vs `my $x = @_`, `($x) x 3` vs `$x x 3`, `($x) = f()`, `() = f()`, `local ($x) = 1`, `my $z = ($x)` and `ok((($y eq "x")))`. Quote the output in the tests' comments.
- [ ] Classify: list every one of the 117 dropped-paren failures with its position class (context-changing: before `=`, before `x`, after my/our/local/state; or not). Put the list in the commit message.
- [ ] RED: write TestDroppedParenPositionsAreClassified. For each context-changing failure, reduce it to a fixture asserting psc's tree matches perl's ops (a list where perl builds a list). Run it and watch any tree defect fail.
- [ ] GREEN: fix each tree defect in the parser, one at a time, running the test after each. If there are none, the test still pins the measured positions as lists.
- [ ] RED: write TestCanonWritesParenOnce in canon_test.go. A `Paren` node on a non-`List` (`(shift)->foo`, `my $z = ($x)`, `ok(($y eq "x"))`) re-emits with exactly one pair, and `((x))` re-emits as `(x)`. Run it and watch it fail.
- [ ] GREEN: make canon write a `Paren` node's parens once. Run it and watch it pass.
- [ ] RED: write TestFaithfulForgivesOnlyExtraNesting in fidelity_test.go. `((x))` in the source against canon's `(x)` is faithful, and so is `(((x)))`. A source pair canon does not write at all (no `Paren` on the node) is still a difference. Run it and watch it fail.
- [ ] GREEN: implement the forgiveness in `Faithful`, limited to a pair nested directly inside another pair around the same node. Run it, TestCanonCatchesMisgrouping and TestCanonQuotientIsNarrow, and watch them all pass.
- [ ] Re-baseline t1canon.ratchet, t1fidelity.ratchet and perlgittfaithful.ratchet in the same commit. The message names each moved file and why it moved; a file going from faithful to unfaithful must be explained by a tree-defect fix or it blocks the commit.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the Concise measurements and writes NEW adversarial probes outside the corpus (e.g. `my ($a, $b) = ...` vs `my $a, $b`, `(@a) x 2`, `local ($_) = ...`, `return (1)`, `print ((1+2)*3)`, a paren around a ternary arm). Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Every context-changing dropped-paren position parses as perl's ops do (`go test ./internal/parse/ -run '^TestDroppedParenPositionsAreClassified$' -count=1 -v | grep -q '^--- PASS: TestDroppedParenPositionsAreClassified'`)
- [ ] canon writes a Paren node's parens exactly once (`go test ./internal/parse/ -run '^TestCanonWritesParenOnce$' -count=1 -v | grep -q '^--- PASS: TestCanonWritesParenOnce'`)
- [ ] Faithful forgives only extra nesting around the same node (`go test ./internal/parse/ -run '^TestFaithfulForgivesOnlyExtraNesting$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivesOnlyExtraNesting'`)
- [ ] canon still catches misgrouping (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)
- [ ] The forgiven quotient stays narrow (`go test ./internal/parse/ -run '^TestCanonQuotientIsNarrow$' -count=1 -v | grep -q '^--- PASS: TestCanonQuotientIsNarrow'`)
- [ ] The T1 canon ratchet matches its re-baselined file (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)
- [ ] The T1 fidelity ratchet matches its re-baselined file (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
- [ ] The perl.git t/ faithful ratchet records the gain (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)
