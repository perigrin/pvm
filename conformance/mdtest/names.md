# Names: braces and packages

Where the variable's NAME is the question rather than what it holds.
Braces around a name are punctuation; `$::` is a package-qualified name,
the scanner row measured at 0.0% clean over fourteen files.

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
