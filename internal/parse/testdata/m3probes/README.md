# m3 probe corpus

Snippets of Perl, each with what the pinned perl (corpus.pin's interpreter,
5.44.0) does with it and what psc is expected to say about it.
`TestM3ProbesHeldToPerl` (`internal/parse/probes_test.go`) checks every
`*.txt` file here, and fails when the directory holds no snippet.

Every parser, `sites.go` or canon fix extends this corpus with its own
probe file, named in that fix's acceptance criteria. A finding a review
leaves only in prose cannot be rechecked; a row here is rechecked by every
later change.

## Format

```
Commentary before the first snippet is free text.

=== keys of a hash ref in scalar context
my $r = { a => 1 };
my $n = keys %$r;
--- facts
rv2hv 1 exact
anonhash 1 exact
faithful
# a comment, such as Deparse,-p output quoted for the reader
```

- `=== NAME` opens a snippet. Everything up to `--- facts` is the snippet's
  source, compiled as a file of its own.
- `--- facts` begins the snippet's facts, one per line, up to the next
  `===`. Blank lines and lines starting with `#` are ignored there.

### Fact lines

- `MARKER COUNT VERDICT`: perl builds COUNT ops of MARKER for the snippet,
  as `internal/parseoracle/testdata/parse_facts.pl` counts them. MARKER is
  one of `srefgen`, `rv2hv`, `match`, `readline` or `anonhash`. A marker
  the snippet does not list is one perl must report no site of. VERDICT is
  psc's expected bucket for that marker, through `parse.Sites` and
  `parseoracle.CompareFacts`:
  - `exact`: psc decides every site as perl does.
  - `wider`: psc hedges where perl decides. This is the only way a hedge
    counts as agreement, so a hedge never stands in silently for a
    decision. A fix that removes the hedge changes the row to `exact`.
- `faithful`: psc parses the snippet and `parse.Faithful` holds, so canon
  re-emits what the source says.
- `known-defect ISSUE-ID`: psc disagrees with the row's verdicts or its
  `faithful` fact, and the git-zhi issue with that full id tracks it.
  perl's counts are still checked. The row fails once psc agrees, so
  the fix that closes the issue drops the line.

Every snippet must compile under the pinned perl. `Deparse,-p` output may
be quoted in a comment, but nothing compares against it.
