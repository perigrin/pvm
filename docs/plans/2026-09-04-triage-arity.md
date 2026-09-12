# Arity-mismatch triage: 53 diagnostics from perl5/t

**Summary: 49 false positives, 4 true positives.**

All false positives share one root cause: the tree-sitter Perl grammar parses a
paren-less list operator as `ambiguous_function_call_expression` holding only its
FIRST argument, and puts the remaining arguments outside the call node. The
existing `collectTrailingListArgs` in `/home/perigrin/dev/pvm/internal/infer/infer.go`
already reclaims those siblings, but only when the call is the FIRST element of
the enclosing `list_expression`. Six distinct syntactic contexts defeat that
guard; they are the six shapes below.

The 4 true positives are perl's own deliberate degenerate-arity tests
(`push @ary;` with no values), which perl itself warns about.

Method notes: every witness was checked with `perl -Mstrict -Mwarnings -c` and
parsed with `./psc parse --raw-markdown`. Witness files are in
`/tmp/claude-1001/-home-perigrin-dev-pvm/70fe3658-d868-4e80-81bb-03eaf4354f54/scratchpad/w/`.

---

## Shape 1: Nested call as a NON-FIRST list element (the known remaining gap)

- **Instances:** 16 of 53
- **Example:** `/home/perigrin/dev/perl5/t/op/bless.t:100` — `    local $SIG{__WARN__} = sub { push @w, join '', @_ };`
- **Verdict:** FALSE POSITIVE (grammar hid args)
- **Valid perl:** yes — `perl -Mstrict -Mwarnings -c w/a.pl` → `syntax OK`
- **Where the args went:** `./psc parse --raw-markdown w/a.pl` shows
  `ambiguous_function_call_expression(function "push", list_expression(array @w, ",", ambiguous_function_call_expression(function "join", string_literal ''), ",", array @a))`.
  The inner `join` call holds only `''`; `@a` is its SIBLING inside push's
  `list_expression`. `collectTrailingListArgs` refuses to claim it because
  `array @w` precedes the join call in that list — the same guard that stops
  `f(1), g(2)` from handing g's argument to f.
- **Suggested fix:** when the preceding sibling is separated from the call by a
  `,` and the call is a paren-less list operator (no `(` token child), let it
  claim the trailing siblings — a paren-less list operator is greedy to the end
  of the enclosing list in real Perl, so there is nothing after it that could
  belong to an earlier element.

Instances:
- `t/op/bless.t:100` — `push @w, join '', @_`
- `t/op/read.t:40` — `push @values, join "", map {chr $_} $_ .. $_ + 4;`
- `t/op/read.t:41` — `push @buffers, join "", map {chr $_} $_ + 5 .. $_ + 20;`
- `t/op/join.t:128` — `for(1,2) { push @_, \join "x", 1 }`
- `t/re/regexp.t:253` — `push @tests, join "\t", $expanded_pat,` (continues over 4 lines)
- `t/op/groups.t:291` — `diag_variable( g => join ',', @g );`
- `t/op/groups.t:292` — `diag_variable( ex_gr => join ',', @extracted );`
- `t/op/sprintf.t:146` — `ok(1, join ' ', grep length, ">$result<", $comment);`
- `t/op/coresubs.t:74` — `inlinable_ok($word, $args_for{$word} || join ",", map "\$$_", 1..$numargs);`
- `t/op/coresubs.t:117` — `: join ",", map "\$$_", 1..$numargs+5+(`
- `t/re/reg_fold.t:165` — `eval join ";\n","plan tests=>". (scalar @tests), @tests, "1"`
- `t/porting/podcheck.t:309` — `$file = lc join '/', File::Spec->splitdir($directories), $file;`
- `t/op/sort.t:143` — `diag "For code points " . join " ", @wrongly_non_utf8;`
- `t/re/regexp_unicode_prop.t:540` — `diag join "\n", @warnings, "\n";`
- `t/test.pl:843` — `join $sep, grep {...} split quotemeta ($sep), $1;`
- `t/io/print.t:72` — `printf "ok 22%n ...\n", substr $n,1,1;`

Note: `t/op/sort.t:143`, `t/porting/podcheck.t:309` and `t/re/regexp_unicode_prop.t:540`
are the same gap reached through a unary/named-op wrapper (`lc`, `diag`, `.`)
rather than a literal comma, but the mechanism is identical: something precedes
the paren-less call in the enclosing list.

---

## Shape 2: `\substr $s, 0, 1` — refgen wraps the call, args escape to the outer list

- **Instances:** 13 of 53
- **Example:** `/home/perigrin/dev/perl5/t/op/substr.t:600` — `    $r[$_] = \ substr $s, $_, 1 for (0, 1);`
- **Verdict:** FALSE POSITIVE (grammar hid args)
- **Valid perl:** yes — `perl -Mstrict -Mwarnings -c w/b.pl` → `syntax OK`
- **Where the args went:** `./psc parse --raw-markdown w/b.pl` shows
  `list_expression(assignment_expression(... refgen_expression("\\", ambiguous_function_call_expression(function "substr", scalar $str))), ",", number 0, ",", number 1)`.
  The call holds only `$str`; `0` and `1` are siblings of the whole
  `assignment_expression`. `collectTrailingListArgs` walks up through
  `assignment_expression` but **not** through `refgen_expression`, so the walk
  bails at `default: return nil` before reaching the `list_expression`.
