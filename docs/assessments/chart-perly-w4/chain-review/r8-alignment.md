# chain-review chart-perly-w4, round 8 (the bound): review-alignment (resumed)

## 1. Recommendation

**Accept.**

I hold no blocking finding. One row, r7-fresh7 FR7-1, I judge to be a defect. Defects are classed blocking, so I promote it. I then explicitly stand aside on it, with grounds in section 3. Its fix is a one-line rewording of an issue that does not start until every other issue has closed. That rewording should be made before execute reaches it.

**What I checked first.** I read the live chain before reading `repair-r7.md`.

- **Chain.** 19 issues at `e3432c0f`. Edges come from `blocked_by` and are unchanged since round 3.
- **Criteria.** I re-ran all 200 against `e3432c0f`, passing each to `sh -c` unmodified. All 200 are red.
- **Residual mapping.** The 66 rows and the crumb are all present. The only mismatches are the five AL6 rows, which are residual.
- **Carries lines.** Unchanged.

**Coverage against the inputs.** Every design goal and anti-goal, every §4.1–§4.5 item and every §5 item has an owning issue. So do perigrin's rulings, including the bounded supersession, narrowing before recognition, the rerun diagnosis, "Hold golden, retire key" and "Relative only". perigrin's goal requirements hold:

- the three stages run in order, wired by hard edges;
- every issue has a model label;
- criteria are keyed on `e3432c0f` or `e3432c0f..HEAD`, never on issue ids;
- every criterion is red before its work;
- the shared README sections have one owner, `bb9b`, which runs last;
- each code issue makes a README-first commit;
- the §7 stops are written into every issue as reported, never tuned toward.

No issue is scope creep.

## 2. Status of my findings

| id | status | grounds |
|---|---|---|
| AL1–AL4, AL12, AL13, AL14, AL17; my promotions PB9, PB17, FR3-4 and r4-fresh3 FR4-1 | released (earlier rounds) | The criteria re-run is 200 of 200 red. The texts that answered these findings are unchanged except where noted below. |
| AL5–AL11, AL15, AL16, AL18 | residual, held | Unchanged. AL18 still applies: 3262's "judge from here" list omits three `shared_test.go` tests. The run-first step covers them. |

**What the repair changed, checked against what it also had to change.**

**d218.**
- **The repointed differentials are named.** All six exist in the tree: three in `forest/precedence_reference_test.go`, `TestPrecedenceCorpusMatchesReference`, `TestPrecedencePerlGitMatchesReference` and `TestWaypoint2RefusalsMatchReference`.
- **Inputs the judged chart refuses get d247's verdict-only rule.** This is design §5's "production now means the recognizer's forest", applied where the forest does not exist.
- **3262's joint criterion stays green at verify.** It runs `TestWaypoint2RefusalsMatchReference`, and 3262's text now notes that d218 later repoints that test.
- **The instruments that stay on the no-precedence mode are named.** All of them exist.
- **The new criterion is red today.** It needs `TestWaypoint2RefusalsMatchReference verdict only: [1-9]`. That holds because the waypoint2 rows include `print $fh [];`-style precedence refusals. FR2-1 established this.

**d166.**
- **Stable narrowed set ids.** A narrowed set id keeps its contents for the recognizer's life. `TestNarrowedSetIdsStableAcrossFiles` checks this in both file orders. Both are red today.
- **It closes a real hole.** The transition cache keys on set id, and a long-lived recognizer serves many files, which is the ratchet's speed side.
- **Nothing else moved.** The rest of d166 is unchanged.

## 3. Promotions

| row | from | to | grounds |
|---|---|---|---|
| r7-fresh7 FR7-1 | residual | **blocking, then stood aside by me (the promoter)** | See below. |

**r7-fresh7 FR7-1: the defect.** `bb9b` item 8 tells the README rewrite to say that "an alternative kept in one context is kept in all" now applies to "`Constrain` only, since production's Build walks by context". That is false under the chain's own criteria.

- **The rule V2 must reproduce.** V1's shortcut keeps, at a node reached under two live contexts, the union of what each context keeps (README around line 2716).
- **V2 must keep the same union.** Design §2 and §5 require V2's forest to equal V1's. d0ba's `TestForestEqualityUnionShortcut` constructs exactly that node: "contexts: 2, distinct non-empty alternative sets: 2". d218's differential then requires V2 to equal V1 on it.
- **Where "Build walks by context" applies.** It only drops the alternatives of a context whose parent chain dies (C6). V1 drops those as unreachable.
- **So the statement is false.** The shortcut still describes production's forest. A README following item 8 literally would be false. It is an instruction to write a false claim into a live document, so it counts as a defect.

**Why I stand aside on it.** The concern is recorded and the chain proceeds.

1. **It can be fixed without a stop.** `bb9b` is the last issue: it is blocked by d2d1, which is blocked by every code issue. Rewording item 8 before it starts is an edit to an unstarted issue. It is not "an amendment of a started issue", so it does not trigger a §7 stop.
2. **The issue itself would catch it.** `bb9b`'s first step derives the claims from the merged diff and the verdict reading. Its text also says "a sentence this issue cannot make true … is reported, not papered over". A writer who reads `TestForestEqualityUnionShortcut` in the diff has the evidence in hand.
3. **The review gate would catch it.** `crochet:review` reads the live documents against the delivery's diff, so a false README sentence is caught there.
4. **No criterion goes the wrong way.** `TestPrecedenceShortcutStated` only requires the phrase to remain in that section. The fix needs no criterion change.

**The fix.** Reword item 8 to say that the union is a property of the shared-node forest and holds for V2's production forest as it does for `Constrain`. Walking by context is what keeps a dead context's alternatives out.

**r7-fresh7 FR7-2: not promoted.** 3282's `TestBuildDropsAlternativeOfDeadContext` names "the union-shortcut fixture", whose two contexts are both live.

- 3282 can add an input of its own with a dying context.
- d0ba's criterion logs only the union-shortcut fixture's own two contexts, so an added input does not break it.
- That makes it a wording gap, not a criterion that cannot pass, and not an amendment of d0ba.
- It stays residual. The suggested fix is "its own constructed input" in place of "the union-shortcut fixture".

**Other rows read and not promoted.** I read every row in r1–r7.

- **r7-fresh7 FR7-9** (d199's size) is pushback's lens.
- **r7-fresh7 FR7-11** (`-race` needs cgo). I measured it: `go test -race` runs here, with cgo enabled and gcc present. CLAUDE.md's "pure-Go" statement is about the build, not about the race detector.
- **The rest of r7-fresh7 FR7-3 to FR7-10, r7-fresh2 FR7-1, r7-fresh4 FR7-1, r7-fresh5 FR7-1 and FR7-2, r7-fresh6 FR7-1 and FR7-2, N16 and N17** are each a gap in wording or a measurement to take. None is a claim that is false and has no other catch.
- **Rows carried from earlier rounds:** unchanged.

## 4. New findings

None.

## 5. What I did not re-examine

- **FR7-1, measured.** I did not build V2 to check the union. My judgement rests on design §2/§5's forest-equality requirement and d0ba's union-shortcut criterion, which together force V2 to keep the union.
- **Unchanged issues.** Only d166, d218 and 3262 changed this round. I did not re-read the other 16 beyond the criteria re-run and the mapping cross-check.
- **perl.git targets,** including `TestPrecedencePerlGitMatchesReference` under d218's new step. Their criteria are gated red today.
- **Whether `TestWaypoint2RefusalsMatchReference`'s verdict-only count will be non-zero.** That rests on FR2-1's finding that `print $fh [];` is a waypoint2 row. I did not rerun it this round.
