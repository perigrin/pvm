---
title: The lexer does not form x=, the word-shaped compound assignment
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0cf64-8436-7f82-bcb5-587f0eba266f
created: 2026-09-23T12:57:39.474460281Z
updated: 2026-09-26T09:00:31.867520986Z
sessions:
- start_sha: 2963c47d4099f32e5bcd3a131d0865f204ee28ea
  end_sha: ""
  commits: 0
  started_at: 2026-09-26T08:58:57.630191229Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T08:58:57.630191229Z
---

`x=` is the repeat-assignment operator and our lexer does not produce the
token. Found by the conformance corpus while writing tier 04's compound
assignment family (issue 01a0cc04-32c9).

## Measured

    my $t = "ab"; $t x= 2;
      Variable($t) Word(x) Operator(=) Number(2)     trailing_tokens

    my $t = "ab"; $t .= "b";
      Variable($t) Operator(.=) Quote("b")           parses

    my $t = "ab"; $t = $t x 2;
      Variable($t) Operator(=) Variable($t) Word(x)  parses

So the binary `x` is lexed correctly as a Word and the compound form is
not lexed at all: `x=` arrives as two tokens and the parser then has a
`Word` where an operator belongs.

## Why it is the odd one of the thirteen

`x` is the only WORD-SHAPED operator in the compound assignment family;
the other twelve are punctuation. A lexer that scans punctuation runs for
compound assignment will never reach it, which is consistent with what
the token dump shows.

Note the spacing hazard that comes with it: `$t x= 2` needs the space
before `x`, because `$tx= 2` is a different variable. Any fix has to form
the token from a word-position `x` followed immediately by `=`, without
absorbing an identifier.

## Where it is recorded

`conformance/04_operators/24_compound_repeat.t` carries the refusal and
bisects it against `.=` and the binary `x`, both of which parse.
