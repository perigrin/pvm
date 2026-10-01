# The tie protocol

`tie` binds a variable to a class and `tied` asks whether it is bound.
Two words one lexer would call one family, and two different argument
grammars.

**Tier 11 oo.** Introduces `anonhash`, `bless`, `emptyavhv`, `isa`,
`method`, `method_named`, `method_super`, `methstart`, `stub`,
`tie`, `tied`. Depends on 08_references.

Both words were UNCLAIMED by every tier of this corpus, and the
placement was measured rather than argued. `07_subroutines` was the
first proposal, on the ground that `TIESCALAR` and `FETCH` are ordinary
named subs -- true, and not sufficient. `tie` REQUIRES the constructor
to return a BLESSED object, and a constructor that does not fails
SILENTLY:

    $ perl -e 'package C; sub TIESCALAR { my $s = "x"; return \$s }
        sub FETCH { return 42 }
        package main; tie my $c,"C"; print "[", $c, "]\n";'
    []

No error, no warning, an empty value. So the blessed constructor is not
optional, and the minimal one emits `emptyavhv` and `bless` -- both of
them this tier's, and both counted against these files because `opsOf`
reads inside a file's own CVs. Tier 11 is the EARLIEST tier that can
hold either word.

Both cases spell the class `$ENV{X} // "Counter"` to keep it a RUNTIME
value so the `tie` cannot fold; `X` is unset when the runner executes
the file. `bless {}, "Counter"` inside the constructor rather than
`$_[0]`: the class arrives as a string these cases already name, and
`$_[0]` would add tier 02's `aelemfast` for nothing the construct is
about.

THE SUBSTRING WORRY IS NOT REAL, and it cost a draft. `tie` is a prefix
of `tied`, so the obvious fear is that a source containing both cannot
count either. Measured against the checker: `checkTokenFact`
(`internal/conformance/fact.go:50`) compares `tokenText == text` on the
token's FULL text and never as a substring, so `one word whose text is
"tie"` counts exactly one even where `tied` also appears.

## `tie` is a list operator taking a VARIABLE

`tie VARIABLE, CLASS, LIST` takes a VARIABLE where every other call in
this tier takes an expression, and hands the trailing LIST to a
constructor it names by string. Three things about the argument list
are true of no other construct here: the first argument is a VARIABLE
and `tie my $x, ...` DECLARES it in the same breath; the second is a
CLASS NAMED BY STRING, resolved at runtime; and the trailing LIST is
passed to the constructor, so the extent of the argument list decides
what the class receives. Measured, main program only:

    b  <0> pushmark s
    c  <0> padsv[$x:41,42] sRM/LVINTRO   <- the VARIABLE, declared here
    d  <+> multideref($ENV{"X"}) sK
    e  <|> dor(other->f) sK/1
    f      <$> const[PV "Counter"] s     <- the CLASS, a runtime string
    g  <$> const[PV "a"] s               <- the LIST, two more elements
    h  <$> const[PV "b"] s
    i  <@> tie vK/3

`vK/3` is the arity: three children under the mark, which is variable,
class and the two-element list flattened. A parser that stopped the
argument extent at the class name emits `vK/2` and the program still
runs, printing the same thing, which is why the arity is recorded here
and the output cannot carry this half of the claim. A parser that read
`tie` as an ordinary named operator gets the second and third right and
the first wrong: `my $x` in argument position is a declaration, and
nothing in the call site says so.

THE OUTPUT IS THE DISPATCH. `42` is `FETCH`'s return value reached by
reading `$x`, and the read looks like an ordinary scalar read in the
source -- `print $x` -- which is the whole point: the call is INVISIBLE
at the call site. A parser that compiled `print $x` as a plain pad read
prints the empty string, because an untied `my $x` is undef.

The token fact pins `tie` as ONE word, the count a lexer folding the
keyword into the declaration beside it -- `tie my` read as one thing --
gets wrong. A negative `no word whose text is "tied"` was drafted and
WITHDRAWN: with full-text comparison there was never a way for `tied`
to appear in a source that does not write it, so the fact could not
fail and asserted nothing.

```perl
package Counter;
sub TIESCALAR { return bless {}, "Counter" }
sub FETCH { return 42 }
package main;
tie my $x, $ENV{X} // "Counter", "a", "b";
print $x, "\n";
```

```behavior
parses: yes
```

```output
42
```

```tokens
one word whose text is "tie"
```

## `tied` is a named unary, in boolean position

`tied` asks a variable whether it is tied and answers with the OBJECT
the tie is bound to, or with undef. Measured, main program only:

    m  <0> padsv[$x:41,43] sRM
    n  <1> tied sK/1            <- arity ONE, a named unary
    o  <|> cond_expr(other->p) lK/1
    q  <0> padsv[$plain:42,43] sRM
    r  <1> tied sK/1
    s  <|> cond_expr(other->t) lK/1

`<1>` is the arity and it is the parse: one child, no pushmark, which
is a named unary and not a list operator, where the `tie` two lines
earlier emits `<@> tie vK/2` -- a list op with a mark. Two words a
lexer would call one family, two different argument grammars, and the
op stream says so.

NO `ref` AND NO `defined`, which an earlier draft assumed were both
required. Measured, `tied $x` in the condition of a ternary is enough:
an untied variable gives undef, which is false, and a tied one gives a
blessed reference, which is true. `ref(tied $x)` would add tier 08's
`ref` and `defined(tied $y)` tier 04's `defined`; both are earlier
tiers and would be legal, and neither is needed.

BOTH CASES ARE WHAT MAKE THE OUTPUT FALSIFYING. `$x` is tied and
`$plain` is not, so the line carries a true answer and a false one. A
parser that read `tied` as always true prints `tt`; one that read it as
always false prints `uu`. `$plain = 1` rather than an undef variable,
because the question is whether the VARIABLE is tied and not whether
its value is defined: an untied undef would let a parser that compiled
`tied` as `defined` pass.

The fact asserted is `tie` and not `tied`, and that is the honest half:
there are TWO `tied`s here -- the true case and the false one -- and
the grammar admits only `one` and `no`, so `one word whose text is
"tied"` would be FALSE of a correct lex. `tie` IS one, and pinning it
says the binding and the question reached the token stream as different
words.

```perl
package Counter;
sub TIESCALAR { return bless {}, "Counter" }
sub FETCH { return 42 }
package main;
tie my $x, $ENV{X} // "Counter";
my $plain = 1;
print tied($x) ? "t" : "u", tied($plain) ? "t" : "u", "\n";
```

```behavior
parses: yes
```

```output
tu
```

```tokens
one word whose text is "tie"
```
