---
title: Derive parse shape from the prototype so a parenless import parses
state: done
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0b84d-caa8-71d2-b42b-0c8445bc6026
created: 2026-09-19T06:15:09.657423509Z
updated: 2026-09-19T06:56:11.023803068Z
sessions:
- start_sha: 7034e553e623c97cdd9406c3d6b9c8ef5c4b2a55
  end_sha: 9cf2ee6df76d20e3fe3e13c8324e0ca79d8137b9
  commits: 1
  started_at: 2026-09-19T06:44:22.901960416Z
  ended_at: 2026-09-19T06:56:11.023803068Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T06:44:22.901960416Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-19T06:56:11.023803068Z
observed_paths:
- internal/parse/call.go
- internal/parse/shape.go
- internal/parse/shape_test.go
---

Derive a call's parse shape from the prototype string, and use it so a
parenless call to an imported sub parses.

Split from 01a0ad3f on 2026-09-19. This is part 3 of 3, and the one the
measurement was about. 01a0ad3f loads the module; 01a0b84d-caa8 reads its
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

- [x] Parse shape is DERIVED from the prototype, never stored (`go test ./internal/parse/ -run '^TestShapeDerivedFromPrototype$' -count=1 -v | grep -q '^--- PASS: TestShapeDerivedFromPrototype'`)
- [x] `(;$)` parses with zero and with one argument (`go test ./internal/parse/ -run '^TestOptionalArgPrototype$' -count=1 -v | grep -q '^--- PASS: TestOptionalArgPrototype'`)
- [x] A leading `&` licenses the block form; a non-initial `&` and `\&` do not. Asserted on the PARSE, never on perl's error text (`go test ./internal/parse/ -run '^TestBlockFormNeedsLeadingAmpersand$' -count=1 -v | grep -q '^--- PASS: TestBlockFormNeedsLeadingAmpersand'`)
- [x] `use Test::More` makes `subtest 'x' => sub {}` parse (`go test ./internal/parse/ -run '^TestImportedListOperator$' -count=1 -v | grep -q '^--- PASS: TestImportedListOperator'`)
- [x] A sub with no prototype swallows a list, and a `($)` one takes a single term (`go test ./internal/parse/ -run '^TestPrototypeArity$' -count=1 -v | grep -q '^--- PASS: TestPrototypeArity'`)
- [x] The T1 rate is re-measured and the change recorded, in BYTES as well as node counts

Spec: ch4 §4.8.3. Blocked by 01a0b84d-caa8 (needs exports and prototypes).


## Done, 2026-09-19 (9cf2ee6d)

Measured over T1 graded, 986 files, with a Test::More whose @EXPORT and
prototypes match the real one:

                         clean   Unknown
    Parse                 460     3,322      no loader
    ParseFile             460     3,322      loader finds nothing
    with Test::More       532     2,531      +72 files, -791 nodes

    full suite            green, 0 FAIL across the repo

All six acceptance criteria verified by running each command as written.

## What `subtest` actually was

Not an unresolved call -- an Unknown SPAN. A bareword followed by a string
never reached the Call path at all, which is why the 691 parenless subtests
in T1 were opaque rather than merely unresolved. With the import resolved it
is a list operator that takes its arguments.

## The committed ratchets did not move, and the middle row is why

T1 is a flat directory of .t files with no .pm beside them, so DirLoader
correctly resolves nothing: `resolved any module` is 0 of 986. That is a fact
about the corpus, not a defect.

TestImportResolutionIsWorthT1Files keeps the third row honest and goes red if
resolution ever stops reducing Unknown nodes. Without it this work reads as
having done nothing, which is exactly the kind of silence the ratchets exist
to prevent.

Repointing the ratchets at ParseFile with a real search path is a separate
decision, and it belongs with 01a0b835 (the oracle repoint) rather than here.

## Two of my assumptions were wrong, and perl said so

Measured against 5.42.0 before encoding, which caught both:

    (@)      swallows the list       sl(1, 2, 3)        NOT one argument
    (\&)     takes exactly one       (&rf((\&foo)), 2)  NOT a list
    (\@\@)   takes two               two(@a, @b)
    (\[$@])  takes one               (g(@a), '???')

A bare `@` or `%` is slurpy; a `\`-escaped one is a single reference slot. My
first counting loop conflated them, and my own test row for `(\&)` was wrong
in the same direction -- corrected against the deparse rather than the other
way round.

## The leading ampersand

    sub bf (&@) { }   bf { $_[0]*2 } (1,2,3)   ->  2,4,6
    sub ni ($&) { }   ni { 1 } (2)             ->  Not enough arguments
                                                   syntax error near "} ("
    sub rf (\&) { }   rf { 1 }                 ->  Type of arg 1 ... must be
                                                   subroutine (not anonymous
                                                   hash ({}))

Both messages of the `($&)` case reproduced, as bson's correction recorded.
The ACs assert the PARSE and never the message text.

## The XS tier still refuses

`knowsShape` is false for a name imported without a visible declaration, so
`first` keeps declining rather than guessing "list operator" -- which would
make `first { $_ > 1 } @a` a syntax error.

An imported sub called with NO arguments now resolves too: `done_testing;`
and `maybe;` are calls whose callee this parser has seen declared, and the
early-return path consulted only the builtin tables before.

## The chain this completes

    01a0ad3f        the loader and the decline path         f066f2ab
    01a0b84d-caa8   exports and prototypes from the text    7034e553
    01a0b84d-ec59   shape derived, and spent at the call    9cf2ee6d

01a0b835 (the oracle repoint) is now unblocked.
