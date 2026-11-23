# crochet:assess round 2: the architecture and direction lens

**Subject:** `docs/plans/2026-11-23-chart-precedence-in-recognizer.md`, revision 2, read at commit c270bbfa3
(branch `docs/chart-precedence-in-recognizer`, worktree cr-close).
**Blob read:** 68371487a68e3f8c1661a80ea830356228a3a163, checked with `git hash-object` on the live file.
**Recommendation: modify.** The remaining changes are small, and none of them changes the direction of the decision.

**A note on bias.** I raised A1–A11, and I am the only one who may release them, so I lean toward finding my own
bar met. To guard against that, each release below quotes the revision-2 text it rests on. Where a release
depends on a property the text asserts but does not secure, I hold the finding, or raise that property as a new
finding.

## What I read, and in what order

I read the subject first and ran my checks against it. I read the diff last.

- **Read in full:** the subject, and the round-1 minute (`docs/assessments/chart-precedence-in-recognizer.md`).
- **Read in part:**
  - the v2 design, through `git show origin/docs/chart-parser-v2-design:docs/plans/2026-10-30-chart-parser-v2-design.md`,
    §3's rules 6 and 12 (identical to the scratchpad copy I used in round 1, checked with `diff`);
  - the C1 report (`chart-recovery/c1-report.md`).
- **Code:**
  - `internal/chart/forest/forestdigest_test.go`: what each golden digests;
  - `internal/chart/recognize/expected.go`;
  - the `Options` doc comment in `recognize/recognize.go`.
- **The diff:** `git diff 7b264fa9c c270bbfa3 -- docs/plans/` is a full rewrite, 138 lines in and 195 out.

## Not re-examined

- **Round-1 sources:** I did not re-read the chart README's speed protocol, the postmortems or the w2 Records.
  I rely on my round-1 quotations from them.
- **Recorded rulings:** I did not check perigrin's 2026-11-14 rulings (D7, "The spec isn't in-volatile") or the
  2026-11-23 rulings in any Records. §1 says they will be recorded at refinement.
- **The recovery plan's D2:** not read.
- **Codebase claims:** I did not verify §3's counts, C1's figures, or whether the fuzz oracle can call the forest's
  constraint without an import cycle (see A13). Those belong to the codebase lens.
- **Nothing run:** no ratchet, no tests and no `make test`, as the brief says.

## A1–A11: release, hold or carry

| id | disposition | against what text, or why |
|---|---|---|
| A1 | **released**, with the residue raised as A14 | §2 names v2 §8 and v2 §2 and amends them "for one thing only". §1 dates every ruling and defines B ("B is §4") and C ("the value on items"). The value is out of scope (§7), so v2 §10's "facts in items" kill is no longer touched. The residue, two v2 clauses that §2 does not name, is A14. |
| A2 | **released** | The header cites v2 by path and branch. The 2026-10-29 first attempt and "design §13" are gone. Minor: the recovery plan is cited only as "perigrin's D2", with no path or branch (the README gives `docs/plans/2026-11-13-chart-parser-recovery-plan.md` on `docs/chart-recovery-plan`). |
| A3 | **released as overtaken**; its descendant is A15 | The value is out, and §3 now says, citing C1 (items per token 4.302 to 4.242), that "the cost of ambiguity is in the forest, not in the items". C1 confirms A3's doubt about items. The design no longer claims items will fall. |
| A4 | **released** | §1: "the three stage keys of §5 are ruled in ('Rule them in')". §5: "reported timing keys … never gated and never compared with the baseline". Removing `Constrain` is out of scope (§7), so the question of the precedence key's end state no longer arises. |
| A5 | **released**; A17 is for perigrin | The exit is now perigrin's goal (§1, ≤ 5), so the return of a speed gate is acknowledged by a ruling. §5 states a reading protocol. Grounding the number is perigrin's call; whether it means anything is A17. |
| A6 | **released**, with the residue raised as A16 | The stops now compare the after-reading's median with the before-reading's minimum or maximum. That is aware of noise, and it is no longer compared with the single-run `baseline.txt`. The README's rule I cited is about the ratchet comparing keys with `baseline.txt`, and a protocol reading follows chart-recovery's practice, so §5 and §6 are consistent with it. "Fires late" matters less in a decision with a 7-day budget and one mechanism. The residue, which session the before-reading is taken in, is A16. |
| A7 | **released** | §5 adds `tables-drift`, `ledger-gaps`, `collapse-over-limit` and `lex-unmapped ≤ 9`. It marks `oracle-exact-chart` as "reported". It carries the population caveat for `items-per-token` and `set-reuse` almost word for word. Minor: `precedence-covered` (may not fall) and `tdt-aligned` are not listed, but the ratchet still gates both. |
| A8 | **released**, with the contract's wording raised as A13 | §4.2 keeps the reference recognizer independent ("stays without precedence"). §4.3 says the filter does not run with lookahead off. `Expected` re-closes under every completion (`expected.go`: "as recognition without lookahead builds it", `b.complete(nil)`), so `Expected` is unchanged, provided the filter does not run under that nil re-closure. The design should say so in one clause. |
| A9 | **released** | §2: "The forest's precedence constraint stays the authority, and is unchanged". §7 puts its removal out of scope. With `Constrain` kept, an under-filtering table does no harm. Over-filtering shows in the referee (`judge` runs through `Build`), in `rd-accepted-chart-refused` and in the resolved digest. |
| A10 | **released** | §4's "What must become true" states properties (exactness, the reference contract, lookahead, accounting, Leo and ambiguous tokens). Function names and representations are gone. §4.5 settles the old §5.3 choice. Minor: "can follow it there" in §4 does not say what "there" keys on. The `indirob` pair needs it, and stop 2 and exactness cover the outcome. |
| A11 | **released** | §6: "The result decides one thing: perigrin's ruling on whether the shift side … is designed next." With no value on items, the round-1 concern about set identity for M2 does not arise (§2: rule 12 "items stay (state, distance) and carry nothing new"). §4.3 covers interleaving (no filter at lookahead 0), and §7 restates the isolation. This answers A11. |

