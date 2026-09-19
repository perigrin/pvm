---
title: Derive parse shape from the prototype so a parenless import parses
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0b84d-caa8-71d2-b42b-0c8445bc6026
created: 2026-09-19T06:15:09.657423509Z
updated: 2026-09-19T06:15:09.657423509Z
---

Derive a call's parse shape from the prototype string, and use it so a
parenless call to an imported sub parses.

Split from 01a0ad3f on 2026-09-19. This is part 3 of 3, and the one the
measurement was about. 01a0ad3f loads the module; 01a0b8xx-A reads its
exports and prototypes; this issue spends them.

## The measurement

In T1: 889 of 986 files `use Test::More`, and `subtest` appears 691 times --
**always** parenless, never `subtest(...)`. Paren forms already parse; what
fails is knowing a bareword is a list operator.

perl agrees it cannot be done without the import:

    $ perl -e 'subtest "name" => sub { 1 };'
    String found where operator expected (Do you need to predeclare "subtest"?)

## Shape is derived, never stored

The prototype determines the shape, measured:

    sub np  { }        no prototype   np(1,2,3) swallows the list
    sub one ($) { }    ($)            `one 1, 2` parses as one(1), 2
    sub nil () { }     ()             `nil + 1`  parses as nil() + 1

A stored `shape` enum would be a second copy of one fact -- the failure mode
this project keeps finding -- and it does not FIT, in at least two ways.

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
The `($&)` case emits BOTH messages; an earlier draft claimed only the arity
one, which was an artifact of truncating perl's output to one line. Corrected
by bson, reproduced here.

The ACs assert the POSITIVE property -- that the block form is unavailable --
rather than which message perl emits. Message text and ordering are not
promised, and an assertion on them would fail a future perl for no semantic
reason.

So a pattern match anchors on a leading `&` NOT preceded by a backslash.

A list/unary/niladic enum has a slot for neither fact and would have to guess
at both. The string does not.

A possible fourth, UNVERIFIED and flagged as such by bson who raised it:
`\@` and `\%` may be auto-reference slots -- the caller writes `@a` and the
sub receives `\@a` -- which would change what is legal at the call site
without changing arity. A lead to measure, not a confirmed fact. Three are
measured; this one is not.

And it is what M4 needs anyway: the prototype is the srefgen signal the whole
fidelity harness was built to measure.

## Why this is not the type-library syntax

Two sessions, independently:

**bson** (perl5-son): the `sub :infix + (Num $x, Num $y) Num;` syntax is
specified in a plan doc only, unimplemented and unparsed -- perl 5.42 rejects
return-type position, typed params, and bodyless signatures, so a declaration
file cannot be a perl file. What exists is `B::SoN::TypeLibrary` with a
programmatic API and rows of `{operands, result}`: types only, no prototype
field, no loader, nothing on disk. Its recommendation was to build this
separately and keep the prototype string as the single field.

**commonplacebook**: searched the vault and found nothing; the spec went to a
scratchpad since cleared, so the artifact is with perigrin. Its independent
point: "a prototype controls parse shape; a type controls which values are
legal -- conflating them would be a design error." Parse shape is a property
of the grammar, not the type system, so it plausibly has a different
lifecycle and source of truth.

If the paper syntax ever gets a parser, a prototype column beside
`operands`/`result` is a plausible convergence. Not blocking on it; a
prototype string is trivially liftable.

## Acceptance

- [ ] Parse shape is DERIVED from the prototype, never stored (`go test ./internal/parse/ -run '^TestShapeDerivedFromPrototype$' -count=1 -v | grep -q '^--- PASS: TestShapeDerivedFromPrototype'`)
- [ ] `(;$)` parses with zero and with one argument (`go test ./internal/parse/ -run '^TestOptionalArgPrototype$' -count=1 -v | grep -q '^--- PASS: TestOptionalArgPrototype'`)
- [ ] A leading `&` licenses the block form; a non-initial `&` and `\&` do not. Asserted on the PARSE, never on perl's error text (`go test ./internal/parse/ -run '^TestBlockFormNeedsLeadingAmpersand$' -count=1 -v | grep -q '^--- PASS: TestBlockFormNeedsLeadingAmpersand'`)
- [ ] `use Test::More` makes `subtest 'x' => sub {}` parse (`go test ./internal/parse/ -run '^TestImportedListOperator$' -count=1 -v | grep -q '^--- PASS: TestImportedListOperator'`)
- [ ] A sub with no prototype swallows a list, and a `($)` one takes a single term (`go test ./internal/parse/ -run '^TestPrototypeArity$' -count=1 -v | grep -q '^--- PASS: TestPrototypeArity'`)
- [ ] The T1 rate is re-measured and the change recorded, in BYTES as well as node counts

Spec: ch4 §4.8.3. Blocked by 01a0b8xx-A (needs exports and prototypes).
