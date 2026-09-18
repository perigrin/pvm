---
title: "Node carries one flag and the tree needs four: Arrow, Paren, Fat, Handle"
state: done
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-17T14:22:29.542545808Z
updated: 2026-09-18T10:58:16.16266301Z
sessions:
- start_sha: 6968d55df55015b70b6730ae5781e9eff4b0c52b
  end_sha: e0d124d1932fa7a5443eb3f763146c1a6171decc
  commits: 1
  started_at: 2026-09-18T10:48:08.065893001Z
  ended_at: 2026-09-18T10:58:16.16266301Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-18T10:48:08.065893001Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-18T10:58:16.16266301Z
observed_paths:
- internal/parse/call.go
- internal/parse/expr.go
- internal/parse/flags_test.go
- internal/parse/parse.go
- internal/parse/term.go
---

`parse.Node` has one flag and §4.14 needs four. The lowering cannot build what the parser threw away, and four separate obligations share this single cause.

## The cause

    type Node struct {
        Kind       Kind
        Start, End int
        Children   []*Node
        Text       string
        Resolved   bool    <- the only flag
    }

`internal/parse/parse.go:213-234`. `Resolved` is documented "only meaningful
for Call; false everywhere else and not read there" — so the struct already
carries a kind-specific bit, and the pattern is established.

Four §4.14.2 rows are PARSER *because of this field list*, not because the
parser cannot see the fact. In each case the parser knows the answer at
construction time and has nowhere to write it down.

## The four

**Arrow** (§4.14.2, rule 2). `$h{k}` and `$h->{k}` are byte-identical trees
after d2ebe02a:

    $h{k};      [source_file statement index term call]
    $h->{k};    [source_file statement index term call]

`internal/parse/expr.go:73` already promises the flag in a comment — "an
`Arrow` flag, not two unrelated shapes" — and the struct cannot honour it.
Rule 2 is why it matters: the two forms read DIFFERENT VARIABLES. Measured on
perl 5.42.0:

    %h = (k => 'hash'); $h = {k => 'ref'};
    print $h{k};     # hash
    print $h->{k};   # ref

25.3% of T1's index nodes (553 of 2,186) are the arrow form.

**Paren** (§4.14.2, rule 1). `($x);` parses to `[term "$x"]` — `finishList`
at `internal/parse/term.go:288` returns the inner node with a widened span.
§4.14 marks `Paren` "KEEP THIS" because §4.10 and §4.12.2 both turn on it:

    ('a') x 3   vs  'a' x 3        list repeat vs string repeat
    my ($x)=f() vs  my $x=f()      list vs scalar context

Span re-derivation is not a substitute: the census found 721 T1 nodes whose
span starts `(` with a first child at the same offset, so the naive test has
721 false positives and the guarded test is unproven.

**Fat** (§4.14.2, `List{Fat}`). Inside a list the fat comma is destroyed —
`term.go:275-276` consumes `,` and `=>` alike:

    (a => 1);   [source_file statement list call term]
    (a, 1);     [source_file statement list call term]

Identical. §4.5.4: `=>` quotes the word to its left, so without the bit the
lowering cannot tell a string from a call. It survives in call arguments
(`f(a => 1)` keeps a `binary "=>"`), which makes the loss inconsistent as
well as lossy.

**Handle** (§4.14.2, `Call.Handle`). `parseFilehandleSlot`
(`internal/parse/call.go:167`) decides the slot at construction and emits a
plain `Term`, so the lowering must re-derive it from child shape:

    print $fh "x";    [call "print" [term "$fh"] [term "\"x\""]]   two children
    print $fh, "x";   [call "print" [binary "," ...]]              one child

`internal/infer/infer.go:1155-1165` records what this cost already: counting
the handle as argument 1 made every typed-handle `print` a false `Str`
mismatch.

## Scope

Add the fields and set them where the parser already knows. Do NOT change
what parses — this is a shape change, and d2ebe02a is the precedent: it
fixed subscript shape at exactly zero effect on the corpus (856 clean /
8,432 Unknown before and after on the then-current sweep; T1 unchanged at
408/4,944). A shape fix that moves the parse rate has changed what parses,
which is a different piece of work.

