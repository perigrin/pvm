<!-- ABOUTME: Specification of Perl's lexical structure — the character- and token-level layer -->
<!-- ABOUTME: for implementers writing a from-scratch Perl lexer in Go using only the standard library. -->

# Chapter 2 — Lexical Structure

## 2.0 Scope, sources, and how to read this chapter

This chapter specifies the character-level and token-level layer of Perl: how a
byte stream becomes a token stream. It is written for someone implementing a
Perl lexer in Go with nothing but the standard library — no tree-sitter, no
CGO, no generated-parser tooling — feeding an LSP that must support partial
re-parses and a downstream type checker (PSC).

Three sources were triangulated. They are not equal:

| Source | Role | Location |
| --- | --- | --- |
| `perl5/toke.c` (14708 lines) + `perly.y` | **Ground truth.** Every rule here was checked against it. | `/home/perigrin/dev/perl5/` |
| PerlOnJava (Java, recursive descent) | A working from-scratch implementation. Useful for "how did someone else structure this". | `/home/perigrin/dev/PerlOnJava/` |
| perl-lsp (Rust) | A working LSP-oriented lexer. Useful for incremental/checkpoint design. | `/home/perigrin/dev/perl-lsp/crates/perl-lexer/` |

Citations look like `toke.c:10568`, `NumberParser.java:214`, `lib.rs:880`.
Where a claim was verified by running perl 5.42.0 directly, it is marked
**[verified]**. Several such checks contradicted one or more of the secondary
implementations; those are recorded as Divergence notes rather than smoothed
over. A spec that is honest about approximation is more useful than one that
claims false completeness.

> **Divergence** notes flag a place where an implementation is known to
> approximate rather than match perl. They are a to-do list of accepted risk,
> not incidental trivia.

### 2.0.1 The one thing to understand before anything else

**Perl cannot be lexed independently of parsing.** There is no token stream
that can be produced by a context-free scanner. The same characters lex
differently depending on whether the parser expects a term or an operator:

```perl
$x = /foo/;     # / begins a regex
$x = $y /foo/;  # / is division, twice — "$y divided by foo divided by..."
print $x %h;    # % begins a hash sigil
print $x % 5;   # % is modulo
```

`toke.c` resolves this with a state variable, `PL_expect`, holding an
`expectation` enum (`perl.h:5972-5985`). Every lexer for Perl must carry an
equivalent. Chapter 3 §3.1 specifies it; §2.2 says how it meets the rest of
the lexer.

A consequence for the LSP use case: **the lexer is not resumable from an
arbitrary byte offset.** Restarting mid-file requires restoring the whole
context stack (expectation state, bracket depth, heredoc queue, sublex stack,
POD/`__DATA__` flags). §2.14 lists what the lexer must expose to a
checkpoint; chapter 6 §6.3 owns the checkpoint itself.

---

## 2.1 The token inventory

### 2.1.1 What `perly.y` actually declares

Perl's real grammar declares roughly 90 terminals. They fall into groups
(`perly.y:46-94`):

**Grammar entry pseudo-tokens** (`perly.y:46`) — injected by the caller to
select a start symbol, never produced from source text:

`GRAMPROG` `GRAMEXPR` `GRAMBLOCK` `GRAMBARESTMT` `GRAMFULLSTMT` `GRAMSTMTSEQ` `GRAMSUBSIGNATURE`

**Single-character punctuation** (`perly.y:49-64`), each its own terminal:

| Token | Char | Token | Char |
| --- | --- | --- | --- |
| `PERLY_AMPERSAND` | `&` | `PERLY_PERCENT_SIGN` | `%` |
| `PERLY_BRACE_OPEN` | `{` | `PERLY_PLUS` | `+` |
| `PERLY_BRACE_CLOSE` | `}` | `PERLY_SEMICOLON` | `;` |
| `PERLY_BRACKET_OPEN` | `[` | `PERLY_SLASH` | `/` |
| `PERLY_BRACKET_CLOSE` | `]` | `PERLY_SNAIL` | `@` |
| `PERLY_COMMA` | `,` | `PERLY_STAR` | `*` |
| `PERLY_DOLLAR` | `$` | `PERLY_COLON` | `:` |
| `PERLY_DOT` | `.` | `PERLY_QUESTION_MARK` | `?` |
| `PERLY_EQUAL_SIGN` | `=` | `PERLY_EXCLAMATION_MARK` | `!` |
| `PERLY_MINUS` | `-` | `PERLY_TILDE` | `~` |
| | | `PERLY_PAREN_OPEN` / `PERLY_PAREN_CLOSE` | `(` `)` |

(The last four appear in the precedence declarations at `perly.y:156-165`
rather than in a `%token` line, but they are terminals all the same.)

**Keyword tokens** (`perly.y:67-81`). Note these are *specific* keywords the
grammar knows structurally — most Perl builtins are **not** here, they arrive
as `FUNC0`/`FUNC1`/`FUNC`/`UNIOP`/`LSTOP` with an opcode payload:

`KW_FORMAT` `KW_PACKAGE` `KW_CLASS` `KW_LOCAL` `KW_MY` `KW_FIELD` `KW_IF`
`KW_ELSE` `KW_ELSIF` `KW_UNLESS` `KW_FOR` `KW_UNTIL` `KW_WHILE` `KW_CONTINUE`
`KW_GIVEN` `KW_WHEN` `KW_DEFAULT` `KW_TRY` `KW_CATCH` `KW_FINALLY` `KW_DEFER`
`KW_REQUIRE` `KW_DO` `KW_USE_or_NO`

Note `use` and `no` collapse into a single terminal `KW_USE_or_NO`
(`perly.y:76`); the distinction is carried in the semantic value.

`sub` splits into **four** terminals depending on named-vs-anonymous and
whether signatures are enabled (`perly.y:80`), and `method` into two
(`perly.y:81`):

`KW_SUB_named` `KW_SUB_named_sig` `KW_SUB_anon` `KW_SUB_anon_sig`
`KW_METHOD_named` `KW_METHOD_anon`

This is a lexer responsibility, not a parser one — the lexer must look ahead
past `sub` to decide.

**Semantic-value tokens** (`perly.y:84-94`):

| Token | Carries | Notes |
| --- | --- | --- |
| `BAREWORD` | identifier | An unrecognized word |
| `METHCALL0` / `METHCALL` | method name | Zero-arg vs general method call |
| `ATTRLIST` | attribute list | `:lvalue :method` |
| `THING` | a constant op | Result of a completed literal (string, number, qw list) |
| `PMFUNC` | pattern-match op | `m//`, `s///`, `tr///`, `qr//` |
| `PRIVATEREF` | pad entry | Lexical variable reference |
| `QWLIST` | list of constants | `qw()` result |
| `FUNC0OP` `FUNC0SUB` `UNIOPSUB` `LSTOPSUB` | op/CV | Named-sub call forms |
| `PLUGEXPR` `PLUGSTMT` | plugin op | `PL_keyword_plugin` results |
| `LABEL` | label name | `FOO:` |
| `PROTOTYPE` | prototype string | `($$;@)` |
| `LOOPEX` | opcode | `next` `last` `redo` |
| `DOTDOT` | opcode | `..` and `...` |
| `YADAYADA` | — | `...` as a statement |
| `FUNC0` `FUNC1` `FUNC` | opcode | Builtins by arity |
| `UNIOP` `LSTOP` `BLKLSTOP` | opcode | Named unary / list operators |
| `POWOP` `MULOP` `ADDOP` | opcode | `**`; `* / % x`; `+ - .` |
| `DOLSHARP` | — | `$#` |
| `HASHBRACK` | — | `{` known to start an anon hash |
| `NOAMP` | — | Sub call without `&` |
| `COLONATTR` | — | The `:` introducing attributes |
| `FORMLBRACK` `FORMRBRACK` | — | Format-picture brackets |
| `SUBLEXSTART` `SUBLEXEND` | — | Interpolation sub-lex boundaries |
| `PHASER` | opcode | `BEGIN` `END` `INIT` `CHECK` `UNITCHECK` |

**Operator-class tokens** from the precedence table (`perly.y:150-180`):

