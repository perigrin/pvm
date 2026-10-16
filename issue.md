---
title: "[I28] psc check exits 0 and says nothing on code perl refuses to compile; the LSP publishes nothing for it either"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.30000356Z
updated: 2026-10-16T01:16:08.30000356Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I28**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

**Ruled 2026-10-16 (perigrin): report refusals.** A parser refusal becomes a psc check (and LSP) diagnostic with its refusal code and span, and psc check exits non-zero; perl refuses the code, so the user should hear it. 01a0c115-a44b's "says nothing about the Hole" negative scenario changes to match (update its test). The refusal replaces the unreachable lowered arity check: close 01a13891 ("The lowered arity check is unreachable") with this issue, deleting `CodeArityMismatch` if nothing else reaches it.

### The verified finding

- **Where:** `internal/infer/infer.go:96-100` types a Hole as Any. Refusal codes on the CST are never surfaced. `test/e2e/psc_check_test.go:86-88` asserts that silence.
- **Repro:**
  - `logicB/h1.pl` (`push(); my $x = length();`) gives rc 0 and no output. origin/pu reported `call to push: expected at least 1 argument(s), got 0 [arity-mismatch]`. perl reports `Not enough arguments for push`.
  - `v/unc.pl` (`if ($x) { print "a";`) gives rc 0. perl reports `had compilation errors`.
- **Impact:** a regression against origin/pu for arity. In the editor, a syntax error shows no diagnostic at all.
- **Ruling needed:** cutover issue 01a0c115-a44b-7393-a828-ad29844b20b4's negative scenario says "psc check on a file perl rejects does not panic and says nothing about the Hole", so silence inside the Hole is planned. But 01a13891-f317-726e-8955-432dfdbd95e5 says "The parser's refusal is the arity diagnostic now; confirm psc check surfaces it". The two disagree, and the maintainer should rule.
- **Fix:** emit one `syntax` diagnostic per top-level Hole, carrying its refusal code.
- **Confidence:** High that it happens; Medium that it is a bug (it depends on the ruling). **NEW in behaviour.**
- **Issue:** 01a13891-f317-726e-8955-432dfdbd95e5 covers the arity case. This finding adds the general refused-syntax case and the LSP side.
- **Found by:** logic-B #15, errors-B #12.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] psc check exits 0 and says nothing on code perl refuses to compile; the LSP publishes nothing for it either is fixed (`go test ./internal/psc/ -run '^TestRefusalIsReported$' -count=1 -v | grep -q '^--- PASS: TestRefusalIsReported'`)
