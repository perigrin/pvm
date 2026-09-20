---
title: "The lowering pass: internal/infer must read internal/parse, not tree-sitter"
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 019d1449-1880-729b-860b-ac29464f6cd7
created: 2026-09-20T22:19:52.755941663Z
updated: 2026-09-20T22:59:21.087904664Z
---

`internal/infer` reads the tree-sitter tree through `internal/parser/`. That
tree is being burned down, so every inference rule written against it is work
that gets deleted. Nothing in `internal/infer` should be extended until it
reads `internal/parse/` instead.

## Why no partial work survives

§6.1.6 counted it, and the count is the argument:

> `internal/infer/*.go` switches on 34 distinct tree-sitter-perl grammar
> names (`subroutine_declaration_statement`,
> `ambiguous_function_call_expression`, `array_element_expression`,
> `hash_deref_expression`, `func0op_call_expression`, `refgen_expression`,
> ...), and `internal/parse/parse.go:146-201` emits 23 coarse kinds with none
> of those names (`statement`, `term`, `binary`, `index`, `declaration`,
> `call`, ...). Porting the facade hands PSC a `Kind()` that returns
> `"binary"` where it expects `"assignment_expression"`. **Ten methods port;
> the vocabulary does not.**

So this is not a migration that can be done incrementally behind a shim. The
vocabulary changes underneath every switch.

## And the old tree is re-derived from TEXT in places

§4.14.2 records where `internal/infer` reaches past the tree entirely:

    findOperatorText    infer.go:645-659
    callElementType     infer.go:1313, telling map from grep by
                        strings.HasPrefix on source

Those are the places new rules would have to reproduce, and they are exactly
what a lowered tree exists to remove.

## Scope

The lowering pass of §6.1.6: a pass over `internal/parse`'s tree producing
the vocabulary §4.14 specifies, which `internal/infer` then consumes instead
of `internal/parser`.

§6.1.6 is explicit about what NOT to build with it: "One lowering pass,
specified in §6.1.6, and no dataflow IR until a feature demands one." Neither
HIR nor PIR as perl-lsp built them.

Key facts by NODE IDENTITY, not byte offset. §6.1.6 item 4, and rust-analyzer
rejected the offset key by name: "offsets change after edits. So, as a rule
of thumb, we avoid using offsets, text ranges or syntax trees as keys and
values for queries". §6.4.3 observes the same defect in PSC's existing
`map[uint32]types.Type`.

## Size

5,012 non-test lines in `internal/infer` at 3b9422eb, 34 grammar names to
re-express. This is a multi-session milestone-sized piece of work and should
probably be decomposed before it is executed -- filed here so that the thing
everything else waits on is ON the chain rather than implied by a spec
section.

## What it unblocks

  019d1449  Expand PSC inference       -- blocked on this, now explicitly
  area 5 of that issue                 -- NarrowByContext has no source of
                                          context until the lowered tree
                                          carries Wantarray and call-site
                                          context

## Acceptance

Deliberately not written. The ACs for a pass this size should come from a
decomposition pass (crochet:refinement) against §6.1.6 and §4.14, not from
one sitting.

Plan: M2/M3 adjacent -- it is what makes §6.1.6's "PSC is being converted to
the new tree regardless" true. Spec: §6.1.6, §4.14, §6.4.3.
