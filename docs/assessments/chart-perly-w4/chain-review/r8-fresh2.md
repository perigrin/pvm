# chain-review round 8 (the bound), review-fresh2: chart-perly-w4

I read the chain live from `/home/user/pvm/.claude/worktrees/v2-refine`: 19 issues, edges unchanged. I diffed every body against my round-7 snapshot (`f2r8-diff.txt`). Only d166, d218, 3262 and the milestone Records changed. I ran my checks before I read `repair-r7.md`.

## 1. Recommendation

**Modify, on one sentence.** My own FR4-1 is answered, and I release it. I hold nothing else blocking. I promote r7-fresh7 FR7-1, because the docs issue tells its builder to write a false statement into the README. The fix is a one-sentence rewording of bb9b's item 8. If the session ends at this bound, this is the one item I would hand to perigrin.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| FR2-1 to FR2-8, FR3-1, FR5-1 | **released** (earlier rounds) | Their fixes are untouched this round. 3262's joint criterion still runs `TestWaypoint2RefusalsMatchReference`. Its prose now says the test passes on the no-precedence mode at 3262, and on the judged forest with verdict-only refusals after d218. Because the criterion only checks for PASS, it stays satisfiable at verify. |
| FR3-2 | **held, residual** | 3262's scope is unchanged. Its session size is still at the edge. |
| FR4-1 (d218) | **released** | **What the fix had to change:** name the tests, say what happens on refused inputs, give them a criterion, and keep 3262's waypoint2 criterion green at verify.<br>**Named.** d218 names the six differentials it repoints: the three in `forest/precedence_reference_test.go`, `TestPrecedenceCorpusMatchesReference`, `TestPrecedencePerlGitMatchesReference` and `TestWaypoint2RefusalsMatchReference`. I grepped all six, and each exists. `TestPrecedenceRefereeMatchesReference` is d247's. Grepping `func Test.*MatchesReference` finds no other sibling the list misses.<br>**Refused inputs.** Each test applies d247's verdict-only rule. The reference must refuse too, and the count is logged as `<test> verdict only: N`. No test errors with "recognition leaves no forest".<br>**Instruments named.** The tests that stay on the no-precedence mode are named as instruments, with the reason: they test the harness or `Constrain` itself.<br>**Criterion.** A new criterion runs the five differentials that need no perl.git and requires each one's verdict-only line, with N > 0 for waypoint2. It is red today, because no such line exists. The perl.git one has a step under `make chart-precedence-reference`.<br>**3262 agrees.** 3262's pinned-list note and its criterion prose now name the hand-off to d218. A literal reading of d218 no longer breaks 3262's criterion, and design §5's sense of "production" is met. |
| FR6-1 | **held, residual** | Unchanged. Near-perly tables still judge nothing, and four other reviewers hold the same point as residual. |
| FR7-1 (mine, d218: 11 toys pass `Feasible`) | **held, residual** | Unchanged and unmeasured. d218's first step would report a failure, so the worst case is a stop. |

## 3. Promotions

| row | from | to | grounds |
|---|---|---|---|
| r7-fresh7 FR7-1 (bb9b item 8) | residual | **blocking** | **What item 8 says.** It tells the builder to rewrite the Known-limits paragraph so that "an alternative kept in one context is kept in all" "now applies (`Constrain` only, since production's Build walks by context)".<br>**Why that is false, from the chain's own text.**<br>• V2's forest still has one node per (symbol, span).<br>• Walking by context (3282, C6) keeps out only alternatives completed under a context whose parent chain dies, and V1 does not keep those either, because they are unreachable.<br>• An alternative attached under one live context sits on the shared node, so every live parent of that node sees it. That is the union, and production keeps it.<br>• d0ba's `TestForestEqualityUnionShortcut`, which d218 makes real for production, requires V2's forest to equal V1's on a node holding two different non-empty alternative sets under two live contexts. That equality is only possible if V2 keeps the union.<br>**So item 8 asks the builder to write a statement that d0ba's and d218's criteria show to be false.**<br>**Why it is blocking.** A false claim in an issue is a defect in what the subject asserts, so I class it blocking, however small the fix. bb9b's catch-all ("a sentence this issue cannot make true without a code change is reported") does not save it. The sentence needs no code change; it is simply wrong. The criterion `TestPrecedenceShortcutStated` passes with either wording, so nothing downstream would catch it.<br>**Fix.** Reword item 8. The union is a property of the shared-node forest, and it holds for V2's production forest as for `Constrain`. Walking by context is what keeps a dead context's alternatives out. |

I read every other residual row in r1-* to r7-*.

- **r7-fresh7 FR7-2** (3282's dead-context test "on the union-shortcut fixture", whose two contexts are both live): I agree it is mis-aimed, but I do not promote it. The criterion only checks a test the 3282 builder writes. The builder can add a constructed input of its own in 3282's own test file without touching d0ba's fixture or its logged counts. So following it does not force an amendment of a closed issue. I would reword it to "on a constructed input with a dead context", but I stand by residual.
- **The rest,** including r7's FR6-3 to FR6-9 restated, N15 and AL rows: true and residual on their own terms.

## 4. New findings

| id | blocking/residual | issue | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR8-1 | residual | 01a21447-d218 | `TestPrecedenceToyForestsMatchReference` is repointed to "the explicit form over each toy's own precedence, then Build". 3262 meanwhile pins `toyForest` to the no-precedence constructor, and `toyForest` "carries the brute force and every toy `Constrain` test". | If the repointed test builds its candidate through `toyForest`, as it may today, it silently stays on the no-precedence mode. Its verdict-only line then reads 0, which d218's criterion accepts for the toys, since it requires N > 0 only for waypoint2. So the repoint could be a no-op that passes. | d218's new criterion (`[0-9]+` for every test except waypoint2). 3262's pinned list. | Require `TestPrecedenceToyForestsMatchReference verdict only: [1-9]`. The leveled toys include bison-refused sentences, for example `nonassoc`'s chain, so a judged candidate must have some. Or have the test log which constructor built its candidate. |

### Round-7 repair re-derived

- **d166, FR6-6.** A narrowed set id keeps its contents for a recognizer's life. `TestNarrowedSetIdsStableAcrossFiles` checks both orders against a fresh recognizer. It is a new test, so its criterion is red today. Nothing else needed to change for it.
- **Records.** The Records entry matches the edits, and the mapping is unchanged.

## 5. What I did not re-examine

- I ran no criterion, and I did not count the bison-refused sentences per toy that FR8-1's suggested floor relies on.
- I did not re-read the issues left unchanged this round: d02f to d135, d199, d1c2, 3282, d1eb, d247, d278, d2a2, d2d1 and bb9b, beyond bb9b's item 8 and the existing statements of C6 and union-shortcut in 3282 and d0ba.
