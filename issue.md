---
title: "A $var before a term-start after print is the filehandle to perl: print $0 &y reads as an infix &"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-10T01:40:32.06078993Z
updated: 2026-10-10T05:34:23.31104023Z
sessions:
- start_sha: 528c1fa24cbd7732a2aa37ec1c0bbfee0828a011
  end_sha: baf52e02b9bfe6cccdacae5aafb9bd5e8d725c68
  commits: 3
  started_at: 2026-10-10T01:47:04.390792455Z
  ended_at: 2026-10-10T03:33:35.252356401Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-10T01:47:04.390792455Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-10T03:33:35.252356401Z
observed_paths:
- internal/lexer/fhslot_test.go
- internal/lexer/lexer.go
- internal/parse/call.go
- internal/parse/canon.go
- internal/parse/chain.go
- internal/parse/chain_test.go
- internal/parse/parse.go
- internal/parse/scalarhandle_test.go
- internal/parse/testdata/fuzz/FuzzParser/0c172203842b1f8c
- internal/parse/testdata/fuzz/FuzzParser/7bd29f70ee42a374
- internal/parse/testdata/fuzz/FuzzParser/b45427e536ce28a7
- internal/parse/testdata/t1fidelity.ratchet
- internal/parse/trailingcomma_test.go
- internal/parse/unaryplus_test.go
- internal/parseoracle/subject_test.go
---

## Context

- paths: `internal/parse/call.go` (`parseFilehandleSlot`), `internal/lexer` (perl decides part of this in toke.c), `internal/parse/testdata/fuzz/FuzzParser/`
- commands: `make fuzz-parser`; `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-10 by `make fuzz-parser` (after 01a12344). Seed parked at scratchpad/pending-seeds/0c172203842b1f8c: `print$0 &y` -- perl reads `$0` as print's filehandle and `&y` as a call (Deparse: `print $0 &y;`); internal/parse reads `print($0 & y())`, and canon's `y()` re-lexes as a transliteration. This is perl's toke.c rule for a `$var` followed by whitespace after a list operator (print/printf/say/exec/system/sort): the variable is the filehandle when what follows looks like the start of a term -- `&`, `*`, `<`, `%`, `'` before an identifier, `.3`, `-1`/`+1`, `?x`, `/re/`, `<<HEREDOC`. Measured: `print $0 -1;` Deparses as `print $0 -1;` (handle).

## Steps

- [ ] Read toke.c's rule (S_intuit_more / the list-operator `$var` case) and measure each trigger with Deparse; failing test first; implement in the parser and/or lexer as the rule requires; commit the seed; ratchets toward perl only; `make fuzz-parser` until one full run passes; `make test`; commit

### The fuzz criterion replays the seeds (2026-10-10)

`make fuzz-parser` explores at random, so as a criterion it can fail on a
new find unrelated to this issue -- and a failing run writes its seed into
testdata, where every later `go test ./internal/parse/` replays it. This
issue's claim is its own inputs, which are committed as fuzz seeds; the
criterion replays them deterministically. The open-ended fuzz claim
belongs to 01a0a70f-0ec3 and 01a0de03, where it stays.

## Acceptance criteria

- [ ] a `$var` before a term-start after a list operator is the filehandle (`go test ./internal/parse/ -run '^TestScalarHandleBeforeATermStart$' -count=1 -v | grep -q '^--- PASS: TestScalarHandleBeforeATermStart'`)
- [ ] every committed fuzz seed replays clean (`go test ./internal/parse/ -run '^FuzzParser$' -count=1 -v | grep -q '^--- PASS: FuzzParser'`)