`OROP` (`or`/`xor`), `ANDOP` (`and`), `NOTOP` (`not`), `ASSIGNOP`, `OROR` (`||`),
`DORDOR` (`//`), `ANDAND` (`&&`), `BITOROP`, `BITANDOP`, `CHEQOP`/`NCEQOP`
(chained vs non-chained equality), `CHRELOP`/`NCRELOP` (relational),
`SHIFTOP`, `MATCHOP` (`=~` `!~`), `UMINUS`, `REFGEN` (`\`), `PREINC` `PREDEC`
`POSTINC` `POSTDEC`, `POSTJOIN`, `ARROW`.

Plus the operator-plugin terminals `PLUGIN_LOW_OP`, `PLUGIN_LOGICAL_OR_LOW_OP`,
`PLUGIN_LOGICAL_AND_LOW_OP`, `PLUGIN_ASSIGN_OP`, `PLUGIN_LOGICAL_OR_OP`,
`PLUGIN_LOGICAL_AND_OP`, `PLUGIN_REL_OP`, `PLUGIN_ADD_OP`, `PLUGIN_MUL_OP`,
`PLUGIN_POW_OP`, `PLUGIN_HIGH_OP` — these exist for `PL_infix_plugin`
(`toke.c:130-132`). A from-scratch implementation may omit them; note the
omission, because CPAN modules like `Syntax::Operator::*` use them.

Note the distinction between `CHEQOP`/`CHRELOP` (chainable, supporting
`$a < $b < $c`) and `NCEQOP`/`NCRELOP` (non-chainable). This is a *lexer*
classification decision, made per-operator.

### 2.1.2 How the two secondary implementations disagree

This is the single most important structural finding of the comparison, and it
should inform the Go design directly.

**PerlOnJava has 7 token types.** The entire enum (`LexerTokenType.java:13-58`):

`WHITESPACE` `NEWLINE` `IDENTIFIER` `NUMBER` `OPERATOR` `STRING` `EOF`

That is not a Perl token inventory — it is a *pre-tokenizer*. PerlOnJava's
lexer deliberately does almost no Perl-specific work; it splits the byte stream
into coarse lexemes and leaves every context-sensitive decision to the parser.
Its own class comment is refreshingly blunt (`Lexer.java:36-42`):

> The Lexer is optimized for speed rather than accuracy. This can lead to
> issues in cases such as: `qq<=>` — An equal-sign string is parsed as `qq` and
> `<=>`; `10E10` — A floating-point number is parsed as `10` and `E10`. The
> Parser is aware of these issues and implements workarounds.

Confirmed by reading the code: `consumeNumber` (`Lexer.java:198-204`) consumes
only `[0-9_]+` — no decimal point, no exponent, no radix prefix. All numeric
structure is reassembled later in `NumberParser.java`. Likewise there is no
heredoc, regex, POD, or `__DATA__` handling in the lexer at all.

**perl-lsp has ~45 token types** (`token.rs:24-127`), and they are Perl-aware:
separate variants for `RegexMatch`, `Substitution`, `Transliteration`,
`QuoteRegex`, `QuoteSingle`, `QuoteDouble`, `QuoteWords`, `QuoteCommand`,
`HeredocStart`, `HeredocBody`, `FormatBody`, `Version`, `Pod`, `DataMarker`,
`DataBody`, plus `InterpolatedString(Vec<StringPart>)` which pre-decomposes
interpolation into literal/variable/expression parts (`token.rs:8-21`).

It also carries two token types that exist purely for **error recovery** —
essential for an LSP and absent from both perl and PerlOnJava:

- `UnknownRest` — emitted when a complexity budget is exceeded, swallowing the
  remainder of input rather than hanging (`token.rs:70`).
- `Error(Arc<str>)` — a malformed-token marker that keeps lexing (`token.rs:126`).

`TokenType::is_trivia()` (`token.rs:131`) and `is_recovery_token()`
(`token.rs:136`) classify these.

**Recommendation for the Go implementation.** Follow perl-lsp's granularity,
not PerlOnJava's. The PSC type checker needs to distinguish `qw()` from a
string from a regex; recovering that downstream from an `OPERATOR` soup is
strictly harder than recording it at scan time, when the delimiter state is
already in hand. Retain trivia (whitespace, comments, POD) as tokens with
spans rather than discarding them — an LSP needs them for formatting, folding,
and comment-attachment, and a `THING` token that has lost its source span
cannot be mapped back to a diagnostic range.

### 2.1.3 Recommended Go token set

```go
type Kind uint8

const (
    // Trivia — retained, never discarded
    Whitespace Kind = iota
    Newline
    Comment       // # to end of line
    Pod           // =word ... =cut
    // Literals
    Number        // all radices, floats, hex floats
    VString       // v5.42.0 and bare 5.42.0
    String        // '' "" q qq, with interpolation parts
    Backtick      // `` qx
    QwList        // qw()
    HeredocStart  // marker; body arrives as HeredocBody
    HeredocBody
    // Patterns
    Match         // m// //
    Subst         // s///
    Trans         // tr/// y///
    QuoteRegex    // qr//
    // Names
    Ident         // bareword, possibly package-qualified
    Variable      // $x @y %z &f *g, incl. punctuation vars
    Label         // FOO:
    // Structure
    LParen; RParen; LBracket; RBracket; LBrace; RBrace
    Semi; Comma; FatComma; Arrow; ColonAttr
    Operator      // everything else, text distinguishes
    // Sections
    DataMarker    // __END__ / __DATA__
    DataBody
    FormatBody
    // Terminal
    EOF
    Error         // recovery: malformed but lexing continues
    UnknownRest   // recovery: budget exhausted, rest is opaque
)
```

Every token carries `Start, End int` byte offsets. Do not store line/column —
derive them from a line-index built once per file, so that edits invalidate one
structure rather than every token.

---

## 2.2 The expectation state machine

### 2.2.1 `PL_expect`

`PL_expect` (`perl.h:5972-5985`) has **eleven** states. Chapter 3 §3.1 owns
them — the table, the transition macros, and the bracket stack that `}` pops.
Every subsection of this chapter that says "in term position" or "at a
statement boundary" means a value of that enum, and the lexer carries it
unchanged. The one state this chapter leans on by name is `XSTATE`, because
POD is recognized only there (§2.5.1).

### 2.2.2 The reduced model, and its cost

perl-lsp collapses the eleven states to five. Chapter 3 §3.9.4 has the
mapping, what it loses, and why the Go lexer implements all eleven.

### 2.2.3 The lexer state `PL_lex_state`

Orthogonal to `PL_expect`. Tracks whether we are inside an interpolating
construct (`toke.c:151-164`):

| State | Value | Meaning |
| --- | --- | --- |
| `LEX_NOTPARSING` | 11 | Not parsing (defined in `perl.h`) |
| `LEX_NORMAL` | 10 | Normal code, not inside `"..."` |
| `LEX_INTERPNORMAL` | 9 | Code within a string, e.g. `"$foo[$x+1]"` |
| `LEX_INTERPCASEMOD` | 8 | Expecting `\U`, `\Q`, `\E` etc. |
| `LEX_INTERPPUSH` | 7 | Starting a new sublex parse level |
| `LEX_INTERPSTART` | 6 | Expecting the start of a `$var` |
| `LEX_INTERPEND` | 5 | End of code, next char is not `[`, `{`, `->` |
| `LEX_INTERPENDMAYBE` | 4 | End of code, next char *is* one of `[`, `{`, `->` |
| `LEX_INTERPCONCAT` | 3 | Expecting anything; start of string or after `\E`, `$foo` |
| `LEX_INTERPCONST` | 2 | Not used |
| `LEX_FORMLINE` | 1 | Expecting a format line |

The values descend deliberately so the dispatch guard is a single comparison
(`toke.c:141-143`).

### 2.2.4 Interpolation is recursive lexing

`"foo $h{$k} bar"` is not scanned as one flat string. Perl re-enters the lexer
on the interpolated region, producing a nested token stream, and synthesizes
operators to join the pieces. `toke.c:2568-2576` states the model:

> `"foo\lbar"` is tokenised as
> `stringify ( const[foo] concat lcfirst ( const[bar] ) )`

The machinery is `S_sublex_start` (`toke.c:2595`), `S_sublex_push`
(`toke.c:2649`), `S_sublex_done` (`toke.c:2754`). `sublex_start` saves the
outer state into `PL_parser->lex_super_state` and sets
`PL_lex_state = LEX_INTERPPUSH` (`toke.c:2624-2630`); `sublex_push` opens a
scope returning a synthetic `(`; `sublex_done` restores and returns `)`. The
`SUBLEXSTART` / `SUBLEXEND` terminals (`perly.y:93`) mark the boundary.

**For Go:** model this with an explicit stack of lexer contexts, and make the
stack part of the checkpoint state (§2.14). Do not use host-language recursion
— an LSP must bound its depth on adversarial input.

---

## 2.3 Character set, encoding, line endings

### 2.3.1 Bytes versus characters

A Perl source file is a byte stream. Whether those bytes are interpreted as
Latin-1 or UTF-8 depends on `use utf8`, which is a *lexically scoped pragma
that takes effect partway through the file*. The lexer must therefore be able
to switch decoding mid-stream.

Throughout `toke.c` this appears as the `UTF` macro, threaded into nearly every
character classification (`isWORDCHAR_lazy_if_safe(s, PL_bufend, UTF)`,
`isIDFIRST_lazy_if_safe(...)`). The `_lazy_if_safe` naming means: treat as
UTF-8 only if the `UTF` flag is set, and never read past `PL_bufend`.

**For Go:** carry a `utf8 bool` on the lexer, flipped when `use utf8;` /
`no utf8;` is scanned. Note this is genuinely lexical — it must be saved and
restored across block scopes, which means the lexer needs a small scope stack
or must accept an approximation. Most real code puts `use utf8;` at file scope,
so a file-level flag is a reasonable v1 with a recorded limitation.

Go's `unicode/utf8` package covers decoding. `utf8.DecodeRune` returns
`RuneError` with size 1 for invalid bytes, which is the correct recovery
behavior — advance one byte and continue.

### 2.3.2 BOM

A UTF-8 BOM at offset 0 is consumed by `swallow_bom` (`toke.c:7691`). The call
is guarded by `bof` — beginning-of-file — computed from the file offset
(`toke.c:7683-7691`). Handle UTF-8 (`EF BB BF`), UTF-16LE, and UTF-16BE BOMs;
perl croaks on UTF-16 without a filter.

### 2.3.3 Line endings

Perl accepts `\n` and `\r\n`. `\r` is treated as whitespace in most positions.
Where it matters most is heredoc terminator matching and `__END__` detection.

Note `PERLIO_USING_CRLF` handling at `toke.c:7684-7688`: the byte offset used
for BOM detection is allowed to be off by one to account for a swallowed CR.

**Normalization is targeted, not global.** Ordinary code is not CRLF-normalized
— `\r` is simply whitespace. But three multi-line-literal scanners *do*
normalize, converting `\r\n`, `\n\r` and bare `\r` all to `\n`:

| Site | Citation |
| --- | --- |
| Heredoc pre-pass over the remaining buffer | `toke.c:11687-11708` |
| Heredoc per-streamed-chunk | `toke.c:11914-11926` |
| `scan_str` body loop | `toke.c:12522-12534` |

This is what makes heredoc terminator matching work on CRLF files — the
terminator comparison never sees a CR. A Go implementation that normalizes
globally will produce wrong spans; one that never normalizes will fail to match
heredoc terminators on Windows files. Normalize in exactly these three places.

> **Divergence (PerlOnJava).** `isAsciiWhitespace` (`Lexer.java:104-106`)
> includes `\r`, and `consumeWhitespace` (`Lexer.java:190-197`) deliberately
> *preserves* it in the token text. The comment (`Lexer.java:98-102`) explains
> why: eval'd code from Template Toolkit needs `\r` preserved inside string
> literals. This is a real constraint worth copying — do not normalize line
> endings destructively at the character layer.

A lone `\r` (classic Mac) is not a line terminator in modern perl.

### 2.3.4 The shebang line

If line 1 begins with `#!`, perl scans it for switches. The relevant handling
follows the `CopLINE(PL_curcop) == 1` check at `toke.c:7712`. Switches like
`-w`, `-T` in the shebang affect compilation. Notably, if the shebang does not
contain the string `perl`, perl may re-exec the named interpreter.

**For a parser/LSP:** recognize the shebang as a comment token, and optionally
extract `-w`/`-T`/`-C` as hints and `-M`, which can enable pragmas that change
the parse (`#!perl -Mstrict`). Do not implement re-exec.

### 2.3.5 `#line` directives

`S_incline` (`toke.c:1929`) parses line directives, resetting the reported file
and line number. The recognized forms:

```perl
#line 42
#line 42 "filename"
# line 42 "filename"
```

This matters for an LSP: a generated file may claim to be another file, and
diagnostics should be mapped accordingly. At minimum, record the directive
rather than silently treating it as an ordinary comment.

Implementation detail: perl stores **N − 1** (`toke.c:1986`), because
`COPLINE_INC_WITH_HERELINES` has already fired at `toke.c:1939`. The number in
the directive names the *next* line, not the current one.

### 2.3.6 VCS conflict markers

A pleasant surprise in modern perl: unresolved merge conflicts are detected and
diagnosed specifically rather than producing a cascade of syntax errors.
`toke.c:9783-9789` (for `<<<<<<<`) and `toke.c:9793-9799` (for `>>>>>>>`) check
for the marker at the start of a line and call `vcs_conflict_marker`.

The condition is: the character is `<` or `>`, `s[1]` is the same character,
the position is at `PL_linestart` or preceded by `\n`, and the following five
characters continue the run — i.e. seven total.

**Recommendation:** implement this. It converts a baffling error cascade into
one clear message, which is disproportionately valuable in an editor.

---

## 2.4 Whitespace and comments

### 2.4.1 Whitespace

Whitespace between tokens is `SPACE_OR_TAB` (defined as `isBLANK_A`,
`toke.c:134`) plus newlines and form feeds. The routine is `S_skipspace`, and
its lookahead-only sibling is `peekspace` — the distinction matters because
`skipspace` can read new lines from the input filehandle, while `peekspace`
must not (see §2.4.4).

Whitespace is generally insignificant, with these exceptions:

1. **Before a quote-like delimiter**, whitespace changes which delimiter is
   chosen — `q (x)` versus `q(x)` (§2.8.2).
2. **In `<<` heredocs**, whitespace between `<<` and a *bare* identifier is
   forbidden (§2.9.2).
3. **Inside `/x` regexes**, whitespace becomes insignificant *within the
   pattern*, which is a regex-engine concern, not a lexer one (§2.11.4).
4. **In formats**, whitespace is part of the picture line (§2.13).
5. **Between the two parts of a bracketing `s{}{}`**, whitespace *and
   comments* are permitted (§2.8.5) — this one surprises people.

### 2.4.2 Comments

`#` begins a comment that runs to the end of the line. The newline itself is
not part of the comment.

**`#` is not a comment in these nine contexts:**

| Context | Notes | Citation |
| --- | --- | --- |
| Inside a string, regex, or quote-like body | Ordinary content | — |
| The *delimiter* of a quote-like operator | `q#foo#` — only when **adjacent** (§2.10.3) | `toke.c:9441-9442` |
| `$#` last-index sigil | `$#array`, `$#{...}`, `$#$ref` | — |
| `$#` the deprecated format variable | Fatal on use since 5.30 | `gv.c:2388` |
| `(?#...)` regex comment group | Skipped wholesale | `toke.c:3714-3732` |
| `#` in an `/x` pattern | A *regex* comment — copied through, not dropped | `toke.c:3743-3751` |
| Quoted heredoc terminators | `<<"E#ND"` keeps the `#` | `toke.c:11646-11653` |
| Format picture lines | Part of the picture | `toke.c:13297` |
| Inside `qw()` | A literal word; warns under `WARN_QW` | `toke.c:6163-6167` |

One further case is an **error**, not a comment: `#` immediately after a sigil
in a subroutine signature (`toke.c:5555-5559`). The source comment states the
rule directly — *`'$#' is banned, while '$ # comment' isn't`*.

The `$#` cases are worth stating explicitly because they are a common lexer
bug:

```perl
my $last = $#array;      # $# is the last-index sigil, not a comment
my $n    = $#{$aref};    # likewise
print "count: $#foo\n";  # inside a string, still the sigil
```

PerlOnJava handles this by making `$#` a two-character operator token
(`Lexer.java:225-228`).

### 2.4.3 Trivia retention

Perl discards comments. An LSP must not. Emit `Comment` and `Whitespace`
tokens with accurate spans and let the parser skip them via a filtered view.
perl-lsp takes this approach — `is_trivia()` (`token.rs:131`) classifies
`Whitespace | Newline | Comment` and the parser skips them on demand.

### 2.4.4 `skipspace` versus `peekspace` — a subtle trap

`skipspace` may consume input *and read further lines*. That makes it unsafe
in lookahead, because it can trigger heredoc body collection and line-number
updates as a side effect.

`toke.c:9441-9455` shows the careful dance required to look ahead for `=>`
across a line boundary without corrupting state: it saves `PL_bufptr` and `s`
as *offsets* (`bufoff`, `soff`) rather than pointers, calls `peekspace`, tests
for `=>`, then restores both from the offsets. The offsets are necessary
because the buffer may be reallocated.

**For Go:** this is easier — use integer indices throughout rather than
pointers, and the whole class of bug disappears. But keep the distinction
between a lookahead that may not have side effects and a consuming skip that
may.

---

## 2.5 POD, `__END__`, `__DATA__`

### 2.5.1 POD — the exact rule

POD is a documentation format embedded in source. A POD block starts with a
line beginning `=` followed by an identifier character, and continues until a
line beginning `=cut` (or end of file).

The precise recognition condition is at `toke.c:9728-9730`:

```c
if (PL_expect == XSTATE
    && isALPHA(tmp)
    && (s == PL_linestart+1 || s[-2] == '\n') )
```

Three conjuncts, all required:

1. **`PL_expect == XSTATE`** — the lexer must be at a statement boundary.
2. **`isALPHA(tmp)`** — the character after `=` must be an ASCII letter.
   So `=cut`, `=pod`, `=head1` qualify; `==`, `=>`, `=~`, `=1` do not.
3. **`s == PL_linestart+1 || s[-2] == '\n'`** — the `=` must be at column 0.

The terminator test is `memBEGINPs(s, ..., "=cut") && !isIDCONT_A(s[4])`
(`toke.c:7696-7697` and `toke.c:9739-9740`) — `=cut` must not be followed by an
identifier character, so `=cutting` does not end a POD block, but `=cut foo`
does.

**A stray `=cut` with no POD open starts a POD block.** `=cut` satisfies all
three conjuncts above — `c` is alpha, and if it is at column 0 in statement
position, POD begins. There is no "are we already in POD" precondition on the
*opening* test.

**[verified]**:

```perl
print "a\n";
=cut
print "b\n";
```
→ prints only `a`. Everything from `=cut` to EOF was swallowed as POD.

This is a realistic editing hazard — deleting a `=pod` line and leaving its
`=cut` silently deletes the rest of the file from the compiler's view. An LSP
should diagnose an unmatched `=cut`.

### 2.5.2 Two POD paths

There are two implementations of POD skipping, and they differ:

- **File path** (`toke.c:7694-7707`): sets `PL_parser->in_pod = 1` and the
  line-reading loop discards lines until `=cut`. Reached at `toke.c:9754-9755`.
- **String/eval path** (`toke.c:9732-9752`): when `PL_in_eval && !PL_rsfp` or
  `PL_lex_state != LEX_NORMAL`, the whole buffer is already in memory, so it
  scans forward for a line starting `=cut` and jumps past it.

Both end with `goto retry`, resuming normal lexing.

### 2.5.3 Can POD appear mid-expression?

**No — and this is worth stating carefully, because it is commonly assumed
otherwise.**

The `PL_expect == XSTATE` guard means POD is recognized *only* at a statement
boundary. Verified against perl 5.42.0:

```perl
# WORKS — POD at a statement boundary
print "a\n";
=pod
docs
=cut
print "b\n";
```
**[verified]** — prints `a` then `b`.

```perl
# FAILS — POD in the middle of an expression
my $x = 1 +
=pod
this is pod
=cut
2;
```
**[verified]** — `syntax error at -e line 2, near "="`.

```perl
# FAILS — POD inside a list, mid-expression
my @x = (1,
=pod
p
=cut
2);
```
**[verified]** — `syntax error`.

```perl
# FAILS — indented, so not column 0
print "a\n";
  =pod
  =cut
print "b\n";
```
**[verified]** — `syntax error`.

So the correct statement is: **POD may appear between statements anywhere in a
file, including between statements inside a block or a sub body — but never
inside an expression, and never indented.** The "mid-expression" claim is a
myth; what is true and surprising is that POD may interrupt a *file* at any
statement boundary, so a lexer cannot simply strip POD in a prepass keyed on
blank-line-then-`=`, nor assume POD only appears after the last statement.

An additional subtlety: because the guard is `PL_expect == XSTATE`, whether a
given `=` at column 0 is POD depends on parser state. A lexer that does not
track expectation *cannot* decide this correctly. This is a concrete case where
perl-lsp's five-mode model (chapter 3 §3.9.4) is structurally unable to match perl.

### 2.5.4 `__END__` and `__DATA__`

Both terminate compilation of the current file. The remainder becomes readable
through a filehandle.

**They are ordinary keywords, not line-anchored markers.** Unlike POD, there is
no column-0 requirement and no preceding-newline requirement — they are matched
by the keyword lookup like any other word.

**[verified]** — indented, mid-line, and still effective:

```perl
print "x\n";  __DATA__
junk
```
→ prints `x`; everything after is data.

| Marker | Filehandle | Notes |
| --- | --- | --- |
| `__END__` | always `main::DATA` | Traditionally in the main script. |
| `__DATA__` | `<current package>::DATA` | Intended for modules; package-scoped. |

**In a `require`d file the two differ.** The guard at `toke.c:8297` means that
when `PL_in_eval` is true *and* `PL_rsfp` is a real file — exactly the `require`
case — `__DATA__` creates the handle but `__END__` does not. This asymmetry is
invisible from main-script testing.

Both are recognized as keywords `KEY___END__` and `KEY___DATA__`, and are
explicitly excluded from the `=>` lookahead at `toke.c:9441`:

```c
if (key && key != KEY___DATA__ && key != KEY___END__
 && (!anydelim || *s != '#')) {
```

That exclusion exists because looking ahead past `__END__` would read into the
data section.

A literal `^D` (chr 4) or `^Z` (chr 26) is also treated as EOF, through the
same `yyl_fake_eof` (`toke.c:9504-9508`; PerlOnJava `DataSection.java:217`).

**For a parser/LSP:** emit `DataMarker` then a single `DataBody` token spanning
to EOF. Do not attempt to lex the body. perl-lsp models exactly this with
`DataMarker` / `DataBody` (`token.rs:74-77`) and an `InDataSection` mode
(`mode.rs:61`).

Note that text after `__END__` is *not* necessarily data — by convention it
often contains POD. An editor may wish to offer POD highlighting there, but the
lexer proper should treat it as opaque.

---

## 2.6 Identifiers

### 2.6.1 Basic form

An identifier begins with a letter or underscore and continues with letters,
digits, or underscores. In `toke.c` terms: `isIDFIRST_lazy_if_safe` to start,
`isWORDCHAR_lazy_if_safe` to continue, both parameterized by the `UTF` flag.

```
ident_start ::= [A-Za-z_]                       (* not utf8 *)
              | [A-Za-z_] | XID_Start           (* under use utf8 *)
ident_cont  ::= [A-Za-z0-9_]                    (* not utf8 *)
              | [A-Za-z0-9_] | XID_Continue     (* under use utf8 *)
identifier  ::= ident_start ident_cont*
```

### 2.6.2 Unicode identifiers

Under `use utf8`, perl accepts identifiers built from Unicode identifier
characters. PerlOnJava models this as ICU's `XID_START` / `XID_CONTINUE`, plus
underscore (`Lexer.java:72-78`):

```java
private static boolean isPerlIdentifierStart(int codePoint) {
    return codePoint == '_' || UCharacter.hasBinaryProperty(codePoint, UProperty.XID_START);
}
```

This is the correct model.

> **Divergence (perl-lsp) — confirmed incorrect.** `unicode.rs:48-68` accepts
> **emoji** as identifier start characters, via an explicit `is_emoji_codepoint`
> table covering U+1F300–U+1F6FF, U+2600–U+26FF and more (`unicode.rs:28-45`).
> `is_perl_identifier_continue` (`unicode.rs:70-84`) additionally accepts ZWJ
> (U+200D), ZWNJ (U+200C) and the combining enclosing keycap (U+20E3).
>
> Real perl rejects this. **[verified]** against perl 5.42.0:
> ```perl
> use utf8; my $🚀 = 1;
> ```
> → `Unrecognized character \x{1f680}; marked by <-- HERE after my $<-- HERE`
>
> A Go implementation should use `XID_Start`/`XID_Continue` semantics and
> **not** copy perl-lsp's emoji extension. Go's standard library exposes
> `unicode.In(r, unicode.L, unicode.Nl, unicode.Other_ID_Start)` and friends;
> `golang.org/x/text` is not required for a close approximation, though the
> stdlib's tables are XID-adjacent rather than exactly XID.

### 2.6.3 Package separators

Two forms:

- **`::`** — the modern separator. `Foo::Bar::baz`.
- **`'`** — the archaic separator, inherited from Perl 4. `Foo'bar` means
  `Foo::bar`.

The apostrophe form is **not deprecated and not removed — it is version-gated.**
It is a default-on named feature, `apostrophe_as_package_separator`, which the
`use v5.42` bundle switches **off**. Measured on 5.42.0: `use v5.38` and
`use v5.40` both still accept it; `use v5.42` does not.

It can also be switched off directly, independent of any version bundle:

```perl
no feature "apostrophe_as_package_separator";
$Foo::bar = 7;
print $Foo'bar;               # Can't find string terminator "'" before EOF
```

So a lexer cannot infer this mode from the version declaration alone — it must
track the feature bit, which `use VERSION` sets as one input among several. The gate is at `toke.c:10808-10811`, with
four call sites.

**[verified]** against perl 5.42.0 — the same code differs by one line:

```perl
# WITHOUT a version declaration -- apostrophe works
package Foo; sub bar {42} package main;
print &Foo'bar(), "\n";        # prints 42
```

```perl
# WITH use v5.42 -- apostrophe is a string delimiter again
use v5.42;
package Foo; sub bar {42} package main;
print &Foo'bar(), "\n";        # Can't find string terminator "'" before EOF
```

**Implementation consequence.** This is a lexer feature flag driven by a
`use VERSION` statement *earlier in the same file*. A lexer must track the
declared version and change identifier scanning accordingly — which means
`use v5.42;` has a lexical effect on everything after it, in the same way
`use utf8` does (§2.3.1). Both belong in the checkpoint state (§2.14.1).

A useful consequence of the lookahead rule: `'` is only a separator when
followed by an identifier-start character. That is why `$'` (the postmatch
variable, §2.7.5) continues to work.

This is a genuine lexing hazard, because `'` is also the single-quote string
delimiter. The disambiguation is positional: after an identifier character
within a name being scanned, `'` is a separator; otherwise it opens a string.
`S_scan_inputsymbol` explicitly permits it in filehandle names
(`toke.c:12117-12118`):

```c
/* allow <Pkg'VALUE> or <Pkg::VALUE> */
d = parse_ident_no_copy(d, e, cBOOL(UTF), NULL, ALLOW_PACKAGE);
```

The `ALLOW_PACKAGE` flag is what enables both separators.

Note the interaction with the fat comma: `$hash{Foo'bar}` and `Foo'bar => 1`
require care. And the pathological case `print $x 'string'` — where `'` must be
a string opener because `$x` is complete.

> **Note (perl-lsp).** `unicode.rs:76` accepts `'` unconditionally as an
> identifier-continue character. That over-accepts: it makes `don't` in a
> context expecting a bareword scan as one identifier. Perl's own handling is
> conditional on `ALLOW_PACKAGE` being requested by the caller.

### 2.6.4 Package-qualified name grammar

```
pkg_sep    ::= "::" | "'"
qualified  ::= pkg_sep? identifier ( pkg_sep identifier )* pkg_sep?
```

A leading `::` means "main": `::foo` is `main::foo`. A *trailing* `::` is legal
and denotes the package itself (`Foo::` as a stash name). Trailing `'` is not.

Repeated separators collapse: `Foo::::bar` is accepted and means `Foo::bar`.

---

## 2.7 Variables and the sigil layer

### 2.7.1 Sigils

| Sigil | Denotes | Example |
| --- | --- | --- |
| `$` | Scalar (or one element of an aggregate) | `$x`, `$a[0]`, `$h{k}` |
| `@` | Array (or a slice) | `@a`, `@a[1,2]`, `@h{'x','y'}` |
| `%` | Hash (or a kv-slice) | `%h`, `%a[1,2]`, `%h{'x'}` |
| `&` | Subroutine | `&foo`, `&$ref` |
| `*` | Glob | `*STDOUT` |
| `$#` | Last index of an array | `$#a`, `$#{$r}`, `$#$r` |

`$#` is the `DOLSHARP` terminal (`perly.y:91`).

Each of `$ @ % & *` is ambiguous with an operator (`%` modulo, `&` bitwise-and,
`*` multiply, `@` nothing, `$` nothing). Resolution is by `PL_expect`. The
warning at `toke.c:7936-7941` shows perl's own uncertainty:

```
"Operator or semicolon missing before %c%s"
"Ambiguous use of %c resolved as operator %c"
```

Emitted from `yyl_safe_bareword` (`toke.c:7930`) when the previous character
was `*`, `%` or `&` and `PL_parser->saw_infix_sigil` is set.

### 2.7.2 Braced forms

```perl
${name}          # same as $name — disambiguates from following text
${ $expr }       # symbolic/hard deref of $expr
${^GLOBAL_PHASE} # a caret-name variable (see 2.7.4)
@{[ ... ]}       # the "baby cart" idiom — interpolate an expression
${\ ($x) }       # scalar-ref interpolation idiom
```

Inside a string, `${name}` is essential to bound the name:

```perl
print "${foo}bar";   # $foo followed by literal "bar"
print "$foobar";     # the variable $foobar
```

`S_scan_ident` (`toke.c:10955`) implements this. It emits the disambiguation
warnings seen at `toke.c:11172-11174` and `toke.c:11247`:

```
"Ambiguous use of %c{%s[...]} resolved to %c%s[...]"
"Ambiguous use of %c{%s} resolved to %c%s"
```

### 2.7.3 `$#` forms

```perl
$#array      # last index of @array
$#{$aref}    # last index of @$aref
$#$aref      # same, no braces
$#{ $h{k} }  # arbitrary expression
```

### 2.7.4 Caret variables

Control-character variables have two spellings:

- **Literal control character**: `$^W` may be written with an actual `Ctrl-W`
  byte (0x17), or as caret-`W`.
- **Braced word form**: `${^WORD}` where `WORD` is uppercase letters and
  underscores. `${^GLOBAL_PHASE}`, `${^TAINT}`, `${^UNICODE}`, `${^CAPTURE}`.

The braced form is the only way to spell multi-character control names. Inside
`${...}`, a leading `^` introduces a caret name and the following characters
are read as a name rather than an expression.

**Caret variables are a lexer encoding, not a separate namespace.** `${^FOO}`
is rewritten to the ordinary symbol whose name begins with the control
character — `^F` becomes byte 0x06, so `${^FOO}` is the symbol `"\006OO"`
(`toke.c:11076-11079`).

**[verified]** — reaching `${^GLOBAL_PHASE}` through its raw symbol name:

```perl
print ${"\007LOBAL_PHASE"}, "\n";   # prints RUN
```

`^G` is 0x07, so `"\007LOBAL_PHASE"` names the same variable. A lexer should
perform this transformation and emit the encoded name, so downstream symbol
resolution sees one namespace rather than two.

### 2.7.5 The full punctuation-variable table

Sourced from `perlvar.pod`. This is a deliverable of this chapter — a lexer
must recognize all of these as *variables*, not as operator sequences.

**Regex-related:**

| Variable | Meaning |
| --- | --- |
| `$1`, `$2`, … `$N` | Capture group N from the last successful match |
| `` $` `` | `$PREMATCH` — text before the last match |
| `$&` | `$MATCH` — text matched |
| `$'` | `$POSTMATCH` — text after the last match |
| `$+` | `$LAST_PAREN_MATCH` — last capture group that matched |
| `$^N` | Text matched by the most recently closed group |
| `@+` | `@LAST_MATCH_END` — end offsets of captures |
| `@-` | `@LAST_MATCH_START` — start offsets of captures |
| `%+` | `%LAST_PAREN_MATCH` — named captures (last matched) |
| `%-` | `%LAST_MATCH_START` — named captures (all, as arrayrefs) |
| `${^CAPTURE}` | Alias for numbered captures as an array |
| `${^CAPTURE_ALL}` | All captures |
| `${^MATCH}` | `$&` under `/p` |
| `${^PREMATCH}` | `` $` `` under `/p` |
| `${^POSTMATCH}` | `$'` under `/p` |
| `${^RE_COMPILE_RECURSION_LIMIT}` | Regex compile recursion limit |
| `${^RE_DEBUG_FLAGS}` | Regex debugging flags |
| `${^RE_TRIE_MAXBUF}` | Regex trie optimization buffer |

**I/O:**

| Variable | Meaning |
| --- | --- |
| `$_` | `$ARG` — the default input and pattern-match variable |
| `@_` | Subroutine arguments |
| `$.` | `$NR` — input line number of the last filehandle read |
| `$/` | `$RS` — input record separator (`undef` = slurp) |
| `$\` | `$ORS` — output record separator |
| `$,` | `$OFS` — output field separator |
| `$"` | `$LIST_SEPARATOR` — separator for interpolated arrays |
| `$;` | `$SUBSEP` — subscript separator for multidim emulation |
| `$\|` | `$OUTPUT_AUTOFLUSH` — autoflush the selected filehandle |
| `$^A` | `$ACCUMULATOR` — format line accumulator |
| `$^L` | `$FORMAT_FORMFEED` |
| `$^` | `$FORMAT_TOP_NAME` |
| `$~` | `$FORMAT_NAME` |
| `$%` | `$FORMAT_PAGE_NUMBER` |
| `$=` | `$FORMAT_LINES_PER_PAGE` |
| `$-` | `$FORMAT_LINES_LEFT` |
| `$:` | `$FORMAT_LINE_BREAK_CHARACTERS` |
| `$#` | Deprecated output number format |
| `$*` | Removed (was multiline matching) |
| `ARGV` | The magic input filehandle |
| `$ARGV` | Name of the current file when reading `<>` |
| `@ARGV` | Command-line arguments |
| `@F` | Split fields under `-a` |
| `${^LAST_FH}` | Last read filehandle |

**Error and status:**

| Variable | Meaning |
| --- | --- |
| `$@` | `$EVAL_ERROR` — error from the last `eval` |
| `$!` | `$OS_ERROR` / `$ERRNO` — errno; numeric or string per context |
| `$^E` | `$EXTENDED_OS_ERROR` — platform-specific error |
| `$?` | `$CHILD_ERROR` — status of the last pipe close, backtick, or `wait` |
| `%!` | Per-errno hash (`$!{ENOENT}`) |
| `${^CHILD_ERROR_NATIVE}` | Native child status |

**Process and interpreter:**

| Variable | Meaning |
| --- | --- |
| `$0` | `$PROGRAM_NAME` — name of the running program |
| `$$` | `$PID` — process ID |
| `$<` | `$UID` — real user ID |
| `$>` | `$EUID` — effective user ID |
| `$(` | `$GID` — real group ID |
| `$)` | `$EGID` — effective group ID |
| `$^C` | `$COMPILING` — true under `-c` |
| `$^D` | `$DEBUGGING` — debug flags |
| `$^F` | `$SYSTEM_FD_MAX` |
| `$^H` | Compile-time hints (internal) |
| `%^H` | Lexical hints hash |
| `$^I` | `$INPLACE_EDIT` — in-place edit extension |
| `$^M` | Emergency memory pool |
| `$^O` | `$OSNAME` — operating system name |
| `$^P` | `$PERLDB` — debugger flags |
| `$^R` | `$LAST_REGEXP_CODE_RESULT` |
| `$^S` | `$EXCEPTIONS_BEING_CAUGHT` |
| `$^T` | `$BASETIME` — script start time |
| `$^V` | `$PERL_VERSION` — a version object |
| `$^W` | `$WARNING` — global warning flag |
| `${^WARNING_BITS}` | Lexical warnings bitmask |
| `$^X` | `$EXECUTABLE_NAME` — the perl binary path |
| `$]` | Perl version as a number |
| `$[` | Removed (was array base index) |
| `${^GLOBAL_PHASE}` | `CONSTRUCT`/`START`/`CHECK`/`INIT`/`RUN`/`END`/`DESTRUCT` |
| `${^TAINT}` | Taint mode status |
| `${^UNICODE}` | Unicode settings from `-C` |
| `${^UTF8CACHE}` | UTF-8 cache debugging |
| `${^UTF8LOCALE}` | UTF-8 locale detected |
| `${^SAFE_LOCALES}` | Locale safety |
| `${^HOOK}` | Hook registry (5.40+) |
| `@INC` | Module search path |
| `%INC` | Loaded modules |
| `%ENV` | Environment |
| `%SIG` | Signal handlers |
| `${^OPEN}` | PerlIO layers from `open` pragma |
| `${^ENCODING}` | Removed |
| `${^WIN32_SLOPPY_STAT}` | Win32 stat behavior |

**Implementation note.** After `$`, the next character determines everything.
Build a table keyed on that character rather than a chain of comparisons. The
hard cases are `$$` (PID versus a scalar deref `$$ref`), `$#` (last-index
versus the deprecated format variable), and `$:` (format break chars versus a
`::` package prefix). All three resolve by peeking one further character:

```perl
$$          # PID  -- next char is not an identifier start or {
$$ref       # deref of $ref
$#          # deprecated format var
$#array     # last index
$#{$r}      # last index of deref
$::x        # $main::x
```

---

## 2.8 Numeric literals

### 2.8.1 The authoritative grammar

`toke.c:12586-12603` carries perl's own grammar in a comment. Reproduced with
its structure preserved:

```
(* integers *)
binary   ::= "0" [bB]      [0-1] ( "_"* [0-1] )*
octal    ::= "0" [oO]?     [0-7] ( "_"* [0-7] )*
decimal  ::=               [0-9] ( "_"* [0-9] )*
hex      ::= "0" [xX]      hexdig ( "_"* hexdig )*

(* floats *)
decfloat ::= [0-9] ("_"* [0-9])* ( "." ( [0-9] ("_"* [0-9])* )? )?
                                 ( [eE] [-+]? [0-9] ("_"* [0-9])* )?
           | "." [0-9] ("_"* [0-9])*
                                 ( [eE] [-+]? [0-9] ("_"* [0-9])* )?
hexfloat ::= "0" [xX] hexdig ("_"* hexdig)*
                     ( "." ( hexdig ("_"* hexdig)* )? )?
                     [pP] [-+]? [0-9] ("_"* [0-9])*     (* REQUIRED *)
```

Two constraints stated at `toke.c:12599-12601`:

- **Hex floats require an exponent**; the fractional part is optional.
- **Decimal floats require the fractional part or the exponent** (or both).

### 2.8.2 Radix dispatch

At `toke.c:12736-12755`:

| Test | Result |
| --- | --- |
| `s[1]` is `x`/`X` | `shift = 4` (hex), `s += 2` |
| `s[1]` is `b`/`B` | `shift = 1` (binary), `s += 2` |
| `s[1]` is `.` or `e`/`E` | `goto decimal` — this is why `0.5` and `0e0` are floats |
| otherwise | `shift = 3` (octal), `s++`; then consume an optional `o`/`O` |

Because the leading `0` is consumed before the `o` test, **`017` and `0o17`
share one code path** (`toke.c:12748-12755`). **[verified]** both yield `15`.

### 2.8.3 Radix digit violations

| Condition | Behavior | Citation |
| --- | --- | --- |
| `8` or `9` in octal | `yyerror("Illegal octal digit '%c'")` — **fatal** | `toke.c:12782-12784` |
| `2`–`7` in binary | `yyerror("Illegal binary digit '%c'")` — **fatal** | `toke.c:12788-12789` |
| `a`–`f` when not hex | Silently stop scanning | `toke.c:12798-12800` |
| No digits after prefix | `yyerror("No digits found for %s literal")` | `toke.c:13015-13028` |

**[verified]** `print 09` → `Illegal octal digit '9' at -e line 1`. Note this
is a compile-time *error*, not a warning — a lexer for an LSP should report it
as a diagnostic and recover by consuming the digit run.

Because `a`–`f` merely stops the scan, `0b1a` lexes as the number `0b1`
followed by the bareword `a`. That is not an error at the lexical layer.

### 2.8.4 Underscores — the rule that is not a grammar rule

**Underscores are never a syntax error anywhere in a number.** Every
misplacement produces at most a `syntax`-category *warning*.

The machinery is at `toke.c:12630-12668`:

- `WARN_ABOUT_UNDERSCORE` (`toke.c:12630`) emits
  `ck_warner(packWARN(WARN_SYNTAX), "Misplaced _ in number")`.
- It is latched by `warned_about_underscore` (`toke.c:12625`), so **at most one
  such warning is emitted per literal**.
- `SUFFER_AN_UNDERSCORE_HERE` (`toke.c:12643`) tolerates an underscore where
  one is not expected, warns, and absorbs all adjacent underscores.
- `SUFFER_AN_UNDERSCORE_JUST_BEFORE_HERE` (`toke.c:12657`) catches trailing
  underscores.

Positions that warn: leading (right after a radix prefix — `toke.c:12759`),
after `.` (`toke.c:13112`), after `e` (`toke.c:13160`), after an exponent sign
(`toke.c:13166`), trailing (`toke.c:12851`, `13102`, `13130`, `13181`), and
doubled (`toke.c:12668`).

**[verified]**: `1_0` → `10` silently; `1__0` → `Misplaced _ in number` then
`10`; `1_` → warning then `1`.

**Underscores are permitted in exponents** — `1e1_0` → `10000000000`
(`toke.c:13173-13180` uses `isDIGIT_or_UNDERSCORE`).

> **Specification consequence.** Any EBNF that *forbids* `1__0` is stricter
> than perl. A Go lexer should accept underscores anywhere within a digit run
> and emit a warning diagnostic, not a parse error. Both PerlOnJava
> (`NumberParser.java:502-504`, a blanket `replaceAll("_","")`) and perl-lsp
> accept them silently — they diverge from perl only on the diagnostic, not on
> the value.

### 2.8.5 Decimal and float details

- Integer digits: `toke.c:13083-13099`.
- Fractional part: `toke.c:13107`, guarded by `*s == '.' && s[1] != '.'`. **The
  `s[1] != '.'` test is what makes `3..5` a range rather than `3.` followed by
  `.5`.** **[verified]** `join(",",3..5)` → `3,4,5`.
- A **trailing `.` is legal**: `5.` → `5`. **[verified]**
- A **leading `.` is legal**: `.5` → `0.5`, reached via the `*s == '.'`
  disjunct at `toke.c:13047`.
- Exponent: `toke.c:13140-13141`. Guard requires `[eE]` (or `[pP]` for hex
  floats) **and** `memCHRs("+-0123456789_", s[1])`. `E` is folded to lowercase
  when copied (`toke.c:13151`) with the comment that some `atof()`
  implementations do not accept `E`.
- **`10E10` is valid.** **[verified]** → `100000000000`.

  > **Divergence (PerlOnJava).** Its own class comment (`Lexer.java:39`) admits
  > `10E10` is mis-lexed as `10` then `E10` at the lexer layer, and repaired in
  > `NumberParser`. A single-pass Go lexer should just handle it correctly.

- **Exponent backtracking** (`toke.c:13183-13193`): if zero exponent digits
  were consumed, restore `s` and `d`. So `1e+` lexes as the number `1` followed
  by the bareword `e`. perl-lsp reproduces this correctly (`lib.rs:1070-1072`).

### 2.8.6 Hex, binary, and octal floats

Detection is `HEXFP_PEEK` (`toke.c:136-139`):

```c
#define HEXFP_PEEK(s)     \
    (((s[0] == '.') && \
      (isXDIGIT(s[1]) || isALPHA_FOLD_EQ(s[1], 'p'))) || \
     isALPHA_FOLD_EQ(s[0], 'p'))
```

It is consulted inside the generic radix digit loop (`toke.c:12841`) and again
at the `out:` label (`toke.c:12857`).

**Important correction.** Perl's own doc comment (`toke.c:12591`) lists only
decimal and hex floats. But because `HEXFP_PEEK` is tested in the *generic*
radix loop regardless of the `shift` value, **binary and octal floats also
work** in perl 5.42.0. **[verified]**:

```perl
printf "%.4f\n", 0x1.8p3;   # 12.0000
printf "%.4f\n", 0b1.1p2;   #  6.0000
printf "%.4f\n", 017.4p1;   # 31.0000
printf "%.4f\n", 0o17.4p1;  # 31.0000
```

All four succeeded. A `p` exponent is a *binary* exponent in every radix.

> This contradicts the common belief (and PerlOnJava's framing) that binary and
> octal floats are a non-perl extension. They are reachable in perl 5.42.0.
> Treat them as supported; they are rare enough that a lexer may reasonably
> emit a "rarely used" hint, but must not reject them.

Hex-float warnings, all `WARN_OVERFLOW`:

| Message | Citation |
| --- | --- |
| `Hexadecimal float: exponent underflow` | `toke.c:12986` |
| `Hexadecimal float: exponent overflow` | `toke.c:12996` |
| `Hexadecimal float: mantissa overflow` | `toke.c:13217` |

### 2.8.7 Overflow and IV/NV promotion

Radix-loop overflow (`toke.c:12814-12822`): when `(x >> shift) != u`, set
`overflowed`, switch to NV accumulation, and warn
`Integer overflow in %s number` where `%s` comes from
`bases[] = {"", "binary", "", "octal", "hexadecimal"}` (`toke.c:12707-12708`).

Decimal promotion (`toke.c:13205-13218`): call `grok_number`; produce a UV/IV
only if the result is exactly `IS_NUMBER_IN_UV`, otherwise fall through to
`floatit = TRUE` and `newSVnv(Atof(...))`.

There is an `assert(!(flags & IS_NUMBER_NEG))` at `toke.c:13210` — **a numeric
literal is never negative.** Unary minus is a separate operator. A Go lexer
must not fold a leading `-` into the number token.

Buffer limit: overflow of `PL_tokenbuf` → `croak("Number too long")`
(`toke.c:13093`, `13122`, `13176`).

### 2.8.8 Version strings

Two forms.

**Explicit `v` prefix.** Routed at `toke.c:9880-9915`. Requires
`isDIGIT(s[1]) && PL_expect != XOPERATOR` (`toke.c:9880`). Taken as a v-string
if a `.` followed by a digit appears (`toke.c:9885`); otherwise there are
fallbacks for `::` labels (`9889`) and `XSTATE` labels (`9894`). Finally at
`toke.c:9906-9915`, a dotless `vNN` in term position is a v-string **only if no
sub of that name is visible**. The source comment: *"avoid v123abc() or
$h{v1}, allow `print v10;`"*.

**[verified]** `my $x = v65; print $x` → `A`. But `print v65, "\n"` →
`No comma allowed after filehandle` — because in that position `v65` is taken
as a filehandle name. Term position matters.

**Bare `N.N.N`.** At `toke.c:13134-13138`, inside the decimal branch:

```c
if (*s == '.' && isDIGIT(s[1])) {
    /* oops, it's really a v-string, but without the "v" */
    s = start;
    goto vstring;
}
```

**The rule: a bare number becomes a v-string when a SECOND `.` is immediately
followed by a digit.** Two or more dots required.

**[verified]**:
- `5.42` → the float `5.42` (one dot)
- `5.42.0` → a v-string, `printf "%vd"` → `5.42.0`
- `1.2.3` → a v-string → `1.2.3`

**The `=>` escape hatch** (`toke.c:13984-13994`): in `Perl_scan_vstring`
(`toke.c:13974`), if the scan did not end on `.` and the next non-space text is
`=>`, return a plain **string** instead of a v-string. This makes
`( v65 => 1 )` produce the key `v65`. **[verified]** — `keys %h` → `v65`.

Guard at `toke.c:13997`: `if (!isALPHA(*pos))` — so `v5x` is not a v-string.

Components use `isDIGIT_or_UNDERSCORE` (`toke.c:13982`), so underscores are
permitted. Overflow warns `Integer overflow in decimal number`
(`toke.c:14019-14021`). The original text is attached as `PERL_MAGIC_vstring`
magic (`toke.c:14037`) — a v-string remembers how it was written.

### 2.8.9 Numeric divergence summary

| Feature | perl 5.42 | PerlOnJava | perl-lsp |
| --- | --- | --- | --- |
| `0o`/`0O` octal | yes (`toke.c:12751`) | yes (`NumberParser.java:187`) | yes (`lib.rs:953`) |
| bare `0NNN` octal | yes (`toke.c:12749`) | yes (`:193`) | **no** — lexes as decimal (`lib.rs:988`) |
| Illegal octal/binary digit | `yyerror`, fatal | different message, different severity | **absent** |
| Hex float `0x1.8p3` | yes **[verified]** | yes (`:354-361`) | **absent** |
| Binary/octal floats | **yes [verified]** | yes | no |
| `Misplaced _ in number` | warn, once per literal | **absent** | **absent** |
| Underscores in exponent | yes (`toke.c:13173`) | yes (`:425`) | yes (`lib.rs:1062`) |
| `10E10` | yes (`toke.c:13151`) | repaired in parser | yes (`lib.rs:1046`) |
| Exponent backtrack | yes (`toke.c:13188`) | throws instead | yes (`lib.rs:1070`) |
| trailing `.` (`5.`) | float **[verified]** | yes (`:199`) | allowlist-conditional |
| `3..5` guard | `s[1] != '.'` | relies on tokenizer | allowlist omission |
| bare `5.42.0` → vstring | yes **[verified]** | yes (`:207-213`) | **no** — real divergence |
| `v` + `=>` → plain string | yes **[verified]** | **no** | n/a |
| v-string underscores | yes (`toke.c:13982`) | yes | **no** |
| `Integer overflow` warning | yes (`toke.c:12820`) | **absent** | absent |
| IV/NV promotion | `grok_number` | `BigInteger.bitLength() > 64` | deferred |

perl-lsp's missing bare-`N.N.N` v-string is user-visible: `use 5.42.0;` lexes
wrongly (`lib.rs:406` runs `try_number` before `try_vstring`, and `try_number`
has no `goto vstring` equivalent).

### 2.8.10 The `should_consume_dot` approximation

perl-lsp replaces perl's one-character test `s[1] != '.'` with an **allowlist
of characters that may follow a consumed dot** (`lib.rs:1004-1033`): a digit,
EOF, a byte `<= b' '`, or one of `; , ) } ] + - * / % = < > ! & | ^ ~ e E`.

> **Divergence (perl-lsp).** This gets `3..5` right by omission (`.` is absent
> from the list) but changes behavior for `5.foo` and for any operator byte
> outside the hardcoded set. `e`/`E` are in the list specifically so `5.e3`
> works. Use perl's one-character lookahead instead — it is simpler *and*
> correct.

---

## 2.9 Heredocs

This is the single hardest construct in Perl's lexer and a known failure point
in every reimplementation. It is specified here in full.

### 2.9.1 Why heredocs are hard

A heredoc introducer appears in the middle of an expression, but its *body*
lives on subsequent lines. The lexer must:

1. Recognize `<<` as a heredoc rather than left-shift.
2. Read the terminator.
3. Emit a token representing the string **immediately**, so the expression
   continues to parse.
4. Defer collecting the body until the current **logical line** ends.
5. Handle several heredocs queued on one line, in order.
6. Handle the body being interpolating, non-interpolating, or command.
7. Keep line numbers correct despite the body being consumed out of order.

Point 4 is the crux: the body starts after the newline that ends the *current
line*, not after the token.

### 2.9.2 Recognizing `<<`

The decision is at `toke.c:7173-7181`, inside `yyl_leftpointy`:

```c
if (PL_expect != XOPERATOR) {
    if (s[1] != '<' && !memchr(s,'>', PL_bufend - s))
        check_unary();
    if (s[1] == '<' && s[2] != '>')
        s = scan_heredoc(s);
    else
        s = scan_inputsymbol(s);
    PL_expect = XOPERATOR;
    TOKEN(sublex_start());
}
```

Three conditions for a heredoc:

1. **`PL_expect != XOPERATOR`** — a term is expected. If an operator was
   expected, `<<` is left-shift (handled at `toke.c:7186-7189`).
2. **`s[1] == '<'`** — two `<` characters.
3. **`s[2] != '>'`** — excludes `<<>>`, the no-magic-open `ARGV` read, which
   goes to `scan_inputsymbol` instead.

**The rule is purely `PL_expect` — there is no lookahead and no
backtracking.** Perl does not inspect what follows `<<` to guess whether it
looks like a terminator.

```perl
print <<EOF;    # heredoc  -- term expected
$x = $y << 2;   # left shift -- operator expected after $y
while (<<>>) {} # ARGV read, not a heredoc
```

> **Divergence (perl-lsp) — a known false positive.** Lacking a real
> expectation state, it guesses (`lib.rs:757-759`):
> ```rust
> if self.mode == LexerMode::ExpectOperator && self.paren_depth > 0 { return None; }
> ```
> The `paren_depth > 0` condition is required because `print $fh <<END` is
> legal at depth 0. The consequence is that a **statement-level shift by a
> bareword constant** is mis-lexed as a heredoc:
> ```perl
> my $mask = 1 << WIDTH;   # perl-lsp: heredoc named WIDTH
> ```
> Perl gets this right unconditionally. This is the clearest single argument
> for implementing the full expectation enum (chapter 3 §3.1) rather than a
> previous-token heuristic.

### 2.9.3 Terminator forms

`S_scan_heredoc` is at `toke.c:11616`.

| Form | Interpolates | Notes |
| --- | --- | --- |
| `<<EOF` | yes | Bare identifier. **No space allowed** between `<<` and the word. |
| `<<"EOF"` | yes | Double-quoted; whitespace before the quote is permitted. |
| `<<'EOF'` | **no** | Single-quoted; literal body. |
| `` <<`EOF` `` | yes, then executed | Command heredoc. |
| `<<\EOF` | **no** | Backslash form — equivalent to `<<'EOF'`. **[verified]** |
| `<<~EOF` | per inner form | Indented; `~` may combine with any of the above. |

The mechanism behind the no-space rule is at `toke.c:11643-11646`: whitespace
after `<<` is skipped into a lookahead pointer `peek`, but `s` is only advanced
to `peek` **if a quote character follows**. For a bare word the scan therefore
resumes at the space, fails the `isWORDCHAR` test at `toke.c:11661`, and croaks.

**[verified]**:

```perl
print << EOF;      # Use of bare << to mean <<"" is forbidden
print << "EOF";    # works -- space before a quote is allowed
```

`~` must come **immediately** after `<<`, before any whitespace
(`toke.c:11640`). `<< ~EOF` is not an indented heredoc.

`<<~` may be combined: `<<~"EOF"`, `<<~'EOF'`, `` <<~`EOF` ``, `<<~\EOF`.

**Label characters.** The bare form consumes `isWORDCHAR` characters
(`toke.c:11661-11670`), which **includes a leading digit**. `<<1` is valid
Perl. **[verified]** — `print <<1;` with terminator `1` works.

> **Divergence (perl-lsp).** Bare labels are restricted to
> `is_perl_identifier_start` (`lib.rs:806`), so `<<1` is rejected and falls
> back to lexing `<<` as a shift operator.

Quoted labels are copied with `delimcpy` up to the closing quote and have **no
word-character restriction** — `<<"E#ND"` keeps the `#`. An unterminated quoted
label croaks `Unterminated delimiter for here document` (`toke.c:11650`). A
label longer than `PL_tokenbuf` croaks
`Delimiter for here document is too long` (`toke.c:11673-11674`).

Internally the label is stored as `"\n" LABEL "\n\0"` (`toke.c:11637`,
`11676-11678`). **That leading sentinel newline *is* the column-0 rule** — the
non-indented terminator search is a plain `memNE` against this buffer, so a
match can only occur at a line start, and because the pattern ends in `\n` the
label must occupy the entire line.

**Trailing whitespace after the terminator is not allowed.** **[verified]**:

```perl
print <<EOF;
b
EOF␣␣␣
```
→ `Can't find string terminator "EOF" anywhere before EOF`

> **Divergence (perl-lsp).** `lib.rs:283` applies
> `line.trim_end_matches([' ', '\t'])` with the comment
> `// Strip trailing spaces/tabs (Perl allows them)`. That comment is wrong —
> perl rejects them, as verified above.

**[verified]** the backslash form suppresses interpolation:

```perl
print <<\EOF;
no $interp here
EOF
```
→ prints `no $interp here` literally.

### 2.9.4 The deferral mechanism

Perl keeps the current line in the SV `PL_linestr`, with `PL_bufptr` the
current scan position and `PL_bufend` the end. `S_scan_heredoc` has **two**
capture mechanisms, selected at `toke.c:11720`:

**(A) In-buffer scan-and-excise** — used for string `eval`, or when inside a
quote-like operator (`toke.c:11720-11862`). The whole text is already in
memory. The body is located, copied out, and then **physically spliced out of
the buffer** (`toke.c:11854`):

```c
Move(s, d, bufend - s + 1, char);
SvCUR_set(linestr, SvCUR(linestr) - (s - d));
```

The remainder of the declaring line closes up over the hole. The function then
returns `s = olds` (`toke.c:11861`) — **the position just past the `<<LABEL`
token on the declaring line.**

**(B) Stream-stealing** — used at file scope (`toke.c:11863-11981`).
`PL_linestr` is swapped for a fresh SV (`toke.c:11875-11877`), lines are pulled
with `lex_next_chunk` and appended until the terminator matches, then the
original line buffer is restored and `s = d` — again the declaring line.

Both mechanisms defer identically: **`scan_heredoc` returns a pointer into the
declaring line just past the introducer, and the body has either been excised
or consumed.** The rest of the logical line then lexes normally, having never
seen the body. That is the mechanism.

`toke.c:10151` warns that `PL_parser->re_eval_start` must be adjusted when
`scan_heredoc` moves the buffer — **any pointer into the line buffer held
across a heredoc scan is invalidated.**

**Line accounting.** `PL_parser->herelines` counts body newlines the parser has
not yet charged to `CopLINE`. It is incremented per body line
(`toke.c:11782`, `11813`, `11824`, `11911`) and decremented once at
`toke.c:11802` because the indented scan overshoots onto the terminator line.
The debt is settled by `COPLINE_INC_WITH_HERELINES` (`toke.c:390-397`) at each
real end-of-line. `PL_multi_start = origline + 1 + herelines` (`toke.c:11720`)
is set *before* scanning, which is what makes a second heredoc on the same line
start after the first one's body.

`toke.c:10151` carries a note that `PL_parser->re_eval_start` must be adjusted
when `scan_heredoc` moves the buffer — a reminder that **any pointer into the
line buffer held across a heredoc scan is invalidated.**

**For Go — the recommended design.** Do not splice buffers. Use a **pending
queue**:

```go
type pendingHeredoc struct {
    terminator []byte
    interpolate bool
    indented    bool    // <<~
    command     bool    // <<`EOF`
    tokenIndex  int     // token to backpatch with the body
}
```

- On seeing `<<TERM`, emit a `HeredocStart` token, push a `pendingHeredoc`, and
  continue scanning the current line normally.
- On reaching the newline that ends the logical line, drain the queue **in
  FIFO order**, reading body lines for each in turn, emitting `HeredocBody`
  tokens and backpatching.
- Resume normal lexing after the last terminator line.

This preserves source offsets exactly — which perl's splicing approach does
not, and which an LSP requires.

### 2.9.5 Multiple heredocs on one line

Legal, and processed in the order the introducers appear.

**[verified]**:

```perl
print <<A, <<B;
first
A
second
B
```
→ prints `first` then `second`.

The bodies appear consecutively: all of A's body, A's terminator, then all of
B's body, B's terminator. A FIFO queue handles this naturally.

### 2.9.6 Heredocs inside expressions and as call arguments

The introducer may appear anywhere a term may.

**[verified]**:

```perl
my $x = join(",", <<X, "tail");
body
X
```
→ produces `body\n,tail`.

Note the body retains its trailing newline; the terminator line itself is not
part of the body.

This is why deferral must key on the **logical line**, not on the enclosing
expression — the expression continues past the heredoc, and may even span
multiple physical lines, but the body starts at the next physical newline.

A pathological but legal case: a heredoc introducer inside another heredoc's
interpolated body. The queue must be per-lexer-context, and saved/restored
across sublex push/pop (§2.2.4).

### 2.9.7 `<<~` indentation stripping

With `~`, the terminator may be indented, and the **common indentation is
stripped from every body line**.

Algorithm, in two phases.

**Phase 1 — find the terminator and record its indentation**
(`toke.c:11776-11803`). Walking back from the candidate terminator to the
previous newline, every character must be `SPACE_OR_TAB` — that is, a space or
a tab only, **not** `\f`, `\v` or `\r` (`toke.c:11786`). That exact byte run
becomes the indent prefix.

**Phase 2 — strip** (`toke.c:11984-12029`):

1. A **wholly empty** line (`*ss == '\n'` immediately, `toke.c:11996`) is
   exempt and emitted as a bare newline.
2. Otherwise the line must begin with the indent prefix, tested with
   `memEQ(ss, indent, indent_len)` (`toke.c:12002`) — a **literal byte-prefix
   comparison**. Excess indentation beyond the prefix is preserved.
3. Any other line is fatal (`toke.c:12016-12021`):
   `Indentation on line %d of here-doc doesn't match delimiter`

Two consequences worth stating explicitly:

- **A tab is not equivalent to any number of spaces.** A tab-indented
  terminator does not match an eight-space-indented body line.
- **A line of *partial* whitespace is NOT exempt** — only a completely empty
  line is. A line with two spaces where the terminator has four is fatal, not
  silently passed through. This is the classic gotcha.

The reported `%d` is 1-based **within the heredoc body**, not a file line
number (`toke.c:11987`, `11998`).

A `<<~EOF` whose terminator sits at column 0 yields a zero-length prefix, so
every `memEQ` of length 0 trivially succeeds — a legal no-op strip.

> **Divergence (PerlOnJava).** `ParseHeredoc.java:220` exempts any
> `line.trim().isEmpty()` line, so partial-whitespace lines that perl rejects
> are silently accepted. Its terminator match uses `String.stripLeading()`
> (`:164`), which strips **any Unicode whitespace**, broader than perl's
> `SPACE_OR_TAB`. Its error message also omits the line number.

> **Divergence (perl-lsp).** Indentation stripping is **not implemented at
> all** — `is_terminator` skips leading space/tab when matching
> (`lib.rs:286-302`), but no strip phase exists and the indent-mismatch error
> does not exist. Acceptable for a span-only lexer; it does not validate `<<~`.

**[verified]**:

```perl
print <<~EOT;
    indented
      more
    EOT
```
→ prints `indented\n  more\n` — the four-space common prefix removed, the
extra two spaces on the second line preserved.

**Tabs versus spaces are compared literally, not expanded.** A terminator
indented with a tab does not match a body line indented with eight spaces.
This is a frequent source of confusion and should be reported clearly.

### 2.9.8 Errors

| Condition | Message |
| --- | --- |
| Terminator never found before EOF | `Can't find string terminator "%s" anywhere before EOF` |
| Body indentation mismatch under `<<~` | `Indentation on line %d of here-doc doesn't match delimiter` |
| Terminator missing / delimiter unterminated | `Use of bare << to mean <<"" is forbidden` (for the removed bare-`<<` form) |

For an LSP, the unterminated case is the common one during editing: the user
has typed the introducer but not yet the terminator. Recovery should treat the
rest of the file as the body, emit a diagnostic, and **not** cascade. This is
exactly the situation perl-lsp's `UnknownRest` token exists for.

### 2.9.9 Heredoc divergences

> **Divergence (PerlOnJava).** There is no heredoc handling in the lexer at
> all — `Lexer.java` has no `<<` case and no deferral machinery. Heredocs are
> handled entirely at the parser layer, reconstructed from the coarse token
> stream. This follows from its "optimized for speed rather than accuracy"
> design (`Lexer.java:36-42`) but means the lexer alone cannot produce correct
> tokens for a file containing heredocs.

> **Divergence (perl-lsp).** `heredoc.rs` is 304 bytes — essentially a stub.
> The token types `HeredocStart` and `HeredocBody` exist (`token.rs:54-57`),
> but the deferral logic is minimal. The lexer carries no heredoc queue in its
> checkpoint state (`checkpoint_impl.rs:23-36` lists `position`, `mode`,
> `delimiter_stack`, `in_prototype`, `prototype_depth`, `after_sub`,
> `after_arrow`, `hash_brace_depth`, `after_var_subscript`, `paren_depth`,
> `current_pos`, `eof_emitted`, `context` — **no pending-heredoc list**).
> Consequently a checkpoint taken while heredocs are pending cannot be
> correctly restored. For an incremental LSP lexer this is a correctness gap,
> and the Go implementation must include the pending queue in its checkpoint.

---

## 2.10 String literals and quote-like operators

### 2.10.1 The operators

| Operator | Meaning | Interpolates | Parts |
| --- | --- | --- | --- |
| `'...'` | Literal string | no | 1 |
| `"..."` | Interpolating string | yes | 1 |
| `` `...` `` | Command execution | yes | 1 |
| `q//` | Literal string | no | 1 |
| `qq//` | Interpolating string | yes | 1 |
| `qw//` | Word list | no | 1 |
| `qx//` | Command execution | yes (unless `'` delimiter) | 1 |
| `qr//` | Compiled regex | yes | 1 + modifiers |
| `m//` | Match | yes | 1 + modifiers |
| `s///` | Substitution | yes | **2** + modifiers |
| `tr///` | Transliteration | no | **2** + modifiers |
| `y///` | Synonym for `tr` | no | **2** + modifiers |

In `'...'` and `q//`, only `\\` and `\'` (or `\` + the delimiter) are escapes.
Everything else is literal — including `\n`, which is a backslash followed by
`n`.

### 2.10.2 Which words take arbitrary delimiters

`S_word_takes_any_delimiter` (`toke.c:5473-5480`) is the definitive list:

```c
static bool
S_word_takes_any_delimiter(char *p, STRLEN len)
{
    return (len == 1 && memCHRs("msyq", p[0]))
            || (len == 2
                && ((p[0] == 't' && p[1] == 'r')
                    || (p[0] == 'q' && memCHRs("qwxr", p[1]))));
}
```

That is: one-character `m`, `s`, `y`, `q`; two-character `tr`, `qq`, `qw`,
`qx`, `qr`. Exactly nine words.

This function gates whether `#` immediately after the word is a delimiter or a
comment — see the guard at `toke.c:9441`, `(!anydelim || *s != '#')`.

### 2.10.3 Delimiter rules

**Any non-word character may be a delimiter.** Word characters (letters,
digits, underscore) may not, because `qx` followed by a letter would be
ambiguous with a longer identifier. There is in fact **no character-class test
on the delimiter inside `scan_str` at all** — it takes the first character
after whitespace unconditionally (`toke.c:12303-12321`). Alphanumerics never
reach it because the *callers* gate them out first (§2.10.2). A control
character is a perfectly valid delimiter.

**Whitespace between the operator and the delimiter is always permitted.**
`scan_str` begins with an unconditional `skipspace` (`toke.c:12296-12299`):

```c
if (isSPACE(*s)) { s = start = skipspace(s); }
```

Newlines count. This holds for *every* delimiter, paired or not.

**[verified]** against perl 5.42.0 — all of these run:

```perl
print q (foo), "\n";        # foo
$_="a"; s /a/b/;            # $_ becomes "b"
print q |x|, "\n";          # x
print "match\n" if m !x!;   # match
```

**The one exception is `#`.** Because `skipspace` also eats comments, `#` is a
delimiter only when it is *immediately adjacent* to the operator word:

```perl
q#foo#      # delimiter is # -- the string "foo"
q #foo#     # q, then a COMMENT to end of line -- error
```

**[verified]**: `print q#foo#` → `foo`; `print q #foo#` →
`Can't find string terminator ";" anywhere before EOF`.

This rule does **not** live in `scan_str`. It is enforced by two caller-side
guards, both carrying the comment *"or `s###` is misparsed"*:

- `toke.c:9331-9332` — a whitespace skip that deliberately does *not* skip
  comments.
- `toke.c:9442` — `(!anydelim || *s != '#')`, where `anydelim` is
  `word_takes_any_delimiter` (§2.10.2).

> **Divergence (perl-lsp) — a documented-but-false premise.** `lib.rs:1875-1881`
> asserts that whitespace before a delimiter is allowed *only* for paired
> delimiters. That is not perl's rule. The code at `lib.rs:1901-1914` therefore
> rejects `s /a/b/`, `q |x|` and `m !x!`, all of which are legal
> **[verified]**. The `op != "s"` carve-outs exist to protect the `-s 'file'`
> filetest, a real ambiguity — but perl resolves it upstream via `PL_expect`
> (after `-`, the lexer is in term position and `s` is never a keyword there),
> not by narrowing the delimiter set.
>
> perl-lsp *does* get the `#` rule right, via `comment_eligible` in
> `skip_comment_gap_after_whitespace` (`lib.rs:2852-2874`) — but applies it
> only on the `s` path, not the `tr`/`y` paths (`lib.rs:3043-3046`).

> **Divergence (PerlOnJava).** `OperatorParser.java:1048-1063` skips whitespace
> but explicitly *not* `#`, so it takes `#` as the delimiter in `q #foo#` where
> perl sees a comment — the exact inverse of the rule. Note the internal
> inconsistency: its *other* gap-skipper, `Whitespace.skipWhitespace`
> (`Whitespace.java:76-96`), handles comments correctly and is used between the
> two halves of `s{}{}`.

**Bracketing pairs nest.** Four pairs are recognized:

| Open | Close |
| --- | --- |
| `(` | `)` |
| `[` | `]` |
| `{` | `}` |
| `<` | `>` |

With a bracketing delimiter, the scanner tracks depth (`toke.c:12430`,
`12464-12500`): `int brackets = 1`, an inner open bracket increments
(`toke.c:12492-12497`), a close decrements, and the literal ends when depth
reaches zero (`toke.c:12466-12468`). Depth tracking is enabled only when
`PL_multi_open != PL_multi_close`.

```perl
q{ outer { inner } outer }   # one string, braces balanced
q( a (b) c )                 # one string, parens balanced
```

**Non-bracketing delimiters do not nest.** With open == close, `brackets` never
rises above zero, so the literal ends at the first unescaped delimiter.

**Delimiter scanning is blind to regex syntax — deliberately.** `scan_str` has
no notion of character classes. A `}` inside `[...]` still closes a `{}`-
delimited pattern:

**[verified]** `m{[}]}` → `Unmatched [ in regex; marked by <-- HERE in m/[ <-- HERE /`

The pattern was terminated at the `}` inside the brackets. This is real,
load-bearing Perl behavior. **A lexer that "helpfully" tracks character classes
during delimiter scanning will accept code perl rejects, and produce different
token boundaries.**

> **Divergence (perl-lsp).** `parse_regex` tracks `in_character_class`
> (`lib.rs:3283-3311`), so it accepts `m{[}]}`. This is a divergence in the
> permissive direction — defensible for an editor that should not cascade
> errors, wrong as a specification of the language.

**Backslash escapes the delimiter** in both cases:

```perl
q/a\/b/     # the string  a/b
q{a\}b}     # the string  a}b
```

The full escape rule is at `toke.c:12440-12462`. Three points that matter:

1. Backslash escaping works for **both** paired and non-paired delimiters.
2. The backslash is **dropped** (leaving a literal delimiter) unless the caller
   passed `keep_bracketed_quoted`. That flag is force-cleared for non-paired
   delimiters at `toke.c:12420-12422`.
3. **`\` as the delimiter itself disables all escaping** (`toke.c:12443`).

Any other `\X` is copied through untouched: `q/a\qb/` yields `a\qb`.

Retention depends on which caller invoked the scan. `keep_bracketed_quoted` is
`TRUE` only for `scan_pat` (`toke.c:11368`) and the *pattern half* of
`scan_subst` (`toke.c:11454`) — because the regex engine must be able to
distinguish an escaped bracket from a structural one:

| Construct | Paired? | `keep_bracketed_quoted` | Result |
| --- | --- | --- | --- |
| `m/a\/b/` | no (forced FALSE) | TRUE | pattern `a/b` — backslash dropped |
| `m{a\}b}` | yes | TRUE | pattern `a\}b` — backslash **kept** |
| `q{a\}b}` | yes | FALSE | string `a}b` — dropped |
| `tr{a\}}{bc}` | yes | FALSE | `tr/a}/bc/` — dropped |

> **Divergence (both).** Neither secondary implementation models this matrix.
> PerlOnJava gates delimiter-unescaping on `isRegex && !isPair`
> (`StringParser.java:233`), narrower than perl's rule. perl-lsp consumes
> backslash-plus-next and **keeps both characters unconditionally**
> (`lib.rs:3119-3126`), so its body text is not the string perl would produce —
> acceptable when only spans are needed, wrong if the value is consumed.

### 2.10.4 The scanner

`S_scan_str` is at `toke.c:12253` (its doc comment starts there and explains
the call graph):

> `yylex()` calls `scan_str()`. `m//` makes `yylex()` call `scan_pat()` which
> calls `scan_str()`. `s///` makes `yylex()` call `scan_subst()` which calls
> `scan_str()`. `tr///` and `y///` make `yylex()` call `scan_trans()` which
> calls `scan_str()`.

