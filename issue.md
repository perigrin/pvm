---
title: Rename Canon to a name that does not suggest perl's Deparse output
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T14:28:58.030100387Z
updated: 2026-10-23T14:32:20.007268722Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/canon_test.go` and every test calling `Canon`, `internal/parse/testdata/t1canon.ratchet`, `docs/specs/perl-parser/` (prose), `docs/plans/2026-09-05-parser-conformance-plan.md` (prose)
- entrypoints: `internal/parse/canon.go`

perigrin, 2026-10-23: "canon is perhaps a poorly chosen name for that since most people would think canon would be the output of `perl -MO=Deparse`." `Canon` is psc's own re-emitter: it writes a parse tree back out as Perl, with the tree's grouping and the source spelling the tree records. It is not perl's canonical form. What it is, in perigrin's words: a deparser, psc's own, from parse tree to Perl source, the way `B::Deparse` goes from op tree to Perl. It is not lowering: `internal/lower` already names the parse-tree-to-inference step, and lowering goes down to a simpler form while this goes back up to source. Proposed name: `Unparse` (`unparse.go`), the standard compiler term for parse tree back to source, unused in the repo. Prose can call it "psc's deparser". Rejected: `Deparse`, because every use of the word in this repo means perl's `B::Deparse`. It appears in about 388 comments quoting `-MO=Deparse` measurements, and in test-local `deparse` closures that run `perl -MO=Deparse` as an oracle (e.g. canon_test.go:397, `-MO=Deparse,-p`). No Go identifier uses it. and `Reemit` (the plan uses "re-emission" for perl's output too). Confirm the name with perigrin before starting.

Deferred until after m3-oracle-at-scale lands, by perigrin's decision, so as not to churn its reviewed chain. That milestone adds more `TestCanon*` names and edits `canon.go` heavily, so do this rename after it merges.

## Steps

- [ ] Also correct the plan's M1 table, which describes TestCanonFidelityRatchet as "perl Deparses canon as it Deparses the source". The test runs `parse.Faithful`, which compares canon's tokens with the source's; no perl is involved (fidelity_test.go:111).

- [ ] Confirm the name with perigrin
- [ ] Rename the function, file, ratchet file and every `TestCanon*` test in one mechanical, behaviour-neutral commit; `make test` passes, and the canon, fidelity and parse ratchets move 0 files
- [ ] Update spec and plan prose ("canon" to the new name), keeping historical quotes intact
- [ ] Commit

## Acceptance criteria

### Positive Scenarios
- [ ] No exported or test identifier is named Canon any more (`sh -c '! grep -rEn "func (Test)?Canon[A-Z(]" internal/parse'`)
- [ ] The T1 re-emission ratchet still holds after the rename (`go test ./internal/parse/ -run '^TestUnparseRatchet$' -count=1 -v | grep -q '^--- PASS: TestUnparseRatchet'`)
