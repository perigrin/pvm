---
title: "[I10] my $x OP= EXPR loses OP in the parser; list and attribute declarations with compound ops lose their shape i..."
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.491801427Z
updated: 2026-10-16T01:17:16.004571748Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:16.004571748Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:16.004571748Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I10**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/parse/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:**
  - `internal/parse/decl.go:791-813` records only the initialiser.
  - `internal/lower/decl.go:31-75`: a non-`=` operator in list form puts an `Assign` into `Decl.Vars`, and the attribute path never sets `ListAssign`.
- **Repro:** `psc parse --tree lowered`:
  - `my $x -= 1;` gives `Decl Vars:$x Init:(NumLit 1)`, the same tree as `my $x = 1`. perl gives -1.
  - `my ($x) ||= 1;` gives `Decl Vars:(Assign Op="||=" ...)`.
  - `my ($x, $y) :shared = f();` gives `Decl Attrs=[:shared]` with no ListAssign, so declContext returns UnknownCtx where perl uses list context.
- **Fix:** record the operator on the Declaration, or build `[binary OP target init]`. In lowering, accept any assign operator in the one-child branch, and set ListAssign when the single Var is a Paren list.
- **Confidence:** High.
- **NEW or PRE-EXISTING:** the parser part is PRE-EXISTING, and this PR's lowering and inference are its first consumers. The lowering part is NEW.
- **Issue:** none. 01a110e9-e23e-7c99-b7db-39ffee6d567d ("Two declaration comma shapes still disagree with perl") is adjacent but different.
- **Found by:** logic-A #3, logic-A #8.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `my $x OP= EXPR` loses OP in the parser; list and attribute declarations with compound ops lose their shape in lowering is fixed (`go test ./internal/parse/ -run '^TestDeclCompoundOperatorKept$' -count=1 -v | grep -q '^--- PASS: TestDeclCompoundOperatorKept'`)
