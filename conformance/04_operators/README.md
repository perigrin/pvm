# 04_operators

Arithmetic, string, comparison, logical, precedence, associativity, the
and/&& cliff.

## Why this tier sits here

An operator needs operands, and it needs to know what kind of thing its
operands are. Tier 01 supplies the operands and tier 02 supplies the
variables that hold them, but tier 03 is the one this tier cannot proceed
without: `@a + 1` is arithmetic on a count and `"$x" . @a` is not, and the
operator is what decides which. Context is not a thing operators consult, it
is a thing operators IMPOSE, so a corpus that introduced operators before
context would be asserting the imposition before it had named what was
imposed.

Check 2 asks whether this tier could move earlier. It could not. Move it
above 03 and every comparison file silently depends on an undeclared notion
of numeric versus string evaluation -- `$a == $b` and `$a eq $b` are
different ops, and the difference is exactly the context the operator forces
on its operands. Move it above 02 and there are no runtime operands at all,
which is worse than a documentation problem: with both sides constant the
optimiser folds the operator out of existence and the tier measures nothing.
That is not a hypothetical, it is the first thing measured here.

What the next tier needs from it is smaller than it looks. Tier 05's `my`
already appears four tiers back as a fixture. What 05 gets from 04 is that
an initialiser can be an expression rather than a literal, which is what
makes the scope of a declaration a question worth asking.

## DEPENDS ON

    03_context

Numeric and string comparison are the same relation over the same two
operands and differ only in the context each imposes. A tier that has not
named that distinction cannot say why `eq` and `==` are two ops.

`my`, `print` and `$ARGV[0]` appear in these files as FIXTURES. The first
two are tier 05's and tier 10's subjects, carried forward from tier 01 for
the same reason they were carried there -- an operator's result has to be
bound and observed. `$ARGV[0]` is tier 02's subject and is load-bearing in a
way the other two are not: it is the cheapest RUNTIME operand available, and
without a runtime operand this tier has nothing to measure. See below.

## INTRODUCES

    add and chr concat defined divide dor eq ge gt index le lt modulo multiply ncmp ne negate not or ord pow repeat scmp seq sge sgt sle slt sne sprintf substr subtract undef xor

## Why those ops, and not the ones the source implies

This list is what `perl -MO=Concise,-exec` EMITS for the files in this tier,
measured, not what reading them suggests. This is the tier where reading and
measuring diverge most, and five places where they do:

- **Every operator here needs a RUNTIME operand or it does not exist.**
  `my $x = 1+2` emits no `add`; it arrives as `const[IV 3] s/FOLD`. The
  optimiser erases the construct the tier is about, so a constant-folded
  arithmetic file and a tier-01 literals file produce identical op lists.
  Tier 01's README states this as a warning about a tier it does not
  contain. Here it is the operating constraint: every file puts
  `$ARGV[0]` on at least one side, and a file that forgets measures its own
  absence. This is why the ops LINT a declared tier and cannot derive one.

- **`concat` AND `multiconcat`, and which one you get is decided by the
  destination, not by the operator.** `my $c = $a . $b` compiles to
  `multiconcat` -- the same op tier 01 claims for interpolation, with no
  `concat` anywhere. The same `.` inside a list (`my @r = ($a . $b)`) or as
  a direct `print` argument compiles to `concat`. Only `concat` is new here;
  `multiconcat` is tier 01's, claimed there. Reading the source for `.` and
  expecting `concat` gets it wrong half the time.

- **`and` is the op for BOTH `&&` and `and`, and the cliff is invisible in
  the op names.** Measured: `my $c = ($a && $b)` and `my $c = ($a and $b)`
  produce byte-identical op streams, down to the order. `||` and `or` are
  likewise one `or` op, and `//` is `dor`. The difference the spec calls a
  cliff is a PARSE difference, and it shows up in the op stream only as a
  changed tree shape. Without the parentheses, `my $c = $a and $b` compiles
  the assignment INSIDE the left operand -- `padsv $a`, `padsv $c`,
  `sassign`, then `and` -- because `=` binds tighter than `and`. Behaviour
  confirms the shape: `my $c = $a && 9` with `$a` true prints 9,
  `my $c = $a and 9` prints 1.

  The `sassign` that appears there is tier 02's op, claimed with the package
  scalar. Its arrival here is still the most useful measurement in the tier,
  because an op showing up only when precedence goes the surprising way is
  what makes the cliff visible at all -- but the op is not new, only its
  reason for being present is.

