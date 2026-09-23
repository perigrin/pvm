---
title: "The bitwise and shift band: three precedence levels at zero"
state: done
urgency: normal
milestone: corpus-complete
blocks:
- 01a0cc04-333a-70f7-a182-d4d0a14cc0de
- 01a0cc04-32c9-722c-8728-d4043e7d3c14
created: 2026-09-23T02:07:02.403766667Z
updated: 2026-09-23T02:54:54.864571063Z
sessions:
- start_sha: 398b8fdf91f7fbfdce79dcab401c879a3e9eeb9c
  end_sha: b1a62eecc58a5d7bd0c1af6e2b54ae4691aaa2c2
  commits: 1
  started_at: 2026-09-23T02:45:17.110911611Z
  ended_at: 2026-09-23T02:54:54.864571063Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T02:45:17.110911611Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-23T02:54:54.864571063Z
observed_paths:
- conformance/04_operators/00_adjacency.t
- conformance/04_operators/16_bitwise.t
- conformance/04_operators/17_bitwise_precedence.t
- conformance/04_operators/18_shift.t
- conformance/04_operators/19_complement.t
- conformance/04_operators/README.md
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier04_test.go
---

Three whole precedence levels with no coverage: perlop's 9 (`<< >>`), 14
(`& &.`) and 15 (`| |. ^ ^.`). Measured across the corpus source, every
`&` is the code sigil (`\&foo`, `&$code`), every `|` is either `//` or
regex alternation, and `^` appears nowhere outside a regex. The `<<` hits
are all heredoc openers and the one `>>` is a format picture line.

That last part is a trap the corpus sets for itself: `<<` is lexed two
entirely different ways and only the heredoc way is present. A lexer that
always read `<<` as a heredoc opener would pass tier 13 and every other
tier.

## What a parser gets wrong here

The band is contiguous with the assignment band above it, so a parser
built to this corpus has a correct expression grammar with a hole exactly
where C-derived operators live. Precedence between them is the claim:
`$a | $b & $c` is `$a | ($b & $c)`, and nothing in the corpus says so.

`~` is a named unary at level 5 and is equally absent; it belongs here
because its only observable use is against the binary forms.

## Placement

The ops are `bit_and`, `bit_or`, `bit_xor`, `complement`, `left_shift`,
`right_shift`, all unclaimed. They belong in `04_operators`, which owns
the precedence table.

Both operands must be RUNTIME values or the optimiser folds the operator
away -- the trap tier 04's README already records for arithmetic, and the
reason its files take `$ENV{X} // <default>`.

## Files

Four: the bitwise trio (`& | ^`) with a precedence pair proving `&` binds
tighter than `|`; the shifts (`<< >>`) with a token fact separating the
shift from a heredoc opener; `~` complement; and one adjacency-style file
placing the band against the arithmetic level so the relative precedence
is claimed rather than assumed.

The stringwise spellings `&. |. ^. ~.` are NOT here -- they are feature
gated and belong with the pragma pair issue.

## Acceptance criteria

- [ ] Tier 04 introduces the bitwise and shift ops, which no tier claimed before (`grep -A2 '^## INTRODUCES' conformance/04_operators/README.md | grep -q 'bit_and'`)
- [ ] The corpus source contains a bitwise operator outside a regex or a code sigil (`awk '/^--- source/{s=1;next} /^--- expect/{s=0} s' conformance/04_operators/*.t | grep -qE '[0-9a-z_)] +\& +[0-9$]'`)
- [ ] Every new file is validated against pinned perl 5.42.0 before it counts (`go test ./internal/conformance/ -run '^TestTierOperatorsPerlValidated$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsPerlValidated'`)
- [ ] The dependency lint passes over the tier (`go test ./internal/conformance/ -run '^TestTierOperatorsLint$' -count=1 -v | grep -q '^--- PASS: TestTierOperatorsLint'`)
- [ ] No op is introduced by two tiers after the new claims land (`go test ./internal/conformance/ -run '^TestOpClaimedByExactlyOneTier$' -count=1 -v | grep -q '^--- PASS: TestOpClaimedByExactlyOneTier'`)
- [ ] The whole conformance suite is green with the new files in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
