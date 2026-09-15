---
title: POD, __END__, format, and s///e replacement sublexing
state: done
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T22:45:28.272209666Z
updated: 2026-09-15T14:02:34.20300231Z
sessions:
- start_sha: aae0b3f5e8184a64b19ed5fbdb5974dba0a430bb
  end_sha: a28f534a27eb8cf78a2413daaa60fd7d56012f33
  commits: 1
  started_at: 2026-09-15T13:57:16.427582776Z
  ended_at: 2026-09-15T14:02:34.20300231Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-15T13:57:16.427582776Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-15T14:02:34.20300231Z
observed_paths:
- internal/lexer/lexer.go
- internal/lexer/pod.go
- internal/lexer/pod_test.go
- internal/lexer/quote.go
---

Three lexer modes that are line-oriented rather than delimiter-oriented, and
one that re-enters the lexer. The M0 plan names all of them —

> it requires heredocs, POD, `__END__`, `q{}` with nested delimiters, and
> `s///e` all to be *recognised* — not understood, but delimited correctly

— and an earlier draft of this chain covered two of the five.

## POD, `__END__`, `__DATA__`

Spec §2.5, §2.15 item 12. A line starting `=` where a STATEMENT could start
begins POD; everything to `=cut` inclusive is trivia. The expect state
decides "where a statement could start", which is why this issue follows the
expect-state one.

`t/base/lex.t:590-630` is a deliberate terminator battery:

    =cute    =cut2    =cut_    =cut

Spec §2.5.1 gives the three-conjunct rule: `=cut` ends POD only when not
followed by an identifier character. `=cute` does NOT end it. A lexer that
matches a `=cut` prefix swallows the wrong span and still round-trips
perfectly, which is exactly the failure the golden token stream exists to
catch.

`t/base/lex.t:491-505` puts `=pod` immediately after `${` + newline and
`map{` + newline. Verified: perl prints `[42]`. That needs `XSTATE` inside a
bare-sigil block, §3.1.4.

6 of the 56 T2 files contain POD or `__END__`.

## format

§0.13 rank 3, 8 corpus files, 3 of them in T2 (`comp/decl.t`,
`comp/form_scope.t`, `comp/parser.t`). Spec §2.13, §2.15 item 16; perl's
`LEX_FORMLINE`.

A `format NAME =` line opens a body that ends at a lone `.` on its own line
— the same shape as a heredoc, and §0.13 says it "wants the same lexer
mode". Picture lines are not Perl and must not be lexed as Perl.

## s///e replacement sublexing

Spec §2.15 items 11 and 15. With `/e`, the replacement is CODE, not a
string: it must be lexed, and it can open constructs that outlive it.

`t/base/lex.t:114` is the case that makes this non-optional:

    $foo =~ s/^not /substr(<<EOF, 0, 0)/e;
      Ignored
    EOF

A heredoc opened inside the replacement, body on the following lines.
Verified: perl prints `[x]`. Treat the replacement as opaque and the `<<EOF`
is missed, the body lexes as barewords, and the file still round-trips.

`/ee` is nested eval of the result; recognise the modifier, do not
re-lex twice.

## Acceptance

- [ ] A `=word` line in statement position opens POD through `=cut` (`go test ./internal/lexer/ -run '^TestPodSpan$' -count=1 -v | grep -q '^--- PASS: TestPodSpan'`)
- [ ] `=cute`, `=cut2` and `=cut_` do not terminate POD; `=cut` does (`go test ./internal/lexer/ -run '^TestPodTerminatorBattery$' -count=1 -v | grep -q '^--- PASS: TestPodTerminatorBattery'`)
- [ ] `=pod` is recognised after `${` and `map{` at line start (`go test ./internal/lexer/ -run '^TestPodAfterBareSigilBlock$' -count=1 -v | grep -q '^--- PASS: TestPodAfterBareSigilBlock'`)
- [ ] `__END__` and `__DATA__` make the rest of the file one trivia token (`go test ./internal/lexer/ -run '^TestEndAndDataSection$' -count=1 -v | grep -q '^--- PASS: TestEndAndDataSection'`)
- [ ] A format body runs to a lone `.` and its picture lines are not lexed as Perl (`go test ./internal/lexer/ -run '^TestFormatBody$' -count=1 -v | grep -q '^--- PASS: TestFormatBody'`)
- [ ] An `/e` replacement is lexed as code, and a heredoc opened inside it takes its body after the line (`go test ./internal/lexer/ -run '^TestSubstitutionEvalSublex$' -count=1 -v | grep -q '^--- PASS: TestSubstitutionEvalSublex'`)

Findings: §0.13 rank 3 (8 files). Spec: §2.5, §2.5.1, §2.13, §2.15 items 11, 12, 15, 16. Plan §5 M0 prose.
