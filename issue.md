---
title: Resolve imports by parsing the module's own source
state: done
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0b835-1270-7b32-9f0a-18fe7c555bae
- 01a0b84d-caa8-71d2-b42b-0c8445bc6026
created: 2026-09-17T02:44:01.961043269Z
updated: 2026-10-09T17:21:13.540952705Z
sessions:
- start_sha: 528b9d835526d7a2c3acb41646c44cf6fd145437
  end_sha: f066f2ab7533480d37f08cd938d35dc99f850cad
  commits: 1
  started_at: 2026-09-19T06:16:22.439105713Z
  ended_at: 2026-09-19T06:36:30.651145868Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T06:16:22.439105713Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-19T06:36:30.651145868Z
observed_paths:
- internal/parse/module.go
- internal/parse/module_test.go
- internal/parse/parse.go
- internal/parse/use.go
---

Give the parser a door that knows where it is, so a `use` can be resolved
against the file's own directory at all.

Split from the original 12-AC issue on 2026-09-19. This is part 1 of 3: the
loading seam and the decline path. Reading exports and prototypes is
01a0b84d-caa8; deriving parse shape and using it at the call site is 01a0b84d-ec59.
The measurement and the argument for parsing the module's source rather than
running perl are unchanged and recorded below.

## Why a path, not just bytes

`Parse(src []byte) *Node` is the whole public surface today and it does no
file IO. A module's location is relative to the file that used it:
`use My::Local` in `t/foo.t` looks for `t/My/Local.pm` and `lib/My/Local.pm`.
Bytes alone can never know that, which is the issue's "local modules, which
are never installed anywhere, get no coverage" requirement.

    func ParseFile(path string) (*Node, error)   // resolves against path's dir
    func Parse(src []byte) *Node                 // unchanged, resolves nothing

`Parse` stays byte-identical in behaviour so every existing caller, ratchet
and fuzz target is untouched. The loader is unexported state on `parser`,
injectable in tests so the suite needs no filesystem.

## The error is the read, and only the read

The parser is TOTAL: every input produces a tree. Syntax it cannot handle
becomes `Unknown` carrying its text; a bareword it cannot identify becomes
`Call{Resolved:false}`. Three things depend on that and none of them tolerate
an error return:

- the ratchets count Unknown NODES, so "parsed 900 statements, 3 opaque" must
  not collapse into "failed"
- round-trip is unconditional -- every tree re-emits its source byte for byte
- the fuzzer asserts a tree comes back for arbitrary bytes

So `ParseFile`'s error is `os.ReadFile` failing on the TOP-LEVEL file: no such
file, permissions, a directory. Once bytes exist, a tree exists.

## A use that cannot be resolved is not a failure

perl dies on `use Nonexistent::Mod`. This parser must not, because parsing
text is not running it. The `use` statement itself is pure syntax -- keyword,
bareword, optional list, semicolon -- and parses regardless of whether the
file exists. What fails is only the ENRICHMENT:

    use Unfindable;          Use node, complete and correct
    subtest 'x' => sub {};   Call{Resolved:false}, no args consumed

versus resolved, where the same call takes its list. Both are valid trees,
both round-trip, one knows more. `Resolved:false` IS the error value: in
band, attached to the node it is about, instead of aborting the file.

Were it an error, the 986-file T1 ratchet would lose every file naming an
uninstalled module -- which is most of T1, since PerlOnJava's tests are not
installed anywhere.

A module found but unreadable is treated the same as not found: silently
`Resolved:false`, because the practical consequence is identical.

## Parse the module's source -- do NOT run perl

To handle `use Test::More`, find `Test/More.pm` on the search path and PARSE
it, with the parser this milestone is building. Recursive, and it terminates.

An earlier draft said to generate a stub from perl's symbol table
(`perl -MTest::More -e 'print prototype(...)'`). Two sessions endorsed that
and both were wrong. perigrin's argument is decisive and is the reason this
project exists:

**reading the symbol table means libperl has parsed the text.** Moving that
from parse time to build time does not remove the dependency, it hides it. A
Perl parser that must run Perl to parse Perl is circular.

It fails three further ways: a repo whose dependencies are not installed
cannot be analysed; the data pins to whichever module version was on the
generating machine; local modules get no coverage.

