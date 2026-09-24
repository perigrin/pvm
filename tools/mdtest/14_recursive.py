"""Port tier 14_recursive to the mdtest topic format."""
import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from gen import topic, check

TIER = '14_recursive'
covered = set()

covered |= topic(TIER, 'eval-replacement.md', 'The replacement that is a program', """
`s///e`, `s///ee` and `eval STRING` -- the three places where a region of
source is handed back to the compiler rather than read as text.

**Tier 14 recursive.** Introduces `entereval`, `substcont`. Depends on
09_regex.

`substcont` is what `/e` actually adds. Plain `s/a/z/` is one `subst` op
taking a constant; `s/a/$n+1/e` is `subst(replstart->)` with the
replacement's own ops -- `padsv`, `const`, `add` -- executed between it and
a `substcont` that loops back for the next match. `substcont` is the
re-entry made visible: the op that returns control to the substitution
after the embedded program has run.

`entereval` is what the SECOND `e` adds, and what a bare `eval STRING`
is. It is the only op in this corpus that compiles Perl the compiler
never saw.

Every replacement here is deliberately unfoldable. Measured,
`s/a/uc("z")/e` compiles to `const[PV "Z"] s/FOLD` and a plain `subst` --
the optimiser erases the `/e` completely, and the case would measure the
same thing as tier 09's plain substitution.
""", [
    ('01_subst_eval.t', 'The `/e` flag: an expression in the replacement half', """
`/e` makes the replacement a Perl EXPRESSION rather than a string: the
lexer must hand the region between the second and third delimiter back to
the parser.

The token claims record where the re-entry has NOT yet happened. At the
lexical layer `s/a/$n+1/e` is ONE quote-like operator, the same as tier
09's `s/a/z/`: the `+` between its delimiters is inside its text, not a
token beside it. A lexer that emitted that `+` as an operator would have
lexed the replacement instead of delimiting it, which is the error this
tier exists to catch -- the region is handed to the PARSER, which lexes it
in a second pass.

`$n` is declared on its own line and so does appear as a variable token
there; only the `+` is unique to the replacement, which is why it is the
one the `no` claim names.
"""),

    ('02_subst_double_eval.t', 'A second `e`: the replacement re-read as Perl', """
A second `e` evaluates the FIRST evaluation's result as Perl: the string
`3*4` is compiled at runtime by code the compiler never saw.

This is the deepest re-entry the corpus contains. `s/b/$c/ee` compiles to
the same `subst(replstart->)` / `substcont` frame as `/e`, with
`entereval` between them: the first `e` evaluates `$c` to the string
`3*4`, the second compiles and runs that string.

The distinction matters lexically because NEITHER `e` is visible to the
lexer as code. The replacement region holds `$c`; the Perl that eventually
runs is `3*4`, which appears in the source only as the contents of a `q{}`
three lines earlier. A lexer cannot reach it at all, and neither can the
optree -- `entereval` is where the source ends.
"""),

    ('06_eval_string.t', '`eval STRING`: the re-entry with no delimiter', """
Every other case in this tier hands back a region the lexer at least had
to FIND: `s///e` has three delimiters, `(?{ })` has a brace pair inside a
pattern, `qr//` freezes one. This has none. The operand is an ordinary
string expression, and what it holds is not known until the expression has
been evaluated.

WHY HERE AND NOT WITH THE BLOCK FORM, since `eval` is one keyword: the two
forms share no op. Measured under 5.42.0, `eval { 1 }` compiles to
`entertry` / `leavetry` -- a control transfer that marks a frame to unwind
to and compiles nothing new, claimed in `06_control`. `eval "1"` compiles
to `entereval`, which compiles Perl the compiler never saw.

Measured, with X unset so the interpolated operand is a runtime value:

    $ perl -MO=Concise,-exec -e 'my $n = $ENV{X} // 2;
      my $r = eval "$n + 1"; print "$r\\n";'
    8  <0> padsv[$n:1,3] s
    9  <+> multiconcat(" + 1",-1,4)[t5] sK/STRINGIFY
    a  <1> entereval[t256] sK/1

THERE IS NO `add` IN THAT STREAM, and that absence is the whole case. The
program adds two numbers and emits no addition op, because the `+` is a
character in a string at compile time and becomes an operator only inside
the second compilation `entereval` triggers. `multiconcat` BUILDS the
program text; `entereval` compiles and runs it. One `e` puts the region's
`add` in the outer stream; two `e`s and a bare `eval` do not.

The string is interpolated rather than constant on purpose. `eval "1 + 1"`
still emits `entereval`, but a parser could constant-fold the operand and
the case would no longer measure that the operand is a VALUE.

The output pins `3`, not `2 + 1`. A parser that treated the string as an
ordinary string expression -- never re-entering -- would assign the text
and print it verbatim: a one-token difference in the tree and a five-byte
difference in the output.
"""),
])

