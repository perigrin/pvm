# Names: braces, carets and packages

Where the variable's NAME is the question rather than what it holds.
Braces around a name are punctuation; `$::` is a package-qualified name,
the scanner row measured at 0.0% clean over fourteen files; `$^O` folds
the caret into the name where a bare `$^` does not.

**Tier 02 variables.** Introduces `aassign`, `aelem`, `aelemfast`,
`aelemfast_lex`, `aelemfastlex_store`, `aslice`, `av2arylen`, `delete`,
`each`, `gv`, `gvsv`, `helem`, `hslice`, `multideref`, `padav`, `padhv`,
`push`, `rv2av`, `rv2hv`, `sassign`, `unshift`, `values`. Depends on
01_literals.

The package spellings are also where the tier's op list stops matching
the source a reader would write first. `sassign` -- the plain scalar
assignment -- is introduced by `$::x = 1`, not by `my $x = 1`, and the
four ops for a constant array subscript are only all reachable once both
a lexical and a package array are in the corpus.

## `${x}` is one variable, not a dereference

GLOSSARY.md says this outright under `variable` -- "`${name}` is
one variable: the braces are punctuation around a name. But `${ $ref }`
is NOT one token, because its contents are an expression requiring a
parser." The two spellings differ by what is inside the braces and by
nothing else, which makes this the tier's sharpest lexing question: a
lexer that sees `${` and commits to a dereference is wrong here, and a
lexer that sees `${` and commits to a name is wrong at tier 08. This
case takes the half that belongs to this tier; `${ $ref }` is tier 08's,
and writing it here would be reaching forward.

BOTH SIGILS ARE BRACED because the brace rule is about the NAME rather
than about scalars: `@{a}` is the same array as `@a`, and a lexer that
special-cased `${` would pass a case that only wrote the scalar form.

The ops are `padsv` and `padav`, identical to the unbraced spellings:
perl resolves the braces away entirely, so the optree cannot tell this
case from one without them. The claim is a LEXICAL one and only the
source records it -- the same situation tier 01 recorded for `-1`, whose
two tokens fold to one constant.

```perl
my $x = 42;
my @a = (7, 8);
print ${x}, "\n";
print scalar(@{a}), "\n";
```

```behavior
parses: yes
```

```output
42
2
```

## `$^O` is one variable and a bare `$^` is another

The caret control variables are the third naming question in this tier,
and the one where the OUTPUT ALONE CANNOT SEE THE ANSWER. `$^O` lexed as
`Variable($^)` and a separate `Word(O)` -- two tokens where perl has
one -- and `my $x = $^O;` still produced a tree that parsed and
round-tripped, canonicalising to `$^;O()`: the punctuation variable, a
semicolon, and a call to a sub named `O`. Unknown=0 on a confidently
wrong tree. That is why this case carries TOKEN FACTS: they fail today
and the output block does not.

The braced spelling `${^TAINT}` was already right, because
`bracedNameFollowsAt` grew a caret branch when `TestLexDotTGoldenStream`
caught `${^TEST}` splitting. One spelling of the rule was repaired and
its bare sibling was left -- the same shape as `$::`, where `$:` alone is
also a real variable and the name only forms when something follows.

WHERE THE NAME STOPS is the whole claim, and perl is the authority.
Measured on 5.42.0 by compiling `my $x = $^C;` for every C: `$^A`
through `$^Z` compile, and so do `$^_`, `$^^` and `$^[`. A LOWERCASE
letter does not -- `$^o` is `Bareword found where operator expected`, so
perl read `$^` and then a word. Digits fail the same way. So the rule is
the uppercase range plus `_`, `^` and `[`, NOT the identifier class: a
lexer that took any identifier byte would swallow the word after a bare
`$^`, which is the format top-of-page name and a real variable.

`$^O` and `$^T` are ASSIGNED here and never printed, because their
values are the platform name and the start time -- neither is the same
twice. They still reach the token stream, which is where this case's
claim lives. `$^W` is the one with a value fixed across platforms, `0`
under the plain `perl FILE` the runner uses, so it carries the output.

Interpolation is deliberately absent: `"$^"` is `$` followed by a
literal caret and warns `Use of uninitialized value $`, a different
question belonging to tier 01. Every op here is a tier 01 or 02
fixture -- `sassign`, `gvsv`, `const`, `print` -- so a ternary or a
`defined` would have reached forward into tiers 04 and 06.

Each caret variable is written EXACTLY ONCE, because a token fact
counts occurrences and the format's only forms are `one` and `no`. So
every value travels out through a package scalar rather than being
assigned and then read back -- including the bare `$^`, whose evidence
is the token fact rather than the output. The negatives carry their
half of the claim: the split produced a `Word` for the letter, so
`no word whose text is "O"` is reachable from this source and fails
whenever the caret stops binding its letter.

```perl
$::w = $^W;
$::o = $^O;
$::t = $^T;
$::c = $^;
print $::w, "\n";
print "read\n";
```

```behavior
parses: yes
```

```tokens
one variable whose text is "$^W"
one variable whose text is "$^O"
one variable whose text is "$^T"
one variable whose text is "$^"
no word whose text is "O"
no word whose text is "W"
no word whose text is "T"
```

```output
0
read
```

## A package array subscripts through `rv2av`

`$a[0]` on a lexical is `aelemfast_lex`. `$::a[0]` is `aelemfast`
behind an `rv2av` over a `gv`. Constant subscript, lexical or package,
read or write -- four ops for what reads as one construct, which is why
the tier's adjacency case carries all four rather than a representative
one.

```perl
@::a = (1, 2, 3);
print $::a[0], "\n";
print scalar(@::a), "\n";
```

```behavior
parses: yes
```

```output
1
3
```

## A package hash reaches storage through `rv2hv`

The element access is `multideref` either way -- the optimiser
folds the glob lookup into the deref chain just as it folds the pad
lookup -- so the difference this case pins is in naming the hash itself,
not in subscripting it. `scalar(keys %::h)` is what forces the bare name
into the optree, and that is `gv` then `rv2hv`.

```perl
%::h = (a => 1, b => 2);
print scalar(keys %::h), "\n";
print $::h{a}, "\n";
```

```behavior
parses: yes
```

```output
2
1
```

## A package scalar is where `sassign` enters the corpus

Tier 01's `my $x = 0.5` emits `padsv_store` and no `sassign` at
all -- the lexical store is one op, not an assignment over a variable.
The package scalar is the first place the two halves separate: `gvsv`
fetches the glob's scalar slot and `sassign` puts the value in it. So
the op a reader would look for in tier 01 is introduced four constructs
into tier 02, by the spelling nobody writes first.

`$::x` rather than `$main::x`: `::` with an empty package name IS
`main`, and the short spelling is the one that makes the lexing question
visible -- whether `$::` is a sigil plus a name that begins with a
separator.

```perl
$::x = 1;
print $::x, "\n";
```

```behavior
parses: yes
```

```output
1
```
