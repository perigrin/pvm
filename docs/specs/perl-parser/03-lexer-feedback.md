<!-- ABOUTME: Specifies the lexer feedback problem in Perl — how the lexer's decisions depend on parser state and on symbol-table facts that only exist at runtime. -->
<!-- ABOUTME: Defines the PL_expect state machine, prototype/BEGIN/source-filter undecidability, and the approximation strategy a static Go parser must adopt, with an empirical verification protocol. -->

# Chapter 3: The Lexer Feedback Problem

## 3.0 The claim this chapter has to defend

Perl's grammar is not context-free, and its lexer is not independent of its
parser. Both halves of that sentence are load-bearing, and they fail in
different ways:

1. **Lexer ← parser.** The lexer cannot tokenize without knowing what the
   parser expects next. `/` is division or the start of a regex depending on
   a single parser-maintained variable. This is *feedback*, and it is
   mechanical: a Go parser can reproduce it exactly, because the information
   flows entirely within the parse.

2. **Lexer ← runtime.** The lexer consults the *live symbol table* — a data
   structure built by executing the program being parsed. `f(@a)` parses
   differently depending on whether a sub named `f` has already been compiled
   and what prototype it carries. This is not feedback; it is a dependency on
   an interpreter that has already run. A static parser cannot reproduce it
   in general, only approximate it.

Category 1 is engineering. Category 2 is a soundness boundary, and this
chapter's job is to draw it precisely, say what falls on each side, and give
the implementer a way to *measure* every approximation against real perl
rather than trusting this document.

The good news is quantitative, and it is measured, not asserted (§3.9): on
perl5's own 620-file test suite, the constructs in category 2 that actually
change a parse appear in roughly 1–8% of files, and the two truly hopeless
ones (source filters) appear in 2 files. The bad news is that the tail is
real and a spec that pretends otherwise is worse than no spec.

---

## 3.1 The `PL_expect` state machine

### 3.1.1 What it is

`PL_expect` is a single enum, declared in `perl.h:5972-5985`, that carries
the parser's expectation into the lexer. It is the entirety of the official
feedback channel — everything else the lexer consults is either lexical
position (`PL_lex_brackets`, `PL_lex_state`) or the symbol table.

```c
typedef enum {
    XOPERATOR,
    XTERM,
    XREF,
    XSTATE,
    XBLOCK,
    XATTRBLOCK, /* next token should be an attribute or block */
    XATTRTERM,  /* next token should be an attribute, or block in a term */
    XTERMBLOCK,
    XBLOCKTERM,
    XPOSTDEREF,
    XTERMORDORDOR /* evil hack */
    /* update exp_name[] in toke.c if adding to this enum */
} expectation;
```

The debug names live at `toke.c:5465-5469`, and reveal a historical wart: the
name array lists `"SIGVAR"` between `POSTDEREF` and `TERMORDORDOR`, which no
longer corresponds to an enum member. Do not use the debug array as the
authoritative list; use `perl.h`. The practical consequence: the array has
twelve entries and the enum eleven, so `exp_name[XTERMORDORDOR]` (index 10)
prints as `"SIGVAR"`. A `-DT` trace that says `SIGVAR` means
`XTERMORDORDOR`.

### 3.1.2 The eleven states

