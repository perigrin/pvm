---
title: The gates the corpus replaces die, including easy_test.go whole
state: done
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c361-6191-74b6-a1c0-263f58d465b3
- 01a0c362-047b-76a4-b0f8-9043eee66baa
- 01a0c362-073a-78c2-ba73-227fb48a26c4
- 01a0c362-0929-795d-ac6b-c1740240e49e
- 01a0c362-0daa-7422-b787-7650ce684c4b
- 01a0c362-127c-73c1-8bb8-a4bf8b66c735
created: 2026-09-21T13:44:43.235895078Z
updated: 2026-09-21T22:10:47.997430711Z
sessions:
- start_sha: dc1bea2c2baed5af0cf71ddf16e379ee18855d30
  end_sha: b41801021927f68fcbf799c60d6b947393d983de
  commits: 1
  started_at: 2026-09-21T21:39:35.991690255Z
  ended_at: 2026-09-21T22:10:47.997430711Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T21:39:35.991690255Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T22:10:47.997430711Z
observed_paths:
- conformance/02_variables/README.md
- conformance/06_control/README.md
- conformance/07_subroutines/00_adjacency.t
- conformance/07_subroutines/08_parenless_extent.t
- conformance/07_subroutines/09_prototype_extent.t
- conformance/07_subroutines/11_code_ref_call.t
- conformance/07_subroutines/README.md
- conformance/08_references/README.md
- conformance/11_oo/README.md
- conformance/13_opaque/README.md
- conformance/README.md
- internal/conformance/claimonce_test.go
- internal/conformance/conformance_test.go
- internal/conformance/file.go
- internal/conformance/hardmarkers_test.go
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier07_callforms_test.go
- internal/conformance/tier07_test.go
- internal/parse/easy_test.go
- internal/parse/t2_test.go
---

The spec commits to this and nothing in the chain owned it. Per the plan
discipline rule, the last 10-20% is where the replacement ships and the
replaced gate survives.

## What the spec says dies

From "What this replaces":

> `TestT1EasyParseRate` and its 70% target, and the T2 100% target.
> "Parses cleanly through tier N" says something; "70% of an alphabetical
> pile that includes Archive::Zip integration tests" does not.

And from Sequencing:

> **`hardMarkers` is scaffolding.** Mine it for tier placements, then delete
> `easy_test.go` whole. It encodes "constructs this parser failed on at
> 4bc971ec", which in a graded corpus is a tier placement rather than a
> filter.

## Both are live today, measured

    internal/parse/easy_test.go:161    func TestT1EasyParseRate
    internal/parse/t2_test.go:196      "The gate's target is 100%"

## The mining half is already happening

Nine tier issue bodies cite `hardMarkers` placements as evidence -- "Three
of the twelve place here", "Four place here". That is the first half of the
spec's instruction. This issue is the second half, and it cannot run until
all twelve markers have a tier, which is why it is blocked by the tier
issues that claim them.

The trigger condition is precise and already encoded: every one of the
twelve markers named in a tier README's INTRODUCES set.

## Acceptance

- [ ] All twelve hardMarkers are placed in a tier README before anything is deleted (`go test ./internal/conformance/ -run '^TestEveryHardMarkerPlaced$' -count=1 -v | grep -q '^--- PASS: TestEveryHardMarkerPlaced'`)
- [ ] easy_test.go is deleted whole, not commented out or skipped (`! test -e internal/parse/easy_test.go`)
- [ ] No reference to TestT1EasyParseRate survives anywhere in the tree (`! grep -rq 'TestT1EasyParseRate' --include='*.go' .`)
- [ ] The T2 100% target is gone from t2_test.go, replaced by the corpus gate (`! grep -q "target is 100" internal/parse/t2_test.go`)
- [ ] The package still builds and its tests pass with the gates removed (`go test ./internal/parse/ -count=1`)
