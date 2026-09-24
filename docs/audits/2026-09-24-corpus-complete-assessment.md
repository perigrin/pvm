# corpus-complete: assessment

**Date:** 2026-09-24
**Milestone:** `corpus-complete`, 13/13 done
**Range:** `cbaacf10..7b6014f5`, 17 commits
**Suite:** green (`make test`)

## What it set out to do

The milestone's body states the target: close six clusters where the
corpus had LITERALLY ZERO coverage, plus one structural hole (no file
pinned a version-gated feature in its DISABLED state). It named 25-35
new corpus files as the size.

The five-axis measurement behind it, verified directly under perl
5.42.0:

    perlsyn          31 /  68 forms        46%
    perlop           41 /  97 operators    42%   12 of 24 levels at ZERO
    perldata         41 / 150 forms        27%
    perlsub          19 / 180 forms        11%
    modern idiom     11 /  34 features     32%
    T1              134 / 282 constructs   48%

## What it delivered

**27 corpus files**, inside the stated range. Measured from the
ratchet's own header:

    at cbaacf10^   185 files, 164 clean, 21 refusing
    at HEAD        212 files, 188 clean, 24 refusing

Every cluster the milestone named is now non-zero: compound assignment
(13 operators), the bitwise/shift band (levels 9/14/15), the range and
flip-flop trio, the compile phase, the symbol table, and the unary
slice. Pattern internals were ruled OUT of scope by measurement rather
than left unaddressed -- a pattern is one token, so no token fact can
reach inside it, and the plan forbids asserting tree shape.

The structural hole is closed: six keywords now have a file measuring
them with their feature OFF, and four of those found real parser gaps.

## What it found that nobody was looking for

This is the part worth recording, because it is larger than the
delivery.

**Forty-nine of the corpus's 115 negative token facts asserted
nothing.** A `no X whose text is "Y"` fact can only fail if `Y` appears
in the source; 49 named text their source never contained. Three
separate rationales produced them, each plausible and each mistaking
what the PARSER does for what the LEXER does -- most instructively a
passage about `->` pasted into four files, derived from `B::Deparse`
output and written as if it were a claim about tokens.

`TestNegativeFactsNameTextTheSourceContains` now makes the rule
mechanical. Negatives went 110 to 63.

**A parser bug costing far more than the case that exposed it.**
`startsTerm` answered true for every `lexer.Word`, so word-spelled
operators (`eq`, `cmp`, `lt`, `x`, `and`) read as term-starters and
both filehandle branches took the slot on `print FOO eq "x"`. Fixing it
moved 14 T1 files and took `base/num.t` from 44 refusals to 6; T2 clean
went 17 of 56 to 19.

## Where the process failed, and what caught it

The gates caught more than the work did, and the same failure recurred:

- **A vacuous fact shipped in the commit that fixed two others.** I
  applied the vacuity test at the operator-table level and not at the
  source level. Seventh instance of the class.
- **A fix repeated the bug it diagnosed.** The finding was a passage
  pasted into four files without checking whether it applied; the
  replacement was pasted into four files without checking whether it
  applied, and was false in two.
- **A four-name exception list was wrong on one entry and missing
  another.** `__CLASS__` is a value only under `use feature 'class'`
  and a bareword filehandle otherwise, so exempting it unconditionally
  contradicted the corpus file edited in the same commit -- which then
  passed for the wrong reason. The root-cause fix (ask what FOLLOWS the
  word, which is perl's own rule) needs no list.
- **A pairing exemption I built to keep ten facts was a
  rationalization.** Deleting `->` from the lexer's operator table
  fails only the POSITIVE facts; not one negative fails. I had the tell
  before the gate ran -- a sibling file did the same job with no such
  fact -- and argued past it.

Three of those four were caught by a PAAD gate rather than by me.

## Open gaps this milestone discovered

Each has an issue; none is silent.

    01a0ce57   the lexer does not form `x=`
    01a0cf64   `-e` lexes as a minus and a word, and the file-test family with it
    01a0d087   ungated `defer` is declined as a statement form
    01a0d0b0   `#line` is lexed as a comment, and its corpus file passes anyway
    01a0d0b1   `print {$fh} "x"` refuses -- the disambiguating form
    01a0d0c7   five refusal citations point at closed issues, and nothing checks

`01a0d0b0` is the sharpest: a file that PASSES while the behaviour it
documents is unimplemented. Its header now says so, but a warning in a
header is not a check.

`01a0d0c7` is the same decay this milestone fixed for four files
(`d95cf4ee`) and never checked corpus-wide. Three cited issues closed
because their MILESTONE finished, not because anything was fixed.

## Deferred, and where to

The plan names two deliberate deferrals and both belong to the NEXT
milestone rather than this one:

- The `t/` sweep, which runs after the corpus exists.
- Adjacency beyond the adjacent pair (tier N with N-3), which the sweep
  is what finds.

Neither is silent and neither was quietly dropped.

## Honest status

The corpus is complete against what this milestone measured, and the
measurement is now a thing the suite enforces rather than prose. It is
NOT complete against Perl: the five-axis numbers above have not been
re-run, and re-running them is the honest next measurement rather than
a claim that can be made from here.

What changed most is not the file count. It is that the corpus's own
claims are checked: 135 Go tests over `conformance/`, 60 of them tier
checks, asserting vacuous facts cannot ship, refusal markers cannot go
stale in the passing direction, no construct is undeclared, and every
token fact speaks the glossary's vocabulary rather than the lexer's.

The gap that remains is the one `01a0d0c7` names: a marker whose ISSUE
closed is not caught, only a marker whose file starts passing. The
corpus checks what it claims about Perl far better than it checks what
it claims about itself.

## Written before the format port, and what it changed

This assessment described a corpus of 212 `.t` files, one construct
each. That corpus no longer exists: it was ported to 65 markdown topic
files in Chalk's mdtest format and the `.t` files were deleted
(`a6b33a2a`). Everything above about what the corpus CLAIMS still
holds -- the port moved claims, it did not change them, and the
equivalence was measured rather than asserted: 212 cases against 212
files, sources byte-identical, 24 recorded refusals on both sides.

Three numbers in this document moved and are corrected above or here:

  - 158 Go tests became 135, and 84 tier checks became 60. TWENTY-FOUR
    TESTS AND TEN HAND-MAINTAINED MAPS WERE DELETED, 2,657 lines, all
    of which derived a construct's identity from a FILENAME and then
    grepped the adjacency body for the spelling that identity mapped
    to. That bridge is only needed by a one-construct-per-file corpus.
  - "every tier holds an adjacency file" is now a topic rather than a
    file, and the check that asserted it was one of the twenty-four.
    The adjacency CASE survives and is still validated against perl,
    checked against the parser, and linted for its op budget; what went
    is the machinery asserting that each construct appears in it, which
    a reader now answers by reading the body.

The port also found six false prose claims in the corpus, three of
which had been in `.t` headers all along and were only noticed because
porting them meant re-measuring them. That is the same finding this
assessment records under "the corpus was lying to itself at scale",
reached by a second route.
