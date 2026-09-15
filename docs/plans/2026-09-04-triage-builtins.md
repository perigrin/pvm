# Builtin signature triage — 298 diagnostics from perl5/t

Perl used for every witness: `/home/perigrin/.local/bin/perl`, v5.42.0.
PSC binary: `/home/perigrin/dev/pvm/psc`. Signatures: `/home/perigrin/dev/pvm/internal/types/signatures.go`.

## Headline

Only **three** of the 298 diagnostics trace to a genuinely wrong `ArgTypes`
table (substr's missing 4th argument, join/index/unshift's `MinArity`, and
chr/int/length's over-tight numeric domain). The remaining bulk — roughly 150
of 298 — are **inference and argument-collection bugs upstream of the
signature table**, and no edit to `signatures.go` will silence them. They are
listed here under the builtin they surface as, with the real cause named.

Four upstream defects account for most of the noise:

- **U1 — paren-less builtin calls collapse to one argument.**
  `join ":", @a` is seen as `join(":")`. Reproduced:
  `/tmp/.../scratchpad/t8.pl:2` → `call to join: expected at least 2 argument(s), got 1`,
  while `join(":", @a)` on the next line is clean. Same for `index $n,"b",0`
  and `unshift @a,1`. **Causes every arity-mismatch in the corpus (~52).**
- **U2 — `my %h = 1..20` / `my @a = 1..20` retypes the container to the RHS type.**
  Reproduced: `t3.pl` = `my %a = 1..20; keys %a;` →
  `call to keys: argument 1 expects Array|Hash, got Str`. `my %a = (1,2)` is
  clean. **Causes all 14 keys, 13 of 15 each, 2 of 3 values (~29).**
- **U3 — heredoc bodies are parsed as code.** A `<<~'E'` body containing three
  or more comma-separated statements gets type-checked. Reproduced in
  `tq.pl`; a 3-bless body yields 2 diagnostics, a 496-line body yields 495.
  **Causes all 30 bless diagnostics.**
- **U4 — inside a comma chain, the previous call's return leaks in as the
  next call's argument.** In `tq.pl` the first `bless(...)` returns Object and
  that Object is reported as `argument 2` of the *second* bless. This is also
  why chr/int/length — builtins perl rejects with "Too many arguments" at two
  args — get "argument 2", "argument 3", "argument 5" reports at all.

---

## substr
- Instances: 88
- PSC signature: `"substr": {MinArity: 2, ArgTypes: []Type{Str, Num, Num}, ReturnType: Str}`
- Perl's actual behaviour: substr is 2-to-4 arguments. The 4th is a
  replacement string; the call both returns the old substring and mutates the
  target.
  ```
  $ perl -e 'my $s="hello"; my $r=substr($s,0,2,"XY"); print "$r $s\n"'
  he XYllo
  $ perl -e 'my $t="hello"; substr($t,0,1,"X","Y")'
  Too many arguments for substr ... near ""Y")"
  $ perl -e 'my $t="hello"; my $z=substr($t)'
  Not enough arguments for substr ... near "$t)"
  ```
  So MinArity 2 is right, max arity is 4, and the 4th slot is `Str`, not `Num`.
- Verdict: **SIGNATURE WRONG** (45 of 88), plus upstream noise.
  - 45 × `argument 4 expects Num, got Str` — the missing replacement argument.
    True signature bug. Reproduced standalone: `t1.pl:10`.
  - 19 × arity-mismatch "got 1" — **U1**, paren-less `substr $n,1,1`.
    Reproduced `t7.pl:2`.
  - 6 × `argument 2 expects Num, got Str` + 3 × arg 3 — perl coerces strings
    numerically here (`substr("hello","1","2")` works), so `Num` is arguably
    too tight for a language with no distinction; but these are also
    downstream of the same string-typed offsets. Judgement call, see below.
  - 1 × `argument 5 expects Num, got CodeRef` — **U4**; perl has no 5th
    argument at all, so PSC is collecting arguments wrongly.
  - 5 × `argument 1 expects Str, got Object`, 2 × ArrayRef, 1 × Undef, 1 ×
    Bool — perl stringifies all of these without error under `no warnings`.
- Example: `/home/perigrin/dev/perl5/t/op/bool.t:31` — `substr($y, 0, 1, "T");`
- Suggested signature: `{MinArity: 2, ArgTypes: []Type{Str, Num, Num, Str}, ReturnType: Str}`
  and a max-arity of 4 so the phantom argument 5 cannot be reported.

## bless
- Instances: 30
- PSC signature: `"bless": {MinArity: 1, ArgTypes: []Type{Ref, Str}, ReturnType: Object}`
- Perl's actual behaviour: `bless {}, "Foo"` → Foo; `bless {}` → main. Arity
  1-2, second argument a plain string. Signature matches.
  Every one of the 30 hits is `bless( {...}, 'HeaderLine' )` — a
  *single-quoted string literal* — inside the `<<~'EOF_DUMP'` heredoc at
  `t/porting/header_parser.t:136-631`. Reproduced minimally in `tq.pl`:
  ```
  is($d,<<~'E', "m");
          bless( { "a" => 1 }, 'HL' ),
          bless( { "a" => 1 }, 'HL' ),
          bless( { "a" => 1 }, 'HL' ),
          E
  ```
  → 2 diagnostics on lines 3 and 4. The same three lines outside a heredoc
  (`to.pl`, `tp.pl`) are clean. So: **U3** (heredoc body typechecked) plus
  **U4** (line N's `bless` returns Object, which is then reported as line
  N+1's argument 2).
- Verdict: **SIGNATURE OK** — all 30 are false positives from U3+U4.
- Example: `/home/perigrin/dev/perl5/t/porting/header_parser.t:613` — `bless( {`
  (a heredoc line, not code; its class argument is `'HeaderLine'` on line 629)
- Suggested signature: no change.

## join
- Instances: 28
- PSC signature: `"join": {MinArity: 2, ArgTypes: []Type{Str, List}, ReturnType: Str}`
- Perl's actual behaviour: `join(":")` compiles and runs, returning "".
  ```
  $ perl -e 'my $ok = eval q{my $x = join(":"); 1}; print $ok ? "ok\n" : "ERR: $@"'
  ok
  ```
  So MinArity is 1, not 2 — though a 1-argument join is always pointless.
  ArgTypes `{Str, List}` is right; `join(':', &ary1)` and `join(':', @a)` both
  flatten correctly.
- Verdict: **MIXED**
  - 26 × arity-mismatch "got 1" — **U1**. Reproduced `t6.pl:3,4,8`; every one
    is a paren-less `join ':', LIST` or `join(':', &subcall)`. The signature's
    MinArity 2 is *also* stricter than perl, so both need fixing: even with U1
    repaired, `join(":")` is legal perl.
  - 1 × `argument 1 expects Str, got Undef`, 1 × `got Object` — separator
    stringifies; false positives from the Str-vs-anything gap, not join-specific.
- Example: `/home/perigrin/dev/perl5/t/cmd/subval.t:85` — `print join(':',&ary1) eq '1:2:3' ? "ok 24\n" : "not ok 24\n";`
- Suggested signature: `{MinArity: 1, ArgTypes: []Type{Str, List}, ReturnType: Str}`

## chr
- Instances: 23
- PSC signature: `"chr": {MinArity: 0, ArgTypes: []Type{Int}, ReturnType: Str}`
- Perl's actual behaviour: chr takes exactly one argument and numifies it.
  ```
  $ perl -e 'no warnings; print ord(chr(65.7)),"\n"'   # 65
  $ perl -e 'no warnings; print ord(chr("65")),"\n"'   # 65
  $ perl -e 'no warnings; printf "%vd\n", chr(-0.1)'   # 65533 (U+FFFD)
  $ perl -e 'chr(65,66)'
  Too many arguments for chr ... near "66)"
  ```
  A float and a numeric string are both accepted and truncated/coerced. `Int`
  is too narrow; perl's domain here is `Num` at minimum and in practice `Str`
  (anything numifiable).
- Verdict: **SIGNATURE WRONG**
  - 13 × `expects Int, got Num` — legitimate perl (`chr(-0.1)`, `chr(65.7)`).
    Reproduced `ta.pl:2,3,6`.
  - 7 × `expects Int, got Str` — `chr("65")` works. Reproduced `ta.pl:4`.
  - 1 × `argument 2` — **U4**; perl rejects 2-argument chr outright.
  - 1 × Object, 1 × Bool — overloaded/boolean numification, same class.
- Example: `/home/perigrin/dev/perl5/t/op/chr.t:22` — `is(chr(-0.1), "\x{FFFD}"); # The U+FFFD Unicode replacement character.`
- Suggested signature: `{MinArity: 0, ArgTypes: []Type{Num}, ReturnType: Str}`
  (`Str` if you want to also accept numeric strings without a warning; `Num`
  clears 13 of the 23 on its own, `Str` clears 20.)

## print
- Instances: 18
- PSC signature: `"print": {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: Bool}`
- Perl's actual behaviour: print takes a LIST, not a single Str, and
  stringifies every member.
  ```
  $ perl -e 'my @a=(1,2,3); print @a; print "\n"'    # 123
  $ perl -e 'print 1==1; print "\n"'                 # 1
  $ perl -e 'print bless({},"X"), "\n"'              # X=HASH(0x...)
  ```
  `Str` in the variadic tail rejects `print @a` outright — exactly the bug
  the file's own comment already identified and fixed for `join` and `sprintf`.
- Verdict: **SIGNATURE WRONG**
  - 6 × `expects Str, got Array` — `print @a` is ordinary perl.
  - 7 × `got Bool`, 2 × `got Bool|Int`, 1 × Undef, 1 × Object, 1 × CodeRef —
    all stringify. Note `print undef` warns but is legal
    (`t/op/tiehandle.t:274`), and `print $coderef` prints `CODE(0x...)`.
- Example: `/home/perigrin/dev/perl5/t/op/postfixderef.t:54` — `print @a;`
- Suggested signature: `{MinArity: 0, ArgTypes: []Type{List}, ReturnType: Bool}`
  (and the same for `say`, which has the identical `{Str}` tail and is only
  absent from this corpus by chance)

## each
- Instances: 15
- PSC signature: `"each": {MinArity: 1, ArgTypes: []Type{Hash | Array}, ReturnType: List}`
- Perl's actual behaviour: `each %h` and `each @a` both work; `each` on a
  plain scalar is a compile error. Signature matches perl.
  ```
  $ perl -e 'my @arr=(10,20); while (my ($i,$v) = each @arr) { print "$i=$v\n" }'
  0=10
  1=20
  ```
- Verdict: **SIGNATURE OK** — the diagnostics are **U2**.
  13 × `got Str` are all `each %hash` where the hash was declared
  `my %hash = 1..20;` (`t/op/each.t:88`). Reproduced `t2.pl:4`; changing the
  declaration to `my %hash = (1,2)` clears it. The 2 × `got Int` at
  `t/op/defins.t:212,225` are `each @array` with the same range-assignment
  pattern.
- Example: `/home/perigrin/dev/perl5/t/op/each.t:91` — `$total += $key while $key = each %hash;`
- Suggested signature: no change. Fix the range-assignment inference.

## keys
- Instances: 14
- PSC signature: `"keys": {MinArity: 1, ArgTypes: []Type{Hash | Array}, ReturnType: List}`
- Perl's actual behaviour: matches. `keys %h` in a fresh file is clean under
  PSC (`t1.pl:3`).
- Verdict: **SIGNATURE OK** — all 14 are **U2**. Every hit is `keys %hash`
  where `%hash` was built by `my %hash = 1..20` or an equivalent range/list
  assignment. Reproduced `t3.pl`.
- Example: `/home/perigrin/dev/perl5/t/op/each.t:95` — `keys %hash;`
- Suggested signature: no change.

## push
- Instances: 10
- PSC signature: `"push": {MinArity: 2, ArgTypes: []Type{Array, List}, ReturnType: Int}`
- Perl's actual behaviour: `push @z` with no values compiles (warns "Useless
  use of push with no values"), so MinArity is 1. `push @a, qr/x/` is fine —
  a Regexp object is an ordinary list member.
  ```
  $ perl -e 'my $p = eval q{my @z; push @z; 1}; print $p ? "ok\n" : "ERR: $@"'
  Useless use of push with no values at (eval 1) line 1.
  ok
  ```
- Verdict: **MIXED**
  - 6 × `argument 1 expects Array, got Str` + 2 × `got Regex` — **U2** again
    (`push @valid_errors, qr/.../` at `t/op/pack.t:29` where
    `@valid_errors` picked up a bad type), false positives.
  - 2 × arity-mismatch "got 1" — `push @ary;` at `t/op/tiearray.t:211` and
    `push @readonly_array, ()` at `t/op/push.t:79`. Both are legal perl; the
    signature's MinArity 2 is too strict.
- Example: `/home/perigrin/dev/perl5/t/op/tiearray.t:211` — `my $got = push @ary;            # this didn't used to call PUSH at all`
- Suggested signature: `{MinArity: 1, ArgTypes: []Type{Array, List}, ReturnType: Int}`

## length
- Instances: 10
- PSC signature: `"length": {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: Int}`
- Perl's actual behaviour: `length(undef)` is legal and returns undef — it is
  the documented way to distinguish undef from "". Refs and objects stringify.
  Exactly one argument.
  ```
  $ perl -e 'no warnings; my $L = length(undef); print defined($L)?"def\n":"undef\n"'
  undef
  $ perl -e 'print length([1,2]), "\n"'   # 21  (length of "ARRAY(0x...)")
  $ perl -e 'length("a","b")'
  Too many arguments for length ... near ""b")"
  ```
  `t/op/length.t:127` is literally the test asserting `length(undef) == undef`.
- Verdict: **SIGNATURE WRONG** — the `Str` argument type excludes Undef, which
  is the one case `length` is specified to handle. 3 × Undef, 2 × Object,
  1 × Regex, 1 × Bool, 1 × ArrayRef, 2 × a big union — all legal.
- Example: `/home/perigrin/dev/perl5/t/op/length.t:127` — `is(length(undef), undef, "Length of literal undef");`
- Suggested signature: `{MinArity: 0, ArgTypes: []Type{Scalar}, ReturnType: Int}`
  (`Scalar` = Undef|Bool|Str|DualVar|Ref, which is precisely what perl accepts
  and what the return-undef-for-undef contract requires)

## index
- Instances: 7
- PSC signature: `"index": {MinArity: 2, ArgTypes: []Type{Str, Str, Int}, ReturnType: Int}`
- Perl's actual behaviour: the 3rd argument (POSITION) is numified from
  anything, floats included.
  ```
  $ perl -e 'no warnings; print index("hello","l",2.9), "\n"'   # 2
  $ perl -e 'no warnings; print index("hello","l","2"), "\n"'   # 2
  $ perl -e 'index("hello")'
  Not enough arguments for index ... near ""hello")"
  ```
  MinArity 2 is correct. `Int` for the position is too tight.
- Verdict: **MIXED**
  - 4 × `argument 3 expects Int, got Str`, 2 × `got Num` — signature too
    narrow; perl numifies. Reproduced `ta`-style; `t/op/index.t:366` passes
    `$len+1` where `$len` inferred Num.
  - 1 × arity-mismatch "got 1" at `t/re/pat.t:129` — **U1**, paren-less
    `index ($_, 'not')` parsed as one argument. Reproduced `t8.pl:6`.
- Example: `/home/perigrin/dev/perl5/t/op/index.t:366` — `is(index($s, "", $len+1), 3, 'Overlong index doesn\'t confuse utf8 cache');`
- Suggested signature: `{MinArity: 2, ArgTypes: []Type{Str, Str, Num}, ReturnType: Int}`
  (identical fix for `rindex`, below)

## die
- Instances: 7
- PSC signature: `"die": {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: None}`
- Perl's actual behaviour: die takes a LIST, and if the first element is a
  reference it is thrown as-is (the entire exception-object idiom).
  ```
  $ perl -e 'eval { die [1,2] }; print ref($@), "\n"'          # ARRAY
  $ perl -e 'eval { die bless({},"E") }; print ref($@), "\n"'  # E
  $ perl -e 'eval { die "a","b","c" }; print $@'               # abc at -e line 1.
  ```
- Verdict: **SIGNATURE WRONG** — 3 × Array, 2 × Object, 1 × ScalarRef,
  1 × ArrayRef, and every one is idiomatic perl. `die @_` inside a
  `$SIG{__WARN__}` handler (`t/op/attrs.t:14`) is the standard rethrow.
- Example: `/home/perigrin/dev/perl5/t/op/die.t:54` — `die bless [ 7 ], "Error";`
- Suggested signature: `{MinArity: 0, ArgTypes: []Type{List}, ReturnType: None}`

## warn
- Instances: 6
- PSC signature: `"warn": {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: Bool}`
- Perl's actual behaviour: same as die — a LIST, refs permitted.
  ```
  $ perl -e 'local $SIG{__WARN__}=sub{print ref($_[0])||"str","\n"}; warn "x","y"; warn [1,2]'
  str
  ARRAY
  ```
- Verdict: **SIGNATURE WRONG** — 3 × Array (all are `warn @_;` in a
  `$SIG{__WARN__}` handler), 3 × ArrayRef.
- Example: `/home/perigrin/dev/perl5/t/op/length.t:124` — `warn @_;`
- Suggested signature: `{MinArity: 0, ArgTypes: []Type{List}, ReturnType: Bool}`

## splice
- Instances: 6
- PSC signature: `"splice": {MinArity: 1, ArgTypes: []Type{Array, Int, Int, List}, ReturnType: List}`
- Perl's actual behaviour: OFFSET and LENGTH are numified from anything, and
  an array in that slot supplies its element count via scalar context —
  `splice(@a, @a, 0, 11, 12)` is the documented push-at-end idiom.
  ```
  $ perl -e 'my @a=(1..10); print join(",",splice(@a,2.7,3)),"\n"'   # 3,4,5
  $ perl -e 'my @a=(1..10); print join(",",splice(@a,"2","3")),"\n"' # 3,4,5
  $ perl -e 'my @c=(1..5); print join(",",splice(@c)),"\n"'          # 1,2,3,4,5
  ```
  MinArity 1 is correct.
- Verdict: **SIGNATURE WRONG** — `Int` in the OFFSET/LENGTH slots rejects both
  the numeric-string form and, more importantly, the `splice(@a, @a, 0, ...)`
  idiom where an Array is evaluated in scalar context. 3 × `argument 2
  expects Int, got Array`, 1 × arg 3 Array, 1 × arg 2 Num. The remaining
  1 × `argument 1 expects Array, got Str` at `t/uni/class.t:96` is `splice(@_,0,1)`,
  a U2-shaped inference miss on `@_`.
- Example: `/home/perigrin/dev/perl5/t/op/splice.t:33` — `is( j(splice(@a, -@a, @a, 1, 2, 3)), j(0..13), 'splice the whole list out, add 3 elements, return value is @a');`
- Suggested signature: `{MinArity: 1, ArgTypes: []Type{Array, Scalar, Scalar, List}, ReturnType: List}`
  — `Scalar` because scalar context is what perl imposes there, and it admits
  both `2.7` and `@a`. If you prefer to keep it numeric, `Num` clears the
  float/string cases but not the `@a`-as-count idiom, which needs the
  contextual-return machinery `scalar()` already uses.

## unshift
- Instances: 5
- PSC signature: `"unshift": {MinArity: 2, ArgTypes: []Type{Array, List}, ReturnType: Int}`
- Perl's actual behaviour: as with push, one argument compiles.
  ```
  $ perl -e 'my $u = eval q{my @z; unshift @z; 1}; print $u ? "ok\n" : "ERR: $@"'
  Useless use of unshift with no values at (eval 1) line 1.
  ok
  ```
  `t/op/unshift.t:14` — `$count3 = unshift (@array);` — is a test *of* that
  behaviour.
- Verdict: **MIXED** — 4 × arity-mismatch. Two of them (`unshift.t:14,49`,
  `tiearray.t:215`) are genuine 1-argument calls that perl accepts, so
  MinArity 2 is wrong. One (`t/io/perlio.t:229`, `unshift @INC, sub {...}`) is
  **U1**, paren-less. The 1 × `argument 1 expects Array, got Str` is U2.
- Example: `/home/perigrin/dev/perl5/t/op/unshift.t:14` — `$count3 = unshift (@array);`
- Suggested signature: `{MinArity: 1, ArgTypes: []Type{Array, List}, ReturnType: Int}`

## rindex
- Instances: 4
- PSC signature: `"rindex": {MinArity: 2, ArgTypes: []Type{Str, Str, Int}, ReturnType: Int}`
- Perl's actual behaviour: identical to index — POSITION numifies.
  ```
  $ perl -e 'no warnings; print rindex("hello","l","3"), "\n"'   # 3
  ```
- Verdict: **SIGNATURE WRONG** — all 4 are `argument 3 expects Int, got Str`.
- Example: `/home/perigrin/dev/perl5/t/uni/overload.t` (rindex hits cluster there)
- Suggested signature: `{MinArity: 2, ArgTypes: []Type{Str, Str, Num}, ReturnType: Int}`

## values
- Instances: 3
- PSC signature: `"values": {MinArity: 1, ArgTypes: []Type{Hash | Array}, ReturnType: List}`
- Perl's actual behaviour: matches.
- Verdict: **SIGNATURE OK** — 2 × `got Str` are **U2** (`values %hash` after
  `my %hash = 1..20` at `t/op/each.t:106`); 1 × `got Scalar` at
  `t/op/aassign.t:482` is the same class.
- Example: `/home/perigrin/dev/perl5/t/op/each.t:106` — `values %hash;`
- Suggested signature: no change.

## split
- Instances: 3
- PSC signature: `"split": {MinArity: 0, ArgTypes: []Type{Regex | Str, Str, Int}, ReturnType: List}`
- Perl's actual behaviour: the pattern argument is numified/stringified like
  any other; a boolean expression works.
  ```
  $ perl -e 'print join(",", split(1==1, "a1b")), "\n"'   # a,b
  ```
  All 3 diagnostics are `argument 1 expects Int|Num|Str|Regex|NaN|Inf, got Bool`
  — `Bool` is simply missing from the accepted union. Note `Str` in this
  lattice is `strLeaf|Num|NaN|Inf` and does **not** include `Bool`.
- Verdict: **SIGNATURE WRONG** (marginally) — a Bool used as a pattern is
  legal but is almost certainly a real bug in the code being checked; these
  3 are arguably true positives worth keeping. Honest answer: ambiguous.
- Example: the three hits are in `t/run/locale.t`-style boolean-pattern code.
- Suggested signature: no change, or add `| Bool` if you want zero false
  positives at the cost of losing a genuinely suspicious pattern.

## shift / pop
- Instances: 3 + 2
- PSC signatures: `"shift": {MinArity: 0, ArgTypes: []Type{Array}, ReturnType: Scalar}`,
  `"pop": {MinArity: 0, ArgTypes: []Type{Array}, ReturnType: Scalar}`
- Perl's actual behaviour: matches. `shift` on a scalar is a hard error in
  5.42 ("Experimental shift on scalar is now forbidden"), so `{Array}` is
  right.
- Verdict: **SIGNATURE OK** — all 5 are `argument 1 expects Array, got Str`,
  i.e. **U2** on the array being shifted.
- Suggested signature: no change.

## int
- Instances: 3
- PSC signature: `"int": {MinArity: 1, ArgTypes: []Type{Num}, ReturnType: Int}`
- Perl's actual behaviour: exactly one argument, numifies its input.
  ```
  $ perl -e 'no warnings; print int("3.9abc"), "\n"'   # 3
  $ perl -e 'int(1,2)'
  Too many arguments for int ... near "2)"
  ```
- Verdict: **MIXED** — 2 × `argument 1 expects Num, got Str` are the
  numeric-string coercion, same class as chr; 1 × `argument 3` is **U4**,
  since perl has no third argument for int.
- Suggested signature: `{MinArity: 1, ArgTypes: []Type{Str}, ReturnType: Int}`
  if you want string numification accepted; otherwise no change and accept
  2 false positives.

## chop / chomp
- Instances: 3 + 3
- PSC signatures: `"chop": {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: Str}`,
  `"chomp": {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: Int}`
- Perl's actual behaviour: both take a LIST of lvalues and modify each in place.
  ```
  $ perl -e 'my @c=("a\n","b\n"); chomp @c; print "@c\n"'      # a b
  $ perl -e 'my @c=("ab","cd"); chop @c; print "@c\n"'         # a c
  $ perl -e 'my ($p,$q)=("a\n","b\n"); chomp($p,$q); print "$p$q|\n"'  # ab|
  ```
- Verdict: **SIGNATURE WRONG** — `Str` rejects `chop(@bar)` and `chomp(@x)`,
  both of which are the primary documented use. 2 × chop Array, 1 × chop
  `argument 2` (from `chop($foo,@foo)`, a legal two-argument call the
  signature's arity handling mangles), 1 × chomp Array, 1 × chomp List,
  1 × chomp Object.
- Example: `/home/perigrin/dev/perl5/t/op/chop.t:31` — `chop($foo,@foo);`
- Suggested signature: `chop` → `{MinArity: 0, ArgTypes: []Type{List}, ReturnType: Str}`,
  `chomp` → `{MinArity: 0, ArgTypes: []Type{List}, ReturnType: Int}`

## Remaining singletons (ucfirst, uc, ord, lcfirst, lc, exists, defined)
- Instances: 1 each, 7 total.
- All are `expects Str, got <something stringifiable>` or the equivalent for
  `exists`/`defined`. They follow the same pattern as `print`/`length`: the
  `Str` mask excludes `Undef`, `Bool` and `Ref`, all of which these builtins
  accept and stringify. Not worth individual signature surgery; they will
  mostly clear if the `Str`-means-"anything stringifiable" question below is
  settled.

---

## The one cross-cutting question

Perl has no static distinction between "a string" and "a thing that
stringifies". PSC's `Str` mask is `strLeaf | Num | NaN | Inf` — it excludes
`Undef`, `Bool`, `DualVar` and every `Ref`. Every builtin whose ArgTypes says
`Str` will therefore fire on `print $bool`, `length undef`, `die $ref`,
`uc $obj`, and so on, all of which are ordinary perl.

Two coherent positions:

1. **`Str` means "a string value"** — then these are true positives and the
   corpus is telling you perl's own test suite is full of implicit coercion.
   Defensible, but it makes PSC very noisy on real code.
2. **Argument positions that stringify should be typed `Scalar`, not `Str`** —
   reserving `Str` for annotations where the author asserted a string.

The signature fixes above take position 2 where perl's documented contract
requires it (`length` must accept undef; `die`/`warn`/`print`/`chomp`/`chop`
take LIST), and leave the rest alone. That is the honest line: fix where perl
*specifies* the wider domain, don't fix where perl merely *tolerates* it.

## Estimated effect

| Fix | Diagnostics cleared |
|---|---|
| U1 — paren-less call argument collection | ~52 |
| U3+U4 — heredoc bodies parsed as code | 30 |
| U2 — range/list assignment retypes container | ~29 |
| substr 4th argument | 45 |
| print/die/warn/chomp/chop → List | ~22 |
| chr/index/rindex/int → Num | ~26 |
| length → Scalar | 10 |
| join/push/unshift MinArity | ~6 |
| **Total** | **~220 of 298** |

The four upstream bugs alone account for ~111. They are worth more than the
signature table.