- **No op for precedence or associativity at all.** `$a + $b * $c` and
  `($a + $b) * $c` emit the same four ops in the same order; only the tree
  differs, and `-exec` order does not show it. `$a ** $b ** $c` is right
  associative and emits two `pow` ops exactly as the left-associative
  `$a - $b - $c` emits two `subtract`. Grouping is what this tier is most
  about and it is the part the op list cannot see, which is why the
  precedence files carry behavioural probes built so the two groupings print
  different things -- `print 2**3**2` gives 512 and `print ((2**3)**2)`
  gives 64.

- **`defined` and `undef` are here because their parses are OPERATOR
  questions, and neither is about the variable it names.** Both were
  unnamed anywhere in the corpus while 29% and 19% of T1's files use
  them.

  `defined` occupies the NAMED UNARY precedence level, below arithmetic
  and above comparison, which is a fact about where its argument stops.
  Measured, `defined $x + 1` is `defined($x + 1)` and prints 1, where
  `(defined $x) + 1` prints 2 -- and the op names are the same both ways,
  `add` and `defined` in a different ORDER, which `opsOf` cannot see for
  the reason stated above about grouping. `09_named_unary.t` carries it.

  Two things about `defined` that a file written from its reputation
  rather than from the interpreter would get wrong, both measured under
  5.42.0. `defined %h` is a FATAL ERROR -- "Can't use 'defined(%hash)'"
  -- so the special rule for hashes was removed rather than being
  something to measure. And `defined &f` is real but is NOT this tier's:
  it compiles to `rv2cv ... /AMPER` feeding `defined`, and `rv2cv` is
  `08_references`'s op, so a file here spelling it would reach four
  tiers forward.

  `undef` is TWO OPERATORS wearing one word, and the split is ARITY.
  Measured, `undef @a` emits `<1> undef` and empties the array, while
  `@a = undef` emits `<0> undef` and leaves the array holding exactly one
  element. Opposite results, identical op NAME -- so the lint sees
  `undef` twice and cannot say that one of them took an operand.
  `10_undef_arity.t` carries it and records that our parser refuses the
  unary spelling alone, with the same `trailing_tokens` as tier 07's
  parenless-call files and for the same missing rule.

- **Five STRING operators are here, and the folding trap splits them
  three against two.** `chr`, `ord`, `index`, `sprintf` and `substr` are
  five of the 35 constructs `namedconstructs_test.go` measures T1 using
  and the corpus naming nowhere. They are operators by this tier's
  standard -- each takes operands and imposes a context on them -- and
  each has a parse question no other tier asks: where a named unary's
  argument stops, how many arguments a call takes, whether a call may sit
  on the left of an `=`.

  The trap this tier opens with decides which of them a file may write
  with constants, and it does NOT follow from anything in the source.
  Measured under 5.42.0:

      sprintf("%03d", 5)   const[PV "005"] s/FOLD     no sprintf op
      ord("A")             const[IV 65]   s/FOLD      no ord op
      chr(65)              const[PV "A"]  s/FOLD      no chr op
      index("hello","l")   const, const, index        SURVIVES
      substr("hello",1,3)  const, const, const, substr SURVIVES

  `chr` was expected to survive on the reasoning that its result depends
  on the encoding pragma in scope. It does not. `index` and `substr`
  survive and the other three do not, and no rule about the operators
  themselves predicts the split. Every file in this tier takes a runtime
  operand regardless, because what a file may RELY on is the constraint
  and the two exceptions are one release's behaviour.

  **`substr` is really THREE ops, and the corpus claims one of them.**
  Measured, `substr($s, 0, 1)` in RVALUE position compiles to
  `substr_left` -- a 5.42.0 optimisation applied at offset zero alone --
  while every nonzero offset, the four-argument form and the lvalue form
  all compile to plain `substr`. `substr_left` is claimed by no tier, so
  `14_substr_arity.t` keeps every offset nonzero and the most ordinary
  spelling of `substr` that anyone writes is the one this corpus cannot
  carry. `15_substr_lvalue.t` reaches offset zero by the two spellings
  that emit plain `substr`.

  **`substr`'s lvalue-ness is a FLAG, not an op.** `substr($s,0,1) = "J"`
  emits `substr[t5] vKS/REPL1ST,3` and NO `sassign` at all -- the
  replacement folds into the op and the assignment stops existing as a
  separate step. The four-argument `substr($s,0,1,"J")` emits
  `substr[t7] sK/4`, the same name, and edits the string identically.
  Only the RETURN value separates them: the four-argument form hands back
  the displaced text and the lvalue form discards it.

  **`sprintf`'s arity is decided inside a string.** `%*d` takes its width
  from the argument list, so one conversion consumes two arguments, and
  the format is one `string literal` to the lexer. The optree does not
  see it either: measured, `sprintf("%03d",$n)` and `sprintf("%*d",$n,$n)`
  both print `sK/2`, which is a private flag rather than an argument
  count, and the real operands are the ops between the `pushmark` and the
  `sprintf`.

  **`index` reports failure as `-1`.** Defined, numeric and TRUE, so
  `//` never fires on it and `if (index(...))` is true for a miss and
  false for a hit at position 0. It is the only builtin in this tier
  whose "not found" answer is an ordinary in-range value.

