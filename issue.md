---
title: "Decide a type for \\$'s scalar container"
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T01:42:07.024701879Z
updated: 2026-10-02T01:42:07.024701879Z
---

RFC 0001 open question 10. The prototype `\$` requires an actual scalar variable -- measured on 5.42, `sref(1)` dies "Type of arg 1 to main::sref must be scalar (not constant item)" -- where `$` takes any expression. Derivation needs a parameter type for the container, distinct from a scalar value, as `Array` is distinct from `List`.

perigrin decides; the decision lands in RFC 0001 as `### The scalar container (*Decided*)`.

## Acceptance criteria

- [ ] The RFC records the decision (`grep -qF '### The scalar container (*Decided*' docs/rfcs/0001-type-libraries.md`)
