# chain-review round 7: pushback (plan quality) on chart-perly-w4

## Drafter: growth check

I diffed the live chain against my round-6 copy (`pb6/`) before reading `repair-r6.md`.

- **Changed issues:** 4 bodies, 3262 (25 lines), d278 (9), d218 (6) and d199 (2).
- **Milestone:** one new Records entry, "Chain-review round 6 repair".
- **Unchanged:** every edge, title and label; there are still 19 issues.

| Edit | Answers |
|---|---|
| 3262: the `ParseWith` guard now panics unless the precedence is what `New(tab)` derives; the guard step is extended; two new `TestParseHasNoPrecedenceSwitch` subtests (empty explicit precedence over the committed tables, and over a copy) | r6-fresh4 FR6-1, r6-fresh5 FR5-1, r6-fresh6 FR6-1 (all blocking) |
| 3262: the infeasible-toy prose and criterion move to `parseStream` inside `forest` | same. This is a knock-on: the new guard would otherwise reject the explicit form the test used through `ParseWith`. |
| 3262: "Not expected to move" split into level-less toys, perly.y-backed tests that judge, and named cost bounds with a rule for reporting to perigrin | r6-fresh5 FR6-1, AL17 (blocking); also absorbs my residual N7 |
| d218: the explicit-form sentence for the leveled toys, and criteria 4 and 6 reworded (commands unchanged) | r6-pushback N14 (blocking) |
| d199: the explicit-form list mirrors N14 | N14 |
| d278: the brute-force paragraph, a TDD step, and `TestBruteForceJudgedOverLeveledToy` | r6-fresh6 FR6-2 (blocking); also answers my residual PB8 |
| Milestone Records | all of the above |

**Verdict.** Every edit answers a blocking finding or is its required knock-on. Folding in PB8 and N7 added no criterion beyond the one FR6-2 needed, so nothing grew without a cause. `criteria-lint.sh` exits 0 over all 19 issues, so every criterion is still red at `e3432c0f`.

The repairer flags one choice: for the brute force, it took the levelled-toy branch rather than a ruling. That is a choice between two acceptable options. I do not count it as growth.

## 1. Recommendation

**Modify.** N14 is released, and so are PB8 and N7, which the repair answered in passing.

I promote one row, **r6-fresh6 FR6-6**, to blocking (section 3), and I give my grounds there. With round 7 of 8 in mind, I promote only that row, and its fix is one criterion in d166:
- the transition cache keys on the terminal-set id;
- d166 gives narrowed sets ids outside the committed tables, and does not say those ids are stable by content within one long-lived recognizer;
- the corpus has no kept library operator, so no differential, ratchet key or verdict reading can ever see the resulting wrong forests.

## 2. Status of my findings

| id | Status | Grounds |
|---|---|---|
| PB1–PB7, PB9–PB11, PB16 | **released** | Unaffected. |
| PB17 | **released** | It is strengthened, not weakened. The guard now also rejects an explicit, for example empty, precedence over perly.y's tables or a copy of them, each with its own subtest. Explicit precedence stays allowed only over non-perly tables, which have no derived levels to bypass. Re-checked: `ParseWith` has one caller outside `forest.go`, `stages_test.go:275`, which passes `recognize.New` over committed tables. That is the derived form, so it is accepted. The infeasible-toy test moved to `parseStream` inside `forest`, so it does not hit the guard. |
| **PB8** | **released** | d278 now has `TestBruteForceJudgedOverLeveledToy`. It requires the derivations `ConstrainReference` keeps over the no-precedence forest to equal the judged explicit-form forest's, with `refused derivations: N` greater than 0, and the level-less brute-force toys to be documented as judging nothing. This is the judged brute-force criterion I asked for in round 1. |
| PB12–PB15, PB18–PB20 | residual | Unchanged. |
| N1, N5, N6, N12 | **released** | Unchanged. |
| **N7** | **released** | 3262 now lists `TestCompletionProbesBounded` (0.6362 against 0.64), `TestLeoMemoEntriesBounded` (0.0395 against 0.040), `TestCacheLookupBounded`, `TestExpectedMemoBounded` and `TestInternOneLookup` as cost bounds. A red bound is "a cost finding, not 'a forest differs'", and is reported to perigrin with its number. That is exactly what N7 asked for. All the named tests exist: I grepped each one. |
| N2–N4, N8–N11, N13, N15 | residual | Unchanged. N15 is the same finding as r6-fresh6 FR6-7: V2's `Parse` accepts non-perly tables unjudged where V1 refused them. |
| FR2-8, FR3-7, FR4-1 (my promotions) | **released** | Unchanged since their release. |
| **N14** | **released** | **What the fix had to change:** d218's prose and its toy criteria; d199's explicit-form list; the hook criterion; and agreement with the infeasible-toy exception. **Checked:** d218 now says that over the leveled toys `Feasible` passes, "production" is the judged recognizer in the explicit form over the toy's own precedence, then Build, with no `Constrain`, "because `Parse` has no toy's precedence". Criteria 4 and 6 name the explicit form for toys, and their commands are unchanged and red. d199's explicit-form list adds "the production differential's toys from the Constrain-removal issue on". The infeasible toys remain excepted and judged by `Constrain`. One count in it is unmeasured: see N16. |

