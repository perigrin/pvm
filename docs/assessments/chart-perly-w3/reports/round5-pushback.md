# Assess round 5: pushback lens, raiser dispositions (assess-pushback)

**Subject:** revision 5, commit 2782fd181, `docs/plans/2026-11-23-chart-stateless-precedence.md`. Its blob is d97415a147b923df589b27bed3cad0317f6bbfe2 (`git rev-parse` at the commit). I act only as raiser here. I am also the drafter of the minute, so I am biased toward finding my own bar met.

## P17: released

I release it against two passages:

- §3's heading: "Recognition and the built forest are unchanged."
- §3.4: "For every input, these equal today's: the parse verdict; the refusal reason and byte; `Expected`; the kept set; the readings count. … A differential test compares all five over … the precedence-refused fixtures (`waypoint2Refusals`, the port's refused cases)."

Both parts of P17 are answered.

- **The width of the test.** It now covers reason, byte and readings, and it includes precedence-refused inputs.
- **The tension between identical output and the recognition contract.** It is gone. Recognition no longer filters, so the reason cannot move from `precedence` to `grammar`, and §6 defers V2 by ruling.

## New in my lens

### P18 (medium-low; §1 and §5): the ruling "get ourselves to 3.92" is recorded, but the decision's exit does not mention it

§1 records "'Let's get ourselves to 3.92 first then look': allocation and the other stages come after this work." In the combined prototype, 3.92 is **V2's** whole-pipeline ratio (minute, "Experiments after round 3": ParseWith V0 5.62, V1 4.16, V2 3.92). Revision 5 ships V1 only, and §6 defers V2.

§5's exit is the relative exit and "≤ 5". It does not mention 3.92.

**If it ships:** refinement can read the 3.92 ruling as this work's target. On the prototype's machine, that is a figure only the out-of-scope V2 reached. If refinement writes it as a criterion, the criterion cannot be met within scope. Alternatively, the work could reach its exit while perigrin expects 3.92, and the result would read as a miss.

The figure was also taken on one machine, which §2 says is not comparable with `baseline.txt`. So "3.92" has no stated reading protocol of its own.

**Modify:** say in one sentence how the 3.92 ruling relates to this decision. For example: it is not this decision's exit, and it is reached only with V2. Or: perigrin's target for the sequence of decisions, read under §4's protocol.

**What would change my mind:** perigrin says 3.92 was meant as the V1-then-V2 sequence's target, and the decision states that. The finding then reduces to that one sentence.
