---
title: "[I20] $x =~ qr/.../ lowers to a qr constructor, typed Regex, not to a match"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.89623416Z
updated: 2026-10-16T01:17:16.391681936Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:16.391681936Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:16.391681936Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I20**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/lower/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/lower/special.go:225-235` (`bind` copies `matchOp`, which can be `qr`). `internal/infer/special.go:116` then types every `Op=="qr"` as Regex.
- **Repro:** `v/qr.pl`, run with `psc check --warn-lossy-coercion`:
  ```perl
  my $r = $x =~ qr/b/; $r + 1
  ```
  It reports `warning: lossy coercion of Regex to Num`, and `($x =~ qr/z/) . "y"` reports `Regex to Str`. perl prints `2 y`. Hover shows Regex.
- **Fix:** in `bind`, map `qr` to `m`.
- **Confidence:** High. **NEW.** **Issue:** none. **Found by:** logic-A #1.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `$x =~ qr/.../` lowers to a qr constructor, typed Regex, not to a match is fixed (`go test ./internal/lower/ -run '^TestBindQrIsAMatch$' -count=1 -v | grep -q '^--- PASS: TestBindQrIsAMatch'`)