## 3. Promotions

| Row | To | Grounds |
|---|---|---|
| **r6-fresh6 FR6-6** | **blocking** | **Proceeding as written admits a cheap implementation that silently produces wrong production forests, and nothing in the chain can detect it.**<br>**The mechanism.** The transition cache keys on the token's terminal-set id: `cacheKey{set: …, span: …, tok: e.Token.TerminalSet, la: la}` at `recognize.go:260`. A recognizer kept across files reuses that cache, and `ParseWith`'s purpose is to serve one ("a recognizer over tab that the caller may keep across files, as the ratchet's long-lived speed side does").<br>**The gap.** d166 gives the 682 kept-operator sets and 19 referee singletons "ids without mutating the shared committed tables". It does not say an id names the same contents for the life of a recognizer. The simplest scheme, numbering narrowed sets per lattice or per file from 1,880 up, gives one id two meanings across files. Then a cache hit in file B replays file A's transition for different terminals: a wrong set, so a wrong forest or a wrong refusal.<br>**Why nothing catches it.** The corpus holds no kept library operator (C12, d166's own text). So the corpus differentials, the ratchet keys and the verdict reading can never exercise two different narrowed sets through one recognizer. d166's criteria are single-file (`TestNarrowKeptLibraryOperator`) or about mutation (`TestNarrowLeavesCommittedTerminalSetsUnmutated`). The defect would ship and surface only on user code that uses library operators.<br>**Why blocking and not residual.** A residual is something the subject does not yet say whose absence leaves proceeding right. Here, proceeding leaves the cheapest correct-looking implementation silently wrong in production, against the milestone's own "forest equality" claim. I would not stand aside on it alone.<br>**Fix (one criterion in d166):** two constructed files whose kept library operators narrow to different sets, parsed in turn through one long-lived recognizer, in both orders, each give a forest equal to a fresh recognizer's. The prose says a narrowed set's id is fixed by its contents for the recognizer's lifetime, or is never used as a cache key. |
| r6-fresh6 FR6-5 (d2a2 misses `main_test.go`'s `resetKeys`) | not promoted | `TestResetKeySet` requires `keys == resetKeys`. Removing the key from `keys` turns it red, and d2a2's close step (`make test`) surfaces that. Removing the same key from `resetKeys` is the obvious fix, consistent with the retirement, not tuning. Residual: add `main_test.go` to d2a2's paths. |
| r6-fresh2/3/4 FR6-1 and FR6-2, r6-fresh6 FR6-7 ("tables never declare levels"; the loud failure becomes silent) | not promoted | On perly.y it stays loud: a failed derivation refuses every input, and the guard now rejects any substitute precedence. The silent branch is non-perly tables only, which no production path builds. That is the same as my N15, and it stays residual. |
| r6-fresh6 FR6-3, FR6-4, FR6-8, FR6-9 | not promoted | These are wording and size. Each is caught by a first step or a close, and none makes proceeding wrong. |
| All earlier residual rows | not promoted | Unchanged. |

## 4. New findings

| id | Class | Issue | Claim | Consequence | Evidence | Suggested action |
|---|---|---|---|---|---|---|
| N16 | residual | d218 | "the leveled toys of `testdata/precedence/toys` that `prectab.Feasible` passes (13 toys, less `empty` and `emptyred`, which are excepted as below)" | That is a count of 11 feasible toys. It has not been measured, because `Feasible` does not exist yet, and `feasibility.md` does not cover the toys directory. Several toys are shaped like the failing pinned ones (`nullpred` and `nullsuf` resemble `nullableShift`). If one also fails `Feasible`, the general rule ("the toys `Feasible` passes") still governs, so no criterion breaks. But the parenthetical becomes false, and a worker may take the mismatch for a stop. | `ls testdata/precedence/toys` (13 files); `toyNames`; d135's `TestFeasibleRefusesFailingToys` covers only the five named toys | Either drop the count, or have d135's `Feasible` test log `Feasible` over every `toyNames` toy, so d218's set is measured before it is used. |
| N17 | residual | d2a2 | `paths:` omits `cmd/ratchet/main_test.go` | This is the same finding as FR6-5. Recorded here because I flagged `main_test.go:22`'s use of the key in round 1 and did not carry it forward. | `main_test.go:22` | Add `main_test.go` to d2a2's paths and to its retirement step. |

## 5. What I did not re-examine

- I ran no V2 code. FR6-6's consequence is reasoned from `recognize.go`'s cache key and d166's text, and N16 from the toy list.
- I did not measure `Feasible` over the 13 toys. The scratch counters at `488e1c7e` would need wiring for that.
- I did not re-read the 15 unchanged bodies beyond confirming they are byte-identical to round 6.
- I did not re-check the residual mapping. The milestone diff shows only a Records entry.
- I did not run the perl.git targets, the fuzz target or `make chart-ratchet`.
