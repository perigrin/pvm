---
title: "Tier 14_recursive: re-enter Perl inside a delimiter"
state: pending
urgency: normal
milestone: m3-conformance-corpus
created: 2026-09-21T09:53:00.137066905Z
updated: 2026-09-21T13:40:22.372444523Z
---

**Introduces:** s///e, (?{ }), qr// with embedded code

Where the lexer must RE-ENTER Perl inside a delimiter, which is harder than
tier 13's delimit-only problem and structurally different. That difference
is why these are two tiers rather than one.

The last tier, and the one most likely to refuse wholesale. That is the
expected outcome, not a failure: the corpus states where the parser is, and
a tier that refuses entirely is a true statement about where it is.

## Acceptance

- [ ] Every file in 14_recursive/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierRecursivePerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierRecursivePerlValidated'`)
- [ ] The dependency lint passes over 14_recursive/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierRecursiveLint$' -count=1 -v | grep -q '^--- PASS: TestTierRecursiveLint'`)
- [ ] 14_recursive/README.md states what the tier introduces, what it depends on, and why check 2 places it here (`test -s conformance/14_recursive/README.md`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierRecursiveRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierRecursiveRefusalsCited'`)
- [ ] The whole corpus is green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
