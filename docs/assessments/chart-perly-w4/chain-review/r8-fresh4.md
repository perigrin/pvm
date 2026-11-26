# chain-review round 8 (the bound): review-fresh4

I read the live chain (19 issues) from `/home/user/pvm/.claude/worktrees/v2-refine` into `fr8/`. I diffed it against my round-7 copy (`fr8/diff.txt`) and ran my checks. Only then did I read `repair-r7.md`.

What changed since round 7:
- The bodies of d166, d218 and 3262, and an appended entry in the milestone's Records.
- No edge, title or label changed, and no issue was added.
- No mapping row or `Carries:` line changed.

## 1. Recommendation: accept

I hold no blocking finding, and I raise no new one. The repair closes my last open point about the perl.git differential (r4-fresh4 FR4-3). It breaks nothing I can find.

## 2. Status of my findings

| Finding | Status | Grounds |
|---|---|---|
| r4-fresh4 FR4-1 | released (in round 6) | No change since. |
| r4-fresh4 FR4-2 | released (in round 7) | No change since. 3262 still names the cost bounds and routes a red bound to perigrin. |
| r4-fresh4 FR4-3 (`make test` never compiles the `precedencereference` test; d218 did not name which reference differentials it repoints) | **released** | d218 now names all six repointed differentials, `TestPrecedencePerlGitMatchesReference` among them. It also adds the step "Where perl.git is present, run `make chart-precedence-reference` and see TestPrecedencePerlGitMatchesReference pass". perl.git is present here (`/opt/corpora/perl5`). So the tagged file is compiled and run at the issue that changes its meaning. Before d218, a compile break in it after d135, d166 or 3262 is still caught only by CI's Ratchets job. The file is pinned and unchanged in meaning until d218, so I accept that. |
| r4-fresh4 FR4-4 (no run-first step on d1c2 and d1eb; unkeyed `Item` literals in `intern_test.go`) | **held, residual** | Unchanged. Its failures are loud: a compile error, or a golden that does not match. |
| r5-fresh4 FR5-1 | released (in round 6) | No change since. |
| r6-fresh4 FR6-1 | released (in round 7) | No change since. The `ParseWith` guard text in 3262 is unchanged by this repair. |
| r6-fresh4 FR6-2 (`Parse` over non-perly.y tables accepts unjudged where V1 refused) | **held, residual** | Unchanged. Production never builds such tables. Several participants hold the same row residual. |
| r7-fresh4 FR7-1 (d278 "reusing" the fuzz oracle's toy across packages) | **held, residual** | Unchanged. It is a wording point that the builder resolves by duplicating the toy or moving it to a shared helper. |

## 3. Promotions

None. I read every residual row from r1 to r7, including all of round 7's.

**r7-fresh7 FR7-1** (bb9b item 8 would make the README say the union shortcut holds "for `Constrain` only"). I agree with the substance. V2's Build keeps one node per (symbol, span), and forest equality, through d0ba's `TestForestEqualityUnionShortcut` and d218's real differential, requires the union at a node reached under two live contexts. So the shortcut holds for production too, and item 8's sentence would be false. I keep it residual for three reasons:
- No criterion asserts the sentence.
- bb9b's own text says "a sentence this issue cannot make true without a code change is reported, not papered over".
- Its first step reads the merged diff, and `crochet:review` reads live documents against the diff.

It should be reworded before bb9b runs.

**r7-fresh7 FR7-2** (3282's dead-context test is "on the union-shortcut fixture", which d0ba defines with two live contexts). The criterion's command runs only `TestBuildDropsAlternativeOfDeadContext`, and nothing ties that test to d0ba's fixture file. So 3282 can add its own dead-context input, and nothing that d0ba's criterion checks has to change. The wording invites a literal reading that would amend d0ba, but that reading is not forced. Residual; reword it to "a fixture of its own beside the union-shortcut fixture".

**The other round-7 rows.** FR7-3 to FR7-11, AL18, r7-fresh2 FR7-1, r7-fresh5 FR7-1 and FR7-2, r7-fresh6 FR7-1 and FR7-2, and N16 and N17 are all scope or wording points, size, or environment notes. Size here means FR7-9; environment means FR7-11, since gcc and cgo are present here, so `-race` runs. In each case the failure mode is a test that goes red, which the builder can see, or a report. None makes a criterion unpassable or green before its work exists.

## 4. New findings

None.

### What I checked in the repair

- **Every test d218 names exists today, so none of its criteria is unpassable.** I found each by `func` name:
  - the six it repoints: `TestPrecedenceToyForestsMatchReference`, `TestPrecedenceOperatorSentencesMatchReference` and `TestPrecedenceChainSentencesMatchReference` in `forest/precedence_reference_test.go`; `TestPrecedenceCorpusMatchesReference` in `cmd/ratchet`; `TestPrecedencePerlGitMatchesReference`, behind the tag; and `TestWaypoint2RefusalsMatchReference` in `chart`;
  - the instruments it keeps on the no-precedence mode: `TestPrecedenceDiffDetectsDifference`, `TestComparePrecedenceUsesScope` and `TestPrecedenceHarnessesDoNotAlias`.
- **The new criteria are red today.**
  - No file under `internal/chart` contains the text "verdict only".
  - `TestNarrowedSetIdsStableAcrossFiles` does not exist.
- **d218's "waypoint2 verdict-only N > 0" clause can pass.** `waypoint2Refusals` (`waypoint2_refusals_test.go:15-44`) holds 11 perl-refused precedence rows, among them `$a <=> $b <=> $c;`, `1 .. 2 .. 3;` and `print $fh [];`. The judged chart refuses those, so N is non-zero once d218 repoints the test. Today the same test errors "recognition leaves no forest" on any refused row, and d218 replaces that with the verdict-only rule.
- **3262's waypoint2 criterion still holds at verify.** Its command is unchanged: it runs `TestWaypoint2RefusalsMatchReference` by name. That test keeps its name and passes before d218 (pinned) and after it (repointed, verdict-only).
- **The d166 addition is self-contained.** The long-lived recognizer test is a new criterion with its own TDD step. It touches no other issue's criteria.

## 5. Not re-examined

- Issues other than d166, d218 and 3262. Their bodies are byte-identical to round 7's.
- The verdict-only rule on scope rows in `TestPrecedenceCorpusMatchesReference`. The criterion allows N to be 0 or more.
- I ran no code this round. My earlier measurements are unchanged, because the tree is unchanged: the two bound tests, and counters 4 and 5 on the toys.