One structural note carried from tier 01, because it bites harder here: the
declared set is a UNION ACROSS THE TIER'S FILES and never a property of one
file. `padrange` absorbs `pushmark` when consecutive `my` declarations fuse,
so a file setting up several operands emits FEWER ops than one setting up
two. Counting ops per file and expecting them to accumulate gets the wrong
answer.

## What writing the files changed

All twenty-eight ops above survived: each is emitted by at least one file
here, so nothing was removed from INTRODUCES. Four corrections the measuring
forced, in the order they bit.

**`$ARGV[0]` is the runtime operand, and `shift` cannot be.** The README
named `$ARGV[0]` as the cheapest without saying what the alternatives cost.
Measured: `$ARGV[0] // 7` emits `aelemfast` and `dor`, both tier 02's or
this tier's. `$ENV{X} // 7` emits `multideref` instead -- also tier 02's, so
also legal, just one op more. `shift // 7` emits `gv`, `rv2av` AND an op
named `shift` that NO tier in the corpus claims, so a file using it fails
the lint outright. The cheapest operand is also the only one of the three
that is available at all.

**Every operand needs its `//` default, and not for tidiness.** With no
arguments `$ARGV[0]` is undef, and `$ARGV[0] + 5` warns on an uninitialised
value. The runner compares stdout byte for byte and a warning goes to
stderr, so the warning would not fail the run -- it would sit there
unnoticed, which is worse. The `//` is what makes the files silent, and it
is why `dor` is emitted by every file in the tier rather than only by
`05_logical.t`.

**False prints as the empty string, and that collides with a pre-commit
hook.** `print "eq ", ($a == $b), "\n"` with a false comparison emits
`eq \n` -- a line ending in a space. The repo's `trailing-whitespace` hook
strips that space out of the `--- expect output` block after it is staged,
so the committed expectation is one byte shorter than perl prints and the
next run reports a `CORPUS BUG` nobody wrote. This is the same class of
trap as the `--- expect output` last-section rule, from a different hook.
The comparison files bracket every result -- `eq [1]` -- so no expected line
ends in whitespace.

**The word-spelled operators are the tier's real finding, and it is a LEXER
finding rather than a parser one.** Perl spells some operators with letters:
`x`, `eq`, `ne`, `lt`, `gt`, `le`, `ge`, `cmp`, `and`, `or`, `not`, `xor`.
Our lexer gives all of them the kind `Word`, where the glossary calls them
operators. Five files here are `STATUS refuses` and four of the five refuse
on exactly that.

The sharpest instance is `06_and_cliff.t`. That file's whole subject is that
`&&` and `and` are the SAME op differing only in precedence -- and our lexer
hands the two halves out as different KINDS, `Operator("&&")` beside
`Word("and")`. Anything downstream deciding on kind alone treats one as an
operator and the other as a bareword. `05_logical.t` makes it unavoidable:
`xor` has no symbolic spelling in perl, so there is no way to write the
construct that dodges the gap.

The token claims are written in the glossary's vocabulary and left FAILING
rather than softened to match what we emit. A claim rewritten to match the
lexer stops measuring the lexer, which is the one thing `--- expect tokens`
exists to do -- see the corpus README on why `5e-1` needs a token assertion
at all.

