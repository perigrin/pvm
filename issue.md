---
title: FuzzParser has no fixpoint invariant, so a canon defect the fuzzer reaches breaks no assertion
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-26T13:59:01.639917324Z
updated: 2026-10-09T06:12:51.43978509Z
sessions:
- start_sha: 0a4028652b10372f819872a4a07866fb7eb15151
  end_sha: 0cbc6dfb48de6e5fb1f7641ecefb081702ab84f9
  commits: 3
  started_at: 2026-10-09T05:31:57.443390482Z
  ended_at: 2026-10-09T06:12:51.43978509Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T05:31:57.443390482Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T06:12:51.43978509Z
observed_paths:
- docs/specs/perl-parser/00-findings.md
- internal/parse/fuzz_test.go
- internal/parse/sites.go
- internal/parse/sites_decided_test.go
- internal/parseoracle/compare_facts.go
- internal/parseoracle/cvwalk_test.go
- internal/parseoracle/ratchet_test.go
- internal/parseoracle/span_test.go
- internal/parseoracle/testdata/parse_facts.pl
- internal/parseoracle/testdata/ratchet/baseline.txt
---

`checkParseInvariants` in `internal/parse/fuzz_test.go` asserts three things --
never panic, never lose a byte, never build an empty node -- and none of them
involves Canon. So a mutation that produces a shape whose canon does not
re-parse finds nothing to fail.

This is not hypothetical. `01a0dde4` was exactly that shape: canon of
`map { $_ + 1 } @a` emitted `map({$_ + 1;}@a)`, which parsed to Unknown=1 and
re-emitted as `map({$_ + 1});}@a);`. The fuzzer's three invariants all held on
it -- no panic, no lost byte, no empty node -- and the defect was found by a
worker measuring a different issue by hand.

## What to add

The property is the one `canonMatches` (`internal/parse/canon_test.go`) already
computes: `canon(parse(canon(parse(S)))) == canon(parse(S))`. Adding it as a
fourth invariant makes the fuzzer able to find this class.

Two things to decide while doing it, neither settled here:

- Whether the emission must also PARSE (Unknown==0), or only be stable. They
  are different assertions: an emission that fails to parse can still be a
  fixpoint, which is the distinction `canonMatches`'s own comment draws. On
  fuzz input, Unknown==0 is certainly too strong -- the seeds include `}` and
  `\x00\x01\x02` -- so a fixpoint-only invariant is the likely answer, and the
  parse check stays in the targeted test.
- Whether the invariant can hold on non-Perl at all. Half-typed and
  non-Perl input is the fuzzer's whole point, and if canon is not a fixpoint on
  garbage today, this lands as a ratchet or a seeded test rather than a
  universal invariant. MEASURE before choosing: run the property over the
  existing seed corpus first and count.

## Context

- paths: `internal/parse/fuzz_test.go` (`checkParseInvariants`),
  `internal/parse/canon_test.go` (`canonMatches`, the fixpoint already written)
- commands: `make fuzz-parser`, `go test ./internal/parse/ -run '^TestFuzzSeeds$' -count=1`
- found by: agent:m1f-1 while fixing 01a0dde4, which asked whether this
  belonged in that change. It does not: that fix is a lexer and parser change
  with a targeted test, and widening a fuzz invariant is a separate question
  with its own measurement.

## Acceptance criteria

- [ ] the canon fixpoint holds over the seed corpus, measured before it became an invariant (`go test ./internal/parse/ -run '^TestFuzzSeedsCanonFixpoint$' -count=1 -v | grep -q '^--- PASS: TestFuzzSeedsCanonFixpoint'`)
- [ ] The seed corpus passes whatever form is chosen (`go test ./internal/parse/ -run '^TestFuzzSeeds$' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
