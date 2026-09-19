---
title: Read a module's exports and prototypes from its own text
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0ad3f-e929-77a8-9b00-2a80b06ec42c
blocks:
- 01a0b84d-ec59-777f-8a00-71e407c9f469
created: 2026-09-19T06:15:01.031946657Z
updated: 2026-09-19T06:44:14.810211129Z
sessions:
- start_sha: f066f2ab7533480d37f08cd938d35dc99f850cad
  end_sha: ""
  commits: 0
  started_at: 2026-09-19T06:36:44.390389021Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T06:36:44.390389021Z
---

Read what a module exports and what each sub's prototype is, from the
module's own text.

Split from 01a0ad3f on 2026-09-19. This is part 2 of 3: 01a0ad3f builds the
loader that gets the module's bytes; this issue reads facts out of them;
01a0b84d-ec59 turns those facts into parse shape at the call site.

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
is 01a0b84d-ec59; this issue stores the string and an explicit null for "no
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

- [x] `our @EXPORT = qw(...)` is read from the module's own text (`go test ./internal/parse/ -run '^TestReadsLiteralExportList$' -count=1 -v | grep -q '^--- PASS: TestReadsLiteralExportList'`)
- [x] `@EXPORT_OK` is read the same way, and a non-literal export list leaves the module unresolved rather than half-read (`go test ./internal/parse/ -run '^TestNonLiteralExportListDeclines$' -count=1 -v | grep -q '^--- PASS: TestNonLiteralExportListDeclines'`)
- [x] A sub's prototype is taken from its declaration, with an explicit null for none, distinct from not-known (`go test ./internal/parse/ -run '^TestReadsPrototypeFromSource$' -count=1 -v | grep -q '^--- PASS: TestReadsPrototypeFromSource'`)
- [x] An import list in the `use` restricts what is imported (`go test ./internal/parse/ -run '^TestImportListRestricts$' -count=1 -v | grep -q '^--- PASS: TestImportListRestricts'`)
- [x] An XS-backed module contributes names but not shapes (`go test ./internal/parse/ -run '^TestXSModuleNamesWithoutShapes$' -count=1 -v | grep -q '^--- PASS: TestXSModuleNamesWithoutShapes'`)
- [x] A local sub shadowing an import wins (`go test ./internal/parse/ -run '^TestLocalSubShadowsImport$' -count=1 -v | grep -q '^--- PASS: TestLocalSubShadowsImport'`)

Spec: ch4 §4.8.3. Blocked by 01a0ad3f (needs the loader).


## Done, 2026-09-19 (7034e553)

    full suite     green, 0 FAIL across the repo
    Parse          untouched -- the whole block is gated on a non-nil resolver
    new grammar    none

All six acceptance criteria verified by running each command as written.

## It was extraction, not parsing

Measured before writing a line, and the issue's prediction held exactly:

    our @EXPORT = qw(ok is subtest);
      declaration "our"  ->  term "@EXPORT", term "qw(ok is subtest)"
    sub ok ($;$) { 1 }
      declaration "sub"  ->  term "ok", prototype "($;$)", block
    sub subtest { 1 }
      declaration "sub"  ->  term "subtest", block          <- no prototype node

The three-way distinction is already STRUCTURAL: a sub with no prototype has
no PrototypeNode child, which is a different tree from one that was never
read. So nothing in the grammar changed.

## What the API is

    type Import struct {
        Name           string
        Prototype      string   // "($)", "(;$)", "(&@)"; empty for none
        PrototypeKnown bool     // false = refuse, do not guess
        Local          bool     // declared in the using file, shadows
    }

    func Imports(root *Node) map[string]Import

`PrototypeKnown` is the field that carries the tiers. `sub subtest { }` is
known with an EMPTY prototype; an XS-backed `first` is not known at all.

## Three forms of the import list, and the empty one is not the missing one

    use M;             @EXPORT
    use M qw(a c);     exactly a and c
    use M ();          nothing -- explicit "load but import nothing"

A COMPUTED import list returns early rather than falling through to the empty
case: `use M (@names)` is not `use M ()`. That one was worth being careful
about -- the two are the same shape in the tree and opposite in meaning.

## Half an answer is worse than none

A computed export list marks the whole module opaque rather than contributing
the names it can see. A name missing from a partial list reads as "not
exported" when the truth is "not known", and a later pass cannot tell those
apart. Pinned by TestNonLiteralExportListDeclines with `map { "gen_$_" }`.

## no and require are not resolved

`no` UNIMPORTS and `require` is a runtime load with no import at all. Both
still parse identically; only the enrichment is skipped.

## The cycle guard gained a cache

f066f2ab's visited set now sits beside a facts cache, so a module used twice
is READ once as well as parsed once. A module currently being resolved
returns "not reached" rather than partial facts -- finite recursion without
claiming the module does not exist.

## Next

01a0b84d-ec59 derives parse shape from these prototypes and spends it at the
call site. It owns the Unknown measured in 01a0ad3f: `subtest 'x' => sub {}`
does not reach the Call path at all today, so the 691 parenless subtests in
T1 are opaque rather than merely unresolved.