- **Suggested fix:** add `refgen_expression` (and `unary_expression`) to the
  wrapper kinds `collectTrailingListArgs` walks up through — a one-line addition
  to the existing `switch` case list.

Instances:
- `t/op/substr.t:600` — `$r[$_] = \ substr $s, $_, 1 for (0, 1);`
- `t/op/substr.t:689` — `sub { $_[0] = 'dea' }->( scalar substr $str, 3, 2 );`
- `t/op/substr.t:782` — `my $y = \substr *foo, 0, 0;`
- `t/op/substr.t:785` — `$y = \substr *foo, 0, 0;`
- `t/op/substr.t:796` — `my $substr = \substr $refee, -2;`
- `t/op/substr.t:828` — `${\substr $refee, 0} = bless ["\x{100}"], o::;`
- `t/op/substr.t:838` — `is ${\substr %h, 0}, scalar %h, '\substr %h';`
- `t/op/substr.t:839` — `is ${\substr @a, 0}, scalar @a, '\substr @a';`
- `t/op/lex_assign.t:194` — `my $pvlv = \substr $str, 0, 1;`
- `t/op/state.t:335` — `state $c = \substr $tintin, $x, 1;`
- `t/op/tie_fetch_count.t:320` — `my $l   =\substr$var,0,1;`
- `t/op/utf8cache.t:68` — `my $l = \substr $x, 0;`
- `t/op/utf8cache.t:77` — `() = ord substr $_[0], 1, 1;` (same walk failure, wrapper is `ord`/`func1op`)

---

## Shape 3: Doubly-nested comparison/concatenation swallow (unwrap only goes one level)

- **Instances:** 9 of 53
- **Example:** `/home/perigrin/dev/perl5/t/op/signatures.t:437` — `sub t130 { join(",", @_).";".scalar(@_) }`
- **Verdict:** FALSE POSITIVE (grammar hid args)
- **Valid perl:** yes — `perl -Mstrict -Mwarnings -c w/m4.pl` → `syntax OK`
- **Where the args went:** this shape only appears when an unbalanced/unclosed
  construct later in the file (e.g. a bare `{` block, a signature the grammar
  cannot parse) pushes the parser into a fallback where a fully parenthesised
  `join(...)` is re-parsed as `ambiguous_function_call_expression`. Minimal
  witness `w/bb.pl` (`sub t130 { join(",", @_).";".scalar(@_) }` followed by a
  bare `{`) parses to
  `ambiguous_function_call_expression(function "join", binary_expression(binary_expression("(", list_expression("," , @_), ")", ".", ";"), ".", scalar(@_)))`.
  The real argument list is TWO `binary_expression` levels down.
  `unwrapSwallowedArgs` unwraps exactly one level and requires the first named
  child to be a `list_expression`; here it is another `binary_expression`, so it
  returns the single argument unchanged. The one-level version (`w/bb2.pl`,
  `join(",", @_).";"`) is handled correctly and reports nothing — confirming the
  limit is unwrap depth, not the shape itself.
- **Suggested fix:** make `unwrapSwallowedArgs` recurse (descend through
  successive `comparisonKinds` nodes until it finds the `list_expression`)
  instead of unwrapping a single level.

Instances (all reproduce only in whole-file context, never line-isolated):
- `t/op/signatures.t:437,602,727,784,796,808,876` (7)
- `t/perf/opcount.t:516` — `. join('',  @args)` inside a multi-line concatenation
- `t/re/fold_grind.pl:1096` — `. ', target="' . join("", @x_target) . '",'`
- `t/cmd/subval.t:85,184` — `print join(':',&ary1) eq '1:2:3' ? ... : ...;`

