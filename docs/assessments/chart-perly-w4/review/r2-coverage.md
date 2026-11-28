# crochet:review round 2: coverage (prefix C)

The subject is unchanged: `git diff e3432c0f 8d8fff38`. I read it in a scratch worktree at 8d8fff38, `wt-coverage`, which is now removed. Nothing in the repository was edited.

**Result:**

- I hold no blocking findings.
- I promote nothing.
- I raise one new RESIDUAL finding, C3.
- C1 and C2 stay RESIDUAL.

## Round 1's 13 residual rows, read through the coverage lens

The question for each row: does it leave a design requirement undelivered, or make a §7 stop fire?

| Row | Read | Promote? | Grounds |
|---|---|---|---|
| C1, chain export at close | Re-checked: `git log e3432c0f..8d8fff38 -- docs/assessments/chart-perly-w4/chain` still shows two exports (`c7157c2e`, `73f94605`) | no | Due at close. Nothing breaks if the delivery is accepted with it open. |
| C2, #649 in the range | Unchanged | no | Bookkeeping only. |
| Q1, `NewJudged` does not check `p.Tab == tab` | Agree | no | The shipped path derives `p` over `tab`, and `ParseWith` refuses an explicit precedence over perly.y. Every V2 requirement is met without the check. |
| Q2, the slipper closure's reduce side is unjudged | **Confirmed, with my own re-run** (below). It also bears on a live document: see C3 | no | The design's reduce side (§4.1) is not applied to completions inside `semiAt`'s closure. Every verdict there is allow on perly.y today, so no forest, key or verdict differs, and no §7 stop fires. The risk is to a future grammar change, which a pin closes. That is residual. |
| Q3, records keyed by parent state; `alt` checks the first item | Agree, unconfirmed | no | The design's §4.4 property, "Build never pairs a completion with a parent the recognizer refused", is held by the forest-equality differentials, which are green: corpus 1881/1881, perl.git 620/620 (my round-1 re-run), waypoint2, referee, and seeds 4 and 44. Quality could not construct a trigger. It should be written down as an invariant, but it does not block. |
| Q4, dead `constrainSource` | Agree | no | Cleanup. |
| P1, test-only API in `forest/equality.go` | Agree | no | It bears on my round-1 note that "test-only" in §2 and §4.5 holds by use, not by file. The design does not require the file split. |
| P2, referee V1 tail in `precedence.go` | Agree | no | As for P1. |
| P3, two stale comments | Agree | no | Code comments, not live documents. |
| P4, three "is this perly.y?" predicates | Agree | no | It could make the differential's two sides judge different grammars, but only on tables that tests construct. The shipped differentials run on the committed tables, where all three agree. |
| P5, duplicated pairing step in `complete` | Agree | no | The branch-set path is covered by `TestJudgedBranchSetsAgree`. Maintenance only. |
| P6, test-only exports and counters | Agree | no | `ConstrainEntries` backs the evidence that "production enters Constrain 0 times". |
| P7, `LeoLink` / `LeoLinkContext` | Agree | no | Nice to have. |

## Evidence re-run for Q2 / C3

In a scratch worktree at 8d8fff38, I added a throwaway test in `internal/chart/prectab` (since deleted) that counts the rules whose `ReduceAt[r*NT+t]` is not 0, over perly.y's committed tables:

```
191 PERLY_BRACE_CLOSE: non-allow 0 of 350 (NT 238)
206 PERLY_SEMICOLON:   non-allow 0 of 350 (NT 238)
EOF 237 (symbols 237): non-allow 0 of 350
```

At 8d8fff38, `recognize/slipper.go` `semiAt` builds `v := &builder{r: r, c: c, pos: pos, interned: true}` and calls `v.complete(la)`. It never sets `v.la`, so `complete` runs with `j == nil`, and none of the closure's completions is judged. Those completions precede `}`'s terminals or EOF.

By contrast, `recognize.go:318–319` sets `b.la` only on the main builder, with the comment "the reduce side is judged under every option".

So Q2 is accurate:

- The only pin is `PERLY_SEMICOLON` (`TestSlipperSemicolonReduceNeverForbidden`). Nothing pins `}` or EOF.
- All three are 0 today.

## New finding

### C3: RESIDUAL. The README states the reduce side without exception, and its reason for leaving the slipper unjudged covers only `;`

- **Where:** `internal/chart/README.md`:
  - l.1985–1988, "The reduce side. A completion is refused where V1's reduce verdict forbids its rule before every terminal of the lookahead set…";
  - l.1979–1981, "a judged recognizer never recognizes unjudged";
  - l.2818–2822, "Nothing judges the slipper's zero-width `;` after recognition, and nothing needs to: no reduction before `;` and no shift of it has a forbidding verdict";
  - l.3285–3288, the same, under `prectab`.

  The code is `recognize/slipper.go` `semiAt`.
- **Evidence:** the code read and the re-run above.
  - The completions in `semiAt`'s closure are the rules completed after the inserted `;`, with the following `}`, or EOF, as their lookahead. They are not judged.
  - The README's stated reason ("no reduction before `;`…") pins the wrong terminal for those completions. Reductions before `;` happen in the main builder, where `;` sits in the closing set's lookahead and is judged. The closure's reductions come before `}` or EOF, and nothing pins those.
- **Grounds:** this is an enumeration the change leaves incomplete.
  - The README's universal reduce-side claim has one unstated exception.
  - The justification it gives for not judging the slipper does not cover that exception.
  - It is RESIDUAL, not BLOCKING. The exception is harmless on perly.y today (0 of 350 for both `}` and EOF), so no forest differs, and the forest-equality differentials confirm that.
  - Q2 is the code side of the same gap. C3 is the live-document side, and the coverage lens owns that side.
- **Remedy:** take one of these, in one change, with the README first:
  - (a) judge the closure: set `v.la` and `v.laKey` in `semiAt` when `r.judge != nil`. The README's universal claim then becomes true.
  - (b) keep it unjudged, and:
    - add `}` and EOF pins beside the `;` pin;
    - add a `shortcut:` comment in `semiAt`;
    - correct the `recognize.go:319` comment;
    - state the exception, and its real reason, in the README's reduce-side bullet and at l.2818 and l.3285.

## Release

The only findings I raised are C1, C2 and C3, and all three are RESIDUAL. I hold no BLOCKING finding, so I have nothing to release.

**blocking held: 0; promoted: none; new: C3 (RESIDUAL).**
