# crochet:assess round 7: the architecture and direction lens

**Subject:** revision 7 of `docs/plans/2026-11-23-chart-stateless-precedence.md`, at 9afff6c63 (the minute is at 1ff6e3f6a).
**Blob:** ce616721c9ed717dc119cf4c5f3329daf802f4cd, checked with `git hash-object`. It matches the expected value.
**Recommendation: accept.** Nothing of mine remains held. Four new findings are low or textual, and none blocks:
- A40: "most" in §2 should be "the largest share";
- A41: name `edit5k` among the in-tree corpora, and give the in-tree aggregate check the same slow-path fallback the generators have;
- A42: name the gated outputs in the first stop;
- A43: one word in the residue of the 3.92 ruling.

**Bias.** I raised A36–A39 and only I may release them, so I lean toward finding my own bar met. Each release below quotes its text.

## What I checked

I read the subject first.

Against the evidence:
- I recomputed every §2 figure from `measurement/combined-runs/result-run{1,2,3}.txt` (the `RESULT` lines), `constrain-measurement.md`, `harness/fresh5/result-t1.txt` and `stateless-compare.txt` (tables built in 86 ms).
- I grepped the round 6 reports for the re-run of the generators.
- I read the diff in `INDEX.md`, which now explains the 11,588 lines against the 11,574 files.
- I confirmed that `reports/round6-architecture.md` matches my scratchpad copy byte for byte.

Against the CI timeout bound:
- `.github/workflows/chart-ratchet.yml`: "advisory", `timeout-minutes: 20`.
- `internal/chart/cmd/ratchet/main.go`: the order of measurement. `measureLex` runs `measureSpeed` and `chartStages` before `measureTDT` fails for lack of fixtures, so in CI the stage keys do run inside that timeout.

I read `git diff --stat 788778ed0 9afff6c63` last.

## Not examined

- I re-ran nothing.
- I did not read round 6's codebase, fresh or pushback reports beyond grepping them for the generators.
- I did not read the minute's round 6 section.
- I did not check:
  - whether CI's lex-overhead load gate stops a run before the stage keys on a busy runner. If it does, the timeout bound goes unexercised there.
  - how long the in-tree aggregate part of §3.4 takes under `make test`.

## A36–A39

| id | disposition | ground |
|---|---|---|
| A36 | **moot, by ruling** | §1: "'Drop the number'. There is no numeric target for V1 and V2 together; the V2 decision sets its own after this decision's verdict reading. This supersedes the 3.92 target." With no number, nothing needs an instrument. |
| A37 | **released** | §2: "In the combined prototype V2 measured 3.85–4.02 over recursive descent. Round 5's independent run measured 4.10, against V1's 4.12 in the same run. Its gain over V1 did not reproduce." V1 now reads "4.03–4.17" for the combined runs and "4.12" for round 5, each from its own source. |
| A38 | **released, by ruling and text** | §3.4 under "Split it": the perl.git files run only in `make chart-precedence-reference`, behind a build tag. "The perl.git target is not run by `make test` … A developer without perl.git runs `make test`." Every `make test` row says "Corpus needed: none". |
| A39 | **released** | §2: "The precedence stage is 0.39–0.42 of today's, per run." `INDEX.md` now says "`list.txt.gz` holds 11,588 lines: 3,627 of them are repository paths under fresh4's since-deleted scratch worktree. Mapped onto an 8c3317079 checkout, the list resolves to the 11,574 files." |

## Assessment of revision 7

### §1 after "Drop the number"

The ruling is recorded with what it supersedes. The surviving part of "Let's get ourselves to 3.92 first then look" is restated as an order: other stages wait for V2. That is the right residue. One word is loose (A43).

### §2: figures against the evidence

Every figure matches its source.

| §2 figure | Evidence |
|---|---|
| `constrain-runs`: 2.33–2.45, 43% | the minute's 2.38 (2.33–2.45), 43.4% |
| `combined-runs`: 2.18–2.23 | V0 precedence 2.2061, 2.1770, 2.2275 |
| `combined-runs`: 39% | those over ParseWith: 39.2%, 39.0%, 39.2% |
| V1 0.73 (0.72–0.74), 4.03–4.17 | V1/V0 per run 0.740, 0.721, 0.734; V1 4.1597, 4.0274, 4.1651 |
| precedence stage 0.39–0.42 | 0.423, 0.406, 0.392 |
| tables about 90 ms | `stateless-compare.txt`: 86 ms |
| V2 3.85–4.02; 4.10 against 4.12 | combined runs; fresh5 |
| "Round 6 re-ran both generators and the 11,574 files" | `reports/round6-fresh.md`: "Seed 4 produced 20,000 random sentences plus 83 handcrafted ones"; `round6-codebase.md` §4 |

The one wording defect is A40: "most".

### §4: before-state, timeout bound and quiet-session stop

All three are consistent with the README and with chart-recovery's practice.

