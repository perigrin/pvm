---
title: "The dependency lint: ops-subset, because the optimiser erases what a tier is about"
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c35d-6e43-7c0d-8828-ef7659402e39
- 01a0c35f-a487-7aa8-934b-4e68dd0fbaa1
blocks:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
created: 2026-09-21T09:48:59.000983906Z
updated: 2026-09-21T15:28:17.858955343Z
sessions:
- start_sha: 6adae2c6f97a74feac2d0ba5139105feb486e7a1
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T15:25:42.377373038Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T15:25:42.377373038Z
---

Without this check the tier numbering is an opinion. With it, "this file
sits at tier 04" is a claim that fails when false.

    ops(file) subset-of union ops(tiers <= N)

The oracle already collects `Ops` from `B::Concise`, so this is
parser-independent and already this project's method. That matters: the
parser cannot supply the construct set, because it cannot classify a tier-09
file while it is still at tier 02.

## The caveat is sharper than "necessary but not sufficient"

**The optimiser erases the construct the tier is about.** `my $x = 1+2`
arrives as `const[IV 3] s/FOLD` with no `add` op. A literals-tier file and
an arithmetic-tier file can produce IDENTICAL op lists. Ops alone cannot
derive a tier -- they can only catch a file using what it did not declare.

Three sources, in layers:

    the file DECLARES   the `# TIER` / `# INTRODUCES` header. Authored
                        files, so the declaration is free.
    ops LINT it         catches undeclared use. Necessary, not sufficient.
    the token stream    classifies tiers 01-03, whose constructs are
                        lexical and where a declaration checks itself.

The CST is NOT a fourth source yet. It becomes one once trusted, and then it
replaces the DECLARATION, not the lint.

## Acceptance

- [ ] A file declaring tier 01 while using a tier-04 construct fails the lint, naming the op and the tier that introduces it (`go test ./internal/conformance/ -run '^TestLintRejectsUndeclaredConstruct$' -count=1 -v | grep -q '^--- PASS: TestLintRejectsUndeclaredConstruct'`)
- [ ] A file whose ops are a subset passes, including one where the optimiser folded the construct away (`go test ./internal/conformance/ -run '^TestLintToleratesFoldedOps$' -count=1 -v | grep -q '^--- PASS: TestLintToleratesFoldedOps'`)
- [ ] The op-to-tier table is derived from the tier READMEs rather than maintained as a second hand-written list (`go test ./internal/conformance/ -run '^TestOpTableDerivedFromReadmes$' -count=1 -v | grep -q '^--- PASS: TestOpTableDerivedFromReadmes'`)
- [ ] An op that no tier claims is reported rather than defaulting to allowed or denied (`go test ./internal/conformance/ -run '^TestUnclaimedOpReported$' -count=1 -v | grep -q '^--- PASS: TestUnclaimedOpReported'`)
