---
title: "`isa Local::Child(...)` refuses: isa is an operator even without its feature"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T18:13:36.275883211Z
updated: 2026-09-30T18:13:36.275883211Z
---

`ok(isa Local::Child('Local::Child'), ...)` refuses. `isa` is in the infix table unconditionally, so parseIndirect never reads it as a method. keywords.c:353 returns KEY_isa only under FEATURE_ISA_IS_ENABLED, and without that feature `isa` is an ordinary word.

Measured with perl 5.42 and -MO=Deparse: the line is `&ok(scalar 'Local::Child'->isa('Local::Child'), ...)`. With `use feature "isa"`, or `use v5.35` or later (feature.pm's 5.35 bundle names isa), `$x isa Foo` is the infix operator; under `use v5.34` it's a syntax error.

PerlOnJava unit/indirect_isa_feature_gate.t:10.

## Acceptance criteria

- [ ] `isa` is a word without its feature and an operator with it (`go test ./internal/parse/ -run '^TestIsaWithoutFeature$' -count=1 -v | grep -q '^--- PASS: TestIsaWithoutFeature'`)
