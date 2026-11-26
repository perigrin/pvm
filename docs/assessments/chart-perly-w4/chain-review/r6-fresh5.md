# chart-perly-w4 chain-review, round 6: review-fresh5 (resumed)

I read the live chain first and diffed it against my round-5 dump (`fr5x/` against `fr6x/`). I then ran my checks and read `repair-r5.md` last.

The repair changed three issues: d199, 3262 and d218. It also added one Records entry. No dependency edge, label or title moved, and no residual-mapping row changed. The criterion count went from 194 to 197. All the cheap criteria (the ones that run neither `go test` nor `make`) are still red on the tree today.

## 1. Recommendation

**Modify.** Two blocking findings are open:
- FR5-1, which I hold in a narrowed form;
- FR6-1, which is new and was introduced by the repair.

## 2. Status of my round-5 findings

**I am biased toward finding my own bar met.** The repair answers the bar I wrote. I scored it against what I wrote, not against what landed near it.

### FR5-1: held, blocking, narrowed

**What I predicted the fix had to change.** The fix would have to touch three places:
- **d135/d199**: where the precedence comes from;
- **3262**: its "not expected to move" list, and the `ParseWith` guard;
- **d218**: the generation-error fate;
- **d278**: the level-less fuzz grammars.

The first three of these landed. d278 already agreed with them.

**The precedence-source half is released.** d199 now states the explicit and derived forms. The derived form recognizes perly.y's grammar by its content, and treats tables whose rules are not perly.y's as declaring no levels, so nothing is forbidden. A failed derivation on perly.y's tables refuses every input. Four criteria pin this.

On the flagged question, I checked whether comparing right-hand sides could switch precedence off on perly.y itself. It cannot. A scratch probe in `internal/chart/grammar`, which I removed afterwards, found:
- G1 has 350 rules and `tab.Rules` has 350;
- there are 0 sequence rules;
- 0 rules differ by right-hand-side names (a naive by-name comparison).

So the stricter match holds today. If a sequence rule or a renamed symbol later broke the match, `TestJudgedJudgesCopyOfCommittedTables/regenerated` and `TestDefaultRecognizerJudgesPrecedence` would go red. A future grammar change therefore cannot switch precedence off silently in CI. Production tables always come from `grammar.Generate(grammar.Committed())` (`cmd/ratchet/main.go:168`, `oracle.go:185`).

**The `ParseWith` half is still held.** In round 5 I wrote: "the `ParseWith` guard checks only the constructor's identity, so a judged recognizer whose precedence came from elsewhere would pass it… say whether `ParseWith` checks that the recognizer's precedence is `tab`'s own." The repair has made this hole concrete:
- **The explicit form is a public way in.** The explicit form must be exported, because forest and chart tests build toy judged recognizers with it.
- **3262 requires `ParseWith` to accept it.** Its criterion `TestParseRefusesInfeasibleGrammar` passes `ParseWith` a judged recognizer over each toy's own precedence.
- **The guard does not catch it.** `TestParseHasNoPrecedenceSwitch` still checks only that the no-precedence constructor is refused.

So a caller can build a judged recognizer over perly.y's tables with an empty `prectab.Precedence`, which has no levels, and pass it to `ParseWith`. Precedence is then off. Since d218 removes `Constrain`, nothing downstream restores it.

That makes 3262's sentence "no recognizer passed to `ParseWith` does either" false. It also contradicts the ruling "There is no precedence off-switch".

**Fix.** `ParseWith` should panic unless both hold:
- the recognizer was built over `tab`;
- where `tab`'s rules are perly.y's, the recognizer uses the derived precedence.

Toys keep the explicit form. Add a subtest `parsewith-refuses-foreign-precedence` to `TestParseHasNoPrecedenceSwitch`.

### FR5-2: held, residual

3262's rerun text is byte-identical. It still reads "First the rerun is un-narrowed: if it dies at all…" and does not restate d166's no-rerun shortcut, which `TestRefusalLexUnmappedNotRerun` pins.

### FR5-3: held, residual

