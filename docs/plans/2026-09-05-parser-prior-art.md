<!-- ABOUTME: Research notes: engineering assessment of two from-scratch Perl parsers, PerlOnJava (Java) and perl-lsp (Rust). -->
<!-- ABOUTME: What each got right, what each got wrong, measured in real line counts, and what a Go implementation should copy. -->

# A1. Prior Art — PerlOnJava and perl-lsp

Moved out of `docs/specs/perl-parser/` on 2026-09-05: these are research
notes on two other projects, not a specification of Perl, and kept inside
the spec they asserted recommendations the spec's own chapters had since
refuted (its round-1 review, `docs/specs/perl-parser/review/r1-cross.md`,
contradictions 10, 11, 13, 16). Where this document and the spec disagree,
the spec wins. "Chapter N" below means `docs/specs/perl-parser/0N-*.md`.

Two people have already done this. One built a Perl compiler for the JVM;
the other built a Perl language server in Rust. Both wrote the lexer and
parser by hand, both hit the same six or seven walls, and both left their
scar tissue in the repository where it can be read.

This appendix is an assessment, not a survey. Every number in it was
measured with `wc -l`, `grep -c`, or read out of a project's own docs on
**2026-09-04**, against these trees:

| Project | Path | Language | Purpose |
|---|---|---|---|
| PerlOnJava | `/home/perigrin/dev/PerlOnJava` | Java | Compile Perl to JVM bytecode |
| perl-lsp | `/home/perigrin/dev/perl-lsp` | Rust | Perl language server |

perl-lsp is the closer analogue — it is an LSP, it is error-tolerant, it is
incremental, and it deliberately refused a C dependency, which is the same
constraint `CGO_ENABLED=0` puts on us. PerlOnJava is the better source for
*semantics*, because a compiler cannot fake a construct it does not
understand; it either emits correct bytecode or the test fails.

The single most valuable artifact across both repositories is perl-lsp's
record of having built **three** parsers and measured them against each
other. Section A1.3 is that result. If you read one section, read that one.

---

## A1.1 PerlOnJava

### A1.1.1 Architecture

```
                    Perl source (String)
                            │
                            ▼
     ┌──────────────────────────────────────────────┐
     │  Lexer.java                        536 lines │
     │  7 token types: WHITESPACE NEWLINE IDENTIFIER│
     │  NUMBER OPERATOR STRING EOF                  │
     │  "optimized for speed rather than accuracy"  │
     └──────────────────────────────────────────────┘
                            │
                 List<LexerToken>  (whole file, eager)
                            │
                            ▼
     ┌──────────────────────────────────────────────┐
     │  Parser + 35 helper classes     24,003 lines │
     │                                              │
     │  Parser.java ── precedence climbing, 24 levels
     │    ├── StatementResolver  ── statement forms │
     │    ├── ParsePrimary       ── terms           │
     │    ├── ParseInfix         ── binary ops      │
     │    ├── StringParser       ── q qq qw m s tr  │
     │    │     └── StringSegmentParser ── interp   │
     │    ├── ParseHeredoc       ── deferred bodies │
     │    ├── PrototypeArgs      ── proto-driven    │
     │    ├── SubroutineParser   ── sub/sig/attrs   │
     │    └── SpecialBlockParser ── BEGIN: RUNS IT  │
     │              │                               │
     │              └──► compiles + executes the    │
     │                   BEGIN block mid-parse,     │
     │                   then resumes with the      │
     │                   mutated symbol table       │
     └──────────────────────────────────────────────┘
                            │
                            ▼
     ┌──────────────────────────────────────────────┐
     │  AST: 30 node classes            2,153 lines │
     │  Stringly-typed: OperatorNode(String op,...) │
     └──────────────────────────────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              ▼                           ▼
     ┌─────────────────┐         ┌─────────────────┐
     │ analysis/       │         │ semantic/       │
     │ 15 visitors     │         │ ScopedSymbol-   │
     │ 5,141 lines     │         │ Table 1,239 ln  │
     └─────────────────┘         └─────────────────┘
              │
              ▼
     JVM bytecode (ASM)  ── or ── register-based interpreter
```

The pipeline is single-pass and **eager**: `Lexer.tokenize()` builds the
entire token list up front, then drops the input string
(`this.input = null; // Throw away input to spare memory`,
`Lexer.java:` in `tokenize()`). There is no incrementality anywhere. For a
compiler this is fine. For us it is a warning, not a model.

### A1.1.2 The thin lexer, and why

The lexer is **536 lines** and knows **7 token types**. The parser is
**24,003 lines**. That 45:1 ratio is the single most important structural
fact about this codebase, and the source explains it directly
(`src/main/java/org/perlonjava/frontend/lexer/Lexer.java`, class javadoc):

> The Lexer is optimized for speed rather than accuracy.
>
> This can lead to issues in cases such as:
>
> `qq<=>`  — An equal-sign string is parsed as `qq` and `<=>`
>
> `10E10`  — A floating-point number is parsed as `10` and `E10`
>
> The Parser is aware of these issues and implements workarounds to handle them.

So the lexer does not tokenize Perl. It tokenizes a *generic* C-family
language and hands the parser a stream that is wrong in exactly the places
Perl is interesting. The parser then re-reads raw characters whenever it
needs the truth — `StringParser.parseRawStringWithDelimiter()` takes the
token list and an index and walks characters itself.

**Is this the right call for an LSP? No.** Two reasons:

1. **Nothing downstream can trust the token stream.** Semantic highlighting,
   folding, and "what token is under the cursor" all want a token stream
   that is *correct*. Here, `10E10` is two tokens and `qq<=>` is a quote
   operator plus a spaceship. Every consumer must re-derive the truth by
   re-parsing.
2. **It forecloses incrementality.** Incremental relexing needs a token
   whose extent and meaning are locally determined. When the parser
   retroactively reinterprets character ranges, there is no stable token
   boundary to restart from.

It *is* the right call for a batch compiler, which is what this is. The
lesson to carry forward is the inverse of what PerlOnJava did: **a
context-sensitive lexer that the parser drives is worth its cost when the
token stream is a product, not just an intermediate.** perl-lsp reached
exactly that conclusion independently (A1.2.2, A1.3).

### A1.1.3 The decomposition — 36 parser files

The file names are close to a table of contents for the Perl parsing
problem. Full census, `wc -l` on
`src/main/java/org/perlonjava/frontend/parser/`:

| Lines | File | Owns |
|---:|---|---|
| 2,260 | `SubroutineParser.java` | `sub`, signatures, attributes, prototypes, `method` |
| 1,894 | `StringSegmentParser.java` | Interpolation: `"$x->[0] @{[ f() ]}"` |
| 1,755 | `OperatorParser.java` | Named unary + list operators |
| 1,560 | `StatementParser.java` | `if`/`while`/`for`/`foreach`, modifiers |
| 1,506 | `PrototypeArgs.java` | Argument parsing *driven by* a prototype string |
| 1,504 | `StatementResolver.java` | Statement dispatch, `{` block-vs-hash |
| 1,445 | `Variable.java` | Sigils, `${...}`, `$#`, special vars |
| 1,214 | `StringParser.java` | Raw delimiter scanning for `q qq qw m s tr y qr` |
| 903 | `ParseInfix.java` | Binary operator application |
| 826 | `StringDoubleQuoted.java` | Escapes: `\x{}`, `\N{}`, `\Q..\E` |
| 822 | `NumberParser.java` | Numeric literals (repairs `10E10`) |
| 786 | `IdentifierParser.java` | Barewords, `::`, `CORE::`, package names |
| 731 | `ClassTransformer.java` | `class`/`field`/`ADJUST` → classic OO desugar |
| 629 | `SignatureParser.java` | `sub f($a, $b = 1, @rest)` |
| 623 | `ParsePrimary.java` | Term dispatch |
| 511 | `ListParser.java` | Comma lists, trailing commas |
| 510 | `SpecialBlockParser.java` | **BEGIN/END/INIT/CHECK — executes them** |
| 450 | `FileHandle.java` | `print $fh ...`, `<FH>`, indirect object |
| 415 | `FormatParser.java` | `format STDOUT = ...` picture lines |
| 387 | `Parser.java` | Precedence climbing loop, parser state flags |
| 383 | `DataSection.java` | `__DATA__` / `__END__` |
| 360 | `ParserTables.java` | Precedence map, INFIX_OP, CORE prototypes |
| 344 | `StatementCopline.java` | Line-number bookkeeping for diagnostics |
| 343 | `ParseHeredoc.java` | Heredoc declaration + deferred body |
| 320 | `CoreOperatorResolver.java` | Builtin name → node |
| 253 | `ParseBlock.java` | `{ ... }` as a block |
| 200 | `ParseMapGrepSort.java` | Block-taking builtins |
| 185 | `Whitespace.java` | Whitespace, comments, POD skipping |
| 174 | `FieldParser.java` | `field $x :param` |
| 16 | `TestMoreHelper.java` | Test::More special-casing (see A1.1.7) |
| 694 | *(6 more: TokenUtils, ParserNodeUtils, FutureAsyncAwait, ConstantOverload, FieldRegistry, StringSingleQuoted)* | |
| **24,003** | **36 files** | |

