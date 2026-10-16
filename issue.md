---
title: Operands perl reads as bare words are declared with :bareword, not special-cased
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-08T17:50:47.011045955Z
updated: 2026-10-16T18:48:35.407390896Z
---

## Context

- paths: `internal/parse/declarations/CORE.tpm`, the parser's special cases for labels, filehandles, sort sub names, class names, hash-subscript keys and `-word` (internal/parse, internal/lexer), `docs/rfcs/0001-type-libraries.md`
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

perigrin, 2026-10-08: `:bareword` is a parameter attribute, a parser hint: the operand is read as a word, not an expression, and the parameter's type says what the word becomes. 01a118fe-8e5f introduces it for `=>` (`Str $lhs :bareword`). Perl reads other operands as words too (measured on 5.42 or with -MO=Deparse): `$h{time}` (hash key, `time` not called), `-foo` ("-foo"), `Foo::Bar->new` (Deparse: `'Foo::Bar'->new`), `next LOOP` (label), `print STDERR 1` (filehandle), `sort by @a` (sub name), `$o->meth` (method name), `require Foo::Bar` (module name). Each is special-cased in the parser today. The goal (every builtin and operator defined in CORE.tpm with a signature that lets it be parsed) wants them declared:

```perl
sub next (Label $label :bareword) None;
sub print (FileHandle $fh :bareword = select(): List[Str] @args = ($_)) Boolean|Undef;
```

## Steps

- [ ] Inventory every place the parser reads an operand as a word; measure each on 5.42
- [ ] For each that a parameter can state, declare it with `:bareword` and its type; move the parser's special case to read the declaration, failing tests first; parses must not change (ratchets)
- [ ] Bring perigrin any case a parameter cannot state (subscripts, `->` method names may belong to the postfix/`->` declaration issues)

## Acceptance criteria

- [ ] every bare-word operand the parser special-cases is either declared `:bareword` in CORE.tpm or listed in RFC 0001 as not a parameter (`go test ./internal/parse/ -run '^TestBarewordOperandsDeclared$' -count=1 -v | grep -q '^--- PASS: TestBarewordOperandsDeclared'`)