(Note: subval.t's two are the `eq`-inside-conditional variant of the same
two-level nesting: `conditional_expression(equality_expression(call, str), ...)`
with the call's list one level deeper than `unwrapSwallowedArgs` looks.)

---

## Shape 4: Paren-less call whose sibling args sit past a comparison operator

- **Instances:** 4 of 53
- **Example:** `/home/perigrin/dev/perl5/t/re/regexp.t:354` — `                    && substr($pat, $i+1, 1) eq '?'`
- **Verdict:** FALSE POSITIVE (grammar hid args)
- **Valid perl:** yes — `perl -Mstrict -Mwarnings -c w/l.pl` → `syntax OK`
- **Where the args went:** the same comparison-swallow that `unwrapSwallowedArgs`
  was written for, but reached through an extra `&&`/`binary_expression` level,
  so the single-level unwrap misses it. Isolated (`w/l.pl`, `1 && substr(...) eq '?'`)
  it reports nothing; in `regexp.t`'s multi-line `&&` chain it reports.
- **Suggested fix:** same recursion fix as Shape 3 — one change covers both.

Instances:
- `t/re/regexp.t:354` — `&& substr($pat, $i+1, 1) eq '?'`
- `t/re/regexp.t:391` — `&& substr($pat, $i+1, 1) eq '^'`
- `t/re/pat.t:129` — `my $e = index ($_, 'not') >= 0 ? '' : 1;`
- `t/op/multideref.t:145` — `::ok(!exists +($r//0)->[$li1]{$lk1}[$li2+$z]{$lk4},'!exists: general');`

For `multideref.t:145` the parse shows `unary_expression("!", func1op_call_expression(exists))`
with the `+($r//0)->...` chain becoming the RIGHT operand of a `binary_expression`
whose operator is the unary `+` — `exists` ends up with zero children. The
leading unary `+` disambiguator is the trigger; perl reads it as a no-op prefix,
the grammar reads it as infix addition.

---

## Shape 5: Heredoc passed as a call argument — body parsed as live code

- **Instances:** 1 of 53
- **Example:** `/home/perigrin/dev/perl5/t/io/perlio.t:229` — `unshift @INC, sub {`
- **Verdict:** FALSE POSITIVE (grammar hid args — and should not be analyzed at all)
- **Valid perl:** yes — the line is inside a `<<'EOP'` heredoc body passed to
  `fresh_perl_like`; `perl -Mstrict -Mwarnings -c w/hd.pl` → `syntax OK`
- **Where the args went:** `./psc parse --raw-markdown w/hd.pl` shows the grammar
  does not support a heredoc as a call argument: it emits
  `ERROR "("` for the call's open paren and then parses the heredoc BODY as
  ordinary statements. The `unshift @INC, sub {...}` inside becomes a real
  `ambiguous_function_call_expression` preceded by a `binary_expression` in the
  enclosing `list_expression`, so the first-element guard declines the `sub {}`
  sibling.
- **Suggested fix:** unclear — the real fix is grammar-level heredoc-as-argument
  support; short of that, suppressing diagnostics inside subtrees that contain an
  `ERROR` node would drop this one and reduce noise generally.

---

## Shape 6: Genuine one-argument `push`/`unshift` — perl's own degenerate-case tests

- **Instances:** 4 of 53
- **Example:** `/home/perigrin/dev/perl5/t/op/tiearray.t:211` — `    my $got = push @ary;            # this didn't used to call PUSH at all`
- **Verdict:** TRUE POSITIVE (really wrong — deliberately so)
- **Valid perl:** yes but warned — `perl -c w/tp.pl` prints
  `Useless use of push with no values at tp.pl line 3.` and
  `Useless use of unshift with no values at tp.pl line 4.`
- **Where the args went:** nowhere. `./psc parse --raw-markdown w/d.pl` shows
  `ambiguous_function_call_expression(function "push", array @ary)` — one
  argument, faithfully. PSC is right.
- **Suggested fix:** none needed. These four are correct diagnostics. Perl's
  suite wraps three of them in `no warnings 'syntax'` precisely because they are
  testing the degenerate no-values case. If the noise matters, PSC could match
  perl and downgrade zero-value `push`/`unshift` to a warning rather than an
  error, but the detection is sound.

Instances:
- `t/op/tiearray.t:211` — `my $got = push @ary;`
- `t/op/tiearray.t:215` — `$got = unshift @ary;`
- `t/op/unshift.t:14` — `$count3 = unshift (@array);` (inside `no warnings 'syntax'`)
- `t/op/unshift.t:49` — `unshift (@alpha);` (inside `no warnings 'syntax'`)

---

## Also examined, NOT a defect

`/home/perigrin/dev/perl5/t/op/substr.t:905` and `:908` — `substr(ta_tindex(), 0, 2)`.
These reproduce only in the whole file, never line-isolated (`w/sf.pl` is clean),
and they are downstream of the same Shape 3 fallback parse triggered elsewhere in
substr.t. Counted under Shape 3's whole-file-context group. They are the only two
lines in the set that perl itself rejects (`substr` outside of string on a
function call is not an lvalue), but PSC is not reporting that — it is reporting
a spurious arity count.

---

## Recommended fix order (by instances retired)

1. **Recurse in `unwrapSwallowedArgs`** — retires Shapes 3 and 4 (13 instances)
   for one small change.
2. **Add `refgen_expression` / `unary_expression` to the wrapper walk in
   `collectTrailingListArgs`** — retires Shape 2 (13 instances), also one small
   change.
3. **Relax the first-element guard for paren-less list operators** — retires
   Shape 1 (16 instances). This is the known gap and the largest single group,
   but it is the change that most needs a regression test proving `f(1), g(2)`
   still counts correctly.
4. **Suppress diagnostics under `ERROR` subtrees** — retires Shape 5 (1) and is
   cheap insurance against every future fallback-parse artifact.
5. **Leave Shape 6 alone** — 4 correct diagnostics.