Nine of these — `StringParser`, `StringSegmentParser`, `StringDoubleQuoted`,
`StringSingleQuoted`, `ParseHeredoc`, `NumberParser`, `DataSection`,
`FormatParser`, `Whitespace` — total **6,272 lines** and are all doing work
that in a conventional compiler belongs to the lexer. That is **26% of the
parser** paying interest on the 536-line lexer.

### A1.1.4 The AST: stringly-typed, 30 classes

`Node` is an interface with four methods: `accept(Visitor)`, `getIndex()`,
`setIndex(int)`, and an untyped annotation bag
(`setAnnotation(String, Object)` / `getAnnotation(String)`).

There are only **30 node classes for all of Perl** because the operator
nodes are generic:

```java
public class BinaryOperatorNode extends AbstractNode {
    public String operator;   // "+", "->", "=~", "x", "isa", ...
    public Node left;
    public Node right;
}
```

Compare perl-lsp, which has **71 `NodeKind` variants**. PerlOnJava trades
exhaustiveness for brevity: adding an operator costs a string, not a class
and a match arm. The bill comes due in the 15 visitors
(`frontend/analysis/`, 5,141 lines), which all dispatch on
`node.operator.equals("...")` — no compiler check that every operator is
handled, and a typo is a silent runtime miss.

**For Go: do not copy this.** Go's type switch over a sealed-ish node
interface gets you most of Rust's exhaustiveness benefit at similar
verbosity to Java. The annotation bag in particular
(`Object getAnnotation(String)`) is a `map[string]any` by another name and
will be a source of unchecked casts. Prefer typed fields.

The one thing worth copying: `AbstractNode` carries `tokenIndex` — an index
into the token list, not a line/column. Position is derived by a separate
`ErrorMessageUtil` that maps index → line. This keeps nodes small and
decouples the AST from source coordinates. For an LSP you need byte offsets
rather than token indices (perl-lsp uses `ByteSpan`), but the principle —
one integer on the node, coordinate conversion elsewhere — is right.

### A1.1.5 Symbol table

`frontend/semantic/`, 1,339 lines in two files:

| Lines | File | Role |
|---:|---|---|
| 1,239 | `ScopedSymbolTable.java` | Nested scopes; `my`/`our`/`state`/`local`; JVM slot assignment; closure capture; **pragma state** |
| 100 | `SymbolTable.java` | Flat name→index map |

The notable design choice is that `ScopedSymbolTable` carries the *pragma*
state — `strict`, `warnings`, and `feature` bits — alongside variable
bindings, as `BitSet` stacks pushed and popped with scope. This is correct
and it is the thing most hand-rolled Perl parsers get wrong: `use strict`
and `use feature 'say'` are lexically scoped, so `say` is a keyword in one
block and a bareword in the next. Any parser that treats the keyword set as
global is wrong on real code.

**Copy this.** Whatever a Go implementation calls its scope stack, the
feature/strict bitset belongs on it, and `parseStatement` must consult it
before deciding whether `say`, `class`, `method`, `field`, `try`, or `defer`
is a keyword.

### A1.1.6 What PerlOnJava got right

1. **The precedence table as a *shape*, not as data.** `ParserTables.java:321-347`
   is 24 levels, from `or`/`xor` at 1 to `->` at 24, plus a `RIGHT_ASSOC_OP`
   set. The *design* is right — precedence as a table rather than a function
   cascade — and Chapter 4 recommends the same structure for exactly that
   reason.

   **Do not copy the contents.** Chapter 4 §4.13 documents four errors in this
   table, verified against `perl -MO=Deparse,-p` on 5.42. The worst is
   `ParserTables.java:338`:

   ```java
   addOperatorsToMap(16, "-d");        // file tests at 16, below "." at 18
   ```

   Perl puts file tests at level 19, *above* concatenation, so `-e $f . '.bak'`
   is `-e($f . '.bak')` — measured. PerlOnJava's level would group it as
   `(-e $f) . '.bak'`. The line also registers only `-d`, omitting the other 25
   file-test operators.

   Take the table from `perly.y:150-182` (32 levels), which Chapter 4 §4.2
   already transcribes. Take the *idea* of a table from here.

2. **Prototype-driven argument parsing as its own module.**
   `PrototypeArgs.java` (1,506 lines) takes a prototype string and parses
   arguments according to it — `$` scalar, `@` slurp, `&` block-or-coderef,
   `\@` auto-reference, `;` optional separator. Because it is one module
   with one input, prototypes for user subs and for the ~200 core builtins
   (`ParserTables.CORE_PROTOTYPES`) go through the same code path. Builtins
   are not special-cased; they are table entries. That is the correct
   factoring and it is worth stealing wholesale.

3. **Executing BEGIN blocks during the parse.** `SpecialBlockParser` compiles
   the BEGIN block, runs it, and resumes parsing with whatever the block did
   to the symbol table. This is what perl does, and it is the only way
   `BEGIN { eval "sub f(\\@) {}" }` can affect the parse of `f(@a)` three
   lines later. **We cannot and should not do this** — an LSP that executes
   the file it is editing is a remote code execution bug. But it documents
   the semantics precisely, and it tells us where the approximation lives.

4. **Scope-carried pragma state** (A1.1.5).

5. **Real test corpus.** 1,508 `.t` files in `src/test/resources/unit/` plus
   425 module tests. `AGENTS.md` mandates that every new unit test be
   validated against system `perl` *before* it is used to drive PerlOnJava
   fixes — the test encodes Perl's behavior, not the implementation's. That
   discipline is the reason the corpus is worth anything.

### A1.1.7 What PerlOnJava got wrong, or finds painful

Marker counts, measured:

| Scope | TODO/FIXME/XXX/HACK |
|---|---:|
| `frontend/` (lexer+parser+AST+analysis+semantic) | **18** |
| all of `src/main/java` | **98** |
| `@Disabled` / `@Ignore` in `src/test` | **12** |

Eighteen markers across 33,281 lines of frontend is genuinely low. The pain
is not in TODO comments; it is in the *shape* of specific functions.

**The brace disambiguator is a second, worse parser.**
`StatementResolver.isHashLiteral()` decides whether `{` opens a block or a
hashref by pre-scanning the raw token stream ahead. It tracks brace depth,
`=>` sightings, `;` sightings, whether the first token is a sigil, whether
the first token is "key-like", plus a hand-rolled quote-state machine
(`inQuotedString`, `quoteDelimiter`, `quoteOpenDelimiter`, `quoteNesting`,
`quoteEscapeNext`, `awaitingQuoteLikeDelimiter`) because the token stream
does not carry string boundaries. The code says so:

> our scanner walks the raw token stream without tracking string boundaries,
> so it must stay conservative. See `perl5_t/t/re/pat.t` for real-world
> misbalanced regex fragments in string arguments.
> — `StatementResolver.java`, `isHashLiteral` comment

And above it, the smoking gun:

```java
private static final Set<String> IDENTIFIER_BEFORE_BARE_STRING_ARG = Set.of(
        "like", "unlike", "ok", "is", "isnt", "cmp_ok", "can_ok", "isa_ok",
        "pass", "fail", "require", "diag", "note", "explain",
        "skip", "warning_like", "warning_is", "warnings_like");
```

