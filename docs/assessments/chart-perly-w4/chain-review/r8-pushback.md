# chain-review round 8 (the bound): pushback (plan quality) on chart-perly-w4

## Drafter: growth check

I diffed the live chain against my round-7 copy (`pb7/`) before reading `repair-r7.md`.

- **Changed bodies (3):** d218 (16 lines), d166 (9 lines) and 3262 (4 lines).
- **Milestone:** one new Records entry, "Chain-review round 7 repair".
- **Unchanged:** edges, titles and labels. There are still 19 issues.

| Edit | Answers |
|---|---|
| d218: names the six repointed reference differentials and the instruments that stay on the no-precedence mode; states the verdict-only rule for inputs the judged chart refuses; adds a perl.git step; adds a criterion running five of the six, each logging `verdict only: N`, with N > 0 for waypoint2 | r6/r7-fresh2 FR4-1 and r6/r7-fresh5 FR5-3 (blocking) |
| 3262: the pinned list names the repointed tests' later owner; the joint criterion's description stops claiming "no-precedence mode" for good (command unchanged) | the same findings. This is a knock-on: without it, 3262's text would contradict d218 at verify. |
| d166: a set-id lifetime statement, a TDD step and `TestNarrowedSetIdsStableAcrossFiles` | r6-fresh6 FR6-6 (my promotion) |
| Milestone Records | both |

**Verdict.** Every edit answers a blocking finding or is its required knock-on. Nothing grew past what the findings asked. `criteria-lint.sh` exits 0 over all 19 issues, so every criterion is still red at `e3432c0f`.

## 1. Recommendation

**Modify.** This is the bound, so the open items below go to perigrin. Each is a one-line fix.

Two blocking findings remain open:

1. **FR6-6 is held, narrowed.** d166's new prose states the invariant correctly. I measured the committed tables, though, and the criterion's fixture cannot fail on the defect it guards against (section 2).
2. **r7-fresh7 FR7-1 is promoted.** bb9b's item 8 tells the docs issue to write into the live README a claim that contradicts d0ba's and d218's forest-equality criteria (section 3).

Everything else I raised is released or residual.

## 2. Status of my findings

