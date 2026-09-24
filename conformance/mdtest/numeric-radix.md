# Numbers by radix

The spellings that say what BASE the digits are in, plus the decimal
integer they all deviate from, plus the separator that is a word
character everywhere else in perl.

**Tier 01 literals.** Introduces `const`, `enter`, `leave`,
`multiconcat`, `nextstate`, `padrange`, `padsv`, `padsv_store`, `print`,
`pushmark`. Depends on nothing -- it is the only tier that can say that.
`my` and `print` appear throughout as FIXTURES rather than as subjects: a
literal has to be bound to something and observed somehow.

The radix is INVISIBLE after compilation. `0b1010`, `0xff`, `0377`,
`0o377` and `4_294_967_296` all arrive as a `const[IV ...]`,
indistinguishable from the same value written in decimal, so the optree
cannot say which spelling was written and output cannot either. Every
case here turns on its token fact for that reason -- except the leading
zero, where a lexer that does nothing special changes the VALUE and
output catches it.

## A decimal integer

The baseline every other spelling in this tier deviates from, and the
tier's simplest construct.

The tier had twelve construct files and none of them was a plain
integer: `0.5` is the FRACTIONAL boundary and the rest are hexadecimal,
binary, octal, exponent and v-string -- every one a deviation from a
baseline no file stated. `= 42;` appeared only in the adjacency body,
which introduces nothing, so the tier's simplest construct was the one
it never asserted. `TestTierLiteralsCoversGlossary` named `decimal
integer` among this tier's boundaries and passed anyway, because it
joined the adjacency source in with the construct files'; excluding that
file made three boundaries fail at once, this being the third.

MEASURED perl 5.42.0:

    $ perl -MO=Concise,-exec -e 'my $x = 42; print "$x\n"'
    ... const[IV 42] ... padsv_store ... multiconcat ... print ...

`const[IV 42]`, an INTEGER const, where `0.5` gives `const[NV 0.5]`.
Same op, different SV type, and the optree is the only place that
difference is visible -- output prints `42` either way. The token fact
is what pins the spelling.

```perl
my $x = 42;
print "$x\n";
```

```behavior
parses: yes
```

```output
42
```

```tokens
one numeric literal whose text is "42"
```

## Binary: the `0b` prefix

A `0b` prefix makes the digits after it binary: `0b1010` is one token
denoting ten.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = 0b1010; print "$x\n"'
    10

After compilation this is `const[IV 10]`, the same op `my $x = 10`
produces, so nothing downstream can say which spelling was written.
`perldata` lists the form under "Scalar value constructors", which is
why GLOSSARY.md carries it as a numeric literal.

```perl
my $x = 0b1010;
print "$x\n";
```

```behavior
parses: yes
```

```output
10
```

```tokens
one numeric literal whose text is "0b1010"
no operator whose text is "b"
```

## Hexadecimal: the `0x` prefix

A `0x` prefix makes the digits after it hexadecimal, letters included:
`0xff` is one token denoting 255.

The trap is that `ff` is also a legal identifier. A lexer scanning the
digit `0`, stopping at the first non-digit and handing `xff` to the word
scanner produces `Number(0) Word(xff)` -- which the parser reads as a
number beside a bareword, not as an error. That is why the negative fact
names the word.

MEASURED perl 5.42.0, the three radix spellings and one decimal are the
SAME value:

    $ perl -e 'printf "%s %s %s\n", 0xff, 0377, 0o377'
    255 255 255

```perl
my $x = 0xff;
print "$x\n";
```

```behavior
parses: yes
```

```output
255
```

```tokens
one numeric literal whose text is "0xff"
no word whose text is "xff"
```

## Octal by leading zero

A leading zero makes the digits after it octal: `0377` is one token
denoting 255, not three hundred and seventy-seven.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = 0377; print "$x\n"'
    255

The radix is carried by a character a decimal literal may also begin
with, so this is the one radix form a lexer can get wrong by doing
NOTHING: scan the digits, read them as decimal, and `0377` becomes 377
with no token boundary out of place and nothing to report. Only the
VALUE differs, which is why this case pins output as well as tokens.

```perl
my $x = 0377;
print "$x\n";
```

```behavior
parses: yes
```

```output
255
```

```tokens
one numeric literal whose text is "0377"
no numeric literal whose text is "377"
```

## Octal by the `0o` prefix

An explicit `0o` prefix is octal too: `0o377` is one token denoting 255,
the same value `0377` denotes. Perl 5.34 added the spelling so octal
need not be signalled by a leading zero alone.

It is a separate case from the leading zero because it is a separate
LEXICAL form -- the two agree on the value and compile to the same
`const[IV 255]`, so the optree cannot tell them apart and only the token
stream can say which was written.

The trap is the hexadecimal one: `o377` is a legal identifier, so a
lexer that stops the number at the first non-digit produces `Number(0)
Word(o377)` and the parser sees a number beside a bareword rather than
an error.

```perl
my $x = 0o377;
print "$x\n";
```

```behavior
parses: yes
```

```output
255
```

```tokens
one numeric literal whose text is "0o377"
no word whose text is "o377"
```

## Underscore separators

Underscores inside a numeric literal are separators, part of the token
and absent from the value: `4_294_967_296` is one token denoting
4294967296.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = 4_294_967_296; print "$x\n"'
    4294967296

The underscore is a WORD character everywhere else in perl, which is
what makes this a boundary rather than a detail. Two wrong answers are
reachable: a lexer that stops a number at the first non-digit produces
`Number(4) Word(_294_967_296)`, and a lexer that scans word characters
greedily after a digit produces one token whose text is right and whose
value is unparseable. The token facts below distinguish both from the
correct answer, which the printed value cannot -- the separators are
gone by the time anything prints.

```perl
my $x = 4_294_967_296;
print "$x\n";
```

```behavior
parses: yes
```

```output
4294967296
```

```tokens
one numeric literal whose text is "4_294_967_296"
no word whose text is "_294_967_296"
```