A hardcoded list of **Test::More function names** inside the core brace
disambiguator, because `like "...\x{100}", qr/.../` desynchronizes the
brace counter otherwise. Plus a 16-line `TestMoreHelper.java`. This is what
a thin lexer costs, compounded: the disambiguator cannot use the lexer's
string knowledge, so it reimplements string scanning, badly, and then
patches the failures with a list of specific CPAN function names.

**Root cause: the lexer does not know where strings end.** Fix that in the
lexer and this entire class of hack disappears. This is the strongest
single argument in either repository for a context-sensitive lexer.

**Parser state is a bag of booleans.** `Parser.java` fields:

```java
public boolean parsingForLoopVariable;
public boolean parsingTakeReference;
public boolean parsingDynamicGlobAssignmentRhs;
public boolean parsingIndirectObject;
public boolean parsingDeclaration;
public boolean parsingFutureAsyncAwaitSub;
public String  futureAsyncAwaitForbiddenContext;
public boolean parsingEvalString;
public boolean isTopLevelScript;
public boolean isInClassBlock;
public boolean isInMethod;
public boolean insideBracedDereference;
public boolean insideBlockOperatorArgument;
public int     heredocSkipToIndex;
public int     heredocNewlineIndex;
```

Fifteen mutable flags on the parser, saved and restored ad hoc. Each one is
a real context sensitivity in Perl — they are not gratuitous — but as flat
mutable state they interact combinatorially and cannot be reasoned about
locally. **For Go: make this one `parseContext` struct passed by value**, so
that entering a sub-context is `ctx2 := ctx; ctx2.inClassBlock = true` and
restoration is automatic on return. Same information, no save/restore bugs.

**Format parsing costs 415 lines plus 5 AST node types** (`FormatNode`,
`FormatField`, `TextFormatField`, `NumericFormatField`,
`MultilineFormatField`, `PictureLine`, `FormatLine`, `ArgumentLine` —
actually 8) for a feature almost nobody uses. Budget it last.

**Correctness ceiling, measured.** From `dev/cpan-reports/cpan-compatibility.md`
(auto-generated 2026-09-04), testing randomly-selected CPAN distributions:

| Metric | Count |
|---|---:|
| Modules tested | 16,920 |
| **Pass** | **8,591 (50.8%)** |
| Fail | 8,151 |
| Skipped | 178 |

That 50.8% is *whole-distribution test-suite pass*, not parse success — it
includes runtime, module, and XS gaps, so it is a floor on the parser, not a
measure of it. But it is a real number from a mature project on real CPAN
code, and it calibrates expectations: after 277,869 lines of Java, half of
random CPAN still does not pass its own tests.

---

## A1.2 perl-lsp

### A1.2.1 Architecture

```
                         bytes (&[u8])
                              │
                              ▼
     ┌────────────────────────────────────────────────────┐
     │  perl-lexer                            10,476 lines│
     │                                                    │
     │  LexerMode: ExpectTerm | ExpectOperator            │
     │           | ExpectDelimiter | InFormatBody         │
     │           | InDataSection                          │
     │                                                    │
     │  ~132 TokenKind variants                           │
     │  builtin_signatures.rs (1,861) + phf_lookup (615)  │
     │  checkpoint/ ── resumable state for incremental    │
     └────────────────────────────────────────────────────┘
                       ▲              │
        mode set by ───┘              ▼  Token + ByteSpan
        the parser      ┌────────────────────────────────────────────┐
                        │  perl-parser-core           43,168 lines   │
                        │                                            │
                        │  engine/parser/                            │
                        │    variables.rs      2,186  sigils, ${...} │
                        │    declarations.rs   1,669  my/our/sub/pkg │
                        │    expressions/                            │
                        │      primary.rs      1,504  terms          │
                        │      postfix.rs      1,474  ->[ ->{ ->@*   │
                        │      precedence.rs   1,051  Pratt          │
                        │      calls.rs          822                 │
                        │      unary.rs          589                 │
                        │      quotes.rs         520                 │
                        │    statements.rs     1,385                 │
                        │    control_flow.rs   1,103                 │
                        │    helpers.rs        1,448                 │
                        │  syntax/                                   │
                        │    quote.rs          1,136                 │
                        │    error/mod.rs      1,282  recovery       │
                        │    heredoc.rs          452  FIFO queue     │
                        │  incremental.rs        696                 │
                        └────────────────────────────────────────────┘
                                        │
                                        ▼  perl-ast: 71 NodeKind
                          ┌─────────────────────────────┐
                          │ ParseOutput {               │
                          │   ast,                      │
                          │   diagnostics  ← never Err  │
                          │ }                           │
                          └─────────────────────────────┘
                                        │
                              ┌─────────┴─────────┐
                              ▼                   ▼
                    ┌──────────────────┐  ┌─────────────────┐
                    │ HIR   8,028 ln   │  │ (AST direct for │
                    │ scope graph,     │  │  syntax-only    │
                    │ bindings, stash, │  │  providers)     │
                    │ compile effects  │  └─────────────────┘
                    └──────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │ PIR   2,651 ln   │  data/call/control
                    │ every node keeps │  flow graph for
                    │ a source anchor  │  static analysis
                    └──────────────────┘
                              │
                              ▼
              perl-semantic-analyzer (27,893) → perl-workspace (35,510)
                              │
                              ▼
                    perl-lsp-rs-core (112,913) → perllsp binary
```

The critical structural feature, and the difference from PerlOnJava: **the
parser sets the lexer's mode.** The lexer is not an independent pass. It is
a subroutine the parser calls with an expectation, which is exactly the
`toke.c`/`perly.y` feedback loop that makes Perl parseable at all.

### A1.2.2 The crate map — and a naming trap

`crates/` holds **48 directories**, 45 of them workspace members.
Two names mislead:

- **`perl-parser` is not the parser.** It is 26,023 lines of *refactoring
  and IDE features* — `refactoring.rs` (3,403), `workspace_refactor.rs`
  (1,590), `workspace_rename.rs` (1,255), `import_optimizer.rs` (1,094),
  `ide/lsp_compat/*`. The parser is **`perl-parser-core`**.
- **`tree-sitter-perl-rs` contains no tree-sitter grammar.** Its own
  `lib.rs` says: *"This is a facade over the v3 native parser
  (`perl-parser-core`); it is **NOT** bindings to the C tree-sitter
  grammar."* Same for `perl-tree-sitter-compat`: *"an adapter, not a
  re-implementation."* They keep tree-sitter's **API shape** — S-expressions,
  `Node::kind()`, highlight captures — over their own engine, so editors
  expecting tree-sitter output still work. That is a genuinely good idea
  (A1.5, lesson 3).

Measured sizes of the crates that matter:

| Crate | Lines | Files | Owns |
|---|---:|---:|---|
| `perl-parser-core` | 43,168 | 80 | **The parser.** Lexer wiring, recursive descent, AST build, error recovery, HIR, PIR |
| `perl-workspace` | 35,510 | 36 | Multi-file index, cross-file symbol queries |
| `perl-semantic-analyzer` | 27,893 | 29 | Scope analysis, type inference, class model, imports |
| `perl-parser` | 26,023 | 59 | Refactoring + IDE providers (**not** the parser) |
| `perl-corpus` | 12,829 | 54 | Test corpus machinery |
| `perl-lexer` | 10,476 | 38 | Mode-aware tokenizer, builtin tables, checkpoints |
| `perl-ast` | 5,740 | 3 | 71 `NodeKind` variants + classification |
| `perl-parser-pest` | 3,981 | 5 | v2 PEG parser (frozen, comparison only) |
| `perl-symbol` | 3,167 | 9 | Symbol model |
| `tree-sitter-perl-rs` | 2,240 | 3 | tree-sitter API facade over v3 |
| `perl-token` | 2,175 | 3 | Token types |
| `perl-position-tracking` | 2,132 | 7 | `ByteSpan`, UTF-16 conversion, line index |
| `perl-diagnostics` | 1,403 | 9 | Diagnostic codes |
| `perl-regex` | 1,197 | 16 | Regex-literal internal structure |
| `tree-sitter-perl-c` | 1,018 | 3 | v1 C grammar bindings (benchmark only) |
| `perl-parser-comparison` | 811 | 5 | **v1/v2/v3 differential harness** |
| `perl-pod` | 719 | 1 | POD skipping |
| `perl-tree-sitter-compat` | 446 | 5 | Compat adapter (zero dependents) |
| `perl-incremental-parsing` | 228 | 1 | Thin; real work is in `perl-parser-core` |
| `perllsp` | 20 | 2 | Binary shim |