d218 is unchanged apart from the generation-error mirror. It still says "repoint the existing production-versus-reference differentials" without naming any. `TestWaypoint2RefusalsMatchReference`'s scope rows (`waypoint2_refusals_test.go:109-122`) would error on a narrowed, judged forest. This is the same finding as r4-fresh2 FR4-1.

### FR5-4: held, residual

d218's `TestForestEqualityCandidateNeverConstrains` still covers "toys" without excepting the infeasible toys. The same issue keeps those toys on `Constrain`.

### FR5-5: released

The repair put the interface in d199 and gave 3262 no new criterion, which is the bar I set. The size has moved to d199 instead: it now has 17 criteria. That is fresh2's FR2-6, and I leave it there.

## 3. Promotions

I read every residual row in r1 to r5 across all participants.

I promote **r3-fresh3 FR3-3** (also pushback N7) to blocking, and hold it as **FR6-1** below. The grounds are that the repair turned it from a risk into a false claim the subject makes.

I promote nothing else. These rows are true and remain residual:
- AL5 to AL11, AL14 to AL16;
- PB8, PB12 to PB20, N2 to N4, N8 to N13;
- FR2-5 to FR2-8, FR3-2, FR3-4 to FR3-7;
- FR4-2 to FR4-4;
- r5-fresh2 FR5-1, r5-fresh3 FR5-1.

PB8 is the closest call: no criterion checks that the brute force runs judged. I would stand aside on it alone.

## 4. New findings

### FR6-1: blocking

**Issue:** 3262.

**The claim.** The repair added this to 3262: "The roughly 113 unqualified `New(` calls inside package `recognize`'s tests run over level-less toy tables, so under the derived form they recognize exactly as before. They are covered by … `TestJudgedOverLevellessTablesRecognizesAsBefore`." The list then ends: "If one of them moves, a forest differs from V1's."

**Why it is false.** Many of those calls run over perly.y's corpus sample (`corpusSample` and `committedTables`), not toys:
- `probes_test.go`: 4;
- `intern_test.go`, `cache_test.go`, `leo_test.go`, `recognize_test.go` (`TestRecognizerReuseMatchesFresh`, `TestInternedSetsSurviveBufferReuse`) and `shared_test.go`;
- the `slipper_test.go` closure tests, `agree_test.go` and `setdigest_test.go`.

From 3262 onwards, all of these judge.

**What could move, and why that is not a forest change.** Two of these tests assert cost bounds with almost no headroom. I measured them today:

| Test | Today | Bound | Headroom |
|---|---|---|---|
| `TestCompletionProbesBounded` | 0.6362 | 0.64 | 0.6% |
| `TestLeoMemoEntriesBounded` | 0.0395 | 0.040 | 1.3% |

Contexts split sets, so judging can move these counts without any forest changing.

**Consequence.** The chain asserts something false. That false assertion routes a legitimate cost move to perigrin as a §7 stop ("a forest differs"). The level-less criterion cannot catch this, because it covers toys only.

**Evidence.**
- Per-file counts of `New(` calls and perly.y-table references in `recognize/*_test.go`.
- The measured values in the table above.
- `probes_test.go:22`, `leo_test.go:222`.

**Suggested action.**
- Split 3262's list. Say that the in-package tests over the perly.y sample follow `New` and judge.
- Name the two cost bounds. A move in either is a cost measurement, reported with its number, not a forest difference.
- Alternatively, pin these tests to the no-precedence constructor until d278.

### FR5-1 (held, narrowed)

This is the `ParseWith` foreign-precedence hole described in section 2.

## 5. Not re-examined

- I ran no `go test` criteria, only the cheap set, the three bound tests and a scratch probe. The probe was a test file added to `internal/chart/grammar` and then removed, and `git status` is clean.
- I did not re-trace d0ba to d10d, d166, d1c2, d3282, d1eb, d247, d278, d2a2, d2d1 or bb9b beyond confirming they are byte-identical to round 5.
- I did not re-check the 66-row mapping row by row. The Records diff is the single added entry.