covered |= topic(TIER, 'embedded-code.md', 'Perl inside a pattern', """
`(?{ })` and a `qr//` that carries one -- a statement written inside a
regex, and a value that carries that statement across statements.

**Tier 14 recursive.** Introduces `entereval`, `substcont`. Depends on
09_regex.

NEITHER CONSTRUCT EMITS AN OP OF ITS OWN, and that is the tier's main
finding. Measured, `$s =~ /a(?{ $n = 1 })b/` compiles to one `match` op
whose pattern is the string `"a(?{ $n = 1 })b"`; the embedded statement is
inside the op's own data, not in the op stream. The full tree does show the
code's optree hanging off the match, but every op in it is marked `-`,
optimised out of the execution path: it lives in a separate CV the regex
engine calls, and `-exec` walks the outer path only. `qr//` with embedded
code measures the same -- one `qr` op, pattern string carrying the block,
indistinguishable from plain `qr//` by anything but its text.

So the construct that most plainly re-enters Perl is INVISIBLE to a
measurement of ops. These cases carry their weight in `tokens`, not in
`INTRODUCES`.

Neither case needs `use re 'eval'` or a warnings pragma. Measured under
5.42.0, a LITERAL code block in a pattern compiles and runs clean under
`use strict; use warnings`; the pragma is required only when the pattern
itself is interpolated from a variable, which neither case does.
""", [
    ('03_code_block.t', '`(?{ })`: a statement inside a pattern', """
The block runs when the engine reaches that point in the match, and the
tokens are the only place this case can make a claim -- the block is
INSIDE the match's one token, not a brace and statements beside it.

The `no` claims name `5` and `/` deliberately. `$k` appears as a real
variable token on its own declaration line, so a bare "no variable" claim
would be false -- but the `5` inside the block appears NOWHERE else, so a
numeric literal with that text can only come from a lexer that read into
the pattern. That is the claim worth making.
"""),

    ('04_qr_with_code.t', '`qr//` carrying a block: code frozen into a value', """
A `qr//` carrying a code block freezes the embedded program into a value:
the code travels with the compiled pattern and runs wherever the object is
later matched.

Tier 09 claims plain `qr//` and says it claims only the non-recursive
form, leaving this tier "only what embedding adds". Measured, what
embedding adds to the `qr` op itself is NOTHING. What changes is
downstream, and it is still tier 09's op: matching against the compiled
object emits `regcomp` then `match`, where matching a literal pattern
emits `match` alone -- which tier 09 already documents as the cost of
interpolation, not of embedding.

So this case's whole subject is that a value can CARRY Perl code across
statements. The assignment and the match are separated on purpose: the
block is written on line 2 and runs on line 4, and `$n` proves it.

The `qr` object is not printed. Its stringification would embed the
block's source, which is deterministic, but printing `$n` is the claim
this case is actually making.
"""),

    ('05_hostile_contents.t', 'Hostile contents: proving the region is parsed, not counted', """
The tier's other cases hold INERT regions. `$n+1`, `$k = 5` and `$n = 7`
are delimited identically by a lexer that counts braces and by one that
parses Perl, so those cases establish that the region is DELIMITED and
leave the RE-ENTRY half of the tier's thesis unasserted. This case is the
inverse of `13_opaque`'s data-not-Perl case: there the hostile content
proved a region was NOT lexed, here it proves a region IS.

MEASURED, AND THE TWO CONSTRUCTS DIFFER. Under 5.42.0 perl uses two
different strategies to find the end of a re-entrant region, and the
difference is observable:

- `(?{ ... })` is PARSED AS PERL. Measured, `/b(?{ $k = length("}}}") })/`
  compiles and prints 3 -- the three braces inside the string do NOT close
  the block. A brace counter would stop at the first and hand the parser
  `(?{ $k = length("}`.
- `s{a}{ ... }e` COUNTS DELIMITERS. Measured, `s{a}{ $n + length("}}") }e`
  is a syntax error in perl itself -- "Unmatched right curly bracket at -e
  line 1" -- because the brace in the string DOES close the replacement.
  The re-entry happens after the region is cut out, not while it is being
  found.

So the substitution's hostile content is PARENS rather than braces:
`s{a}{ $n + length("))") }e` compiles and prints 3bc, and a reader that
balanced some bracket other than its own delimiter would truncate it. That
is the strongest falsifiable claim the construct admits, and writing braces
there instead would pin a perl syntax error as though it were our bug.

The braced `s{}{}e` spelling rather than `s///e` is forced by the same
measurement from the other side: with a `/` delimiter, a `/` inside a
string in the replacement is a syntax error in perl too. A delimiter that
cannot appear inside the region leaves no hostile content to write.

Every replacement here is unfoldable: `$n + length("))")` keeps
`substcont`, where a constant replacement would fold to `const s/FOLD` and
erase the `/e` entirely. Measured, `length("))")` itself folds to
`const[IV 2]`, but the `add` around it does not, so the frame survives.

The token claims are where this case carries its weight, because the
optree cannot see the distinction at all: the `(?{ })` and the `qr//`
block are inside their ops' PATTERN STRINGS, one `match` and one `qr`. The
`no` claims name the hostile characters -- a lexer that stopped early would
have spilled the region's tail out as ordinary tokens, and `length` would
appear as a word beside the quote rather than inside it.
"""),
])