So all quote-like scanning funnels through one delimiter engine, with
`scan_pat` (`toke.c:11357`), `scan_subst` (`toke.c:11438`) and `scan_trans`
(`toke.c:11531`) layering modifier parsing on top.

### 2.10.5 Two-part operators

`s`, `tr` and `y` take two delimited parts.

**With a non-bracketing delimiter, the parts share it and there are three
occurrences total:**

```perl
s/foo/bar/
tr/a-z/A-Z/
y|abc|xyz|
```

**With a bracketing delimiter, each part gets its own pair, and the second pair
may use a *different* delimiter:**

```perl
s{foo}{bar}
s{foo}(bar)
s[foo]<bar>
tr{a-z}{A-Z}
```

**Between the two parts, whitespace *and comments* are permitted** — including
newlines. This surprises most implementers.

**[verified]**:

```perl
$_ = "ab";
s{a}
# comment here
{X};
print "$_\n";
```
→ prints `Xb`. The comment between the two bracketed parts is skipped.

This is only true for the bracketing form. With a shared non-bracketing
delimiter there is no gap in which whitespace could appear.

**Implementation note.** After scanning the first part of a bracketed
two-part operator, run the full whitespace-and-comment skip before looking for
the opening delimiter of the second part. A naive `skip spaces only` will fail
on real code.

