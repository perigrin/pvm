---
title: format and write have no statement form, and the lexer already reads the body
state: in-progress
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T06:58:32.800892893Z
updated: 2026-09-26T07:01:56.803102854Z
sessions:
- start_sha: d1537f9631443c467a73f3e56cf382da4eeab2ae
  end_sha: ""
  commits: 0
  started_at: 2026-09-26T07:01:56.803102854Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T07:01:56.803102854Z
---

`format NAME = ... .` and the `write` that uses it refuse with
`unimplemented_statement`. Two corpus cases, tier 13's own subject, and
nothing in any milestone tracks them.

Filed while measuring the path to "the entire corpus parses with no
errors": of the eight root causes behind the seventeen refusing cases,
six have issues and two do not. This is one of the two.

## Measured, 2026-09-26 at d1537f96

    formats.md/A format declaration and the `write` that uses it
      refuses: unfiled   unimplemented_statement
    formats.md/A picture line that is not Perl
      refuses: unfiled   unimplemented_statement

Both `refuses: unfiled`, which is the corpus saying the case is its own
record -- correct while nothing tracked it, and what this issue replaces.

## What the lexer already does, so the gap is narrower than it looks

The lexer HAS a `FormatBody` kind and produces it. From the tier-13
adjacency body's token dump:

    Word("format") Word("STDOUT") Operator("=")
    FormatBody("a report line\n.\n")

So the body is already delimited, and `conformance/GLOSSARY.md` has a
`format body` category for asserting about it. `conformance/README.md`
records why: the `format NAME =` introducer is ordinary tokens -- two words
and an operator -- and only the body is opaque.

The refusal is `unimplemented_statement`, which means the PARSER has no
statement form for it, not that the lexer cannot read it.

## What perl does

    format STDOUT =
    a fixed report line
    .
    write;

compiles and prints. A format declaration compiles to nothing until
something calls `write`, which is why tier 13 places it with the
constructs the optree cannot see and why a token fact is the only
assertion available about the body.

## Scope

A statement form for `format NAME = <body> .` and for `write`, enough that
both cases parse. The body stays opaque -- do NOT lex inside it; that is
the tier's whole point.

`write` with no argument is a named unary over the default filehandle;
`write FH` takes one. Measure which forms the two cases use before
implementing either.

## Context

- paths: `internal/parse/stmt.go`, `internal/lexer/scan.go` (`scanFormatBody`, already present), `conformance/mdtest/formats.md`
- commands: `go test ./internal/parse/ -count=1`, `go test ./internal/conformance/ -count=1`
- entrypoints: whatever emits `unimplemented_statement` for a `format` keyword in statement position

## Acceptance criteria

- [ ] The format declaration case parses (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q 'PASS: TestCorpus/formats.md/A_format_declaration_and_the_.write._that_uses_it'`)
- [ ] The picture-line case parses (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -qE 'PASS: TestCorpus/formats.md/A_picture_line_that_is_not_Perl'`)
- [ ] Neither case leaves the sentinel behind (`! grep -q 'refuses: unfiled' conformance/mdtest/formats.md`)
- [ ] REGRESSION GUARD, passes today: the parse suite holds (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the conformance suite holds (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
