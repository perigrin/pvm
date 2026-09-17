---
title: Resolve imports by parsing the module's own source
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-17T02:44:01.961043269Z
updated: 2026-09-17T02:55:34.44568822Z
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

## Parse the module's source -- do NOT run perl

To handle `use Test::More`, find `Test/More.pm` on the search path and PARSE
it, with the parser this milestone is building. Recursive, and it terminates.

An earlier draft of this issue said to generate a stub file from perl's own
symbol table (`perl -MTest::More -e 'print prototype(...)'`). Two sessions
endorsed that and both were wrong. perigrin's argument is decisive and it is
the reason this project exists:

**reading the symbol table means libperl has parsed the text.** Moving that
from parse time to build time does not remove the dependency, it hides it.
A Perl parser that must run Perl to parse Perl is circular.

It fails in three further ways beyond the circularity:

- a repo whose dependencies are not installed cannot be analysed at all
- the data pins to whichever module version was on the generating machine
- local modules, which are never installed anywhere, get no coverage

Everything needed is in the text. Measured on Test/More.pm, 5.42.0:

    our @EXPORT = qw(ok use_ok ... pass fail ... subtest ...)   a literal qw
    sub ok ($;$) {        prototype on the sub line
    sub is ($$;$) {
    sub pass (;$) {
    sub subtest {         no prototype, visibly absent

    39  static `sub NAME` declarations
     0  glob-assignment generations (`*{...} = sub`)

This is the same answer the Chalk side reached for C headers, from the other
direction: for C you parse the header, for Perl you parse the module. Both
are static reads of declaration syntax, and Perl is the easier case because
the file is already in a language this parser handles -- no second grammar.

## The quarantine list

Statically visible in the ordinary case, opaque only where something computes
at runtime. Decide what the parser does on each NOW rather than retrofitting:

- subs built dynamically: `*{"main::$name"} = sub {...}`. Test::More does
  none, which is why it works here.
- `@EXPORT_OK` and `%EXPORT_TAGS`, usually literal so usually fine
- an `import` override that COMPUTES exports rather than delegating to
  Exporter

Same shape as the tie and magic quarantine. Where a module hits one of these,
its imports stay unknown and calls to them stay `Call{Resolved:false}` --
which is correct, not a gap.

## What this is NOT

Not a symbol table, and not a generated snapshot. Reading a module's source
IS import resolution for the static cases, which is why the limits are
narrower than the earlier draft claimed:

- `use Test::More qw(ok)` imports a subset, and the import list is in the
  `use` statement, so this IS resolvable rather than assumed
- a local sub shadowing an import is visible in the file being parsed
- `require` + manual import is not covered
- anything on the quarantine list above stays unresolved

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

## Where the text runs out: XS

A module whose subs are implemented in C has names in the text and bodies in
a shared library. Measured on List/Util.pm, 5.42.0:

    4   static `sub NAME` declarations
    2   XSLoader/bootstrap calls
    our @EXPORT_OK = qw(all any first min max ... reduce ...)   literal

So the EXPORT list is readable and the prototypes are not. At runtime perl
reports `first &@`, `reduce &@`, `sum @`, `max @` -- and `&@` is the leading
ampersand that licenses the block form. None of that is in the file.

That gives three tiers, and they need different answers:

    name known, prototype known      Test::More      resolve fully
    name known, prototype unknown    List::Util      REFUSE, do not guess
    name unknown                     computed import Call{Resolved:false}

The middle tier is the new one and refusing is not conservatism, it is the
only safe move. Guessing "list operator" for `first` makes the block form a
syntax error:

    sub fake (@) { }   fake { $_ > 1 } (1,2)   ->  syntax error near "} ("

while `first { $_ > 1 } @a` is valid and returns 2. A wrong shape turns
working code into an error, which is the guess the plan forbids -- a tree
that is not a parse of its source.

So an XS-backed import contributes its NAME (the call is known to exist, so
a bareword call to it is a Call rather than an Unknown) and no shape: the
call parses in its parenthesised form and a parenless one stays
`Call{Resolved:false}`. That is strictly better than today, where the name is
unknown too.

## Acceptance

- [ ] A module named in `use` is located on the search path and parsed (`go test ./internal/parse/ -run '^TestResolvesModuleSource$' -count=1 -v | grep -q '^--- PASS: TestResolvesModuleSource'`)
- [ ] `our @EXPORT = qw(...)` is read from the module's own text (`go test ./internal/parse/ -run '^TestReadsLiteralExportList$' -count=1 -v | grep -q '^--- PASS: TestReadsLiteralExportList'`)
- [ ] A sub's prototype is taken from its declaration, with an explicit null for none (`go test ./internal/parse/ -run '^TestReadsPrototypeFromSource$' -count=1 -v | grep -q '^--- PASS: TestReadsPrototypeFromSource'`)
- [ ] Parse shape is DERIVED from the prototype, never stored (`go test ./internal/parse/ -run '^TestShapeDerivedFromPrototype$' -count=1 -v | grep -q '^--- PASS: TestShapeDerivedFromPrototype'`)
- [ ] `(;$)` parses with zero and with one argument (`go test ./internal/parse/ -run '^TestOptionalArgPrototype$' -count=1 -v | grep -q '^--- PASS: TestOptionalArgPrototype'`)
- [ ] A leading `&` licenses the block form; a non-initial `&` and `\&` do not. Asserted on the PARSE, never on perl's error text (`go test ./internal/parse/ -run '^TestBlockFormNeedsLeadingAmpersand$' -count=1 -v | grep -q '^--- PASS: TestBlockFormNeedsLeadingAmpersand'`)
- [ ] `use Test::More` makes `subtest 'x' => sub {}` parse (`go test ./internal/parse/ -run '^TestImportedListOperator$' -count=1 -v | grep -q '^--- PASS: TestImportedListOperator'`)
- [ ] An import list in the `use` restricts what is imported (`go test ./internal/parse/ -run '^TestImportListRestricts$' -count=1 -v | grep -q '^--- PASS: TestImportListRestricts'`)
- [ ] A module the parser cannot find, or one that computes its exports, leaves calls `Call{Resolved:false}` rather than guessing (`go test ./internal/parse/ -run '^TestUnresolvableImportDeclines$' -count=1 -v | grep -q '^--- PASS: TestUnresolvableImportDeclines'`)
- [ ] No perl subprocess is involved, at parse time or build time (`go test ./internal/parse/ -run '^TestNoPerlDependency$' -count=1 -v | grep -q '^--- PASS: TestNoPerlDependency'`)
- [ ] An XS-backed module contributes names but not shapes, and a parenless call to one stays unresolved (`go test ./internal/parse/ -run '^TestXSModuleNamesWithoutShapes$' -count=1 -v | grep -q '^--- PASS: TestXSModuleNamesWithoutShapes'`)
- [ ] The T1 rate is re-measured and the change recorded

Spec: ch4 §4.8.3.
