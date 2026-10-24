---
title: The last-semicolon forgiveness hides a hash misread as a block
state: done
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-5f77-78ef-a0c7-538e0a241cdc
created: 2026-10-24T04:11:53.607169795Z
updated: 2026-10-24T14:40:06.338559036Z
sessions:
- start_sha: 1d4a6f4e304f58c864263cc69e31979df173b738
  end_sha: 39c57492239c1d5954f394b3eb49cd8f2bd675f4
  commits: 1
  started_at: 2026-10-24T13:37:57.683278124Z
  ended_at: 2026-10-24T14:39:31.022841924Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-24T13:37:57.683278124Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-24T14:39:31.022841924Z
observed_paths:
- .github/workflows/parse-fidelity.yml
- internal/parse/canon.go
- internal/parse/fidelity.go
- internal/parse/fidelity_test.go
---

Found by the independent review of NL-c (01a1692f-bf96, exec/nlc 3de59a88) on 2026-10-24. **perigrin's ruling (2026-10-24): narrow it in M3.** Faithful keeps the last-semicolon forgiveness, but it applies only where the TREE says the braces are a code block. It never applies to `{` … `}` that the tree reads as a hash, a subscript, or anything else. This issue moves into m3-oracle-at-scale and lands before the M3-dir unfaithful list (01a16602-5f77) is made, so that uni/variables.t is judged honestly.

## Context

- paths: `internal/parse/fidelity.go` (the last-semicolon rule), `internal/parse/canon.go`, and the parser's Block node
- entrypoints: `faithfulTokens`' last-semicolon rule

The last-semicolon forgiveness can hide a hash or a subscript that psc misreads as a block.
- A hand-built tree for `$h = {a => 1};` with its AnonHash replaced by a Block: canon writes `$h = {a => 1;};`, and Faithful returns true.
- At token level, `$h{a}` against `$h{a;}` is also forgiven.
- perl 5.44 rejects both `{ a => 1; }` and `$h{a;}` as syntax errors.

So Faithful, and every ratchet built on it, cannot see a block/hash misparse.

The reviewer's proposed fix follows the call-paren precedent: the tree records whether the source wrote a block's last `;` (as ArgParen records call parens), canon writes it back, and the last-semicolon rule is deleted. That would leave Faithful forgiving only extra paren nesting (01a1664d-75d4).

A narrower option: forgive the last `;` only inside a node the tree knows is a Block, never inside `{` … `}` that the tokens alone can't classify.

**Refined 2026-10-24, after the worker's STOP. perigrin ruled: narrow it, and fix the parser.** The three wrong-tree files (uni/variables.t, hash_ref_disambiguation.t, cpan_ceph_tie_runtime.t) cannot be caught here: psc's tree already says Block, and that is the misread itself. So this issue closes only the token-level holes, where a `;` sits inside braces the tree reads as a hash or a subscript (`$h{a;}`, `{a => 1;}`). NO file is expected to drop. Any file that goes faithful→unfaithful is a stop. The misreads are fixed in the parser instead: uni/variables.t by 01a16e20-f266, an M3 issue that blocks 5f77, and the two T1 files by v0.1 issues 01a16e21.

**Approach:** canonTokens already carries per-token marks (ParenNode). Add a mark for the `;` that canon writes as a code block's last statement terminator, and have last-semicolon forgive only a `;` carrying that mark. Then TestFaithfulForgivenessIsClosed's replay checks the mark.

## Acceptance criteria

### Positive Scenarios

- [ ] A `;` inside an AnonHash or a subscript is never forgiven, at canon-token level (`go test ./internal/parse/ -run '^TestFaithfulForgivesSemicolonOnlyInBlocks$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] `$h{a;}` against `$h{a}`, and a hash constructor written `{a => 1;}`, are unfaithful; a real block's last `;` is still forgiven (`go test ./internal/parse/ -run '^TestFaithfulLastSemicolonOnlyInBlocks$' -count=1 -v | grep -q '^--- PASS'`)
