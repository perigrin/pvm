# Increment, negation and the file tests

perlop's levels 3, 5 and 10 -- between them one file, and it was about
something else.

**Tier 04 operators.** Introduces `preinc`, `postinc`, `predec`,
`postdec`, `ftis`. Depends on 03_context.

`postinc` moved here from 07_subroutines, which had claimed it as a side
effect of a file about argument aliasing and held it "for a gap rather
than on merit". perlop puts `++` four tiers before subroutines exist.

## Pre and post, increment and decrement

Four spellings separated by their RETURN VALUE rather than their effect:
a parser reading `$a++` as `++$a` prints the same final `$a` and a
different second field. `preinc` and `postinc` are separate ops, not one
op with a flag.

```perl
my $a = $ENV{X} // 5;
my $b = $a++;
my $c = ++$a;
my $d = $a--;
my $e = --$a;
print "$a $b $c $d $e\n";
```

```behavior
parses: yes
```

```output
5 5 7 7 5
```

```tokens
no operator whose text is "+"
```

## The magic string increment

A successor function on strings with no arithmetic in it: `"Az"` becomes
`"Ba"`, `"zz"` becomes `"aaa"`, `"a9"` becomes `"b0"`. perlop documents
it as a special case of `++` and no other operator in the language
behaves this way.

```perl
my $a = $ENV{X} // "Az";
my $b = $ENV{Y} // "zz";
my $c = $ENV{Z} // "a9";
$a++;
$b++;
$c++;
print "$a $b $c\n";
```

```behavior
parses: yes
```

```output
Ba aaa b0
```

```tokens
no operator whose text is "+"
```

## Unary `+` computes nothing and disambiguates

It is the only way to stop `print (...)` being read as a complete call.
`print (1+2)*3` is `(print(1+2))*3` -- the parens after a list operator
are its ARGUMENT LIST, so print takes `1+2`, prints 3, and the `*3`
multiplies print's return value and is discarded.

NEITHER READING IS AN ERROR. Both compile, both print a number, and the
number is wrong only if you meant the other one.

This case declares NO token fact, which is a decision. It carried one
and the fact was vacuous: `no operator whose text is "+("` names a
spelling absent from the lexer's operator table, so no input could ever
produce it. No replacement is honest either -- every numeric literal in
the source appears exactly twice, the source writes a bare `+`, and it
contains no `=` at all. The output is the whole claim and it is enough.

```perl
print (1+2)*3;
print "\n";
print +(1+2)*3;
print "\n";
```

```behavior
parses: yes
```

```output
3
9
```

## The file-test operators

`-e` is a NAMED UNARY whose name is punctuation, and the whole family
(`-e -d -f -s -z -r -w -x -M -A -C`) was absent from the corpus.

WHAT A PARSER GETS WRONG IS THE MINUS. `-e $f` is a file test; `-$e` is
negation; `-bareword` is the string `"-bareword"`. Three readings of one
character, decided by what follows it.

Both spellings are written, and the parenthesised one is what makes the
fork actual: `e($f)` would be a call to an undeclared sub, so that form
has no reading where `-e` is a minus applied to something.

This case REFUSES, and in a shape the corpus had not recorded before:
the parser emits no Unknown at all. Our lexer splits `-e` into
`Operator(-) Word(e)`, the parser reads that as a negated bareword and
builds a tree, and only the token fact fails.

```perl
my $f = $ENV{X} // "/etc/hostname";
print "[", (-e $f), "][", (-e($f)), "]\n";
```

```behavior
parses: yes
refuses: 01a0cf64-8436-7f82-bcb5-587f0eba266f
```

```output
[1][1]
```

```tokens
no operator whose text is "-"
```
