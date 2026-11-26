# chain-review round 7: review-fresh4

I read the live chain (19 issues) from `/home/user/pvm/.claude/worktrees/v2-refine` into `fr7/`. I diffed it against my round-6 copy (`fr7/diff.txt`) and ran my checks. Only then did I read `repair-r6.md`.

What changed since round 6:
- The bodies of d199, d218, d278 and 3262, and an appended entry in the milestone's Records.
- No edge, title or label changed, and no issue was added.
- No mapping row or `Carries:` line changed.

## 1. Recommendation: accept

My blocking finding r6-fresh4 FR6-1 is answered, and I release it. I found no new blocking finding. What remains is residual.

## 2. Status of my findings

| Finding | Status | Grounds |
|---|---|---|
| r6-fresh4 FR6-1 (an explicit precedence through `ParseWith` is an off-switch) | **released** | **What the fix had to reach.** A fix had to do four things: (a) bind `ParseWith`'s guard to the precedence the recognizer carries, not to the constructor that built it; (b) catch an explicit precedence over a *copy* of perly.y's tables as well as over the committed value; (c) take the toy-refusal test off `ParseWith`; (d) leave every existing `ParseWith` caller passing.<br>**(a) Landed.** 3262: "A `recognize.Recognizer` records which constructor built it and, when judged, the precedence it judges with. `ParseWith` panics unless that precedence is the one `New(tab)` would derive for `tab`."<br>**(b) Landed.** The guard uses the derived form's content test, and the criterion has subtests `parsewith-refuses-empty-explicit-precedence-over-committed-tables` and `…-over-copy-of-committed-tables`.<br>**(c) Landed.** `TestParseRefusesInfeasibleGrammar` now uses `parseStream` inside package `forest`.<br>**(d) Holds.** The only callers are `forest.Parse`, which passes `New(tab)`, and `TestPipelineStagesReproduceParseWith`, which passes `recognize.New(tab, …)` over the same `tab`. Both pass the guard.<br>**Remaining edges.** Explicit precedence over non-perly.y tables is still allowed through `ParseWith`. No perly.y precedence is bypassed there, and it is the same input class as my residual FR6-2. A derivation that fails on perly.y's tables gives `New(tab)` no precedence, and `Parse` passes that same recognizer, so it matches the guard by construction. |
| r6-fresh4 FR6-2 (`Parse` over non-perly.y tables accepts unjudged where V1 refused) | **held, residual** | Unchanged in substance. 3262 now says outright that "for tables whose rules are not perly.y's, the explicit form stays allowed". Production builds its tables from `grammar.Committed()` (`grammar/refuse.go:30`), so it never reaches this class. Fresh2, fresh3, fresh6 (FR6-7) and pushback (N15) hold the same row residual. |
| r5-fresh4 FR5-1 | **released** (in round 6) | No change since. |
| r4-fresh4 FR4-1 | **released** (in round 6) | No change since. |
| r4-fresh4 FR4-2 (the two cost bounds near their limits) | **released** | 3262 now names `TestCompletionProbesBounded` (0.6362 against 0.64) and `TestLeoMemoEntriesBounded` (0.0395 against 0.040), along with `TestCacheLookupBounded`, `TestExpectedMemoBounded` and `TestInternOneLookup`, as "cost bounds over perly.y, which judge from here". It also states: "If a cost bound goes red, the builder reports the measured number to perigrin and does not change the bound without perigrin's ruling. A cost move is a cost finding, not 'a forest differs'." That is the decision I asked for, made in advance. I re-ran both tests and got the same numbers as before. |
| r4-fresh4 FR4-3 (`make test` never compiles the `precedencereference` test; d218 does not name which reference differentials it repoints) | **held, residual** | Unchanged. CI's Ratchets job is the backstop. |
| r4-fresh4 FR4-4 (no run-first step on d1c2 and d1eb; unkeyed `Item` literals in `intern_test.go`) | **held, residual** | Unchanged. A compile error is loud, and the builder fixes it on sight. |

## 3. Promotions

None. I read every residual row from r1 to r6. The new rows since round 5:
- **r6-fresh2 FR6-1, r6-fresh3 FR6-1, r6-fresh6 FR6-7 and r6-pushback N15.** These are the same "non-perly.y tables judge nothing" class as my FR6-2. It is held residual by consensus, and no production path reaches it.
- **r6-fresh6 FR6-3 to FR6-6, FR6-8 and FR6-9.** Ambiguities in wording or scope whose failure mode is a test that goes red, which the builder can see. Nothing passes silently.
- **r6-fresh6 FR6-4.** "`forestdigest_test.go` is pinned to the no-precedence constructor". The file also holds `TestResolvedForestDigestGolden`, which reads `Parse` and so follows production. Read per test, the pin applies to the forest digest alone. Residual.

## 4. New findings

| id | blocking/residual | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR7-1 | residual | 01a21447-d278 | The brute force gains "a §3-feasible toy with levels, reusing the fuzz oracle's toy". | The fuzz toy lives in package `recognize`'s test files. The brute force lives in package `forest`'s `forest_test.go`, which cannot import another package's `_test.go` helpers. "Reusing" therefore means writing the same toy twice, or moving it to a non-test helper. Either works. | `recognize/fuzz_test.go`; `forest/forest_test.go` is `package forest`. | Say "the same grammar as the fuzz oracle's toy", or name where the shared definition lives. |

### Checks that support the repair's toy claim

d218 now says the production differential over the bison toys judges through the explicit form "over the leveled toys … that `prectab.Feasible` passes (13 toys, less `empty` and `emptyred`)". I measured counters 4 and 5, the two most likely to fail on a toy, over all 13 toys in `toyNames`. I used a scratch copy of the tree in `fr7/wt/`; the repository is untouched.

- **Counter 4**, forbidden shifts not of `at`'s own next terminal: emptyred 2, every other toy 0.
- **Counter 5**, all-nullable rules with a non-allow reduce verdict: empty 1, every other toy 0. That includes `nullpred` and `nullsuf`.

The claim holds for these two counters. I did not measure counters 1 to 3 or 6 on the toys.

## Checked and found sound

- **New criteria are red.** None of the new test names or subtest names appear in the tree. The text "judges nothing" appears in neither `forest_test.go` nor `fuzz_test.go` today, so d278's grep clause is red too.
- **Unchanged structure.** Edges, the residual mapping and the `Carries:` lines are unchanged.
- **3262's three-way split.** The "judge from here, expected to hold" list names the perly.y-backed tests in package `recognize` that I found missing in round 4: cache, intern, postdot, slipper, reuse and shared.

## 5. Not re-examined

- Issues other than d199, d218, d278 and 3262. Their bodies are byte-identical to round 6's.
- Feasibility counters 1 to 3 and 6 on the bison toys.
- I ran no V2 code. The only tests I ran were my two bound tests and the scratch toy-counter test.
