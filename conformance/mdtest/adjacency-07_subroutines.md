# Every construct, each beside another

One body holding every construct this tier introduces, each adjacent to
another -- and adjacent to 06_control, the tier this one depends on.

**Tier 07 subroutines.** Introduces `anoncode`, `argcheck`,
`argdefelem`, `argelem`, `entersub`, `leavesub`, `lock`, `return`, `warn`.
Depends on 06_control.

The tier's other cases are one construct each, which is what makes them
diagnosable: when the signature case refuses, the construct that refused
is the only one present. That same property is why a corpus of such cases
cannot reach an ADJACENCY bug -- a parser that handles every construct
alone and mis-handles a pair goes green over the pair.

## The whole tier in one body

Present here: a named sub, a signature with a default, `@_` read
positionally, `@_` WRITTEN through, five call forms, an anonymous sub,
`return` both early and trailing, `lock`, and the builtin extent pair.

THE ALIASING WRITE IS THE ONE CONSTRUCT WHOSE EFFECT IS VISIBLE FROM
OUTSIDE THE SUB. `bump($seen)` returns nothing anyone looks at, and
`$seen` is 1 in the printed line only because `$_[0]++` reached the
caller's variable. It sits next to a signatured sub on purpose: `pick`
binds its arguments by signature and `bump` by `@_`, so both of the
tier's argument protocols are in one body, which is the pair a parser
handling each alone can still get wrong. They are two subs rather than
one because, measured 5.42.0, reading `@_` inside a signatured sub warns
-- `Use of @_ in scalar with signatured subroutine is experimental` --
and a warning on stderr is noise this corpus has no section to pin.

THE PAIRING WITH 06_CONTROL IS NOT DECORATION. The `for` loop with `next`
inside the body is the earlier tier's construct sitting directly against
this tier's `return`, which is the pair the DEPENDS ON line names. A
`return` reached from inside a loop has to unwind the loop as well as the
sub, and nothing else in the corpus puts those two exits next to each
other.

THE PROTOTYPE NEEDS A FEATURE TOGGLE, and the reason is the sharpest
single measurement here. `sub g ($)` is a PROTOTYPE only where the
signatures feature is OFF. This body says `use v5.36`, which turns
signatures on, and under it perl reads the same three characters as a
SIGNATURE and enforces arity instead -- measured 5.42.0,
`use v5.36; sub g ($) { "g" } print g 1, 2;` gives
`Too many arguments for subroutine 'main::g' (got 2; expected 1)`, while
without the pragma `sub g ($) { "g[$_[0]]" } print g 1, 2` prints
`g[1]2`. The same three characters, two different features, and only the
feature state decides which. The `no feature "signatures"` /
`use feature "signatures"` pair around `sub g` is what lets both readings
live in one body: `pick` is declared above it and keeps its signature,
`g` is declared inside it and gets a prototype.

A BLOCK WOULD HAVE BEEN TIDIER AND IS MEASURABLY WRONG.
`{ no feature "signatures"; sub g ($) {...} }` compiles the bare braces
as a loop -- `enterloop`, `stub`, `leaveloop` -- and `stub` is an op
11_oo introduces, so the dependency lint correctly refuses a file
reaching four tiers forward to declare a sub. The file-scope toggle emits
no ops at all.

THE LAST TWO PRINTED LINES ARE THE ADJACENCY THAT MATTERS. `print f 1, 2`
and `print g 1, 2` are the same call-site shape and print different
things: `f` is greedy and takes both arguments, while `g`'s prototype
cuts the extent to one and the `2` falls through to the enclosing
`print`. A parser that handled each case alone and got the pair wrong
would go green over two separate cases and fail here, which is the whole
reason an adjacency case exists.

`@greedy` and `@cut` stand the BUILTIN extent question beside the
user-sub one. `warn "a", "b"` is greedy and yields ONE value;
`warn("a"), "b"` is cut by the paren and yields TWO -- the `12` on the
third printed line. A user sub's extent is decided by its DECLARATION and
a builtin's by a PAREN at the call site, and a parser can implement
either and not the other. `local $SIG{__WARN__} = sub { }` is what makes
the pair observable at all: `warn` writes to STDERR, which the pinned
output does not read, and an installed handler is called INSTEAD of that
write, leaving `warn`'s RETURN VALUE as the only thing to count.

WHY THIS REFUSES, and why that is the corpus working rather than a
regression: it parsed clean while it held only parenthesised calls. The
two parenless call sites are what our parser declines, at the same
`trailing_tokens` site as the two extent cases -- it reads the callee as
a complete term and then finds a number with no operator between them.
Measured at dc1bea2c: three Unknown nodes, all `trailing_tokens`, where
each extent case alone produces one. The marker comes off when the
parenless form lands, and the extent cases' markers come off with it.

THIS IS ALSO THE TIER'S SHARPEST STATEMENT OF WHAT THE OP LINT CAN AND
CANNOT SEE. The source below contains a signature, a parameter default, a
loop, a `next`, three `return`s, an aliasing write through `$_[0]` and an
ampersand call, and the MAIN program's op stream contains exactly two ops
this tier introduces -- `anoncode` and `entersub` -- and not one op from
06_control either, because the loop is inside the sub too. Everything
else compiled into a CV that `-MO=Concise,-exec` does not print.
Measured with the sub bodies included, the same source adds seventeen
ops: `aelemfast and argcheck argdefelem argelem enteriter eq iter join
leaveloop leavesub lt multiconcat next postinc return rv2av unstack`.
That difference is the size of the blind spot, in one file.

```perl
use v5.36;
sub pick ($n, $label = "small") {
    return $label if $n < 10;
    for my $i (1 .. 2) {
        next if $i == 1;
        return "big-" . $i;
    }
    return "none";
}
sub bump { $_[0]++ }
sub answer { 42 }
sub f { return "f[" . join("-", @_) . "]" }
no feature "signatures";
sub g ($) { return "g[" . $_[0] . "]" }
use feature "signatures";
my $anon = sub { pick($_[0]) . "/" . &pick($_[1], "tiny") };
my $seen = 0;
bump($seen);
my $n = $seen;
lock($n);
local $SIG{__WARN__} = sub { };
my @greedy = (warn "a", "b");
my @cut = (warn("a"), "b");
print pick(1), " ", pick(50), " ", $anon->(2, 3), " ", $seen, "\n";
print &answer, " ", answer(), " ", answer, "\n";
print scalar @greedy, scalar @cut, "\n";
print f 1, 2;
print "\n";
print g 1, 2;
print "\n";
```

```behavior
parses: yes
refuses: 01a0dd6f-9891-7095-9b02-d382a25fbb98
refusal: trailing_tokens
```

```output
small big-2 small/tiny 1
42 42 42
12
f[1-2]
g[1]2
```
