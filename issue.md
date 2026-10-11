---
title: continue BLOCK is declined, and every op it needs is already claimed
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T17:48:11.618449654Z
updated: 2026-10-11T16:22:47.407914302Z
sessions:
- start_sha: 20c20c3c9f6b7772b719bd7affc1d50f419e1e75
  end_sha: b2084901525559c23fe0eef21f146e164b5c4cc9
  commits: 2
  started_at: 2026-09-26T18:36:10.823402761Z
  ended_at: 2026-09-26T19:45:19.698155701Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T18:36:10.823402761Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-26T19:45:19.698155701Z
observed_paths:
- conformance/mdtest/loops.md
- conformance/mdtest/names.md
- internal/conformance/testdata/corpus.ratchet
- internal/lexer/caret_test.go
- internal/lexer/scan.go
- internal/lexer/testdata/lex.t.tokens
- internal/parse/control.go
- internal/parse/control_test.go
- internal/parse/parse.go
- internal/parse/t2_test.go
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1fidelity.ratchet
---

`continue BLOCK` after a `while` is declined: `continue` is still in
`statementKeywords` at `internal/parse/parse.go:800`. Valid perl, present in
perl.git `t/cmd/`, and absent from every corpus perl block.

Contributed by the B::SoN session, which round-trips it today -- so this is a
case where their implementation is ahead of ours.

## Measured

    my $i = 0;
    my @seen;
    while ($i < 3) {
      push @seen, $i;
    } continue {
      $i++;
    }
    print "@seen\n";

    perl 5.42.0   ->  0 1 2
    ours          ->  Unknown=1, canon keeps the block verbatim:
                      "... while ($i < 3) {push(@seen , $i);} continue {\n  $i++;\n}print(...)"

Canon's fixpoint HOLDS, so this is a plain decline rather than a wrong tree --
`continue` sits in `statementKeywords`, whose docstring says the list is
shortened as each form becomes readable.

## Every op is already claimed, so no tier change is needed

From `perl -MO=Concise,-exec`:

    and const enter enterloop gvsv join leave leaveloop lt multiconcat
    nextstate padav padsv padsv_store preinc print push pushmark unstack

Checked against the tiers: `enterloop` and `unstack` at 06_control, `leaveloop`
at 07_subroutines, `preinc` at unary, `padav`/`push` at 02_variables. Nothing
unclaimed.

That is the contrast with `given`/`when` (`01a0dec7`), which needs FIVE new ops
claimed before the corpus can hold a case at all. `continue` needs none: there
is no `continue` op. The destination is metadata on `enterloop`
(`enterloop(next->v last->12 redo->d)`) plus an `unstack` at the bottom, which
is why grepping the introduces lines for `continue` finds nothing and nothing is
missing.

## The corpus does not have it

Measured over `perl` BLOCK contents, not file text -- a word in prose does not
count:

    continue {        0 block lines in conformance/mdtest/

## Scope

1. Read `continue BLOCK` as part of the loop it follows. `while`, `until` and
   `foreach` all take one; measure whether they share a path.
2. Remove `continue` from `statementKeywords` and update that docstring, which
   lists what has already gone -- it is the measure of progress through the
   milestone and must not go stale.
3. Canon must re-emit the block. The fixpoint holds today only because the
   statement is Unknown and echoes its bytes; once it parses, canon has to
   place it. Run `go test ./internal/parse/ -run '^TestCanon' -count=1` EARLY.
4. A corpus case, tier 06. `continue` runs on `next` but NOT on `last`, which
   is the property worth asserting rather than the bare block.

## Context

- paths: `internal/parse/parse.go:800` (`statementKeywords`),
  `internal/parse/control.go` (the loop forms), `internal/parse/canon.go`
- commands: `go test ./internal/parse/ -count=1`,
  `go test ./internal/conformance/ -count=1`
- files: `/home/perigrin/dev/perl5/t/cmd/*.t`
- related: `01a0dec7` (the switch family needs new ops; this does not),
  `01a0d087` (ungated `defer`, the last `statementKeywords` entry removed)

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

### Criteria repaired (2026-10-11)

The milestone verify of 2026-10-11 could not run some of these criteria and
failed others on stale text. Commands written as bare backtick spans are now in
`(...)`, test names are anchored with a `--- PASS` check, a whole-suite run is
removed (perigrin: "Why are we doing 73 full test suite runs?!?"), and a grep of
text that later work moved is pointed at where the fact now lives.

## Acceptance criteria

- [ ] `continue BLOCK` after a while loop parses (`go test ./internal/parse/ -run '^TestContinueBlock$' -count=1 -v | grep -q '^--- PASS: TestContinueBlock'`)
- [ ] `continue` is gone from statementKeywords and the docstring says so (`! awk '/^var statementKeywords = map/,/^}/' internal/parse/parse.go | grep -qE '^[[:space:]]*"continue"[[:space:]]*:' && grep -q 'continue. is GONE' internal/parse/parse.go`)
- [ ] Canon re-emits the block and holds its fixpoint (`go test ./internal/parse/ -run '^TestCanon' -count=1 -v | grep -q '^--- PASS: TestCanon'`)