Not in scope: `$::x`, `qx//`, `<<>>`, signatures, statement-position
`sub {}`. Those are parser gaps with their own causes.

## Representation

One `bool` per fact is the obvious shape and matches `Resolved`. A bitfield
is the alternative. Decide it in the issue rather than by default, and record
which in the code — four bools on every node in a 272,455-node tree is 272KB
of padding that a `uint8` would not spend. §6.10.4 owns the memory budget.

## Acceptance

- [x] `$h{k}` and `$h->{k}` differ by exactly one field (`go test ./internal/parse/ -run '^TestArrowBitDistinguishesSubscripts$' -count=1 -v | grep -q '^--- PASS: TestArrowBitDistinguishesSubscripts'`)
- [x] `($x)` is distinguishable from `$x` without re-deriving from spans (`go test ./internal/parse/ -run '^TestParenBitSurvives$' -count=1 -v | grep -q '^--- PASS: TestParenBitSurvives'`)
- [x] `('a') x 3` and `'a' x 3` are distinguishable, and `my ($x) = f()` from `my $x = f()` (`go test ./internal/parse/ -run '^TestParenBitCarriesContext$' -count=1 -v | grep -q '^--- PASS: TestParenBitCarriesContext'`)
- [x] `(a => 1)` and `(a, 1)` differ, inside lists, anon arrays and anon hashes (`go test ./internal/parse/ -run '^TestFatBitSurvivesInAggregates$' -count=1 -v | grep -q '^--- PASS: TestFatBitSurvivesInAggregates'`)
- [x] `print $fh "x"` marks its handle without the lowering counting children (`go test ./internal/parse/ -run '^TestHandleBitIsSet$' -count=1 -v | grep -q '^--- PASS: TestHandleBitIsSet'`)
- [x] Each bit is false where it does not apply, and no consumer reads it there (`go test ./internal/parse/ -run '^TestFlagsAreFalseWhereInapplicable$' -count=1 -v | grep -q '^--- PASS: TestFlagsAreFalseWhereInapplicable'`)
- [x] Round-trip is unaffected (`go test ./internal/parse/ -run '^TestTreeRoundTrips$' -count=1 -v | grep -q '^--- PASS: TestTreeRoundTrips'`)
- [x] T1 does not move: 408 clean, 4,944 Unknown, measured on `unit/*.t` at top level (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [x] The whole suite is green (`go test ./internal/... 2>&1 | grep -qv FAIL`)

The T1 AC uses `TestParseRatchet`, committed at 4bc971ec. Until it exists this issue cannot be completed — correctly, since
"changed no parse outcome" is the claim that needs the ratchet to mean
anything.

Spec: §4.14.2, §4.14.3, rules 1 and 2. Plan: M1.


## Done, 2026-09-18 (e0d124d1)

    T1      434 clean / 4,606 Unknown, unmoved
    T2      16 of 56, unmoved
    oracle  40 exact, 1 WRONG (the M4 pin), unmoved

Unmoved is the expected result: this adds fields to trees that already
parsed. A change here that moved the parse rate would mean it had changed
what parses.

BOOLS, NOT A BITFIELD, and the choice was measured rather than defaulted.
`unsafe.Sizeof`: Node with one bool is 72 bytes, with five bools 72, with a
`bool + uint8` bitfield 72. The flags land in padding the struct already
had, so the bitfield costs the same and reads worse.

## A latent bug fixed in passing

`finishList` copied Kind, Text and Children into its replacement node and
dropped everything else -- so `($h->{k})` would have lost its Arrow the
moment Arrow existed. It now carries every flag across.

## One AC was stale when I started

The last criterion read "T1 does not move: 408 clean, 4,944 Unknown". Those
were the numbers when this issue was filed; two commits since
(f905b4e9, 6968d55d) moved them to 434 / 4,606. The criterion is satisfied
in spirit -- T1 did not move ACROSS THIS CHANGE -- and the ratchet enforces
it against the current baseline rather than a quoted one, which is why the
ratchet exists.

Handle is set at the call site rather than in parseFilehandleSlot's three
branches: one place that cannot be forgotten when a fourth handle form
arrives.

All four mutation-checked: disabling any one fails its own test and only
its own.