Whole workspace: **460,219 lines of Rust** across `crates/`.
`perl-lsp-rs-core` alone is 112,913.

### A1.2.3 HIR and PIR — the two IR layers

Both live in `perl-parser-core/src/`. Their own module docs define them.

**HIR** (`hir/`, 8,028 lines across 5 files) — *"the first compiler-substrate
layer above raw parser nodes. It keeps stable language constructs, parser
anchors, source ranges, and scope graph proof data without changing LSP
provider behavior."*

HIR is where the AST stops being syntax and starts being *program facts*.
Its exported type list names them: `ScopeGraph`, `ScopeFrame`, `Binding`,
`BindingReference`, `PackageStash`, `StashGraph`, `PackageInheritanceEdge`,
`PrototypeTable`, `BarewordTable`, `CompileEffect`, `CompilePhaseBlock`,
`ModuleRequest`, `ExportDeclaration`, `DynamicBoundary`.

Two of those are the interesting ones for us:

- **`PrototypeTable` / `PrototypeFact`** — prototypes are collected as
  *facts* with confidence, not applied during the parse. This is the static
  approximation of PerlOnJava's "run the BEGIN block": you cannot know the
  prototype, so you record what you can infer and how sure you are.
- **`DynamicBoundary`** — an explicit marker for "static analysis stops
  here": `eval STRING`, symbolic refs, `AUTOLOAD`, source filters. Naming
  the boundary is better engineering than silently producing a confident
  wrong answer.

`disposition.rs` (1,110 lines) is the context propagator — scalar vs list vs
void — which is precisely what PSC will need.

**PIR** (`pir/`, 2,651 lines across 4 files) — *"a source-anchored tooling
intermediate representation... It models data access, calls, and control
flow for editor tooling and static analysis without executing Perl."*

PIR is a graph (`PirNode`, `PirEdge`, `PirGraph`) for reachability and
dataflow questions. Its invariant is enforced in the doctest:

```rust
// Every source-derived node preserves a source anchor.
assert!(pir.nodes.iter().all(|n| n.source_anchor.is_anchored()));
```

The module doc is unusually candid about scope: *"PIR v0 is a
compiler-substrate data layer only: it does not replace HIR, does not
promote provider behavior, does not claim determinism, and does not run real
Perl."*

**Assessment.** AST → HIR is a real separation and worth having: syntax
nodes answer "what did the user type", HIR answers "what does this program
mean", and an LSP needs both — go-to-definition needs HIR, but
selection-range needs AST. **AST → HIR → PIR is one layer more than a first
implementation needs.** PIR is 2,651 lines and, per its own doc, changes no
provider behavior. Build AST and a scope graph. Add a dataflow IR when a
feature demands it.

### A1.2.4 What perl-lsp got right

1. **The mode-aware lexer.** `LexerMode` with five states, documented in
   `crates/perl-lexer/src/mode.rs` and ADR-0014. The transition table is
   small enough to copy directly:

   | Previous token | Next mode | Example |
   |---|---|---|
   | identifier | `ExpectOperator` | `$x / 2` |
   | number | `ExpectOperator` | `10 / 3` |
   | closing paren/bracket | `ExpectOperator` | `) / 2` |
   | keyword / word-operator | `ExpectTerm` | `if /p/`, `x and /p/` |
   | operator | `ExpectTerm` | `=~ /test/` |
   | opening paren/bracket | `ExpectTerm` | `( /regex/` |

2. **The FIFO heredoc queue** (ADR-0024). Declarations push a
   `PendingHeredoc` onto a `VecDeque`; bodies are collected at the statement
   boundary, in order. This handles `my ($a,$b) = (<<E1, <<E2);` correctly
   and keeps the parser single-pass. `PendingHeredoc` carries `label`,
   `allow_indent` (`<<~`), `quote: QuoteKind`, `decl_span`, and `body_start`
   — with a comment explaining why `body_start` cannot be derived:

   > A later argument in the same expression may be parsed after this
   > heredoc's terminator, so this cannot be derived from statement
   > completion.

   **Copy this design exactly.** It is the correct answer and it is cheap:
   the heredoc module is 452 lines.

3. **Errors are diagnostics, never `Err`.** `Parser::parse_with_recovery()`
   returns `ParseOutput { ast, diagnostics }` and *always* produces a tree.
   `perl-ast-v2` has a `MissingKind` enum — `Expression`, `Statement`,
   `Identifier`, `Block`, `ClosingDelimiter(char)`, `Semicolon`, `Condition`
   — so an error node says what was *expected*, which is what completion
   needs at a half-typed cursor. This is non-negotiable for an LSP and the
   API shape is right.

4. **Budget guards on pathological input.** `MAX_REGEX_BYTES` (64 KB),
   loop budgets, graceful degradation via an `UnknownRest` token. A language
   server cannot hang. Design the escape hatch in from the start.

5. **Byte spans with UTF-16 conversion isolated.** `perl-position-tracking`
   (2,132 lines) keeps `ByteSpan` in the AST and converts to LSP's UTF-16
   positions at the boundary (`convert.rs`, `mapper.rs`, `line_index.rs`).
   Getting this wrong means every position is off on any file with an emoji
   in a comment. Isolate it in one package.

6. **The differential harness.** `perl-parser-comparison` runs the same
   inputs through v1/v2/v3 and asserts on the *disagreement table*. Expected
   verdicts are encoded, so both regressions and improvements force an
   intentional edit. 50 cases, all passing.

7. **Corpus ratchets with machine-generated receipts.** This is the best
   measurement practice in either repository, and it is worth copying as a
   *process*, not just a number. Baselines are JSON receipts in `.ci/`,
   regenerated per merge, carrying commit, timestamp, and Perl version:

   | Corpus | Files | Clean | Rate | Error nodes | Receipt |
   |---|---:|---:|---:|---:|---|
   | Ubuntu system Perl 5.38.2 | 7,095 | 7,047 | **99.3%** | **0** | `.ci/parser-corpus-baseline.json` (2026-05-18, `f201b498c`) |
   | CPAN top 1000 | 9,372 | 8,931 | **95.3%** | 3,015 | `.ci/cpan-corpus-baseline.json` (2026-04-09, `fb9484be`) |
   | Project corpus | 96 | 96 | 100% | 0 | `test_corpus/` + `perl-corpus` |

   The system-Perl run parses 7,095 files in **4.8 seconds** with zero error
   nodes. The CPAN run is the honest one — 435 files still fail, and the
   failure buckets are a ready-made worklist:

   | Bucket | Count | | Bucket | Count |
   |---|---:|---|---|---:|
   | `unexpected_comma_expr` | 48 | | `unclosed_paren` | 34 |
   | `unclosed_paren_identifier` | 43 | | `unclosed_brace_semicolon` | 28 |
   | `unclosed_brace` | 41 | | `expected_left_brace` | 24 |
   | `unexpected_token_in_expr` | 41 | | `unclosed_brace_eof` | 24 |
   | `unexpected_assign_expr` | 34 | | `expected_colon` | 20 |

   Note what dominates: **unclosed delimiters**, which is what you get when
   a quote-like construct's delimiter scan goes wrong and swallows the rest
   of the file. That is the same failure class PerlOnJava patches with its
   Test::More list. It is the hardest part of Perl, and it is still the top
   failure bucket after 460k lines of Rust.

   Separately, the v1/v2/v3 differential
   (`crates/perl-parser-comparison/tests/corpus_differential.rs`) ratchets
   against a 1,268-file corpus: *"v3 clean: 1242, v3 errors: 26,
   crashes: 0"* (97.9%), floor at 1,179. `test_corpus/` holds 1,323 files
   including `real_projects/`, `real_world/`, and adversarial cases
   (`obscure_perl_constructs.pl`, `parser_stress_cases.pl`,
   `source_filters.pl`, `unknown_rest_unterminated_heredoc.pl`).

   Node-kind coverage is tracked too: **66/69 (95.7%)**, with *"0 actionable
   never-seen; 3 recovery-only allowlisted"*
   (`docs/project/status/parser.md`). Knowing which AST nodes your corpus
   has never exercised is a metric worth stealing outright.

