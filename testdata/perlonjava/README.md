<!-- ABOUTME: Provenance, license and thanks for the in-tree copy of PerlOnJava's unit tests (T1). -->
<!-- ABOUTME: unit/ is upstream's src/test/resources/unit byte for byte; t1corpus.pin records the same revision. -->

# PerlOnJava's unit tests

## Thank you, Flavio

This directory holds the test suite of [PerlOnJava](https://github.com/fglock/PerlOnJava),
Flavio Glock's implementation of Perl 5 on the JVM. pvm did not write these
tests, and could not have measured its parser without them.

Thank you, Flavio, for all of this amazing work: for PerlOnJava itself, a
from-scratch Perl that compiles to JVM bytecode and that you have carried
through more than nine thousand commits since its first in January 2010, and
for this suite in particular.

What makes the suite so valuable to us is how it was built. Each top-level
file exercises one construct and is named for it: `anonymous_code_attribute_order.t`,
`array_autovivification.t`, `boolean_bitwise_xor.t`. When pvm's parser fails
one of them, the filename names the feature, and a failure becomes a bug report
rather than a hunt. And because PerlOnJava's own rules say a unit test must
encode standard Perl behaviour and be checked against perl before anyone relies
on it, these files describe Perl, not one implementation of it. That is exactly
what a from-scratch parser needs to be measured against.

These 986 files are T1, the corpus pvm's recursive-descent parser is graded
on. The T1 ratchets in `internal/parse` count against them, and a change
that makes one construct better while making the corpus worse is caught here
first. Our lexer cites them too (`nested_heredoc.t`, `string_interpolation.t`,
`postderef.t`, `glob_slot_hash_deref.t`), and on at least one hard question,
how to scan a quote whose delimiter is a backslash, PerlOnJava had it right
where two other implementations we compared did not.

The copy lives in this repository so every checkout and every CI run has it.
Any credit for what pvm's parser gets right on these files is shared with you.

## Provenance

| | |
|---|---|
| Upstream | https://github.com/fglock/PerlOnJava |
| Revision | `73be162b596cc948da61448e2a3caabb726017de` |
| Source path | `src/test/resources/unit/` |
| Copied | 2026-11-07 |
| Files | 986 top-level `*.t` (T1), 1,531 in all, with subdirectories and helper modules |

`unit/` is that tree whole and unmodified, byte for byte. pvm's tests read only
the top-level `*.t` files, but the copy is the complete upstream directory.
`internal/parse/testdata/t1corpus.pin` records the same revision, and
`TestT1CopyMatchesThePin` fails if the two disagree or the top level stops
holding 986 files.

The bytes are kept exact on purpose: some files have trailing whitespace or no
final newline, and the parse ratchets measure them as they are. So
`.pre-commit-config.yaml` excludes this directory from the whitespace and
end-of-file fixers, and `.gitattributes` marks it `-text`. Do not edit a file
here; a fix belongs upstream.

## License

PerlOnJava is copyright Flavio Glock, and is free software under the same
terms as Perl: the GNU General Public License, version 1 or later, or the
Artistic License, at your option. The three upstream license files ship here
unmodified:

- `LICENSE.md`, the statement of those terms;
- `Artistic`, the Artistic License;
- `Copying`, the GNU General Public License.

## Refreshing the copy

In one commit:

1. Replace `unit/` with `src/test/resources/unit/` from the new upstream
   revision, whole, and copy the three license files again.
2. Update the revision in `internal/parse/testdata/t1corpus.pin` and in this
   README.
3. Regenerate `internal/parse/testdata/t1.ratchet` against the new tree, as the
   pin's own comment says, along with the other T1 baselines beside it
   (`t1canon.ratchet`, `t1fidelity.ratchet`): a baseline measured across a skew
   in the corpus is an anecdote, not a result.

Then run `make test` with `T1_CORPUS` unset, so the tests read this copy.
