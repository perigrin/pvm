---
title: "[C2] psc check and the LSP fetch from MetaCPAN when a version file says dev, latest-stable or @…, even with PVM_..."
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:06.994042355Z
updated: 2026-10-16T01:17:15.278282659Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:15.278282659Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:15.278282659Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **C2**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/perl/version.go:280-283` → `resolveLatestVersionWithDev` (:318-334), which calls `cpan.NewMetaCPANProvider()` with `context.Background()`.
- **Repro:**
  ```
  cd v/sec2; echo dev > .perl-version
  PVM_SKIP_NETWORK_CALLS=1 strace -f -e trace=connect psc check a.pl
  ```
  It makes 11 `connect` calls, including `1.1.1.1:53` and `151.101.x.217`.
- **Impact:**
  - This breaks the ruling that psc makes no network fetches.
  - An untrusted repository can make the checker phone home.
  - In the LSP the fetch runs under `LSPServer.mu` (write lock) on every keystroke, which stalls hover and diagnostics for up to the 30 s client timeout.
- **Fix:** the same as C1. resolveLatestVersionWithDev should also honour `SkipNetworkFetch` and `PVM_SKIP_NETWORK_CALLS`.
- **Confidence:** High. **NEW.** **Issue:** none. **Found by:** security (S2).

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] psc check and the LSP fetch from MetaCPAN when a version file says `dev`, `latest-stable` or `@…`, even with PVM_SKIP_NETWORK_CALLS=1 is fixed (`go test ./internal/psc/ -run '^TestPscNeverFetchesTheNetwork$' -count=1 -v | grep -q '^--- PASS: TestPscNeverFetchesTheNetwork'`)