### 2.10.6 Escapes

`S_scan_const` (`toke.c:3217`) processes escapes. The governing rule for
patterns is at `toke.c:3826-3837`, and it is the opposite of what most
implementers assume:

```c
/* In a pattern, process \N, but skip any other backslash escapes. ...
   we don't want to translate an escape sequence into a meta symbol and
   have the regex compiler use the meta symbol meaning, e.g. \x{2E} would
   be confused with a dot. */
else if (PL_lex_inpat
        && (*s != 'N' || s[1] != '{' || regcurly(s + 1, send, NULL)))
{
    *d++ = '\\';
    goto default_action;
}
```

**Inside a pattern, almost every escape is passed through untranslated**, so
the regex engine sees the original text. Translating `\x{2E}` to `.` at lex
time would silently turn a literal period into the any-char metacharacter.

| Escape | `'...'` `q//` | `"..."` `qq//` | pattern | `tr///` |
| --- | --- | --- | --- | --- |
| `\\`, `\<delim>` | yes | yes | yes | yes |
| `\n` `\t` `\r` `\f` `\b` `\a` `\e` | **no** | lexed | **passed through** | lexed |
| `\0` … `\777` octal | no | lexed | **passed through** | lexed |
| `\o{...}` | no | lexed | **passed through** | lexed |
| `\x41`, `\x{263A}` | no | lexed | **passed through** | lexed |
| `\c X` | no | lexed | **passed through** | lexed |
| `\N{NAME}`, `\N{U+...}` | no | lexed | **lexed — the sole exception** | lexed |
| `\N` bare, `\N{3,5}` | — | — | passed through (`regcurly` guard) | — |
| `\l` `\u` `\L` `\U` `\Q` `\F` `\E` | no | **sublex** | **sublex** | *literal, no case ops* |
| `\d` `\w` `\s` `\D` `\W` `\S` `\B` | no | no | **engine only** | no |
| unknown `\X` | copied | copied, warns if alphanumeric | passed through | copied |

