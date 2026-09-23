---
title: Four parser gaps the corpus recorded, tracked so their citations point somewhere live
state: pending
urgency: normal
milestone: corpus-complete
created: 2026-09-23T17:10:01.515942446Z
updated: 2026-09-23T17:10:01.515942446Z
---

Four parser gaps found by the `corpus-complete` milestone, each recorded
by a corpus file that cites a now-closed issue. This issue is where they
are tracked so the citations point somewhere live.

## The gaps

    conformance/11_oo/16_classname_ungated.t   trailing_tokens
    conformance/11_oo/17_defer_ungated.t       unimplemented_statement
    conformance/12_packages/10_compile_tokens.t  trailing_tokens
    conformance/12_packages/11_line_directive.t  trailing_tokens

Plus one from the compound-assignment slice, already filed separately as
`01a0ce57-db92-78e5-bbe4-7c28e74db6f3` (`x=` is never lexed).

## What each needs

**`__CLASS__` ungated** is the `isa` finding in a different keyword.
Measured, perl reads `print __CLASS__;` as `print __CLASS__ $_` -- a
filehandle print over a glob. Our parser leaves trailing tokens instead.

**`defer BLOCK` ungated** is the sharpest of the set and the one whose
claim is currently unexercised. Measured, the block runs FIRST, because
ungated `defer` is an indirect method call whose brace group is an
anonymous hash evaluated eagerly; gated, it runs at scope exit. Same two
lines, opposite order. `17_defer_ungated.t` pins the ungated ordering by
output and our parser refuses the file, so the ordering claim is
validated against perl and not yet against us.

**`__PACKAGE__`** and the **`#line` directive** are tier 12's two. The
directive is the more consequential: a `#` at column zero matching
perl's pattern is not a comment, and a lexer that treats it as one
reports every subsequent error location wrong, silently.

## Why this issue exists rather than a relaxed criterion

`01a0cc04-339f`'s acceptance criterion originally demanded the defer
ordering be "pinned by output" -- meaning our parser handles the file.
It was edited after the issue closed to accept a recorded refusal
instead. The tier relocation in that same edit was forced by the op
budget and legitimate; the PASS-to-SKIP relaxation was not covered by
that reason and rode across on its back.

The honest repair is to track the gap rather than lower the bar, which
is what this issue is. When the parser learns these constructs, the four
files start passing and the ratchet reports it.
