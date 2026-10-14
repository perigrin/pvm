---
title: "\\(@a) is typed ArrayRef, but perl takes a reference to each element"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-14T02:07:19.798045969Z
updated: 2026-10-14T02:07:19.798045969Z
---

## Context

- paths: `internal/infer/lowered_special.go` (`ref`), legacy `internal/infer/infer.go` (`inferRefgenType`)
- commands: `go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$'`; perl 5.42 `/home/perigrin/.local/bin/perl`

Found 2026-10-14 by the special-form port (01a12a85-3386), which copied it so the golden `special/ref_paren` matches: `\(@a)` is typed ArrayRef. perl takes a reference to EACH element: `my @a = (1, 2); my $r = \(@a); print ref($r), " ", $$r` prints `SCALAR 2` (the last element's ref, in scalar context), and `my @rs = \(@a)` gives 2 refs. The lowered tree already distinguishes the two (Paren is a bit on the Var, per perigrin's ruling with this exact case: `\(@a)` = 3 vs `\@a` = 1), so the fix is local. Fix after the cutover (01a0c115-df67) with the other copied defects; the golden edit gets a `Diagnostics-changed:` trailer.

## Steps

- [ ] Failing test first; a Ref of a parenthesised aggregate is a list of ScalarRefs (ScalarRef in scalar context); golden under a trailer; `make test`; commit

## Acceptance criteria

- [ ] `\(@a)` is ScalarRefs, not an ArrayRef (`go test ./internal/infer/ -run '^TestRefOfParenAggregateIsElementRefs$' -count=1 -v | grep -q '^--- PASS: TestRefOfParenAggregateIsElementRefs'`)
