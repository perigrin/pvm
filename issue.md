---
title: "`isa Local::Child(...)` refuses: isa is an operator even without its feature"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T18:13:36.275883211Z
updated: 2026-09-30T18:20:23.239976808Z
sessions:
- start_sha: 089821909379f1e6551fefc022bf8887661d4cf4
  end_sha: 1a15ffd68adb98185d45dc93c6585a6cc8a8b33e
  commits: 1
  started_at: 2026-09-30T18:13:38.399245594Z
  ended_at: 2026-09-30T18:20:23.239976808Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T18:13:38.399245594Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-30T18:20:23.239976808Z
observed_paths:
- internal/parse/indirect.go
- internal/parse/isaindirect_test.go
- internal/parse/testdata/t1.ratchet
- internal/parse/use.go
---

`ok(isa Local::Child('Local::Child'), ...)` refuses. `isa` is in the infix table unconditionally, so parseIndirect never reads it as a method. keywords.c:353 returns KEY_isa only under FEATURE_ISA_IS_ENABLED, and without that feature `isa` is an ordinary word.

Measured with perl 5.42 and -MO=Deparse: the line is `&ok(scalar 'Local::Child'->isa('Local::Child'), ...)`. With `use feature "isa"`, or `use v5.35` or later (feature.pm's 5.35 bundle names isa), `$x isa Foo` is the infix operator; under `use v5.34` it's a syntax error.

PerlOnJava unit/indirect_isa_feature_gate.t:10.

## Acceptance criteria

- [ ] `isa` is a word without its feature and an operator with it (`go test ./internal/parse/ -run '^TestIsaWithoutFeature$' -count=1 -v | grep -q '^--- PASS: TestIsaWithoutFeature'`)
