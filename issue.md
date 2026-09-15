---
title: Fix psc check CLI to produce diagnostics
state: pending
urgency: normal
milestone: psc-type-alignment
created: 2026-03-22T02:43:24.620376-04:00
updated: 2026-09-15T21:02:04.542168956Z
---

The `psc check` CLI exits 0 with no output on files containing known type errors.
The Go-level E2E tests (`TestPSCCheckDiagnostics`) pass — they invoke `psc.NewCommand()`
directly and get arity-mismatch and type-mismatch diagnostics on stderr.

The CLI help shows flags (--recursive, --format, --strict, --dump-ast, --show-inferred)
that do not exist in `internal/psc/check_command.go`. The router or help system is
intercepting the check subcommand and routing it through a different code path.

## Acceptance

- [x] `./build/psc check test/e2e/testdata/check/diagnostics.pl` produces diagnostic output
- [x] Exit code is non-zero when diagnostics are found
- [x] CLI output matches Go API test expectations
- [x] Root cause of CLI/API disconnect is documented

## Verified fixed, 2026-09-15

Re-tested against the tree at e2f29afa. Every acceptance criterion now holds,
fixed by work between the March filing and now rather than by this pass.

    $ ./psc check test/e2e/testdata/check/diagnostics.pl
    ...:1:1: error: call to push: expected at least 2 argument(s), got 0 [arity-mismatch]
    ...:4:6: error: call to push: argument 1 expects Array, got Int [type-mismatch]
    ...:9:6: error: call to keys: argument 1 expects Array|Hash, got Int [type-mismatch]
    ...:10:6: error: call to push: argument 1 expects Array, got Hash [type-mismatch]
    ...:11:8: error: call to length: argument 1 expects Str, got Array [type-mismatch]

    exit code with diagnostics: 1
    exit code on clean.pl:      0
    go test ./test/e2e/ -run TestPSCCheck: ok

The phantom flags are gone: `--recursive`, `--format`, `--dump-ast` and
`--show-inferred` are each rejected with "Unknown flag".

`--strict` survives because it is REAL, not phantom. It is declared at
check_command.go:75, bound at :19, and passed into infer.Options at :35,
where infer.go:107 swaps TypeSatisfies for TypeSatisfiesStrict. Measured
through the CLI:

    $ ./psc check strict3.pl                 # my $n = Foo->compute + 1;
    (silent)
    $ ./psc check --strict strict3.pl
    strict3.pl:1:9: error: left operand of "+": expected Num, got Unknown [coercion-mismatch]

Two corrections to my own earlier reading of this issue, recorded because
both were wrong in ways that would have wasted the next person's time:

1. I reported `psc check` still silently exiting 0 on a type error. My test
   file was `my @a = (1,2,3); my $n = keys(@a) + 1;` — and `keys(@a)` is
   LEGAL Perl (verified: it returns the index list, so the expression is 4).
   PSC accepts it correctly. The fixture this issue names was always
   diagnosed.

2. I reported `--strict` as "accepted but produces nothing". It produces
   nothing on files where nothing is un-inferred, which was every file I
   tried until I read infer_test.go:2486 and used the shape it documents.

One real gap remains, already recorded in that test's own comment rather than
here: `my $r = Foo->compute` gives $r the sigil type Scalar rather than the
RHS's Unknown, so strict does not see through an assignment. Closing it means
propagating Unknown through assignment narrowing. That is its own issue, not
this one.
