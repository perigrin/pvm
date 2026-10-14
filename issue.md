---
title: A reference as a repeat count is silent by default, but perl repeats by its address and dies
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-14T21:25:54.564466402Z
updated: 2026-10-14T21:25:54.564466402Z
---

## Context

- paths: `internal/infer/lowered_ops.go` (`lossyCoercion`, repeat count check)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

Found 2026-10-14 by the operator false-positive fixes (01a13b5c-0164): under perigrin's ruling that a reference in a numeric slot is a lossy-coercion opt-in, `"ab" x {}` is silent by default. perl -w gives no warning either, but it repeats by the reference's address and dies "Out of memory in perl:util:safesysrealloc". That is a crash, not a lossy value. `TestRepeatCountAcceptsNum` holds it only under the opt-in today.

Question for perigrin: keep a reference as a repeat count (and any other slot where the numified address is used as a size or count) as a default diagnostic, as an exception to the lossy class?

## Steps

- [ ] perigrin rules; if kept, failing test first; a Ref/Object repeat count is reported by default; `make test`; commit

## Acceptance criteria

- [ ] a reference as a repeat count is reported by default (`go test ./internal/infer/ -run '^TestRefRepeatCountIsReported$' -count=1 -v | grep -q '^--- PASS: TestRefRepeatCountIsReported'`)
