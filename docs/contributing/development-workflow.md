---
stability: 1
covers: []
---

## Setup

Install Go 1.24+, git, and clone the repository. Run `go build -o git-zhi ./cmd/git-zhi/`
to verify the build.

## Branching

Create feature branches from `pu`. Merge back to `pu` via pull request.

## Fuzzing

`-fuzz` takes one target and runs until its budget is spent, so the fuzzers
are make targets of their own, not part of `make test`:

- `make fuzz-lexer` runs `FuzzLexer` (internal/lexer) to 1,000,000 execs.
- `make fuzz-parser` runs `FuzzParser` (internal/parse) to 1,000,000 execs.
- `make fuzz-incremental` runs `FuzzIncremental` and then
  `FuzzIncrementalSequence` (internal/parse) to 10,000,000 execs each. They
  check spec §7.7's property: an incremental re-parse after one edit, or after
  each edit of a burst, equals a fresh Session parse of the edited bytes.

`make test` replays every committed seed under each package's
`testdata/fuzz/`, the eight §7.7.3 edits among them, but never runs a
campaign. A find is a seed to commit with its fix.

## Review

All changes require a pull request. Tests must pass. No `--no-verify`.

## Deploy

Build the binaries with `go build` and place them on `$PATH`.
