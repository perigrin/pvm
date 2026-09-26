---
title: A __DATA__ section has no statement form, and it is the section not the handle
state: in-progress
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T07:00:39.376307576Z
updated: 2026-09-26T07:02:01.342336494Z
sessions:
- start_sha: d1537f9631443c467a73f3e56cf382da4eeab2ae
  end_sha: ""
  commits: 0
  started_at: 2026-09-26T07:02:01.342336494Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T07:02:01.342336494Z
---

A `__DATA__` or `__END__` section refuses with `not_a_term`. The lexer
reads it correctly; the parser has no statement form for the token.

Two corpus cases, tier 13's own subject, and nothing tracked them.

## The obvious diagnosis is wrong, and that matters

Both failing cases read `<DATA>`, so the natural reading -- and the one
the earlier record carried -- is "the parser will not take `<DATA>` as a
term". Measured, that is FALSE:

    <DATA> with section      Unknown=1
    <DATA> no section        Unknown=0     <- the readline alone is fine
    <FOO> with section       Unknown=1
    <STDIN> with section     Unknown=1
    print 1; __DATA__ x      Unknown=1     <- NO readline at all
    <DATA> after __END__     Unknown=1

The discriminator is the SECTION, not the handle. `<DATA>`, `<FOO>` and
`<STDIN>` all parse without one and all refuse with one, and a program
with a section and no readline refuses too.

`<DATA>` appeared in every failing case because a program that reads a data
section usually has one, which is exactly the coincidence that makes a
count-then-attribute reading unsafe.

## What the lexer already does

    my @lines = <DATA>;
    print @lines;
    __DATA__
    one
    two

lexes to, among others:

    Readline("<DATA>")
    DataSection("__DATA__\none\ntwo\n")

Both correct. `DataSection` is a lexer Kind, the glossary has a
`data section` category, and tier 13 exists for constructs the optree
cannot see. So the work is a statement form, not lexing.

## A second, separate failure on the same case

The case also fails a token fact:

    one data section whose text is "__DATA__\none\ntwo\n\n"

Note the TRAILING BLANK LINE the fact wants; the lexer emits
`"__DATA__\none\ntwo\n"` without it. Whether the fact or the lexer is
right is a measurement nobody has made -- perl's `DATA` filehandle yields
`one\n` and `two\n`, and whether the section token should carry a final
newline beyond that is a boundary question the glossary's `data section`
entry does not settle.

Decide that BEFORE the statement form, because a statement form that makes
the parse clean leaves the case still failing on the fact, and a reader
would then think the fix did not work.

## Scope

1. Establish whether the section token should include a trailing newline.
   Measure against perl, and record the answer in the glossary's
   `data section` entry, which is silent on it.
2. A statement form for `__DATA__`/`__END__` so the section terminates the
   program rather than refusing. The body stays opaque.

## Context

- paths: `internal/parse/stmt.go`, `internal/lexer/scan.go` (`scanDataSection`, already present), `conformance/mdtest/pod-and-data.md`, `conformance/GLOSSARY.md`
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/conformance/ -count=1`
- entrypoints: whatever emits `not_a_term` when a `DataSection` token reaches statement position

## Acceptance criteria

- [ ] The data-section case parses (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -qE 'PASS: TestCorpus/pod-and-data.md/A_data_section_read_back_through_.DATA.'`)
- [ ] The not-a-program case parses (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q 'PASS: TestCorpus/pod-and-data.md/A_data_section_holding_bytes_that_are_not_a_program'`)
- [ ] Neither case leaves the sentinel behind (`! grep -q 'refuses: unfiled' conformance/mdtest/pod-and-data.md`)
- [ ] The glossary settles the trailing-newline boundary (`grep -qiE 'trailing newline|final newline' conformance/GLOSSARY.md`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the conformance suite holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
