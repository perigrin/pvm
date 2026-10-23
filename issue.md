---
title: Order cause-table rule 5 after the canon-adds-paren rules
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T14:43:13.184804128Z
updated: 2026-10-23T14:43:13.184804128Z
---

## Context

Raised by a fresh reader (fresh4), m3-oracle-at-scale chain-review round 8 (follow-up).

In 5f77's 9-rule class table, rule 5 (EOF on either side goes to stmt-boundary) comes before rules 7 and 8. So a canon that adds a paren after the source's last token is classed stmt-boundary. Today only lib/universal.t hits it.

## Steps

- [ ] Limit rule 5 to cases where canon isn't `(`, or state that the routing is intended, with a fixture

## Acceptance criteria

### Positive Scenarios
- [ ] A canon-added paren at EOF is classed canon-adds-paren (`go test ./internal/parse/ -run '^TestFaithfulQueueEOFParenIsCanonAddsParen$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueEOFParenIsCanonAddsParen'`)
