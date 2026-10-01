---
title: The builtin table is parsed from declarations/CORE.pmt
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-01T15:14:51.218241251Z
updated: 2026-10-01T15:14:53.546333211Z
sessions:
- start_sha: a3095b73ecc19394831f35b4980b7ed48bb1d9be
  end_sha: ""
  commits: 0
  started_at: 2026-10-01T15:14:53.546333211Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-01T15:14:53.546333211Z
---

The builtin table is the type library for the language itself, TypeScript's lib.d.ts. perigrin (2026-10-01): it is a file of declarations, `sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__) Object;`, parsed to build the default table. Today that table is coreprotos.go, a Go map generated from perl.

declarations/CORE.pmt holds the 188 keywords with a prototype, as `sub NAME :prototype(PROTO);` (types come later, on these lines), and coreTable parses it. coreprotos.go goes. `:prototype(...)` is read as a prototype wherever a sub is declared; it is the spelling a sub with a signature needs. That fixes two perl.git files: op/signatures.t's `sub t020 :prototype($)` and op/taint.t's `sub tainted :prototype($)` are named unaries, and perl 5.42 reads `t020 222, $a = 333` as `t020(222), $a = 333`.

## Acceptance criteria

- [ ] CORE.pmt agrees with perl for every keyword (`go test ./internal/parse/ -run '^TestCoreDeclarationsArePerls$' -count=1 -v | grep -q '^--- PASS: TestCoreDeclarationsArePerls'`)
- [ ] A :prototype attribute is a prototype (`go test ./internal/parse/ -run '^TestPrototypeAttribute$' -count=1 -v | grep -q '^--- PASS: TestPrototypeAttribute'`)
- [ ] coreprotos.go is gone (`sh -c '! test -e internal/parse/coreprotos.go'`)