8. **Naming the boundary rather than guessing.** Source filters are
   *detected* — `helpers.rs:311`, `fn is_filter_module()` matching `Filter`,
   `Filter::Util::Call`, `Filter::Simple`, `Filter::cpp`, `Filter::exec`,
   `Filter::sh`, `Filter::tee`, `Filter::decrypt` — and the documented
   answer is to stop:

   > **Workaround**: Review files using source filters manually or with
   > Perl's own tools (`perl -c`).
   > — `docs/reference/KNOWN_LIMITATIONS.md:504`

### A1.2.5 What v3 still cannot do

Measured markers, `crates/*/src`:

| Marker | Count | Note |
|---|---:|---|
| TODO/FIXME/XXX/HACK, raw | 157 | Misleading |
| …excluding `perl-ci-hygiene` | **22** | 131 of the 157 are inside the TODO-scanner itself |
| `#[ignore]` in tests | 47 | 10 in `perl-parser`, but 9 of those are *commented-out* dead markers from when `s///` did not work |
| Live lib tests | 6,470 | 17 tracked ignores, 100% pass |

Twenty-two real markers across ~460k lines is exceptionally low, and it is
low for a structural reason: `cargo xtask todos` ratchets the count of
*unlinked* TODOs, so markers get converted into issue links rather than
accumulating in source. That is a good gate; note it as a technique.

The genuine v3 gaps, from `docs/reference/KNOWN_LIMITATIONS.md` (5 items,
*"2% of edge case tests"*) — all ⚠️ Partial, nothing ❌:

| Gap | Detail |
|---|---|
| Complex prototypes | `sub f($$$$;&@)` parses but *"AST accuracy varies"* |
| Deep nested interpolation | `@{[ map { @{[ grep {...} @$_ ]} } @nested ]}` *"may fail"* |
| Formats | Basic works; `^<<<` continuation and `@<<<` picture lines *"may have issues"* |
| Emoji identifiers | ZWJ sequences, variation selectors not validated |
| `5.` decimals | Parses; AST representation imprecise |

Plus two from `docs/reference/FEATURES.md` that contradict the "~100%" claim:

- **`my $x = Foo::Bar->new();` fails.** Bareword qualified names work as a
  statement but *"FAILS in expression"* position.
- **`my_function arg1, arg2;` fails.** User-defined subs without parens are
  not supported; ~70 builtins are special-cased.

That second pair matters more than the exotic ones. Calling a user sub
without parens is ordinary Perl, and it is exactly the case that needs the
prototype table — which is why PerlOnJava's `PrototypeArgs` is 1,506 lines
and perl-lsp's is a fact table with confidence levels.

Machine-readable gaps:
`crates/perl-corpus/fixtures/parser_accuracy/manifest.json` labels
**9 of 51 fixtures** `unsupported_constructs: 1` (`autoload_boundary`,
`postderef_boundary`, `signatures_basic`, `malformed_heredoc_recovery`,
`unterminated_heredoc`, `bad_heredoc_terminator`, …) and marks
**13 `dynamic_boundaries`** across 7 fixtures (`eval_string_boundary`,
`typeglob_alias`, `dynamic_require_boundary`, …). Also live:
`variables.rs:2065` — *"known limitation for `&$var` without braces"* — and
`pir/lower.rs` keeps a running `unsupported_construct_counts` map.

### A1.2.6 What perl-lsp got wrong

**They over-split into 132 crates and had to undo it.** ADR-0008 (2025-01-15)
adopted a microcrate architecture. ADR-0041 (2026-04-14, 38 KB) reverses it.
The workspace `Cargo.toml` today carries **130 lines** reading
`# (perl-lsp-X absorbed into perl-lsp-rs-core::Y — Wave GN)`. The
post-mortem is worth quoting because it is rare to find one this honest:

> The original microcrate bet had four explicit goals: [decoupled versions,
> smaller publish surface, faster compile times, agent-sized work units].
> After ~18 months of operating the microcrate workspace, **only goal #4
> (agent work) delivered as expected.**
>
> - **Decoupled versions never happened.** The 132 crates are too internally
>   coupled; in practice they move together...
> - **Per-update publish surface enlarged.** A leaf change in `perl-token`
>   cascades through ~40 crates...
> - **Compile times didn't improve.** ...132 crates means 132 separate
>   codegen units even when most could share.
> - **Public surface bloat is real.** ...The external story is really
>   4 products + a handful of reusable kernels — the rest is implementation
>   plumbing that became permanent public artifacts.

Some of that is crates.io-specific (published crates cannot use path-only
deps). But the core failure — *architectural seams do not need to be
distribution units* — applies directly to Go packages, where the equivalent
mistake is a `internal/` tree with forty single-file packages and import
cycles you resolve by inventing `internal/types` grab-bags.

**Two stale documents that would mislead a reader.** Both found by the
comparison agent, both worth knowing because you will hit them:

- `docs/adr/0017-workspace-exclusion-strategy.md` claims
  `tree-sitter-perl-c` and `tree-sitter-perl-rs` are excluded from the
  workspace over libclang. They are members today.
- `docs/benchmarks/BENCHMARK_REPORT.md` (July 2025) recommends *"Make the
  pure Rust parser the default in the next release"* — but "pure Rust
  parser" there means **v2 pest**, which was subsequently abandoned. It is a
  historical artifact that reads like current guidance.

**The prose documentation lies; the generated receipts do not.** This is a
process failure worth naming, because it will happen to any project that
writes coverage claims by hand.

| Source | Claim | Kind |
|---|---|---|
| `.ci/parser-corpus-baseline.json` | 99.3% (7047/7095), 0 error nodes | Generated receipt |
| `.ci/cpan-corpus-baseline.json` | 95.3% (8931/9372), 3015 error nodes | Generated receipt |
| `corpus_differential.rs` | 97.9% (1242/1268) | Test assertion |
| `docs/reference/KNOWN_LIMITATIONS.md` | *"~100%"* | Hand-written |
| `docs/reference/FEATURES.md` | *"~99.5%"* — and its own header says 99.996% | Hand-written, self-contradictory |
| `SCOUT_CORPUS_TEST_STRATEGY.md` (2026-03-19) | **72.6%** (5139/7095) | Hand-written, **27 points stale** |

`SCOUT_CORPUS_TEST_STRATEGY.md` quotes the *same 7,095-file corpus* at
72.6% clean with 28,383 error nodes. The live receipt from two months later
says 7,047 clean and zero error nodes. Same denominator, so it is a stale
snapshot rather than a different measurement — but anyone citing that
document is 27 points wrong. Its manifest count is off by 2.4× too (claims
1,849 modules; `.ci/cpan-corpus-manifest.txt` is 4,492 lines).

**The lesson is not "perl-lsp is dishonest" — it is the opposite.** Their
generated receipts are excellent and internally consistent. The lesson is
that a hand-written percentage in a Markdown file has a half-life of about
a quarter. Generate the number or do not publish it.

One soft spot in the live receipts:
**`whitespace_invariance_rate = 0.3`** (n=44) — the parse is not stable
under trailing-whitespace changes in 70% of sampled cases. For an LSP that
re-parses on every keystroke, that is the metric to watch, and it is the
only visibly bad scorer in the set. `strict-clean subset` reports
`insufficient_data`.

**Duplicate/competing modules.** `perl-ast` (5,740 lines) and `perl-ast-v2`
(549 lines) coexist, v2 described as *"an updated AST that uses Range
instead of SourceLocation to support incremental parsing"* — a migration
that has not landed. Similarly, `perl-parser` contains five separate
incremental implementations (`incremental_v2.rs` 2,195,
`incremental_document.rs` 1,511, `incremental_advanced_reuse.rs` 1,311,
`incremental_checkpoint.rs` 1,303, `incremental_simple.rs` 323) plus
`perl-parser-core/src/incremental.rs` (696) plus an
`perl-incremental-parsing` crate (228). That is **7,567 lines across seven
implementations** of one feature. Pick one.

