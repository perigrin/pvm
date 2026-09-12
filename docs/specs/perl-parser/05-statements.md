<!-- ABOUTME: Chapter 5 of the Perl language specification for parser implementors: statements, -->
<!-- ABOUTME: declarations, program structure, error recovery, and incremental re-parse boundaries. -->

# Chapter 5 — Statements, Declarations, and Program Structure

**Audience.** You are writing a hand-rolled recursive-descent Perl parser in Go
using only the standard library. This chapter tells you what a *statement* is,
what forms declarations take, how a program is structured, and — because the
target is a language server — where you may safely stop and restart parsing.

**Ground truth.** Grammar productions are cited from `perly.y` in a Perl blead
tree at `PERL_REVISION 5 / PERL_VERSION 45` (`patchlevel.h:40-41`). Lexer
behaviour is cited from `toke.c` in the same tree. Where a construct's shape is
decided by the *lexer* rather than the grammar — and in Perl that is often —
the chapter says so explicitly, because a recursive-descent parser has no
separate lexer pass to hide behind.

**Empirical claims are checked, not assumed.** Where this chapter asserts that
something parses, errors, or produces a particular result, it was run against
perl v5.42.0. That discipline caught six errors during drafting, in three
categories:

* **The grammar is more permissive than the language.** `sub f ($x =) {}`
  reduces in `perly.y` but is rejected by a later check (§5.6.2).
* **A plausible-sounding rule is simply false.** A lexical sub is *not* visible
  in its own body (§5.5.7); stacked labels do not all attach to the loop
  (§5.2.2).
* **A reference implementation diverges from Perl.** PerlOnJava strips a leading
  underscore when deriving a `:reader` name, and ignores constructor arguments
  for `//=`/`||=` field defaults. Perl does neither (§5.7.3, §5.7.6).

That last category is the reason to treat the secondary references as evidence
about *engineering*, not about *semantics*. If you extend this chapter, run the
example.

Two secondary references are cited for engineering decisions rather than
semantics:

* **PerlOnJava** — `src/main/java/org/perlonjava/frontend/parser/*.java`. A
  full recursive-descent Perl parser in Java. Structurally the closest existing
  thing to what you are building.
* **perl-lsp** — `crates/perl-parser-core/src/engine/parser/*.rs`. A
  recursive-descent Perl parser built for a language server, with real error
  recovery.

---

## 5.0 The one thing to understand first

Perl's grammar is not context-free, and the place this hurts most is statement
parsing. `perly.y` is a bison grammar, but it only works because `toke.c` does
enormous amounts of work before a token ever reaches the parser: it looks
ahead, it consults the lexical pad, it consults the current feature bundle, and
it *rewrites the token stream* by pushing synthetic tokens onto a queue
(`force_next`).

Concretely, `toke.c` will hand the grammar different token types for the
identical source text depending on state:

| Source | Token emitted | Deciding state |
| --- | --- | --- |
| `sub foo {...}` | `KW_SUB_named` | signatures feature **off** |
| `sub foo {...}` | `KW_SUB_named_sig` | signatures feature **on** |
| `sub {...}` | `KW_SUB_anon` / `KW_SUB_anon_sig` | same |
| `method foo {...}` | `KW_METHOD_named` | always implies signatures |
| `continue` | `KW_CONTINUE` | next non-space char is `{` |
| `continue` | `OP_CONTINUE` | anything else |
| `BEGIN` | routes to `yyl_sub` | `PL_expect == XSTATE` |
| `BEGIN` | plain bareword | any other expectation |

See `toke.c:5862` (`is_sigsub = is_method || FEATURE_SIGNATURES_IS_ENABLED`),
`toke.c:5962-5975` (the four-way token choice), `toke.c:8385-8395`
(`continue`), and `toke.c:8313-8322` (phasers).

**Design consequence for your Go parser.** Do not build a lexer that produces a
token stream independent of the parser. Build a *scanner* that the parser
drives, and thread a parse-state struct through it. The minimum state:

```go
// ParseState is threaded through the whole parse. It is the Go equivalent of
// the PL_expect / PL_hints / feature-bits soup in toke.c.
type ParseState struct {
    Expect   Expect     // the eleven-state PL_expect enum, chapter 3 §3.1
    BrackStack []Expect // PL_lex_brackstack: what `}` restores, chapter 3 §3.1.4
    Features FeatureSet // signatures, say, isa, try, class, module_true, ...
    Strict   StrictBits
    InClass  bool       // `field`/`method`/`ADJUST` legal only here
    Pad      *ScopeChain// lexical names in scope; needed for bareword decisions
    Package  string     // current `package`, for name qualification
}
```

`Expect` is not optional decoration. `toke.c` uses `PL_expect` to decide
whether `{` opens a block or an anonymous hash, whether `/` is division or a
regex, and whether `BEGIN` is a phaser or a bareword. Chapter 3 §3.1 specifies
the enum, its transitions, and the bracket stack; carry it unchanged.

---

## 5.1 Program structure

### 5.1.1 The top level

A Perl compilation unit is a sequence of statements (`perly.y:187-197`, the
`GRAMPROG` alternative). There is no required `main`, no required package
declaration, no required anything. An empty file is a valid program.

```ebnf
Program   = StmtSeq EOF_or_Section ;
StmtSeq   = { FullStmt } ;
```

`perly.y:824-832`:

```
stmtseq
	:	empty
	|	stmtseq[list] fullstmt
```

Left-recursive in bison; in recursive descent this is simply a loop. Note that
`stmtseq` *can* be empty, and that `fullstmt` can itself produce nothing (the
empty statement, §5.2.4). Your loop must therefore tolerate a statement parse
returning `nil` without treating that as an error or as end-of-input.

### 5.1.2 Where the file actually ends

The token stream ends at physical EOF or at the first `__END__` / `__DATA__`
keyword. Recognition — including the **[verified]** fact that neither marker
is line-anchored — is chapter 2 §2.5.4; do not re-derive it here. On the
parser side, `toke.c:8295-8299` routes both `KEY___DATA__` and `KEY___END__`
to `yyl_fake_eof`, the same function used for real EOF, so the parser cannot
tell them apart. Everything after the marker is *not parsed*:

```go
type File struct {
    Stmts   []Stmt
    Data    *DataSection // nil if no __END__/__DATA__
}

type DataSection struct {
    Marker  string // "__END__" or "__DATA__"
    Package string // package current at the marker; DATA handle lives here
    Span    Span   // byte range of the content, NOT tokenised
}
```

The `Package` field exists because `__DATA__` opens `<current package>::DATA`
while `__END__` always opens `main::DATA` (§2.5.4; PerlOnJava records the
same rule in `dataHandleName`, `DataSection.java:201`, and the `require`d-file
asymmetry at `:318`). `<DATA>` reads **raw bytes**, so the span must be taken
from the undecoded source; PerlOnJava extracts it from the raw file for
exactly this reason (`DataSection.java:333-336`).

> **Divergence (PerlOnJava).** `DataSection.java:261-268` requires the marker
> at the start of a line. perl does not — §2.5.4, **[verified]** mid-line and
> indented.

### 5.1.3 POD

POD recognition — column 0, `isALPHA` after `=`, only at `XSTATE`, the stray
`=cut` hazard — is chapter 2 §2.5. For an LSP keep POD in the tree as trivia
rather than discarding it, so that hover and folding work:

```go
type Pod struct {
    Span     Span
    Command  string // "head1", "over", "cut", ...
    Terminated bool // false if EOF was reached first
}
```

### 5.1.4 The shebang line

Chapter 2 §2.3.4. Treat it as trivia but record any switches — `-M` can enable
pragmas that change the parse.

---

## 5.2 Statement forms

### 5.2.1 The statement hierarchy

`perly.y` splits statements into three layers, and you should mirror this
split exactly because it makes labels and statement modifiers fall out cleanly.

```ebnf
FullStmt  = BareStmt
          | Label FullStmt ;            (* labels nest *)
BareStmt  = (* the 24 alternatives below *) ;
```

`perly.y:846-868`:

```
fullstmt:	barestmt
	|	labfullstmt
	;

labfullstmt:	LABEL barestmt
	|	LABEL labfullstmt[list]
	;
```

Note the second `labfullstmt` alternative: **labels stack syntactically**. `A:
B: C: for (...) {...}` parses, and all three labels are in the tree — so your
AST should hold a `[]Label`, not a single optional label.

But only the **innermost** label attaches to the loop. Verified against
v5.42.0 with `perl -MO=Deparse -e 'A: B: for (1..3){ last B }'`:

```perl
A: B: ;
foreach $_ (1 .. 3) {
    last B;
}
```

The outer labels degenerate to empty labeled statements. `last B` works;
`last A` dies with `Label not found for "last A"`. PerlOnJava reaches the same
result by keeping only the last label of a run (`ParseBlock.java:135-141`),
though it also records each as a `LabelNode` statement — which is the more
faithful shape, and the one to copy: keep every label for hover and
go-to-definition, but resolve `last`/`next`/`redo` only against the innermost.

The full `barestmt` alternative list is `perly.y:877-903`. As of this tree it
is a flat list of 24 named nonterminals, alphabetically sorted — a recent
refactor that is a gift to anyone writing a parser, because each alternative
now has a name you can map one-to-one onto a Go function:

| `perly.y` nonterminal | Line | Construct |
| --- | --- | --- |
| `PLUGSTMT` | 878 | keyword-plugin statement |
| `bare_statement_block` | 269 | bare block `{ ... }` + optional `continue` |
| `bare_statement_class_declaration` | 277 | `class NAME;` |
| `bare_statement_class_definition` | 293 | `class NAME { ... }` |
| `bare_statement_default` | 316 | `default { ... }` |
| `bare_statement_defer` | 324 | `defer { ... }` |
| `bare_statement_expression` | 331 | `EXPR;` with modifiers |
| `bare_statement_field_declaration` | 339 | `field $x;` |
| `bare_statement_for` | 347 | all `for`/`foreach` forms |
| `bare_statement_format` | 487 | `format NAME = ...` |
| `bare_statement_given` | 503 | `given (...) { }` |
| `bare_statement_if` | 516 | `if (...) { } elsif ... else ...` |
| `bare_statement_null` | 531 | `;` |
| `bare_statement_package_declaration` | 539 | `package NAME;` |
| `bare_statement_package_definition` | 555 | `package NAME { ... }` |
| `bare_statement_phaser` | 573 | `ADJUST { ... }` |
| `bare_statement_sub_signature` | 604 | `sub`/`method` with signature |
| `bare_statement_sub_traditional` | 636 | `sub` with prototype |
| `bare_statement_try_catch` | 661 | `try { } catch ($e) { } finally { }` |
| `bare_statement_unless` | 683 | `unless (...) { }` |
| `bare_statement_until` | 698 | `until (...) { }` |
| `bare_statement_utilize` | 713 | `use` / `no` |
| `bare_statement_when` | 731 | `when (...) { }` |
| `bare_statement_while` | 743 | `while (...) { }` |
| `bare_statement_yadayada` | 758 | `...;` |

Your dispatch function is a switch on the first token, plus a few lookaheads.
Everything not matching a keyword falls through to the expression statement.

```go
func (p *Parser) parseStatement() Stmt {
    labels := p.parseLabels()          // may be empty; labels stack
    s := p.parseBareStatement()
    if s == nil { return nil }         // empty statement
    s.setLabels(labels)
    return s
}
```

### 5.2.2 Labels

```ebnf
Label = IDENT ":" ;   (* not "::" *)
```

A label is an identifier followed by a single colon that is *not* part of `::`
and *not* the colon of a ternary. The lexer resolves this; in recursive descent
you resolve it with a two-token lookahead at statement position only:

```go
// A label is only recognised where a statement may begin.
func (p *Parser) atLabel() bool {
    if p.tok.Kind != IDENT { return false }
    if p.peek(1).Kind != COLON { return false }
    if p.peek(2).Kind == COLON { return false } // Foo::bar, not a label
    return true
}
```

Two guards, both from PerlOnJava's `ParseBlock.parseLabel`
(`ParseBlock.java:224-235`), that you will need:

* **Reject quote-like operators.** `m:...:` is a match with `:` delimiters, not a
  label named `m` (`:227-230`). Same for `s`, `tr`, `y`, `q`, `qq`, `qw`, `qr`.
* **Reject `sub`** (`:233-235`).

Labels are conventionally uppercase but that is style, not grammar. A label may
precede *any* bare statement, not only a loop — `FOO: print "hi";` parses. It
is only meaningful on loops and bare blocks, where `last`/`next`/`redo` can
target it.

### 5.2.3 Expression statements

`perly.y:331-337` and `perly.y:933-952`:

```
bare_statement_expression
	:	sideff PERLY_SEMICOLON
```

```ebnf
ExprStmt = Expr [ StmtModifier ] ";" ;
```

The `sideff` nonterminal is where statement modifiers live (§5.4). A plain
expression statement is `sideff` with no modifier.

The name `sideff` — "side effect" — is a hint about Perl's evaluation model:
the statement's value is discarded, so a statement whose expression has no side
effect is a `Useless use of ... in void context` warning. Your parser should
not enforce that, but PSC-style analysis may want the flag.

### 5.2.4 The empty statement

`perly.y:531-537`:

```
bare_statement_null
	:	PERLY_SEMICOLON
		{
			$$ = NULL;
			parser->copline = NOLINE;
		}
```

A lone `;` is a valid statement producing no op. `;;;;` is four empty
statements. Crucially the action returns `NULL`, and `fullstmt` at
`perly.y:846-849` guards on that:

```
fullstmt:	barestmt
			{
			  $$ = $barestmt ? newSTATEOP(0, NULL, $barestmt) : NULL;
			}
```

