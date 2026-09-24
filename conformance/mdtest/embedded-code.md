# Perl inside a pattern

`(?{ })` and a `qr//` that carries one -- a statement written inside a
regex, and a value that carries that statement across statements.

**Tier 14 recursive.** Introduces `entereval`, `substcont`. Depends on
09_regex.

NEITHER CONSTRUCT EMITS AN OP OF ITS OWN, and that is the tier's main
finding. Measured, `$s =~ /a(?{ $n = 1 })b/` compiles to one `match` op
whose pattern is the string `"a(?{ $n = 1 })b"`; the embedded statement is
inside the op's own data, not in the op stream. The full tree does show the
code's optree hanging off the match, but every op in it is marked `-`,
optimised out of the execution path: it lives in a separate CV the regex
engine calls, and `-exec` walks the outer path only. `qr//` with embedded
code measures the same -- one `qr` op, pattern string carrying the block,
indistinguishable from plain `qr//` by anything but its text.

So the construct that most plainly re-enters Perl is INVISIBLE to a
measurement of ops. These cases carry their weight in `tokens`, not in
`INTRODUCES`.

Neither case needs `use re 'eval'` or a warnings pragma. Measured under
5.42.0, a LITERAL code block in a pattern compiles and runs clean under
`use strict; use warnings`; the pragma is required only when the pattern
itself is interpolated from a variable, which neither case does.

## `(?{ })`: a statement inside a pattern

The block runs when the engine reaches that point in the match, and the
tokens are the only place this case can make a claim -- the block is
INSIDE the match's one token, not a brace and statements beside it.

The `no` claims name `5` and `/` deliberately. `$k` appears as a real
variable token on its own declaration line, so a bare "no variable" claim
would be false -- but the `5` inside the block appears NOWHERE else, so a
numeric literal with that text can only come from a lexer that read into
the pattern. That is the claim worth making.

```perl
my $s = "abc";
my $k = 0;
$s =~ /b(?{ $k = 5 })/;
print "$k\n";
```

```behavior
parses: yes
```

```output
5
```

```tokens
no numeric literal whose text is "5"
no operator whose text is "/"
```

## `qr//` carrying a block: code frozen into a value

A `qr//` carrying a code block freezes the embedded program into a value:
the code travels with the compiled pattern and runs wherever the object is
later matched.

Tier 09 claims plain `qr//` and says it claims only the non-recursive
form, leaving this tier "only what embedding adds". Measured, what
embedding adds to the `qr` op itself is NOTHING. What changes is
downstream, and it is still tier 09's op: matching against the compiled
object emits `regcomp` then `match`, where matching a literal pattern
emits `match` alone -- which tier 09 already documents as the cost of
interpolation, not of embedding.

So this case's whole subject is that a value can CARRY Perl code across
statements. The assignment and the match are separated on purpose: the
block is written on line 2 and runs on line 4, and `$n` proves it.

The `qr` object is not printed. Its stringification would embed the
block's source, which is deterministic, but printing `$n` is the claim
this case is actually making.

```perl
my $n = 0;
my $r = qr/a(?{ $n = 7 })b/;
my $s = "ab";
$s =~ $r;
print "$n\n";
```

```behavior
parses: yes
```

```output
7
```

```tokens
one quote-like operator whose text is "qr/a(?{ $n = 7 })b/"
one operator whose text is "=~"
no numeric literal whose text is "7"
```

## Hostile contents: proving the region is parsed, not counted

The tier's other cases hold INERT regions. `$n+1`, `$k = 5` and `$n = 7`
are delimited identically by a lexer that counts braces and by one that
parses Perl, so those cases establish that the region is DELIMITED and
leave the RE-ENTRY half of the tier's thesis unasserted. This case is the
inverse of `13_opaque`'s data-not-Perl case: there the hostile content
proved a region was NOT lexed, here it proves a region IS.

MEASURED, AND THE TWO CONSTRUCTS DIFFER. Under 5.42.0 perl uses two
different strategies to find the end of a re-entrant region, and the
difference is observable:

- `(?{ ... })` is PARSED AS PERL. Measured, `/b(?{ $k = length("}}}") })/`
  compiles and prints 3 -- the three braces inside the string do NOT close
  the block. A brace counter would stop at the first and hand the parser
  `(?{ $k = length("}`.
- `s{a}{ ... }e` COUNTS DELIMITERS. Measured, `s{a}{ $n + length("}}") }e`
  is a syntax error in perl itself -- "Unmatched right curly bracket at -e
  line 1" -- because the brace in the string DOES close the replacement.
  The re-entry happens after the region is cut out, not while it is being
  found.

So the substitution's hostile content is PARENS rather than braces:
`s{a}{ $n + length("))") }e` compiles and prints 3bc, and a reader that
balanced some bracket other than its own delimiter would truncate it. That
is the strongest falsifiable claim the construct admits, and writing braces
there instead would pin a perl syntax error as though it were our bug.

The braced `s{}{}e` spelling rather than `s///e` is forced by the same
measurement from the other side: with a `/` delimiter, a `/` inside a
string in the replacement is a syntax error in perl too. A delimiter that
cannot appear inside the region leaves no hostile content to write.

Every replacement here is unfoldable: `$n + length("))")` keeps
`substcont`, where a constant replacement would fold to `const s/FOLD` and
erase the `/e` entirely. Measured, `length("))")` itself folds to
`const[IV 2]`, but the `add` around it does not, so the frame survives.

The token claims are where this case carries its weight, because the
optree cannot see the distinction at all: the `(?{ })` and the `qr//`
block are inside their ops' PATTERN STRINGS, one `match` and one `qr`. The
`no` claims name the hostile characters -- a lexer that stopped early would
have spilled the region's tail out as ordinary tokens, and `length` would
appear as a word beside the quote rather than inside it.

```perl
my $s = "abc";
my $n = 1;
my $k = 0;
$s =~ s{a}{ $n + length("))") }e;
$s =~ /b(?{ $k = length("}}}") })/;
my $r = qr/c(?{ $k = $k + length("}}") })/;
$s =~ $r;
print "$s $k\n";
```

```behavior
parses: yes
```

```output
3bc 5
```

```tokens
one quote-like operator whose text is "s{a}{ $n + length(\"))\") }e"
one quote-like operator whose text is "qr/c(?{ $k = $k + length(\"}}\") })/"
no word whose text is "length"
no operator whose text is "+"
```
