---
title: "Types paper: parametric containers, tuples, lvalue lists, and multi and context-indexed function types"
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T01:42:09.165064836Z
updated: 2026-10-05T21:12:01.61813549Z
---

## Context

- paths: `~/dev/types-paper/2-perl-types-formal.md` (a separate git repository, `~/dev/types-paper`, not pvm)
- commands: `sh -c 'grep -qiE "^##+ .*parametric" $HOME/dev/types-paper/2-perl-types-formal.md'`
- entrypoints: "## The Type Lattice" (its "### List Types" and "### Completeness of the container types"), "## Complete Type Membership Definition", "## Subtyping Relationship" and "## Join and Meet" in `2-perl-types-formal.md`; RFC 0001's "Open questions" 1, 3, 4 and 5 (`docs/rfcs/0001-type-libraries.md`)

perigrin (2026-10-01): parametric types go into the paper (~/dev/types-paper) first, so pvm and perl5-son extend one lattice. RFC 0001's open questions 1, 3, 4 and 5 wait on it:

- `List[T]` (values in flight, covariant) against `Array[T]` (a container, invariant); `Hash[T]` flattening to `List[Str|T]`; `ArrayRef[T]` from `Array[T]`; whether `T <: List[T]`.
- Lvalue lists: `chomp`/`chop` modify their arguments in place.
- Tuples: `each` returns `(Str, T)` for a hash.
- Function types with several signatures, selected by arity, argument type or context.
- Blessing and reftype: an `IO::File` object is a blessed reference of reftype `GLOB`, and the lattice files every blessed reference under `Object`, losing the `GLOB`. RFC 0001's type-names section keeps such objects out of `FileHandle` until this is settled.
- Blessing and reftype, a measured precedent (2026-10-05): with the lattice oracle's own predicates, `*STDOUT{IO}` and `*{$fh}{IO}` are references of reftype IO blessed into IO::File, so they classify as Object -- exactly as qr// (reftype REGEXP, blessed Regexp) does, which the lattice already makes a leaf UNDER Object (Regex). An IO reference is the same shape; whether it becomes a leaf under Object beside Regex (making FileHandle = Glob|GlobRef|<IO reference>) is the question. pvm commit 02df00ad witnesses `*STDOUT{IO}` on the oracle's Object <: Ref edge.

Membership in the paper is defined through interpretation (`v ∈ T` iff interpreting `v` through `T` is lossless and its contracts hold), and the parametric section should keep that definition.

The work is a new `##` section of `2-perl-types-formal.md` whose heading names parametric types, placed after "## The Type Lattice", with subsections for: container types and variance (`List[T]`, `Array[T]`, `Hash[T]`, `ArrayRef[T]`, and whether `T <: List[T]`); lvalue lists; tuples; function types with several signatures selected by arity, argument type or context; and blessing against reftype. Each states membership and subtyping through interpretation, as the scalar types do. This issue blocks no pvm issue: the RFC's open questions stay open until it lands, and the pvm chain builds only what is decided.

## Steps

- [ ] Draft the parametric section's heading and the container-types-and-variance subsection, membership through interpretation; commit in `~/dev/types-paper`
- [ ] Verify the heading: `sh -c 'grep -qiE "^##+ .*parametric" $HOME/dev/types-paper/2-perl-types-formal.md'`
- [ ] Add the lvalue-lists, tuples, function-types and blessing-against-reftype subsections; commit each
- [ ] Run the paper's lattice verifier (`appendix-a/lattice-check.pl`, in `~/dev/types-paper`), and review the section with perigrin before marking this done

## Acceptance criteria

- [ ] The paper has a parametric-types section (`sh -c 'grep -qiE "^##+ .*parametric" $HOME/dev/types-paper/2-perl-types-formal.md'`)