So an empty statement does not even get a statement op — and note the
`parser->copline = NOLINE` in the action: an empty statement also *resets* the
pending line number, so it does not shift the line attributed to the next
statement. PerlOnJava models this explicitly (`ParseBlock.java:148-157`, "lone
semicolons reset the flag without producing a COP, matching perl's
`bare_statement_null`"). If you care about `caller`/`warn`/`die` line fidelity,
you need the same rule.

In your AST you have two reasonable choices: return `nil` and skip it, or emit
an `EmptyStmt` node. **For an LSP, emit the node.** You need the span so that a `;` deletion is a
localised edit rather than a whole-file reparse (chapter 6 §6.5).

Note that a statement being `NULL` is *also* how `package NAME;`, `use`, `sub
NAME {...}`, `format`, and `ADJUST` report themselves — they are compile-time
effects with no runtime op. Do not conflate "produced no op" with "produced no
syntax". Your AST needs all of them.

### 5.2.5 Blocks as statements

`perly.y:269-275`:

```
bare_statement_block
	:	block
		cont
```

```ebnf
BlockStmt = Block [ "continue" Block ] ;
Block     = "{" StmtSeq "}" ;
```

A bare block is a **loop that executes exactly once**. This is not a curiosity:
`last`, `next`, and `redo` work inside it (`next` and `last` both exit; `redo`
restarts). That is why `cont` — the `continue` block — is attached here at all
(`perly.y:972-976`).

```perl
OUTER: {
    last OUTER if $skip;
    do_work();
}
```

Contrast with `do BLOCK`, which is *not* a loop (§5.3.6).

`block` vs `mblock` (`perly.y:775-780` vs `perly.y:799-804`) differ only in
whether they start a full or partial lexical scope (`block_start(TRUE)` vs
`block_start(FALSE)`). This matters for `my` visibility in constructs like
`if (my $x = f()) { ... }`, where `$x` must be visible in the block. For a
parser producing an AST, both are just `Block`; record the distinction only if
you are doing scope resolution, in which case: `mblock` shares the scope opened
by the enclosing condition, `block` opens its own.

### 5.2.6 The yada-yada statement

`perly.y:758-764`:

```
bare_statement_yadayada
	:	YADAYADA PERLY_SEMICOLON
```

`...;` is a statement that dies with "Unimplemented" at runtime. It parses only
in statement position, which distinguishes it from the `...` range operator.
Since `..` and `...` are also operators, the lexer decides by expectation.

```go
// `...` at statement position is the yada-yada; elsewhere it is the range op.
```

---

### 5.2.7 Statement line attribution

Every statement in Perl carries a line number, used by `caller`, `warn`, `die`,
and the debugger. Getting it right is not optional for a language server, since
it is what makes a stack trace map to the right line.

`perly.y` threads `parser->copline` through nearly every statement action —
`parser->copline = (line_t)$KW_IF` (`perly.y:527`), `= (line_t)$KW_FOR`
(`perly.y:377`), `= NOLINE` for the empty statement (`perly.y:535`), and the
`if (parser->copline > (line_t)$PERLY_BRACE_OPEN)` clamp in every block rule
(`perly.y:776-777`, `:800-801`, `:1226-1227`, `:1245-1246`). The clamp exists so
a multi-line block reports its opening brace, not its last line.

The rule is subtler than "the line the statement starts on", and PerlOnJava
models it in a dedicated class, `StatementCopline.java`, whose class comment
(`:11-47`) is the best specification available. Two behaviours:

**1. Only some tokens arm the line.** Only tokens that reach `TERM()`/`LOP()` in
`toke.c` set the pending line. PerlOnJava keeps an explicit `NON_ARMING_WORDS`
set (`StatementCopline.java:59-79`) containing `my`, `return`, `if`, `sub`, and
most named unary operators — and *deliberately excludes* the quote-like
operators `q qq m s tr qr`, because those arm via `sublex_start()`
(`:53-57`). `ARMING_OPERATORS` (`:86-90`) is sigils, `)`, `]`, quote characters,
`++`, `--`. The consequence, from their comment: `sub f { return\n $x\n ->m; }`
reports the `$x` line, not the `return` line.

**2. A brace-terminated statement swallows one lookahead token.** Deciding that
a compound statement has ended requires reading one token past its `}` — which
is the *next* statement's first token — and building the next statement's line
record discards whatever that token armed. So a compound statement shifts the
line attributed to its successor.

Constructs that produce no line record do not consume anything and so do not
shift: a named `sub`, `package NAME;`, a phaser. PerlOnJava's
`leavesLookaheadSwallowed` (`ParseBlock.java:67`) exempts nodes annotated
`compileTimeOnly` or `noReturnValue` for exactly this reason.

If you do not need `caller`-accurate line numbers, record the start line of each
statement and move on. If you do, budget for this; it is fiddly and it is the
kind of thing that is much cheaper to build in than to retrofit.

## 5.3 Control flow

### 5.3.1 Braces are mandatory

This is the single biggest structural difference from C, and it simplifies your
parser enormously. Every conditional and loop body in Perl is a `block` or
`mblock` — a `{ ... }` — never a single statement:

```
bare_statement_if
	:	KW_IF PERLY_PAREN_OPEN remember mexpr PERLY_PAREN_CLOSE mblock else
```

(`perly.y:516-529`.) There is no production anywhere that lets a bare statement
be a conditional body. Consequences:

* No dangling-else ambiguity. Ever.
* `if ($x) print "hi";` is a syntax error, and you can produce a *good* error
  for it: "syntax error near `print`; Perl requires braces around the body of
  `if`".
* Your `parseIf` can call `parseBlock` unconditionally and error cleanly if the
  next token is not `{`.

The parentheses around the condition are equally mandatory.

### 5.3.2 `if` / `elsif` / `else` / `unless`

```ebnf
IfStmt     = "if" "(" Expr ")" Block ElseTail ;
UnlessStmt = "unless" "(" Expr ")" Block ElseTail ;
ElseTail   = ε
           | "else" Block
           | "elsif" "(" Expr ")" Block ElseTail ;
```

`perly.y:955-969`:

```
else
	:	empty
	|	KW_ELSE mblock
	|	KW_ELSIF PERLY_PAREN_OPEN mexpr PERLY_PAREN_CLOSE mblock else[else.recurse]
```

`elsif` is right-recursive; in your parser it is a loop or a tail call, whichever
you prefer. Note there is no `elsunless`.

`unless` is *not* sugar that your parser should desugar. `perly.y:683-696`
builds the same `newCONDOP` but with the branches swapped:

```
			$$ = block_end($remember, newCONDOP(OPpSTATEMENT<<8,
                                        $mexpr, $else, op_scope($mblock)));
```

Compare `if` at `perly.y:525-526`, which passes `op_scope($mblock), $else`. For
an LSP you want to preserve the surface form — a user renaming `unless` to `if`
expects the parse to reflect what they typed — so keep them as distinct nodes
with a flag, or distinct node types.

`unless ... elsif ...` is legal Perl. It is confusing Perl, but it parses,
because `else` in the grammar is shared.

```go
type IfStmt struct {
    Negated bool   // true for `unless`
    Cond    Expr
    Then    *Block
    // Elifs are flattened; each has its own Cond and Block.
    Elifs   []struct{ Cond Expr; Body *Block }
    Else    *Block // nil if absent
    Span    Span
}
```

### 5.3.3 `while` and `until`

```ebnf
WhileStmt = "while" "(" [ Expr ] ")" Block [ "continue" Block ] ;
UntilStmt = "until" "(" Expr ")"     Block [ "continue" Block ] ;
```

`perly.y:743-756` and `perly.y:698-711`. Two asymmetries worth knowing:

**`while` accepts an empty condition; `until` does not.** The `while`
production uses `texpr` (`perly.y:998-1003`):

```
texpr	:	%empty /* NULL means true */
			{ YYSTYPE tmplval;
			  (void)scan_num("1", &tmplval);
			  $$ = tmplval.opval; }
	|	expr
```

So `while () { ... }` is an infinite loop, equivalent to `while (1)`. The
`until` production uses `iexpr` (`perly.y:1006-1008`), which requires an `expr`
and wraps it in `invert(scalar(...))`. `until () { }` is a syntax error.

**`continue` blocks.** Both accept a trailing `continue BLOCK`
(`perly.y:972-976`), which runs after each iteration including after a `next`.
This is the only way to get C's `for(;;i++)` semantics with `next` in a `while`
loop.

The lexer disambiguates `continue` at `toke.c:8385-8395`:

```c
    case KEY_continue:
        /* We have to disambiguate the two senses of
          "continue". If the next token is a '{' then
          treat it as the start of a continue block;
          otherwise treat it as a control operator.
         */
        s = skipspace(s);
        if (*s == '{')
            PREBLOCK(KW_CONTINUE);
        else
            FUN0(OP_CONTINUE);
```

The bare `continue` operator (5.10+, `given`/`when` machinery) falls through to
the current `when` block. In your parser: after seeing `continue` in statement
or block-tail position, skip whitespace and comments and peek for `{`.

### 5.3.4 C-style `for`

```ebnf
CForStmt = "for" "(" [ Expr ] ";" [ Expr ] ";" [ Expr ] ")" Block ;
```

`perly.y:347-378`. The three clauses are `mnexpr`, `texpr`, `mnexpr` — init and
increment may be empty (`nexpr` → `empty | sideff`, `perly.y:992-995`), and the
test may be empty meaning true.

`for (;;) { }` is the idiomatic infinite loop.

**A C-style `for` never takes a `continue` block.** Look at the production: no
`cont`. Only the foreach forms have one. This is easy to get wrong.

The grammar sets `parser->expect = XTERM` after each semicolon
(`perly.y:353-355`, `perly.y:358-360`), overriding the `XSTATE` the lexer set
at the `;` (chapter 3 §3.1.2). In your parser this is the reminder
that the `/` after a `for` semicolon is a regex, not division.

**Disambiguating C-style from foreach.** The lexer does this at
`toke.c:7433-7507` (`yyl_foreach`), and it does it by *scanning ahead in the raw
source*, not by parsing:

```c
    if (PL_expect == XSTATE && isIDFIRST_lazy_if_safe(s, PL_bufend, UTF)) {
        ...
        if (LIKELY(memBEGINPs(p, (STRLEN) (PL_bufend - p), "my"))) {
            core_valid = TRUE;
            paren_is_valid = TRUE;
```

The algorithm, which you should copy:

1. Skip whitespace after `for`/`foreach`.
2. If the next thing is an identifier, check for `CORE::` prefix, then for
   `my` / `our` / `state`.
3. After `my`, optionally skip a package name (`for my Dog $x (...)` — the old
   typed-lexical syntax; `toke.c:7485-7494`).
4. If we saw `my` and the next char is `(`, this is the 5.36 multi-var foreach
   (`toke.c:7496-7498`).
5. Otherwise the next char must be `$` or `\`; if not, croak `Missing $ on loop
   variable` (`toke.c:7499-7502`).
6. If the char right after `for (` is not an identifier at all, it is either
   C-style or `for (LIST)`.

For the C-style/`for (LIST)` split you need one more decision, which the
grammar makes by lookahead over the parenthesised group: **scan to the matching
`)` and check whether the group contains two top-level `;`.** If yes, C-style;
if no, foreach over a list. You must respect nesting and quoting during that
scan, which is why this is a scanner-level job, not a token-level one.

```go
// Returns true if the parenthesised group starting at the current `(` contains
// exactly two semicolons at paren-depth 1. Must skip strings, regexes,
// heredocs, comments, and nested brackets.
func (p *Parser) forIsCStyle() bool
```

### 5.3.5 `foreach` — all five forms

`perly.y:379-485` gives six alternatives beyond the C-style one. Grouped:

```ebnf
ForeachStmt =
    "for" "my" ScalarVar "(" Expr ")" Block [ "continue" Block ]
  | "for" "my" "(" ScalarList ")" "(" Expr ")" Block [ "continue" Block ]
  | "for" ScalarVar "(" Expr ")" Block [ "continue" Block ]
  | "for" MyRefgen MyVar "(" Expr ")" Block [ "continue" Block ]
  | "for" "\" RefgenTopic "(" Expr ")" Block [ "continue" Block ]
  | "for" "(" Expr ")" Block [ "continue" Block ] ;
```

**Form 1 — lexical loop variable** (`perly.y:379-391`):

```perl
for my $x (@list) { ... }
```

`my_scalar` (`perly.y:1803-1805`) introduces `$x` into the pad. The variable is
scoped to the loop and is an *alias* to each element — assigning to `$x`
modifies `@list`.

**Form 2 — multi-var foreach, new in 5.36** (`perly.y:392-410`):

```perl
for my ($k, $v) (%hash) { ... }
for my ($a, $b, $c) (@triples) { ... }
```

The list is `my_list_of_itervars` → `list_of_itervars` (`perly.y:1808-1827`),
which accepts scalars and `\`-prefixed refaliases separated by commas, with a
permitted trailing comma. Each iteration consumes `n` elements from the list,
where `n` is the number of variables. If the list length is not a multiple of
`n`, the trailing variables are `undef` on the final pass.

The grammar has a special case worth reproducing (`perly.y:404-407`):

```c
			if ($my_list_of_itervars->op_type == OP_PADSV)
				/* degenerate case of 1 var: for my ($x) ....
				   Flag it so it can be special-cased in newFOROP */
				$my_list_of_itervars->op_flags |= OPf_PARENS;
```

`for my ($x) (@list)` is *not* the same as `for my $x (@list)`: the
parenthesised single-variable form takes one element per pass but does **not**
alias, whereas the bare form does. Preserve the parens in your AST.

**Form 3 — package/global loop variable with implicit localisation**
(`perly.y:411-422`):

```perl
for $x (@list) { ... }
```

If `$x` is a package variable, it is implicitly `local`ised for the duration of
the loop and restored afterwards — note `op_lvalue($scalar, OP_ENTERLOOP)`. If
`$x` is an already-declared lexical, it is used directly and *not* localised
(there is nothing to localise). Your parser cannot distinguish these two
without scope information, so record the syntax and let the analyser decide.
This is one of the places `ParseState.Pad` earns its keep.

**Form 4 and 5 — refaliasing** (`perly.y:423-473`):

```perl
for \my $x (@refs) { ... }     # my_refgen my_var
for my \$x (@refs) { ... }     # my_refgen is `my \` OR `\ my`
for \$x (@refs)    { ... }     # REFGEN refgen_topic
```

`my_refgen` (`perly.y:1838-1840`) accepts `my \` and `\ my` in either order.
`refgen_topic` (`perly.y:1834-1836`) is a `my_var` (scalar, array, or hash) or
an `amper` (`&foo`). This is the `refaliasing` feature; the loop variable
becomes an alias to the *referent* of each element.

**Form 6 — implicit `$_`** (`perly.y:474-484`):

```perl
for (@list) { print; }
```

No variable, so `newFOROP(0, NULL, ...)`. `$_` is aliased and **implicitly
localised** — the outer `$_` is saved and restored. Semantically important, and
worth a distinct AST shape so tooling can tell users what `$_` refers to.

```go
type ForeachStmt struct {
    // Exactly one of these describes the loop variable.
    Var      *VarExpr   // form 1 (with Decl != nil) or form 3
    Vars     []*VarExpr // form 2, multi-var; len>=1
    Parens   bool       // form 2 syntax used (matters even for len==1)
    Decl     DeclKind   // DeclNone | DeclMy | DeclOur | DeclState
    Refalias bool       // forms 4 and 5
    // Var == nil && Vars == nil  =>  form 6, implicit $_

    List     Expr
    Body     *Block
    Continue *Block // nil if absent
    Span     Span
}
```

**`for` and `foreach` are exactly synonymous.** `toke.c:8552-8554` routes both
`KEY_for` and `KEY_foreach` to `yyl_foreach`. Record which spelling was used for
formatting fidelity; treat them identically otherwise.

### 5.3.6 `do BLOCK` and why `do BLOCK while` is not a loop

`perly.y:1557-1562`:

```
termdo	:       KW_DO term	%prec UNIOP                     /* do $filename */
	|	KW_DO block	%prec PERLY_PAREN_OPEN               /* do { code */
			{ $$ = newUNOP(OP_NULL, OPf_SPECIAL, op_scope($block));}
```

`termdo` is reachable from `term` (`perly.y:1567`), **not** from `barestmt`.
That single fact explains everything:

* `do BLOCK` is an **expression**. Its value is the last statement's value.
* `do { ... } while (COND);` is therefore an *expression statement with a
  `while` statement modifier* (§5.4), parsed through
  `sideff: expr KW_WHILE condition` at `perly.y:943-944`.
* Because it is not a loop op, `last`, `next`, and `redo` **do not work** in it.
  Perl warns `Exiting subroutine via last` or simply does the wrong thing.
* Perl special-cases it so the body runs at least once, but that is a runtime
  detail of `newLOOPOP` with `OPf_PARENS`.

The standard workaround, which your diagnostics should know about:

```perl
LOOP: {
    do {
        ...
        last LOOP if $done;     # works: bare block IS a loop
    } while ($cond);
}
```

Also note `toke.c:7509-7536` (`yyl_do`): `do` followed by `{` is
`PRETERMBLOCK(KW_DO)`; `do` followed by a bareword that is not a keyword and is
followed by `(` becomes `do &subname(...)` — a deprecated way to call a sub.
`do EXPR` where EXPR is a filename does a runtime `require`-like file
inclusion.

### 5.3.7 Loop control: `last`, `next`, `redo`, `goto`, `dump`

`perly.y:1661-1665`:

```
	|	LOOPEX  /* loop exiting command (goto, last, dump, etc) */
			{ $$ = newOP($LOOPEX, OPf_SPECIAL);
			    PL_hints |= HINT_BLOCK_SCOPE; }
	|	LOOPEX term[operand]
```

`LOOPEX` is declared `%nonassoc` at the very bottom of the precedence table
(`perly.y:151`), one step above `PREC_LOW`. It reaches the tree through `term`,
so **loop control is an expression, not a statement form.** `last if $done;`
parses as an expression statement with an `if` modifier.

```ebnf
LoopEx = ("last" | "next" | "redo") [ Label | Expr ]
       | "goto" ( Label | "&" SubName | Expr )
       | "dump" [ Label ] ;
```

The operand is a `term`, which means the label is parsed as a bareword and only
*later* interpreted as a label. `last $x` is legal and computes the label name
at runtime. For an LSP, this means "go to definition" on a loop label works
only when the operand is a literal bareword — which is the 99% case, so
special-case it:

```go
type LoopCtlExpr struct {
    Kind  LoopCtlKind // Last, Next, Redo, Goto, Dump
    Label string      // set only if operand was a literal bareword
    Expr  Expr        // set otherwise; nil if no operand
    Span  Span
}
```

Because `LOOPEX` binds so loosely, `last unless $ok` works but `last + 1` parses
as `last(+1)`. That is intentional and matches list operators.

### 5.3.8 `given` / `when` / `default`

Deprecated since 5.38, removed from the default feature set, but still present
in the grammar and still exercised by the core test suite. Parse it.

```ebnf
GivenStmt   = "given" "(" Expr ")" Block ;
WhenStmt    = "when" "(" Expr ")" Block ;
DefaultStmt = "default" Block ;
```

`perly.y:503-514`, `perly.y:731-741`, `perly.y:316-322`.

`when` also exists as a statement modifier (`perly.y:950-951`):

```
	|	expr[body] KW_WHEN condition
			{ $$ = newWHENOP($condition, op_scope($body)); }
```

Both require `use feature 'switch'`, which is *not* in any bundle from `:5.36`
onwards (see the bundle table in `lib/feature.pm:843-870`). Your parser should
parse them regardless and let the analyser flag the missing feature — a parser
that refuses to parse deprecated syntax is useless for a language server
looking at old code.

`when` inside a `for` loop is legal and uses the loop's `$_` as the topic;
`given` merely sets `$_` for its block.

### 5.3.9 `defer`

`perly.y:324-329`:

```
bare_statement_defer
	:	KW_DEFER mblock
```

```ebnf
DeferStmt = "defer" Block ;
```

Requires `use feature 'defer'` (experimental, not in any bundle). The block runs
when the enclosing block is left, by any route including `die`. Deferred blocks
run in reverse order of registration. Note it takes an `mblock`, sharing the
enclosing scope's `my` introduction.

### 5.3.10 Exception handling

**`eval BLOCK` and `eval EXPR`.** These do not appear as statement forms.
`eval` is a `UNIOP`, reaching the tree through `perly.y:1668-1673`:

```
	|	UNIOP                                /* Unary op, $_ implied */
	|	UNIOP block                          /* eval { foo }* */
	|	UNIOP term[operand]                           /* Unary op */
```

So `eval { ... }` is an expression whose operand is a block; `eval "..."` is an
expression whose operand is a term. `eval { }` is compile-time-checked; `eval
""` is a runtime compile. Both set `$@`. Both need a trailing `;` when used as
statements, unlike `if`/`while`/`sub`, and forgetting it is a classic bug:

```perl
eval { risky() }    # missing semicolon
or die $@;          # ... actually fine, this parses as one expression
```

**`try` / `catch` / `finally`.** `perly.y:661-681`:

```ebnf
TryStmt = "try" Block "catch" "(" ScalarVar ")" Block [ "finally" Block ] ;
```

Points to note:

* `catch` is **mandatory**. There is no `try { } finally { }` without a catch.
* The `(VAR)` on `catch` is mandatory, and the grammar goes out of its way to
  give a good error when it is missing (`perly.y:812-821` plus the check at
  `perly.y:667-671`):

  ```
  catch_paren:	empty
  			/* not really valid grammar but we detect it in the
  			 * action block to throw a nicer error message */
  ```

  Copy this technique. Accept the malformed form in the grammar and diagnose it
  in the action, so that recovery is clean and the message is specific.
* The catch variable is introduced with `parser->in_my = KEY_catch`
  (`perly.y:816`), making it a lexical scoped to the catch block.
* `finally` is optional (`perly.y:979-983`). Before 5.40 a `try` without
  `finally` warned as experimental; as of 5.40 it does not
  (`lib/feature.pm:703-709`).
* `try` is in the `:5.40` bundle and later; before that it needs `use feature
  'try'`.

### 5.3.11 Loop and conditional summary table

| Construct | Braces required | Parens required | `continue` block | Is a loop? |
| --- | --- | --- | --- | --- |
| `if` / `unless` | yes | yes | no | no |
| `while` / `until` | yes | yes | yes | yes |
| `for (;;)` | yes | yes | **no** | yes |
| `foreach` (all forms) | yes | yes | yes | yes |
| bare `{ }` | — | — | yes | **yes** (once) |
| `do { }` | yes | — | no | **no** |
| `given` / `when` | yes | yes | no | no |
| `try` / `catch` | yes | yes on catch | no | no |
| `defer` | yes | — | no | no |

---

## 5.4 Statement modifiers

```ebnf
StmtModifier = ("if" | "unless" | "while" | "until" | "for" | "foreach" | "when") Expr ;
```

`perly.y:933-952` is the whole story:

```
sideff	:	error
			{ $$ = NULL; }
	|	expr[body]
	|	expr[body] KW_IF condition
	|	expr[body] KW_UNLESS condition
	|	expr[body] KW_WHILE condition
	|	expr[body] KW_UNTIL iexpr
	|	expr[body] KW_FOR condition
	|	expr[body] KW_WHEN condition
	;
```

### 5.4.1 They cannot be stacked

Look at the production: the left-hand side of each alternative is `expr`, and
the whole alternative reduces to `sideff`, which is **not** `expr`. There is no
path from `sideff` back into `expr`. Therefore:

```perl
print "x" if $a for @list;   # SYNTAX ERROR
```

You cannot chain modifiers. The workaround is a `do BLOCK`:

```perl
do { print "x" if $a } for @list;   # legal
```

Your parser must reject the stacked form with a clear message. The natural
implementation: parse the modifier, then if the next token is another modifier
keyword rather than `;`, emit "statement modifiers cannot be stacked" and
recover to the semicolon.

### 5.4.2 Precedence

The modifier binds looser than **everything** in the expression. Its left
operand is the entire `expr`, including low-precedence `and`/`or`/`not`
(`perly.y:1254-1263`). So:

```perl
$a = 1 or die  if $x;    # ($a = 1 or die)  if $x
```

In recursive descent this falls out for free: parse a full expression at the
lowest precedence level, then check for a modifier keyword.

```go
func (p *Parser) parseExprStatement() Stmt {
    e := p.parseExpr(PrecLowest)   // includes `and`, `or`, `not`
    if mod, ok := p.tryStmtModifier(); ok {
        st := &ModifiedStmt{Body: e, Mod: mod}
        if p.atStmtModifierKeyword() {
            p.errorf("statement modifiers cannot be stacked")
            p.recoverToStatementEnd()
        }
        p.expectSemicolonOrBlockEnd()
        return st
    }
    p.expectSemicolonOrBlockEnd()
    return &ExprStmt{X: e}
}
```

### 5.4.3 Semantics that differ from the block forms

* **`for`/`foreach` modifier**: `newFOROP(0, NULL, $condition, $body, NULL)`
  (`perly.y:947-949`) — the iterator is always `$_`, always implicitly
  localised. You cannot name a variable: `print for my $x (@l)` is a syntax
  error.
* **`while`/`until` modifier**: `newLOOPOP(OPf_PARENS, 1, ...)`
  (`perly.y:943-946`). The `OPf_PARENS` flag is what makes `do BLOCK while`
  run at least once. For any other expression the condition is tested first.
* **`if`/`unless` modifier**: compiles to `newLOGOP(OP_AND, ...)` /
  `newLOGOP(OP_OR, ...)`. Semantically `EXPR if COND` really is `COND and
  EXPR`.
* **`until` uses `iexpr`**, so the condition is inverted at parse time
  (`perly.y:1006-1008`).
* **A modifier statement cannot declare a `my` visible afterwards.**
  `my $x = 1 if $cond;` is undefined behaviour — Perl documents it as such.
  Worth a lint diagnostic.

### 5.4.4 The `error` alternative

`sideff: error { $$ = NULL; }` at `perly.y:933-934` is the *only* error
production in the entire grammar. This is Perl's whole recovery story: when a
statement fails to parse, discard it and resynchronise at the next `;`. Section
5.13 explains why you should do considerably better than this.

---

## 5.5 Subroutines

### 5.5.1 The four sub token types

Repeating §5.0 because it is the crux: `toke.c:5957-5975` chooses among four
tokens based on (a) whether a name followed and (b) whether signatures are
enabled:

```c
    if (!have_name) {
        ...
        if (is_method)
            TOKEN(KW_METHOD_anon);
        else if (is_sigsub)
            TOKEN(KW_SUB_anon_sig);
        else
            TOKEN(KW_SUB_anon);
    }
    force_ident_maybe_lex('&');
    if (is_method)
        TOKEN(KW_METHOD_named);
    else if (is_sigsub)
        TOKEN(KW_SUB_named_sig);
    else
        TOKEN(KW_SUB_named);
```

You do not need four token types. You need one `sub` token and a
`p.state.Features.Has(FeatSignatures)` check at the point where you decide
whether `(` after the name introduces a prototype or a signature.

### 5.5.2 Named subs, traditional form

`perly.y:636-659`:

```ebnf
SubDecl = "sub" SubName [ Prototype ] { Attribute } ( Block | ";" ) ;
```

* `subname` is `BAREWORD | PRIVATEREF` (`perly.y:1045-1047`) — the `PRIVATEREF`
  case is a lexical sub (§5.5.7).
* `proto` is `empty | PROTOTYPE` (`perly.y:1050-1053`); the lexer produces the
  `PROTOTYPE` token by calling `scan_str` at `toke.c:5923-5933`, i.e. it scans a
  balanced `(...)` as an opaque string.
* `optsubbody` is `subbody | ";"` (`perly.y:1217-1220`) — the `;` case is a
  **forward declaration**.

### 5.5.3 Named subs, signature form

`perly.y:604-634`:

```ebnf
SigSubDecl = ("sub" | "method") SubName { Attribute } ( SigBlock | ";" ) ;
SigBlock   = [ Signature ] Block ;
```

Note the ordering difference from the traditional form. In the traditional
form, `proto` comes **before** `subattrlist`; in the signature form there is no
prototype slot at all, and `subattrlist` comes before the body. The signature
itself lives *inside* `sigsubbody` (`perly.y:1241-1250`):

```
sigsubbody:	remember optsubsignature PERLY_BRACE_OPEN
			{ PL_parser->sig_seen = FALSE; }
		stmtseq PERLY_BRACE_CLOSE
```

Which means the source order is:

```perl
sub name :attr ($x, $y) { ... }
#        ^attrs ^signature
```

**Attributes come before the signature.** Writing them the other way round —
`sub name ($x) :attr { }` — is an error, and the grammar sets
`parser->expect = XATTRBLOCK` after a signature specifically so that the lexer
can scan the misplaced attributes and produce a helpful message rather than
collapsing (`perly.y:1202-1211`):

```
                            /* tell the toker that attrributes can follow
                             * this sig, but only so that the toker
                             * can skip through any (illegal) trailing
                             * attribute text then give a useful error
                             * message about "attributes before sig",
```

Do the same. It is exactly the kind of error a developer makes constantly, and
a language server that recovers from it keeps the rest of the file analysable.

### 5.5.4 Anonymous subs

`perly.y:1533-1555`, under `anonymous` (which also covers `[...]` and `{...}`
constructors):

```ebnf
AnonSub = "sub" [ Prototype ] { Attribute } Block          (* no signatures *)
        | "sub" { Attribute } [ Signature ] Block           (* signatures on *)
        | "method" { Attribute } [ Signature ] Block ;
```

The grammar deliberately includes the *bodyless* alternatives purely to produce
a good error (`perly.y:1541-1542`, `1546-1547`, `1553-1554`):

```
	|	KW_SUB_anon     startanonsub proto subattrlist            %prec PERLY_PAREN_OPEN
			{ yyerror("Illegal declaration of anonymous subroutine"); YYERROR; }
```

Third instance of the same technique — accept the wrong thing, diagnose in the
action. It is the dominant error-handling idiom in `perly.y` and it should be
the dominant idiom in your parser too.

`sub` in expression position always means an anonymous sub. `sub` in statement
position followed by an identifier means a named sub; followed by `{` it is
still an anonymous sub, now in void context (and `Useless use of anonymous
subroutine` fires).

### 5.5.5 Prototypes

A prototype is a parenthesised string that changes how *calls to the sub are
parsed* — but only for calls that are (a) to an already-declared sub and (b)
without an `&` sigil. That conditional is why prototypes are widely disliked
and why your parser must handle them: `sub mymax(\@) {...}` changes the parse of
`mymax @list` downstream.

The lexer treats a prototype as an opaque scanned string (`toke.c:5923-5929`),
then validates it (`validate_proto`) and pushes it as a `PROTOTYPE` token. The
grammar never looks inside it.

Prototype characters, for the record:

| Char | Meaning |
| --- | --- |
| `$` | scalar; imposes scalar context |
| `@` | slurpy list; consumes the rest |
| `%` | slurpy hash; consumes the rest |
| `&` | code ref; as *first* param allows `func { ... } @args` |
| `*` | typeglob/filehandle slot; a bareword arrives as the plain string of its name, not a glob (chapter 3 §3.5.2) |
| `\X` | reference to X; `\@` takes `@list` and passes `\@list` |
| `\[$@%&*]` | reference to any of the listed types |
| `;` | subsequent params are optional |
| `+` | scalar-or-reference; `\[@%]` if array/hash, else `$` |
| `_` | like `$` but defaults to `$_` |
| (empty) | takes no args; the sub is inlineable as a constant |

The `&` first-position case shows up in the grammar at `perly.y:1330-1340`
(`LSTOPSUB startanonsub block optlistexpr`) — this is how `first { $_ > 3 }
@list` parses without a comma.

`BLKLSTOP` at `perly.y:1281-1284` is the newer builtin equivalent (`all { ... }
@args`).

**Prototypes and signatures are mutually exclusive per sub.** The grammar
enforces this structurally — `bare_statement_sub_traditional` has a `proto`
slot and no signature; `bare_statement_sub_signature` has a signature and no
`proto` slot. To attach a prototype to a signatured sub you use the `:prototype`
attribute:

```perl
use feature 'signatures';
sub mymax :prototype(\@) ($aref) { ... }
```

### 5.5.6 Attributes

`perly.y:1056-1079`:

```ebnf
AttrList = ":" [ Attr { [","] Attr } ] ;
Attr     = IDENT [ "(" balanced-text ")" ] ;
```

```
subattrlist
	:	empty
	|	COLONATTR ATTRLIST
	|	COLONATTR
```

The lexer produces `COLONATTR` for the introducing colon and `ATTRLIST` for the
attribute text; the argument in parens is scanned as balanced text and is **not
parsed as Perl**. `:Foo(bar baz => {})` is fine; the contents are handed to the
attribute handler as a string.

Built-in sub attributes:

* `:lvalue` — the sub may be assigned to.
* `:method` — the sub is a method; suppresses `Ambiguous call` warnings and
  affects `AUTOLOAD` behaviour. Not the same as the `method` keyword.
* `:prototype(...)` — attach a prototype to a signatured sub.
* `:const` — for anon subs; evaluate the body once at closure-creation time.

`perly.y:1061-1062` shows the interaction with signatures:

```c
			  if(attrlist && !PL_parser->sig_seen)
			      attrlist = apply_builtin_cv_attributes(PL_compcv, attrlist);
```

Built-in attributes are applied only when no signature has been seen —
reinforcing that attributes must precede the signature.

Attributes also apply to variables (`my $x :shared`, `perly.y:1719-1724`) and to
classes and fields (§5.7).

Separator note: attributes may be separated by whitespace *or* commas.
`:lvalue :method` and `:lvalue,method` both work. Handle both.

### 5.5.7 Lexical subs

```perl
my sub helper { ... }
state sub counter { ... }
our sub exported { ... }
```

The grammar sees these through `subname: PRIVATEREF` (`perly.y:1046`) and
dispatches to `newMYSUB` instead of `newATTRSUB` (`perly.y:626-629`):

```c
			$subname->op_type == OP_CONST
				? newATTRSUB($startsub, $subname, NULL, $subattrlist, body)
				: newMYSUB(  $startsub, $subname, NULL, $subattrlist, body)
```

So the discriminator is: **is the name already in the pad?** `toke.c:5884-5893`
does that lookup:

```c
        if (memchr(tmpbuf, ':', len) || key != KEY_sub
         || pad_findmy_pvn(
                PL_tokenbuf, len + 1, 0
            ) != NOT_IN_PAD)
            sv_setpvn(PL_subname, tmpbuf, len);
        else {
            sv_setsv(PL_subname,PL_curstname);
            sv_catpvs(PL_subname,"::");
            sv_catpvn(PL_subname,tmpbuf,len);
        }
```

A name containing `::` is never a lexical sub. A name found in the pad is.

Visibility is the subtle part, and it is the opposite of what most people
expect: **a lexical sub is not visible inside its own body.** The name is
introduced into the pad only once the declaration statement completes, so a
naive recursive definition fails at runtime. Verified against perl v5.42.0:

```perl
my sub fact { my $n = shift; $n <= 1 ? 1 : $n * fact($n-1) }
say fact(5);
# Undefined subroutine &main::fact called at ... line 1.
```

`state sub` behaves identically. The documented idiom is to forward-declare, so
the name is in the pad before the body is parsed:

```perl
my sub fact;                                     # declaration: name enters pad
sub fact { my $n = shift; $n <= 1 ? 1 : $n * fact($n-1) }
say fact(5);                                     # 120
```

PerlOnJava models this correctly and explicitly, adding `&subName` to the symbol
table *after* the body is parsed to make the sub "invisible inside itself"
(`StatementResolver.java:411-413, 500-505`).

For your parser this means the pad update for a lexical sub happens **at the end
of the declaration**, not at the name. Get the order wrong and you will resolve
`fact` inside the body to the lexical rather than reporting it unresolved.

Two further wrinkles PerlOnJava documents. `our sub` creates a package sub *plus*
a lexical alias carrying the fully-qualified name, so it resolves across
subsequent `package` switches (`StatementResolver.java:370-390`). And `state
sub`, plus `my sub` at file scope, has its assignment executed at compile time
via a synthetic `BEGIN`, so that `use overload => \&foo` can see it (`:513-537`).

`toke.c:5901-5906` is the error path: `my sub` with no name croaks `Missing
name in "my sub"`.

For your parser: track declared lexical subs in `ParseState.Pad` alongside
lexical variables, because a bareword matching a lexical sub name parses as a
call to it rather than as a string.

### 5.5.8 `AUTOLOAD` and `DESTROY`

`toke.c:8313-8322`:

```c
    case KEY_AUTOLOAD:
    case KEY_DESTROY:
    case KEY_BEGIN:
    case KEY_UNITCHECK:
    case KEY_CHECK:
    case KEY_INIT:
    case KEY_END:
        if (PL_expect == XSTATE)
            return yyl_sub(aTHX_ PL_bufptr, key);
        return yyl_just_a_word(aTHX_ s, len, orig_keyword, c);
```

`AUTOLOAD` and `DESTROY` are grouped with the phasers in the lexer, but they are
**ordinary subs**, not phasers. They route through `yyl_sub` only so that `sub`
may be omitted:

```perl
AUTOLOAD { ... }     # same as sub AUTOLOAD { ... }
DESTROY  { ... }     # same as sub DESTROY  { ... }
```

The `PL_expect == XSTATE` guard means `$h{AUTOLOAD}` still parses as a hash key.
Reproduce that guard — it is why `Expect` must be in your parse state.

Semantically: `AUTOLOAD` is called for undefined methods/subs, with the full
name in `our $AUTOLOAD`. `DESTROY` is the destructor. Neither is special to the
parser beyond the optional `sub`.

### 5.5.9 Recommended sub AST

```go
type SubDecl struct {
    Kind      SubKind    // SubNamed, SubAnon, SubMethod, SubLexical
    Name      string     // "" for anonymous
    DeclScope DeclKind   // DeclNone | DeclMy | DeclState | DeclOur
    Proto     *Prototype // nil unless traditional form with a prototype
    Attrs     []Attr
    Sig       *Signature // nil if no signature (or traditional form)
    Body      *Block     // nil for a forward declaration
    Span      Span
    NameSpan  Span       // for rename / go-to-definition
}

type Prototype struct {
    Text string // raw, between the parens; not decomposed
    Span Span
}

type Attr struct {
    Name string
    Args string // raw balanced text, or "" if no parens
    Span Span
}
```

`Body == nil` distinguishes a forward declaration. Do not collapse it with an
empty block — `sub f;` and `sub f {}` mean different things.

---

## 5.6 Signatures

The definitive grammar is `perly.y:1087-1214`. Read it top to bottom; it is
compact and complete.

```ebnf
Signature   = "(" [ SigList ] ")" ;
SigList     = SigElem { "," [ SigElem ] } ;    (* trailing comma allowed *)
SigElem     = SigScalar | SigSlurpy ;
SigScalar   = [ ":" ] "$" [ IDENT ] [ AssignOp SigDefault ] ;
SigSlurpy   = ("@" | "%") [ IDENT ] ;
SigDefault  = ε | Term ;
AssignOp    = "=" | "//=" | "||=" ;
```

### 5.6.1 Positional parameters

```perl
sub f ($x, $y) { ... }
```

`sigscalarelem` at `perly.y:1122-1132` calls `subsignature_append_positional`.

### 5.6.2 Optional parameters and defaults

```perl
sub f ($x, $y = 42) { ... }
sub f ($x, $y //= compute()) { ... }
sub f ($x, $y ||= 0) { ... }
sub f ($x, $y =) { ... }        # legal! default is "no default"
```

`perly.y:1133-1142` captures the `ASSIGNOP` value, and `optsigscalardefault`
(`perly.y:1145-1149`) allows the default to be **empty**:

```
optsigscalardefault:
                %empty
                        { $$ = newOP(OP_NULL, 0); }
        |       term
```

This is a case where the grammar is more permissive than the language. `sub f
($x =) {}` reduces successfully, but a later check rejects it — verified against
v5.42.0:

```
$ perl -e 'use v5.36; sub f($x=){}'
Optional parameter lacks default expression at -e line 1, near "=)"
```

Write `$x = undef` if that is what you mean. **The lesson generalises: `perly.y`
alone is not the specification.** Several constructs reduce in the grammar and
are then rejected by an action, a `croak` in `toke.c`, or a check in `op.c`.
Where this chapter states that something parses, it has been checked against a
real perl; where you extend it, do the same.

The three assignment operators differ in *when* the default applies:

| Op | Default applies when | Test performed |
| --- | --- | --- |
| `=` | the argument was not passed at all | **arity**: `@_ < n` |
| `//=` | not passed, **or** passed as `undef` | **value**: `$x // default` |
| `||=` | not passed, **or** passed as false | **value**: `$x || default` |

The arity-versus-value distinction is real and observable: with `=`, passing an
explicit `undef` keeps the `undef`; with `//=` it is replaced. PerlOnJava's
codegen makes it explicit — `generateDefaultAssignment`
(`SignatureParser.java:311-336`) emits `@_ < (paramIndex+1) && ($var = default)`
for `=`, and a plain `BinaryOperatorNode(op, variable, defaultValue)` for the
other two. **Record which operator was written; do not normalise it away.**

The default expression is a full `term`, evaluated at call time, in a scope
where **earlier parameters are already bound**. So this works:

```perl
sub rect ($w, $h = $w) { ... }     # square by default
```

Which means your parser must introduce each parameter into the pad *as it is
parsed*, not all at the end. `parser->in_my = KEY_sigvar` at `perly.y:1154` and
`perly.y:1188` is doing exactly this.

### 5.6.3 Slurpy parameters

```perl
sub f ($x, @rest)  { ... }
sub f ($x, %opts)  { ... }
sub f ($x, @)      { ... }   # slurpy but unnamed: "accept and discard"
sub f ($x, %)      { ... }
```

`sigslurpelem` (`perly.y:1100-1113`) via `sigslurpsigil` (`perly.y:1093-1097`)
and `sigvar` (`perly.y:1087-1091`) — `sigvar` may be `%empty`, which is how the
nameless `@`/`%` forms work. Nameless positional `$` works the same way: `sub f
($, $, $x) {...}` skips the first two arguments.

**A slurpy parameter may not have a default.** The grammar contains explicit
error productions (`perly.y:1105-1112`):

```
        |     sigslurpsigil sigvar ASSIGNOP
                        {
			    yyerror("A slurpy parameter may not have a default value");
                        }
        |     sigslurpsigil sigvar ASSIGNOP term
                        {
			    yyerror("A slurpy parameter may not have a default value");
                        }
```

Fourth instance of the accept-then-diagnose idiom. Note there are **two**
productions, one with a term and one without, so that both `(@r = )` and `(@r =
1)` get the good message rather than a generic syntax error. Copy this level of
care.

Constraints the grammar does not encode but `subsignature_append_slurpy`
enforces:

* At most one slurpy parameter.
* It must be last.
* A mandatory (non-defaulted) parameter may not follow an optional one.

Enforce these in your parser after collecting the list — they are simple
post-checks and they produce far better messages than a grammar-level rejection.
`SignatureParser.java` is the best worked example: it catches "param after
slurpy" twice, once at the top of the *next* parameter
(`SignatureParser.java:149-151`) and once eagerly inside the slurpy handler
(`:249-259`), where it distinguishes "Multiple slurpy parameters not allowed"
(next token is `@`/`%`) from "Slurpy parameter not last". It also rejects `$#`
with "'#' not allowed immediately following a sigil in a subroutine signature"
(`:211`), `$_` as a parameter name (`:175-177`), and `$b += 1` with "Illegal
operator following parameter in a subroutine signature" (`:186-192`).

### 5.6.4 Named parameters

The tree at hand supports `:$name` syntax (`perly.y:1123-1142`, the `optcolon`
productions and `subsignature_append_named`):

```perl
sub f ($x, :$verbose = 0, :$debug = 0) { ... }
f(1, verbose => 1);
```

`optcolon` is `perly.y:1115-1119`. This is very new — check
`FEATURE_*_IS_ENABLED` gating in the tree you target before relying on it. Parse
it, and record it, but flag it as version-gated.

Two semantic wrinkles, from PerlOnJava's implementation
(`SignatureParser.java:379-424`). Named parameters desugar to a shared
`my %__named_args__ = @_;` plus per-parameter `delete`, which means the
**maximum-arity check is not emitted** when named parameters are present
(`:475-488`) — key/value pairs are unbounded. And the default-operator mapping
changes: for named parameters `=` and `//=` **both** become `//` (`:418-424`),
so the arity-versus-value distinction of §5.6.2 does not apply to them.

Also note `"Named parameters cannot be slurpy"` (`:157-159`) and `"Named
parameters must actually have a name"` (`:180-182`) — `:$` is not legal.

### 5.6.5 Signature is a lexical scope wrinkle

`subsigguts` (`perly.y:1184-1213`) wraps the whole signature in `ENTER`/`LEAVE`
and sets `parser->sig_seen = TRUE` at the end. The generated code is prepended
to the sub body (`perly.y:1247-1248`):

```c
			  $$ = block_end($remember,
				op_append_list(OP_LINESEQ, $optsubsignature, $stmtseq));
```

So a signature is *not* a separate scope from the body — parameters are ordinary
lexicals of the body's scope. `my $x` inside a body when `$x` is also a
parameter is a `"my" variable $x masks earlier declaration` warning.

### 5.6.6 Recommended signature AST

```go
type Signature struct {
    Params []Param
    Span   Span
}

type Param struct {
    Sigil   byte   // '$', '@', '%'
    Name    string // "" for the nameless placeholder forms
    Named   bool   // `:$x` form
    Default Expr   // nil if none
    DefOp   string // "=", "//=", "||="; "" if no default
    Span    Span
}

// Post-parse validation, not grammar:
//   - at most one slurpy, and it must be last
//   - no mandatory param after an optional one
//   - slurpy may not have a default (diagnose, don't reject the parse)
func (s *Signature) Validate(d *Diagnostics)
```

Compute arity from this and expose it — `min` is the count of params before the
first default, `max` is `len(Params)` or unbounded if slurpy. That is what
signature help and call-arity diagnostics need.

---

## 5.7 Packages and classes

### 5.7.1 `package`

```ebnf
PackageStmt = "package" IDENT [ Version ] ";"
            | "package" IDENT [ Version ] Block ;
```

`perly.y:539-553` and `perly.y:555-571`. There is a comment there you must read
before writing your parser (`perly.y:544-548`):

```
		/* version and package appear in the reverse order to what may be
		 * expected, because toke.c has already pushed both of them to a stack
		 * by calling force_next() from within force_version().
		 * When the parser pops them back out again they appear swapped
		 */
```

This is an artefact of the token-pushing mechanism, **not** of the surface
syntax. Source order is `package NAME VERSION`. In a recursive-descent parser
you read them in source order and this whole wrinkle vanishes. Mentioned here
only so you are not confused when reading `perly.y`.

The lexer side is `toke.c:8858-8862`:

```c
    case KEY_package:
        s = force_word(s, BAREWORD, ALLOW_PACKAGE);
        s = skipspace(s);
        s = force_strict_version(s);
        PREBLOCK(KW_PACKAGE);
```

`force_strict_version` means the version must be a strict v-string or numeric
literal (`1.234`, `v1.2.3`), not an arbitrary expression.

Semantics:

* `package NAME;` — sets the current package until the end of the enclosing
  block or file. **Multiple packages per file are normal.**
* `package NAME BLOCK` — sets the current package for the block only. Preferred
  in modern code because the scope is explicit.
* `package NAME VERSION;` — also sets `$NAME::VERSION`.
* Package names may contain `::` (and `'` as a separator under
  `apostrophe_as_package_separator`, in bundles up to `:5.40`).
* Nesting is textual only. `package Foo; package Foo::Bar;` creates no
  relationship between the two beyond the name.

`__PACKAGE__` is a compile-time constant that expands to the current package
name as a string. It is a term, not a statement.

```go
type PackageStmt struct {
    Name    string
    Version string  // "" if absent
    Body    *Block  // nil for the `;` form
    Span    Span
}
```

For an LSP, maintain a stack of `(package, scopeEnd)` so that any position in
the file can be mapped to its enclosing package. The `;` form's scope ends at
the enclosing block's `}` or EOF; the `BLOCK` form's ends at its own `}`.

### 5.7.2 `class`

New in 5.38, still experimental, and **not in any feature bundle** — `use
v5.42;` alone does not enable it, you need an explicit `use feature 'class';`
(verified against v5.42.0). That makes `class` a feature gate you cannot skip:
without it, `class Point { ... }` is a syntax error, and with the gate off a
parser that accepts it silently will mis-parse code that meant `class` as an
ordinary bareword.

`perly.y:277-291` and `perly.y:293-314` mirror the `package` productions
exactly, with an attribute list added:

```ebnf
ClassStmt = "class" IDENT [ Version ] { Attribute } ";"
          | "class" IDENT [ Version ] { Attribute } Block ;
```

Lexer at `toke.c:8376-8383`:

```c
    case KEY_class:
        ck_warner_d(packWARN(WARN_EXPERIMENTAL__CLASS), "class is experimental");

        s = force_word(s, BAREWORD, ALLOW_PACKAGE);
        s = skipspace(s);
        s = force_strict_version(s);
        PL_expect = XATTRBLOCK;
        TOKEN(KW_CLASS);
```

`PL_expect = XATTRBLOCK` is the state that lets `:isa(...)` be scanned as an
attribute rather than as a ternary colon. In your parser, the class name is
followed by an attribute-parsing context.

Class attributes:

* `:isa(Parent)` — single inheritance. **One superclass only.**
* `:isa(Parent 2.345)` — with a minimum version; will `require` the parent if
  not already loaded (`pod/perlclass.pod:230-238`).

A `class NAME;` (unit-class) declaration is **not** a no-op: it still needs a
constructor generated, so it behaves as though an empty class block followed
(PerlOnJava synthesises exactly that, `StatementParser.java:1191-1229`).

Scoping inside a class block has a wrinkle worth knowing before you design your
scope handling. PerlOnJava uses a deliberate two-scope structure
(`StatementParser.java:1343-1351`) and **delays the inner scope exit** so that
generated accessors and the constructor are registered while class-level
lexicals are still visible (`:1379-1382`, `:1424-1440`), with the synthetic
members registered after the exit (`:1443-1458`). If you generate anything from
`:reader`/`:param`, you will hit the same ordering problem.

Inside a `class` block, three new statement forms become legal, and *only*
inside one — `toke.c:8543` calls `croak_kw_unless_class("field")` and
`perly.y:579` does the same for `ADJUST`. Track `ParseState.InClass`.

### 5.7.3 `field`

`perly.y:1761-1783`, with `fieldvar` at `perly.y:1744-1759`:

```ebnf
FieldStmt = "field" ("$"|"@"|"%") IDENT { Attribute } [ AssignOp Term ] ";" ;
```

```perl
class Point {
    field $x :param :reader = 0;
    field $y :param :reader = 0;
    field @history;
    field %meta;
}
```

Field attributes (`pod/perlclass.pod:241-310`):

| Attribute | Effect |
| --- | --- |
| `:param` | initialise from a named constructor argument of the same name |
| `:param(alt_name)` | same, under a different name |
| `:reader` | generate a reader method named after the field |
| `:reader(get_x)` | reader under a different name |
| `:writer` | generate a writer method `set_NAME`; **scalar fields only** |
| `:writer(set_it)` | writer under a different name |

Interaction rules that a good parser reports on:

* A `:param` field **without** a default is a **required** constructor
  parameter (`pod/perlclass.pod:251-253`).
* With `=` the default applies when the parameter was not passed; with `//=`
  also when passed as `undef`; with `||=` also when passed as false
  (`pod/perlclass.pod:119-123`).
* `:reader` on an array or hash field returns the list, not a reference
  (`pod/perlclass.pod:280`).
* `:writer` on a non-scalar field is a compile-time error
  (`pod/perlclass.pod:307-309`).

The field initialiser is parsed in a special scope —
`class_prepare_initfield_parse()` at `perly.y:1775` — where `$self` and earlier
fields are visible. Field initialisers run in declaration order at construction
time, so `field $b = $a * 2;` after `field $a = 3;` works.

A field must be registered in the scope **as it is parsed**, under both its
plain name and a namespaced form, so that a later field's default can reference
an earlier field (`field $two = $one + 1`). PerlOnJava registers each field three
times (`FieldParser.java:87-100`): as `field:NAME`, as `$NAME`/`@NAME`/`%NAME`,
and in a global registry for inheritance. Field attributes are best stored
generically (`"attr:" + name` → value, `FieldParser.java:172`) rather than
enumerated at parse time — the attribute set is still growing.

```go
type FieldDecl struct {
    Sigil   byte // '$', '@', '%'
    Name    string
    Attrs   []Attr
    Default Expr   // nil if none
    DefOp   string // "=", "//=", "||="
    Span    Span
}
```

### 5.7.4 `method`

`method` is parsed by the same `bare_statement_sub_signature` production as a
signatured `sub` (`perly.y:604-634`), selected by `sigsub_or_method_named`
(`perly.y:766-772`). Differences:

* **`method` always implies signatures**, regardless of the feature bundle:
  `toke.c:5862`, `bool is_sigsub = is_method || FEATURE_SIGNATURES_IS_ENABLED;`
* `class_prepare_method_parse(PL_compcv)` (`perly.y:614-615`) injects `$self`
  as an implicit lexical. It is not in the signature and must not be written
  there.
* `__CLASS__` (`toke.c:8309-8311`) is available inside a method and gives the
  *invocant's* class, which under inheritance differs from `__PACKAGE__`.
* Anonymous methods exist: `perly.y:1548-1552`, `KW_METHOD_anon`.

```perl
class Counter {
    field $n = 0;
    method inc ($by = 1) { $n += $by; return $self; }
    method value { return $n; }
}
```

Fields are directly visible as lexicals inside methods — no `$self->{n}`.

### 5.7.5 `ADJUST`

`perly.y:573-602`, the `bare_statement_phaser` production:

```ebnf
AdjustStmt = "ADJUST" Block ;
```

`toke.c:8324-8331`:

```c
    case KEY_ADJUST:
        ck_warner_d(packWARN(WARN_EXPERIMENTAL__CLASS), "ADJUST is experimental");

        /* The way that KEY_CHECK et.al. are handled currently are nothing
         * short of crazy. We won't copy that model for new phasers, but use
         * this as an experiment to test if this will work
         */
        PHASERBLOCK(KEY_ADJUST);
```

`ADJUST` blocks run at construction, after fields are initialised, in source
order (`pod/perlclass.pod:331-334`). They see `$self` and all fields. Multiple
`ADJUST` blocks per class are allowed and run in declaration order.

Note the grammar routes `ADJUST` through `startsub` and `optsubbody`
(`perly.y:574-586`), so `ADJUST;` — a bodyless phaser — parses. It does nothing.

`ADJUST` is the one construct in the class family that both references agree
must be **context-gated** rather than feature-gated: perl-lsp requires
`in_class_body > 0` (`statements.rs:114-123`) and PerlOnJava throws "ADJUST
blocks are only allowed inside class blocks"
(`SpecialBlockParser.java:87-91`). PerlOnJava also registers `$self` in a fresh
scope before parsing the body so strict-vars checking works (`:96-105`), and —
importantly — **does not execute the block at parse time**, unlike the other
phasers: it is wrapped as an anonymous sub and handed to the class transformer
(`:157-173`).

### 5.7.6 Class construction ordering

For diagnostics, the order at `new()` time is:

1. `:param` fields bound from constructor arguments.
2. Field initialisers evaluated in declaration order (for fields not bound by
   `:param`, or bound-but-defaulted).
3. `ADJUST` blocks in declaration order.

PerlOnJava's `ClassTransformer` is the clearest worked implementation, and three
of its decisions are worth knowing even though you are only parsing.

**The generated constructor** (`generateConstructor`,
`ClassTransformer.java:221`) is `my $self = $class->SUPER::new(%args)` when there
is a parent and `bless {}, $class` otherwise (`:256-286`), then per-field
initialisation, then each ADJUST invoked as `$adjustSub->($self)` (`:304-318`),
then `return $self`.

**`:param` with a default behaves exactly like a signature default** (§5.6.2),
which is the useful thing to know since it means one rule covers both. Verified
against v5.42.0:

| Declaration | `new(x => 3)` | `new()` | `new(x => undef)` | `new(x => 0)` |
| --- | --- | --- | --- | --- |
| `field $x :param = 7` | 3 | 7 | **undef** | **0** |
| `field $x :param //= 7` | 3 | 7 | **7** | 0 |
| `field $x :param ||= 7` | 3 | 7 | 7 | **7** |

So `=` is presence-based and `//=`/`||=` are value-based, matching
`pod/perlclass.pod:119-123`. Note this differs from PerlOnJava, whose codegen
ignores `%args` entirely for `//=` and `||=`
(`ClassTransformer.java:451-505`) — follow Perl.

A `:param` field with **no** default is required; omitting it dies with
`Required parameter 'x' is missing for "P" constructor`. Without `:param`, an
`@` field defaults to `[]` and a `%` field to `{}`
(`ClassTransformer.java:464-490`).

**Beware `:reader` name derivation — the references diverge from Perl here.**
PerlOnJava strips one leading underscore, so `field $_x :reader` yields a reader
named `x` (`defaultAccessorName`, `:675-677`). Real Perl does not: verified
against v5.42.0, `field $_x :reader` generates `_x`, and calling `->x` dies with
`Can't locate object method "x"`. Follow Perl. `:writer` generates `set_<name>`
in both.

Note also what `method` does *not* do: PerlOnJava prepends `my $self = shift
@_;` (`transformMethod`, `:576-585`) and inserts the signature after it
(`:592-601`), but explicitly does **not** alias fields into the method body,
with a comment that automatic injection would break lexical scoping
(`:587-590`). Real Perl does make fields directly visible; if you are resolving
symbols, fields are in scope in every method of the class.

The core tests live in `t/class/` — `field.t`, `accessor.t`, `construct.t`,
`method.t`, `inherit.t`, `phasers.t`. Use them as a conformance suite.

---

## 5.8 Compile-time blocks (phasers)

```ebnf
Phaser = ("BEGIN"|"END"|"INIT"|"CHECK"|"UNITCHECK") Block ;
```

Reached through `yyl_sub` (`toke.c:8313-8322`) with `sub` optional, so both
`BEGIN { }` and `sub BEGIN { }` parse identically. The `PL_expect == XSTATE`
guard means these are only phasers at statement position.

### 5.8.1 Ordering

This is the part everyone gets wrong. The order is:

| Phaser | When | Order among peers |
| --- | --- | --- |
| `BEGIN` | as soon as the block is compiled | **FIFO** — source order |
| `UNITCHECK` | after the enclosing compilation unit finishes compiling | **LIFO** |
| `CHECK` | after the whole program finishes compiling | **LIFO** |
| `INIT` | just before the runtime phase begins | **FIFO** |
| `END` | after the program finishes running | **LIFO** |

Mnemonic: `BEGIN` and `INIT` run in source order; `UNITCHECK`, `CHECK`, and
`END` run in reverse.

`UNITCHECK` is per-file (per `require`, per string `eval`), which makes it the
right phaser for module-level checks; `CHECK` is once per program.

### 5.8.2 Why `BEGIN` matters to the parser

`BEGIN` runs **during compilation**, which means it can change how the rest of
the file parses. `use` is literally implemented as a `BEGIN` block —
`perly.y:713-716`:

```
bare_statement_utilize
	:	KW_USE_or_NO
		startsub
		{ CvSPECIAL_on(PL_compcv); /* It's a BEGIN {} */ }
```

A `BEGIN` block that calls `Some::Module->import` or manipulates `%^H` or
installs a keyword plugin changes the grammar for everything after it. A static
parser cannot follow this in general.

**The pragmatic strategy**, which both PerlOnJava and perl-lsp adopt and which
you should adopt: recognise the *known* pragmas by name and model their effects
(§5.9); treat arbitrary `BEGIN` blocks as opaque and parse on with unchanged
state. Record that an opaque `BEGIN` was seen so that downstream diagnostics can
be softened — "this file contains a `BEGIN` block that may alter parsing" is a
better story than confidently wrong errors.

```go
type PhaserStmt struct {
    Kind SpecialBlockKind // Begin, End, Init, Check, UnitCheck, Adjust
    Body *Block
    Span Span
}
```

---

## 5.9 `use`, `no`, `require`, and how pragmas change the parse

### 5.9.1 Grammar

`perly.y:713-729`:

```ebnf
UseStmt = ("use" | "no") ( Version | ModuleName [ Version ] ) [ ImportList ] ";" ;
```

The lexer builds this at `toke.c:5435-5463` (`S_tokenize_use`), which is worth
reading in full because it is short and it settles every ambiguity:

```c
    if (PL_expect != XSTATE)
        /* diag_listed_as: "use" not allowed in expression */
        yyerror(form("\"%s\" not allowed in expression",
                    is_use ? "use" : "no"));
    PL_expect = XTERM;
    s = skipspace(s);
    if (isDIGIT(*s) || (*s == 'v' && isDIGIT(s[1]))) {
        s = force_version(s, TRUE);
        if (*s == ';' || *s == '}'
                || (s = skipspace(s), (*s == ';' || *s == '}'))) {
            NEXTVAL_NEXTTOKE.opval = NULL;
            force_next(BAREWORD);
        }
        else if (*s == 'v') {
            s = force_word(s, BAREWORD, ALLOW_PACKAGE);
            s = force_version(s, FALSE);
        }
    }
    else {
        s = force_word(s, BAREWORD, ALLOW_PACKAGE);
        s = force_version(s, FALSE);
    }
```

Reading it as a decision tree:

1. `use` / `no` is **only** legal at statement position. In expression position
   it is an error, not a bareword.
2. If the next token starts with a digit, or is `v` followed by a digit → it is
   a **version**: `use 5.036;`, `use v5.36;`.
3. Otherwise it is a **module name**, optionally followed by a version:
   `use Foo::Bar 1.23 qw(a b);`.
4. The odd `else if (*s == 'v')` branch handles `use v5 Foo` — a version
   followed by a module. Rare.

The import list is `optlistexpr` (`perly.y:719`), an arbitrary expression
evaluated at `BEGIN` time. `use POSIX qw(floor ceil)` and `use constant PI =>
3.14159` both fit this slot.

**`no` is `use` with the sense inverted** — the same production, distinguished
by `pl_yylval.ival = is_use` (`toke.c:5461`). `no strict 'refs'` calls
`strict->unimport('refs')`.

### 5.9.2 `require`

`require` is *not* a statement form. It is `KW_REQUIRE` reaching the tree
through `term` (`perly.y:1674-1677`):

```
	|	KW_REQUIRE                              /* require, $_ implied */
	|	KW_REQUIRE term[operand]                         /* require Foo */
```

It has its own precedence level (`perly.y:170`, `%nonassoc KW_REQUIRE`, between
`UNIOP` and `SHIFTOP`). Differences from `use`:

| | `use` | `require` |
| --- | --- | --- |
| When | compile time (`BEGIN`) | run time |
| Calls `import` | yes | no |
| Argument | bareword or version, lexed specially | any expression |
| Is a | statement | expression |

`require Foo::Bar;` with a bareword converts to a path (`Foo/Bar.pm`);
`require $file;` with a string uses the string as a path. `require 5.010;`
checks the Perl version.

### 5.9.3 The pragmas that change the parse

This is the section that determines whether your parser is correct on real code.

**`use strict` / `no strict`.** Scoped to the enclosing block. Three
sub-pragmas: `refs`, `subs`, `vars`. `strict subs` is the one that affects
parsing: with it on, a bareword that is not a declared sub is an error rather
than a string. Your parser should parse it either way and let the analyser
decide — but it must *track* the strict state per scope, because the diagnostic
depends on it.

**`use warnings`.** Does not affect the parse. Track it for diagnostics.

**`use feature`.** Directly changes the grammar. The features that matter:

| Feature | Parse effect |
| --- | --- |
| `signatures` | `sub f (...)` is a signature, not a prototype |
| `say` | `say` is a list operator, not a bareword |
| `state` | `state` is a declarator |
| `switch` | `given`/`when`/`default` are statement forms |
| `try` | `try`/`catch`/`finally` are statement forms |
| `defer` | `defer` is a statement form |
| `class` | `class`/`field`/`method`/`ADJUST` are statement forms |
| `isa` | `isa` is an infix operator |
| `postderef_qq` | `$aref->@*` interpolates in strings |
| `bitwise` | `&`/`|`/`^` gain string variants `&.` etc. |
| `module_true` | a module need not end with a true value |
| `apostrophe_as_package_separator` | `Foo'bar` means `Foo::bar` |
| `refaliasing` | `\$x = \$y` aliases |
| `multidimensional` | `$h{$a,$b}` joins keys with `$;` |
| `indirect` | `new Foo(...)` is a method call |
| `bareword_filehandles` | `print FH ...` works |

The last four are *disablements* — features that were always on and are now
turned off by newer bundles.

**Feature bundles.** `use v5.36;` is shorthand for enabling a whole bundle plus
`use strict`. The bundle contents are in `lib/feature.pm:843-870`. The important
transitions:

```
:5.36   +signatures  +isa  -indirect  -multidimensional
        -bareword_filehandles  -switch
:5.38   +module_true  -smartmatch(deprecated)
:5.40   +try
:5.42   -apostrophe_as_package_separator
```

So `use v5.36;` enables signatures, which means every `sub f (...)` in the rest
of that scope is a signature, not a prototype. **This is the single most
important pragma effect for your parser.**

Also: `use VERSION` where VERSION >= 5.11 implies `use strict`. And `use
v5.36` and later implies `use warnings`.

**`use lib`.** Changes `@INC` at compile time. Affects module resolution, not
parsing. An LSP must honour it for go-to-definition across modules.

**`use constant`.** `use constant PI => 3.14;` installs `PI` as a sub with an
empty prototype, making `PI` a bareword that parses as a call. Without knowing
this, `PI + 1` might parse as `PI(+1)`. Model `use constant` specially:
recognise the `NAME => value` and `{ NAME => value, ... }` forms and register
the names as zero-arity subs.

**`use parent` / `use base`.** Set `@ISA`. No parse effect; big analysis effect
(method resolution).

**`use overload`.** No parse effect.

**Anything else.** A module's `import` can install keyword plugins
(`PLUGSTMT`, `perly.y:878`) or prototypes for exported subs. You cannot follow
this statically. Do what everyone does: maintain a table of known modules whose
exports you model (`List::Util`, `Carp`, `POSIX`, `Try::Tiny`, `Moose`,
`Moo`...), and degrade gracefully otherwise.

### 5.9.4 Threading feature state through the parser

```go
type FeatureSet uint64

const (
    FeatSignatures FeatureSet = 1 << iota
    FeatSay
    FeatState
    FeatSwitch
    FeatTry
    FeatDefer
    FeatClass
    FeatIsa
    FeatModuleTrue
    FeatBitwise
    FeatPostderefQQ
    FeatApostropheSep
    FeatRefaliasing
    // disablements, stored as positive bits meaning "still on":
    FeatIndirect
    FeatMultidim
    FeatBarewordFilehandles
)

var bundles = map[string]FeatureSet{
    "5.10": ...,
    "5.36": FeatSignatures|FeatSay|FeatState|FeatIsa|FeatPostderefQQ|
            FeatBitwise|FeatApostropheSep,
    "5.38": bundles5_36 | FeatModuleTrue,
    "5.40": bundles5_38 | FeatTry,
    "5.42": bundles5_40 &^ FeatApostropheSep,
}
```

Feature state is **lexically scoped to the enclosing block**, like `strict`. Push
a copy on block entry, pop on exit. `use feature` inside an `if` block affects
only that block.

```go
func (p *Parser) parseBlock() *Block {
    saved := p.state.Features
    savedStrict := p.state.Strict
    defer func() { p.state.Features = saved; p.state.Strict = savedStrict }()
    // ... parse statements ...
}
```

The one exception: `BEGIN` blocks execute immediately, so a `use` inside a
`BEGIN` block at file scope affects the rest of the file. In practice `use` is
almost always at file scope already.

```go
type UseStmt struct {
    No      bool   // `no` rather than `use`
    Module  string // "" for `use VERSION`
    Version string // "" if absent
    Args    Expr   // nil if no import list; `()` is a distinct empty list
    Span    Span
}
```

The distinction between `use Foo;` (Args == nil, calls `import` with no args)
and `use Foo ();` (Args == empty list, suppresses `import` entirely) is
semantically real. Do not collapse them.

---

## 5.10 `format` and `write`

Legacy report formatting. Rare in modern code, present in the core test suite,
and structurally interesting because it is the one place Perl has a *different
lexical mode*.

`perly.y:487-501` and the body at `perly.y:787-792`:

```ebnf
FormatStmt = "format" [ Name ] "=" NEWLINE FormatLines "." NEWLINE ;
FormatLines = { PictureLine [ ArgumentLine ] } ;
```

```
formblock:	PERLY_EQUAL_SIGN remember PERLY_SEMICOLON FORMRBRACK formstmtseq PERLY_SEMICOLON PERLY_DOT
```

```perl
format STDOUT =
@<<<<<<<<<<<<<<<   @>>>>>>>>>   @####.##
$name,             $id,         $balance
.
```

Structure:

* `format NAME =` — if `NAME` is omitted it defaults to `STDOUT`. The lexer
  handles the name at `toke.c:5912-5920`, forcing a `BAREWORD` then
  `PREBLOCK(KW_FORMAT)`.
* The `=` must be the last thing on its line.
* Lines alternate: a **picture line** containing field specifiers, then an
  **argument line** containing a comma-separated list of expressions supplying
  values for the fields.
* A lone `.` on a line ends the format.

Note that "picture line then argument line" is the *convention*, not something
the syntax enforces. PerlOnJava classifies each line independently with a single
regex — `FIELD_PATTERN = Pattern.compile("[@^]([<>|#*]+|\\*|#+\\.?#+?)")`
(`FormatParser.java:30`); a line that matches is a picture line, one that does
not is an argument line (`:257-266`). There is no alternation check.

Field specifiers in a picture line:

| Spec | Meaning |
| --- | --- |
| `@<<<` | left-justified text, width = number of chars |
| `@>>>` | right-justified |
| `@\|\|\|` | centred |
| `@###.##` | numeric with the given decimal places |
| `@*` | multi-line text |
| `^<<<` | as `@<<<` but chops the source variable; used for fill mode |
| `~` | suppress the line if all fields are blank |
| `~~` | repeat the line until all fields are exhausted |

The lexer enters format mode via `PL_lex_formbrack` (`toke.c:9764-9770`):

```c
                ENTER_with_name("lex_format");
                SAVEI8(PL_parser->form_lex_state);
                SAVEI32(PL_lex_formbrack);
                PL_parser->form_lex_state = PL_lex_state;
                PL_lex_formbrack = PL_lex_brackets + 1;
                PL_parser->sub_error_count = PL_error_count;
                return yyl_leftcurly(aTHX_ s, 1);
```

In this mode, **whitespace and newlines are significant**, and the ordinary
tokeniser is bypassed. Your Go parser needs the same: `format` switches your
scanner into a line-oriented mode until the terminating `.`.

Two implementation notes worth having. Field **width is the specifier length
excluding the leading `@`/`^`** — `@<<<` is a 4-character field of width 3 in
PerlOnJava's accounting (`FormatParser.java`), which differs from Perl's own; get
this right against `t/op/write.t` rather than against either reference.

And argument lines need a *nested* parse: PerlOnJava spins up a fresh lexer and
parser over just that line's text (`FormatParser.java:375-380`), then loops
`parseExpression` split on commas, falling back to treating the whole line as a
string literal if anything throws (`:407-411`) so a bad argument line never
fails the enclosing parse. That fallback is good recovery design — copy it.

`write` is an ordinary named unary operator (`write` or `write FILEHANDLE`), not
a statement form.

```go
type FormatStmt struct {
    Name  string      // "STDOUT" if omitted
    Lines []FormatLine
    Span  Span
}

type FormatLine struct {
    Picture string // raw picture text, with fields
    Args    []Expr // parsed from the following argument line; nil if none
    Span    Span
}
```

Recovery note: an unterminated format (no `.`) swallows the rest of the file.
Cap it at the next line that looks like a statement start in column 0 and report
"unterminated format".

---

## 5.11 The `{` problem

At statement position, `{` may open a **bare block statement**
(`bare_statement_block`, `perly.y:269`) or an **anonymous hash constructor**
(`anonymous: HASHBRACK ...`, `perly.y:1536-1537`) as an expression statement.
`toke.c` decides in `yyl_leftcurly` from `PL_expect` plus a lookahead
heuristic. **Chapter 3 §3.3 specifies that heuristic; reproduce it verbatim
there and call it from here.** Two consequences people get backwards, both
**[verified]** with `B::Deparse` on 5.42.0: `{}` is an anon hash
(`toke.c:6715`), and `{ $x => 1 }` is a block. Record the guess on the node so
a "did you mean a hash?" quick fix (`+{` / `{;`) can be offered when the block
body fails to parse. The same predicate serves `map`/`grep`/`sort` argument
position (chapter 4 §4.9.2).

---

## 5.12 A statement-parsing skeleton

Pulling §5.2–§5.11 together:

```go
func (p *Parser) parseStatement() Stmt {
    start := p.pos()

    // Trivia first: POD, __END__/__DATA__, comments.
    if p.atPod()      { return p.parsePod() }
    if p.atEndMarker(){ return p.parseDataSection() }

    labels := p.parseLabels()

    switch p.tok.Kind {
    case SEMI:
        p.next()
        return &EmptyStmt{Span: p.span(start)}

    case LBRACE:
        if !p.looksLikeAnonHash() { // ch3 §3.3
            return p.finishBlockStmt(labels, start) // + optional `continue`
        }
        // fall through to the expression statement path

    case KW_IF, KW_UNLESS:
        return p.parseIf(labels, start)
    case KW_WHILE, KW_UNTIL:
        return p.parseWhile(labels, start)
    case KW_FOR, KW_FOREACH:
        return p.parseFor(labels, start)   // dispatches C-style vs foreach
    case KW_SUB:
        if p.peekIsIdentLike(1) { return p.parseNamedSub(start) }
        // else: anonymous sub in expression position; fall through

    case KW_PACKAGE:
        return p.parsePackage(start)
    case KW_USE, KW_NO:
        return p.parseUse(start)
    case KW_TRY:
        if p.state.Features.Has(FeatTry) { return p.parseTry(labels, start) }
    case KW_DEFER:
        if p.state.Features.Has(FeatDefer) { return p.parseDefer(start) }
    case KW_GIVEN, KW_WHEN, KW_DEFAULT:
        return p.parseGivenWhen(labels, start)
    case KW_CLASS:
        if p.state.Features.Has(FeatClass) { return p.parseClass(start) }
    case KW_FIELD:
        if p.state.InClass { return p.parseField(start) }
    case KW_METHOD:
        if p.state.InClass || p.state.Features.Has(FeatClass) {
            return p.parseMethod(start)
        }
    case KW_FORMAT:
        return p.parseFormat(start)
    case IDENT:
        // Phasers: BEGIN/END/INIT/CHECK/UNITCHECK, plus AUTOLOAD/DESTROY.
        // Only at statement position, only when followed by `{`.
        if k, ok := specialBlockKind(p.tok.Text); ok && p.peekIs(1, LBRACE) {
            return p.parsePhaser(k, start)
        }
    case ELLIPSIS:
        if p.peekIs(1, SEMI) { p.next(); p.next(); return &YadaStmt{...} }
    }

    return p.parseExprStatement(labels, start)
}
```

**`do` is deliberately absent from this switch.** It is a primary expression
(§5.3.6), so `do { ... } while (...)` arrives through `parseExprStatement` and
is recognised as a `while` *modifier* on a block-valued expression. Both
references do it this way — perl-lsp routes `Do` through
`expressions/primary.rs:480-490`, and PerlOnJava through
`CoreOperatorResolver.java:127`. `require` is likewise an expression operator
(`CoreOperatorResolver.java:128`), not a statement.

`eval` is absent for the same reason (§5.3.10).

Note which keywords are gated on features. A `try` in a file without `use
feature 'try'` is a bareword sub call — `try { ... } catch { ... };` is exactly
how `Try::Tiny` works, using prototypes. Getting this wrong breaks every file
that uses `Try::Tiny`.

For an LSP, a defensible refinement: parse the *statement form* even when the
feature is off, and emit a diagnostic instead of silently reinterpreting. Which
you prefer depends on whether you value correctness on `Try::Tiny` code (parse
as call) or better errors on 5.34 code missing the feature (parse as statement).
The safe answer: check whether a `try` sub is in scope; if it is, parse as a
call; otherwise parse as the statement form.

---

## 5.13 Error recovery

This is the section that decides whether your parser is usable in an editor.
Both reference implementations get parts of it right and parts of it wrong, and
the wrong parts are as instructive as the right ones.

### 5.13.1 What Perl itself does, and why it is not enough

The entire error-recovery apparatus in `perly.y` is one production
(`perly.y:933-934`):

```
sideff	:	error
			{ $$ = NULL; }
```

Bison's `error` token discards states until it can shift `error`, then discards
input until it can shift the following token — here, a `;`. So Perl's recovery
is: **skip to the next semicolon and continue.** Plus a global error counter
(`PL_error_count`, `toke.c:85`) that aborts after 10 errors, and a
`sub_error_count` (`toke.c:2627`, `toke.c:9769`) used to suppress cascading
errors inside a sub or format whose header already failed.

Adequate for a compiler that reports and exits. Inadequate for a language
server, where a file is *usually* mid-edit and therefore *usually* invalid, and
where the user still expects completion, hover, and outline to work in the parts
that are fine.

### 5.13.2 The contract: errors are data, not control flow

perl-lsp states the rule explicitly in its module header
(`engine/parser/mod.rs:10-13`):

```
//! - **Returns `Ok(ast)` with ERROR nodes** for most parse failures (recovered errors)
//! - **Returns `Err`** only for catastrophic failures (recursion limits, etc.)
```

`result.is_err()` is the wrong error check; callers inspect `parser.errors()`.
Adopt this. In Go:

```go
// Parse always returns a File spanning the whole input. Diagnostics are
// collected separately. It returns an error only for conditions that make
// further parsing meaningless: recursion limit, cancellation.
func Parse(src []byte, opts Options) (*File, []Diagnostic, error)
```

The invariant to hold: **the AST always spans the entire file**, whatever
happened. Every byte is inside some node, even if that node is an error node.
This is what makes folding, outline, and semantic highlighting degrade smoothly
instead of blanking out.

### 5.13.3 Three recovery mechanisms, in order of preference

perl-lsp implements three, and the ordering is right. Copy the ordering.

**(a) Insert what is missing.** When a closing delimiter is absent but the next
token is a strong statement-start signal, *synthesise* the delimiter rather than
erroring. `expect_closing_delimiter` (`helpers.rs:940-963`):

```rust
if self.is_delimiter_recovery_point() {
    let pos = self.current_position();
    let site = Self::recovery_site_for_closer(kind);
    self.errors.push(ParseError::Recovered {
        site, kind: RecoveryKind::InsertedCloser, location: pos });
    return Ok(());
}
```

This is the best mechanism because the resulting tree has the shape the user
intended. `foo($a, $b\nmy $x = 1;` recovers as a complete call with a
diagnostic, not as wreckage.

One subtlety worth stealing (documented at `helpers.rs:977-982`): a *sibling*
closer — a `)` when you wanted `]` — counts as a recovery point but is **not
consumed**, so the enclosing frame consumes it normally. Without that rule,
nested recovery eats delimiters it does not own.

**(b) Insert a missing operand.** An infix operator with nothing usable to its
right gets a placeholder node rather than aborting the expression
(`helpers.rs:918-929`):

```rust
self.errors.push(ParseError::Recovered {
    site: RecoverySite::InfixRhs, kind: RecoveryKind::MissingOperand, location: op_pos });
Some(Node::new(NodeKind::MissingExpression, SourceLocation { start: pos, end: pos }))
```

`$x = ` followed by end-of-line is the single most common state of a file being
typed in. Handling it as a well-formed assignment with a missing RHS is what
lets completion fire at the cursor.

The gate is `is_infix_rhs_absent` (`helpers.rs:~760-868`), which treats
declaration-starting keywords as "cannot be an RHS": `my our state sub package
use no if unless elsif else while until for foreach class method format begin
end check init unitcheck given defer`. Note the escape hatches it needs —
`CHECK()` / `INIT()` without a following `{` (`helpers.rs:795-808`), `given`
without `(` (`:812-817`), `defer` without `{` (`:821-826`) — because all of
those are legal bareword sub calls. Every keyword you add to a "cannot start an
expression" set needs the corresponding escape hatch, or you break real code.

**(c) Panic-mode resynchronisation.** Last resort. `synchronize`
(`helpers.rs:1033-1064`):

```rust
fn synchronize(&mut self) -> bool {
    let mut skipped = 0;
    while !self.tokens.is_eof() && skipped < 100 {
        if self.is_sync_point() {
            if matches!(self.peek_kind(),
                Some(TokenKind::Semicolon) | Some(TokenKind::RightBrace)
                | Some(TokenKind::RightParen) | Some(TokenKind::RightBracket)) {
                let _ = self.consume_token();
            }
            return true;
        }
        let _ = self.consume_token();
        skipped += 1;
    }
    false
}
```

**Read the consume-on-closer branch carefully — it fixes a real infinite loop.**
The comment at `helpers.rs:1041-1047` records the bug: an orphan closer at top
level made `parse_statement` fail on the same token, `synchronize` return `true`
without advancing, and the loop spin forever. **Any resynchronisation routine
that can return "recovered" without consuming a token will hang your parser.**
Assert on it:

```go
func (p *Parser) synchronize() bool {
    before := p.tokenIndex
    ok := p.synchronizeInner()
    // A successful sync that consumed nothing means the caller will retry the
    // same token and fail identically. That is an infinite loop.
    if ok && p.tokenIndex == before && !p.atEOF() {
        panic("synchronize made no progress")
    }
    return ok
}
```

### 5.13.4 Resynchronisation points

perl-lsp's `is_sync_point` (`helpers.rs:1014-1031`) is: any recovery boundary
(`;`, any closing delimiter, EOF — `perl-token/src/kind.rs:779-781`), plus `my
our local state`, `sub package use`, `if unless elsif else`, `while until`, `for
foreach`.

**`sub` and `package` are resync anchors.** perl-lsp includes them in
`is_sync_point`, in `is_delimiter_recovery_point`, and in `is_infix_rhs_absent`.
Note that perl-lsp is *not* line-start sensitive — its parser has no column
concept. That is a simplification you should not copy, and here is why.

Ranked by confidence, for a parser that does track columns:

**Tier 1 — column-0 anchors.** `sub` or `package` (also `class`, `format`,
`__END__`, `__DATA__`, and the phaser names) as the first non-whitespace token
on a line, at column 0. Convention in Perl files is near-universal: top-level
subs and packages start at column 0. This is the strongest structural signal
available, and it is what rescues the second half of a file after a lost brace.

**Tier 2 — line-start keywords at any indentation.** The rest of the
`is_sync_point` set. Strong, but these appear mid-expression too (statement
modifiers), so require them at a line start.

**Tier 3 — punctuation.** `;` at the current brace depth; `}` closing the block
we believe we are in.

**Tier 4 —** `}` at depth 0, EOF.

```go
func (p *Parser) recoverToStatementBoundary(depth int) Span {
    start := p.pos()
    for !p.atEOF() {
        switch {
        case p.tok.Kind == SEMI && p.braceDepth == depth:
            p.next()
            return p.spanFrom(start)
        case p.tok.Kind == RBRACE && p.braceDepth <= depth:
            return p.spanFrom(start) // caller closes the block
        case p.tok.Column == 0 && isTier1Anchor(p.tok):
            // Hard reset: a column-0 `sub`/`package` means we almost certainly
            // lost a brace upstream. ponytail: heuristic, but it is the single
            // highest-value rule in the recovery path.
            p.braceDepth = 0
            return p.spanFrom(start)
        case p.atLineStart() && isTier2Anchor(p.tok):
            return p.spanFrom(start)
        }
        p.next()
    }
    return p.spanFrom(start)
}
```

Cap the skip. perl-lsp uses a hardcoded 100 tokens. A cap prevents a single
error from eating a whole file; when the cap is hit, give up on that statement,
emit one error node covering the skipped range, and let the outer loop try again
from wherever you stopped.

Skip to the *nearest* sync point at the current brace depth, never further.
Skipping to the next `sub` when a `;` is three tokens away throws away a whole
statement the user could have had completion in.

### 5.13.5 Accept, then diagnose

`perly.y` uses this idiom at least five times: `catch` without parens
(`perly.y:812-821`), anonymous sub without a body (`perly.y:1541-1554`), slurpy
with a default (`perly.y:1105-1112`), attributes after a signature
(`perly.y:1202-1211`), `use` in expression position (`toke.c:5439-5442`). In
every case the malformed form is *in the grammar*, and the action produces a
specific message.

It is strictly better than grammar-level rejection: the parse continues with a
correctly-shaped node, and the message names the actual mistake instead of
"syntax error near `)`".

Note the care in the slurpy case — there are **two** productions,
`sigslurpsigil sigvar ASSIGNOP` and `sigslurpsigil sigvar ASSIGNOP term`, so
that both `(@r =)` and `(@r = 1)` get the good message. Match that level of
care for the errors your users actually make:

| Malformed input | Message |
| --- | --- |
| `if ($x) stmt;` | "`if` requires a block; Perl does not allow a bare statement" |
| `sub f ($x) :attr {}` | "attributes must precede the signature" |
| `for my $x @list {}` | "missing parentheses around the `foreach` list" |
| `try {} finally {}` | "`try` requires a `catch` block" |
| `catch {}` | "`catch` requires a `(VAR)`" |
| `print "x" if $a for @l` | "statement modifiers cannot be stacked" |
| `sub f (@r = 1) {}` | "a slurpy parameter may not have a default value" |
| `sub f (@r, $x) {}` | "slurpy parameter not last" |
| `sub f ($x = 1, $y) {}` | "mandatory parameter follows optional parameter" |
| `field $x;` outside a class | "Cannot 'field' outside of a 'class'" (perl's own wording) |
| `ADJUST { }` outside a class | "Cannot 'ADJUST' outside of a 'class'" |
| `for my $x (;;)` | "C-style `for` cannot have a loop variable" |
| `while ()` vs `until ()` | `while ()` is legal (infinite); `until ()` is not |

PerlOnJava's `SignatureParser` is the best worked example of this in either
codebase: it distinguishes "Multiple slurpy parameters not allowed" from
"Slurpy parameter not last" by peeking at whether the offending token is a
sigil (`SignatureParser.java:249-259`), and it rejects `$#` with "'#' not
allowed immediately following a sigil in a subroutine signature"
(`SignatureParser.java:211`). Specific messages are cheap to write and they are
most of what makes a language server feel competent.

### 5.13.6 Orphaned `else` and friends

perl-lsp does something worth copying. When recovery has eaten an `if`, the
stranded `else` block would otherwise be invisible to the LSP. So
`parse_orphaned_else` (`control_flow.rs:906-951`) and `parse_orphaned_elsif`
(`:952-1026`) wrap the block in an `If` node with a **synthetic true
condition**, keeping the body in the tree and analysable
(documented at `control_flow.rs:895-905`).

Generalise the principle: **when a construct's header is unrecoverable but its
body is a well-formed block, keep the body.** The body is where the symbols
are. Do the same for a `sub` whose signature failed, an `if` whose condition
failed, a `foreach` whose list failed.

### 5.13.7 Unclosed constructs

**Unclosed brace.** perl-lsp's `parse_block` (`statements.rs:1131-1214`) never
returns `Err` for this. It parses statements until EOF, then
(`statements.rs:1200-1211`):

```rust
if s.peek_kind() == Some(TokenKind::RightBrace) {
    s.expect(TokenKind::RightBrace)?;
} else {
    let pos = s.current_position();
    s.errors.push(ParseError::syntax(
        "Unclosed block: expected '}' but reached end of input", pos));
}
let end = s.previous_position();
Ok(Node::new(NodeKind::Block { statements }, SourceLocation { start, end }))
```

A well-formed `Block` node containing everything parsed so far, plus one
diagnostic. Every test in `unclosed_block_recovery_tests.rs` asserts
`result.is_ok() && !parser.errors().is_empty()`.

Two improvements on that:

1. **Report at the opening brace, not at EOF.** perl-lsp locates the diagnostic
   at the current position, which is EOF — useless in an editor, which will
   underline the last character of the file. Record the position of every
   unmatched `{` on a stack and report there, with a secondary note at EOF.
2. **Use the column-0 heuristic during the parse** so that later top-level subs
   were parsed at the right nesting level in the first place.

perl-lsp's `test_recovery_on_sub_keyword_in_unclosed_block`
(`unclosed_block_recovery_tests.rs:193`) documents the cost of not doing (2):
given `sub foo { my $x = 1;\nsub bar { my $y = 2; }`, `bar` ends up nested
*inside* `foo`'s body, and `foo` absorbs `bar`'s closing brace. Nested named
subs are legal Perl, so this is not strictly wrong — but it is not what the user
meant, and the resulting outline is wrong. A column-0 check gets it right.

**Unterminated string, heredoc, or regex.** Worse than an unclosed brace,
because the rest of the file lexes as string content. Bound the damage:

* Quoted string with `'` or `"`: stop at end of line unless the line ends in a
  backslash. Report "unterminated string" there. Multi-line strings are legal,
  but an unterminated one is far more common in an editor, and stopping at the
  line end confines the damage to one line. This is a deliberate trade — mark it
  with a comment.
* Heredoc: if the terminator is never found, treat the body as running to the
  next Tier-1 anchor.
* Regex: apply the string rule.

**Locate the heredoc diagnostic at the `<<LABEL`, not at EOF.** perl-lsp gets
this wrong — `drain_pending_heredocs_from` (`heredoc.rs:198-201`) pushes
"Unterminated heredoc: {label}" located at `self.src_bytes.len()`. An editor
underlines the end of the file instead of the declaration the user needs to fix.
The declaration position is in hand when the heredoc is queued; carry it.

**Unterminated POD** legitimately runs to EOF (POD at the end of a module
usually has no `=cut`) — not an error. **Unterminated `format`** is an error;
stop at the next Tier-1 anchor.

### 5.13.8 Error node placement

Attach errors to the **narrowest** enclosing node. If a sub's signature fails
but the body is fine, the `SubDecl` should exist with a valid `Body` and an
error in place of `Sig`. Document symbols, folding, and go-to-definition all
keep working for that sub.

```go
type SubDecl struct {
    // ...
    Sig    *Signature
    SigErr *ErrorNode // non-nil if the signature failed to parse
}
```

**An error in a child must not destroy the parent.**

### 5.13.9 Do not declare error variants you do not emit

A small discipline, and both codebases show why it matters. perl-lsp declares
`RecoveryKind::{InsertedCloser, MissingOperand, TruncatedChain,
InferredSemicolon}` (`syntax/error/mod.rs:169-178`) and `NodeKind::{Error,
MissingExpression, MissingStatement, MissingIdentifier, MissingBlock}`. Only
`InsertedCloser`, `MissingOperand`, `Error`, and `MissingExpression` are ever
constructed. The `Error` node's `partial: Option<Box<Node>>` field
(`perl-ast/src/ast.rs:2453-2462`) is hardcoded `None` at both construction sites
— so the `||` branches in the unclosed-block tests written to accept a
partial-carrying error node can never fire.

The codebase does have a drift-guard test asserting the reserved variants stay
unemitted (`engine/parser/mod.rs:527-676`), which is honest. But the better
answer is not to declare them. Similarly, perl-lsp's `ParseBudget` /
`BudgetTracker` (`syntax/error/mod.rs:203-346`) is a complete, tested API that
the parser never consults — `parse_with_recovery` (`mod.rs:421-448`) fabricates
a tracker after the fact from `self.errors.len()`. The limits actually in force
are hardcoded constants elsewhere.

Define the error vocabulary you emit. Add variants when you emit them.

---

## 5.14 Safe re-parse boundaries

Incremental re-parsing — checkpoints, damage triage, the repair loop, the
splice-equals-fresh-parse test, and when to build any of it — is chapter 6.
This section states the one thing the grammar can say with authority and
chapter 6 §6.5.2 consumes: **which Perl constructs can be re-parsed in
isolation.**

A construct is a safe re-parse boundary if, given an edit strictly inside it,
the parse of everything outside is unchanged. Requirements:

1. It is **brace-delimited** and the braces are balanced.
2. It **cannot alter parse state** visible outside — no `use`, no `BEGIN`, no
   `package NAME;`, no `sub` declaration adding a prototype, no `use constant`.
3. It contains no **unterminated** string, heredoc, regex, POD, or format.
4. It contains no `__END__` / `__DATA__`.

| Boundary | Notes |
| --- | --- |
| **Sub body** | best boundary; cannot change how later code parses |
| **Method body** | same |
| **Bare block** | same |
| **Loop / conditional body** | same |
| **`class` block** | plus: a new `:reader` field adds a method name |
| **`package NAME BLOCK`** | safe; the `package NAME;` form is **not** |
| **Single statement** | do not anchor here: the bookkeeping costs more than the parse (chapter 6 §6.5.2) |

**Never safe:** the file top level; anything containing a `use`, a `BEGIN`, or a
`package NAME;`.

To re-parse a boundary in isolation the parser must restore the `ParseState`
(§5.0) in effect at its opening brace — feature set, strict bits, current
package, lexical scope chain. Capture it when the region is first parsed;
retrofitting this is a rewrite.

---

## 5.15 Cross-checking against the reference implementations

Where the two references disagree with each other or with `perly.y`, the
disagreements are informative. A short list, since you will hit all of these.

**Statement dispatch placement.** perl-lsp dispatches in
`parse_statement_inner` (`statements.rs:153-530`). PerlOnJava's naming is
inverted from what you would guess: `StatementResolver.parseStatement` is the
dispatcher, and `StatementParser` holds the per-construct implementations.
Neither is a name-resolution pass. Pick one name and be consistent.

**`until` desugaring.** perl-lsp emits no `Until` node — `parse_until_statement`
(`control_flow.rs:232-269`) produces `While { condition: Unary{op:"!"},
keyword: Some("until") }`, so consumers must read `keyword`. PerlOnJava does the
same (`StatementParser.java:102-104`). `perly.y` also inverts, via `iexpr`
(`perly.y:1006-1008`). **Do not follow them.** For an LSP, keep the surface form:
a user toggling `while`↔`until` expects the tree to reflect what is on screen,
and formatters and refactorings need the original spelling. Carry a `Negated`
flag on one node type, or use two node types — either preserves the source.

**`unless`.** Same argument. `perly.y:692-693` swaps the branches; keep them
unswapped and carry a flag.

**`do BLOCK while`.** Neither reference models the Perl-specific "runs at least
once, and `last` does not work" semantics in the parser. perl-lsp produces a
plain `StatementModifier { statement: Do{..}, modifier: "while" }`
(`control_flow.rs:770-786`). PerlOnJava detects it purely structurally —
`boolean isDoWhile = expression instanceof BlockNode`
(`StatementResolver.java:948`) — and passes the flag to its loop node. Follow
PerlOnJava: set the flag at parse time so the analyser can warn about
`last`/`next`/`redo` inside it without re-deriving the shape.

**Multi-var foreach.** PerlOnJava has no dedicated parse path for `for my ($k,
$v) (%h)`. It falls out of the generic machinery: `parsePrimary` returns a
`ListNode` for the parenthesised variables, stored unchanged in
`For1Node.variable`, and the n-at-a-time stride is implemented only in the
backend (`backend/jvm/EmitForeach.java:472`). That is a defensible split, but
for an LSP you want the arity visible in the tree — a `Vars []*VarExpr` field —
so that hover on `$v` can say "second of two loop variables".

**Feature gating.** perl-lsp threads **no** feature state through the parser at
all. `use v5.36` is parsed and recorded on the `Use` node and changes nothing
downstream (`declarations.rs:751-777`); `class`, `field`, `method`, `try`,
`defer`, `given` are always available, disambiguated by lookahead
(`statements.rs:288, 343, 363, 376`); signatures parse unconditionally. The one
context gate is `ADJUST`, which requires `in_class_body > 0`
(`statements.rs:114-123`).

PerlOnJava does the opposite: keywords yield null when the feature is disabled
(`StatementResolver.java:106-140`), falling back to bareword parsing, and
`use VERSION` really does enable the bundle, rounding the minor version up to
even and adding implicit `use strict` at ≥5.12 and `use warnings` at ≥5.35
(`StatementParser.java:828-856`).

**Prefer PerlOnJava's approach**, with a caveat. Parsing the superset is
attractive and it is what perl-lsp chose, but it means you can never report
"`class` used without `use feature 'class'`", and — more importantly — you
cannot resolve the prototype-versus-signature ambiguity of §5.18 item 1, which is not
optional. The caveat: where a keyword could plausibly be a sub call
(`try { } catch { };` under `Try::Tiny`), check the pad for a sub of that name
before committing to the statement form.

**`use` executes at parse time.** PerlOnJava actually runs it —
`ModuleOperators.require` then `runSpecialBlock(parser, "BEGIN", ...)`
(`StatementParser.java:876, 986`) — with pragma stacks snapshotted and restored
around the synthetic BEGIN (`:981-991`). You are writing a static parser and
will not do this. What to take from it is the *shape*: model the known pragmas'
effects on `ParseState` explicitly, and note where an unknown module's `import`
could have changed something you cannot see.

**`use Foo` vs `use Foo ()`.** PerlOnJava's `hasEmptyLiteralList`
(`StatementParser.java:895-897`) requires both that tokens were consumed *and*
that the AST is statically empty — so `use Foo @list` with an empty runtime
`@list` still calls `import()`. Keep the three cases distinct in your AST:
`Args == nil` (`use Foo;`), `Args == empty list` (`use Foo ();`), and
`Args == non-empty`.

**The `{` heuristic.** PerlOnJava's `isHashLiteral`
(`StatementResolver.java:998-1364`) re-derives the decision with its own
string scanner (`docs/plans/2026-09-05-parser-prior-art.md` §A1.1.7); the rule
to implement is `toke.c`'s, chapter 3 §3.3.
Its decision ladder (`:1322-1364`) agrees with `toke.c` on the two points
people get wrong: empty `{}` is a **hash** (`toke.c:6715`), and a lowercase
bareword followed by a comma is deliberately **not** a hash indicator
(`:1035-1045`; `toke.c:6810-6814`) because `foo` may be a call. Most of the rest of
the function (`:1073-1188`) exists to avoid false signals from inside strings —
the cost of deciding this in the parser instead of the lexer.

**Labels.** PerlOnJava detects them in `ParseBlock.parseLabel`
(`ParseBlock.java:224`) with two guards worth copying: quote-like operators are
rejected so `m:...:` is not read as a label (`:227-230`), and `sub` is rejected
(`:233-235`). Multiple consecutive labels are consumed in a loop with the last
winning (`ParseBlock.java:135-141`) — note that `perly.y:861-867` keeps *all* of
them, which is the more faithful behaviour.

**Signature parameter registration order.** Both references register each
parameter in the symbol table *immediately*, before parsing its default
(`SignatureParser.java:171-173`). This is what makes `sub f($a, $b = $a)` work.
`perly.y` does the same via `parser->in_my = KEY_sigvar` (`perly.y:1154, 1188`).
Get this right or defaults referencing earlier parameters break.

**Signature default semantics differ by operator, and PerlOnJava's codegen makes
the difference explicit** (`SignatureParser.java:311-336`): `=` is **arity-based**
— `@_ < (paramIndex+1) && ($var = default)` — so passing an explicit `undef`
keeps the `undef`. `//=` and `||=` are **value-based**, emitted as
`$var //= default`, so an explicit `undef` or false *is* replaced. For named
parameters the mapping changes again (`:418-424`): `=` and `//=` both become
`//`. Record the operator; do not normalise it away.

**Recovery loop duplication.** perl-lsp has the same recovery loop written out
three times — `parse_program` (`statements.rs:21-58`), `parse_block`
(`:1146-1191`), `parse_given_block` (`control_flow.rs:829-871`). Write it once:

```go
// recoverStatementInto handles one failed statement: record the error, append
// an error node, resynchronise. Returns false if recovery gave up.
func (p *Parser) recoverStatementInto(out *[]Stmt, err error) bool
```

---
## 5.16 Conformance checklist

Test against the core suite. The directories that matter for this chapter:

| Path | Covers |
| --- | --- |
| `t/comp/` | parser and compilation basics |
| `t/op/` | statement and loop semantics |
| `t/op/sub.t`, `t/op/anonsub.t` | subroutines |
| `t/op/signatures.t` | signatures |
| `t/op/lexsub.t` | lexical subs |
| `t/op/for.t`, `t/op/each.t` | foreach forms |
| `t/op/loopctl.t` | `last`/`next`/`redo` with labels |
| `t/op/try.t` | try/catch/finally |
| `t/op/defer.t` | defer |
| `t/op/switch.t` | given/when |
| `t/op/write.t` | format |
| `t/class/` | 5.38 class syntax — `field.t`, `method.t`, `accessor.t`, `construct.t`, `inherit.t`, `phasers.t` |
| `t/lib/croak/` | error message text |

The `t/class/` suite is the best available specification for the new class
syntax, since the pod is still catching up.

Build order is chapter 6 §6.11; milestone gates are chapter 7 §7.8.

---

## 5.17 Summary of AST node shapes

```go
type Stmt interface{ Node; stmtNode() }

// Statement forms
type EmptyStmt   struct{ Span Span }
type ExprStmt    struct{ X Expr; Span Span }
type ModifiedStmt struct{ Body Expr; Mod StmtModifier; Span Span }
type BlockStmt   struct{ Body *Block; Continue *Block; Labels []Label; Span Span }
type YadaStmt    struct{ Span Span }

// Control flow
type IfStmt      struct{ Negated bool; Cond Expr; Then *Block
                         Elifs []Elif; Else *Block; Span Span }
type WhileStmt   struct{ Negated bool; Cond Expr /*nil = infinite*/
                         Body *Block; Continue *Block; Labels []Label; Span Span }
type CForStmt    struct{ Init, Cond, Post Expr; Body *Block
                         Labels []Label; Span Span }
type ForeachStmt struct{ /* see §5.3.5 */ }
type GivenStmt   struct{ Topic Expr; Body *Block; Span Span }
type WhenStmt    struct{ Cond Expr /*nil = default*/; Body *Block; Span Span }
type TryStmt     struct{ Try *Block; CatchVar *VarExpr; Catch *Block
                         Finally *Block; Span Span }
type DeferStmt   struct{ Body *Block; Span Span }

// Declarations
type SubDecl     struct{ /* see §5.5.9 */ }
type PackageStmt struct{ /* see §5.7.1 */ }
type ClassStmt   struct{ Name, Version string; Attrs []Attr
                         Body *Block /*nil for `;` form*/; Span Span }
type FieldDecl   struct{ /* see §5.7.3 */ }
type PhaserStmt  struct{ /* see §5.8.2 */ }
type UseStmt     struct{ /* see §5.9.4 */ }
type FormatStmt  struct{ /* see §5.10 */ }

// Trivia and errors
type Pod         struct{ /* see §5.1.3 */ }
type DataSection struct{ /* see §5.1.2 */ }
type ErrorStmt   struct{ /* see §5.13.2 */ }
```

Every node carries a `Span`. Declarations additionally carry a `NameSpan`, so
that rename and go-to-definition have a precise target without re-scanning.

---

## 5.18 The five things most likely to break your parser

Ranked by how much time they will cost you.

1. **`sub f (...)` — prototype or signature?** Decided by a feature bit set by a
   `use` statement possibly hundreds of lines earlier, possibly inside a block.
   You must thread feature state through the parser from the start; retrofitting
   it is a rewrite. `toke.c:5862`.

2. **`{` — block or hash?** No amount of local lookahead settles it in general.
   Implement the heuristic of chapter 3 §3.3 (called from §5.11), record the
   guess, and offer a quick fix.
   Users write `+{` and `{;` precisely because Perl gets this wrong too.

3. **`for`'s six shapes.** C-style versus six foreach variants, requiring a
   balanced-paren scan for the semicolon count plus the `yyl_foreach` lookahead
   for the `my`/`\`/`$` prefix. `toke.c:7433-7507` and `perly.y:347-485`.

4. **`do BLOCK while` is not a loop.** It reaches the tree through `term`
   (`perly.y:1567`), not `barestmt`, so it is an expression statement with a
   `while` modifier. `last`/`next`/`redo` silently do the wrong thing in it.
   Diagnose this; users hit it constantly.

5. **Statement modifiers cannot be stacked, and a bareword can be anything.**
   `sideff` has no path back to `expr` (`perly.y:933-952`), and a bareword may be
   a sub call, a string, a filehandle, a label, a package name, or a class name
   depending on `strict subs`, the pad, declared subs, and imported names. The
   bareword problem is not solvable in the parser alone — expose the ambiguity
   in the AST and let the analyser resolve it with scope information.