**Carried to the value's later design:** none of mine is in §6's list (C1, C4, C5, C8, C9, P1, P2, P3, P10), and
none should be. I agree those nine concern only the value. A3's question about left-associative items is answered
by C1's data, so it does not need to carry.

## New findings

### A12. §4.1 and §6 stop 1: "the forest digests are unchanged" is false by construction, and the stop fires on the first build

- **Evidence.** `forestdigest_test.go` holds two goldens.
  - `forest-digest.golden` is "the hash of its forest's text, the forest Build makes" (`forestDigest`, which
    calls `built`). That is the forest *before* the constraint.
  - `resolved-forest-digest.golden` is the forest `Parse` leaves after the constraint and resolve.
- **The contradiction.** §4 says "The forest build refuses the same completions, so it never packs a derivation
  through one". `Build`'s forest therefore loses alternatives, and `forest-digest.golden` must move.
  - §4.1 says "The forest digests (`forest/testdata`) are unchanged".
  - Stop 1 fires on "a forest digest changes".
  - `TestForestDigestGolden` itself says "Never regenerate it to make a change pass: a change that moves it has
    changed the forest the chart builds".
- **Consequence if it ships.** The first issue either trips a stop by design, or regenerates a golden that the
  test forbids regenerating. Both put the gate in the wrong place.
- **Modify.** Hold `resolved-forest-digest.golden` unchanged; that golden is the exactness property. Let
  `forest-digest.golden` move, under a ruling. Require the diff to show only removed alternatives, each of them
  one that `Constrain` removed.

### A13. §4.2: the reference contract "exactly" is false for the `%nonassoc` and error pairs

- **The design's contract.** §4.2: "the recognizer accepts exactly what the reference accepts and the precedence
  constraint leaves a derivation for".
- **Why it fails.** The decision builds only the reduce side.
  - Take `1 .. 2 .. 3` (an error resolution). The filter refuses completing `1 .. 2` before `..`. The other
    grouping, `1 .. (2 .. 3)`, is blocked only on the shift side, so recognition still accepts the file, and
    `Constrain` then refuses it.
  - `++$x++` behaves the same way. Error resolutions "fail both sides" (README, "Forbid"), and this decision
    builds one of the two.
- **What the true contract is.** Constraint-accepted ⊆ recognizer-accepted ⊆ reference-accepted. Equality holds
  only with lookahead off.
- **Consequence if it ships.** A fuzz oracle written to the stated equality fails on the first `%nonassoc`
  chain. Or it gets weakened ad hoc, which is the outcome A8 warned about.
- **An open question for the codebase lens.** The fuzz oracle lives in package `recognize`. Holding the
  constraint there means calling into `forest`, which imports `recognize`.
- **Modify.** State the contract as that sandwich.

### A14. §2: the amendment misses two v2 clauses, and says nothing on where the live text changes

- **The clauses.** v2 §3 (on `origin/docs/chart-parser-v2-design`) says:
  - "Rule 6 is sharpened, not relaxed: a constraint derived from perl's own declarations may … refuse a span
    where none does, **after the chart (§8)**";
  - rule 12: "Scope … collapses what remains, after the chart in waypoint 2, **as precedence does (§8)**".
- **Why §2 is wrong as written.** Both clauses place precedence after the chart. §2 says "rule 12 is untouched"
  and does not name rule 6. Rule 12's (state, distance) clause is indeed untouched; its "as precedence does"
  clause is not.
- **What the decision does not say.**
  - v2 lives on an unmerged branch, so the decision does not say where the amended-by relation is written.
  - It does not say which statements in the live README become false: `Options`' "Neither changes acceptance"
    (§4.3 names this one), Status's "the precedence constraint over the forest", and The precedence constraint's
    "reads only the forest".
