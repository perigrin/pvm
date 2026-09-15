# PSC triage against perl's t/ corpus

Date: 2026-09-04
Status: findings only — nothing fixed, nothing committed to `internal/`

## What was run

`psc check` over all 620 files in `~/dev/perl5/t`: **1183 diagnostics**, 7
minutes single-threaded. Four subagents triaged one class each, every
false-positive claim verified by running perl rather than by reasoning.

| bucket | diagnostics | verdict |
|---|---|---|
| numeric positions | 762 | ~580 false positives, ~0 real bugs in perl's tests |
| builtin calls | 298 | ~220 clearable; signature table is NOT the main cause |
| string/regex | 123 (94 real) | ~85 false positives from 3 mechanisms |
| arity | 53 | 49 false, 4 TRUE |

## The number that matters more than any of these

**89 of 227 `t/op` files have parse errors — 39%, 756 error nodes.**

Every diagnostic above is computed on whatever survived parsing. The
triage measures the checker; this measures whether the checker is being
shown the program. Coverage and mutation detection are both bounded by it,
so it belongs at the top of the list rather than in it.

## Findings, highest leverage first

### 1. Upstream parser bugs, not signature bugs (~111 + all arity)

The builtins agent's headline: no `signatures.go` edit can silence these.

- **Paren-less calls collapse to one argument.** `join ":", @a` is seen as
  `join(":")`; `join(":", @a)` on the next line is clean. Accounts for
  *every* arity mismatch in the corpus, and the same shape hits `index`,
  `unshift`, `substr`. Partially addressed already this session; the
  remaining shapes are below.
- **Heredoc bodies are parsed as live code.** All 30 `bless` diagnostics
  are inside one `<<~'EOF_DUMP'` in `t/porting/header_parser.t`. A 496-line
  body yields 495 diagnostics. Compounded by a second bug: in a comma
  chain the previous `bless`'s Object return leaks in as the next call's
  argument 2.
- **`format NAME = ... .` blocks unrecognised.** The terminating lone `.`
  is read as concatenation and swallows the next statement, so diagnostics
  point at lines containing no `.` at all.
- **`/x` lexed as the binary `x` operator.** Silent for `/i`, `/g`, `/m`,
  `/s` only because those letters are not operators.
- **`my %h = 1..20` retypes the container to the RHS type.** `my %a = (1,2)`
  is clean; `my %a = 1..20` makes `keys %a` report "got Str".

### 2. Lattice and signature fixes (~150 from two, cheap)

- **`Bool` is not a subtype of `Int`** (87). `"not " x ($a ne $b)` is *the*
  dominant idiom in perl's own tests. Verified: `"not " x ("x" ne "y")`
  works and the boolean is 1.
- **Bitwise string operators** (81). `"ok \xFF\xFF\n" & "ok 19\n"` returns
  a STRING per perlop. Verified independently.
- **`=~` declares `{Left: Str, Right: Regex}`** (33). `Str` is not a
  subtype of `Regex` and `coercion.go` has no `Regex` target, so every
  string pattern mismatches. Perl compiles strings into patterns freely.
- **Magic string ranges** (239). `'a'..'z'`, `"-3".."0"`. PSC types `..` as
  `(Int,Int)`; perl picks numeric-vs-string increment at runtime. Verified:
  `("a".."e")` gives `a b c d e`.
- **`substr` takes a 4th argument** (45–46). The replacement string.
  Verified: `substr($s,0,2,"XY")` returns "he" and leaves `$s` as "XYllo".
  The signature has three, so the variadic tail repeats `Num`.
- **`print`/`die`/`warn`/`chomp`/`chop` typed `{Str}`, all take a LIST** (~22).
- **`length` must accept `Scalar`** — `length(undef)` returning undef is the
  documented contract, and `t/op/length.t:127` is the test asserting it.
- **`join`/`push`/`unshift` `MinArity: 2` is stricter than perl.**
  `t/op/unshift.t:14` is a test OF the one-argument form.

### 3. Arity shapes still open (26 of 49, two small edits)

- `\substr $s, 0, 1` (13) — `collectTrailingListArgs` walks up through
  `assignment_expression` and `binary_expression` but not
  `refgen_expression`, so it hits `default: return nil` before reaching the
  enclosing list. Adding `refgen_expression` and `unary_expression` retires
  all 13.
- Doubly-nested swallow (9 + 4) — `unwrapSwallowedArgs` unwraps exactly one
  level. Making it recurse fixes both.
- Nested call as a non-first list element (16) — the known gap;
  `push @w, join '', @_`. Needs a real rule about which call consumes a
  trailing list, not a heuristic.

### 4. A flow-analysis bug worth its own entry

**Void stringification clobbers the variable's type.** `t/base/num.t` does
`$a = 1; "$a";` and PSC treats the rvalue interpolation as a store, so
every later `$a + 1` sees `Str`. The file's own comment calls this a
deliberately planted trap. This will misfire well beyond this corpus.

## True positives — PSC was right

- 4 genuine 1-arg `push`/`unshift` in `tiearray.t` and `unshift.t`. Perl
  itself says "Useless use of push with no values", and three sit inside
  `no warnings 'syntax'` because they test the degenerate case.
- The undef group in string positions. Verified with an EXPLICIT
  `my $u = undef` rather than an unassigned scalar: perl emits "Use of
  uninitialized value" for concat, range, match and `eq`.
- `split` with a `Bool` pattern (3) — legal but probably a real bug in the
  checked code.

## Policy questions, not bugs

- Reference and overloaded-object stringification (35) are
  coercible-but-not-subtype BY DESIGN. `CoercionMismatch` fires on exactly
  `IsCoercible && !TypeSatisfies`, so it cannot distinguish accidental
  stringification from intentional. Several flagged tests exist
  specifically to assert that stringification works.
- PSC ignores lexical warning state. `range.t` and `repeat.t` contain
  neither `use warnings` nor `no warnings`, so perl is silent where PSC
  escalates to `error` — and `range.t:136` reads
  `# undef should be treated as 0 for numerical range`. PSC is flagging the
  subject of the test.

## Unresolved

`libperl.t:188` reports its type as the entire lattice, and `$nm_style`'s
union contains a stray `0x80000000` — the `None` sentinel leaking into type
display. Marked unclear rather than guessed at.

## Method note

Every false-positive claim in the four reports has a perl run behind it,
with witnesses kept. The agents were briefed on the trap this session hit
three times: an unassigned `my $x` IS undef, so a witness testing "what
happens when this is defined" must assign explicitly. One agent
specifically noted keeping a separate genuinely-unassigned variable so the
two cases could not be confused.

Counts are attributed by message shape from 3–5 sampled instances per
group, not verified line-by-line.

Full reports: `scratchpad/triage/report-{numeric,builtins,string,arity}.md`
