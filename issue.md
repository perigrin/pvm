---
title: The last-semicolon forgiveness hides a hash misread as a block
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-5f77-78ef-a0c7-538e0a241cdc
created: 2026-10-24T04:11:53.607169795Z
updated: 2026-10-24T12:53:22.08427014Z
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

**Why it matters for M3:** three files are counted faithful while psc builds the wrong tree for them. All three fail `perl -c` near `;}`: uni/variables.t (in an M3 dir), hash_ref_disambiguation.t (T1) and cpan_ceph_tie_runtime.t (T1). Once this issue lands they become unfaithful. That is correct, and the drop is expected and re-baselined in the same commit, naming exactly these files. Any other file that drops is a stop: report it to perigrin.

**Approach:** canonTokens already carries per-token marks (ParenNode). Add a mark for the `;` that canon writes as a code block's last statement terminator, and have last-semicolon forgive only a `;` carrying that mark. Then TestFaithfulForgivenessIsClosed's replay checks the mark.

## Acceptance criteria

### Positive Scenarios

- `go test ./internal/parse/ -run '^TestFaithfulCatchesBlockForHashMisread$' -count=1 -v | grep -q '^--- PASS'` (a hand-built AnonHash-as-Block tree is unfaithful)
- `go test ./internal/parse/ -run '^TestFaithfulLastSemicolonOnlyInBlocks$' -count=1 -v | grep -q '^--- PASS'` (`$h{a;}` against `$h{a}`, and a hash constructor written `{a => 1;}`, are unfaithful; a real block's last `;` is still forgiven)
