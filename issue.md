---
title: Canon emits a block argument it cannot re-lex, so its output does not re-parse
state: pending
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0d087-28dd-711f-a0d5-54cb8515910c
created: 2026-09-26T13:25:34.422410713Z
updated: 2026-09-26T13:25:51.202987955Z
---

Canon writes a `Block` that sits in an argument list as `{` + statements + `}`,
and each Statement child gets its own `;`. Re-lexed, that `;` terminates a
statement and the `}` is orphaned, so canon's own output does not re-parse.

Measured at `5c84df85`, with NO change to the parser. Both lines are canon
output for source the parser already reads cleanly:

    my @b = map { $_ + 1 } @a;    Unknown=0  canon "my @b = map({$_ + 1;}@a);"
    my @b = map({$_ + 1;}@a);     Unknown=1  canon "my @b = map({$_ + 1});}@a);"

The second line is the first line's canon fed back in. It does not parse, and
its own canon is mangled -- so the canon FIXPOINT (canon of canon equals canon)
is broken today for every `map`, `grep` and `sort` block.

## The mechanism

    $ token dump of `map({$_;}@a);`
      Word "map"  Operator "("  Operator "{"  Variable "$_"
      Semicolon ";"  CloseBracket "}"  Variable "@a"  CloseBracket ")"

The `{` carries `OpensBlock=false`. `map` is in `blockTaking`, but the lexer's
`nextIsOpenBrace` lookahead only fires when the brace is ADJACENT to the word
-- here a `(` sits between them, so the brace is classified from XTerm and
reads as an anonymous hash. `parseListOpBlock` asks the token for `OpensBlock`
(by design, `canon.go`'s own argument) and correctly declines, so the block is
never a block and the `;` ends the statement.

So the two readings disagree: canon EMITS a parenthesised call whose block
argument it cannot itself re-lex. Either canon must not parenthesise a call
whose first argument is a block, or it must not terminate the block's last
statement -- the measurement does not decide which, and the fix should.

## Why this is invisible today, and what it blocks

`TestCanonTokenIdentity` holds only the files `t1canon.ratchet` records as
clean, and 302 of 986 already disagree. The `map`/`grep`/`sort` files are among
the 302, so nothing in the suite currently fails on this.

It becomes visible the moment a new `Block` starts reaching an argument list.
`01a0d087` (ungated `defer`) measured exactly that: making `WORD BLOCK ARG`
parse also makes `print {$fh} ...` parse -- an improvement, Unknown 1 -> 0 --
and 46 previously-clean T1 files then break canon round-trip, all with the same
diagnostic, `"Semicolon ;"` where `"CloseBracket }"` was expected. 45 of those
46 contain no `defer` at all. That issue cannot land until this one does.

## Context

- paths: `internal/parse/canon.go` (the `Block` case, ~line 226, and the
  `blockForm` terminator decision, ~line 80), `internal/lexer/lexer.go:414`
  (`nextBraceIsBlock`, which needs the brace adjacent to the word),
  `internal/parse/call.go:275` (`parseListOpBlock`, which asks the token)
- commands: `go test ./internal/parse/ -run '^TestCanon' -count=1`
- blocks: `01a0d087-28dd-711f-a0d5-54cb8515910c`
- found by: agent:m1e-1 while measuring `01a0d087`

## Acceptance criteria

- [ ] Canon of a call with a block argument re-parses (`go test ./internal/parse/ -run '^TestCanonBlockArgumentRoundTrips$' -count=1 -v | grep -q '^--- PASS: TestCanonBlockArgumentRoundTrips'`)
- [ ] The canon ratchet holds or improves (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 | grep -q '^ok'`)
- [ ] The canon token identity gate holds (`go test ./internal/parse/ -run '^TestCanonTokenIdentity$' -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the parse ratchets hold (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
