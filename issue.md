---
title: BEGIN is in 99 T1 files and no corpus file, with the rest of compile phase
state: in-progress
urgency: normal
milestone: corpus-complete
created: 2026-09-23T02:07:02.60006476Z
updated: 2026-09-23T16:51:36.322538387Z
sessions:
- start_sha: 1afd0f341b23e124940e7ed14440f2338370dd79
  end_sha: ""
  commits: 0
  started_at: 2026-09-23T16:51:36.322538387Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T16:51:36.322538387Z
---

`BEGIN` appears in 99 of T1's 986 files and in none of the corpus's 179.
Nor does `END`, `__PACKAGE__`, `__FILE__`, `__LINE__`, `use constant`,
`use overload`, or `use parent`/`use base`. Measured: zero occurrences of
each across the corpus source.

These are the constructs where parsing and execution INTERLEAVE -- where
a construct changes what the rest of the parse sees. It is the
highest-frequency block in the whole T1 gap list and the corpus makes no
claim about any of it.

## What a parser gets wrong here, and what catches it

Measured, they emit almost nothing: `BEGIN { $x = 1 }`, `__PACKAGE__` and
`use constant PI => 3` all compile to a bare `const` and the statement
machinery. So the op lint cannot see them and OUTPUT can only see their
effect.

The claim is therefore mostly a TOKEN one, plus ordering:

- `BEGIN` is a bareword followed by a block with no semicolon and no
  parens -- structurally identical to a sub call with a hashref. The
  token fact separates the phaser from the call.
- `__PACKAGE__`, `__FILE__`, `__LINE__` are barewords that are NOT
  barewords; they are replaced at compile time.
- `use constant PI => 3` installs a bareword that afterwards parses as a
  zero-argument sub rather than a string. That is a construct changing
  the grammar for everything after it, and the only way to see it is to
  use `PI` later in the same file.

`# line NNN "file"` belongs here too: a `#` at column zero matching that
pattern is NOT a comment, and it rewrites `__LINE__` and `__FILE__`.
Measured, the directive makes `__LINE__` report 200 and `__FILE__` report
`bzzzt`. A lexer that treats every `#` as a comment gets every subsequent
error location wrong, silently.

## Files

Four, in `12_packages`, which owns `use` and compile-time loading:
`BEGIN`/`END` ordering pinned by output; the three compile-time tokens;
`use constant` with a later use of the installed bareword; and the
`# line` directive with `__LINE__` proving it took effect.

## Acceptance criteria

- [ ] Tier 12 holds a file whose output pins BEGIN running before the main program (`awk '/^--- source/{s=1;next} /^--- expect/{s=0} s' conformance/12_packages/*.t | grep -q 'BEGIN'`)
- [ ] A file pins that a # line directive rewrites __LINE__, which a comment-skipping lexer gets wrong (`awk '/^--- source/{s=1;next} /^--- expect/{s=0} s' conformance/12_packages/*.t | grep -q '__LINE__'`)
- [ ] Every new file is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierPackagesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierPackagesPerlValidated'`)
- [ ] The dependency lint passes over the tier (`go test ./internal/conformance/ -run '^TestTierPackagesLint$' -count=1 -v | grep -q '^--- PASS: TestTierPackagesLint'`)
- [ ] The tier's adjacency file holds every construct it introduces (`go test ./internal/conformance/ -run '^TestTierPackagesAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierPackagesAdjacency'`)
- [ ] The whole conformance suite is green with the new files in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
