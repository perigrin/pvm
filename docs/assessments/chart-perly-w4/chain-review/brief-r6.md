# chain-review chart-perly-w4: round 6 (resumed participants)

The subject is the live chain in `/home/user/pvm/.claude/worktrees/v2-refine`. It has 19 issues. Round 2's repair made 01a21463-3262 the single switch-over to the judged recognizer.

**What changed.** Read `/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/v2/review/repair-r5.md`. It is the repairer's log: one line per edit, each naming the finding it answers. The repairer answered the one held blocking finding, r4-fresh3 FR4-1, the same one raised as r5-fresh4 FR5-1 and r5-fresh5 FR5-1. d199 now states where the judged recognizer's precedence comes from:
- toys with levels pass it explicitly;
- `New` uses perly.y's precedence only when the tables match perly.y rule for rule, by content;
- level-less tables forbid nothing;
- a failed derivation on perly.y's tables refuses every input.

3262 and d218 mirror this. One call is flagged for you: the rule-for-rule match also compares right-hand sides, which is stricter than today's derivation check. Hunt especially for an off-switch this leaves, or for a toy suite it still breaks.

**How to verify.** Re-derive from the live chain, not from the log.
- Read the chain first, run your checks, then read the log.
- Before checking whether a fix landed, say what else it would have to change, and check that too.
- Look for what the repair broke: dependency edges, the residual mapping in the milestone's Records, carried ids lost in a split, and criteria that went green early.
- Re-run whatever proved your own findings.
- Read every residual row from round 1, every participant's rows, from r1-* to r5-*. Promote any row you now judge blocking, and give your grounds.

**For each of your round-1 findings, give one status:**
- released, with grounds;
- held, with what is still wrong;
- stood aside, with grounds.

Declare each new finding blocking or residual.

**Report.** Write it to `/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/v2/review/r6-<your-name>.md`. Include, in order:
1. your recommendation;
2. the status of each of your findings;
3. promotions;
4. new findings, as a table;
5. what you did not re-examine.

Make no edits. Your final reply is only the path, your recommendation and a count.
