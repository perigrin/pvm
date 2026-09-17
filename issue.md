---
title: "Module stubs: the prototype string, generated from perl"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-17T02:44:01.961043269Z
updated: 2026-09-17T02:49:19.349366446Z
---

Teach the parser what a `use` imported, so a parenless call to an imported sub
parses.

Measured in T1: 889 of 986 files `use Test::More`, and `subtest` appears 691
times -- **always** parenless, never `subtest(...)`. Paren forms already
parse; what fails is knowing a bareword is a list operator.

perl agrees it cannot be done without the import:

    $ perl -e 'subtest "name" => sub { 1 };'
    String found where operator expected (Do you need to predeclare "subtest"?)

## One field: the prototype string

Not a shape enum. The prototype DETERMINES the shape, measured:

    sub np  { }        no prototype   np(1,2,3) swallows the list
    sub one ($) { }    ($)            `one 1, 2` parses as one(1), 2
    sub nil () { }     ()             `nil + 1`  parses as nil() + 1

So shape is derived, not stored. A stored `shape` field would be a second
copy of one fact -- the failure mode this project keeps finding.

It also does not FIT, in at least two ways.

`pass` and `fail` are `(;$)`, zero-or-one:

    $ perl -e 'sub p (;$) {...} p; p("x");'    both parse

A LEADING `&` licenses the block-call form, and without one that form is not
merely different but a SYNTAX ERROR:

    sub bf (&@) { }    bf { $_[0]*2 } (1,2,3)   ->  2,4,6
    sub ao (&)  { }    ao { 42 }                ->  42
    sub na ($@) { }    na { 1 } (2,3)           ->  syntax error near "} ("

Measured, and the LEADING position is what matters. A non-initial `&` is an
ordinary code-ref slot that does not license the block form:

    sub ni ($&) { }   ni { 1 } (2)
      Not enough arguments for main::ni ... near "} ("
      syntax error ... near "} ("

    sub rf (\&) { }   rf { 1 }
      Type of arg 1 to main::rf must be subroutine (not anonymous hash ({}))

The mechanism is that `{1}` is read as an anonymous hash rather than a block.
The `($&)` case emits BOTH messages; an earlier draft of this issue claimed
only the arity one, which was an artifact of truncating perl's output to one
line. Corrected by bson, reproduced here.

The ACs assert the POSITIVE property -- that the block form is unavailable --
rather than which message perl emits. Message text and ordering are not
promised, and an assertion on them would fail a future perl for no semantic
reason.

So a pattern match anchors on a leading `&` NOT preceded by a backslash.

A list/unary/niladic enum has a slot for neither fact and would have to
guess at both. The string does not.

A possible fourth, UNVERIFIED and flagged as such by bson who raised it:
`\@` and `\%` may be auto-reference slots -- the caller writes `@a` and the
sub receives `\@a` -- which would change what is legal at the call site
without changing arity. A lead to measure, not a confirmed fact. Three are
measured; this one is not.

And it is what M4 needs anyway: the prototype is the srefgen signal the whole
fidelity harness was built to measure.

## Generated, not authored

    perl -MTest::More -e 'print prototype(\&Test::More::ok)'     ->  $;$

perl's own symbol table is the source of truth, the same authority the oracle
already uses. perl-lsp hardcodes a Rust array
(`perl-workspace-core/src/test.rs`); a hand list drifts, and there is prior
art in this project of exactly that going wrong.

Scrape `@EXPORT`/`@EXPORT_OK` rather than what one `import` call installed --
bson's point: `use Test::More` takes import args, so a single invocation is
not the defensible source.

## What this is NOT

A heuristic keyed on the `use` statement, not import resolution. Recorded in
the generated file's header so the T1 rate does not start claiming more than
it measures:

- `use Test::More qw(ok)` imports a subset; we assume the default list
- a local sub shadowing an import stays `Call{Resolved:false}`, correctly
- `require` + manual import is not covered

## Consulted, and why this is not the type-library syntax

Two sessions, independently:

**bson** (perl5-son): the `sub :infix + (Num $x, Num $y) Num;` syntax is
specified in a plan doc only, unimplemented and unparsed -- perl 5.42 rejects
return-type position, typed params, and bodyless signatures, so a declaration
file cannot be a perl file. What exists is `B::SoN::TypeLibrary` with a
programmatic API and rows of `{operands, result}`: types only, no prototype
field, no loader, nothing on disk. Its recommendation was to build this
separately and keep the prototype string as the single field.

**commonplacebook**: searched the vault and found nothing; the spec went to a
scratchpad that has since been cleared, so the artifact is with perigrin. Its
independent point: "a prototype controls parse shape; a type controls which
values are legal -- conflating them would be a design error." Parse shape is
a property of the grammar, not the type system, so it plausibly has a
different lifecycle and source of truth.

If the paper syntax ever gets a parser, a prototype column beside
`operands`/`result` is a plausible convergence. Not blocking on it; a
prototype string is trivially liftable.

## Acceptance

- [ ] A generator produces a stub file from perl's symbol table (`make module-stubs`)
- [ ] The stub records name and prototype string, with an explicit null for none (`go test ./internal/parse/ -run '^TestStubRecordsPrototypes$' -count=1 -v | grep -q '^--- PASS: TestStubRecordsPrototypes'`)
- [ ] Parse shape is DERIVED from the prototype, not stored (`go test ./internal/parse/ -run '^TestShapeDerivedFromPrototype$' -count=1 -v | grep -q '^--- PASS: TestShapeDerivedFromPrototype'`)
- [ ] `(;$)` parses with zero and with one argument (`go test ./internal/parse/ -run '^TestOptionalArgPrototype$' -count=1 -v | grep -q '^--- PASS: TestOptionalArgPrototype'`)
- [ ] `use Test::More` makes `subtest 'x' => sub {}` parse (`go test ./internal/parse/ -run '^TestImportedListOperator$' -count=1 -v | grep -q '^--- PASS: TestImportedListOperator'`)
- [ ] A leading `&` licenses the block form; a non-initial `&` and `\&` do not. Asserted on the PARSE, never on perl's error text (`go test ./internal/parse/ -run '^TestBlockFormNeedsLeadingAmpersand$' -count=1 -v | grep -q '^--- PASS: TestBlockFormNeedsLeadingAmpersand'`)
- [ ] The stub file's header records that this is a heuristic, not import resolution
- [ ] The T1 rate is re-measured and the change recorded

Spec: ch4 §4.8.3. Blocked on perigrin confirming the format is ours to define.
