---
title: "[I22] The term-position * split refuses or misparses ** and *= after __LINE__, do {}, eval {} and unresolved cons..."
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.994388424Z
updated: 2026-10-16T01:17:16.066990383Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:16.066990383Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:16.066990383Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I22**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/parse/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/lexer/scan.go:1105-1114`, the new term-position star added in this PR (the diff confirms it).
- **Repro:**
  - `my $y = __LINE__ ** 2;` lowers to `Stmts: Hole`.
  - `my $y = do {1} ** 2;` lowers to `Hole`.
  - In `v/pi.pl`, `use Math::Trig; my $y = pi ** 2;` lowers to `Hole` when `pi` is unresolved, which is common because psc parse has no @INC and psc check often has an empty one.
  - `print *STDOUT **2;` lowers to `Binary * (Var *STDOUT) (Var *2)`.
  - perl Deparse gives `(my($y) = 1.0)`, `do {1} ** 2`, and `print((*STDOUT ** 2))`.
- **Fix:**
  - Put `__LINE__`, `__FILE__`, `__PACKAGE__` and `__SUB__` in the term-ending table.
  - Set XOperator after the `}` of `do`/`eval BLOCK`.
  - Keep `**` as one token after a Word.
- **Confidence:** High for the behaviour. For NEW-ness, the lexer diff shows the branch is new; I did not rebuild origin/pu's parse to compare canon, so I relied on logic-A's comparison. **NEW.** **Issue:** none. **Found by:** logic-A #6.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] The term-position `*` split refuses or misparses `**` and `*=` after `__LINE__`, `do {}`, `eval {}` and unresolved constants is fixed (`go test ./internal/parse/ -run '^TestPowerAfterNiladicTerm$' -count=1 -v | grep -q '^--- PASS: TestPowerAfterNiladicTerm'`)
