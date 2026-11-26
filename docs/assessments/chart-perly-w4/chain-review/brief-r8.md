# chain-review chart-perly-w4: round 8 (resumed participants)

The subject is the live chain in `/home/user/pvm/.claude/worktrees/v2-refine`. It has 19 issues. Round 2's repair made 01a21463-3262 the single switch-over to the judged recognizer.

**What changed.** Read `/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/v2/review/repair-r7.md`. It is the repairer's log: one line per edit, each naming the finding it answers. The repairer answered two blocking findings.

- **d218** now names the six reference differentials it repoints. Inputs the judged chart refuses follow d247's verdict-only rule. This answers r6/r7-fresh2 FR4-1 and r6/r7-fresh5 FR5-3.
- **d166** states that a narrowed set id is stable for a recognizer's lifetime, with `TestNarrowedSetIdsStableAcrossFiles`. This answers r6-fresh6 FR6-6.

**This is round 8, the bound.** If it does not converge, the session ends here, and the open items go to perigrin.

The orchestrator flags two of round 7's residual rows for you to read when you consider promotions:
- **r7-fresh7 FR7-1:** item 8 of the docs issue may write a false statement into the README.
- **r7-fresh7 FR7-2:** 3282's test needs a fixture that d0ba defines differently.

Judge them on their merits. This flag is not a request to promote them.

**How to verify.** Re-derive from the live chain, not from the log.
- Read the chain first, run your checks, then read the log.
- Before checking whether a fix landed, say what else it would have to change, and check that too.
- Look for what the repair broke: dependency edges, the residual mapping in the milestone's Records, carried ids lost in a split, and criteria that went green early.
- Re-run whatever proved your own findings.
- Read every residual row from round 1, every participant's rows, from r1-* to r7-*. Promote any row you now judge blocking, and give your grounds.

**For each of your round-1 findings, give one status:**
- released, with grounds;
- held, with what is still wrong;
- stood aside, with grounds.

Declare each new finding blocking or residual.

**Report.** Write it to `/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/v2/review/r8-<your-name>.md`. Include, in order:
1. your recommendation;
2. the status of each of your findings;
3. promotions;
4. new findings, as a table;
5. what you did not re-examine.

Make no edits. Your final reply is only the path, your recommendation and a count.
