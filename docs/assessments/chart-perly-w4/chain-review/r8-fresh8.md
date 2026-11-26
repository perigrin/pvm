# chain-review round 8: fresh reader (review-fresh8)

Subject: milestone `chart-perly-w4`, read live from `/home/user/pvm/.claude/worktrees/v2-refine` (19 issues, all pending). No `r1-*` to `r8-*` file or `repair-*` file was read, and I did not rely on the milestone's chain-review Records.

## Recommendation: accept

I found no blocker. I checked each of the four blocker kinds against the live chain and the tree at `e3432c0f`:

- **Criteria that cannot pass, or pass before their work exists.** Every criterion is gated on one of these, so each is red at `e3432c0f`: a test name that does not exist yet, a file that is absent, or a guard that requires a new test or `prectab` (for example the `resolved-forest-digest`, `refused-where.txt` and `baseline.txt` diffs).
  - Every existing symbol a criterion names exists today: `TestItemIsStateDistanceR2`/`F2`, `TestPrecedenceIgnoredByRecognition`, `TestRecognitionKeysUnchangedByPipeline`, `TestForestGrammarDataOncePerTables`, `TestPrecedenceGenerationErrorRefusesEveryFile`, `grammar.Derived`, `buildParsePrecedence`, the MarkFirstFilter and `nullableShift` pinned toys, the 13 toys, and `FUZZTIME` on `fuzz-chart-recognize`.
  - The measured numbers that criteria quote match the tree or the minute: 1,296 referee sentences with 19 refused (I ran this), 30 waypoint2 rows including precedence rows, 27 answer keys in `baseline.txt` under d2d1's exclusion regex, the w3 `.meta` format, and the 1,212/1,224 core items.
  - The awk, sed, `git log` and `git show --name-only` mechanics in d2a2, d278 and d2d1 behave as intended against the repository's history shape. That history has no merge commits in golden logs.
- **Ordering.**
  - The three stages are wired by hard edges: d02f → d05d → d089 → d0ba → {d0e1, d10d} → d135.
  - Stage 3 then runs d166 → d199 → d1c2 → 3282 → 3262 → {d1eb, d218} → {d247, d278, d2a2} → d2d1 → bb9b.
  - I traced each intermediate state, and each can close green:
    - `recognize.New` stays unjudged until 3262.
    - The V1 side is fixed to the no-precedence constructor at d199.
    - `ParseWith` keeps `Constrain` until d218.
    - The stage harness keeps `Constrain` until d2a2, after d218, and running it twice is idempotent.
    - The referee keeps `Constrain` until d247.
  - None of the ratchet's 27 answer keys reads `Recognized`, so narrowing and judging recognition in `chartInputs` (d166, 3262) moves only the exempt keys `items-per-token` and `set-reuse`. `lex-unmapped` comes from `lex.Lex` alone (`lexUnmapped`), so neither change touches it.
- **Absent corpora or other repositories.** Every perl.git criterion is guarded by `[ ! -d "${PERL5_CORPUS:-/opt/corpora/perl5}/t" ]` or forces `PERL5_CORPUS=/nonexistent`. `488e1c7e`, the scratch-code source, is an ancestor of `e3432c0f`. pu has moved past `e3432c0f` only by #649, which touches nothing under `internal/chart` or the Makefile.
- **Gaps against the design, rulings and residual rows.** The milestone maps all 66 residual rows plus the F2-11 crumb, and every id in the minute's list appears in the table. Design §4.4, §4.5, §5 and §7 each have an owning issue. The rulings are carried: "Narrow before recognition", "Keep 'precedence'", "Lift rule 12", "Regenerate everything" narrowed by "Hold golden, retire key", "Diagnose by rerun" and "Relative only".

## Findings

| id | blocking/residual | issue id | claim | consequence | evidence | suggested action |
|---|---|---|---|---|---|---|
| FR8-1 | residual | 01a21447-d199 | The reduce-side issue has grown back to about the size round 1 split it for. It has 17 criteria and 8 TDD cycles, and it touches `recognize.go`, `sets.go`, `cache.go`, `leo.go`, `leo_memo.go`, Build, `equality.go` and `prectab`. It adds two constructors, an explicit and a content-derived precedence source, the `Feasible` refusal path, refusal records, Leo, and a corpus differential. | The issue risks spilling over one session. A partial close would leave the precedence-source half untested against a real judgement. | The body is 16,189 characters, the largest after 3262. Rounds 4 to 6 added the derived form, the `Feasible` refusal and the copy-of-tables criteria. | No chain change is required. Tell the executor that a stall here should be reported to the orchestrator rather than closed partly done. If it is split later, the precedence-source half is the separable piece: constructors, explicit and derived forms, `Feasible` and failed derivation. |
| FR8-2 | residual | 01a2144c-bb9b | Two of the docs issue's criteria will probably be green before the issue starts: "No V1-only claim survives" and "The Makefile no longer says production runs Constrain". | Each code issue's first commit must correct the README lines its change falsifies, so these lines go before the docs issue starts: d135 removes `forest.NewPrecedence`, 3262 removes `TestPrecedenceIgnoredByRecognition`, d2a2 removes the retired key's row, and d218 may fix the Makefile comment. These criteria are then regression guards, not evidence of the docs issue's own rewrite. The issue's other criteria (V2 decision, branch, rule 12 and rule 2, lookahead, linear lattices, `prectab` section, verdict citation, every new test named) are red at `e3432c0f`, so the issue is still checked. | I confirmed each of the issue's other criteria is red on the README and Makefile at `e3432c0f`. | None needed. Optionally, note in the issue that these two are guards. |
| FR8-3 | residual | 01a21447-d089 | The set-1626 observer counts tokens whose committed terminal-set id is 1626. It is written before narrowing exists (d166), and no later issue reruns it through the narrowed lattice. The §7 stop is stated post-narrowing. | Suppose a lexed set held HASHBRACK and PERLY_BRACE_OPEN plus a `PLUGIN_*` reading, and narrowed to set 1626's contents. The observer would not see it. The risk is low, because `{` is not a library operator. | d089's Context and d166's set-id mechanism. | I did not check the committed sets for such a superset. The builder of d166 could confirm that no narrowed set equals 1626's contents under another id. |

Count: 0 blocking, 3 residual.

## What I did not examine

- I ran only the referee and waypoint2 tests. I did not run `make test`, the counters or any criterion command against a built change, since none exists yet.
- I did not recompute the feasibility counts (649, 18, 674, 29, 1,880, 63). I took them from `feasibility.md` and the minute.
- I did not judge whether the design's forest-equality claim holds through Build-by-context and V1's union shortcut. That is a design §7 stop to be observed, not a chain defect.
- I did not check whether every toy other than `empty` and `emptyred` passes `prectab.Feasible`. d218's wording excepts "every toy Feasible refuses", so another refusal would extend the exception rather than break a criterion.
- I did not check whether 3262's list of tests that move is complete. Its first step runs `make test` and reports any red test missing from the list, by design.
- I did not read the full ledger rows behind each residual. I read only the one-line forms in the minute's "Residual rows for refinement".
