---
title: "Tier 08_references: deref, arrow, anonymous constructors"
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c361-6191-74b6-a1c0-263f58d465b3
blocks:
- 01a0c362-0daa-7422-b787-7650ce684c4b
created: 2026-09-21T09:52:57.12910003Z
updated: 2026-09-21T13:40:24.094729075Z
---

**Introduces:** `\`, deref, arrow, anonymous constructors, `${ }`, `@{ }`

After subroutines, following *Modern Perl*: dereference syntax is harder
than a sub declaration even though `\@a` alone looks simpler.

Two `hardMarkers` place here -- deref-brace and deref-at.

## This tier is a prerequisite for 11_oo, which is the non-obvious edge

`bless` takes a REFERENCE and a string. Every blessed-hash object is this
tier plus a string, which is why the object tier depends on this one and NOT
on packages. That dependency is real and satisfied three tiers earlier.

A known defect to cover rather than discover: a brace before `->` is a
hashref, and the tree currently misreads two anonhash cases (M1 issue
01a0bae2).

## Acceptance

- [ ] Every file in 08_references/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierReferencesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierReferencesPerlValidated'`)
- [ ] The dependency lint passes over 08_references/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierReferencesLint$' -count=1 -v | grep -q '^--- PASS: TestTierReferencesLint'`)
- [ ] 08_references/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`test -s conformance/08_references/README.md`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierReferencesRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierReferencesRefusalsCited'`)
- [ ] The whole corpus is green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