- **Consequence if it ships.** The same unrecorded reversal as A1, smaller in scope. Review at the end finds live
  text contradicting the build.
- **Modify.** Add rule 6's clause and rule 12's "as precedence does" clause to §2. Name the live statements that
  change.

### A15. §3: the evidence does not say how much B can remove, so the re-measure cannot be read

- **The gap.** §3 cites C1's "operator groupings were 47.9% of the forest's packed alternatives". That figure
  counts every alternative at ambiguous operator nodes, the correct ones included. It is not the share B removes.
- **What B cannot remove (reasoned, not measured).**
  - In a same-level left-associative chain such as `$a . $b . $c . $d`, the wrong groupings lean right, and the
    shift side blocks them. bison *reduces* `$b . $c` before `.`, so the reduce-side filter refuses nothing.
  - B removes groupings that complete too early: right-associative chains (`=`, `**`, `?:`), a tighter operator
    after a looser one (`$a + $b * $c`), and the error pairs' reduce half.
  - The chains common in real Perl (`.`, `+`, `&&`, `||`) are mostly the kind B leaves alone.
- **Consequence if it ships.** perigrin's "Do just B first and let's re-measure" produces a number with no
  account of how much of the forest B could ever touch. A small gain could then be misread as "precedence in
  recognition doesn't pay", when it actually means "the reduce side was the small half".
- **Modify.** Record two numbers beside the reading, as figures in Records rather than keys:
  - the packed alternatives the filter removed, as a share of C1's operator alternatives;
  - the precedence stage key's before and after.

### A16. §5: "before … and after … in one session" cannot both hold as written

- **The conflict.** The before-reading is taken "on `pu` with the stage keys", before any change. The
  after-reading is taken once the decision is built, up to 7 days later. Those two readings are in different
  sessions.
- **The precedent.** chart-recovery's What Will We Change says: "Record the host kernel with each quiet reading;
  a change voids the before reading. Compare only interleaved pairs from one session."
- **Undefined term.** §5 does not define "quiet". The README's working definition is a 1-minute load at or below
  the CPU count.
- **Consequence if it ships.** Stop 3 compares readings across sessions, and a change of kernel or load decides
  it.
- **Modify.**
  - Re-take the before-reading at the end, alternating with the after-reading in one session (the pu commit is
    kept).
  - Record the commit, the load, the core count and the kernel.
  - Define "quiet".

### A17 (for perigrin, not the author). §1 and §6: the exit of 5 may be met with no change

- **The readings so far.** `baseline.txt`'s single run reads 4.7915. Checkpoint 3's two readings were 5.18 and
  5.11, before REVIEWFIX. The before-reading on pu is not yet taken.
- **If it reads at or below 5,** the exit holds on the unchanged tree, so the exit says nothing about B. The test
  that actually means something is stop 3's second half: the after-reading's median below the before-reading's
  minimum.
- **If it reads well above 5** and B removes only the reduce-side share (A15), the exit may be unreachable within
  scope (§7 excludes speed work on resolve and readings).
- **For perigrin to rule.** Should the exit be read relative to the before-reading, for example "below its
  minimum, and at or below 5"? Or does an exit already met at the start count as met?

## Summary

| id | section | status | what it needs |
|---|---|---|---|
| A1 | §1, §2 | released (residue A14) | |
| A2 | header | released | minor: give the recovery plan's path |
| A3 | §3 | released, overtaken (see A15) | |
| A4 | §5 | released | |
| A5 | §6 | released (see A17) | |
| A6 | §6 | released (residue A16) | |
| A7 | §5 | released | minor: list `precedence-covered` |
| A8 | §4.2, §4.3 | released (wording A13) | one clause: `Expected`'s re-closure runs no filter |
| A9 | §2, §7 | released | |
| A10 | §4 | released | minor: say what "there" keys on |
| A11 | §6, §7 | released | |
| A12 | §4.1, §6 | new, must fix | Build's golden moves by construction; hold the resolved digest instead |
| A13 | §4.2 | new, must fix | the contract is a sandwich, not equality |
| A14 | §2 | new, should fix | rule 6's and rule 12's "after the chart" clauses; live statements |
| A15 | §3 | new, should fix | record B's share of what it can remove, so the re-measure can be read |
| A16 | §5 | new, should fix | one session; re-take the before-reading at the end; define "quiet" |
| A17 | §1, §6 | for perigrin | the exit of 5 may hold before any change |

**Recommendation: modify.**
- The narrowing to B answers A1–A11, and I have released all eleven.
- Revision 2 introduced two contradictions that break its own gates. A12: the stop on forest digests fires by
  construction. A13: the reference contract is an equality the decision cannot meet.
- A14–A16 are wording and protocol fixes.
- A17 is perigrin's to rule.
