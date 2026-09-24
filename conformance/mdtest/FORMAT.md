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

**A CASE IS ONE COMPILATION UNIT.** Exactly one ```perl block; a second
is an error, not an overwrite and not a merge.

Ty's mdtest (astral-sh/ruff, `crates/ty_test`) merges consecutive
unnamed code blocks into one file. This format deliberately does not,
because PERL'S COMPILATION UNIT IS THE FILE: `my` scope, BEGIN
ordering, `use strict`'s lexical effect, `__DATA__` and constant
folding are all per unit. Measured:

    $ perl -MO=Concise -e 'my $x = "abc";'
      padsv_store[$x] <- const[PV "abc"]        folded into the slot

    $ perl -MO=Concise -e 'my $x = "abc" . $0;'
      multiconcat("abc",3,-1)[$x] <- gvsv[*0]   runtime, and it survives

(B::SoN reports the same split in its own vocabulary: the first emits no
VarDecl node, the second keeps one.)

Those are two fixtures a reader expects to be independent. Merged into
one unit they can fold across the boundary, so what you read is not
what perl compiled.

Rejecting rather than overwriting is what makes this ENFORCED rather
than remembered. It also makes one-construct-per-case a property of the
format -- the property that makes PerlOnJava's 986-file corpus useful
for ranking -- rather than a convention an author has to hold.

**`output` is a fenced block, not a behavior key.** It is byte-exact
stdout and is routinely several lines. An EMPTY block pins empty output;
an ABSENT block pins nothing at all. Those are different claims.

**A topic declares its tier** in a `**Tier NN name.**` line in its
intro prose. Declared rather than derived: the optimiser erases the very
construct a case is about, so the op-budget lint checks a declaration
rather than making one.

**A case may use only ops its tier or an earlier one introduces.** That
is what keeps the corpus graded, and `lintOps` enforces it.

**A NON-PARSING CASE CAN ASSERT WHAT A PARSING ONE CANNOT.** `parses:
no` means perl rejects the program, so perl builds no optree for it and
it emits NO OPS -- which means the op budget cannot bind it.

That is a structural capability, not a completeness argument. Tier 04's
`1 .. 2 .. 3` spells `..`, whose ops belong to 03_context; a PARSING
case spelling it in tier 04 would fail the lint, and this one cannot.
So the two kinds of case assert disjoint sets of facts rather than
being harder and easier versions of the same thing.

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

## Info strings: a path if it has a dot, otherwise a language tag

    ```perl          the unnamed entry point. Always legal.
    ```fileA.pm      a named file, when blocks belong together.

The disambiguation is the dot, and it needs no registry: `perl` has no
dot and no path we would write lacks one.

BARE `perl` STAYS FIRST-CLASS rather than becoming shorthand. Almost
every case in this corpus is one file, and a path is a CLAIM -- writing
`foo.pm` says this compiles as a module, needs a trailing `1;`, and
resolves through `@INC`. A case asserting none of that should not have
to pick a name implying it, so the ABSENCE of a path is itself
information: one program, and nothing depends on where it lives.

Multi-file cases are not used here yet. Settled with the B::SoN session
so the two corpora agree before either needs it.

## Why blocks, and not inline assertions

Ty puts its assertions inline -- `reveal_type(x)` with a trailing
`# revealed: <type>`, and `# error: [rule-code]` at the line that
raises.

THE REASON THIS FORMAT USES BLOCKS IS THAT THEY ORGANISE FACTS BY KIND.
What perl prints, whether perl compiles it, what the token stream
contains: three kinds of claim about one program, and a block each
keeps them separable, greppable, and machine-written by the tool that
measures them.

THE TRADEOFF IS LOCALITY. An inline assertion sits AT the expression it
is about; a block sits in a list that must reference a location some
other way. Every claim in this corpus is about a WHOLE PROGRAM -- what
it prints, whether it parses, what its tokens are -- so nothing here
pays that cost yet. A per-expression claim (a type at a site, a
diagnostic at a column) would, and that is the trigger to revisit.

A WRONG REASON THAT WAS HERE, recorded because a spec outlives the
decision it justified and the next reader reads the reason. An earlier
version argued that several implementations "answer independently", so
an inline comment would have to encode WHICH implementation it binds
and a block therefore "says silence structurally". Both halves are
wrong. This is a CONFORMANCE SUITE: the fixture states what is true
about the program, and each implementation checks itself against the
parts it supports -- one set of agreed facts, not several namespaced
answers. And silence needs no structure either way: an implementation
with no type layer skips a type assertion whether it sits in a block or
in a comment.

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
