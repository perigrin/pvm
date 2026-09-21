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
    # STATUS refuses as of 38c95d23. Issue 01a0c13f-97f5-7f98-b32d-07245ec6ddfe.
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

**A marker needs a blank line before it**, or another marker, or the start
of the file. Every corpus file was already written this way -- measured at
`dc1bea2c`, all 472 markers across all 134 files have one -- so the rule
costs nothing to follow and it is the only thing standing between the
format and a line of Perl.

The split is textual and a heredoc body, POD block or `__DATA__` section is
OPAQUE to it, so a line inside one beginning `--- ` was read as a real
marker. The two spellings failed differently and only one failed safely:

    my $t = <<'END';
    --- not a marker
    END

was rejected as an unknown section, loudly. But

    my $t = <<'END';
    --- expect output
    END
    print $t;

was accepted SILENTLY: the source truncated at that line, leaving an
unterminated heredoc opener, and the rest of the program became the pinned
output. A valid Perl program read as a differently-shaped corpus file with
nothing said. The blank line is what tells the two apart.

It is a weak check on purpose. A heredoc body may contain a blank line and
then `--- expect output` and still slip through. Closing the hole properly
would mean lexing the source to find its heredoc openers -- which is the
parser this corpus exists to test, so a corpus file could not be read while
that parser was broken. The rule makes the reachable accident loud and
leaves the rest to `13_opaque/README.md`, which records what the split
cannot see.

## The comment block

Read by people, except for one line. `# STATUS refuses` marks a file the
parser does not handle yet; the runner skips such a file and names the
refusal rather than failing. A file marked refusing that PASSES is an
error, because a stale marker is how a corpus stops measuring anything.

A refusal cites an issue id or, when the refusal was found by writing the
file, cites the file itself. The second is a complete citation rather than
a placeholder: the file already holds the source, the measured perl
behaviour and the tokens we produce instead, and an issue would be a
second copy of that, free to go stale.

**Cite the id WHOLE**, all thirty-six characters:

    # STATUS refuses as of 0ce515cb. Issue 01a0c13f-97f5-7f98-b32d-07245ec6ddfe.

The short `01a0c13f-97f5` form still parses, so files written before this
rule keep working, but it is not unique. These ids are UUIDv7 and their
leading 8 characters are a millisecond timestamp, so issues filed in one
batch collide there by construction -- measured across the 27 ids in
`docs/plans/2026-09-21-deferred-chain-m1-m2.md`, the 8-character prefix
gives 10 distinct values and the 13-character prefix 23. Uniqueness rests
on 16 bits of randomness. A citation that cannot be looked up
unambiguously cannot be verified, which is what the runner needs of it.

### Naming the refusal

A refusing file may also name WHICH of the parser's refusal sites it waits
on, with a `Refusal <code>.` clause on the same STATUS line:

    # STATUS refuses as of 9102578c. Issue 01a0c35e-.... Refusal missing_operand.

The codes are `parse.RefusalSites` in `internal/parse/refusal.go`, which
lists every one with what it means and where its site lives.

A code is a stable identifier, never a message. A message is prose and
changes when someone rewords it; a file that named one would break on an
edit that changed nothing about the parser. A code changes only when the
reason the parser declines changes -- which is exactly the event a
refusing file wants to be told about.

**A file that names a code and refuses with a different one FAILS**, the
same as a stale marker and for the same reason: it still skips on a claim
its own header no longer describes, so it has stopped measuring what it
documents. Without codes this was invisible -- every refusal read as
"1 Unknown node(s)", so a refusal that changed CAUSE while staying a
refusal looked identical to one that had not moved.

**The clause is optional.** A file that names no code promises nothing
about which site declines and keeps skipping as before, which is what the
files written before this rule do. A code is an additional promise, not a
new requirement.

## Three questions this format settles

**Trailing newlines: exactly one is stripped from `--- expect output`,
and the blank separator line is what supplies it.** Write

    --- expect output
    0.5

    --- expect tokens

The body is `0.5\n\n`, the strip removes the separator, and what remains is
`0.5\n` -- exactly the four bytes `print "$x\n"` emits. Omit the blank line
and the body is `0.5\n`, leaving `0.5`, which no such program produces; the
result is a `CORPUS BUG` report rather than a parse error, so it is worth
getting right.

**So `--- expect output` must never be the last section in a file.** At end
of file the blank separator line is trailing whitespace, and the repo's
`end-of-file-fixer` pre-commit hook strips it -- rewriting the file after it
is staged, so the commit carries an output one byte shorter than perl
prints and the author is blamed for a `CORPUS BUG` they did not write. The
format does not care about section order, so put any other section after it
and the blank line sits in the middle of the file where the hook has no
quarrel with it. `TestExpectOutputIsNeverTheLastSection` enforces this.

The reason for the rule, measured against perl 5.42.0: `print "0.5\n"`
emits four bytes and
`print "0.5\n\n"` emits five, so the two are genuinely different outputs
and the format has to be able to express both. Stripping one newline makes
the common case -- `print "$x\n"` -- read naturally, and a deliberate
trailing blank line is still written as two newlines and survives. Nothing
is unexpressible, so the convenience is free.

**A file that prints NOTHING says so with an empty `--- expect output`
section, which is not the same as having none.** Write the marker and the
blank separator line and nothing between them:

    --- expect output

    --- expect tokens

The body is `\n`, the strip removes the separator, and the pin is the empty
string -- a claim that perl prints zero bytes, checked like any other. A
file with NO `--- expect output` section makes no claim about output at
all, which is the normal case for a `parsent` file. The runner tells the
two apart by whether the section is present, never by whether the pinned
text is empty, so a construct whose whole point is that it prints nothing
can be pinned as precisely as one that prints `0.5\n`.

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

These are the corpus's only defence against a mis-lex, and the reason is
worth stating because it is not obvious.

Our lexer reads `5e-1` as three tokens -- `Number("5e")`, `Operator("-")`,
`Number("1")` -- and the parser then sees valid subtraction and produces no
Unknown node at all. Measured under 5.42.0, none of the other checks can
see it:

- **Behaviour cannot.** Both the correct reading and the subtraction
  evaluate to `0.5`, so the program prints the same thing either way.
- **The optree cannot.** `perl -e 'my $x = 5e - 1'` is a SYNTAX ERROR
  ("Bareword found where operator expected"), so perl never builds the
  wrong tree for B::Concise to be compared against. There is nothing to
  diff.
- **A round-trip cannot.** It reassembles the same source bytes whatever
  grouping produced them.

Only an assertion about the token stream distinguishes the two, which is
why this section exists and why it is written against the glossary rather
than against `lexer.Kind`.
