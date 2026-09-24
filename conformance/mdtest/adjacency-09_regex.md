# Every regex construct, each beside another

One body holding every construct this tier introduces, each adjacent to
another, and every delimiter form the tier teaches.

**Tier 09 regex.** Introduces nothing of its own; it is the mixture that
is the subject. Depends on 01_literals.

The tier's other cases are one construct each, which is what makes them
diagnosable. That same property is why a corpus of such cases cannot
reach an ADJACENCY bug -- and in a tier whose operand is not Perl, the
adjacency bug is the one to expect. A lexer that delimits a pattern by
scanning for the next `/` is correct on every case here taken alone and
wrong the moment an `s{a}{z}` sits between two matches.

## The whole tier in one body

The constructs sit consecutively: a constant match, a negated match
through a bracketing delimiter, a match through a non-bracketing one, a
match through the comment character, a match whose pattern NESTS its own
delimiter, an interpolated match, a `qr//`, four substitutions delimited
by brackets, slashes, hashes and nested brackets, a `tr///`, a `split`
on a pattern, and a `//g` match read through `pos` -- then one print
that reads every result.

THE LAST THREE ARE WHERE THE ADJACENCY CLAIM EARNS ITS KEEP a second
time. `tr/./Z/` is a second two-region quote-like, and a lexer whose
table says only `s` takes two regions mis-terminates it and then
mis-terminates everything after it -- a failure a one-construct-per-case
`tr` test cannot reach, because there is nothing after it to break.
`split / /` puts a pattern in an ARGUMENT SLOT immediately following
that, which is the position a lexer recovering from a botched `tr` is
most likely to read as division. The pattern is a single space, the
least distinguishable body a slash-delimited pattern can have.

EVERY DELIMITER FORM THE TIER TEACHES APPEARS HERE, which is the part
that needed fixing. The body shipped holding `m{}`, `s{}{}` and `qr//`
alone -- so the one bug it exists to reach was unreachable for every
form most likely to produce it.

Three of the added forms are the ones that cannot NEST. `m!abc!`,
`s/c/y/` and `s#d#w#` each close with the character they opened with, so
a lexer has no bracket depth to count and must simply stop at the next
occurrence. The fourth is the opposite case and is here for the
contrast: `m{a{b}c}` and `s{a{b}c}{ok}` carry their own opening brace
inside the pattern, where a lexer MUST count depth. That pair of rules
is what GLOSSARY.md records, and neither `m{zzz}` nor `s{a}{z}`
exercises either half.

The hash forms are the sharpest pair. `#` is Perl's comment character,
so a lexer that has not yet recognised the `m` or `s` has already
discarded the rest of the line. Putting them BETWEEN other delimiter
forms rather than alone is the adjacency claim in its strongest version:
recovering from `m#abc#` is not enough if the `s#d#w#` three statements
later is then read as a comment.

The tier's dependency on 01_literals is present rather than decorative:
every pattern here is matched against a string literal bound to a pad
slot, and the final print interpolates every result into one
double-quoted string. Pairing with 08_references instead would assert
nothing.

The substitutions run LAST on purpose. They mutate `$s`, which the five
matches above them read; running any earlier would make those results
depend on statement order in a way that hides a mis-parse behind a
plausible-looking output. Each of the three that touch `$s` mutates a
DIFFERENT character -- `a`, `c` and `d` -- so the final `zbyw` records
that all three ran, where two substitutions of one character would leave
the second's failure invisible. The nested substitution needs its own
target, `$n`, because its pattern is five characters `$s` does not hold.

```perl
my $s = "abcd";
my $p = "b";
my $hit = $s =~ /abc/;
my $miss = $s !~ m{zzz};
my $bang = $s =~ m!abc!;
my $hash = $s =~ m#abc#;
my $nest = "a{b}c" =~ m{a{b}c};
my $interp = $s =~ /$p/;
my $re = qr/abc/;
$s =~ s{a}{z};
$s =~ s/c/y/;
$s =~ s#d#w#;
my $n = "a{b}c";
$n =~ s{a{b}c}{ok};
my $tr = "a.c";
my $cnt = ($tr =~ tr/./Z/);
my @f = split / /, "p q";
my $pos = "abcabc";
$pos =~ m/b/g;
my $at = pos($pos);
print "$hit $miss $bang $hash $nest $interp $re $s $n $tr $cnt @f $at\n";
```

```behavior
parses: yes
```

```output
1 1 1 1 1 1 (?^:abc) zbyw ok aZc 1 p q 2
```

```tokens
one quote-like operator whose text is "m{zzz}"
one quote-like operator whose text is "m!abc!"
one quote-like operator whose text is "m#abc#"
one quote-like operator whose text is "m{a{b}c}"
one quote-like operator whose text is "s{a}{z}"
one quote-like operator whose text is "s/c/y/"
one quote-like operator whose text is "s#d#w#"
one quote-like operator whose text is "s{a{b}c}{ok}"
one quote-like operator whose text is "qr/abc/"
one quote-like operator whose text is "tr/./Z/"
one quote-like operator whose text is "m/b/g"
```