**[verified]**: `"\x41"` deparses to `'A'`; `/\x41/` deparses to `/\x41/`.

`\N{...}` is the one escape resolved at lex time even in patterns
(`toke.c:3831`, `3978`). The transformation is documented at `toke.c:3147`:

```
\N{FOO}  => \N{U+hex_for_character_FOO}
```

`charnames` resolution happens in the lexer; the engine sees the codepoint
form.

Case and quote operators are excluded from `tr///` by the
`PL_lex_inwhat != OP_TRANS` guard (`toke.c:3813`).

In an `s///` replacement, `\1`–`\9` are rewritten to `$1`–`$9` with the warning
`\%d better written as $%d` (`toke.c:3798-3810`).

Unknown alphanumeric escapes warn `Unrecognized escape \%c passed through`
(`toke.c:3841-3849`).

**The critical division:** `\L`, `\U`, `\Q`, `\l`, `\u`, `\E` are **lexer**
constructs — they become `lc`, `uc`, `quotemeta`, `lcfirst`, `ucfirst` op calls
via the sublex mechanism (§2.2.4), in both strings *and* patterns. `\d`, `\w`,
`\s` and friends are **regex engine** constructs the lexer passes through.

### 2.10.7 `qw` splitting

`qw(a b c)` produces a list of constant strings, split on whitespace, with no
interpolation and no escape processing beyond the delimiter. It yields the
`QWLIST` terminal (`perly.y:84`). Leading and trailing whitespace is ignored;
runs of whitespace are one separator.

