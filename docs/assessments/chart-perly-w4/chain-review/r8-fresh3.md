# chain-review round 8 (the bound): fresh3, as a participant (chart-perly-w4)

## 1. Recommendation

**Modify.** The fix is one clause, in two issues, and it can go to perigrin as the session's open item.

Both of round 7's repairs landed, and I found nothing they broke in edges, Records or early-green criteria. But d218's repair now spells out the repointed differentials' candidate as a recipe: the judged recognizer, then Build. That recipe omits narrowing the lattice. Design §4.1 says that, without narrowing, the recognizer keeps completions V1 drops, and that forest equality is claimed only for narrowed input. So I am promoting r6-fresh6 FR6-3 to blocking, in the concrete form the repair gave it (§3).

My own findings are all released, apart from one residual I hold (FR6-1).

**Method.**

- I read the live chain at `e3432c0f` first: 19 issues, with edges and labels unchanged.
- I diffed every body and the milestone against my round-7 snapshot. Three issues changed (d166, d218 and 3262), plus Records.
- I checked that the nine tests d218 now names all exist: `forest/precedence_reference_test.go`, `cmd/ratchet/precedence_reference_test.go`, `precedence_perlgit_test.go`, `waypoint2_refusals_test.go` and `forest/precedence_pinned_test.go`.
- I read the harness d218 repoints, `comparePrecedence` (`cmd/ratchet/precedence_reference_test.go:31-62`). It recognizes `lex.Linear(s)`, the un-narrowed lattice.
- I re-read design §3, line 107, and §4.1, line 126.
- I ran `go test -race` here, for r7-fresh7 FR7-11. It works: gcc is present and `CGO_ENABLED=1`.
- After that I read `repair-r7.md` and every row from r7-*.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| r3-fresh3 FR3-1 to FR3-7 | released | FR3-1, FR3-2, FR3-4, FR3-6 and FR3-7 were released in rounds 4 to 6, and FR3-3 in round 7. Nothing this round touches them. 3262's cost-bound list and the `memory_test` and referee fates are unchanged. |
| r4-fresh3 FR4-1, r5-fresh3 FR5-1 | released | Unchanged. d218's restated text keeps the fate of `TestPrecedenceGenerationErrorRefusesEveryFile` word for word. |
| r6-fresh3 FR6-1 | **held, residual** | Unchanged. Tables that declare levels but whose rules are not perly.y's are still parsed with no judgement, where V1 refused them loudly. No production caller builds such tables. It goes to perigrin as a recorded difference from V1, not a defect in the plan. |

## 3. Promotions

**r6-fresh6 FR6-3 → blocking.** It now applies to d218's repointed differentials as well as to 3282.

**What the chain asserts.** d218 now says the following for each repointed production-versus-reference differential:

- "the candidate is the judged recognizer's forest, then Build, with no `Constrain`: `recognize.New` over perly.y's tables";
- "the reference is `ConstrainReference` over the no-precedence forest, with `LibraryOperators` under each file's scope where the test applied it before";
- "where both accept, the kept alternatives must be equal".

A new criterion runs `TestPrecedenceCorpusMatchesReference` over the corpus, and `TestPrecedencePerlGitMatchesReference` covers perl.git. The harness lexes and recognizes `lex.Linear(s)` (`comparePrecedence`), which is not narrowed. Nothing in d218 tells the candidate to recognize the narrowed lattice that production builds from d166 onward.

**What the design says.**

- §4.1 (line 126): "A recognizer given an un-narrowed lattice applies the same rule. At the 117 tokens of §3 it keeps completions V1 drops. Forest equality is claimed for narrowed input, which is what production gives it."
- §3 (line 107) counts those 117 tokens on the corpus, the same input this criterion runs over, and 1,110 tokens carrying `PLUGIN_*` readings.

So d218, followed literally, asserts forest equality on the input where the design says equality is not claimed. `LibraryOperators` applied after Build prunes the `PLUGIN_*` alternatives. It does not undo a completion that the mixed lookahead set kept and that the parent then continued on the real terminal.

**Consequence.** Read literally, the d218 criterion either cannot pass, or passes only if none of the 117 kept completions is reachable from the root. Nothing establishes that, and the design declines to claim it. If it does go red, d218's text tells the worker that "a forest differs from V1's" is a §7 stop, which would send a false stop to perigrin. 3282's `TestForestEqualityCorpusJudgedRecognizer` (judged recognizer and Build, no `Constrain`, over the corpus) has the same gap, which was FR6-3's original target.

d199's reduce-side corpus differential is safe, because it runs `Constrain` after Build, and that removes the extra completions.

**Why this is blocking now when I kept it residual in round 7.** In round 7 I argued that a worker who follows the design would narrow. This round's repair replaced the generic "repoint at the recognizer's forest" with an explicit candidate recipe, and that recipe omits narrowing. It also built an explicit criterion on it. That makes it a defect in what the issue asserts, a criterion the design does not back, rather than a wording gap.

**Suggested action.** Add one clause to d218 and to 3282: the candidate recognizes the narrowed lattice production builds (d166), under the file's own scope. Over the toys and sentences, which carry no library operators, narrowing is the identity (d166's `TestNarrowWithoutLibraryOperatorsIsIdentity`), so nothing else changes. If perigrin prefers the un-narrowed reading, the alternative is to state the 117-token exception and have the test check it.

**Rows flagged by the orchestrator, judged and kept residual.**

- **r7-fresh7 FR7-1** (docs item 8). The README would carry a false sentence, about the union of alternatives versus walking by context. But bb9b's own rule says that a sentence it cannot make true is reported, and the writer reads the merged diff. The false sentence would need to survive a rewrite that is explicitly grounded in that diff. Residual: a reword of item 8 fixes it, and no code depends on it.
- **r7-fresh7 FR7-2** (3282's dead-context case "on the union-shortcut fixture"). d0ba's fixture is two live contexts, so 3282 needs an input of its own. Its criterion names only a new test, `TestBuildDropsAlternativeOfDeadContext`, so the builder can add that input beside the fixture without touching d0ba's criterion or its logged line. Residual, as wording.

**Other r7 rows weighed:**

- **FR7-4** (slipper's second scan site): the slipper scans only `;`, which d05d pins as carrying no precedence level, so no shift verdict there can forbid. Residual.
- **FR7-11** (`-race` needs cgo): it runs here. Residual.
- **FR7-9** (d199's size): large, but every piece has a stated criterion. Residual.
- **FR7-3, FR7-5 to FR7-8 and FR7-10**, and the rows in r7-fresh2, r7-fresh4, r7-fresh5 and r7-fresh6, AL18, N16 and N17: none leaves a state that cannot close green.

## 4. New findings

None beyond the promotion in §3.

| id | blocking/residual | issue id | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| (none) | | | | | | |

## 5. Not re-examined

- I ran no V2 code. Whether any of the 117 un-narrowed kept completions reaches a root, and so makes the promoted criterion fail rather than merely go unbacked, is not measured.
- I did not re-run `criteria-lint.sh`. The new criteria, `TestNarrowedSetIdsStableAcrossFiles` and the `verdict only:` log lines, name a test and log lines that do not exist today, so they are red.
- I did not check the node matching the repointed differentials need. `comparePrecedence` compares by node index on one shared forest, and a separately built candidate needs d0ba's structural comparison. I judge that to be implementation detail.
- I did not re-examine the 16 issues left unchanged this round beyond confirming they are byte-identical to my round-7 snapshot.
