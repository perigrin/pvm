---
title: A bareword hash subscript is a string, not a call
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T12:19:07.059614571Z
updated: 2026-09-19T12:19:07.059614571Z
---

A bareword hash subscript is stored as a Call, so the key looks like a
function invocation. Found by canonical re-emission (01a0a8d8), which makes
it visible:

    $ $h->{k} = 1;
    canon -> $h->{k()} = 1;

    index {
      term $h
      call k            <-- should be a Term holding the string "k"

Perl autoquotes a bareword subscript, and does so EVEN WHEN a sub of that
name is in scope. Measured on 5.42.0:

    $ perl -e 'sub k { "other" } my %h = (k => 1); print $h{k}, "\n";'
    1

So `$h{k}` is the string key and `$h{k()}` is the sub's return value. They
are different programs, and the tree currently says the second.

## Why this matters beyond the emission

`Call{Resolved:false}` is how §4.8.3 records "a call to something I have not
seen". A hash key is not a call at all, so every bareword subscript in the
corpus is currently a false entry in that set -- it inflates the unresolved
count and, through sites.go, the hedge the parse oracle scores.

## Scope

Both bracket forms and the arrow form:

    $h{k}       $h->{k}
    $h{-k}      leading minus autoquotes too
    $h{k()}     an explicit call is NOT autoquoted and must stay a Call

The `=>` case is already right: exports.go reads Fat correctly, and
`{a => 1}` gives a Call with Fat set. That flag is how the fat comma's own
autoquoting is recorded, and the same reasoning applies here.

## Acceptance

- [ ] A bareword subscript is a Term, not a Call (`go test ./internal/parse/ -run '^TestBarewordSubscriptIsAString$' -count=1 -v | grep -q '^--- PASS: TestBarewordSubscriptIsAString'`)
- [ ] An explicit `$h{k()}` stays a Call (`go test ./internal/parse/ -run '^TestBarewordSubscriptIsAString$' -count=1 -v | grep -q '^--- PASS: TestBarewordSubscriptIsAString'`)
- [ ] The canon ratchet improves, measured in files (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)

Plan: M1. Spec: §4.14.