**Format support is still `⚠️ Partial` in v3** — the same feature PerlOnJava
spent 415 lines and 8 node types on. Both projects rank it last. So should we.

---

## A1.3 The result that matters: three parsers, measured

perl-lsp built the same parser three times and kept a harness that runs all
three. This is the most valuable data in either repository because it is the
only place where the architectural question is *answered with numbers*
rather than argued.

| | v1 | v2 | v3 |
|---|---|---|---|
| Approach | tree-sitter (C grammar + `scanner.c`) | pest PEG grammar | **hand-written recursive descent + mode-aware lexer** |
| Crate | `tree-sitter-perl-c` (1,018) | `perl-parser-pest` (3,981) | `perl-parser-core` (43,168) + `perl-lexer` (10,476) |
| Status | Archived, benchmark only | Frozen — *"NOT in default gate"*, *"No new features"* | **Production** |

**Coverage** (`docs/reference/KNOWN_LIMITATIONS.md:63`):

| Parser | Core Perl 5 | Modern Perl 5.38+ | Edge cases | Overall |
|---|---:|---:|---:|---:|
| v3 Native | 100% | 100% | 98% | ~100% |
| v2 Pest | 100% | 100% | 95% | ~99.996% |
| v1 tree-sitter | 95% | **0%** | 60% | ~95% |

v1 scores **0% on modern Perl** — no `class`, `method`, `field`, `try`/`catch`.

**Speed** (same doc):

| Input | v1 (C) | v2 (Pest) | v3 (Native) |
|---|---:|---:|---:|
| 1 KB | ~100 µs | ~200 µs | **~50 µs** |
| 10 KB | ~500 µs | ~1.5 ms | **~200 µs** |
| 100 KB | ~5 ms | ~15 ms | **~2 ms** |
| Incremental | ❌ | ❌ | **✅ <1 ms** |

The hand-written parser is *fastest*, not merely most correct. Incremental
re-parse: 65 µs for a simple edit at 96.8–99.7% node reuse
(`docs/benchmarks/BENCHMARK_RESULTS.md`).

### Why tree-sitter lost

`docs/articles/research/TREE_SITTER_BREAKAGE.md` is the post-mortem, with
`scanner.c` line counts as receipts:

| Construct | Lines of `scanner.c` |
|---|---:|
| `/` disambiguation | ~150 |
| Heredocs | ~200 |
| Quote-like operators | ~250 |
| Special variables | ~100 |
| Formats | ~75 |
| **Total external scanner** | **~975** |

For comparison the doc cites ~200 lines for tree-sitter-javascript, ~300 for
python, ~150 for rust. And it was *still wrong*. The stated root cause:

> Tree-sitter can encode substantial syntactic context through grammar
> state, GLR conflicts, precedence, external tokens, `valid_symbols`, and
> serialized external-scanner state, but the external scanner still cannot
> directly query parse-stack internals.

Its closing line:

> **tree-sitter works when the lexer can be context-free. Perl's lexer is
> context-sensitive by design.** A hand-written recursive descent parser
> with a mode-based lexer is the only architecture that correctly handles
> Perl's grammar.

The doc carries an honest claim boundary: results describe *"the vendored
`tree-sitter-perl-c` target used by perl-lsp at the time of measurement"*,
not current upstream.

### Why pest lost — and this is the subtle one

Pest scored **~99.996%**. It lost anyway, on **failure mode**. From
`crates/perl-parser-comparison/tests/recovery_differential.rs:47`:

> **v1 fails loudly** — it produces ERROR nodes that an LSP diagnostics
> layer can detect and recover from. **v2 fails silently** — it accepts the
> input, returns `Ok`, and produces a plausible-looking but structurally
> wrong AST. Silent failure is more dangerous for an LSP than loud failure:
> a tool that reports `${^MATCH}` as `${` will give wrong hover text and
> wrong go-to-definition targets without any signal that something went
> wrong.

The recovery suite marks v2's apparent recovery wins with an asterisk
because they are cases where it *"'recovers' by silently misparsing the
input."*

**This is the most important lesson in the appendix.** A grammar formalism
that always succeeds will report success on Perl it does not understand.
Chapter 1's distinction between *coverage* and *fidelity* is exactly this
failure, and perl-lsp paid ~4,000 lines to learn it. A Go implementation
must be able to say "I do not know", and its corpus harness must assert on
tree *shape*, not on absence of errors.

### The compressed conclusion

Perl's lexer needs parser state. Tree-sitter's architecture forbids that, so
the grammar's complexity migrated into a 975-line stateful C scanner that
was still wrong. Pest could express the grammar but had no way to fail
honestly. Hand-written recursive descent with a parser-driven lexer mode won
because it is the only one of the three where the lexer can ask the parser
what it is looking at.

And then they kept tree-sitter's *interface* as a facade
(`tree-sitter-perl-rs`) so editors would still work. Discard the engine,
keep the API.

---

## A1.4 The hard cases, side by side

| Construct | PerlOnJava | perl-lsp | Recommendation for Go |
|---|---|---|---|
| **Heredocs** | `ParseHeredoc.java` (343). Parser-side deferred collection, with two `Parser` fields (`heredocSkipToIndex`, `heredocNewlineIndex`) to re-sync the token stream after BEGIN reordering. Backtick heredocs need a special case because the lexer splits `` <<`LABEL` `` into three tokens. | `syntax/heredoc.rs` (452) + ADR-0024. **FIFO `VecDeque<PendingHeredoc>`**, bodies collected at statement boundary. Carries `label`, `allow_indent`, `quote`, `decl_span`, `body_start`. | **perl-lsp's FIFO queue.** Cleaner, handles multiple heredocs per statement natively, ~450 lines. |
| **Prototypes** | `PrototypeArgs.java` (1,506). Prototype string *drives* argument parsing. Same path for user subs and ~200 core builtins via `CORE_PROTOTYPES`. Correct semantics because BEGIN actually runs. | `PrototypeTable` / `PrototypeFact` in HIR — collected as facts with confidence, not applied during parse. Marked `⚠️ Partial` for complex protos like `sub f(&@)`. | **PerlOnJava's table-driven arg parser** for builtins (you know those statically); **perl-lsp's fact model** for user subs (you cannot run BEGIN). |
| **`$x[` vs `$x [`** | `Variable.java` (1,445) + `insideBracedDereference` flag. | `engine/parser/variables.rs` (2,186) + `expressions/postfix.rs` (1,474). Lexer emits distinct token kinds. | Both spend ~1.5–2.2k lines. Budget it. Distinct token kinds from the lexer is cleaner. |
| **Regex vs divide** | Parser-side. `StringParser.parseRawStrings(..., boolean isRegex)` — caller decides, based on which parse function is running. | **`LexerMode::ExpectTerm` / `ExpectOperator`**, ADR-0014, with a 6-row transition table and a dedicated 486-line test file (`slash_ambiguity_tests.rs`). | **perl-lsp's mode machine's *shape*** — a parser-fed enum with a transition table, small and testable — with perl's **eleven** states as its contents (chapter 3 §3.1). The five-state version loses `XTERMORDORDOR`, `XREF` and `XATTRBLOCK` (chapter 3 §3.9.4). |
| **`{` block vs hash** | `StatementResolver.isHashLiteral()` — pre-scan with a *reimplemented* quote-state machine plus a hardcoded Test::More function list. See A1.1.7. | `parse_hash_or_block_inner()`, context-aware, using the lexer's real string boundaries. | **perl-lsp's** shape — lexer-fed, real string boundaries; the rule itself is `toke.c`'s nine steps (chapter 3 §3.3), including the `XREF` skip and the pointer-does-not-move case. PerlOnJava's version is the appendix's clearest cautionary tale, and it is caused by the thin lexer. |
| **Barewords / indirect object** | `IdentifierParser.java` (786) + `FileHandle.java` (450) + `parsingIndirectObject` flag. Full support (`print $fh "text"`). | Curated builtin list + `is_indirect_call_pattern()`. v3 supports it; v1 and v2 do **not**. | Curated list. It is the only tractable approach — the general case needs the symbol table perl builds at runtime. |
| **BEGIN blocks** | `SpecialBlockParser.java` (510). **Compiles and executes them mid-parse**, then resumes with the mutated symbol table. Correct. Also runs `use`. | Parsed as phase blocks (`statements.rs:393,1348`), *not* executed. `CompileEffect` / `CompilePhaseBlock` / `CompileDirective` record what it *would* do, with a `CompileConfidence`. | **perl-lsp's.** Executing user code in a language server is an RCE. Record the effect, mark confidence, move on. |
| **Source filters** | Implemented — `FilterUtilCall.java` provides `Filter::Util::Call` in Java, and `ErrorMessageUtil` re-syncs the token list after a filter rewrites the source. | **Detected and refused.** `is_filter_module()` matches 8 filter modules; documented workaround is `perl -c`. Corpus has `source_filters.pl` as a known-limitation fixture. | **perl-lsp's.** Detect, emit a diagnostic saying analysis is degraded, do not pretend. |
| **`__DATA__` / `__END__`** | `DataSection.java` (383) | `LexerMode::InDataSection` | Lexer mode. One line of state. |
| **Formats** | `FormatParser.java` (415) + 8 AST node types | `LexerMode::InFormatBody`; still `⚠️ Partial` | Lexer mode, and do it **last**. |
| **POD** | `Whitespace.java` (185) | `perl-pod` crate (719) | Lexer-level skip, preserved as trivia for the LSP. |

