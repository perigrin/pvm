---
title: "split has no prototype but typed positional parameters: typed Perl has no spelling for it"
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-06T12:42:47.53409794Z
updated: 2026-10-06T12:42:47.53409794Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/corenoproto_test.go` (`TestCoreSplitDeclared`, skipped), RFC 0001 ("CORE.pmt: the builtin table", "Builtins that keep their own parse")
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

From 01a0fd9b-5017-78d0-ae38-331cbe6a9866, which left `split` undeclared because perl contradicts every available spelling. Measured on 5.42: `prototype("CORE::split")` is undef and `split $a, $b` deparses as a list operator, `split(/$a/, $b, 0)`; yet its arguments are positional and typed -- the string is taken in SCALAR context (`split /,/, @a` over two elements yields from "2"), and a fourth argument is "Too many arguments for split".

- The typed line `(Regex|Str $pattern = ' ', Str $string = $_, Int $limit = 0) List` derives `;$_$`, a prototype perl does not report: TestCoreDeclarationsArePerls fails and split's parse would change.
- The RFC's form for a sub with no prototype, `(List @args)`, says the string is flattened, which perl shows is false.
- `:unary` (named unary, derives no prototype) is the wrong parse shape; an invocant colon and a multi do not apply.

So typed Perl needs a way to say "these parameters are typed and positional, but the declaration derives no prototype and parses as a list operator" -- perhaps an attribute in the family of `:unary` (e.g. `:listop`), or a rule that a builtin in "Builtins that keep their own parse" derives none.

## Steps

- [ ] perigrin rules on the spelling; record it in RFC 0001
- [ ] Implement with a failing test; declare split; move its golden row into TestCoreNoPrototypeBuiltinsMatchMeasuredSignatures and unskip TestCoreSplitDeclared; `make test`; commit

## Acceptance criteria

- [ ] split is declared with its measured types and no derived prototype (`go test ./internal/parse/ -run '^TestCoreSplitDeclared$' -count=1 -v | grep -q '^--- PASS: TestCoreSplitDeclared'`)
