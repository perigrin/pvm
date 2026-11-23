# Assess round 3: pushback lens, raiser dispositions (assess-pushback)

- **Subject:** revision 3, commit 3e4d0c43b, "Chart parser: the precedence constraint without maps". I read it with `git show 3e4d0c43b:docs/plans/2026-11-23-chart-precedence-in-recognizer.md`.
- **Role:** I act only as raiser of P13–P15. I am the drafter of the minute, and I am biased toward finding my own bar met.

## Dispositions

| id | disposition | against what text, or why |
|---|---|---|
| P13 | **Released** | §1: "'Relative and ≤ 5': the exit is met when the after-reading's median is below the before-reading's minimum, taken interleaved in one session, and at or below 5". §4: "Exit: … perigrin's relative exit met by `speed-ratio-aggregate-longlived`". The exit can no longer be met with zero gain. |
| P14 | **Moot: the approach is retired** | Revision 3 makes no change to recognition or to `forest.Build`, so nothing moves `forest-digest.golden` by design. §5 retires precedence in the recognizer. Holding both goldens unchanged (§3.1) is now consistent with a change confined to `Constrain`. |
| P15 | **Moot: the approach is retired** | There is no recognizer filter and no reference contract in revision 3 (§5). |

## Does revision 3 raise anything in my lens?

P16 (low, §3.2): "no hash map" doesn't say whether it means Go's built-in `map` or any hashed structure. One implementer could replace `map` with a hand-rolled hash table, and another could use dense tables indexed by node and context id. Whether §3.2 is met would then be decided by whoever implements first. The fix is to state the property, for example "no Go `map` and no hashed lookup", or "lookups by dense index".
