# chain-review chart-perly-w4: round 7 (resumed participants)

The subject is the live chain in `/home/user/pvm/.claude/worktrees/v2-refine`. It has 19 issues. Round 2's repair made 01a21463-3262 the single switch-over to the judged recognizer.

**What changed.** Read `/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/v2/review/repair-r6.md`. It is the repairer's log: one line per edit, each naming the finding it answers. The repairer answered four blocking findings, in 3262, d218, d199 and d278.

- **`ParseWith` off-switch** (r6-fresh4 FR6-1, r6-fresh5 FR5-1, r6-fresh6 FR6-1). `ParseWith` now panics unless the recognizer's precedence is what `New(tab)` would derive. Subtests cover empty precedence over the committed tables and over a copy of them.
- **3262's list** (r6-fresh5 FR6-1, AL17). The list is split three ways: toys, perly.y-backed tests that judge, and named cost bounds. A red bound is reported to perigrin with its number, never changed without a ruling.
- **Leveled toys** (r6-pushback N14). They are judged through the explicit form, except `empty` and `emptyred`.
- **Brute-force reference** (r6-fresh6 FR6-2). The brute force gains a feasible leveled toy, and a criterion requires a non-zero refusal count.

This is round 7 of a bound of 8.

**How to verify.** Re-derive from the live chain, not from the log.
- Read the chain first, run your checks, then read the log.
- Before checking whether a fix landed, say what else it would have to change, and check that too.
- Look for what the repair broke: dependency edges, the residual mapping in the milestone's Records, carried ids lost in a split, and criteria that went green early.
- Re-run whatever proved your own findings.
- Read every residual row from round 1, every participant's rows, from r1-* to r6-*. Promote any row you now judge blocking, and give your grounds.

**For each of your round-1 findings, give one status:**
- released, with grounds;
- held, with what is still wrong;
- stood aside, with grounds.

Declare each new finding blocking or residual.

**Report.** Write it to `/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/v2/review/r7-<your-name>.md`. Include, in order:
1. your recommendation;
2. the status of each of your findings;
3. promotions;
4. new findings, as a table;
5. what you did not re-examine.

Make no edits. Your final reply is only the path, your recommendation and a count.
