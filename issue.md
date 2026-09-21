---
title: The gates the corpus replaces die, including easy_test.go whole
state: pending
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c361-6191-74b6-a1c0-263f58d465b3
- 01a0c362-047b-76a4-b0f8-9043eee66baa
- 01a0c362-073a-78c2-ba73-227fb48a26c4
- 01a0c362-0929-795d-ac6b-c1740240e49e
- 01a0c362-0daa-7422-b787-7650ce684c4b
created: 2026-09-21T13:44:43.235895078Z
updated: 2026-09-21T13:45:11.528812405Z
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