| id | Status | Grounds |
|---|---|---|
| PB1–PB11, PB16, PB17 | **released** | Unaffected by this round. PB17's guard and its callers are unchanged. |
| PB12–PB15, PB18–PB20 | residual | Unchanged. |
| N1, N5, N6, N7, N12, N14 | **released** | Unchanged. d218's new repoint text keeps N14's explicit form for toys ("the explicit form over each toy's own precedence"). |
| N2–N4, N8–N11, N13, N15–N17 | residual | Unchanged. N16 is the same finding as r7-fresh2 FR7-1 and r7-fresh5 FR7-2 (the unmeasured "13 less 2" count of feasible toys). N17 is the same as FR6-5 (`main_test.go`'s `resetKeys`). |
| FR2-8, FR3-7, FR4-1 (my promotions) | **released** | Unchanged. |
| **r6-fresh6 FR6-6** (my promotion) | **held, narrowed** | Details in the three paragraphs below this table. |

**What the fix had to do.** The fix had four parts. It had to state that a narrowed id keeps its contents for a recognizer's life, and it has: "An id is never reassigned to a different set". It had to add a criterion that **fails on the defect**. It had to add a TDD step. And it had to leave the mutation criterion intact. Three of the four landed. The second did not.

**What I measured.** The criterion's fixture is "two files that narrow to different terminal sets (one with a library operator in scope, one with it out of scope)". I measured what those two narrowings produce on the committed tables, using a `go test -overlay` file so the tree stayed untouched and `git status` stayed clean. Of the 1,880 sets, 62 carry a `PLUGIN_*` reading. Stripping every `PLUGIN_*` reading, which is the out-of-scope case, gives **a set already in the committed tables in 61 of 62 cases**. Keeping one reading, the in-scope case, gives a set outside them in 682 cases, which matches d166's 682.

**Why the criterion cannot fail.** The out-of-scope file therefore uses a committed id. The cheap defect is numbering new ids per file from 1,880 up. Under that defect, the in-scope file's new id never meets a second new id with different contents. So the cache never serves a wrong transition, and `TestNarrowedSetIdsStableAcrossFiles` passes on the defective implementation as well as the correct one. That is a criterion that passes whether or not its work exists, which makes it a defect, not a residual. **Fix:** both files must keep a library operator in scope, each narrowing a token to a *different* set outside the committed 1,880, for example two operators, or one operator with different kept readings. The test logs that both ids lie past the committed count and that their contents differ. |

## 3. Promotions

| Row | To | Grounds |
|---|---|---|
| **r7-fresh7 FR7-1** | **blocking** | **Two parts of the chain contradict, and the live README would carry a false claim.** bb9b's item 8 tells the docs issue to rewrite the "Known limits" paragraph so that "an alternative kept in one context is kept in all" applies to `Constrain` **only**, "since production's Build walks by context". The rest of the chain asserts the opposite. V2's forest keeps one node per (symbol, span), so a node reached under two *live* contexts keeps the union of their alternatives, as V1's does. That is exactly what d0ba's `TestForestEqualityUnionShortcut` requires V2 to equal ("one node holds two contexts with two different non-empty alternative sets"), and what d218 makes real. Walking by context removes only alternatives whose context's chain *dies* (C6), and V1's forest does not keep those either, because they are unreachable. So item 8, followed literally, writes into a live document a claim that the milestone's own criteria refute. `TestPrecedenceShortcutStated` reads the phrase from that very section, so the docs issue edits a guarded sentence on false grounds. I would not stand aside on this alone: the live layer must be true now, and bb9b exists to make it true. **Fix:** reword item 8 to say that the union holds for V2's production forest as for `Constrain`, being a property of the shared-node forest, and that walking by context drops only dead-context alternatives. |
| r7-fresh7 FR7-2 (3282's dead-context test "on the union-shortcut fixture", which d0ba defines with two live contexts) | not promoted | The builder has to add a dead-context site. Adding an input to the fixture file is a fixture change, not a criterion change. It need not touch d0ba's assertion either, which is logged over the node it constructs. At worst, 3282's first step reports it. Residual. Suggested action: let 3282 add its own input. |
| r7-fresh5 FR7-1 (the `ParseWith` guard when `New(tab)`'s derivation fails) | not promoted | A recognizer built by `New(tab)` is by construction "what `New(tab)` would derive", so a faithful guard accepts it, and `Recognize`'s error then becomes the refusal. Only a careless implementation, one that compares non-nil precedences, breaks `TestPrecedenceGenerationErrorRefusesEveryFile`, and that test is in `make test`. Residual. |
| r7-fresh6 FR7-2 (the brute force compares against `ConstrainReference`, not `Constrain`) | not promoted | The pinned divergence needs an empty rule with a non-allow verdict, which counter 5 makes infeasible, so a feasible toy avoids it. If one does not, the first step reports it. Residual. |
| r7-fresh7 FR7-9 (d199 size) and FR7-3 – FR7-8, FR7-10, FR7-11; AL18; r7-fresh4/6 FR7-1 (toy reuse across packages) | not promoted | Each is size, wording or test placement. None is a claim the chain makes wrongly, and each is caught at a first step or a close. FR7-6, the missing answer-key step in d218 and d247, is the same concern as my old PB9. d218 and d247 cannot move `terminals-per-token`, and the remaining keys are compared at the verdict reading. Residual. |
| All earlier residual rows | not promoted | Unchanged. |

## 4. New findings

| id | Class | Issue | Claim | Consequence | Evidence | Suggested action |
|---|---|---|---|---|---|---|
| N18 | residual | d218 | The repointed tests compare the judged forest with `ConstrainReference` over the no-precedence forest. | V2 is held equal to `Constrain`, and `Constrain` and `ConstrainReference` are pinned to diverge on one toy (`TestPrecedenceReferenceKnownDivergence`). The repointed toy differential `TestPrecedenceToyForestsMatchReference` passes today only because it handles that toy. That handling must survive the repoint. | `forest/precedence_pinned_test.go`; d218's repointed list | One clause in d218: "toys on which `Constrain` and `ConstrainReference` are pinned to diverge keep their pinned handling". |

## 5. What I did not re-examine

- I ran no V2 code. The only measurement this round is the overlay count of narrowed sets against the committed tables.
- I did not re-read the 16 unchanged bodies beyond confirming they are byte-identical to round 7.
- I did not check how `TestPrecedenceToyForestsMatchReference` handles the divergent toy today. N18 is a reading of names, not of the test body.
- I did not re-check the residual mapping. The milestone diff shows only a Records entry.
- I did not run the perl.git targets, the fuzz target or `make chart-ratchet`.
