# crochet:assess round 8: architecture and direction lens

**Subject:** revision 8 of `docs/plans/2026-11-23-chart-stateless-precedence.md`, at commit 336067853. The minute
is at 146411199.
**Blob:** ed91e46239b29283ad5c82cd390411621b52a2cd, checked with `git hash-object`. It matches the expected blob.
**Recommendation: accept.** I hold nothing. From my lens this revision is a fixed point. A44 and A45 are new and
low. Refinement can carry them, and neither is a block.

**On bias.** I raised A40 to A43, and only I may release them, so I am inclined to find my own bar met. Each
release below quotes the text it rests on.

## Order of reading and checks

1. I read the subject in full.
2. I traced its new figure ("105–137 ms") to `reports/round7-fresh.md`, L4: "I … timed `NewStateless` five times
   at load 0.16: 105–137 ms. `NewPrecedence` itself took a further 297–343 ms".
3. I read round 7's fresh L5, which §3.4's new wording answers.
4. I re-read what revision 8 left unchanged: §3.1–§3.3, §3.5, the exit, the remaining stops and §6.
5. I checked the new text against the README rules I quoted in earlier rounds: the speed protocol, the 10% rule,
   checkpoint 3, and the isolation paragraph at README lines 370–382.
6. Last, I ran `git diff --stat 9afff6c63 336067853`.

## Not examined

- I re-ran nothing.
- I did not read round 7's codebase or pushback reports, or the minute's round 7 section, beyond the INDEX diff.
- I did not check how long one ratchet run takes after the stage keys are added. Six runs make up the verdict
  reading.

## A40 to A43

| id | disposition | ground |
|---|---|---|
| A40 | **released** | §2: "`Constrain` takes the pipeline's largest share." |
| A41 | **released** | §3.4's row now reads "Every recognized file of the ratchet's in-tree corpora (T1, Guacamole parses and parsent, std5k, edit5k)". The fallback now covers every row: "If a `make test` check proves slow, its bulk moves behind the build tag." |
| A42 | **released** | §5's first stop: "a §3.4 check differs from the reference, or a forest digest or ratchet count key moves". |
| A43 | **released, answered by ruling** | §1: "'V2's verdict reading (Recommended)'. Allocation and the other stages wait until V2 is built and measured." §5 says the same: "wait for V2's verdict reading". |

## Revision 8's new text

**§1, the rulings list.** It is correct and complete for the rulings now in force.

- "V2's verdict reading" replaces the "3.92 first" residue. Dropping the superseded ruling from the in-force list
  is right, because the minute keeps it.
- Two rulings share the words "Rule them in". The path ruling carries "(Recommended)" and its subject, and the
  key ruling names §4's stage keys. They are distinguishable in context.

**The v2 §11 departure.** It reads: "v2 §11's 'only `internal/chart` changes' is departed from only by the
`Makefile` and CI paths perigrin ruled in."

- This is accurate. v2 §11 governs code ("Only `internal/chart` changes. Every other package is read through what
  it already exports"). The documents under `docs/` follow chart-perly-w2's practice.
- §6's limit list matches the departure.
- The ruling is scoped to "the issue that adds that target", which is the keying I suggested in round 6.

**§4, the verdict reading.** It is consistent with the README's timing rules and with chart-recovery.

- "One measurement is one ratchet run's printed value of a key … the protocol's own median of its three internal
  measurements". This matches the README: "A figure is the median of three measurements".
- Three quiet runs of the before-state and three of the after-state, interleaved in one session, is chart-recovery's
  verdict reading: "One quiet reading on pu cd855330 (three ratchet runs, 08:10–08:16 UTC)".
- Each run's commit, load, cores and kernel are recorded. chart-recovery also asked for the kernel ("Record the host
  kernel with each quiet reading").
- Recording the exit key `speed-ratio-aggregate-longlived` with every stage key closes the gap between §1's exit and
  what is actually read.
- No figure is compared with `baseline.txt`, as the README requires ("Timing keys … are never compared with the
  baseline").
- Each ratchet run also applies lex-overhead's own load gate, which agrees with the definition of "quiet".
- The new preparation hook ("Today `Protocol.pass` starts the clock before anything per pass runs") is a correct
  observation. It belongs in the decision because it touches the speed protocol, which the README calls a contract.

**§5, the added stop.** "Or a forest digest or ratchet count key moves" is A42's fix. The timeout stop now names
`chart-ratchet.yml`, the job my round 7 check showed reaches the stage keys before failing on missing fixtures. The
list is complete for what §3 and §4 promise.

**What revision 8 broke.** Nothing I could find.

- §3.4's "the forest its caller passes to the constraint today" fixes round 7's L5, for `ParseWith` and the referee.
- §2's "the outputs themselves were not kept" is honest about round 6's re-run.
- The remaining unchanged text is as I released it in rounds 5 to 7.

## New findings (both low)

### A44. §4: the verdict reading does not record each run's aggregate exclusions (low)

- **The problem.** `speed-ratio-aggregate-longlived` is taken over an aggregate whose membership depends on load.
  The README says: "The share moves with load, and so does which files the rule excludes: a gate run at load 16 to
  20 read 17.7% and also excluded `T1/require_module_return.t` … Re-read it rather than trust it."
- **Why it is low.** All six runs are quiet and interleaved, so the populations should match. Nothing in the record
  would show it if they did not. The ratchet already prints each excluded file on stderr.
- **Consequence.** If one run's aggregate differed, a before-against-after comparison could reflect a different file
  set rather than the change.
- **Modify, optional.** Add "and each run's excluded files" to "What it records".

### A45. "What this does not amend": the README's isolation paragraph is not among the live documents to change (low)

- **The problem.** The decision lists two live documents to change first: "the chart README's precedence-constraint
  section, and the ratchet key table". The README's isolation paragraph (lines 370–382) is the live record of every
  ruling that admitted a path outside `internal/chart`, for example MAPDISPATCH's `Makefile` and `ci.yml`.
  Revision 8 adds such a ruling and does not name that paragraph.
- **Consequence.** The live record of path rulings silently falls behind. At worst a reader concludes the Makefile
  target was never ruled in.
- **Modify, optional.** Add "and its isolation paragraph". Refinement can also do this; it does not change the
  decision.

## Summary

| id | status |
|---|---|
| A40 | released |
| A41 | released |
| A42 | released |
| A43 | released (ruling) |
| A44 | new, low, optional: record each run's excluded files |
| A45 | new, low, optional: name the README isolation paragraph among the live documents |

**Held: none.** Every A-row I raised across rounds 1 to 7 is released or moot.

**Fixed point: confirmed from the architecture lens.** A44 and A45 are recorded for refinement. They are not
conditions of acceptance.

**Recommendation: accept.**
