---
title: Four parser gaps the corpus recorded, tracked so their citations point somewhere live
state: done
urgency: normal
milestone: corpus-complete
created: 2026-09-23T17:10:01.515942446Z
updated: 2026-09-24T20:47:33.216860306Z
sessions:
- start_sha: 95cf8935c0ed1334c63e25964bff53ff6c7b9837
  end_sha: fb2c85f6d26d5a529699c52bdcf5918891d0c8a9
  commits: 2
  started_at: 2026-09-23T22:36:45.951226458Z
  ended_at: 2026-09-23T23:16:31.834743736Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T22:36:45.951226458Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-23T23:16:31.834743736Z
observed_paths:
- conformance/11_oo/16_classname_ungated.t
- conformance/11_oo/17_defer_ungated.t
- conformance/12_packages/10_compile_tokens.t
- conformance/12_packages/11_line_directive.t
- internal/conformance/testdata/corpus.ratchet
- internal/parse/call.go
- internal/parse/compiletoken_test.go
- internal/parse/t2_test.go
---

Four parser gaps found by the `corpus-complete` milestone, each recorded
by a corpus file citing this issue.

## What they turned out to be

THREE WERE ONE BUG, and not the bug any of the three files thought.
`isBarewordHandle` used perl's all-caps filehandle convention, because a
static parser has no symbol table to ask. `__PACKAGE__`, `__LINE__`,
`__FILE__` and `__CLASS__` are all-caps, so each was taken into
`print`'s filehandle slot and the comma after it was stranded.

Perl draws the line itself, and the two readings are not both legal:

    $ perl -MO=Deparse -e 'print FOO, "\n";'
    No comma allowed after filehandle at -e line 1.

    $ perl -MO=Deparse -e 'print __PACKAGE__, "\n";'
    print __PACKAGE__, "\n";

Fixed in `831eaa1d` as a four-name exception, with a test asserting
`print STDERR "x"` still takes the handle.

The measurement corrected the issue's own framing three times:

- `print FOO, "\n"` refuses identically, so this was never about
  compile-time tokens as such.
- `__PACKAGE__` alone always parsed; only the handle slot was wrong.
- The `#line` directive ALREADY WORKED. `11_line_directive.t` refused
  on its `print __LINE__, " ", __FILE__` line, not on the directive it
  is named for.

Corpus 185 clean to 188, 27 refusing to 24. T2 `comp/package.t` 7 to 6.

THE FOURTH IS A DIFFERENT KIND OF THING and is now `01a0d087`.
`17_defer_ungated.t` refuses because `defer` sits in
`statementKeywords` for the GATED form, which is unimplemented, and
that entry also declines the ungated reading. Removing it is not
enough: `WORD BLOCK ARG` has no statement-level reader, the general
shape is indirect object notation this parser does not implement
(`new Foo;` refuses too), and removing the entry exposes a canon
round-trip defect on `defer.t` that the refusal was masking. Three
pieces, scoped in that issue; this file's citation now points there.

## The first fix was wrong, and the gate found it

`831eaa1d` fixed this with a four-name exception list
(`compileTimeToken`). That was wrong on one entry and incomplete on
another, and both faults had one cause: a list encodes a judgement per
name where perl encodes none.

`__CLASS__` is a value only under `use feature 'class'` and a bareword
FILEHANDLE otherwise -- `print __CLASS__, "x"` is "No comma allowed
after filehandle", the same diagnostic the fix cited as its proof for
`FOO`. Exempting it adopted the gated reading with no feature state,
and contradicted `11_oo/16_classname_ungated.t`, edited in that same
commit, whose subject is that ungated `print __CLASS__;` IS a
filehandle print. That file passed for the wrong reason. `__SUB__`
behaves identically and was missing.

`f4e35d7a` replaced the list with perl's own rule: a comma after a
handle is an error, so what FOLLOWS the word decides the slot. The
`$fh` branch already applied that test; the bareword branch did not.

That exposed an older bug behind it -- `startsTerm` called every
`lexer.Word` a term-starter, so word operators (`eq`, `cmp`, `lt`,
`and`) denied nothing and both branches took the slot on
`print FOO eq "x"`. The `$fh` half predated this milestone. Fixing it
moved 14 T1 files, took `base/num.t` from 44 refusals to 6, and took T2
clean from 17 of 56 to 19.

So this issue's third criterion names `TestCommaDeniesTheFilehandleSlot`
rather than the `TestCompileTokensAreNotFilehandles` the first fix
added: that test asserted the wrong rule and was replaced.

## Acceptance criteria

- [ ] The three filehandle-slot refusals are gone and the files pass (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q 'PASS: TestCorpus/compile-phase.md'`)
- [ ] A bareword filehandle still takes the slot (`go test ./internal/parse/ -run '^TestBarewordHandleStillTakesTheSlot$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] A comma after a bareword denies the filehandle slot, so compile-time tokens parse as arguments (`go test ./internal/parse/ -run '^TestCommaDeniesTheFilehandleSlot$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] Every remaining refusal citation names a live issue (`! grep -rh 'STATUS refuses' conformance/ | grep -q 01a0cf3e`)
- [ ] The whole conformance suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
