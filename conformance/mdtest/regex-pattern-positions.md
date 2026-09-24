# Patterns in unexpected positions

A pattern in an ARGUMENT slot and a regex operator on the LEFT of an
assignment. Both are positions a parser modelled on ordinary expressions
gets wrong, and in both the failure is silent.

**Tier 09 regex.** Introduces `match`, `pos`, `qr`, `regcomp`, `split`,
`subst`, `trans`. Depends on 01_literals.

Each case is arranged so the OUTPUT reports the mis-parse, because the
optree will not: `split`'s two spellings emit byte-identical ops while
printing different answers, and `pos` is the same op in rvalue and
lvalue position, differing only in perl's printed flags. Both subjects
read `$ENV{X}`, unset when the runner executes them, because a `split`
or a match on a constant is a folding candidate and a folded one
measures nothing.

## `split`'s first argument is a pattern

`split /,/, $s` puts a regex literal in an argument slot, where a parser
reading `/` as division produces a tree perl never builds.

THE FAILURE IS SILENT. `split /,/, $s` and `split ",", $s` behave
identically on every ordinary input, so a parser that mistook the
pattern for a division, recovered, and produced a string would print
exactly what a correct one prints.

Except on ONE input, and it is the case's whole measurement. A single
SPACE as split's first argument is perl's awk-compatibility special
case: the STRING `" "` means "split on runs of whitespace and discard
leading whitespace", while the PATTERN `/ /` means what it says, one
space. Measured under 5.42.0 on `"  a b "`, `split / /` gives `||a|b`
and `split " "` gives `a|b`. So the two spellings are a PARSE apart, and
perl itself distinguishes them at parse time rather than at runtime.

AND THE OPTREE CANNOT SEE IT. Measured, both spellings emit the same op
with the same pattern text printed inside it -- `split(/" "/ => @p:2,3)
[t4] vK/LVINTRO,ASSIGN,LEX,IMPLIM` -- byte-identical, for two programs
that print different things. That is this tier's standing argument in
its sharpest form, because here the erasure hides a difference the
output can still see.

The token facts are the only place the distinction can be asserted, and
they are COUNTED rather than merely forbidden. `no operator whose text
is "/"` forbids the division reading: a lexer that scanned `/ /` as two
divide operators around a space would emit two `Operator("/")` tokens
where the pattern belongs. `one string literal whose text is "\" \""`
is the counting half and the sharper of the two -- this source holds
exactly ONE quoted string spelled `" "`, the second split's argument,
and the first split's `/ /` must NOT be one. A lexer that read a
slash-delimited pattern as a string literal would produce TWO tokens
matching that text and fail the count. Measured, ours emits `Quote("/
/")` for the pattern and `Quote("\" \"")` for the string, and only the
second is a string literal by the glossary's rule.

Neither fact can be written as `one word whose text is "split"`: the
source calls `split` twice, on purpose, since the whole measurement is
the two spellings side by side.

```perl
my $s = $ENV{X} // "  a b ";
my @pat = split / /, $s;
my @str = split " ", $s;
print scalar(@pat), " [@pat] ", scalar(@str), " [@str]\n";
```

```behavior
parses: yes
```

```output
4 [  a b] 2 [a b]
```

```tokens
no operator whose text is "/"
one string literal whose text is "\" \""
```

## `pos` is a named operator in lvalue position

`pos` is LVALUE-CAPABLE and tied to the regex engine: it reads and
WRITES the match position `//g` leaves behind on a scalar.

THE LVALUE FORM IS WHY THIS IS A PARSE AND NOT A CALL. `pos($s) = 0`
puts a named operator on the LEFT of an assignment, which almost nothing
in Perl does -- measured, the optree puts the `pos` op underneath
`sassign`'s left arm, `<1> pos[t3] sKRM*/1` then `<2> sassign vKS/2`. A
parser that modelled `pos` as an ordinary named unary returning a value
would refuse that assignment or silently discard it, and the second case
is the dangerous one: the program still runs.

THE OUTPUT REPORTS THE DISCARD, which is the case's whole design. `//g`
on a scalar advances a stored position, so two successive `/b/g` matches
against `"abcabc"` find the two `b` characters in turn. Measured,
without the reset the two positions are `2 5`; with it they are `2 2`.
So a parser that dropped `pos($s) = 0` prints `2 5` where this case pins
`2 2`, and the discarded statement is the only thing that could account
for it.

The op is the same in both positions -- rvalue `my $a = pos($s)` and
lvalue `pos($s) = 0` differ in the flags perl prints, not in the op --
so the op claim is earned by either spelling and the LVALUE claim is
behavioural.

THE TOKEN FACTS ARE ABOUT THE MODIFIER, not about `pos`. `pos` lexes as
an ordinary `Word` and this source holds three of them, so the facts
grammar's `one`/`no` cannot pin it. What the facts CAN pin is the thing
`pos` is useless without: `/b/g` carries a MODIFIER, and the modifier is
part of the quote token. `no word whose text is "g"` forbids the
modifier escaping the quote -- a lexer that stopped at the closing
delimiter would leave `g` behind as a stray `Word("g")`, a program that
still parses and still runs, printing `2 2` for the wrong reason. `no
word whose text is "b"` forbids the pattern BODY escaping it, which is
what a lexer reading the slashes as division would produce. Together
they say the three characters `b/g` were consumed by the quote and by
nothing else.

A positive fact is not available for the glossary's reason, not by
omission: the source spells `m/b/g` twice, deliberately, since the
measurement IS the two matches, so `one quote-like operator whose text
is "m/b/g"` would be false at a count of two.

```perl
my $s = $ENV{X} // "abcabc";
$s =~ m/b/g;
my $a = pos($s);
pos($s) = 0;
$s =~ m/b/g;
my $b = pos($s);
print "$a $b\n";
```

```behavior
parses: yes
```

```output
2 2
```

```tokens
no word whose text is "g"
no word whose text is "b"
```