```perl
qw(a b  c)          # ("a", "b", "c")
qw(
  alpha
  beta
)                   # ("alpha", "beta")
```

A common warning: `qw` with commas — `qw(a, b)` yields `("a,", "b")`, and perl
warns `Possible attempt to separate words with commas`.

---

## 2.11 Regular expression literals

### 2.11.1 Forms

```perl
m/pattern/flags
/pattern/flags          # bare form -- only in term position
m{pattern}flags
qr/pattern/flags
s/pattern/replacement/flags
tr/searchlist/replacementlist/flags
?pattern?               # obsolete match-once form
```

The bare `/.../` form is only recognized when a term is expected. See §2.11.3.

### 2.11.2 Modifiers

Modifiers follow the closing delimiter. Full sets:

**`m//` and `qr//`:**

| Flag | Meaning |
| --- | --- |
| `m` | `^`/`$` match at internal line boundaries |
| `s` | `.` matches newline |
| `i` | Case-insensitive |
| `x` | Extended — whitespace and `#` comments ignored |
| `xx` | As `x`, and whitespace inside bracketed classes also ignored |
| `p` | Preserve `${^MATCH}`, `${^PREMATCH}`, `${^POSTMATCH}` |
| `a` | ASCII-restricted character classes |
| `aa` | As `a`, and forbid ASCII/non-ASCII case folding |
| `d` | Default (dual ASCII/Unicode) semantics |
| `l` | Use the current locale |
| `u` | Use Unicode semantics |
| `n` | Non-capturing groups — `()` does not capture |
| `g` | Global match |
| `c` | With `/g`, do not reset pos on failure |
| `o` | Compile pattern once (largely obsolete) |
| `e` | *(`s///` only)* |
| `r` | Return the modified copy, leave the original alone |