---

## A1.5 Recommended Go package layout

Mapped from both projects, with the parts that failed removed.

```
internal/
├── source/          # []byte, line index, byte↔(line,col), UTF-16 for LSP
│   ← perl-lsp perl-position-tracking (2,132)
│   Isolate UTF-16 here. Nothing else converts.
│
├── token/           # Token kind constants, Token{Kind, Span}, keyword tables
│   ← perl-lsp perl-token (2,175)
│   Rich kinds (~130), not PerlOnJava's 7.
│
├── lexer/           # Mode-aware tokenizer. THE core decision.
│   ← perl-lsp perl-lexer (10,476), NOT PerlOnJava's 645
│   Expect: perl's eleven PL_expect states (chapter 3 §3.1),
│          plus delimiter / format / data / heredoc contexts
│   Owns: quote-like delimiter scanning, heredoc declaration
│         recognition, POD skip, __DATA__, numeric literals,
│         string extents.
│   Exposes SetMode() to the parser. That coupling is the design.
│
├── ast/             # Node types. Typed structs + type switch.
│   ← perl-lsp perl-ast (5,740); ~70 kinds
│   Every node: Span. No annotation bag.
│
├── parser/          # Recursive descent + Pratt
│   ← perl-parser-core/engine/parser (~15,000 of its 43,168)
│   parser.go        driver, parseContext struct (NOT 15 bool fields)
│   precedence.go    32-level table ← perly.y:150-182 (NOT PerlOnJava, see A1.1.6)
│   expr.go          primary / postfix / unary / infix
│   stmt.go          statements, modifiers, phase blocks
│   decl.go          my/our/state/local, sub, package, class/field/method
│   variable.go      sigils, ${...}, $#, slices, special vars   (~1,500)
│   quote.go         q qq qw m s tr y qr + interpolation        (~2,000)
│   heredoc.go       FIFO pending queue ← ADR-0024              (~450)
│   proto.go         prototype-driven args ← PerlOnJava         (~1,000)
│   recover.go       error nodes, MissingKind, sync points
│
├── scope/           # Lexical scopes + pragma/feature bitsets
│   ← PerlOnJava ScopedSymbolTable (1,239)
│   The feature bitset lives HERE and gates keyword recognition.
│
├── hir/             # Scope graph, bindings, packages, prototype facts,
│   │                 compile effects, DynamicBoundary markers
│   ← perl-lsp hir/ (8,028), trimmed
│   Skip PIR until a feature demands it.
│
└── corpus/          # Differential harness + ratchet
    ← perl-parser-comparison + perl-corpus
    Assert on tree SHAPE vs perl -MO=Concise, not on error absence.
```

**Deliberately omitted**, with reasons:

| Not included | Why |
|---|---|
| `internal/pir` | perl-lsp's own doc: *"does not promote provider behavior"*. 2,651 lines, zero behavior change. Add when a dataflow feature needs it. |
| Separate `internal/incremental` | perl-lsp has seven implementations totalling 7,567 lines. Incrementality is a property of the lexer's checkpoint state and the parser's node reuse — it belongs *in* those packages. |
| `internal/ast/v2` | perl-lsp's unfinished AST migration. Get spans right the first time. |
| Forty micro-packages | ADR-0041. Go packages are architectural seams; they do not need to be one-file. |
| A `format` package | Both projects rank formats last. So should we. |

**Package-count guidance from ADR-0041:** ~10 internal packages, not 45.
Use directories and unexported identifiers for seams within a package. Go's
lack of a `pub(crate)` equivalent means a package *is* your visibility
boundary — which argues for fewer, larger packages, exactly the direction
perl-lsp moved after 18 months.

---

## A1.6 Line-count budget

### Measured, both projects

| Subsystem | PerlOnJava (Java) | perl-lsp (Rust) |
|---|---:|---:|
| Lexer | 645 | 10,476 |
| Parser | 24,003 | ~28,000 (parser-core minus hir/pir/tests) |
| AST | 2,153 (30 classes) | 5,740 (71 kinds) |
| Symbol table / scope | 1,339 | ~3,200 (perl-symbol + scope_analyzer) |
| Analysis / IR | 5,141 (15 visitors) | 10,679 (HIR 8,028 + PIR 2,651) |
| Position tracking | (in ErrorMessageUtil) | 2,132 |
| **Frontend total** | **33,281** | **~52,000** |
| Semantic analysis | (in runtime) | 27,893 |
| Backend / server | 49,946 (bytecode) | 112,913 (LSP core) |
| Runtime | 191,475 | n/a |
| **Whole project** | **277,869** | **460,219** |
| Test corpus | 1,933 `.t` files | 1,323 corpus files |

The two frontends differ by 1.6×, and nearly all of that gap is the lexer
(645 vs 10,476) plus the IR layers (5,141 vs 10,679). The *parsers* are
within 15% of each other: **24,003 vs ~28,000**. That convergence, from two
independent teams in different languages, is the most reliable estimate in
this appendix. **Parsing Perl costs about 25,000 lines. There is no clever
way around it.**

### Go estimate

Go is more verbose than Rust (no enums-with-payload, no `?`, explicit error
returns) and less verbose than Java (no getters, no ceremony, composite
literals, multiple returns). Calibrate at roughly **0.9× Java, 1.25× Rust**
for equivalent logic. Since the two references bracket the answer, use them
as bounds rather than trusting the multiplier.

| Package | Estimate | Basis |
|---|---:|---|
| `internal/source` | 1,200 | perl-lsp 2,132; drop the wire/rope machinery |
| `internal/token` | 1,500 | perl-lsp 2,175, mostly tables |
| `internal/lexer` | **7,000** | perl-lsp 10,476 minus builtin tables (2,476) and tests |
| `internal/ast` | 3,500 | 70 node structs; Go structs are terser than Rust enums here |
| `internal/parser` | **16,000** | Both references land ~25k; Go saves on the lexer-repair code PerlOnJava needs |
| ↳ `variable.go` | 1,500 | both projects: 1,445 / 2,186 |
| ↳ `quote.go` | 2,000 | PerlOnJava 1,214+1,894+826; perl-lsp 1,136 |
| ↳ `heredoc.go` | 450 | perl-lsp 452 |
| ↳ `proto.go` | 1,000 | PerlOnJava 1,506, minus BEGIN execution |
| ↳ `precedence.go` | 400 | PerlOnJava 360 |
| ↳ `recover.go` | 1,200 | perl-lsp syntax/error 1,282 |
| `internal/scope` | 1,500 | PerlOnJava 1,339 |
| `internal/hir` | 4,000 | perl-lsp 8,028, trimmed hard |
| `internal/corpus` | 1,500 | harness only |
| **Frontend total** | **~36,000** | |

Plus tests. Both projects run roughly 1:1 test-to-source in the parser
(perl-parser-core has ~5,000 lines of `*_tests.rs` inside `src/`, and
`slash_ambiguity_tests.rs` alone is 486). **Budget ~30,000 lines of test.**

