---
title: Tier 01 written against the finished tooling, which is what proves it works
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c35f-a487-7aa8-934b-4e68dd0fbaa1
- 01a0c35e-d2b6-7993-9059-5675239aea65
- 01a0c35e-66f9-71b2-b5f3-3dd0ef1cba05
- 01a0c35f-57cc-7a0f-83ac-53654f46f235
blocks:
- 01a0c361-6191-74b6-a1c0-263f58d465b3
- 01a0c361-c7d7-7ea4-9c93-b8bcc80a108e
- 01a0c362-047b-76a4-b0f8-9043eee66baa
- 01a0c362-073a-78c2-ba73-227fb48a26c4
- 01a0c362-127c-73c1-8bb8-a4bf8b66c735
- 01a0c362-0a97-7c14-82fe-27d76ee5c862
- 01a0c362-0c0b-7f6a-b807-fa62daefe02e
created: 2026-09-21T09:50:44.159860455Z
updated: 2026-10-23T19:18:29.324390838Z
sessions:
- start_sha: 7711154e001e09d6d7b489eaa6d325851e7a4e30
  end_sha: 9750d03b51c601e083873e26c01cb22054bb120b
  commits: 1
  started_at: 2026-09-21T20:03:38.532861151Z
  ended_at: 2026-09-21T20:21:11.857566311Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T20:03:38.532861151Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T20:21:11.857566311Z
observed_paths:
- conformance/01_literals/00_adjacency.t
- conformance/01_literals/01_binary.t
- conformance/01_literals/03_hexadecimal.t
- conformance/01_literals/04_leading_decimal.t
- conformance/01_literals/08_signed_exponent.t
- conformance/01_literals/05_negative.t
- conformance/01_literals/06_octal_leading_zero.t
- conformance/01_literals/07_octal_prefix.t
- conformance/01_literals/09_trailing_decimal.t
- conformance/01_literals/10_underscore_separators.t
- conformance/01_literals/11_vstring_bare.t
- conformance/01_literals/12_vstring_v.t
- conformance/01_literals/README.md
- conformance/02_variables/01_array.t
- conformance/GLOSSARY.md
- internal/conformance/conformance_test.go
- internal/conformance/glossary_test.go
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier01_test.go
---

The tooling issues before this one are each separately checkable, and none
of them proves the tooling works together. Tier 01 does. It is the first
tier whose constructs are lexical, which means the token stream classifies
it and a declaration would be checking itself.

## What the spike established, and what it did not

Three files exist: `02_decimal.t` passes, `03_leading_decimal.t` and
`04_signed_exponent.t` refuse. Enough to show the format holds, nowhere near
enough to be a tier. Numbering starts at 02 because 01 was never written --
regenerate rather than preserve the gap.

## The finding this tier carries forward

`5e-1`, `5e+1`, `5E-1` and `1.5e-3` all split into three tokens; `5e1` is
fine. `perl -e 'print 5e'` is a syntax error, so `Number("5e")` is not a
token perl would ever produce.

**The parser cannot see this.** It gets `Number Operator Number`, reads
valid subtraction, and produces ZERO Unknown nodes. Behaviour cannot see it
either -- both print `0.5`. Nor can B::Concise, because perl never builds
the wrong tree to compare against. The token assertion is the only check
that catches it, which is the measured argument for the whole token layer
and belongs in this tier's README.

Criteria updated 2026-10-23: e8ee9001 and a6b33a2a replaced the tier-keyed tests with topic files; each remapped command now names the test that holds the same property (old -> new: `TestCorpus/01_literals` -> `TestCorpus` over the tier-01 topics (adjacency-01_literals.md, numeric-point.md, numeric-radix.md, strings.md, vstrings.md)). Left unchanged: `TestTierLiteralsAdjacency`, `TestTierNumberingRegenerates`, because no current test holds that property.

Criteria updated 2026-10-23 (agent:ccx-adj, issue 01a1696e): the deleted per-tier adjacency tests are replaced by one generic test, `TestAdjacencyCoversItsTier`, whose per-tier subtest checks that the tier's adjacency topic emits every op its README's INTRODUCES list names, that each case in it is multi-statement, and that it emits an op of the declared prerequisite. Old -> new: `TestTierLiteralsAdjacency` -> `TestAdjacencyCoversItsTier/01_literals`; `TestTierNumberingRegenerates` -> `TestEveryTierDeclaresFileOrder`. These commands pass once branch ccx/adj merges.

Obsolete by perigrin's ruling 2026-10-23: the criterion that files are renumbered from 01 with no gap asserted a property of numbered `.t` files. The corpus is topic files now and topics carry no numbers, so each tier instead declares its ordering convention under `## FILE ORDER` in its README, which `TestEveryTierDeclaresFileOrder` checks.

## Acceptance criteria

- [ ] Tier 01 covers the numeric and string literal boundaries the glossary names, including v-strings and the -1-is-two-tokens case (`go test ./internal/conformance/ -run '^TestTierLiteralsCoversGlossary$' -count=1 -v | grep -q '^--- PASS: TestTierLiteralsCoversGlossary'`)
- [ ] Tier 01 declares its topic ordering in its README; topics carry no numbers, so there is nothing to renumber (`go test ./internal/conformance/ -run '^TestEveryTierDeclaresFileOrder$' -count=1 -v | grep -q '^--- PASS: TestEveryTierDeclaresFileOrder'`)
- [ ] 03_leading_decimal.t's citation of 01a0c13f-97f5 resolves to a live issue (`git-zhi issue show 01a0c13f-97f5 >/dev/null`)
- [ ] Every file in 01_literals/ is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierLiteralsPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierLiteralsPerlValidated'`)
- [ ] The dependency lint passes over 01_literals/: every construct used was introduced at or before this tier (`go test ./internal/conformance/ -run '^TestTierLiteralsLint$' -count=1 -v | grep -q '^--- PASS: TestTierLiteralsLint'`)
- [ ] The tier's adjacency topic puts every construct 01_literals introduces in one body, each adjacent to another (`go test ./internal/conformance/ -run '^TestAdjacencyCoversItsTier$/^01_literals$' -count=1 -v | grep -q '^    --- PASS: TestAdjacencyCoversItsTier/01_literals'`)
- [ ] 01_literals/README.md states what the tier introduces, what it depends on, and why check 2 places it here, in the form the dependency lint reads (`go test ./internal/conformance/ -run '^TestEveryTierHasAReadme$' -count=1 -v | grep -q '^--- PASS: TestEveryTierHasAReadme'`)
- [ ] Each refusing file cites a refusal code rather than a message (`go test ./internal/conformance/ -run '^TestTierLiteralsRefusalsCited$' -count=1 -v | grep -q '^--- PASS: TestTierLiteralsRefusalsCited'`)
- [ ] The tier's own files are green with skips limited to documented refusals (`go test ./internal/conformance/ -run '^TestCorpus$/^(adjacency-01_literals|numeric-point|numeric-radix|strings|vstrings)\.md$' -count=1 -v | grep -q '^--- PASS: TestCorpus'`)