**`s///`** takes all of the above plus:

| Flag | Meaning |
| --- | --- |
| `e` | Evaluate the replacement as Perl code |
| `ee` | Evaluate, then `eval` the result as a string |
| `r` | Non-destructive — return the result |

**`tr///` / `y///`** takes a *different* set:

| Flag | Meaning |
| --- | --- |
| `c` | Complement the search list |
| `d` | Delete unreplaced characters |
| `s` | Squash duplicate replaced characters |
| `r` | Non-destructive — return the result |

Note `c`, `d`, `s`, `r` mean entirely different things for `tr` than for `m`.
A lexer must dispatch modifier validation on the operator.

The sets are defined in `regexp.h:459-471` and parsed by `S_pmflag`
(`toke.c:11258-11354`):

```c
#define STD_PAT_MODS     "msixxn"
#define CHARSET_PAT_MODS "a" "d" "l" "u"
#define EXT_PAT_MODS     "o" "p" "n"
#define QR_PAT_MODS  STD_PAT_MODS EXT_PAT_MODS CHARSET_PAT_MODS
#define M_PAT_MODS   QR_PAT_MODS  "gc"
#define S_PAT_MODS   M_PAT_MODS   "e" "r"
```

**There are four distinct sets, not one.** `x` appears twice in
`STD_PAT_MODS` — that encodes `/xx`, counted by `x_mod_count`
(`regexp.h:405-413`): the first `x` sets `EXTENDED`, a second adds
`EXTENDED_MORE`.

**Modifiers must be immediately adjacent** — the flag loops read raw characters
with no whitespace skipping. `s{a}{b}` followed by a newline then `i` is a
syntax error.

**Charset modifiers are mutually exclusive** (`toke.c:11337-11349`):

| Condition | Diagnostic |
| --- | --- |
| Two different charset modifiers | `Regexp modifiers "/%c" and "/%c" are mutually exclusive` |
| `a` twice | Legal — upgrades to ASCII-more-restricted |
| `a` three times | `Regexp modifier "/a" may appear a maximum of twice` |
| Any other repeated | `Regexp modifier "/%c" may not appear twice` |

**The unknown-modifier rule is asymmetric** (`toke.c:11274-11285`):

- If the offending character **is** a word character →
  `Unknown regexp modifier "/%s"`, and parsing *continues* so more errors can
  be collected.
- If it is **not** a word character → not an error at all; the modifier run
  simply ends. `m/a/ + 1` is fine.

**`tr///` has no modifier validation at all.** Its loop
(`toke.c:11556-11573`) simply stops at an unrecognized character, leaving it to
be lexed as the next token. `tr/a/b/z` therefore produces a *generic* syntax
error (`Bareword found where operator expected`), not an unknown-modifier
diagnostic.

> **Divergence (both).** PerlOnJava validates against **one flat set**
> `gcr?noimsxpadeulET` (`RegexFlags.java:106`), so it accepts `m//r` and
> `qr//g`, which perl rejects; and it never validates `tr///` modifiers at all.
> perl-lsp's lexer consumes **all alphanumerics including digits** as modifiers
> (`lib.rs:3230`), which perl never does; validation moved to its parser, again
> with one set and no charset-exclusion or `xx` handling.

The `e` flag has a structural consequence: **with `/e`, the replacement part is
Perl code, not a string.** It must be lexed as code. With `/ee` it is code
producing a string that is then evaluated. `scan_subst` (`toke.c:11438`)
handles this by setting up the replacement to be parsed as an expression.

### 2.11.3 Slash disambiguation

The rule is `PL_expect`. When a term is expected, `/` starts a regex; when an
operator is expected, `/` is division and `//` is defined-or.

```perl
$x = /foo/;      # regex -- term expected after =
$x = $y / 2;     # division -- operator expected after $y
$x = $y // 2;    # defined-or
@a = grep { /x/ } @b;   # regex -- term expected inside the block
print $h{k} / 2; # division
```

The `XTERMORDORDOR` state (`perl.h:5983`, commented `/* evil hack */`) exists
for cases where both readings remain live.

perl-lsp documents its heuristic table at `mode.rs:19-28` (reproduced in
chapter 3 §3.9.4) and notes a 64KB `MAX_REGEX_BYTES` cap with graceful degradation to
`UnknownRest` (`mode.rs:32-35`) — a reasonable LSP safety valve, and an
explicit departure from perl, which has no such limit.

### 2.11.4 The `/x` modifier

Under `/x`, whitespace in the pattern is ignored and `#` introduces a comment
running to end of line.

**This IS a lexer concern, not merely a regex-engine one.** `S_scan_const`
handles it directly at `toke.c:3742-3750`:

```c
else if (*s == '#'
         && PL_lex_inpat
         && !in_charclass
         && ((PMOP*)PL_lex_inpat)->op_pmflags & RXf_PMf_EXTENDED)
{
    while (s < send && *s != '\n')
        *d++ = *s++;
}
```

The comment's characters are **copied but not scanned for interpolation** —
which is the entire point. Under `/x`, a `$var` inside a `#` comment is not
interpolated; without `/x`, it is.

Note the `in_charclass` guard, tracked by a backslash-parity scan at
`toke.c:3698-3712`, so `[#]` under `/x` is not mistaken for a comment.

Delimiter scanning is unaffected — the pattern still ends at its delimiter:

```perl
m/ foo # comment /x
```

ends at the final `/`, because `scan_str` found the delimiter before
`scan_const` ever ran.

> **Divergence (perl-lsp).** No equivalent of `toke.c:3742-3750` exists, so
> `/x`-comment content is not excluded from interpolation scanning. A `$var`
> inside an `/x` comment will be reported as a variable reference.

### 2.11.5 Embedded code blocks

`(?{ CODE })` and `(??{ CODE })` embed Perl code in a pattern.

These require the code to be lexed as Perl, and perl handles it by tracking
`PL_parser->re_eval_start` — the note at `toke.c:10151` warns that
`S_scan_heredoc` adjusts it, i.e. a heredoc inside an embedded code block is a
real (if perverse) case that perl accounts for.

For a from-scratch implementation:

- At minimum, track brace depth inside `(?{ ... })` so the pattern's closing
  delimiter is found correctly. A `}` inside a string inside the code block
  must not be miscounted.
- Ideally, recursively lex the block as Perl code and emit real tokens, so the
  type checker can see it.
- Note that `use re 'eval'` is required for these to be permitted with
  interpolated patterns — a compile-time check, not a lexical one.

`(?#...)` is a plain regex comment and needs no code handling.

### 2.11.6 The match-once `?pattern?` form

Two forms must be distinguished; they have different fates.

**Bare `?pattern?` is removed.** `case '?'` at `toke.c:9810-9819` is now
unconditionally the ternary operator. **[verified]**:

```perl
$_="a"; print "ok\n" if ?a?;   # syntax error at -e line 1, near "if ?"
```

**`m?pattern?` still works.** `scan_pat` honors it at `toke.c:11373-11392`:

```c
if (PL_multi_open == '?') pm->op_pmflags |= PMf_ONCE;
```

This is the **only** place `PMf_ONCE` is set. The PMOP is registered in the
stash's `PERL_MAGIC_symtab` so that `reset` can find and re-arm it.

**[verified]**:

```perl
$_="a"; print "ok\n" if m?a?;   # prints ok
```

So `?` is a legal delimiter for `m`, but never an implicit match introducer.
Recognize bare `?...?` only to produce a good diagnostic for old code.

---

## 2.12 The `<...>` family

`S_scan_inputsymbol` (`toke.c:12070`) handles everything after a `<` in term
position. Its doc comment (`toke.c:12057-12066`) enumerates the forms:

