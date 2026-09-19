---
title: A bareword hash subscript is a string, not a call
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T12:19:07.059614571Z
updated: 2026-09-19T12:44:18.034166879Z
sessions:
- start_sha: 4bb5af573f0937de0b63a88bc5e43954d31f75cd
  end_sha: ""
  commits: 0
  started_at: 2026-09-19T12:24:37.607130755Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T12:24:37.607130755Z
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

- [x] A bareword subscript is a Term, not a Call (`go test ./internal/parse/ -run '^TestBarewordSubscriptIsAString$' -count=1 -v | grep -q '^--- PASS: TestBarewordSubscriptIsAString'`)
- [x] An explicit `$h{k()}` stays a Call (`go test ./internal/parse/ -run '^TestBarewordSubscriptIsAString$' -count=1 -v | grep -q '^--- PASS: TestBarewordSubscriptIsAString'`)
- [x] The canon ratchet improves, measured in files (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)

Plan: M1. Spec: §4.14.


## Done, 2026-09-19 (657d0caf)

    T1 Call nodes   29,757 -> 28,695
    T1 unresolved   13,784 -> 12,759
    full suite green, 0 FAIL across the repo

## The cost was not the emission

Call{Resolved:false} is how §4.8.3 records "a call to something I have not
seen", and sites.go scores that set as the parser's hedge. A hash key is not
a call at all, so 1,025 T1 entries were false -- inflating exactly the number
01a0b835 measured when the M1 subject reported 87 WRONG against tree-sitter's
2. Whether it moves that measurement is 01a0b835's to re-run.

## The rule is narrower than the issue assumed

A SLICE is not autoquoted, which I had wrong in the first draft of the test:

    $ perl -e 'sub a { "z" } my %h=(a=>1,b=>2); my @s=@h{a,b};
               print join(",", map { $_ // "undef" } @s), "\n";'
    undef,2

`a` called a() and looked up "z"; `b` had no sub and autoquoted. So the rule
belongs to a lone word filling a BRACE subscript, not to braces generally.
Testing the node gives that for free: a comma leaves a Binary there, which is
not a Call.

Array subscripts are not autoquoted either -- `$a[k]` with `sub k { 1 }`
gives element 1.

## The span, not the child count

`$h{k()}` has no children either, because an empty argument list leaves none.
My first attempt used `len(Children) > 0` and the negative rows of
TestBarewordSubscriptIsAString caught it. An explicit call's node covers its
parentheses; a bareword covers only the word.

## The ratchet did not move, and that is correct

`$h->{k()}` was a canon fixpoint before and `$h->{k}` is one after -- the
tree changed, the emission's stability did not. Acceptance criterion 3
assumed otherwise. TestAutoquoteShrinksTheUnresolvedSet holds the real gain
instead, with a one-sided bound so a later change that legitimately resolves
more calls does not fail it.
