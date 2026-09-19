---
title: Read a module's exports and prototypes from its own text
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0ad3f-e929-77a8-9b00-2a80b06ec42c
created: 2026-09-19T06:15:01.031946657Z
updated: 2026-09-19T06:15:01.031946657Z
---

Read what a module exports and what each sub's prototype is, from the
module's own text.

Split from 01a0ad3f on 2026-09-19. This is part 2 of 3: 01a0ad3f builds the
loader that gets the module's bytes; this issue reads facts out of them;
01a0b8xx-B turns those facts into parse shape at the call site.

## Everything needed is in the text

Measured on Test/More.pm, 5.42.0:

    our @EXPORT = qw(ok use_ok ... pass fail ... subtest ...)   a literal qw
    sub ok ($;$) {        prototype on the sub line
    sub is ($$;$) {
    sub pass (;$) {
    sub subtest {         no prototype, visibly absent

    39  static `sub NAME` declarations
     0  glob-assignment generations (`*{...} = sub`)

Two facts per module, then: the export list, and a prototype per sub. Both
are ordinary declaration syntax this parser already handles -- `PrototypeNode`
exists today (decl.go:130-141) and carries the prototype string.

## One field: the prototype string

Not a shape enum. The prototype DETERMINES the shape, measured:

    sub np  { }        no prototype   np(1,2,3) swallows the list
    sub one ($) { }    ($)            `one 1, 2` parses as one(1), 2
    sub nil () { }     ()             `nil + 1`  parses as nil() + 1

So shape is derived, not stored. A stored `shape` field would be a second
copy of one fact -- the failure mode this project keeps finding. Deriving it
is 01a0b8xx-B; this issue stores the string and an explicit null for "no
prototype", which is distinct from "not known".

That three-way distinction is the whole point of this issue and it is what
the tiers below turn on.

## The three tiers

    name known, prototype known      Test::More      resolve fully
    name known, prototype unknown    List::Util      REFUSE, do not guess
    name unknown                     computed import Call{Resolved:false}

The middle tier is the XS case. Measured on List/Util.pm, 5.42.0:

    4   static `sub NAME` declarations
    2   XSLoader/bootstrap calls
    our @EXPORT_OK = qw(all any first min max ... reduce ...)   literal

The EXPORT list is readable and the prototypes are not. At runtime perl
reports `first &@`, `reduce &@`, `sum @`, `max @` -- none of it in the file.

So an XS-backed import contributes its NAME and no shape. The call parses in
its parenthesised form and a parenless one stays `Call{Resolved:false}`.
Strictly better than today, where the name is unknown too.

Refusing is not conservatism, it is the only safe move -- guessing "list
operator" for `first` makes the block form a syntax error:

    sub fake (@) { }   fake { $_ > 1 } (1,2)   ->  syntax error near "} ("

while `first { $_ > 1 } @a` is valid and returns 2. A wrong shape turns
working code into an error: a tree that is not a parse of its source.

## The import list restricts

`use Test::More qw(ok)` imports a subset, and the list is in the `use`
statement, so this is resolvable rather than assumed. No list means the
module's own `@EXPORT`.

A local sub shadowing an import is visible in the file being parsed and wins.
`require` plus manual import is not covered.

## Acceptance

- [ ] `our @EXPORT = qw(...)` is read from the module's own text (`go test ./internal/parse/ -run '^TestReadsLiteralExportList$' -count=1 -v | grep -q '^--- PASS: TestReadsLiteralExportList'`)
- [ ] `@EXPORT_OK` is read the same way, and a non-literal export list leaves the module unresolved rather than half-read (`go test ./internal/parse/ -run '^TestNonLiteralExportListDeclines$' -count=1 -v | grep -q '^--- PASS: TestNonLiteralExportListDeclines'`)
- [ ] A sub's prototype is taken from its declaration, with an explicit null for none, distinct from not-known (`go test ./internal/parse/ -run '^TestReadsPrototypeFromSource$' -count=1 -v | grep -q '^--- PASS: TestReadsPrototypeFromSource'`)
- [ ] An import list in the `use` restricts what is imported (`go test ./internal/parse/ -run '^TestImportListRestricts$' -count=1 -v | grep -q '^--- PASS: TestImportListRestricts'`)
- [ ] An XS-backed module contributes names but not shapes (`go test ./internal/parse/ -run '^TestXSModuleNamesWithoutShapes$' -count=1 -v | grep -q '^--- PASS: TestXSModuleNamesWithoutShapes'`)
- [ ] A local sub shadowing an import wins (`go test ./internal/parse/ -run '^TestLocalSubShadowsImport$' -count=1 -v | grep -q '^--- PASS: TestLocalSubShadowsImport'`)

Spec: ch4 §4.8.3. Blocked by 01a0ad3f (needs the loader).
