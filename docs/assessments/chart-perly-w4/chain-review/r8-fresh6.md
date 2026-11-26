# chain-review round 8 (the bound): review-fresh6, resumed

I read the subject live from `/home/user/pvm/.claude/worktrees/v2-refine` at `e3432c0f`. It has 19 issues. The dependency edges and labels are unchanged from round 7.

I diffed every body against my round-7 dump before reading `repair-r7.md`. Only d166, d218 and 3262 changed, plus one new entry in the milestone Records. No mapping row moved. The diff matches the log.

## 1. Recommendation: modify

Two items are blocking, and each needs a one-sentence fix.

- **FR6-6 is answered in prose, but the new criterion does not test it.** Its fixture puts a narrowed set id the committed tables lack in only one of its two files. So the per-file id scheme the finding warns about passes the test. I measured this (see below).
- **I promote r7-fresh7 FR7-1.** The docs issue tells its writer to put a false sentence into the live README.

## 2. Status of my findings

### FR6-6: held, blocking

FR6-6 was promoted by r7-pushback, which holds it as promoter. I am its raiser, and this is my status.

**What the fix had to reach:**
- a lifetime statement for narrowed ids;
- a criterion whose fixture makes a cross-file id collision possible: two files, each with a narrowed set the committed tables lack, with different contents;
- red before the work.

**What landed.** The statement landed in d166's Context. The criterion `TestNarrowedSetIdsStableAcrossFiles` uses "two files that narrow to different terminal sets (one with a library operator in scope, one with it out of scope)", in both orders.

**What is still wrong.** I measured the committed tables with a throwaway test. I added it to `internal/chart/forest`, ran it, then deleted it, and the tree is clean.

- 62 committed sets carry `PLUGIN_*` readings.
- With every `PLUGIN_*` reading dropped (the out-of-scope case), all 62 narrow to a set that is already committed. So 0 are absent.
- With one reading kept, 682 sets are absent, which is d166's own number.

So the out-of-scope file uses committed ids only. Only the in-scope file ever has an id beyond 1,880. Consider a recognizer that numbers new sets per file from 1,880, which is the defect. It never serves file A's new-id transition to file B, because B has no new id. So both `a-then-b` and `b-then-a` pass on it.

The criterion is green on the implementation it exists to reject. It passes before its work, the stability, exists, which is a defect in what d166 asserts.

**Fix.** Both files keep a library operator in scope, with different kept operators, or the same operator over different base sets. Then each file narrows to a set the committed tables lack, and the two sets differ. The test logs `new sets: N` per file, with N ≥ 1, and asserts that the two files' new sets differ in contents.

### Earlier findings, still open

| id | status | grounds |
|---|---|---|
| FR6-3 | held, residual (wider) | d218's new repoint text gives the candidate as "the judged recognizer's forest, then Build, with no `Constrain`". It attaches `LibraryOperators … where the test applied it before` to the *reference* sentence only. Read literally, `TestPrecedenceCorpusMatchesReference`'s candidate is not narrowed and keeps dead `PLUGIN_*` readings that the reference drops. Design §4.1 claims equality for narrowed input only. A careful builder will narrow, as production does. It is the same kind of slip as F8-P4, so it stays residual. The fix is one clause: "over the lattice narrowed under the file's scope", in d199, 3282 and d218. |
| FR6-4 | held, residual | Unchanged. `forestdigest_test.go` is listed as pinned, but its recognizer is `built()` in `resolve_test.go`, which 3262 lists as judging. |
| FR6-5 | held, residual | Unchanged. This is the same finding as r7-pushback N17. |
| FR6-7 | held, residual | Unchanged. This is the same finding as N15 and its siblings. I do not promote it, for the grounds in round 7. |
| FR6-8 | held, residual | Unchanged. This is the same finding as r7-fresh7 FR7-9 on d199. |
| FR6-9 | held, residual | Unchanged. This is the same finding as r7-fresh7 FR7-10. |
| FR7-1 | held, residual | Unchanged. The fuzz toy lives in a `recognize` test file, so `forest` cannot import it. This is the same finding as r7-fresh4 FR7-1. |
| FR7-2 | held, residual | Unchanged. The brute force judges with `ConstrainReference`, which has a pinned divergence from `Constrain`. |

## 3. Promotions

### r7-fresh7 FR7-1: promoted to blocking. I hold it.

bb9b, item 8, tells the writer to say that "an alternative kept in one context is kept in all" now applies to "`Constrain` only, since production's Build walks by context".

The chain itself requires V2's production forest to keep that union:

- d0ba's `TestForestEqualityUnionShortcut` fixture has one node holding two live contexts with two different non-empty alternative sets. V1 keeps their union.
- d218's real differential requires production's forest to equal V1's on that fixture.
- 3282's walk by context drops only an alternative whose context's parent chain *dies* (C6). It does not drop one of two live contexts' alternatives.

So the sentence item 8 prescribes is false of the delivered code. bb9b also says a sentence it cannot make true is reported. But item 8 states the conclusion, so a writer following it literally writes the falsehood into the live layer.

The discernment rules class an instruction to assert something false as a defect, whatever its size. The fix is to reword item 8. The union is a property of the shared-node forest and holds for production too. Walking by context is what keeps a dead context's alternatives out. `TestPrecedenceShortcutStated`'s phrase stays.

### Read and not promoted

I read every residual row in r1-* through r7-*.

- **r7-fresh7 FR7-2 (3282 needs a dead-context fixture that d0ba's fixture lacks).** Residual. 3282 can add its own constructed input in its own test without touching d0ba's criterion or fixture file. Saying so is a wording fix, not a stop.
- **r7-fresh7 FR7-4 (the slipper's second scan and add path).** Residual. `;` carries no level, which d05d pins, so the slipper's scan is never forbidden. If the slipper drops a context, the referee-sentence and corpus differentials catch it as a forest difference.
- **r7-fresh7 FR7-11 (`-race` needs cgo).** Residual. `CGO_ENABLED=1` and `gcc` are present here, and the criterion runs locally, not in CI.
- **Everything else** holds the grounds already recorded, including N15, the AL5–AL18 rows, PB12–PB20, and the r7-fresh2, r7-fresh4 and r7-fresh5 rows.

### Repair side effects checked

- d218's five-test criterion is red today, because no `verdict only:` lines exist yet.
- `TestPrecedencePerlGitMatchesReference` keeps its name, which CI's step greps for.
- 3262's joint criterion runs `TestWaypoint2RefusalsMatchReference` by PASS only, so it stays green after d218 repoints it.

## 4. New findings

None beyond the held FR6-6 defect described in section 2.

## 5. What I did not re-examine

- I ran no criterion. The one thing I executed was the throwaway set-membership probe above, against the committed tables.
- I did not re-read any body except d166, d218 and 3262. The rest are byte-identical to round 7.
- I did not check the numbers d218 quotes: 13 toys, and `verdict only` counts above 0 on waypoint2.
