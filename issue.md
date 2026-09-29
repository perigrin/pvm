---
title: The lexer splits &. |. ^. ~. and ^^ into two tokens, though the parser's table already places them
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T05:02:14.224247164Z
updated: 2026-09-29T05:08:46.680415374Z
sessions:
- start_sha: f29b92cd4d9de9fefb99e4b55f458393f2f566e2
  end_sha: b8fecf7e33510c839f9187567c7e28afc742ae60
  commits: 1
  started_at: 2026-09-29T05:02:15.26079209Z
  ended_at: 2026-09-29T05:08:46.680415374Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T05:02:15.26079209Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T05:08:46.680415374Z
observed_paths:
- internal/lexer/scan.go
- internal/lexer/stringbitwise_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
---

The lexer never produces the string-bitwise operators `&.`, `|.`, `^.`, `~.`,
their assignments `&.=`, `|.=`, `^.=`, or logical xor `^^` / `^^=` (5.40) --
each splits into two tokens, although the parser's precedence table already
places every one of them:

    my $x = 22 &. 66;       canon  my $x = 22 & .;66;
    my $w = $a ^^ $b;       canon  my $w = $a ^ ^;$b;

Measured on 5.42.0 with -MO=Deparse (bitwise feature on): `$a |. $b`, `~.$s`
and `$a ^^ $b` are kept as written.

The four string-bitwise operators exist only under the `bitwise` feature; toke.c
gates them on FEATURE_BITWISE_IS_ENABLED, and unfeatured `$a |.5` is `$a | 0.5`.
The lexer does not track that feature, so it takes the operator unless the `.`
is immediately followed by a digit; `$a |.5` under the feature is the ceiling.

Reach: op/bop.t (15 Unknowns) and op/lop.t (9).

## Acceptance criteria

- [ ] The string-bitwise operators, their assignments, and ^^ / ^^= are single tokens, and a `.` before a digit still belongs to the number (`go test ./internal/lexer/ -run '^TestStringBitwiseAndLogicalXor$' -count=1 -v | grep -q '^--- PASS: TestStringBitwiseAndLogicalXor'`)
- [ ] REGRESSION GUARD: the lexer's golden token stream is unchanged (`go test ./internal/lexer/ -count=1`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