The fifth refusal is `00_adjacency.t`, and it is the adjacency argument
working exactly as tier 01 predicted: three Unknown nodes over a body whose
every construct appears in a sibling file that parses. Only the mixture
refuses.

**A note for tier 03.** `reverse sort @nums` in the adjacency file emits
`sort` alone -- the optimiser folds the reverse into the sort's direction
flag and no `reverse` op survives. Tier 03 claims `reverse` and must emit it
from a file of its own; this tier cannot do it for them.

## What checking the tier changed

The four corrections above came from writing the files. These four came
from writing `internal/conformance/tier04_test.go` against them, which is
a different kind of pressure: the files say what they measure, and the
tests ask whether anything would notice if they stopped.

**`%nonassoc` is not a promise that repetition is rejected, and eight of
the eleven levels prove it.** `precedence.go` counts eleven `%nonassoc`
levels in `perly.y` and `AssocNone`'s comment says a binding power "stops
the recursion but still accepts the input", so the parser checks nonassoc
separately. Reading that as "repetition is an error at these eleven
levels" is an inference, and measured against 5.42.0 it is wrong three
times over:

    level  7 LSTOP    print print @a            -e syntax OK
    level 19 UNIOP    scalar scalar @a          -e syntax OK
    level 20 REQUIRE  require require           -e syntax OK
    level  2 LOOPEX   last LOOP last LOOP       syntax error
    level 11 DOTDOT   1 .. 2 .. 3               syntax error
    level 27 POSTINC  $a++ ++                   Can't modify postincrement

A conflict needs a left operand to fight over, and those three levels are
list operators and named unaries -- prefix forms whose second occurrence
is the first one's ARGUMENT. Five more levels carry no lexable operator
at all. Only level 11 is a nonassoc that means what the keyword suggests.

Level 27 is the sharpest of the three rejections, because it is not a
rejection by the grammar: perl PARSES `$a++ ++` and declines it at the
lvalue check. `expr.go` already excludes `++` and `--` from
`parseNonassoc` because they are postfix and take no right operand; this
measurement says that exclusion is right for a second reason.

**The op budget, not perl, is what bounds this tier's operator coverage.**
Of perly.y's 31 adjacent precedence pairs, nine have a pseudo-token or an
XS plugin hook on one side and no source reaches them. Sixteen more need
an operator whose op NO TIER in this corpus claims -- `..` is
flip/flop/range, `++` is preinc/postinc, `|` is `bit_or`, `<<` is
`left_shift`, `?:` is `cond_expr`, `=~` is `match`. A fixture for any of
those would fail the dependency lint, correctly. Six pairs are measurable
here, and they are the six where both levels' operators are ops this tier
itself claims: 4-5, 5-6, 12-13, 16-17, 22-23 and 25-26.

**Half the precedence fixtures written by eye measured nothing.** A
grouping fixture is only a measurement if its two forced readings print
DIFFERENT things; one whose readings agree exercises both operators and
would go on passing against a parser that grouped the other way. Three of
the six first attempts failed that, all three logical:

    $b or $a and $b     both groupings print 0
    not $y and $x       both groupings print 1
    $b || $a && $b      both groupings print 0

One truth assignment -- first operand true, other two false -- rescues all
three, found by trying all eight rather than by reasoning. The tier's own
`07_precedence.t` and `08_associativity.t` were checked the same way and
do discriminate: `14`/`20` and `512`/`64` differ under the two readings.

**`not` is TIGHTER than `and`, which is the opposite of how it reads.**
`not` is level 6 and `and` is level 5, so in `not $y and $z` the default
grouping is `(not $y) and $z`. Writing the pair the way the source reads
-- `not` as the outer word -- puts the two forced readings the wrong way
round, and the fixture then fails the precedence claim while passing the
discrimination one. This is the `and`/`&&` cliff from the other side: the
word-spelled operators are not where their spelling suggests.

## FILE ORDER

    accidental

The numbers are the order the files happened to be written in. This README
cites its files by name and never by position, and no claim in it would
become false if the files were renumbered.

`accidental` is a record of debt, not a convention. Renumbering this tier
into `derived` order costs nothing on disk but regenerates the ratchet,
which is a separate change; the declaration exists so the state is written
down rather than rediscovered by the next agent whose numbering test fails.
