# The corpus file format

One construct per file. Every file is validated against perl before it
counts, because Perl 5 has no specification but its implementation.

Design: `docs/plans/2026-09-21-graded-conformance-corpus.md`.

## A file

    #!perl
    # TIER 01_literals
    # INTRODUCES numeric literal
    # USES my, print
    # MEASURED perl 5.42.0
    # STATUS refuses as of 38c95d23. Issue 01a0c13f-97f5.
    --- source
    my $x = .5;
    print "$x\n";
    --- expect parses
    --- expect output
    0.5
    --- expect tokens
    one numeric literal whose text is ".5"
    no operator whose text is "."

Everything before the first `--- ` marker is the comment block. Everything
after is sections.

## Sections

| section | meaning |
|---|---|
| `--- source` | the Perl the case runs. Required. |
| `--- expect parses` | the parser must accept it |
| `--- expect parsent` | the parser must refuse it |
| `--- expect output` | what perl prints, byte for byte |
| `--- expect tokens` | declared lexical claims, one per line |

**Exactly one of `expect parses` / `expect parsent` is required.** A file
asserting neither says nothing; a file asserting both is a contradiction.
Both are errors rather than something the reader resolves, because
resolving a contradiction by field order would make it invisible -- the
runner consults `parses` first, so a file claiming both would quietly
measure only half of what it says.

**Sections may appear in any order.** A repeated section is an error, not a
last-one-wins overwrite: two `--- expect output` blocks mean the author
believed both, and picking one silently discards a real disagreement.

**An unknown section name is an error.** A typo that parsed as "no
assertion" would make a file weaker without saying so.

## The comment block

Read by people, except for one line. `# STATUS refuses` marks a file the
parser does not handle yet; the runner skips such a file and names the
refusal rather than failing. A file marked refusing that PASSES is an
error, because a stale marker is how a corpus stops measuring anything.

A refusal cites an issue id (`Issue 01a0c13f-97f5`) or, when the refusal
was found by writing the file, cites the file itself. The second is a
complete citation rather than a placeholder: the file already holds the
source, the measured perl behaviour and the tokens we produce instead, and
an issue would be a second copy of that, free to go stale.

## Three questions this format settles

**Trailing newlines: exactly one is stripped from `--- expect output`.**
Measured against perl 5.42.0, `print "0.5\n"` emits four bytes and
`print "0.5\n\n"` emits five, so the two are genuinely different outputs
and the format has to be able to express both. Stripping one newline makes
the common case -- `print "$x\n"` -- read naturally, and a deliberate
trailing blank line is still written as two newlines and survives. Nothing
is unexpressible, so the convenience is free.

**`STATUS refuses` stays in the header rather than being derived.** The
runner does know whether a file refuses: it ran the parser. But a derived
status cannot disagree with reality, and disagreement is the event worth
catching -- a file that starts passing must fail loudly so the marker gets
removed. A header that can be wrong is what makes "it is no longer wrong"
observable. The cost is a field that can go stale; the runner's
stale-marker check is what pays it.

**`USES` is documentation and is not checked.** `INTRODUCES` is the tier
lint's input and is verified against the op stream. `USES` records what
else the file leans on so a reader knows why a tier-01 file mentions
`print`. Nothing verifies it, so it can rot -- that is accepted rather
than hidden, because the alternative is a second dependency declaration
that must agree with the first, and two lists that must agree are how
drift starts. If `USES` ever needs to be trusted, derive it from the
tokens rather than checking the comment.

## Token facts

Written in the vocabulary of `GLOSSARY.md`, not of our lexer, so a claim
survives a lexer refactor:

    one numeric literal whose text is ".5"
    no operator whose text is "."

These are the corpus's only defence against a mis-lex. Neither behaviour,
nor the optree, nor a round-trip can see one: `5e-1` mis-lexed as
`5e - 1` still prints `0.5`, perl never builds the wrong tree to compare
against, and a round-trip reassembles the same bytes however they were
grouped.
