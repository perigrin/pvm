---
title: The typeglob sigil appears three times and all three are sprintf widths
state: done
urgency: normal
milestone: corpus-complete
created: 2026-09-23T02:07:02.645428825Z
updated: 2026-09-23T17:37:12.987242398Z
sessions:
- start_sha: 97c3893fd4ff1f1546c7984579545e96f6d2c33f
  end_sha: c2adc1f75f71cd96e8cf5067a659da2199c28462
  commits: 1
  started_at: 2026-09-23T17:19:22.707327103Z
  ended_at: 2026-09-23T17:37:12.987242398Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T17:19:22.707327103Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-23T17:37:12.987242398Z
observed_paths:
- conformance/10_io/00_adjacency.t
- conformance/10_io/09_glob_assign.t
- conformance/10_io/10_glob_ref.t
- conformance/10_io/11_autoload.t
- conformance/12_packages/10_compile_tokens.t
- conformance/12_packages/12_use_constant.t
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/vacuous_test.go
---

The typeglob sigil does not appear in the corpus. Measured: `*` occurs
three times across 1,141 lines of source and all three are `sprintf`
width specifiers -- `"%0*d"`, `"%*d"`, `"%-*d"`. Zero typeglobs.

Also absent: `local *FH`, `\*STDOUT`, `*name = sub {}`, `*foo{THING}`,
`AUTOLOAD`, `DESTROY`. In T1 these run to roughly 100 file-uses.

## What a parser gets wrong here

`*` is a SIGIL introducing a fourth namespace, and it is the same
character as multiplication. `*alias = sub {...}` installs a sub at
runtime; `$a * $b` multiplies. Position decides, exactly as it does for
`%` and `<` -- the expect-state mechanism the lexer already has.

A corpus with no typeglob cannot tell a lexer that treats `*` as always
multiplication from one that gets it right. That is a one-character
lexical fork with no coverage at all.

`AUTOLOAD` and `DESTROY` are the other half: bareword sub names that are
special to the language rather than to the program, and which a class
body must not parse as ordinary subs.

## Files

Three, in `08_references`, which owns `\` and dereference spellings and
whose subject is exactly "a reference to a thing that is not a value":

- `*name = sub {...}` with a call afterwards proving installation, plus a
  token fact separating the glob sigil from multiplication in the same
  file.
- `\*STDOUT` and `local *FH`, the reference and dynamic-scope forms.
- `AUTOLOAD` with a call to an undefined sub, pinned by output.

`DESTROY` needs an object going out of scope, which is tier 11's; it goes
there rather than here, in the same file family as `bless`.

## Acceptance criteria

- [ ] The corpus source contains a typeglob, which it did not before -- the only three * uses were sprintf widths (`awk '/^--- source/{s=1;next} /^--- expect/{s=0} s' conformance/08_references/*.t | grep -qE '(^|[^*%0-9-])\*[A-Za-z]'`)
- [ ] A file pins that a glob assignment installs a callable sub, by calling it (`go test ./internal/conformance/ -run '^TestCorpus/08_references' -count=1 -v | grep -q '^    --- PASS: TestCorpus/08_references'`)
- [ ] Every new file is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierReferencesPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierReferencesPerlValidated'`)
- [ ] The dependency lint passes over the tier (`go test ./internal/conformance/ -run '^TestTierReferencesLint$' -count=1 -v | grep -q '^--- PASS: TestTierReferencesLint'`)
- [ ] The tier's adjacency file holds every construct it introduces (`go test ./internal/conformance/ -run '^TestTierReferencesAdjacency$' -count=1 -v | grep -q '^--- PASS: TestTierReferencesAdjacency'`)
- [ ] The whole conformance suite is green with the new files in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
