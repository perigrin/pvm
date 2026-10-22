---
title: "Round-5 review residue: suggestions, cuts and gaps pu shares"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-22T13:10:12.451980646Z
updated: 2026-10-22T15:18:58.65129162Z
---

crochet:review round 5 of PR #488 at 7201d0dd (2026-10-22). Quality report: scratchpad/review5-core-module-types-paad.md; probes scratchpad/r5probes/ (RAW-FINDINGS.md), r4/p/, r3probes/, r2probes/. perigrin, 2026-10-22: one narrow round, then close -- fix only these, with Opus, each reviewed against pu, r2 (2db182e6), h3 (692a560c) and h4 (7201d0dd) with perl 5.42 before merge, then a final quality lens over this round's diff alone; every other finding is a follow-up. A probe pu, r2, h3 or h4 gets right must still be right.

Filed by perigrin's ruling (2026-10-22: one narrow round, then close). From scratchpad/review5-core-module-types-paad.md: B3 (calls.go:522-536 `no warnings 'uninitialized'` ignored for call arguments under the flag), B4 (calls.go:191-192 `fmod(())` loses its arity error), A2 (symbols.go:296-310 a later nested `my @r` hides an earlier write; logicA/hs1.pl), A3 (ops.go:273-317 every user-sub/method argument counts as written -- pu shares; the precise fix belongs to 01a142c9), A4 (ops.go:309 chomp counted as a write of a reference), S1 (infer.go:247 + symbols.go:492-503 ClearElementShadows full scan per node: uc-16000.pl 13.1s vs 1.3s), St1 (nested foreach quadratic: 2.0s vs pu 0.38s at depth 1000), S2 (ternaries past the parser's depth limit, pu shares), C1 (Symbol.tpm:47 qualify_to_ref does not admit the Glob delete_package returns), E2 (ci.yml:56-58 stale comment), E3 (returntypes_test.go:248,:275 the pending-audit criterion cannot fail), D1 (docs/contributing/core-module-declarations.md:80-81 quotes removed Cwd.tpm lines). Coverage round 5: state/exit_die.pl (an always-exiting branch drops element shadows), state/s4anon.pl (a write in a never-run anon sub kills the shadow; 01a142c9 covers named subs), logicB/c106-c107 (no warnings not composed with excusesMismatch), and3000 memory (~2.3 GB) -- all pu-shared; contributing guide :274-279 ("provably-sound arity" silent, and no --warn-possible-mismatch) and :252-256 (Perl542 PATH fallback, PVM_REQUIRE_PERL542, FromPvm = built by pvm). Ponytail round 5 (~-26 lines): one inPlaceTargets for invalidateWrites/bodyWritesAlias; one userSub predicate for call/writtenArgs/callReturnsList; slices.Insert/Delete in put/drop; ClearElementShadows uses remove; clearArraysIn and invalidateElementShadow share arrayNamed. From the 01a163bb-c491 confirm review: always-run loops are reported as if they may run zero times -- `while (1) { ...; last }` (a10, the idiomatic one), `for my $i (1)` (a8), map over a non-empty literal (a7), scratchpad/rv491/; the mirror of A1, shared with pu and 692a560c. Narrow fix: mayNotRun excludes a constant-true while condition and a foreach over a non-empty literal list.

## Steps
- [ ] Fix or re-file each item; none left untracked

## Acceptance criteria

### Positive Scenarios
- [ ] TestReviewResidueTracked (`go test ./internal/infer/ -run '^TestReviewResidueTracked$' -count=1 -v | grep -q '^--- PASS: TestReviewResidueTracked'`)

### Negative Scenarios
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