covered |= topic(TIER, 'adjacency-14_recursive.md', 'Every re-entrant construct, each beside another', """
One body holding every construct this tier introduces, each adjacent to
another -- plus the deferred form `(??{ })`, which appears only here.

**Tier 14 recursive.** Introduces nothing of its own; it is the mixture
that is the subject. Depends on 09_regex.

The tier's other cases are one construct each, which is what makes them
diagnosable. That same property is why they cannot reach the bug this tier
is most exposed to. A lexer that handles re-entry by special-casing ONE
level -- remember you are inside a replacement, lex to the closing
delimiter, hand it back -- passes every isolated case and still has no
stack. Four re-entrant constructs in one statement sequence is what asks
whether the mechanism nests or merely remembers.

THE PAIRING WITH THE EARLIER TIER is with `09_regex`, this tier's declared
dependency: `qr/c/` is tier 09's plain compiled pattern, and the
`(??{ $i })` beside it is this tier's deferred re-entry interpolating that
object mid-match. Adjacent, in one pattern, which is where a lexer that
delimits `(?{` by counting braces rather than by parsing would find the
extra `?` and stop.

`(??{ })` -- the DEFERRED form -- appears here and nowhere else, because it
introduces no op and no token category the other cases do not already
cover. Measured, `/a(??{ $inner })/` is one `match` op whose pattern string
holds the block, the same shape as `(?{ })`: the difference between the two
is WHEN the engine runs the code and what it does with the result --
`(?{ })` runs it for effect, `(??{ })` uses its return value as a pattern
to match right there. That distinction is entirely inside the regex engine
and invisible to both the optree and the token stream, so it earns a place
in the adjacency case rather than a case of its own.

No `use re 'eval'` and no warnings suppression anywhere here. Measured,
LITERAL code blocks compile and run clean under `use strict; use
warnings`; the pragma is required only when the pattern itself is
interpolated from a variable. `(??{ $i })` interpolates a variable INTO the
block, not the block into the pattern, which is why it too is exempt.
""", [
    ('00_adjacency.t', 'The whole tier in one body', """
The string walks `abc` -> `2bc` (the `/e` replacement evaluates `$n+1`) ->
`212c` (the `/ee` replacement evaluates `$c` to `3*4`, then evaluates THAT
to `12`). `$k` is 5 because the `(?{ })` block ran when the engine reached
it, and `hit` prints because the deferred block returned the compiled
`qr/c/` and the engine matched it.

The `no` claims are what the mixture buys. `+` and `5` are each unique to
a re-entrant region here, so either one surfacing as an outer token means
a lexer read into a region it was supposed to delimit -- and with four
such regions in sequence, it names which layer of re-entry lost its place.
"""),
])

check(TIER, covered)
