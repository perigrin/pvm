---
title: Four parser gaps the corpus recorded, tracked so their citations point somewhere live
state: in-progress
urgency: normal
milestone: corpus-complete
created: 2026-09-23T17:10:01.515942446Z
updated: 2026-09-23T23:09:18.954607855Z
sessions:
- start_sha: 95cf8935c0ed1334c63e25964bff53ff6c7b9837
  end_sha: ""
  commits: 0
  started_at: 2026-09-23T22:36:45.951226458Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T22:36:45.951226458Z
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

## Acceptance criteria

- [ ] The three filehandle-slot refusals are gone and the files pass (`go test ./internal/conformance/ -run '^TestCorpus$/11_oo/16|^TestCorpus$/12_packages/1' -count=1 -v | grep -q 'PASS: TestCorpus/12_packages/10'`)
- [ ] A bareword filehandle still takes the slot (`go test ./internal/parse/ -run '^TestBarewordHandleStillTakesTheSlot$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] Compile-time tokens parse as arguments to print (`go test ./internal/parse/ -run '^TestCompileTokensAreNotFilehandles$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] Every remaining refusal citation names a live issue (`! grep -rh 'STATUS refuses' conformance/ | grep -q 01a0cf3e`)
- [ ] The whole conformance suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
