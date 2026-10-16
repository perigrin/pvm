---
title: "[D1] Specs and comments still describe the tree-sitter path and pre-lowering state as current"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:38.047206225Z
updated: 2026-10-16T01:17:19.411338936Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:19.411338936Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:19.411338936Z
---

## Context

- found by: the agentic review of PR #486 (plan alignment); report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`
- paths: `docs/specs/perl-parser/04-expressions.md` (§4.14), `docs/specs/perl-parser/05-statements.md` (§5.17), `docs/specs/perl-parser/06-incremental-lsp.md`, `internal/lower/positions.go:96`, `internal/lower/node.go:374`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486. These are the spec and comment deviations the review confirmed: text that describes the tree-sitter path or the pre-lowering state as current. Update each to describe what ships (keep dated measurements as history, marked as such). The statement vocabulary needs a home: either §4.14 gains the statement kinds as they ship, or §5.17 is revised to match -- pick the smaller edit. Record the given/when/try/defer/format Hole deferral, `field`, and the latency benchmark as open items with their issue IDs.

#### Deviations (stale steering text; I checked each cited line)
- **04-expressions.md:1492-1499** says, in the present tense, "`internal/infer/infer.go:184` keys its annotation map on StartByte", and cites `nearestPatternBefore (infer.go:1724-1751)`.
  - Annotations are keyed by `lower.ID`.
  - `nearestPatternBefore` no longer exists in internal/infer.
  - `lower/positions.go:96` and `positions_test.go:41` still cite `nearestPatternBefore, infer.go:1726`.
- **04-expressions.md:2082-2104 (§4.14.6 "Still open")** is stale:
  - It says "The consumer handles none of the four" (local, state, our, field); infer handles our, state and local.
  - It cites `infer.go:1569` and `:1741`, which are deleted sites.
  - It treats the Arrow question as open; that is moot now.
  - It lists the three consumer defects as untracked; they are fixed, yet 01a121c3-b0bc-7e12-a196-1d1fd563ab80 is still pending.
- **04-expressions.md:2020** says "The lowering needs a PostCond bit ... this list has neither". `WhileStmt.PostCond` exists.
- **The statement vocabulary** (`node.go:374`, "§5.17's kinds") does not match 05-statements.md §5.17's shapes. Neither spec describes what ships.
- **04-expressions.md:1608 and :1537** are incomplete:
  - LoopEx also carries `return`, with its value in `Label`.
  - Unary.Op uses `neg` and `pos` (plus `not` and `~.`).
- **Unrecorded erasures:**
  - `my Foo $f` drops its class.
  - A label on a non-loop statement is dropped.
  - A package version is dropped when a block follows it (S10).
  - A one-link `cmp_chain` lowers to Binary.
- **06-incremental-lsp.md is stale:**
  - :13 says infer "consumes the tree-sitter tree through the internal/parser façade".
  - :465 states item 7 as the rule.
  - :699 says the node-identity scheme is "Undesigned" and the key is "infer.go:184".
  - :701 lists the defects as open.
  - :1133 cites `walkUseStatement (infer.go:22-53)`, which is now `useStmt` in flow.go (contract-B #11).
  - §6.4.3 and §6.10 describe a `map[uint32]` StartByte key.
- **Plan naming.** docs/plans/2026-09-05-parser-conformance-plan.md:580 defines "M2" as "Incremental, and correct under edit", while this milestone is `m2-lower-the-tree`. The conformance ladder's M2 gate stays open.

---

## Steps

- [ ] Edit each cited line; cite issue IDs for the open items; `make test`; commit

## Acceptance criteria

- [ ] no spec or comment describes the deleted tree-sitter path or StartByte keys as current (`! grep -qE 'nearestPatternBefore|infer\.go:184|keys its annotation map on StartByte' docs/specs/perl-parser/04-expressions.md internal/lower/positions.go && ! grep -q 'consumes the tree-sitter tree' docs/specs/perl-parser/06-incremental-lsp.md`)
