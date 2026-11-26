# chain-review round 6: refinement repair on chart-perly-w4

## Scope

This round answers the four blocking findings that were still open after round 6. No residual finding is answered, including r6-fresh6 FR6-3 to FR6-9 and r6-pushback N15.

All edits went through the `git zhi` CLI (git-zhi 0.5.2). The repository, git config and branches were not touched. No issue was added or split, and no edge, title or label changed, so there are still 19 issues.

## How each finding was answered

### 1. The explicit-precedence off-switch through `ParseWith`

Raised as r6-fresh4 FR6-1, r6-fresh5 FR5-1 and r6-fresh6 FR6-1. Issue: 01a21463-3262.

**The guard.** A recognizer now records, when it is judged, the precedence it judges with. `ParseWith` panics unless that precedence is the one `New(tab)` would derive for `tab`:

- **Tables whose rules are perly.y's, rule for rule.** This uses the derived form's content test, so a copy of the committed tables counts. The recognizer must judge with the derived precedence. A no-precedence recognizer panics. So does an explicit precedence, for example an empty one built with `prectab.NewPrecedence(tab, nil, …)`.
- **Any other tables.** The explicit form stays allowed. Those tables have no derived levels to bypass.

**The criterion.** `TestParseHasNoPrecedenceSwitch` gains two required subtests:

- `parsewith-refuses-empty-explicit-precedence-over-committed-tables`
- `parsewith-refuses-empty-explicit-precedence-over-copy-of-committed-tables`

The guard step names both refusals.

**The moved test.** `TestParseRefusesInfeasibleGrammar` now runs through `parseStream` inside package `forest`, as r6-fresh4 suggested. It no longer goes through the public path.

### 2. 3262's false claim about the in-package `New(` calls

Raised as r6-fresh5 FR6-1 and r6-alignment AL17. Issue: 01a21463-3262.

The single "not expected to move" bullet is split into three:

**(a) Level-less toys.** These recognize as before. d199's `TestJudgedOverLevellessTablesRecognizesAsBefore` covers them.

**(b) Tests that judge from 3262 and are expected to hold.** These run over perly.y's committed tables, its sample or the corpus. I found them with a grep for `committedTables(` and `corpusSample(`:

- `TestCacheMatchesNoCache`
- `TestInternedSetsSurviveBufferReuse` and `TestRecognizerReuseMatchesFresh`
- `TestPostdotIndexMatchesSetScan` and `TestPostdotIndexConcurrentRecognizers`
- `TestRecognizersShareGrammarTables`
- the four slipper tests
- `startforms_test.go`, `copy_test.go`, `resolve_test.go` and the rest of `memory_test.go`

If one of these goes red because a forest or an acceptance differs, that is a §7 stop. `agree_test.go` and `setdigest_test.go` were already pinned in 3262.

**(c) Cost bounds over perly.y**, with today's numbers:

| Test | Today | Bound |
|---|---|---|
| `TestCompletionProbesBounded` | 0.6362 goTo searches per token | 0.64 |
| `TestLeoMemoEntriesBounded` | 0.0395 memo entries per node | 0.040 |

`TestCacheLookupBounded`, `TestExpectedMemoBounded` and `TestInternOneLookup` are in the same group.

**The rule for (c).** If a cost bound goes red, the builder reports the measured number to perigrin. The bound does not change without perigrin's ruling. A cost move is a cost finding, not "a forest differs".

### 3. r6-pushback N14

Issues: 01a21447-d218 and 01a21447-d199.

**d218.** The issue now states what "production" means over the leveled toys of `testdata/precedence/toys` that `prectab.Feasible` passes. Those are the 13 toys, less `empty` and `emptyred`, which stay excepted. For each such toy, the candidate is the judged recognizer in the explicit form over the toy's own precedence, then Build, with no `Constrain`. The reason is that `Parse` has no toy precedence, and its derived form judges nothing over non-perly.y tables.

Criteria 4 and 6 now describe the toys this way. Their commands are unchanged.

**d199.** Its explicit-form list adds "the production differential's toys from the Constrain-removal issue on", with the same sentence.

### 4. r6-fresh6 FR6-2

Issue: 01a21447-d278.

The brute force runs over level-less toys (`forestToys` and `slipperToys`), which judge nothing. As for `FuzzRecognize`, it gains a §3-feasible toy with levels, reusing the fuzz oracle's toy. Over that toy it does three things:

- counts the derivations `ConstrainReference` keeps over the no-precedence forest;
- compares that count with the forest from the judged recognizer (explicit form) and Build;
- states in `forest_test.go` that its level-less toys judge nothing.

The step "Regenerate the brute-force reference's expectations" is replaced, because there is no expectations file; the new step is a TDD block. A new criterion requires `TestBruteForceJudgedOverLeveledToy` to pass, `refused derivations: N` with N > 0, and the `judge(s) nothing` statement in `forest_test.go`.

## Change log

| Issue | Change | Finding |
|---|---|---|
| 01a21463-3262 | The `ParseWith` guard paragraph is rewritten; the guard step is extended; the `TestParseHasNoPrecedenceSwitch` criterion gets two subtests. | r6-fresh4 FR6-1, r6-fresh5 FR5-1, r6-fresh6 FR6-1 |
| 01a21463-3262 | The infeasible-toy prose and criterion move to `parseStream` inside `forest`. | same |
| 01a21463-3262 | "Not expected to move" is split into three bullets, (a), (b) and (c). The cost-bound bullet names the bounds and the rule for perigrin. | r6-fresh5 FR6-1, r6-alignment AL17 |
| 01a21447-d218 | Adds the explicit-form sentence for the leveled toys; reword criteria 4 and 6 (commands unchanged). | r6-pushback N14 |
| 01a21447-d199 | The explicit-form list mirrors N14. | r6-pushback N14 |
| 01a21447-d278 | Adds the brute-force paragraph, step and criterion. | r6-fresh6 FR6-2 |
| Milestone Records | Adds a "Chain-review round 6 repair" entry. No mapping row changed. | all |

Every other line of each edited body is unchanged from the live body read at the start of this round. The four stored bodies and the milestone body were re-read and match their source files. Labels and the resolution command are unchanged.

## Verification

- **New criteria are red.** I ran the 5 new or changed criterion commands at e3432c0f with `PERL5_CORPUS=/nonexistent`. All 5 exit 1. The two d218 commands are unchanged; I re-ran them anyway.
- **Every test and file named exists.** I checked the named tests in `recognize/` (cache, intern, leo, probes, recognize, shared, slipper) and the two bound constants, `completionProbesBound = 0.64` and `leoMemoEntriesBound = 0.040`. `forestToys` and `slipperToys` are in `forest_test.go`. `testdata/precedence/toys` holds 13 toy files.
- **Lint passes.** `internal/chart/offline/criteria-lint.sh` exits 0 over all 19 issues.

## Unchanged

- The critical chain is still 15 issues long.
- **Open choice for perigrin.** r6-fresh6 FR6-2 offered "keep the brute force level-less, as a ruling". I took the other branch: the brute force gains a leveled toy. If perigrin prefers the ruling, the new d278 step and criterion come out together.
