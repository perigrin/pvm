---
title: "Sigils are structural: the lexer swallows whole dereferences"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-17T03:04:49.380252705Z
updated: 2026-09-17T03:05:12.885823133Z
---

The lexer emits a whole dereference as ONE `Variable` token, so the parser
cannot build the tree §4.4.2 specifies. Measured:

    "${$h->{k}};"   ->  Variable "${$h->{k}}"   Semicolon
    "@{[ 1, 2 ]};"  ->  Variable "@{[ 1, 2 ]}"  Semicolon
    "$$x;"          ->  Variable "$$x"
    "@{$x};"        ->  Variable "@{$x}"

The second is an anonymous array constructor with a comma list inside a
dereference, delivered as a single lexeme. The parser then produces:

    "${$h->{k}};"   ->  [source_file statement term]
    "@{[ 1, 2 ]};"  ->  [source_file statement term]

One leaf, no children. That is not an imprecise tree, it is a WRONG one: the
parse claims there is no structure where there demonstrably is. Round-trip
cannot see it because every byte is present inside the leaf -- the same blind
spot canonical re-emission exists to close.

## What the grammar says

`perly.y:1884-1894` has six separate productions, each with its own
constructor, and `indirob` is RECURSIVE:

    indirob : BAREWORD | scalar %prec PREC_LOW | block | PRIVATEREF ;
    scalar  : PERLY_DOLLAR        indirob  { newSVREF }
    ary     : PERLY_SNAIL         indirob  { newAVREF }
    hsh     : PERLY_PERCENT_SIGN  indirob  { newHVREF }
    star    : PERLY_STAR          indirob  { newGVREF }
    amper   : PERLY_AMPERSAND     indirob  { newCVREF }
    arylen  : DOLSHARP            indirob  { newAVREF }

So `${...}` contains an arbitrary expression, `$$x` nests, and `@{[ 1, 2 ]}`
holds a real constructor. §4.4.2's own sketch assumes the parser sees the
sigil separately:

    func (p *Parser) parseSigil(sigil string) Node {
        case p.at("{"):    return &Deref{Sigil: sigil, Expr: p.parseBlockAsExpr()}
        case p.atSigil():  return &Deref{Sigil: sigil, Expr: p.parseSigil(...)}
        default:           return &Var{Sigil: sigil, Name: p.parseVarName()}
    }

`p.at("{")` and `p.atSigil()` cannot work when the lexer has already consumed
the whole chain.

## Two node kinds, not one Term

`Var{Sigil, Name}` and `Deref{Sigil, Expr}`. The sigil is STRUCTURAL, not a
spelling: `@a` and `$x` and `%h` reduce through different productions with
different constructors. An earlier decision collapsed them into one `Term`
with the sigil in `Text`, justified as "a kind per keyword does not make
walkers more precise" -- that reasoning is right for `my` versus `our`, which
IS a spelling, and wrong here.

Corroborated by the consumer: `internal/infer` already switches on `scalar`
(10 uses), `array` (10), `hash` (9) as distinct kinds.

## Acceptance

- [ ] The lexer emits a sigil as its own token, not a whole chain (`go test ./internal/lexer/ -run '^TestSigilIsItsOwnToken$' -count=1 -v | grep -q '^--- PASS: TestSigilIsItsOwnToken'`)
- [ ] `$x`, `@x`, `%x`, `$#x` produce Var nodes carrying their sigil (`go test ./internal/parse/ -run '^TestVarCarriesSigil$' -count=1 -v | grep -q '^--- PASS: TestVarCarriesSigil'`)
- [ ] `$$x` and `@$x` nest as Deref over Var (`go test ./internal/parse/ -run '^TestSigilChainNests$' -count=1 -v | grep -q '^--- PASS: TestSigilChainNests'`)
- [ ] `${ $h->{k} }` holds its inner expression as children (`go test ./internal/parse/ -run '^TestBlockDerefHoldsExpression$' -count=1 -v | grep -q '^--- PASS: TestBlockDerefHoldsExpression'`)
- [ ] `@{[ 1, 2 ]}` holds an AnonArray, not a leaf (`go test ./internal/parse/ -run '^TestDerefHoldsConstructor$' -count=1 -v | grep -q '^--- PASS: TestDerefHoldsConstructor'`)
- [ ] `$$x[0]` parses as `${$x}[0]`, per the %prec PREC_LOW rule (`go test ./internal/parse/ -run '^TestDerefSubscriptBinding$' -count=1 -v | grep -q '^--- PASS: TestDerefSubscriptBinding'`)
- [ ] 620/620 still round-trip, and the lexer ratchet and lex.t golden stream are re-baselined in the same commit (`go test ./internal/lexer/ -run '^TestLexerRatchet$' -count=1 -v | grep -q '^--- PASS: TestLexerRatchet'`)

Spec: ch4 §4.4.2, perly.y:1884-1894.