| Form | Meaning |
| --- | --- |
| `<>` | Read from `ARGV` |
| `<<>>` | Read from `ARGV` without magic open (no shell metacharacter handling) |
| `<FH>` | Read from filehandle `FH` |
| `<pkg::FH>` | Package-qualified filehandle |
| `<pkg'FH>` | Package-qualified, archaic separator |
| `<$fh>` | Read from the filehandle in `$fh` |
| `<*.h>` | Filename glob |

### 2.12.1 The readline-versus-glob decision

At `toke.c:12112-12128`. The algorithm:

1. Copy everything up to `>` into a buffer (`delimcpy`, `toke.c:12088`).
2. If it starts with `$` and has more characters, skip the `$`
   (`toke.c:12111`) — because *"except for the `$` at the front, a scalar
   variable and a filehandle look the same"* (`toke.c:12105-12109`).
3. Parse an identifier with `ALLOW_PACKAGE` (`toke.c:12118`).
4. **If any text remains unconsumed, it is a glob, not a readline**
   (`toke.c:12126-12132`) — re-scan with `scan_str` and set `OP_GLOB`.

So `<$fh>` is a readline (the whole thing parses as an identifier after the
`$`), but `<$fh{x}>` is a glob (the `{x}` does not parse as part of an
identifier). The doc comment says so explicitly (`toke.c:12100-12104`):

> Remember, only scalar variables are interpreted as filehandles by this code.
> Anything more complex (e.g., `<$fh{$num}>`) will be treated as a `glob()`
> call.

### 2.12.2 Errors

| Condition | Message | Citation |
| --- | --- | --- |
| Content longer than `PL_tokenbuf` | `Excessively long <> operator` | `toke.c:12097` |
| No closing `>` before newline or EOF | `Unterminated <> operator` | `toke.c:12099` |
| Glob form unterminated | `Glob not terminated` | `toke.c:12130` |

Note the newline restriction: `end` is set to the next `\n` or `PL_bufend`
(`toke.c:12078-12080`), so **a `<...>` construct may not span lines.**

---

## 2.13 Formats

`format` declarations introduce a picture-line mini-language. `S_scan_formline`
is at `toke.c:13271`, and `LEX_FORMLINE` (`toke.c:164`) is the dedicated lexer
state.

```perl
format STDOUT =
@<<<<<<<<   @>>>>>   @#####.##
$left,      $right,  $number
.
```

Rules:

- The body begins after the `=` and the following newline.
- It ends at a line containing exactly a single `.`.
- Picture lines and argument lines alternate; the lexer does not interpret the
  picture characters, it delimits lines.
- `FORMLBRACK` and `FORMRBRACK` (`perly.y:92`) bracket the argument
  expressions.
- `$^A` is the accumulator (§2.7.5).

perl-lsp models this with `InFormatBody` mode (`mode.rs:58`) and a `FormatBody`
token (`token.rs:60-61`). Formats are rare in modern code; emitting the body as
one opaque token with a correct span is an acceptable v1.

---

## 2.14 What the lexer must expose to a checkpoint

Incremental lexing — the checkpoint struct, where checkpoints are recorded,
the restart-and-resync loop — is chapter 6 §6.3. This section lists only the
state that is the lexer's to expose, and the two lexer behaviours an LSP needs
that perl does not have.

### 2.14.1 What must be checkpointed

Everything that affects how the next character is interpreted:

- `PL_expect` and the bracket stack (chapter 3 §3.1, §3.1.4).
- `PL_lex_state` (§2.2.3) and the sublex context stack (§2.2.4).
- The pending-heredoc queue (§2.9.4). **Must** be included.
- Whether we are inside POD or past `__END__`/`__DATA__` (§2.5).
- `use utf8` (§2.3.1) and the feature bits the lexer reads — the `'`
  separator (§2.6.3).
- `PL_last_lop_op`, for `sort` and the filehandle cases (chapter 3 §3.4.4),
  and whether the previous token was a list or named-unary operator
  (`PL_last_lop`, `PL_last_uni`), which the `{` and `(` rules read
  (chapter 3 §3.1.4, §3.3 step 3).

> **Divergence (perl-lsp).** Its checkpoint (`checkpoint_impl.rs:23-36`)
> captures `position`, `mode`, `delimiter_stack`, `in_prototype`,
> `prototype_depth`, `after_sub`, `after_arrow`, `hash_brace_depth`,
> `after_var_subscript`, `paren_depth`, `current_pos`, `eof_emitted` and a
> `context` enum — but **no pending-heredoc queue and no sublex stack**. Two
> further approximations are visible in the same file: the `Format` context
> stores `start_position: self.position.saturating_sub(100)` with the comment
> `// Approximate` (`checkpoint_impl.rs:14`), and the `QuoteLike` context
> stores `operator: String::new()` with the comment
> `// Would need to track this` (`checkpoint_impl.rs:18`). A restore inside a
> quote-like construct therefore loses which operator opened it.

### 2.14.2 Safe restart points

Chapter 6 §6.3.1. The lexer-side condition is `Expect == XSTATE` with every
stack above empty and no pending heredoc.

### 2.14.3 Budgets and recovery

An LSP must never hang. Two mechanisms, both borrowed from perl-lsp:

- **A byte budget per construct.** perl-lsp caps regex literals at 64KB
  (`mode.rs:33`). On exceeding it, emit `UnknownRest` and stop
  (`token.rs:69-70`).
- **An `Error` token that does not stop lexing** (`token.rs:126`). A malformed
  number or an unknown escape should produce a diagnostic and a token, then
  scanning continues.

Perl itself has neither — it croaks. For a compiler that is correct; for an
editor it is unusable.

### 2.14.4 Recovery points

Parser-side resynchronisation is chapter 5 §5.13.4.

---

## 2.15 Implementation checklist

Everything the lexer must do, grouped by dependency. Build order across the
whole system is chapter 6 §6.11; milestone gates are chapter 7 §7.8.

1. **Byte reader with a line index.** UTF-8 decode with `RuneError` recovery.
   BOM handling. `\r\n` normalization that preserves offsets.
2. **Trivia.** Whitespace, `#` comments, `#line` directives, shebang. Retain
   as tokens with spans.
3. **Expectation state machine.** All eleven states (chapter 3 §3.1).
   Everything else depends on it.
4. **Identifiers and package-qualified names.** Both separators, `::` and `'`.
   XID-based Unicode under a `utf8` flag. **Do not accept emoji** (§2.6.2).
5. **Sigils and punctuation variables.** Table-driven on the character after
   the sigil (§2.7.5). Braced and caret forms.
6. **Numbers.** All radices, underscores-as-warnings-not-errors, floats,
   `p`-exponent floats in every radix, the `s[1] != '.'` range guard.
7. **V-strings.** Both the `v` prefix and the bare two-dot form. The `=>`
   escape hatch.
8. **Simple strings.** `'...'`, `"..."`, `` `...` ``, with escape processing
   split by context.
9. **Quote-like operators.** The nine-word table (§2.10.2), the delimiter
   engine, bracketing and nesting, the `#`-delimiter asymmetry.
10. **Two-part operators.** `s`, `tr`, `y`, including whitespace *and comments*
    between bracketed parts (§2.10.5).
11. **Regex modifiers.** Per-operator modifier sets (§2.11.2). `/e` making the
    replacement code.
12. **POD, `__END__`, `__DATA__`.** POD only at `XSTATE`, column 0, alpha after
    `=` (§2.5.3).
13. **Heredocs.** The pending queue, FIFO drain at logical-line end, all six
    terminator forms, `<<~` stripping (§2.9).
14. **`<...>` family.** Readline versus glob by the leftover-text rule
    (§2.12.1).
15. **Interpolation sublexing.** The explicit context stack (§2.2.4).
16. **Formats.** `LEX_FORMLINE`, terminate on a lone `.` (§2.13).
17. **Checkpointing.** Everything in §2.14.1, heredoc queue included.
18. **Budgets and recovery tokens.** `Error` and `UnknownRest` (§2.14.3).
19. **VCS conflict markers.** Seven-character run at line start (§2.3.6).

### 2.15.1 Test corpus

Every **[verified]** example in this chapter is a test case with a known
expected result. In addition, exercise:

- Multiple heredocs on one line, in an argument list, with `<<~`.
- A heredoc whose body contains what looks like another heredoc introducer.
- `s{...}` with a comment before `{...}`.
- `q#...#` versus `q #...#`.
- `$#array`, `$#{$r}`, `$#$r`, and `$#` alone.
- `5.42` versus `5.42.0` versus `v5.42` versus `( v65 => 1 )`.
- `/` in every position: after `=`, after `)`, after a bareword, inside
  `grep {}`.
- POD at a statement boundary inside a sub body; POD mid-expression (must
  error); indented POD (must error).
- `use utf8` turning on mid-file.
- An unterminated heredoc, string, and regex at EOF — each must produce one
  diagnostic and no cascade.

---

## 2.16 Summary of divergences

Collected for reference. Each is a place where a secondary source should not be
copied.

| # | Source | Divergence | Section |
| --- | --- | --- | --- |
| 1 | PerlOnJava | 7 token types only; a pre-tokenizer, not a Perl lexer | §2.1.2 |
| 2 | PerlOnJava | `10E10` mis-lexed at the lexer layer, repaired downstream | §2.8.5 |
| 3 | PerlOnJava | `qq<=>` mis-lexed as `qq` + `<=>` | §2.1.2 |
| 4 | PerlOnJava | No heredoc handling in the lexer at all | §2.9.9 |
| 5 | PerlOnJava | No `Misplaced _ in number` warning | §2.8.4 |
| 6 | PerlOnJava | No `Integer overflow` warning | §2.8.9 |
| 7 | PerlOnJava | No `=>` escape hatch for v-strings | §2.8.8 |
| 8 | perl-lsp | Accepts emoji as identifiers; perl rejects them **[verified]** | §2.6.2 |
| 9 | perl-lsp | Accepts `'` unconditionally as identifier-continue | §2.6.3 |
| 10 | perl-lsp | 5 modes cannot express `XSTATE`; POD rule unreachable | ch3 §3.9.4 |
| 11 | perl-lsp | No bare `N.N.N` v-strings — `use 5.42.0` mis-lexed | §2.8.9 |
| 12 | perl-lsp | No hex floats; no bare `0NNN` octal; no illegal-digit check | §2.8.9 |
| 13 | perl-lsp | `should_consume_dot` allowlist approximates `s[1] != '.'` | §2.8.10 |
| 14 | perl-lsp | Checkpoint omits the heredoc queue and sublex stack | §2.14.1 |
| 15 | perl-lsp | Checkpoint `Format.start_position` marked `// Approximate` | §2.14.1 |
| 16 | perl-lsp | Checkpoint `QuoteLike.operator` left empty | §2.14.1 |
| 17 | both | Neither implements `PL_infix_plugin` operator terminals | §2.1.1 |
| 18 | perl-lsp | Claims whitespace before a delimiter needs a paired delimiter; rejects `s /a/b/`, `q \|x\|`, `m !x!` — all legal **[verified]** | §2.10.3 |
| 19 | PerlOnJava | Takes `#` as a delimiter with a gap; perl sees a comment | §2.10.3 |
| 20 | perl-lsp | Tracks character classes in regex delimiter scanning; accepts `m{[}]}` which perl rejects **[verified]** | §2.10.3 |
| 21 | both | Collapse perl's four modifier sets into one; accept `m//r`, `qr//g` | §2.11.2 |
| 22 | perl-lsp | Lexer consumes digits as regex modifiers | §2.11.2 |
| 23 | both | Neither models the `keep_bracketed_quoted` escape matrix | §2.10.3 |
| 24 | perl-lsp | Rejects `<<1` (digit heredoc label), which perl accepts **[verified]** | §2.9.3 |
| 25 | perl-lsp | Accepts trailing whitespace after a heredoc terminator; perl rejects it **[verified]**, and its code comment claims otherwise | §2.9.3 |
| 26 | perl-lsp | No `<<~` indentation stripping and no mismatch error | §2.9.7 |
| 27 | PerlOnJava | Exempts partial-whitespace lines under `<<~`; perl makes them fatal | §2.9.7 |
| 28 | perl-lsp | No `/x` comment handling — interpolates inside `/x` comments | §2.11.4 |
| 29 | perl-lsp | `<<` heuristic `mode==ExpectOperator && paren_depth>0` mis-lexes `1 << WIDTH` at statement level as a heredoc | §2.9.2 |

### Corrections to widely held beliefs

Each established by running perl 5.42.0 directly:

1. **Binary and octal floats work** (`0b1.1p2`, `017.4p1`). The `p` exponent is
   not hex-only — `HEXFP_PEEK` is tested in the generic radix loop (§2.8.6).
2. **POD cannot appear mid-expression.** It requires `PL_expect == XSTATE`,
   column 0, and an alpha after `=`. It *can* appear at any statement boundary,
   which is the genuinely awkward part (§2.5.3).
3. **A stray `=cut` starts a POD block** and silently swallows the rest of the
   file (§2.5.3).
4. **Whitespace before a quote delimiter is always allowed**, for every
   delimiter kind. Only `#` is special, and only by adjacency (§2.10.3).
5. **The archaic `'` package separator is not deprecated — it is
   version-gated**, switching off under `use v5.42`+ (§2.6.3).
6. **`__END__`/`__DATA__` are keywords, not line-anchored markers** — no
   column-0 requirement, unlike POD (§2.5.4).
7. **`m?pattern?` still works**; only *bare* `?pattern?` was removed (§2.11.6).
8. **`/x` is a lexer concern**, not purely a regex-engine one — it suppresses
   interpolation inside pattern comments (§2.11.4).
9. **Delimiter scanning is character-class-blind by design.** `m{[}]}` is a
   syntax error in perl; a lexer that "fixes" this diverges (§2.10.3).