This is the same answer the Chalk side reached for C headers, from the other
direction: for C you parse the header, for Perl you parse the module. Both
are static reads of declaration syntax, and Perl is the easier case because
the file is already in a language this parser handles.

## Termination

`use` is recursive by nature and a cycle is ordinary (`A` uses `B` uses `A`).
A visited set keyed by resolved path, plus a parsed-module cache, makes the
recursion terminate and stops re-parsing Test::More once per file that uses
it.

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

## Acceptance criteria

- [x] A module named in `use` is located on the search path and parsed (`go test ./internal/parse/ -run '^TestResolvesModuleSource$' -count=1 -v | grep -q '^--- PASS: TestResolvesModuleSource'`)
- [x] `ParseFile` returns an error for an unreadable top-level file and never for anything found during resolution (`go test ./internal/parse/ -run '^TestParseFileErrorIsTheReadOnly$' -count=1 -v | grep -q '^--- PASS: TestParseFileErrorIsTheReadOnly'`)
- [x] A `use` naming a module that cannot be found still parses, and calls to its imports stay `Call{Resolved:false}` rather than guessing (`go test ./internal/parse/ -run '^TestUnresolvableImportDeclines$' -count=1 -v | grep -q '^--- PASS: TestUnresolvableImportDeclines'`)
- [x] A `use` cycle terminates, and a module used twice is parsed once (`go test ./internal/parse/ -run '^TestModuleCycleTerminates$' -count=1 -v | grep -q '^--- PASS: TestModuleCycleTerminates'`)
- [x] No perl subprocess is involved, at parse time or build time (`go test ./internal/parse/ -run '^TestNoPerlDependency$' -count=1 -v | grep -q '^--- PASS: TestNoPerlDependency'`)
- [x] `Parse([]byte)` behaviour is unchanged: the T1 and T2 ratchets do not move on this commit

Spec: ch4 §4.8.3. Blocks 01a0b835 (the oracle repoint needs prototypes).


## Done, 2026-09-19 (f066f2ab)

    ParseFile        added, resolves against the file's own directory
    Parse            unchanged -- every ratchet holds
    full suite       green, 0 FAIL across the repo

All six acceptance criteria verified by running each command as written.

## What landed

    func ParseFile(path string) (*Node, error)   // resolves against path's dir
    func ParseWithLoader(src []byte, load Loader) *Node
    func Parse(src []byte) *Node                 // unchanged

`Loader` is the seam: `func(module string) ([]byte, bool)`. Production is
`DirLoader`, which searches directories using perl's own spelling
(`My::Mod` -> `My/Mod.pm`); tests hand back bytes they wrote themselves, so
the suite needs no installed module and touches no fixed layout.

The bool is "found", not "succeeded". A module that cannot be located and one
that cannot be read are the same answer, because the consequence is
identical, and TestParseFileErrorIsTheReadOnly pins that with a chmod 000
file.

## Termination, and where the mark goes

The visited set is keyed by module name and marked BEFORE parsing rather than
after. The module's own `use` statements are read during that parse, so a
cycle returns through the same door; marking after would recurse forever. It
doubles as the cache -- Test::More is parsed once, not once per file that
uses it.

## Measured while writing the test, and it belongs to the third issue

    subtest 'x' => sub { 1 };   ->   Unknown, not Call{Resolved:false}

A bareword followed by a string does not reach the Call path at all, so the
691 parenless subtests in T1 are OPAQUE rather than merely unresolved. That
is a bigger gap than the issue described and it is 01a0b84d-ec59's to close.

My first draft of TestUnresolvableImportDeclines asserted
`Call{Resolved:false}` and failed, which was the test reaching into the next
issue rather than a defect here. Corrected to assert the negative -- no
RESOLVED call to subtest -- which holds now and keeps holding once the shape
work lands.

## loaded is unexported, and on the root alone

It says what the PARSE did, not what the source contains. A consumer walking
children must not find it hanging off an interior node, it plays no part in
the span arithmetic, and it never affects round-trip. Read through
`LoadedModules(root)`; empty for a tree from `Parse`.

## Next

01a0b84d-caa8 reads exports and prototypes out of the bytes this issue
delivers. 01a0b84d-ec59 spends them at the call site, and owns the Unknown
measured above.