**Total: 35,000–40,000 lines of Go for the frontend, plus ~30,000 of test.**

Milestones, ordered so each is independently useful:

| Milestone | Cumulative | Gets you |
|---|---:|---|
| Lexer + token + source | 9,700 | Semantic highlighting, folding |
| + AST + core parser (no quotes/heredoc/proto) | 22,000 | Outline, most navigation |
| + quotes, heredocs, interpolation | 26,500 | Real-world files parse |
| + prototypes, error recovery, scope | 30,000 | Completion, diagnostics |
| + HIR + corpus harness | 36,000 | Cross-file, measurable fidelity |

---

## A1.7 Licensing

Not legal advice. What follows is what the files say and the obvious reading.

### PerlOnJava

`LICENSE.md`, verbatim:

> PerlOnJava is free software; you can redistribute it and/or modify it
> under the terms of either:
>
> a. the GNU General Public License as published by the Free Software
>    Foundation; either version 1, or (at your option) any later version, or
>
> b. the "Artistic License" which comes with this Kit.
>
> Copyright (c) Flavio Glock

Both license texts ship: `Copying` (GPL, 13 KB) and `Artistic` (6.2 KB).
This is "same terms as Perl 5" — a **dual license, licensee's choice**.

**Reading code for reference:** unrestricted. Neither license limits reading,
and no license can restrict learning a technique. The 24-level precedence
table is a transcription of `perlop`, i.e. facts about Perl, not creative
expression.

**Copying code:** copyleft. Under either branch, distributing a derivative
carries obligations — GPLv1+ requires source distribution under the same
terms; the Artistic License requires prominent notice of modification and
one of several distribution provisions. **A permissively-licensed Go project
cannot copy PerlOnJava source and stay permissive.** Reimplement from the
documented behavior, not from the file.

**Test corpus:** `src/test/resources/unit/` holds 1,508 `.t` files. These are
project-authored (`AGENTS.md` requires validating each against system `perl`),
so they carry the project's dual license, i.e. the copyleft applies to them
too. Note separately that PerlOnJava's `AGENTS.md` also references
`perl5_t/` — Perl's own test suite, which is not in this checkout and carries
Perl's own license.

### perl-lsp

Two files: `LICENSE-MIT` and `LICENSE-APACHE`. This is the standard Rust
**MIT OR Apache-2.0** dual license, licensee's choice.

`LICENSE-MIT`, verbatim:

> MIT License
>
> Copyright (c) 2026 Steven Zimmerman, CPA
>
> Permission is hereby granted, free of charge, to any person obtaining a
> copy of this software and associated documentation files (the "Software"),
> to deal in the Software without restriction, including without limitation
> the rights to use, copy, modify, merge, publish, distribute, sublicense,
> and/or sell copies of the Software, and to permit persons to whom the
> Software is furnished to do so, subject to the following conditions:
>
> The above copyright notice and this permission notice shall be included in
> all copies or substantial portions of the Software.

**Reading code for reference:** unrestricted.

**Copying code:** permitted under MIT with attribution — retain the copyright
notice and permission text. Apache-2.0 additionally grants patent rights and
requires a NOTICE-style statement of changes. Either way, **a Go project can
port perl-lsp logic** provided it carries the notice. Note that porting
Rust to Go is a translation, not a copy, and the safe practice is to carry
the attribution anyway.

**Test corpus:** `test_corpus/` (1,323 files) is inside the repository and
therefore under the same MIT/Apache-2.0 terms — **reusable with attribution**.
This is the single most practically valuable thing either license permits.
Its adversarial fixtures (`obscure_perl_constructs.pl`,
`parser_stress_cases.pl`, `source_filters.pl`, `heredoc_depth.pl`,
`unknown_rest_unterminated_heredoc.pl`, `low_frequency_nodekinds.pl`) encode
18 months of "things that broke the parser" and would take a long time to
rediscover.

One caveat: `test_corpus/real_projects/` and `test_corpus/real_world/` may
contain code vendored from third parties under *their* licenses. Check
provenance before redistributing those specific subdirectories.

### Practical summary

| | PerlOnJava | perl-lsp |
|---|---|---|
| License | Artistic 1.0 **or** GPL v1+ | MIT **or** Apache-2.0 |
| Read for reference | Yes | Yes |
| Copy code into **PVM specifically** | Plausible — see below | **Yes**, with attribution |
| Copy code into a permissive project | **No** — copyleft | **Yes**, with attribution |
| Best used for | Semantics: what Perl *means*, prototype tables, precedence, pragma scoping | Architecture: mode lexer, heredoc queue, error recovery, **and the test corpus** |

#### PVM's own license changes the PerlOnJava answer

PVM is licensed **Artistic License 2.0** (`LICENSE`, "Copyright (c) 2025
Chris Prather"). PerlOnJava offers the Artistic License as one of its two
branches. These are the same license family Perl itself ships under, so the
blunt reading — "copyleft, therefore off limits" — is too strong for this
project in particular.

That is a materially different position from a generic MIT-licensed Go
project, and it is worth a deliberate decision rather than an assumption in
either direction. Artistic 1.0 and Artistic 2.0 are not the same document,
and their modification-and-distribution clauses differ. This appendix quotes
what the licenses say and stops there; the choice of whether to copy, and
under which branch, belongs to the project owner and not to this
specification.

The conservative path remains available and costs little: read PerlOnJava for
semantics and reimplement in Go. The parser is being rewritten anyway, so
almost nothing would be copied verbatim even if it were unambiguously
permitted.

**Test corpora are the more valuable question.** perl-lsp's corpus is
permissively licensed and directly reusable with attribution. PerlOnJava's
~1,935 `.t` files carry its dual license — but note that many derive from
perl's own suite, which is itself Artistic-or-GPL, so their provenance needs
checking file by file rather than in bulk.

---

## A1.8 Top lessons

1. **The lexer must be context-sensitive and parser-driven.** PerlOnJava's
   536-line lexer pushed ~6,300 lines of character re-scanning into the
   parser and produced `isHashLiteral`, which reimplements string scanning
   and then hardcodes Test::More function names to patch the failures.
   perl-lsp's `LexerMode` (a parser-fed enum with one small transition
   table) solves the same problems structurally — with perl's eleven states
   as its contents, not perl-lsp's five (chapter 3 §3.1, §3.9.4). This is also *why tree-sitter failed*: the
   external scanner cannot query the parse stack. Build the mode machine on
   day one; it cannot be retrofitted.

2. **A parser that cannot fail is worse than one that fails loudly.** Pest
   scored 99.996% coverage and lost anyway, because it returned `Ok` with a
   structurally wrong AST. Never let the harness measure "parsed without
   error" alone. Assert on tree shape.

3. **Do not over-decompose.** ADR-0008 → ADR-0041, 132 crates collapsed to
   ~30, with only 1 of 4 goals met after 18 months. Ten Go packages, not
   forty. Architectural seams are not distribution units.

4. **Copy the heredoc FIFO queue's design; take precedence from `perly.y`.**
   The heredoc queue is small, solved, and in the appendix above. The
   precedence *structure* — a table, not a function cascade — is worth copying
   too, but its *contents* must come from `perly.y:150-182`, not from
   PerlOnJava's table, which Chapter 4 §4.13 shows has four measured errors
   (A1.1.6). ~850 lines you do not have to design, none of it copied blind.

5. **Budget ~25,000 lines for the parser.** Two independent teams in
   different languages converged there (24,003 and ~28,000). Plan around it
   rather than being surprised by it.

6. **Name the boundaries you cannot cross.** perl-lsp's `DynamicBoundary`,
   `CompileConfidence`, and `is_filter_module()` are better engineering than
   a confident wrong answer. PerlOnJava can execute BEGIN blocks because it
   is a compiler; a language server cannot, and should say so rather than
   guess.

7. **Keep the tree-sitter API even after discarding tree-sitter.**
   `tree-sitter-perl-rs` is a facade over the hand-written parser that
   preserves S-expressions and `Node::kind()`, so downstream editor
   integrations keep working. For PVM, which currently ships a gotreesitter
   binding, this is a migration path rather than a rewrite.
