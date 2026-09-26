---
title: "A bareword hash key that names a quote-like operator lexes as that operator: $h{m} swallows to the next brace"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T23:58:31.367571262Z
updated: 2026-09-26T23:58:31.367571262Z
---

A hash subscript whose key is a quote-like operator NAME lexes as that operator
with `}` as its delimiter, so `$h{m}` swallows to the next `}` and emits a
spurious `};`. Perl autoquotes the key.

Found by re-bucketing first-failure shapes at `34737c46` (205 dirty files, 186
distinct shapes) and isolating the recurring tail shapes.

## Measured at `34737c46`

    source                U   canon
    my $s = $t->{m};      1   my $s = $t->{m};};      <- the extra `};`
    my $s = $t->{q};      1   my $s = $t->{q};};
    my $s = $t->{s};      1   my $s = $t->{s};};
    my $s = $t->{y};      1   my $s = $t->{y};};
    my $s = $h{m};        1   my $s = $h{m};};        <- the ARROW is irrelevant
    $t->{m} = 1;          1   $t->{m} = 1;};          <- as an lvalue too

    my $s = $t->{'m'};    0   correct                 <- quoting the key fixes it
    my $s = $t->{mm};     0   correct                 <- TWO letters is fine
    my $s = $t->{x};      0   correct                 <- `x` is not affected
    my $s = $t->{a};      0   correct
    my $s = $t->{key};    0   correct
    my $s = $t->[0];      0   correct                 <- array subscript fine
    my $s = "x $t->{m} y"; 0  correct                 <- INTERPOLATED is fine

So the rule is: a subscript holding exactly one bareword that happens to name a
quote-like operator. `m`, `q`, `s`, `y` reproduce; `x` does not, because `x` is
the repeat operator rather than a quote-like one.

## Perl autoquotes it, always

    perl -MO=Deparse -e 'my %h=(m=>1,q=>2,s=>3,y=>4); print $h{m},$h{q},$h{s},$h{y};'
      ->  print $h{'m'}, $h{'q'}, $h{'s'}, $h{'y'};

    perl -MO=Deparse -e 'my %h; my $a=$h{tr}; my $b=$h{qq}; my $c=$h{qw};'
      ->  my $a = $h{'tr'};  my $b = $h{'qq'};  my $c = $h{'qw'};

`tr`, `qq` and `qw` autoquote too, so the fix must cover every quote-like name
and not just the four that happen to reproduce with a one-character body. The
lexer is reading `{m}` as a match whose delimiter is `}`: the body runs to the
NEXT `}`, which is why the emission gains one and the enclosing construct loses
its closer.

Note `"x $t->{m} y"` is already CORRECT, so whatever path handles interpolated
subscripts already knows this rule. That is the same shape as the caret variable
(`${^TAINT}` was right while `$^O` was wrong) and the trivia set diverging in
four copies: one spelling of a rule was repaired and its sibling left, because
nothing tied them together.

## Reach

Small in first-failure terms and larger in truth. Bucketing attributes it to a
handful of files, but bucketing counts each file once at its FIRST Unknown, and
`$h{s}`-shaped keys appear throughout perl's suite -- `$test->{m}` in
`op/sprintf2.t` is inside a parenthesised call to a RESOLVED sub, which is how it
was found. The `&$coderef` fix measured 22 files against an attributed 7 for the
same reason: an attributed reach is a floor.

## Scope

1. A subscript containing exactly one bareword is a STRING. Find where the lexer
   decides `m`/`q`/`s`/`y`/`tr`/`qq`/`qw` opens a quote-like operator and make it
   decline when the previous significant token is a subscript opener (`{` of a
   hash element) and the next is the matching `}`.
2. Measure which names are affected rather than assuming the four above: perl's
   quote-like set is `q qq qw qx m qr s tr y`. Check each as a bare key.
3. The INTERPOLATED path already gets this right. Find out why and reuse that
   decision rather than writing a second rule -- if the two answers can diverge,
   that is the defect one level up.
4. `$h{m}` vs `$h{ m }` (whitespace) and `$h{m => 1}` (a fat comma after) may
   differ. Measure before scoping.

## Context

- paths: `internal/lexer/scan.go` (`scanQuoteLike`), `internal/lexer/expect.go`
  (the term/operator decision), and whatever handles subscripts inside
  interpolation
- commands: `go test ./internal/lexer/ -count=1`,
  `go test ./internal/parse/ -count=1`
- related: `01a0df00` (the caret variable -- braced spelling right, bare one
  wrong, the same divergence), `01a0dd25`-era trivia duplication

## Acceptance criteria

- [ ] `$h{m}` parses with the key as a string (`go test ./internal/parse/ -run '^TestBarewordSubscriptKey$' -count=1 -v | grep -q '^--- PASS: TestBarewordSubscriptKey'`)
- [ ] Every quote-like name works as a bare key (`go test ./internal/parse/ -run '^TestBarewordSubscriptKey$' -count=1 | grep -q '^ok'`)
- [ ] A real match still lexes as a match (`go test ./internal/lexer/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the interpolated spelling stays correct (`go test ./internal/parse/ -count=1 | grep -q '^ok'`)
- [ ] REGRESSION GUARD, passes today: the corpus stays complete (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
