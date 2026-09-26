---
title: A one-dot v-string is three tokens, and use v5.36 depends on that
state: done
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T05:18:56.979600412Z
updated: 2026-09-26T06:48:24.013187741Z
sessions:
- start_sha: a21b05dea5e6ec80fb8f1326a65685e6bbcd4087
  end_sha: 9ba0064cb73a9259867e7248efa2ad215b8a8646
  commits: 2
  started_at: 2026-09-26T06:08:52.015927486Z
  ended_at: 2026-09-26T06:48:24.013187741Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T06:08:52.015927486Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-26T06:48:24.013187741Z
observed_paths:
- conformance/mdtest/vstrings.md
- internal/lexer/ident.go
- internal/lexer/scan.go
- internal/lexer/scanrows_test.go
- internal/lexer/vnum_test.go
---

`v5.36` is a v-string in perl and three tokens to us. The two-dot forms
were fixed by `01a0db78`; the one-dot form was deliberately left, because
correcting it means teaching `internal/parse/use.go` to accept a version
that arrives as ONE token.

Filed by the worker that fixed the two-dot forms, which recorded it in
`conformance/mdtest/vstrings.md` under "STILL OPEN" and said, correctly,
that the corpus is not a tracker.

## Measured, 2026-09-26

    $ perl -e 'my $v = v5.36; print length($v)'
    2

Two characters, so perl holds a string of ordinals. We lex it as:

    Word("v5") Operator(".") Number("36")

The two-dot forms now lex as one `Quote`, which is what `01a0db78`
delivered. `scanVString` declines on one dot ON PURPOSE, and that is the
load-bearing part: `use v5.36` is reassembled from those three tokens by
`internal/parse/use.go`, so making the one-dot form a single token breaks
version parsing until that reassembly is taught the new shape.

## The trap that makes this worth an issue rather than a comment

`01a0db78` hit this exact class in its two-dot work and it nearly shipped:
making `v5.42.0` one token SILENTLY BROKE SIGNATURES. `noteSignatures` read
the version out of the three-token split, and with one token nothing read
it, so under `use v5.42.0` a signature `($a, $b)` started lexing as a
Prototype -- a lexical change that alters what the rest of the file means.

No existing test caught it, because every suite used `v5.36`, the one-dot
form this issue is about. The worker found it by probing, then wrote
`TestVStringVersionStillEnablesSignatures` covering both directions.

So whoever takes this must assume the same shape: the one-dot form is
load-bearing for `use`, and the check that would catch a regression may not
exist yet. Write it first.

## Scope

Make `v5.36` one token and keep `use v5.36` working, including its
feature-gating side effects. At minimum:

  - a failing test that `v5.36` is one token, and
  - a failing test that `use v5.36` still enables signatures, BEFORE the
    lexer change, since that is the one that broke last time.

`versionAtLeast` already strips a leading `v` and stops at the second dot,
so it may need no change; measure rather than assume.

## Context

- paths: `internal/lexer/scan.go` (`scanVString`), `internal/lexer/ident.go` (`noteSignatures`, `versionAtLeast`), `internal/parse/use.go`, `conformance/mdtest/vstrings.md`
- commands: `go test ./internal/lexer/ -count=1`, `go test ./internal/parse/ -count=1`
- entrypoints: `scanVString`, which currently declines on one dot

## Acceptance criteria

- [ ] A one-dot v-string is one token (`go test ./internal/lexer/ -run '^TestOneDotVString$' -count=1 -v | grep -q '^--- PASS: TestOneDotVString'`)
- [ ] `use v5.36` still enables signatures (`go test ./internal/lexer/ -run '^TestVStringVersionStillEnablesSignatures$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] The corpus no longer records this as open (`! grep -qi 'STILL OPEN' conformance/mdtest/vstrings.md`)
- [ ] REGRESSION GUARD, passes today: the lexer suite holds (`go test ./internal/lexer/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
