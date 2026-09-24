# Arrays

Naming an array, counting it, and reaching one element of it. Four ways
of writing a subscript, which perl compiles to four different ops.

**Tier 02 variables.** Introduces `aassign`, `aelem`, `aelemfast`,
`aelemfast_lex`, `aelemfastlex_store`, `aslice`, `av2arylen`, `delete`,
`each`, `gv`, `gvsv`, `helem`, `hslice`, `multideref`, `padav`, `padhv`,
`push`, `rv2av`, `rv2hv`, `sassign`, `unshift`, `values`. Depends on
01_literals.

`scalar(@a)` throughout rather than `print "@a"`. Measured, the
interpolation emits `join` and a `gvsv` for `$"` -- the `gvsv` is `$"`,
which the interpolation reads -- and `print "$a[0]"` emits `stringify`.
None of those is naming a variable, and a case that printed `"@a"` would
drag two unclaimed ops into the tier.

## An array holds a list

The tier's baseline, the way `01_binary.t` is tier 01's: every
other case here starts by filling an array or a hash, so a regression
here explains all of them at once rather than being diagnosed nine
times.

The ops are `padav` for the array itself and `aassign` for the list
assignment. There is no `scalar` op: `scalar(@a)` compiles to the padav
in scalar context, so the keyword leaves no trace of its own.

```perl
my @a = (1, 2, 3);
print scalar(@a), "\n";
```

```behavior
parses: yes
```

```output
3
```

## A constant subscript reads as one op and writes as another

`$a[0]` on the right of `=` is not the same op as `$a[0]` on the
left. This is the asymmetry the tier README calls out: read and write
are the same three characters in the source, and a parser that treats
them as one node is not wrong -- but perl does not, and a corpus that
never writes an element cannot see the second op at all.

The write is `aelemfastlex_store` and the read is `aelemfast_lex`. Both
are the `_lex` spellings because `@a` is a lexical; the package
spellings are in the names topic.

```perl
my @a = (1, 2, 3);
$a[0] = 9;
print $a[0], "\n";
```

```behavior
parses: yes
```

```output
9
```

## `$#a`, and the subscript the optimiser will not fold

Two claims in one body because the second only exists in terms of
the first. `$#a` alone is `av2arylen`. `$a[$#a]` is the case where the
subscript is an EXPRESSION rather than a constant, so perl builds a
plain `aelem` instead of the `aelemfast_lex` above -- which makes
`aelem`, the op a reader would expect to be this tier's centre, an edge
case reachable only by writing the subscript this way.

```perl
my @a = (1, 2, 3);
print $#a, "\n";
print $a[$#a], "\n";
```

```behavior
parses: yes
```

```output
2
3
```

## An array slice: the sigil decides, not the name

`@a[0, 2]` is the `@` sigil on an array name with a LIST
subscript, returning several elements rather than one. The sigil is the
whole point: `$a[0]` and `@a[0]` name the same array and differ only in
what they hand back, so a lexer that binds the sigil to the name and a
parser that binds it to the subscript form disagree here and nowhere
else in the tier.

`13`, not `1 3`: the slice's two elements arrive as separate arguments
to `print` and `$,` is unset, so nothing separates them -- the same
reason tier 01's adjacency file prints `qw(a b c)` as `abc`.

```perl
my @a = (1, 2, 3);
print( (@a[0, 2]), "\n");
```

```behavior
parses: yes
```

```output
13
```
