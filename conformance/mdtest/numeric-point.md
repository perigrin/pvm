# The decimal point, the exponent and the sign

Where a numeric literal STARTS and STOPS. Three of these five cases are
about a `.` that is also the concatenation operator, one is about the
sign of an EXPONENT, and one is about a `-` that is never part of the
number at all.

**Tier 01 literals.** Introduces `const`, `enter`, `leave`,
`multiconcat`, `nextstate`, `padrange`, `padsv`, `padsv_store`, `print`,
`pushmark`. Depends on nothing -- it is the only tier that can say that.
`my` and `print` appear throughout as FIXTURES rather than as subjects: a
literal has to be bound to something and observed somehow.

MEASURED perl 5.42.0, and this is the reason the token layer exists:

    $ perl -e 'printf "%.17g %.17g %.17g\n", .5, 0.5, 5e-1'
    0.5 0.5 0.5

Three spellings, the SAME VALUE to 17 significant digits. No behavioural
probe can see which was written, so each case here carries a token
assertion and the output pin is only a guard against a lexer that got
the value wrong as well.

## A decimal point between digits

`0.5` is one token, not `0` `.` `5`.

This is the case that already works, and it is here for two reasons. It
is the baseline the two cases after it deviate from -- `.5` and `5e-1`
are the same construct with the integer part removed and an exponent
sign added -- so a regression here would explain both without being
separately diagnosed. And a corpus whose every case refuses cannot
demonstrate that passing is reachable.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = 0.5; print "$x\n"'
    0.5

```perl
my $x = 0.5;
print "$x\n";
```

```behavior
parses: yes
```

```output
0.5
```

```tokens
one numeric literal whose text is "0.5"
no operator whose text is "."
```

## A leading decimal point

`.5` is one token, not the concatenation operator `.` followed by `5`.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = .5; print "$x\n"'
    0.5

This refused from 38c95d23 until the lexer learned the rule: our lexer
produced `Operator(.) Number(5)`, because `scanNumber` required a
leading DIGIT. Fixed under issue 01a0c13f-97f5-7f98-b32d-07245ec6ddfe,
which made a `.` before a digit start a number where a TERM is expected
and leave it as concatenation where an operator is. The token fact below
was written while it still refused, and it is what the fix had to
satisfy. Nothing in it changed.

```perl
my $x = .5;
print "$x\n";
```

```behavior
parses: yes
```

```output
0.5
```

```tokens
one numeric literal whose text is ".5"
no operator whose text is "."
```

## A trailing decimal point

A trailing point with no digits after it is still part of the numeric
literal: `1.` is one token, not `1` followed by the concatenation
operator.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = 1.; print "$x\n"'
    1

The mirror image of the leading point, and the pair is why both are
here: `.5` puts the dot where a lexer expects an operator and `1.` puts
it where a lexer expects more digits. A scanner that requires a digit on
BOTH sides of the point gets each of them wrong in a different way, and
a corpus holding only one of the two would diagnose that as a single
fault.

The output cannot distinguish the spellings -- `1.` prints `1`, exactly
as `1` does -- so the token assertion is what this case turns on.

```perl
my $x = 1.;
print "$x\n";
```

```behavior
parses: yes
```

```output
1
```

```tokens
one numeric literal whose text is "1."
no operator whose text is "."
```

## A leading minus is not part of the literal

`-1` is two tokens, a negation operator and the literal `1`. This is the
asymmetry the signed exponent is the other half of: the sign of an
EXPONENT is part of the token; a sign in front of the number never is.

MEASURED perl 5.42.0:

    $ perl -e 'my $x = -1; print "$x\n"'
    -1

    $ perl -MO=Concise -e 'my $x = -1;' 2>&1 | grep const
    const[IV -1] s/FOLD

The second measurement is why this case exists. The optree holds ONE
FOLDED CONSTANT: the optimiser has already applied the negation and
erased the operator, so nothing downstream of compilation can tell `-1`
from a hypothetical single negative literal. Behaviour cannot see it
either -- both would print `-1`. The token stream is the only place the
two tokens are still two, which is this tier's argument for the token
layer stated where it is cheapest to check.

Our lexer gets this right today: `Operator(-) Number(1)`. The case is
here to keep it right, not to report that it is wrong.

```perl
my $x = -1;
print "$x\n";
```

```behavior
parses: yes
```

```output
-1
```

```tokens
one operator whose text is "-"
no numeric literal whose text is "-1"
```

## The sign of an exponent is part of the literal

`5e-1` is one token, not `5e` minus `1`.

MEASURED perl 5.42.0:

    $ perl -e 'print 5e-1, "\n"'
    0.5

    $ perl -e 'print 5e, "\n"'
    Bareword found where operator expected (Missing operator before "e"?)
    syntax error at -e line 1, near "5e"

The second measurement is what makes the split WRONG rather than merely
different: `5e` is not a number, so a lexer emitting `Number("5e")` has
produced a token perl would reject.

REFUSES as of 38c95d23. Our lexer produces `Number(5e) Operator(-)
Number(1)`, and the same split affects `5e+1`, `5E-1` and `1.5e-3`.
Unsigned `5e1` lexes correctly, which is why it went unnoticed. The
refusal is LEXICAL, so the parser returns no Unknown node and there is
no refusal code to name.

No issue: this case is the record. The construct was found by writing
it, so there is nowhere earlier for it to have been filed, and
duplicating the token stream into a tracker would give it a second place
to go stale.

```perl
my $x = 5e-1;
print "$x\n";
```

```behavior
parses: yes
refuses: unfiled
```

```output
0.5
```

```tokens
one numeric literal whose text is "5e-1"
no operator whose text is "-"
```