- **Before-state.** "That commit is the before-state": the stage-key commit, as in chart-recovery, where M came before any R or F issue. It is read interleaved with the after-state in one session at the end, so the README's rule that timing keys are never compared with the baseline holds.
- **Timeout bound.** It is real. In `main.go` the speed and stage keys run inside `measureLex`, before `measureTDT` exits 2 for want of fixtures in CI, so CI's 20-minute timeout does cover them. It sits beside the README's local budget (300 s, already overrun; input 01a1ec3c), which still governs the development machine. The commit "reports the keys' cost to a ratchet run", and that report covers the local side.
- **Quiet-session stop.** "A quiet session cannot be had … the reading records the load, and perigrin rules on it." This matches the README's checkpoint 3 ruling ("A reading records its load rather than requiring a quiet one", perigrin, 2026-11-15) and keeps the decision with perigrin rather than with the reader.
- **New work.** The section now names what the keys need: a forest copy, and exported `libraryOperators`, resolve and readings. It also notes that the stage keys do not sum to the pipeline. Both are honest.

### §5: is the stop list complete?

It is nearly complete. The list covers:
- a mismatch against the reference;
- a state-dependent key the guard cannot settle;
- no precedence gain;
- a gain without the exit;
- no quiet session;
- the timeout;
- a needed criterion change;
- twice the budget, with a report at the budget itself.

One gated output is not named as a stop: A42.

### §6: path naming

§6 names the `Makefile` target and the CI step "for perigrin to rule in at refinement, keyed to the issue that adds the target". That is the w2 precedent improved, and it takes my round-6 suggestion. Under "Split it" only the perl.git part needs those paths, which is consistent with the rest.

## New findings

### A40. §2 says `Constrain` takes "most" of the pipeline's time (low)

- **The text:** "`Constrain` takes most of the pipeline's time". Its own table gives 39% and 43%.
- **Consequence:** small. A reader takes "most" to mean a majority and overestimates what this decision can save.
- **Modify:** "the largest share of the pipeline's time".

### A41. §3.4's in-tree row omits `edit5k`, and the in-tree aggregate has no fallback if it proves slow (low)

- **The omission.** The row reads "(T1, Guacamole, std5k)". The README lists the in-tree corpora as `guacamole`, `std5k`, `edit5k` and `T1`, and the superset population takes files from all of them.
- **No fallback.** §3.4 gives the seeded generators a fallback ("If it proves slow, the bulk moves behind the build tag"). The in-tree part, about 1,650 files judged twice, has none. ABSENTFAST held the chart suite to 50 s. The prototype prepared 2,211 forests in 17 to 19 s.
- **Consequence:** the population is mis-stated, or `make test` grows with no stated recourse.
- **Modify:** write "every in-tree corpus of the superset population (T1, Guacamole, std5k, edit5k)", and extend the "if it proves slow" sentence to this row.

### A42. §5's first stop names only the reference comparison, not the other gated outputs (should fix)

- **The text.** §3.4 ends: "What stays unchanged: both forest digests and every ratchet count key." The first stop is "a §3.4 check differs from the reference".
- **The gap.** A digest that moves, or a ratchet count key that moves (for example a refusal byte, or `precedence-covered`), is not a difference "from the reference". The ratchet would exit 1 on a count key regardless.
- **Precedent.** Revision 3's first stop was "any output in §3.1 changes", and its §3.1 named the digests and count keys.
- **Consequence:** an unexpected digest change has no stop to report it, so it is left to a worker's judgement whether to regenerate. The golden's own comment forbids regenerating it.
- **Modify:** "a §3.4 check differs from the reference, or a forest digest or ratchet count key moves".

### A43. §1: "until the V2 decision is made" is ambiguous (low)

- **The text.** §1 says allocation and the other stages "wait until the V2 decision is made". §5 says they "wait".
- **The ambiguity.** "Made" could mean decided (accepted at assessment) or delivered (built and its verdict read). perigrin's words were "get ourselves to … first then look", which reads as delivered.
- **Consequence:** small. Whether allocation work may start while V2 is in execution is left to interpretation.
- **Modify:** "until the V2 decision's verdict reading", or perigrin says which he meant.

## Summary

| id | status |
|---|---|
| A36 | moot (ruling "Drop the number") |
| A37 | released |
| A38 | released (ruling "Split it" and §3.4) |
| A39 | released |
| A40 | new, low: "most" should be "the largest share" |
| A41 | new, low: `edit5k`; slow-path fallback for the in-tree row |
| A42 | new, should fix: the first stop should include a moved digest or count key |
| A43 | new, low: "made" should be "verdict reading" |

**Held: none.** Every A-row of mine from rounds 1 to 6 is released or moot.

**Recommendation: accept.**
- A42 is one clause and worth making before refinement.
- A40, A41 and A43 are wording.
- None of the four changes what the decision builds, measures or stops on.
