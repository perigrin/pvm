# The mdtest topic format

A topic file holds ONE PERL PROGRAM PER CASE and several
implementations' answers about it. It is Chalk's mdtest format, which
this corpus adopted because the two were already the same idea with
different block names.

## The blocks

    ## <case title>          one case; prose under it says what it is for

    ```perl                  THE PROGRAM. Shared -- every implementation
    my $x = 1 + 2;           reading this corpus reads this block.
    ```

    ```behavior              what perl does with it, plus our record
    parses: yes              yes | no -- whether perl -c accepts it
    refuses: <issue id>      OPTIONAL: our parser refuses; names the issue
    refusal: <code>          OPTIONAL: which refusal code it produces
    ```

    ```output                what PERL PRINTS, byte-exact. The ground
    3                        truth: no implementation defines it and all
    ```                      of them are checked against it.

    ```tokens                the LEXICAL answer, which is this parser's.
    one numeric literal whose text is "1"
    ```                      GLOSSARY.md's vocabulary, not the lexer's.

    ```ir                    the GRAPH answer, which is B::SoN's. We do
    ...                      not fill it and do not validate it.
    ```

## The rules

**An unknown block tag is IGNORED.** B::SoN fills `ir`, we fill
`tokens`, and neither is the other's to check. A reader that rejected
what it did not own would force every implementation to implement every
other one's answer.

**`output` is a fenced block, not a behavior key.** It is byte-exact
stdout and is routinely several lines. An EMPTY block pins empty output;
an ABSENT block pins nothing at all. Those are different claims.

**A topic declares its tier** in a `**Tier NN name.**` line in its
intro prose. Declared rather than derived: the optimiser erases the very
construct a case is about, so the op-budget lint checks a declaration
rather than making one.

**A case may use only ops its tier or an earlier one introduces.** That
is what keeps the corpus graded, and `lintOps` enforces it.

**A `refuses:` record must still refuse.** A case that records a refusal
and now passes is reported, not silently skipped -- a stale record hides
a regression.

## What a case asserts, and what it must not

Four claims: perl compiles it (or does not), perl prints exactly this,
our parser agrees, and the token stream has these properties.

NO CASE ASSERTS TREE SHAPE. The plan is explicit -- "No corpus file
asserts CST shape, and none should" -- because that coupling makes the
corpus unusable by anyone whose tree differs. The `ir` block is not a
counterexample: it is a subset claim by a different implementation about
its own graph, and this reader never checks it.

## Token facts

Written in `conformance/GLOSSARY.md`'s vocabulary so a claim survives a
lexer refactor. Two forms only, `one` and `no`:

    one numeric literal whose text is ".5"
    no operator whose text is "."

A NEGATIVE IS A CLAIM ONLY IF IT CAN FAIL, and that needs two things:
the spelling must be something the lexer can emit, and it must be
REACHABLE FROM THIS SOURCE. `no operator whose text is "<="` over a
source containing no `<=` can never fail and asserts nothing.
`TestNegativeFactsNameTextTheSourceContains` enforces the second; 49 of
115 negatives failed it when it was written.
