# 14_recursive

`s///e`, `(?{ })`, and `qr//` carrying embedded code.

## Why this tier sits here

Every tier before this one has an operand smaller than the language. A
literal's operand is a sequence of characters, an operator's operands are
two expressions, a regex's operand is a pattern -- a second language, as
tier 09 says, but a second language, not this one. Here the operand is
PERL. `s/a/$n+1/e` puts an expression in the replacement; `(?{ $k = 5 })`
puts a statement inside a pattern; `qr/a(?{ ... })b/` freezes such a
pattern into a value that carries its code wherever it is interpolated.

That is a different demand on the lexer than anything earlier. Tier 13
asks it to find the END of a region it must not read -- a heredoc body, a
`qx` string, POD. That is hard, but it is one scan. Here the lexer must
find the end of a region and then HAND IT BACK, so the same lexer runs
again inside its own output, and the inner run can itself contain a
`s///e`. The nesting has no fixed depth. A lexer that special-cases one
level of re-entry passes every file in this tier except an adversarial
one, and the adjacency file is where that shows.

Check 2 asks whether this tier could move earlier. Measured, it could move
a long way: `my $s="abc"; my $n=1; $s =~ s/a/$n+1/e;` needs tier 01's
literal, tier 02's pad slot, tier 04's `add` and tier 09's `subst`, and
nothing from 03, 05, 06, 07, 08 or 10 through 13. There is no dependency
on `13_opaque` at all -- the two tiers solve different lexing problems and
neither is built on the other.

It sits last for a reason the dependency graph cannot express. The
construct's operand is the whole language, so every tier before it is
INSIDE ITS OWN SUBJECT: a `s///e` replacement may hold a method call
(tier 11), a dereference (tier 08), a heredoc (tier 13). The tier does not
depend on the thirteen before it; it CONTAINS them. Last is where a tier
goes when the honest statement of its scope is "everything already
written, one level down."

## DEPENDS ON

    09_regex

Not `13_opaque`, which is merely N-1. Nothing here is a heredoc, a format,
a `qx` or a POD block, and nothing here needs one. The pairing would
assert a dependency that does not exist -- the failure mode
`conformance/README.md` and tier 09 both warn about.

`09_regex` is the real floor, and it is exact rather than approximate:
every construct in this tier is a regex with something added. `s///e` is
tier 09's `s///` with a flag. `(?{ })` is tier 09's `m//` with a block in
the pattern. `qr//` with embedded code is tier 09's `qr//` with the same
block. Remove the regex and there is no construct left -- `/e` is not a
thing that exists apart from `s`. Tier 09 also declares this boundary from
its own side, claiming plain `qr//` and leaving this tier "only what
embedding adds", so the two tiers agree on where the line falls.

`04_operators` is the near miss. The measured files here use `add` in a
replacement expression, because a replacement that is a bare constant gets
folded away and measures nothing. But that is a choice of FIXTURE: the
expression could be any expression, and the tier's subject is that the
region is an expression at all, not which operator it holds.

## INTRODUCES

    entereval substcont

## Why those ops, and not the ones the source implies

Two ops, for three constructs, and the gap is this tier's main finding.

- **`(?{ })` and `(??{ })` emit NO op of their own.** Measured,
  `$s =~ /a(?{ $n = 1 })b/` compiles to one `match` op whose pattern is the
  string `"a(?{ $n = 1 })b"` -- the embedded statement is inside the op's
  own data, not in the op stream. The full (non-`-exec`) tree does show the
  code's optree hanging off the match, but every op in it is marked `-`,
  optimised out of the execution path: it lives in a separate CV the regex
  engine calls, and `-exec` walks the outer path only. So the construct
  that most clearly re-enters Perl is INVISIBLE to a measurement of ops.
  That is corpus lesson 3 at its sharpest -- the tier's subject is a lexical
  fact, and these files carry their weight in `expect tokens`, not in
  `INTRODUCES`.

- **`qr//` with embedded code adds nothing over plain `qr//` either.** Same
  measurement: one `qr` op, pattern string carries the block. What changes is
  downstream -- matching against the compiled object emits `regcomp` then
  `match`, where matching a literal pattern emits `match` alone. Both are
  tier 09's ops, and tier 09 already documents `regcomp` as what
  interpolation costs. Embedding code into a `qr` buys no op at all.

- **`substcont` is what `/e` actually adds.** Plain `s/a/z/` is one `subst`
  op taking a constant. `s/a/$n+1/e` is `subst(replstart->)` with the
  replacement's own ops -- `padsv`, `const`, `add` -- executed between it
  and a `substcont` that loops back for the next match. `substcont` is the
  re-entry made visible: it is the op that returns control to the
  substitution after the embedded program has run.

- **`entereval` is what the SECOND `e` adds.** `s/b/$c/ee` is the same
  `subst`/`substcont` frame with `entereval` in the middle: the first `e`
  evaluates the replacement as an expression yielding a string, the second
  evaluates THAT string as Perl at runtime. It is the only op in this
  corpus that compiles Perl the compiler never saw, which is why it is
  claimed rather than left as a footnote to `/e`.

- **No `add` claimed, though these files emit one.** `add` is tier 04's,
  used here as the replacement expression's body. A constant replacement
  would be folded to `const` and emit no `substcont` either -- `s/a/uc("z")/e`
  measures as `const[PV "Z"] s/FOLD` and a plain `subst`, the `/e` gone
  entirely. Every file here has a deliberately unfoldable replacement for
  that reason.

Neither `use re 'eval'` nor a warnings pragma is needed for anything in
this tier. Measured under 5.42.0, a LITERAL code block in a pattern
compiles and runs clean under `use strict; use warnings`; the pragma is
required only when the pattern itself is interpolated from a variable,
which no file here does. No file is marked `# STATUS refuses` on those
grounds.