| State | Meaning | Set by (examples) | The question it answers |
|---|---|---|---|
| `XOPERATOR` | A complete term was just consumed; an infix/postfix operator must follow. | `TERM()` macro (`toke.c:254`); `FUN0`/`FUN1` (`262`, `264`); the end of a variable (`yyl_dollar` `5793`, `yyl_snail` `7052`, `yyl_percent` `6419`, `yyl_star` `6360`), of a string (`S_sublex_done`, `2762`, `2821`), of `qw` (`6144`), of a readline/heredoc (`7180`); `}` whose stack entry was `XOPERATOR` (§3.1.4). One exception: after `$fh` directly following a list operator, `yyl_dollar` re-sets `XTERM` when what follows looks like a term (`5797-5845`: `print $fh -1`) | `/` is divide, `{` is a subscript, `<` is less-than |
| `XTERM` | A value is required here. | `OPERATOR()` macro (`toke.c:249`) and the operator macros (`265-278`); `UNI()` named unary operators (`296`); `LOP(f, XTERM)` list operators (`2152`, `2188`); `(` unless it directly follows a list or named-unary operator (`7146-7149`); an anon-hash `{` (`OPERATOR(HASHBRACK)`, `6655`); `use`/`no` (`5443`); `sort` (`9041` — then `LOP(OP_SORT, XREF)` at `9043` overwrites it unless `force_word` queued a comparator name, §3.1.3 `LOP`); `}` whose stack entry was `XTERM` | `/` starts a regex, `{` is an anon hash, `<` starts a readline |
| `XREF` | A dereference block/term follows a sigil, or an indirect-object slot. | `PREREF()` (`toke.c:253`) from a bare sigil — `$` (`5699`), `$#` (`5687`), `@` (`7034`), `%` (`6410`), `&` (`6951`), `*` (`6363`) — i.e. `${`, `@{`, `&{`; `LOP(f, XREF)` for `print`/`printf`/`say` (`8837`, `8841`, `8952`), `map`/`grep` (`8756`, `8585`) and `sort` (`9043`); `S_intuit_method` on `foo $bar` (`5106`); `S_scan_ident` at the end of an interpolated identifier (`11231`) | `{` after a bare sigil is a deref block, never an anon hash — `yyl_leftcurly` skips the §3.3 heuristic for it (`6719-6721`); after `print`/`map`/`grep` the `{` *does* go through the heuristic, because it directly follows a list operator |
| `XSTATE` | Start of a statement. | After `;` (`toke.c:9669`); after every `{` that opens a block — all non-`HASHBRACK` arms of `yyl_leftcurly` (`6685`, `6691`, `6696`, `6837`, `6841`); after a `}` whose stack entry was `XSTATE` (`6862`); after a keyword-plugin statement (`9352`); around `format` lines (`9826`, `13337`) | A bareword followed by `:` is a **label**; `{` goes through the §3.3 heuristic (`{ a => 1 }` at statement start is a hash) |
| `XBLOCK` | A `{` here opens a block, unconditionally. | `PREBLOCK()` (`toke.c:251`): `else` (`8475`), `continue` (`8393`), `try`/`catch`/`finally` (`9127`, `8371`, `8550`), `defer` (`8443`), `default` (`8439`), `package NAME` (`8862`), `format` (`5919`), a `)` directly followed by `{` — the end of an `if`/`while`/`for` condition (`7164`), a sub with an `&`-first prototype (`6636`); `PHASERBLOCK` (`255`) for `BEGIN`/`END` &c.; the `)` closing a signature (`5632`); `XATTRBLOCK` once `:` has been seen (`6471`) | `{` is a block, never a hashref |
| `XATTRBLOCK` | An attribute list (`:lvalue`) or a block. | Unconditionally after the name in `sub NAME` / `method NAME` / `format NAME` (`yyl_sub`, `toke.c:5879`) and after `class NAME` (`8382`); the state says attributes *may* follow. `yyl_colon` turns it into `XBLOCK` on `:` (`6471`) | `:` is an attribute marker, not the ternary colon |
| `XATTRTERM` | An attribute list or a block, in term position (anon sub). | After `sub`/`method` with no name following (`yyl_sub`, `toke.c:5907`). `yyl_colon` turns it into `XTERMBLOCK` on `:` (`6474`) | same, in an expression |
| `XTERMBLOCK` | A term whose `{` opens a **block** but which is otherwise a term. | `PRETERMBLOCK()` (`toke.c:252`) for `do {` (`7514`); `eval {` (`8495`); a regex code block `(?{ ... })`, which the lexer hands to the parser as a synthetic `do {` (`10114`); an anon sub's attribute list (`XATTRTERM` → `XTERMBLOCK`, `6474`); a `format` argument line starting with `{` (`13343`) | `eval {`, `do {` — the `{` is a block, never a hash |
| `XBLOCKTERM` | A `{` opens a block, but a term is the overall result. | `yyl_just_a_word`, when a bareword with no known sub is followed by `$` or `{` and the `indirect` feature is on — the `METHCALL0` path (`toke.c:8215-8224`) | `foo {...}` / `foo $x` as an indirect method call (`{...}->foo`) |
| `XPOSTDEREF` | Immediately after `->` when the lexer has already seen that a postfix-deref form follows. | `yyl_hyphen`, when `->` is followed by `$*`, `$#*`, `@*`, `@[`, `@{`, `%*`, `%[`, `%{`, `&*`, `**` or `*{` (`toke.c:6285-6293`) | `$` `@` `%` `&` `*` are postfix-deref, not operators |
| `XTERMORDORDOR` | "Evil hack" (perl's own word). Term expected, but `//` must still lex as defined-or. | `FTST()` macro (`toke.c:261`) — filetest operators; `UNIDOR()` (`297`) — named unary operators with an optional argument: `getc`, `pop`, `pos`, `readline`, backtick-`readpipe`, `readlink`, `shift`, `undef`, `umask` (`8594`, `8850`, `8853`, `8923`, `8926`, `8938`, `9009`, `9154`, `9163`) | `-e $f // die`, `shift // 0` — is `//` an empty regex or defined-or? |

`XTERMORDORDOR` exists because filetest operators (`-e`, `-f`) and the
`UNIDOR` operators (`shift`, `pop`, …) are named unary operators that
legitimately take no argument, so after `-e` or `shift` both a term and an
operator are grammatical. Only four places in toke.c treat it differently
from `XTERM`: `yyl_slash` reads `//` as defined-or (`toke.c:7060`; a single
`/` is still a regex), `yyl_tilde` reads `~~` as smartmatch (`7129`),
`yyl_leftcurly` treats `{` as an anon hash exactly as in `XTERM` (`6652`),
and the v-string check at `9905-9907`. Everything else treats it as a term
position. Implement it as a distinct state, not as `XTERM`; collapsing it
silently mis-lexes `-e $f // die` and `shift // 0` (both verified with
`Deparse` to be defined-or).

### 3.1.3 Transitions

The transitions are not written as a table in toke.c — they are encoded in
the *return macros* at `toke.c:249-303`. This is the single most useful
structural fact for a Go implementation: **almost every transition is a
property of the token being emitted, not of a hand-written state graph.**
`grep -c 'PL_expect *=[^=]' toke.c` finds 98 assignments, 24 of them the
macro definitions themselves; every one of the eleven enum members is a
target somewhere, and most of the 74 remaining sites flow through these
macros:

| Macro | Definition site | Sets `PL_expect` to | Emitted for |
|---|---|---|---|
| `TERM(t)` | `toke.c:254` | `XOPERATOR` | a completed value: variable, literal, `)`, `]` |
| `OPERATOR(t)` | `toke.c:249` | `XTERM` | an infix or prefix operator |
| `PREBLOCK(t)` | `toke.c:251` | `XBLOCK` | tokens that must be followed by a block |
| `PREREF(t)` | `toke.c:253` | `XREF` | a bare sigil (`$` `@` `%` `&` `*` with no name) |
| `PRETERMBLOCK(t)` | `toke.c:252` | `XTERMBLOCK` | `do {` (`toke.c:7514`) |
| `Aop/Mop/BAop/BOop/SHop/PWop/PMop/ChEop/...` | `toke.c:265-278` | `XTERM` | arithmetic/bitwise/shift/match/comparison operators |
| `FUN0(f)`, `FUN0OP`, `FUN1(f)` | `toke.c:262-264` | `XOPERATOR` | zero-argument functions (`time`, `wantarray`); one-argument functions written `f(`  |
| `UNI(f)` | `toke.c:296` (via `UNI3`, `285`) | `XTERM` | named unary operators |
| `UNIDOR(f)` | `toke.c:297` | `XTERMORDORDOR` | named unary operators with an optional argument (`shift`, `pop`, …) |
| `UNIBRACK(f)` | `toke.c:303` | **unchanged** (`have_x = 0`) | `eval` — which is why `KEY_eval` sets `XTERMBLOCK`/`XTERM` by hand first (`8495-8505`) |
| `UNIPROTO(f,opt)` | `toke.c:298` | `XTERM` (expands to `OPERATOR`) | a user sub whose prototype makes it a named unary (§3.5.3) |
| `FTST(f)` | `toke.c:261` | `XTERMORDORDOR` | filetest operators |
| `LOOPX(f)` | `toke.c:257` | `XOPERATOR` if a label token was forced, else `XTERM` | `next`, `last`, `redo`, `goto`, `dump` |
| `LOP(f,x)` | `toke.c:2152` (`S_lop`, `2174`) | caller-supplied `x` — but **only if no token is already queued** (`PL_nexttoke`, `2186-2188`) | list operators |
| `OLDLOP(f)` | `toke.c:382` | `XTERM` | `return` (`8893`) |
| `POSTDEREF(f)` | `toke.c:256` | `XOPERATOR` (in `S_postderef`, `toke.c:2234`; assignments at `2244`, `2258`) | postfix deref |

**Design consequence for Go.** Model this as a function
`nextExpect(tok TokenKind, cur Expect) Expect` driven by a table keyed on
token kind, with a small number of documented exceptions (`LOOPX`, `LOP`,
`yyl_leftcurly`, `yyl_rightcurly`). Do not write an 80-case switch mirroring
toke.c's assignment sites; you will get the same behaviour with a fraction of
the code, and the exceptions are what you should spend review effort on.

### 3.1.4 The bracket stack: how `}` restores expectation

`PL_expect` alone is insufficient, because `}` must restore whatever
expectation held before the matching `{`. Perl keeps a parallel stack,
`PL_lex_brackstack`, pushed in `yyl_leftcurly` (`toke.c:6643`) and popped in
`yyl_rightcurly` (`toke.c:6853`, the pop itself at `6862`):

```c
PL_expect = (expectation)PL_lex_brackstack[--PL_lex_brackets];
```

The value pushed is *the expectation that will hold after the closing brace*,
and it differs by opening context (`toke.c:6650-6703`):

| `PL_expect` at `{` | Pushed for the matching `}` | Token emitted for `{` |
|---|---|---|
| `XTERM`, `XTERMORDORDOR` | `XOPERATOR` | `HASHBRACK` (anon hash) |
| `XOPERATOR` | `XOPERATOR` | `PERLY_BRACE_OPEN` (subscript) |
| `XATTRTERM`, `XTERMBLOCK` | `XOPERATOR` | `PERLY_BRACE_OPEN` |
| `XATTRBLOCK`, `XBLOCK` | `XSTATE` | `PERLY_BRACE_OPEN` (block) |
| `XBLOCKTERM` | `XTERM` | `PERLY_BRACE_OPEN` |
| default (`XSTATE`, `XREF`, `XPOSTDEREF`) | `XTERM` if the `{` directly follows a list op (`PL_oldoldbufptr == PL_last_lop`), else `XOPERATOR` (`6700-6703`); **rewritten to `XSTATE`** at `6840` when the §3.3 heuristic decides "block" in a non-`XREF` state | disambiguated (§3.3) |

The expectation *inside* the braces is set at the same time: `XTERM` for an
anon hash (`OPERATOR(HASHBRACK)`, `6655`), `XSTATE` for every block arm
(`6685`, `6691`, `6696`, `6837`, `6841`), and `XTERM` for the two `XREF`
term outcomes of §3.3 step 8 (`6824`, `6833`).

A Go parser needs this stack. A single scalar expectation cannot survive
nesting.

---

## 3.2 The disambiguation table

Every entry below is decided by `PL_expect` *plus*, in the starred rows,
information the lexer cannot get from the token stream.

| Char | `XOPERATOR` | `XTERM` / term-ish | Other states | toke.c |
|---|---|---|---|---|
| `/` | divide; `//` is defined-or | start of match regex | `XTERMORDORDOR`: `//` is defined-or, else regex | `yyl_slash`, `7058` |
| `?` | ternary `?:` | (historically match delim; now a syntax error) | — | `9810` |
| `{` | subscript | anon hash (`HASHBRACK`) | `XBLOCK`/`XATTRBLOCK`/`XTERMBLOCK`/`XATTRTERM`/`XBLOCKTERM`: block; `XSTATE`/`XREF`/`XPOSTDEREF`: the §3.3 heuristic, except that a bare sigil's `{` (`${`, `@{`) is always a block | `yyl_leftcurly`, `6643` |
| `}` | — | — | pops `PL_lex_brackstack` | `yyl_rightcurly`, `6853` |
| `<` | less-than; `<<` shift; `<=>` | `<<HEREDOC` if `s[1]=='<' && s[2]!='>'`, else `<FH>` readline | — | `yyl_leftpointy`, `7169` |
| `>` | greater-than; `>>` shift | (only reached as operator) | — | `yyl_rightpointy`, `7221` |
| `*` | multiply; `**` power | glob `*name` (calls `scan_ident`) | `XPOSTDEREF`: postfix deref `->@*` | `yyl_star`, `6353` |
| `%` | modulo | hash `%name`\* | `XPOSTDEREF`: postfix deref | `yyl_percent`, `6391` |
| `&` | bitwise and; `&&` | sub call `&name` | `XPOSTDEREF`: postfix deref | `yyl_ampersand`, `6901` |
| `+` | add; `++` is postincrement | unary plus; `++` is preincrement | — | `yyl_plus`, `6325` |
| `-` | subtract; `--` postdec | unary minus; **filetest** `-e` if next is a single letter + non-word (`6202`) and not followed by `=>` (`6213`)\* | — | `yyl_hyphen`, `6200` |
| `~` | `~~` smartmatch, only while the `smartmatch` feature is enabled (`7128`; off under `use v5.42`) | complement (`~` or `~.` under `bitwise`) | `XTERMORDORDOR`: `~~` as in `XOPERATOR` | `yyl_tilde`, `7125` |
| `(` | function-call parens | grouping / list | after list op: keeps `oldbufptr` for `print(STDOUT 1)` | `yyl_leftparen`, `7144` |
| `:` | ternary colon | — | `XATTRBLOCK`/`XATTRTERM`: attribute; `XSTATE`+bareword: label | `yyl_colon`, `6456` |
| bareword | error ("no operator expected") | sub / string / filehandle / class / label\* | — | `yyl_just_a_word`, `8023` |

Starred rows need symbol-table knowledge and are covered in §3.3–§3.6.

### 3.2.1 Worked examples

```perl
# / — pure PL_expect, fully static
my $x = 10 / 2;         # XOPERATOR after $x's value... -> divide
print "hit" if /foo/;   # 'if' -> OPERATOR() -> XTERM     -> regex
my @m = grep { /x/ } @l;# '{' opens block -> XSTATE       -> regex
$a =~ /x/;              # =~ is an operator -> XTERM      -> regex
$n = $c // 3;           # XOPERATOR + '//'                -> defined-or
-e $f // die;           # FTST -> XTERMORDORDOR + '//'    -> defined-or

# { — block vs hashref
my $h = { a => 1 };     # after '=' -> XTERM              -> anon hash
my $h = { $x => 1 };    # after '=' -> XTERM              -> anon hash (no heuristic)
{ a => 1 }              # stmt start -> XSTATE, §3.3      -> anon hash
{ $x => 1 }             # stmt start -> XSTATE, §3.3      -> block
sub f { 1 }             # sub NAME -> XATTRBLOCK          -> block
map { $_ } @l;          # map -> LOP(XREF), §3.3 heuristic -> block
map {; $_ } @l;         # leading ';' disambiguates to block by force
map { +{ a=>1 } } @l;   # unary + forces the inner one to a hashref
map { "$_" => 1 } @l;   # string then => -> HASHBRACK      -> syntax error at @l
print {$fh} "x";        # print -> LOP(XREF), heuristic   -> block
@{ "foo", "bar" }       # bare sigil -> XREF, no heuristic -> block

# < — readline vs less-than
while (my $l = <STDIN>) {}   # after '=' -> XTERM         -> readline
if ($a < $b) {}              # after $a  -> XOPERATOR     -> less-than
my $t = <<"END";             # XTERM, s[1]=='<'           -> heredoc
print $fh $a <=> $b;         # XOPERATOR                  -> spaceship
```

The `map { +{ a=>1 } }` idiom exists precisely because perl's own
disambiguation (§3.3) is a heuristic that programmers learned to override by
hand. That is a strong signal: **where perl needs a manual override, your
static parser will need one too, and the same `+` works.**

---

## 3.3 `{` — the heuristic perl uses when expectation is not enough

When `PL_expect` falls through to `default` in `yyl_leftcurly`
(`toke.c:6698-6842`) — that is, in `XSTATE`, `XREF` or `XPOSTDEREF`: at
statement start, after `print`/`map`/`grep`, after a bare sigil, inside
`eval ""` — perl runs a lookahead heuristic. It is worth reproducing because it
is one of the few places perl documents its own guesswork in comments:

> This hack serves to disambiguate a pair of curlies as being a block or an
> anon hash. Normally, expectation determines that, but in cases where we're
> not in a position to expect anything in particular (like inside `eval""`)
> we have to resolve the ambiguity. This code covers the case where the first
> term in the curlies is a quoted string. Most other cases need to be
> explicitly disambiguated by prepending a "+" before the opening curly.
> — `toke.c:6723-6737`

The algorithm:

1. Push the stack entry (§3.1.4) and skip whitespace after `{`.
2. If the next char is `}` → anon hash (`HASHBRACK`), `toke.c:6706-6715`
   (inside an interpolated string, `${}` is instead a syntax error, `6707`).
3. If `PL_expect == XREF` **and** the `{` does not directly follow a list
   operator — i.e. `${`, `@{`, `%{`, `&{`, `*{` — skip steps 4–7 entirely
   and go to step 8 (`toke.c:6719-6721`). `@{ "foo", "bar" }` is a block
   even though a string followed by `,` would otherwise say "hash".
4. If the next char is `'`, `"` or `` ` `` → scan past the string, honouring
   backslash escapes (`toke.c:6740-6746`).
5. Else if it starts `q`, `qq`, `qx` (followed by a non-word char) → if the
   next non-space is `=>` it is an anon hash at once (`6760-6762`, `q =>`);
   otherwise scan past the quote-like construct, tracking nested bracket
   delimiters (`toke.c:6747-6788`). A plain word starting with `q` (`qux`)
   is scanned as a bareword (`6789-6797`).
6. Else if the next char is a word character, scan the bareword
   (`toke.c:6799-6806`). Otherwise (`$`, `-`, `(`, `+`, …) the scan pointer
   does not move, so the next step looks at that character itself — which
   is why `{ $x => 1 }` is a block.
7. Skip whitespace. If the next char is `,` (and the first char was `q` or
   not lowercase) **or** `=>`, it is an anon hash (`toke.c:6810-6814`).
8. Otherwise, in `XREF` (`toke.c:6815-6838`): if followed by another `{` or
   by `sub` with an attribute colon, it is a term (`PL_expect = XTERM`,
   `6824`, `6833`), so `${{...}}{k}` and `&{sub:attr...}` work; else a
   statement (`PL_expect = XSTATE`, `6837`), so `map {no strict; ...}`
   works.
9. Otherwise: a block — `PL_expect = XSTATE` and the stack entry pushed in
   step 1 is overwritten with `XSTATE` (`toke.c:6839-6841`).

Note the asymmetry at step 7: `{ foo, ... }` with a *lowercase* bareword is
**not** treated as a hash, because `foo` is more likely a function call. This
is a deliberate lean, not an oversight. Verified against perl 5.42.0 with
`Deparse`, all at statement start: `{ a => 1 }`, `{ Foo, 1 }`, `{ "foo", 1 }`,
`{ q(foo), 1 }` are anon hashes; `{ foo, 1 }`, `{ $x => 1 }`, `{ lc($_) => 1 }`
are blocks.

**Static verdict.** Fully implementable — it reads only source text. It is
also *wrong* sometimes in real perl, and your Go parser being wrong in
exactly the same places is correct behaviour. Copy the algorithm, do not
improve it.

---

## 3.4 `$x[...]` vs `$x [...]`, and `%h` vs `%`

### 3.4.1 What `S_intuit_more` actually decides

`S_intuit_more` (docblock `toke.c:4553`, function `4576-5041`) answers:
"after this variable, is the following `[` or `{` a subscript, or something
else?" Perl calls it from `yyl_dollar` (`toke.c:5707`), `yyl_percent`
(`6413`), `yyl_snail` (`7039`), and from two interpolation sites (`10129`,
`11103`), and the consequences are structural — in `yyl_snail`, a `TRUE`
return with a following `{` **rewrites the sigil**:

```c
if (*s == '{')
    PL_tokenbuf[0] = '%';   /* @h{...} is a hash slice, not an array */
```

The decision procedure, in order (`toke.c:4598-4649`):

1. Already inside brackets (`PL_lex_brackets`) → TRUE.
2. Begins `->[` or `->{` → TRUE.
3. Begins `->$*`, `->$#*`, `->@*`, `->@[`, `->@{` and postderef_qq enabled → TRUE.
4. Not `{` or `[` → FALSE.
5. **Not inside a pattern → TRUE.** Outside a regex, `[` and `{` after a
   variable are always subscripts.
6. Inside a pattern, `{` matching `regcurly` (i.e. `{2,3}`) → FALSE (a quantifier), `4637`; any other `{` → TRUE.
7. Inside a pattern, `[`:
   - `[]` or `[^` → FALSE (character class), `4648-4649`.
   - **Symbol-table consultation** (`toke.c:4653-4695`): for `$foo[`, check
     whether a scalar `$foo` and/or an array `@foo` exist. Under `strict vars`
     this is decisive — if only `@foo` exists it is a subscript, if only
     `$foo` exists it is a character class, if neither exists it is an error.
   - Otherwise, fall into the weighting heuristic.

**Step 5 is the load-bearing simplification for a static parser.** Outside a
regex — which is the overwhelming majority of `$x[` occurrences in real code
— the answer is unconditionally "subscript," with no symbol table needed.

### 3.4.2 The weighting heuristic

Only reached for `[` *inside a pattern*. `toke.c:4696-5040`. Perl's own
comment (`4714`): *"this is terrifying, and it mostly works. See GH #16478."*
Weights (negative = subscript, non-negative = character class):

| Condition | Δweight | Line |
|---|---|---|
| no `]` before end of buffer | immediate TRUE (subscript) | 4698 |
| digits only, length ≤ 2 | immediate TRUE (subscript) | 4711 |
| first char is `$` | −1 (init) | 4728 |
| otherwise | +2 (init) | 4732 |
| repeat of `@`, `&`, `$` | −10 × times seen | 4762 |
| sigil + known multi-char global identifier | −100 | 4807 |
| sigil + unknown identifier | −10 | 4832 |
| `$` + punct-var char, then `]`/`}`/`)`/space/`=` | −10 | 4844 |
| `$` + punct-var char, otherwise | −1 | 4846 |
| `\w`, `\d`, `\s`, `\]` | +100 | 4863 |
| any other `\` once a `'` or `"` has been seen | +1 | 4869 |
| `\` + `abcfnrtvx` | +40 | 4896 |
| `\` + digits | +40 | 4899 |
| `\` at end of string | +100 | 4918 |
| `-\` | +50 | 4928 |
| `a-`, `A-`, `0-`, `1-`, `!-`, ` -` (not at first position) | +30 | 4934 |
| `-z`, `-Z`, `-7`, `-9`, `-~` | +30 | 4943 |
| leading `-` then digit or `$` | −5 | 4948 |
| non-word char then two alphas spelling a keyword | −150 | 4987 |
| consecutive code points (`ab`, `12`) | +5 | 5003 |
| repeated character | −(times seen) | 5016 |

Result: `weight >= 0` → character class (FALSE), `5039`; else subscript (TRUE).

The `−100` at line 4807 and the `strict vars` branch at 4653 are the only two
symbol-table dependencies. Both are reachable **only inside a regex**.

### 3.4.3 `%h` vs `%` modulo

Handled entirely by `PL_expect` in `yyl_percent` (`toke.c:6391-6423`):
`XOPERATOR` → modulo; `XPOSTDEREF` → postfix deref; anything else → sigil. Then `intuit_more` runs, and if
it returns TRUE with a following `[`, the sigil is rewritten `%` → `@`
(`%h[...]` is a *hash slice returning values*, ergo an array-ish construct).

```perl
my $r = $total % 3;   # XOPERATOR -> modulo
my %h = (a => 1);     # after 'my' -> XTERM -> hash sigil
@h{qw(a b)} = (1,2);  # @ + intuit_more sees '{' -> sigil rewritten to %
```

Fully static. No approximation needed.

### 3.4.4 `sort $coderef @list`

`sort` is parsed as a list operator via `S_lop` (`toke.c:2174-2196`), whose
documented rule is:

> - if we have a next token, then it's a list operator (no parens) for which
>   the next token has already been parsed; e.g., `sort foo @args`
> - if the next thing is an opening paren, then it's a function
> - else it's a list operator

The `sort SUBNAME LIST` / `sort $coderef LIST` form works because
`yyl_just_a_word` has a special case: `PL_last_lop_op == OP_SORT` forces the
indirect-object reading **regardless of whether the bareword is a known sub**
(`toke.c:8153-8156`):

```c
if (
    ( !immediate_paren && (PL_last_lop_op == OP_SORT
     || (!c.cv && ...)))
```

That `PL_last_lop_op == OP_SORT` short-circuit is a gift: **`sort` is the one
indirect-object site that needs no symbol table.** Verified:

```console
$ perl -MO=Deparse -e 'my $cr = sub {1}; my @l=(3,1); my @s = sort $cr @l;'
my(@s) = (sort $cr @l);      # $cr is the comparator, not the first element
```

Contrast with `sort $x, @l` — the comma makes `$x` an ordinary list element.
The distinguishing signal is purely syntactic (absence of a comma after the
first term), so a static parser gets this right.

---

## 3.5 Prototypes: the definition-order dependency

### 3.5.1 The problem, demonstrated

This is the cleanest possible demonstration that perl's parse depends on
execution order. The two programs differ only in statement order:

```console
$ perl -MO=Concise -e 'sub f(\@){} my @a; f(@a)'
9  <1> entersub vKS ->a
5     <0> pushmark s ->6
7     <1> srefgen sKM/1 ->8          # <-- @a became a REFERENCE
6        <0> padav[@a:3,4] lRM ->7
8     <#> gv[IV \"\@"] s ->9

$ perl -MO=Concise -e 'my @a; f(@a); sub f(\@){}'
8  <1> entersub[t3] vKS/TARG ->9
6     <0> padav[@a:1,4] lM ->7       # <-- @a flattened; NO srefgen
```

The prototype is applied only when the definition precedes the call. This is
not a warning or a subtlety — it is a different op tree, and therefore
different types flowing into PSC.

### 3.5.2 The prototype character set

| Char | Effect on the call-site parse | Example | Call site becomes |
|---|---|---|---|
| `$` | Force scalar context on one argument | `sub f($)` | `f @a` → `f(scalar @a)` |
| `@` | Slurp all remaining args in list context | `sub f(@)` | greedy; nothing after it matters |
| `%` | Slurp all remaining args (identical to `@` at parse time) | `sub f(%)` | greedy |
| `&` | Code ref; **if first, allows a bare block with no comma** | `sub f(&@)` | `f { ... } @list` |
| `*` | Typeglob/filehandle slot: a bareword is accepted (no strict-subs error, not a sub call) and arrives as the plain string of its name; a glob or globref arrives as a globref | `sub f(*)` | `f STDIN` → `f('STDIN')` (verified: `ref \$_[0]` is `SCALAR`, value `STDIN`) |
| `;` | Separates required from optional | `sub f($;$)` | second arg may be omitted |
| `\X` | Take a **reference** to the next argument, which must be of type X | `sub f(\@)` | `f(@a)` → `f(\@a)` |
| `\[XYZ]` | Reference to any of the listed types | `sub f(\[@%])` | `f(%h)` → `f(\%h)` |
| `+` | Like `\[@%]` if the argument is a hash/array, else like `$` | `sub f(+)` | polymorphic |
| `_` | Like `$`, but defaults to `$_` when omitted | `sub f(_)` | `f()` → `f($_)` |

The `\X` and `+` forms are the only ones that change the *shape* of the AST
(inserting an `srefgen`). `$`, `_`, `*` and `+` change how far argument
parsing extends. `&` in first position changes the *token grammar*, allowing
a block where a term was expected.

### 3.5.3 How toke.c uses them

Two independent consumers:

**`yyl_subproto` (`toke.c:6588-6640`)**, reached only from
`yyl_constant_op` (`7973`) once `yyl_just_a_word` has found a CV, decides
the *token class* of the sub name, which determines its grammatical
precedence. An empty prototype makes the name a zero-argument term outright
(`TERM(FUNC0SUB)`, `6596`) — this is what `use constant` relies on:

```c
if ((( *proto == '$' || *proto == '_' || *proto == '*' || *proto == '+')
     && proto[1] == '\0')
 || ( *proto == '\\' && proto[1] && proto[2] == '\0'))
{
    UNIPROTO(UNIOPSUB, optional);   /* named unary operator precedence */
}
...
if (*proto == '&' && *s == '{') {
    PREBLOCK(LSTOPSUB);             /* allows f { ... } with no comma */
}
```

A single-character prototype makes the sub a **named unary operator**, which
binds tighter than a comma. `sub f($); f $x, $y` parses as `f($x), $y`, not
`f($x, $y)`. The precedence class comes from `perly.y:169` (`%nonassoc UNIOP
UNIOPSUB`), which sits between comparison and shift operators — so `f $a + 1`
is `f($a + 1)` while `f $a > 1` is `f($a) > 1`.

**`S_intuit_method` (`toke.c:5087-5094`)** uses the prototype to *suppress*
the indirect-object reading:

```c
if (cv && SvPOK(cv)) {
    const char *proto = CvPROTO(cv);
    if (proto) {
        while (*proto && (isSPACE(*proto) || *proto == ';')) proto++;
        if (*proto == '*') return 0;    /* takes a filehandle: not a method */
    }
}
```

### 3.5.4 What happens when the sub has not been seen

`yyl_just_a_word` reaches the `c.cv` lookup at `toke.c:8113-8119`, gets
`NULL`, and `yyl_subproto` is never consulted. What happens next depends only
on the following character, and it is **not** a list-operator parse:

| `NAME` unknown, followed by | Result | toke.c | Verified (5.42.0) |
|---|---|---|---|
| `(` | sub call, unprototyped (`PERLY_AMPERSAND` + `BAREWORD`) | `8195-8212` | `f(@a)` → `entersub`, no `srefgen` |
| `=>` | string | `8177-8193` | — |
| `$var` or `{`, `indirect` on | indirect method call, `XBLOCKTERM` | `8214-8224` | `f $x, 1` → `($x->f, 1)` |
| bareword, `indirect` on | `intuit_method` → method call | `8227-8250` | `f bar` → `'bar'->f` |
| `;` or `}` | bareword string (error under `strict subs`) | `8261-8271` | `f;` → `'f'` |
| anything else (`@a`, `1`, `"s"`) | bareword string, then the parser fails on the term | `8261-8271` | `f @a` → *"Array found where operator expected (Do you need to predeclare "f"?)"* |
| any of the above with `use v5.36` (`indirect` off) | as above minus the two method rows, which become syntax errors | `8215`, `5071` | `f $x, 1` → syntax error |

So there is no deferred fixup: perl commits at the call site. But the
"conservative default" is only the parenthesised form. `f(@a)` before
`sub f(\@)` compiles as an unprototyped call and stays that way. `f @a`
before the definition does not compile at all, and `f $x` becomes a method
call on `$x`.

**What this means for a static parser.** Parse `NAME(...)` to an unknown
`NAME` as a plain call, applying no prototype — that is exactly perl. For
paren-less `NAME term` to an unknown `NAME`, perl has no list-operator
reading; a Go parser that accepts it as one is *more permissive* than perl.
That is a reasonable recovery for an LSP, but it must carry a diagnostic,
because the construct is exactly what perl reports as *"Do you need to
predeclare"*. The divergence from perl on prototypes proper is then confined
to *backward* calls to prototyped subs — and that case is statically visible,
because the definition is in the file, textually earlier.

### 3.5.5 The approximation

A two-pass design captures nearly all of it:

- **Pass 1 (cheap, regex-adjacent):** scan for `sub NAME (PROTO)` and
  `use constant`, building a name → prototype map, recording the byte offset
  of each definition.
- **Pass 2 (real parse):** at a call site to name N at offset O, apply N's
  prototype **only if** its definition offset < O and it is in the same file.
  Otherwise parse unprototyped.

This reproduces perl exactly for same-file definitions, which is the common
case. It diverges for imported prototypes (§3.6).

**Do not** apply prototypes from later definitions, however tempting. Doing
so is *less* accurate than ignoring them, because it disagrees with perl on
code that perl parses one specific way.

---

## 3.6 `BEGIN`, `use`, and why this is undecidable in general

### 3.6.1 The mechanism

`use Foo;` is defined as `BEGIN { require Foo; Foo->import; }`. `BEGIN` runs
the enclosed code *as soon as it is compiled*, before the rest of the file is
parsed. That code can:

- define subs (giving them prototypes that affect later calls),
- install globs (`*name = sub(\@){...}`),
- import prototyped subs from another module,
- install a keyword plugin (`PL_keyword_plugin`, `toke.c:9338-9361`),
- install an infix operator plugin (`PL_infix_plugin`, `toke.c:9363-9387`),
- install an overloaded-constant handler (`$^H{...}`, consumed by
  `S_new_constant` at `toke.c:10568` via `call_sv`),
- install a source filter (§3.7),
- `require` a file whose name is computed at runtime.

Verified, glob installation changing a call-site parse:

```console
$ perl -MO=Deparse -e 'BEGIN { *g = sub(\@){}; } my @a; g(@a);'
sub BEGIN { *g = sub (\@) { }; }
my @a;
&g(\@a);          # the prototype was applied
```

### 3.6.2 The undecidability argument

The reduction is immediate and total:

```perl
BEGIN {
    if (some_arbitrary_computation()) {
        eval 'sub f(\@) {}';    # f gets a prototype
    } else {
        eval 'sub f(@) {}';     # f does not
    }
}
my @a;
f(@a);              # parses as f(\@a) or f(@a) — depends on the halting
                    # behaviour of some_arbitrary_computation()
```

Determining the parse of the final line requires deciding the outcome of
arbitrary Perl, so the parse of a Perl program is undecidable by reduction
from the halting problem. There is no clever engineering around this; it is
a theorem. A static parser must therefore be *approximate by construction*,
and the honest question is not "can we be correct" but "what is the residual
error rate on real code, and can we detect when we are in the bad region."

### 3.6.3 What proportion of real code this affects

Measured over `perl5/t/**/*.t` (620 files), which is a deliberately
adversarial corpus — it is the test suite whose job is to exercise perl's
darkest corners:

| Construct | Files | % of corpus |
|---|---|---|
| Contains `BEGIN` at all | 487 | 78.5% |
| `BEGIN` doing only `@INC`/`chdir` boilerplate (parse-irrelevant) | 448 | 72.3% |
| **`BEGIN` defining a sub or installing a glob (parse-relevant)** | **8** | **1.3%** |
| `use constant` | 32 | 5.2% |
| `sub NAME(PROTO)` definitions | 47 | 7.6% |
| String `eval "..."` | 84 | 13.5% |
| `__DATA__` / `__END__` | 50 | 8.1% |
| `format` declarations | 13 | 2.1% |
| Source filters | 2 | 0.3% |

The headline number — 78.5% of files contain `BEGIN` — is the misleading one,
and quoting it would be exactly the kind of overpromise-in-reverse this spec
should avoid. The number that matters is **1.3%**: the fraction where `BEGIN`
does something that can change how subsequent code parses. The rest is
`BEGIN { unshift @INC, 't/lib' }`.

Two caveats against reading this as good news:

1. Application code is *less* prototype-heavy than perl's test suite, but
   *more* likely to import from modules (`List::Util::reduce` has prototype
   `&@`). Cross-file imports are the dominant real-world case and are not
   measured here.
2. `use constant` (5.2%) creates subs with prototype `()`, which makes the
   name a **term** rather than a list operator. `use constant PI => 3; my $x
   = PI - 1;` parses as `PI() - 1` only because `PI` has an empty prototype
   (`TERM(FUNC0SUB)`, `toke.c:6596`). Without that knowledge perl itself
   would read `'PI' - 1` — a bareword, then an operator — and a static parser
   that treats unknown barewords as list operators would read `PI(-1)`;
   both are wrong. This is the highest-
   frequency prototype-dependent construct in practice, and it is also the
   easiest to special-case: `use constant` is syntactically recognisable.

### 3.6.4 Approximation strategy for `use`

Tiered, in increasing cost:

| Tier | Handles | Cost |
|---|---|---|
| 0 | Hard-code the prototypes of core-adjacent modules: `List::Util` (`reduce`/`first`/`any`/`all` = `&@`), `Scalar::Util`, `POSIX`, `Carp` | a static table; trivial |
| 1 | `use constant NAME => ...` and `use constant { A=>..., B=>... }` → register `NAME` with prototype `()` | source-visible; cheap |
| 2 | `use parent`/`use base` → record inheritance for PSC; **no parse impact** | cheap |
| 3 | Resolve `use Foo` against a workspace index of `Foo.pm` parsed by the same parser, extracting `@EXPORT`/`@EXPORT_OK` and prototypes | needs a module index and a dependency graph; this is where an LSP earns its keep |
| 4 | Actually execute `BEGIN` | out of scope — it is a Perl interpreter |

Tier 3 is the right stopping point for an LSP. It is unsound (a module can
build `@EXPORT` at runtime) but it is *sound in practice* for the ~99% of
modules that use a literal `our @EXPORT = qw(...)`.

PerlOnJava, by contrast, sits at tier 4: it is a runtime, so it looks up
`GlobalVariable.getGlobalCodeRef(fullName)` during parsing
(`ParsePrimary.java:468`) and gets the real answer. Note the comment at
`ParsePrimary.java:357`, which is a hard-won bug report worth heeding:

> Check if the sub exists at compile time. Do NOT use `getGlobalCodeRef()`
> here: it autovivifies an undefined CV placeholder, and
> `RuntimeScalar.getDefinedBoolean()` returns true for every CODE slot — so
> `::e` was always treated as `main::e`.

Even with a live symbol table, "does this sub exist" is easy to get wrong.
A static parser has no such placeholder problem, because its map contains
only names it actually saw.

---

## 3.7 Source filters: the hard stop

### 3.7.1 The mechanism

A source filter is a coderef inserted into `PL_rsfp_filters` by
`Perl_filter_add` (`toke.c:5164-5230`). Every subsequent line of source is
passed through it by `Perl_filter_read` (`toke.c:5268`) before the lexer sees
it, reached from `S_filter_gets` (`toke.c:5387-5395`):

```c
S_filter_gets(pTHX_ SV *sv, STRLEN append)
{
    if (PL_rsfp_filters) {
        ...
        if (FILTER_READ(0, sv, 0) > 0)
```

The filter is arbitrary Perl. `Filter::Simple` is commonly used to implement
whole alternative syntaxes. The bytes on disk have, in the general case, no
relationship to the bytes perl parses.

**A correction worth making explicit**, because it is a natural place to go
wrong: the `call_sv` at `toke.c:10568` is *not* the source-filter call site.
It is inside `S_new_constant`, which dispatches overloaded constant handlers
registered in `$^H` (`use overload`, `use bignum`). That is a real
lexer-feedback mechanism and worth knowing about — it lets a pragma change
what the literal `1.5` compiles to — but it operates per-literal, not on the
source text. Source filters go through `FILTER_READ`.

### 3.7.2 The only honest response

Undecidable, and unlike prototypes there is no useful partial credit: you
cannot analyse *any* of the file, because you do not know what the file
contains. Detect and bail.

Detection is a `use` of a known filter module. perl-lsp does exactly this
(`crates/perl-parser-core/src/engine/parser/helpers.rs:312`):

```rust
/// Check if a module is a source filter (security risk)
fn is_filter_module(module: &str) -> bool {
    matches!(module,
        "Filter" | "Filter::Util::Call" | "Filter::Simple" | "Filter::cpp"
        | "Filter::exec" | "Filter::sh" | "Filter::tee" | "Filter::decrypt")
}
```

Recommended behaviour for the Go parser:

1. On `use <filter module>`, mark the file `Unanalyzable{Reason: SourceFilter}`.
2. Emit **no** diagnostics for the file — false errors on a file you cannot
   read are worse than silence.
3. Still provide document symbols from a best-effort lexical scan, clearly
   flagged as low-confidence. An LSP that goes completely dark is a worse
   experience than one that offers degraded navigation.
4. Do not feed the file to PSC. Type inference over a file whose text you do
   not have is not conservative, it is fictional.

The list above is necessarily incomplete — anyone can write a filter module —
so treat `filter_add` appearing in a `use`d local module as a signal too, if
your workspace index reaches that far. Two files in perl5/t use filters.

---

## 3.8 Barewords: the full resolution order

A bareword is the hardest single token in Perl, because it can be a sub call,
a string, a filehandle, a class name, a label, a package qualifier, or a hash
key. Perl resolves it in `yyl_keylookup` (`toke.c:9305`) then
`yyl_just_a_word` (`toke.c:8023`). The order is strict and each step can
terminate.

### 3.8.1 In `yyl_keylookup` — before it is even "a word"

| # | Test | Result | Line |
|---|---|---|---|
| 1 | Followed by `::` and word is `CORE` | `CORE::` operator | 9324 |
| 2 | Followed by `::` otherwise | fall to `just_a_word` as package name | 9326 |
| 3 | Followed by `=>` (fat comma) | **auto-quoted string** | 9333-9335 |
| 4 | Registered keyword plugin claims it | `PLUGSTMT` / `PLUGEXPR` | 9338-9361 |
| 5 | Registered infix plugin claims it | custom infix operator | 9363-9387 |
| 6 | `PL_expect == XSTATE` and followed by `:` (not `::`) | **LABEL** | 9389-9396 |
| 7 | Lexical sub in scope (`pad_findmy_pvn` on `&name`), unless `XOPERATOR` | lexical sub call | 9399-9433 |
| 8 | A built-in keyword (`keyword()`) | that keyword's handler, via `yyl_word_or_keyword` | 9436, 9456 |
| 9 | Keyword **and** followed by `=>` (possibly on the next line) | auto-quoted string | 9441-9453 |
| 10 | otherwise | `yyl_word_or_keyword` → `yyl_just_a_word` | 9456 |

Step 3 before step 4, and step 9 after step 8, together mean: **`=>` always
wins over everything, including keywords.** `print => 1` is the string
`"print"`. This makes `=>` the single most reliable disambiguator in the
language and worth checking early in a Go implementation.

Step 6 is why a label is only recognised at statement start. `FOO: while(){}`
is a label; `$h{FOO}: ...` is not.

### 3.8.2 In `yyl_just_a_word`

| # | Test | Result | Line |
|---|---|---|---|
| 1 | `PL_expect == XOPERATOR` | error/warning: two terms in a row | 8033 |
| 2 | Followed by `'` or `::` | extend the package-qualified name | 8046-8062 |
| 3 | Name ends in `::` | bareword package name, done | 8068-8081 |
| 4 | Directly after a list/unary op **and** in an indirect-object slot (`XREF`, or that op takes `OA_FILEREF`) | try `intuit_method` (§3.8.3) | 8122-8145 |
| 5 | …and no `(` follows, and (prior op is `sort`, **or** no known sub exists and prior op is not `map`/`grep`) | **indirect object / filehandle** | 8156-8172 |
| 6 | `_` following a filetest operator | bareword | 8162-8164 |
| 7 | Followed by `=>` | auto-quoted string | 8177-8193 |
| 8 | Followed by `(` | **sub call**, unconditionally | 8195-8212 |
| 9 | Followed by `$` or `{`, no known sub, indirect enabled | **method call** (`METHCALL0`) | 8214-8224 |
| 10 | Followed by bareword or `$`, `intuit_method` says yes | method call | 8227-8250 |
| 11 | A known sub exists (`c.cv`) | sub call (constant-folded if constant) | 8253-8259 |
| 12 | `strict subs` in effect | compile error (`OPpCONST_STRICT`) | 8263 |
| 13 | otherwise | **bareword string** | 8261-8271 |

Steps 5, 9, 10 and 11 all consult the symbol table. Step 8 does not:
**a bareword followed by `(` is always a sub call**, which is the escape
hatch a static parser leans on hardest.

### 3.8.3 `S_intuit_method`

`toke.c:5046-5144`. Its own docblock is the clearest statement of the rules:

> - Not a method if foo is a filehandle.
> - Not a method if foo is a subroutine prototyped to take a filehandle.
> - Not a method if it's really `Foo $bar`.
> - Method if it's `foo $bar`.
> - Not a method if it's really `print foo $bar`.
> - Method if it's really `foo package::` (interpreted as `package->foo`).
> - Not a method if bar is known to be a subroutine (`sub bar; foo bar`).
> - Not a method if bar is a filehandle or package, but is quoted with `=>`.

Plus one purely static gate at the top (`toke.c:5071`):

```c
if (!FEATURE_INDIRECT_IS_ENABLED)
    return 0;
```

**`no feature 'indirect'` — the default under `use v5.36` and later — deletes
this entire function.** That is the most valuable single fact in this
section. Any file with `use v5.36;` or newer, or an explicit
`no feature 'indirect';`, has no indirect-object syntax at all, and every
symbol-table dependency in steps 4, 5, 9 and 10 above evaporates. A static
parser is *exactly correct* on barewords in modern Perl.

`isUPPER(*PL_tokenbuf)` at `toke.c:5100` encodes the community convention
that `new Foo` is a class method but `foo $bar` is a method call on `$bar`;
capitalisation is load-bearing in perl's own heuristic, which licenses a Go
parser using the same signal.

---

## 3.9 The approximation strategy

This is the section the rest of the chapter exists to support. For each
construct: what perl does, what a static parser can do, how it fails, and how
to verify.

### 3.9.1 The verification protocol

Every approximation here is empirically checkable, because perl will report
its own parse decisions. Four probes:

| Probe | Command | Reveals |
|---|---|---|
| Prototype | `perl -e 'require F; print prototype(\&F::f)'` | the exact prototype string |
| Op tree | `perl -MO=Concise -e '...'` | `srefgen` presence = prototype applied |
| Round-trip | `perl -MO=Deparse -e '...'` | perl's parse, re-rendered as Perl |
| Syntax only | `perl -c file` | whether it parses at all |

`Deparse` is the workhorse: it renders perl's actual parse back into source,
so a differential test is "does our AST, pretty-printed, match `Deparse`
output modulo normalization?" `Concise` is the tiebreaker when Deparse output
is ambiguous — as with prototypes, where Deparse shows `f(@a)` for both
parses but Concise shows the `srefgen`.

Verified example of the whole loop:

```console
$ perl -e 'sub f(\@%){} print prototype(\&f)'
\@%
$ perl -e 'print prototype("CORE::push")'
\@@
```

Core ops have prototypes too, and `prototype("CORE::name")` gives them — a
free source of truth for the built-in table.

### 3.9.2 The table

| Construct | (a) What perl does | (b) Static approximation | (c) Failure mode when wrong | (d) How to verify |
|---|---|---|---|---|
| `/` divide vs regex | `PL_expect` | Reproduce `PL_expect` exactly | Regex lexed as division → cascading garbage to end of statement | `Deparse`; corpus diff |
| `{` block vs hashref | `PL_expect`, else the lookahead heuristic (§3.3) | Reproduce both, including the heuristic verbatim | Block parsed as hashref → spurious syntax error | `Deparse`: `{;` vs `+{` in output |
| `<` readline vs `<` | `PL_expect` | Reproduce | `if ($a < $b)` swallowed as readline | `Deparse` |
| `%` modulo vs hash | `PL_expect` | Reproduce | Rare; symmetric to `/` | `Deparse` |
| `$x[` in code | `intuit_more` → TRUE at step 5 (not in pattern) | Always a subscript outside regex | none — perl does the same | `Deparse` |
| `$x[` in regex | Weight heuristic + symbol table | Port the weights; skip the two symbol-table branches (treat unknown ids as "not found", −10 not −100) | Character class read as subscript or vice versa → wrong regex semantics, no syntax error | `Concise`: `helem`/`aelem` op present or absent |
| `%h[...]`, `@h{...}` sigil rewrite | `intuit_more` in `yyl_snail`/`yyl_percent` | Reproduce; purely syntactic | Wrong container type to PSC | `Deparse` shows `@h{...}` vs `%h{...}` |
| `sort $cr @l` | `PL_last_lop_op == OP_SORT` short-circuit | Special-case `sort`: first term with no following comma is the comparator | Comparator read as list element | `Deparse`: `sort $cr @l` vs `sort($cr, @l)` |
| Prototype, same file, def **before** call | Applies prototype | Two-pass; apply when def offset < call offset | — correct | `Concise`: `srefgen` |
| Prototype, same file, def **after** call | Does **not** apply | Do not apply | — correct, bug-compatible | `Concise`: no `srefgen` |
| Prototype, imported | Applies (module already compiled) | Tier 0–3 (§3.6.4): hard-coded table, then workspace index | Missing `srefgen`; `f {...} @l` for `&@` protos may be a hard syntax error | `prototype(\&F::f)` after `require` |
| `use constant NAME => v` | Sub with `()` prototype; name is a term | Recognise `use constant` syntactically, register `()` | `PI - 1` read as `PI(-1)` | `Deparse` |
| Bareword + `(` | Always a sub call | Same | — correct | — |
| Bareword + `=>` | Always a string | Same | — correct | — |
| Bareword at stmt start + `:` | Label | Same (needs `XSTATE`) | Label read as ternary | `Deparse` |
| Bareword, `no feature 'indirect'` | `intuit_method` disabled | Detect `use v5.36+` / `no feature 'indirect'`; then bareword rules are fully static | — correct | `perl -c` |
| Bareword, indirect enabled, unknown sub | `intuit_method` → indirect object | Heuristic: `isUPPER` first char + following `$var`/bareword → method call, mirroring `toke.c:5100` | `new Foo` vs `new(Foo)` | `Deparse` |
| `print FH LIST` | `OA_FILEREF` on the prior op | Table of filehandle-taking builtins (`print`, `printf`, `say`, `open`, `close`, `binmode`) | Filehandle read as first list element | `Deparse` |
| `BEGIN` defining subs | Executes it | Scan `BEGIN` blocks for literal `sub NAME` and `*name = sub` and register them; ignore conditionals | Missing prototype → unprototyped parse (perl's own forward-reference behaviour) | `Concise` |
| `BEGIN` with computed defs | Executes it | **Cannot.** Mark the file low-confidence | Silent wrong parse | none — this is the undecidable core |
| String `eval "..."` | Parses at runtime | Do not parse the string; treat as opaque | No symbols from inside | — |
| Source filter | Rewrites source | **Bail out** (§3.7) | Everything | detect `use Filter::*` |
| `__DATA__` / `__END__` | Everything after is data | Stop lexing at the marker | Data parsed as code → garbage errors | trivial |
| `format` | Custom sub-lexer, `.` terminated | Skip to a lone `.` on its own line | Format body parsed as code | `perl -c` |
| Keyword/infix plugins | `PL_keyword_plugin` | **Cannot.** Detect known plugin modules and degrade | Syntax error on valid code | detect the `use` |
| Overloaded constants (`$^H`) | `S_new_constant` calls a handler | Ignore — affects values, not the parse shape | Wrong literal type only | — |

### 3.9.3 The confidence-level design

Rather than a boolean parsed/failed, carry a per-file confidence, since an
LSP must decide how loudly to complain:

| Level | Condition | LSP behaviour |
|---|---|---|
| `Exact` | No prototypes, no parse-relevant `BEGIN`, no filters, `no feature 'indirect'` | Full diagnostics; feed PSC |
| `High` | Same-file prototypes only, or imports resolved from the workspace index | Full diagnostics; feed PSC |
| `Degraded` | Unresolved imports of prototyped subs; `BEGIN` with conditional definitions | Suppress "undefined sub" and arity diagnostics; feed PSC with widened types |
| `Unanalyzable` | Source filter, keyword plugin | Lexical symbols only; no diagnostics; do not feed PSC |

The rule that keeps this honest: **a parser that is uncertain must widen, not
guess.** PSC's lattice has a top element; use it. A wrong concrete type is
worse than an unknown one, because it produces a confident false diagnostic,
and confident false diagnostics are how a language server loses its users.

### 3.9.4 Comparison: how the two reference implementations approximate

**PerlOnJava** is a runtime, so it does not approximate — it consults the
live symbol table during parsing, exactly like perl.
`PrototypeArgs.consumeArgsWithPrototype` (`PrototypeArgs.java:252`) receives
a real prototype string and dispatches on it, and `ParsePrimary.java:468`
looks up `GlobalVariable.getGlobalCodeRef(fullName)`. That its
`PrototypeArgs.java` is 1506 lines is the measure of the problem's true size
*when you have complete information*. Its structure is worth stealing even if
its information source is not: note `isNamedUnaryPrototype`
(`PrototypeArgs.java:157`) recognising `$`, `_`, `;$`, `;_` — precisely
`yyl_subproto`'s `UNIPROTO` condition at `toke.c:6603-6615`.

PerlOnJava also carries pragmatic special cases a static parser will need:
`isListUtilCallback` (`PrototypeArgs.java:36`) hard-codes `reduce`, `any`,
`all`, `first`, `pairmap` and friends — tier 0 of §3.6.4, in production.

**perl-lsp** approximates aggressively and states so. Its
`crates/perl-lexer/src/mode.rs:39-63` collapses perl's eleven states into five:

```rust
pub enum LexerMode {
    ExpectTerm,       // slash starts a regex
    ExpectOperator,   // slash is division
    ExpectDelimiter,  // quote-like operator delimiter; '#' is not a comment
    InFormatBody,
    InDataSection,
}
```

The mapping is `XTERM|XREF|XSTATE|XBLOCK|XTERMBLOCK|XATTR* → ExpectTerm` and
`XOPERATOR|XPOSTDEREF → ExpectOperator`, with `XTERMORDORDOR` lost. Its own
documented transition table (`mode.rs:19-28`) is by *previous token*, not by parser state — a
strictly weaker model, though it captures the common cases:

| Previous token | Next mode |
|---|---|
| identifier, number, closing paren/bracket | `ExpectOperator` |
| keyword, word-operator, operator, opening paren/bracket | `ExpectTerm` |

`ExpectDelimiter` is an addition perl handles differently (via `PL_lex_state`
and `word_takes_any_delimiter`, `toke.c:5472`), and it is a good idea: it is
what stops `s#a#b#` from being read as a comment. Keep it as a distinct
state in Go.

perl-lsp's honest bail-out on filters (`helpers.rs:312`) and its budget
guards (`MAX_REGEX_BYTES`, 64KB, with graceful `UnknownRest` degradation) are
the right instincts for an LSP: never hang, always return something.

**Recommendation for the Go parser: eleven states, not five.** The five-state
model will mis-lex `-e $f // die` (no `XTERMORDORDOR`), `print {$fh} $x`
(no `XREF`), and `sub f :lvalue {}` (no `XATTRBLOCK`). It cannot express
`XSTATE`, so perl's rule that POD is recognized only at a statement boundary
(chapter 2 §2.5.1) is unreachable, and it decides `{` with ad-hoc fields
(`hash_brace_depth`, `after_sub`, `after_arrow`; `checkpoint_impl.rs:29-34`)
instead of the state — defensible for an LSP that must never hang, wrong for
a compiler. The extra six states
cost a handful of enum values and a bracket stack; they buy exact
compatibility on the entire category-1 problem, which is the part that is
actually solvable. Do not economise here — economise on the symbol-table
approximations, where the effort has diminishing returns.

---

## 3.10 Incremental re-parse: what feedback costs an LSP

The chapter's conclusions constrain partial re-parsing, so state the
constraint explicitly.

**Lexer state is not recoverable from a byte offset.** You cannot resume
lexing at the edit point, because tokenizing requires `PL_expect`, the
bracket stack, `PL_lex_state`, and (for `$x[` inside regexes) the accumulated
prototype/sub map. A resumable lexer state must therefore carry `Expect`,
the bracket stack, `PL_lex_state`, the heredoc queue and `PL_last_lop_op`
(for the `sort` and filehandle cases of §3.4.4); the struct itself is chapter
6 §6.3.1. Snapshot it at **statement boundaries** (`Expect == XSTATE`, both
stacks empty) and you get safe restart points. In practice statement
boundaries at brace depth 0 are frequent enough that a re-parse after an edit
touches one sub, not the file.

**Prototype changes invalidate the whole file after the definition point.**
Editing `sub f(\@)` to `sub f(@)` changes how every subsequent call site
parses. The dependency is one-directional (later text depends on earlier),
so the invalidation rule is simple: an edit at offset O invalidates parse
results from O to EOF, but nothing before it. Combined with statement
snapshots, this is an O(rest-of-file) worst case and O(one statement)
typically — acceptable.

**A `BEGIN` block edit invalidates the file and every file that `use`s it.**
This is the one place the LSP needs a cross-file dependency graph.

---

## 3.11 Verdict: is this feasible?

Yes, with a stated and measured error rate, and the split is clean:

**Solvable exactly (category 1).** The `PL_expect` machine, the bracket
stack, `intuit_more` outside patterns, the `{` heuristic, `=>` auto-quoting,
labels, `sort`, bareword-plus-paren, heredocs, `__DATA__`, formats. This is
the large majority of the disambiguation surface, it is all in this chapter,
and a careful Go implementation should reach ~100% on it. It is also the part
where getting it wrong is catastrophic (a mis-lexed regex corrupts everything
downstream), so it deserves the implementation effort.

**Solvable in practice (category 2, benign).** Same-file prototypes,
`use constant`, hard-coded core module prototypes, workspace-indexed imports.
Correct on the overwhelming majority of real code; the residue degrades to
perl's own forward-reference behaviour, which is the *right* failure mode —
you get the parse perl gives a program whose subs are declared later.

**Not solvable (category 2, malignant).** Computed `BEGIN` blocks, source
filters, keyword plugins, runtime `@EXPORT` manipulation, string `eval`
defining subs. Provably undecidable. Detect, degrade, and say so. Measured at
1.3% + 0.3% of perl5/t, and lower in application code.

The design failure to avoid is not the 1.6% — it is *pretending* the 1.6%
does not exist, shipping confident diagnostics on files the parser
misunderstood, and training users to ignore the diagnostics. Build the
confidence levels of §3.9.3 in from the start; retrofitting honesty into a
parser that was designed to always answer is much harder than designing for
"I don't know" from the beginning.

---

## 3.12 Implementation checklist

Acceptance list; build order is chapter 6 §6.11.

- [ ] `Expect` enum with all 11 states, including `XTERMORDORDOR`.
- [ ] Transition driven by emitted-token class (the macro table of §3.1.3), not per-site assignment.
- [ ] `BrackStack []Expect`, pushed in the `{` handler per the §3.1.4 table.
- [ ] `intuit_more`: step 5 (outside pattern → subscript) first; weights only for `[` inside patterns.
- [ ] `{` disambiguation heuristic of §3.3, verbatim, including the lowercase-bareword asymmetry.
- [ ] Bareword resolution in the exact order of §3.8.1 and §3.8.2; `=>` checked first.
- [ ] Detect `use v5.36+` / `no feature 'indirect'` and disable `intuit_method` entirely.
- [ ] Two-pass prototype collection with definition-offset ordering (§3.5.5).
- [ ] `use constant` → register `()` prototype.
- [ ] Tier-0 hard-coded prototype table (`List::Util` &c.); `prototype("CORE::x")` to generate the builtin table.
- [ ] Source-filter and keyword-plugin detection → `Unanalyzable`, no diagnostics.
- [ ] Per-file confidence level plumbed into both diagnostics and PSC.
- [ ] Differential test harness: parse → pretty-print → compare against `perl -MO=Deparse`, with `-MO=Concise` for prototype cases.
- [ ] Corpus run over `perl5/t/**/*.t` (620 files) with a per-file pass/fail/degraded report, tracked over time.

The last item is the one that keeps this spec honest. Every claim in this
chapter is a testable prediction; a spec whose predictions are never checked
decays into folklore within a release or two.
