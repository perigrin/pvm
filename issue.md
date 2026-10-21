---
title: "Method-parameter declarations to re-measure before method checking lands: IO::Handle read/sysread, the *_separator invocants, Data::Dumper::Dump, IO::Socket::IP accept"
state: pending
urgency: normal
milestone: v0.1
blocked_by:
- 01a15ce6-5e2f-7794-8572-2a7bd2852304
created: 2026-10-21T21:06:00.725628471Z
updated: 2026-10-21T21:06:00.725628471Z
---

crochet:review round 3 V18; destination given by 01a16016-481d. Round 2's S18 is tracked in 01a15ce6-5e2f (method-argument checks), whose body does not list the declarations below.

`methodCall` (internal/infer/calls.go) checks no arguments, so these method-parameter declarations are held only by the return/parameter oracle. Each turns into a false positive when method checking lands, and must be re-measured against perl 5.42 first:

- IO::Handle `read` and `sysread`: `Str $buf`.
- The `*_separator` entries declared `(Str $class:)`.
- `Data::Dumper::Dump`.
- IO::Socket::IP `accept`.

## Steps
- [ ] Re-measure each declaration above with perl 5.42 -w, called the ways its POD shows
- [ ] Correct each that perl accepts and the declaration refuses

## Acceptance criteria

### Positive Scenarios
- [ ] Method-parameter declarations fit perl's calls (`go test ./internal/infer/ -run '^TestMethodParameterDeclarationsFitPerl$' -count=1 -v | grep -q '^--- PASS: TestMethodParameterDeclarationsFitPerl'`)

### Negative Scenarios
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
