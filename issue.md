---
title: BEGIN is in 99 T1 files and no corpus file, with the rest of compile phase
state: done
urgency: normal
milestone: corpus-complete
created: 2026-09-23T02:07:02.60006476Z
updated: 2026-09-24T20:47:48.689291732Z
sessions:
- start_sha: 1afd0f341b23e124940e7ed14440f2338370dd79
  end_sha: df9f271d0a74bb508132371df7b7f0d40a632f0b
  commits: 1
  started_at: 2026-09-23T16:51:36.322538387Z
  ended_at: 2026-09-23T17:04:29.609130989Z
- start_sha: 97c3893fd4ff1f1546c7984579545e96f6d2c33f
  end_sha: c2adc1f75f71cd96e8cf5067a659da2199c28462
  commits: 1
  started_at: 2026-09-23T17:27:59.570556235Z
  ended_at: 2026-09-23T17:34:42.078255112Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T16:51:36.322538387Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-23T17:04:29.609130989Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-09-23T17:27:59.511251306Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T17:27:59.570556235Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-23T17:34:42.078255112Z
observed_paths:
- conformance/10_io/00_adjacency.t
- conformance/10_io/09_glob_assign.t
- conformance/10_io/10_glob_ref.t
- conformance/10_io/11_autoload.t
- conformance/11_oo/14_state_ungated.t
- conformance/11_oo/15_class_ungated.t
- conformance/11_oo/16_classname_ungated.t
- conformance/11_oo/17_defer_ungated.t
- conformance/12_packages/00_adjacency.t
- conformance/12_packages/09_begin_end.t
- conformance/12_packages/10_compile_tokens.t
- conformance/12_packages/11_line_directive.t
- conformance/12_packages/12_use_constant.t
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier11_test.go
- internal/conformance/tier12_test.go
- internal/conformance/vacuous_test.go
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

- [ ] Tier 12 holds a file whose output pins BEGIN running before the main program (`python3 tools/mdtest/tiersrc.py 12_packages | grep -q 'BEGIN'`)
- [ ] A file pins that a # line directive rewrites __LINE__, which a comment-skipping lexer gets wrong (`python3 tools/mdtest/tiersrc.py 12_packages | grep -q '__LINE__'`)
- [ ] Every new file is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierPackagesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierPackagesPerlValidated'`)
- [ ] The dependency lint passes over the tier (`go test ./internal/conformance/ -run '^TestTierPackagesLint$' -count=1 -v | grep -q '^--- PASS: TestTierPackagesLint'`)
- [ ] The tier's adjacency file holds every construct it introduces (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q 'PASS: TestCorpus/adjacency-'`)
- [ ] The whole conformance suite is green with the new files in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
