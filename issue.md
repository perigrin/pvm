---
title: Troubleshooting help points at type annotations and a pvm build that is not available
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-01T16:35:40.021701398Z
updated: 2026-10-16T18:48:30.111292986Z
---

internal/data/help/troubleshooting.md's "Build Issues" entry says "Check PSC configuration and fix type annotations" and points at `pvm build --check-only`. Both are stale: type annotations in source were dropped in the March pure-Go rewrite (PVM infers types; typed Perl is accepted only in .tpm declaration files, RFC 0001), and `pvm build` now reports "The build command requires type-system components that are not available in this build." Found while rewriting the types help page; not changed there.

## Acceptance criteria

- [ ] The troubleshooting page names no type annotations (`sh -c '! grep -q "type annotations" internal/data/help/troubleshooting.md'`)
