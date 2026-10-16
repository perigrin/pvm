---
title: "[C4] Deep bracket nesting overflows the Go stack, a fatal error no recover can catch"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.084568278Z
updated: 2026-10-16T01:17:15.939684722Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:15.939684722Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:15.939684722Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **C4**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/parse/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

**Ruled 2026-10-16 (perigrin): fix it.** A nesting-depth cap in the parser turns input deeper than the cap into an Unknown/Hole rather than overflowing the stack; the LSP cannot recover a fatal error.

### The verified finding

- **Where:** `internal/parse` recursion: parseExpr → parseTerm → parseParenList/parseBracketed → parseExpr. There is no depth limit.
- **Repro:**
  ```
  perl -e 'print "my \$x = ", "(" x 400000, "1", ")" x 400000, ";\n"' > deep.pl
  ```
  `perl -c deep.pl` says `syntax OK`. `psc check deep.pl` prints `runtime: goroutine stack exceeds 1000000000-byte limit / fatal error: stack overflow`.
- **Impact:** one 800 KB file in a workspace kills the LSP, and an editor restarts it straight back into the same crash. The input is pathological, so it is rated Critical only because the rubric says a crash reachable from the LSP is Critical. The caller may downgrade it.
- **Fix:** add a nesting-depth counter in the parser (parseBlock, parseParenList, parseBracketed, statement). Past a cap, return a refusal (Unknown) instead of recursing.
- **Confidence:** High. **NEW in reach.** **Issue:** none. **Found by:** security (S3).

---

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Deep bracket nesting overflows the Go stack, a fatal error no recover can catch is fixed (`go test ./internal/parse/ -run '^TestDeepNestingIsBounded$' -count=1 -v | grep -q '^--- PASS: TestDeepNestingIsBounded'`)
