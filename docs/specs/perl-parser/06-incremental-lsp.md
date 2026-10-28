<!-- ABOUTME: Chapter 6 of the Perl parser spec — incremental re-parsing and LSP server architecture in pure Go. -->
<!-- ABOUTME: Damage/repair from an edit range, UTF-8/UTF-16 position tracking, live-buffer error recovery, and the concurrency model. -->

# Chapter 6: Incremental Parsing and LSP Architecture

This chapter is about architecture, not grammar. Chapters 1–5 tell you how to
parse Perl once. This one tells you how to parse it again, ten milliseconds
later, after the user typed one character — and how to wrap that in a language
server that stdlib-only Go can actually ship.

The target consumer is PSC, the type checker already living in
`internal/infer/`. The parser is a component of PSC, not a sibling project:
`internal/infer/` reads the tree `internal/parse/` builds, through the
lowering pass §6.1.6 specifies (`internal/lower`, chapter 4 §4.14), since
the parse tree's vocabulary is not inference's (§6.4.1). The tree-sitter
path it read before is retired. Everything here is shaped by one constraint: **PSC must not
notice that the tree came from an incremental parse.** A tree assembled by
splicing is either indistinguishable from a fresh parse of the same bytes, or
it is a bug. Chapter 6 exists to make that guarantee cheap.

**Target: a keystroke in a 5,000-line Perl file produces a new, PSC-consumable
tree in under 10 ms at p99.** Every recommendation below is downstream of that
number.

---

## 6.1 Why Naive Incremental Parsing Fails for Perl

### 6.1.1 The tree-sitter model, and where it breaks

tree-sitter's incremental algorithm — which `gotreesitter` implements
faithfully — is elegant and worth understanding before rejecting parts of it.
The mechanism is in `/home/perigrin/dev/gotreesitter/incremental.go` (`advance()`, lines 128-175):

1. `Tree.Edit(InputEdit{StartByte, OldEndByte, NewEndByte})` marks nodes
   overlapping the edit as `dirty` and shifts byte offsets after it.
2. On the next parse, a `reuseCursor` walks the old tree in pre-order, offering
   the LR parser old subtrees that begin exactly at the current lookahead byte.
3. `reuseTargetState` checks whether the parser's current LR state has a shift
   (for leaves) or goto (for non-leaves) that accepts the old node's symbol
   *and lands in the state the node was originally built in* (`n.parseState`).
   If yes, the whole subtree is pushed onto the stack in O(1) and the token
   source skips to `n.EndByte()`.

The correctness argument is: **an LR parse state is a complete summary of
everything to the left.** If the parser reaches the same state at the same
byte, the subtree that grew there before will grow there again.

There is one detail in `gotreesitter` worth stealing outright. `advance()`
does:

```go
dirtyHere := cur.dirty
if dirtyHere {
    if nodeBytesEqual(cur.startByte, cur.endByte, c.oldSource, c.newSource) {
        cur.dirty = false   // undo path: bytes came back, reuse is safe
        dirtyHere = false
    }
}
```

A dirty node whose bytes are byte-identical in old and new source is
un-dirtied. That is the undo/redo fast path, and it is nearly free. Keep it.

**Perl breaks the correctness argument, because in Perl the LR state is not a
complete summary of everything to the left.** Perl's parser carries mutable
side state that the LR automaton does not model:

- `PL_expect` — is `/` a division sign or the start of a regex? Is `{` a block
  or an anonymous hash? The answer lives in a lexer variable, not the parse
  stack.
- The pending-heredoc queue — a `<<END` seen on this line changes how the
  *next* line is lexed.
- The symbol table — `sub foo;` predeclared, or a `use` that ran at compile
  time, changes whether `foo $x` parses as a call with arguments.
- Prototypes — `sub mymax(\@)` makes `mymax @list` pass a reference. The
  prototype is compile-time state that alters the parse of later calls.

`gotreesitter` cannot see any of this, so its reuse decisions are made on
incomplete information. Sometimes it gets lucky. Sometimes it silently produces
a tree that does not match a fresh parse — and that is the failure mode you
must design against, because it is invisible until PSC reports nonsense.

### 6.1.2 The invalidation triggers, concretely

This is the enumeration the rest of the chapter is built on. Each of these is a
case where an edit at byte *N* changes the correct parse of bytes far after *N*.

**Class A — lexer-state triggers (invalidate to end of file, or to a provable resync point).**

| # | Trigger | Example | Damage extent |
|---|---------|---------|---------------|
| A1 | Unterminated string/quote-like | typing `"` in `my $x = 1;` | Everything after, until the next quote of the same kind — potentially EOF |
| A2 | New or removed heredoc introducer | adding `<<END` | The current line's remainder is normal, but the *following* lines become heredoc body |
| A3 | Heredoc terminator edited | changing `END` to `ENDX` | Body extends to the next matching terminator or EOF |
| A4 | Unbalanced delimiter in quote-like | `s{foo}{bar}` → `s{foo}{bar` | To EOF or next `}` |
| A5 | `PL_expect` flip at `/` | `$x /$y/` vs `$x / $y / $z` | Division vs. regex retokenizes the whole rest of the expression |
| A6 | `PL_expect` flip at `{` | `map { ... }` vs `+{ ... }` | Block vs. anon-hash changes statement structure |
| A7 | POD block boundary | inserting `=pod` at column 0 | Everything to `=cut` becomes documentation |
| A8 | `__END__` / `__DATA__` inserted | typing `__END__` | Everything after stops being code entirely |
| A9 | Comment `#` inserted | `#` before `my $x` | Rest of line is not code; if the line ended a heredoc terminator, cascades further |
| A10 | Format block `format STDOUT =` | | To the terminating `.` line. perl-lsp names this its own fallback reason, `ContextSensitiveFormat` (`incremental.rs:164`): format bodies need lexer state that token replay cannot reconstruct |

**Class B — semantic triggers (parse of later text depends on an earlier declaration).**

| # | Trigger | Example | Damage extent |
|---|---------|---------|---------------|
| B1 | Prototype added/changed | `sub f(\@)` | Every call site of `f` in the file, and in files that `use` it |
| B2 | `use` / `no` pragma added | `use feature 'signatures'` | Enables `sub f($x, $y)` parsing file-wide from that point |
| B3 | Source filter `use` | `use Filter::…` | Entire file — abandon incrementality, full reparse |
| B4 | `sub` predeclaration | `sub foo;` | `foo LIST` becomes a call with args |
| B5 | Package change | `package Foo;` | Scoping of unqualified names to end of block/file |
| B6 | `BEGIN { }` block edited | | Anything, in principle — treat as file-wide |
| B7 | Constant via `use constant` | | Bareword becomes a term |
| B8 | Overload / `use overload` | | Operator parse unchanged, but PSC types change file-wide |

**Class C — structural triggers (grammar nesting).**

| # | Trigger | Damage extent |
|---|---------|---------------|
| C1 | Unbalanced `{`/`}` | Block nesting shifts for the entire remainder |
| C2 | Unbalanced `(`/`)` | Expression nesting to the closing paren or EOF |
| C3 | Deleted `;` | Two statements merge; damage to the next `;` |
| C4 | Deleted `}` closing a sub | The following subs become nested |

The design consequence is blunt: **you cannot decide reuse by looking only at
the edit range.** You must classify the edit first. Sections 6.4 and 6.5 do
exactly that.

### 6.1.3 The prior art's honest verdict

Look at what perl-lsp actually ships. There are ~7,000 lines of incremental
machinery across `incremental_v2.rs` (2,195), `incremental_document.rs`
(1,511), `incremental_advanced_reuse.rs` (1,311), and
`incremental_checkpoint.rs` (1,303). Then read the module doc on the shipping
server's text sync
(`crates/perl-lsp-rs/src/runtime/text_sync.rs:1-9`):

> We advertise `TextDocumentSyncKind::Incremental` (2): the client sends
> range-based text edits which are applied to the in-memory Rope via
> `apply_changes`. After applying the edits the *entire* document is reparsed
> — incremental *parsing* is future work.

The incremental engine is behind a cargo feature and, per
`text_sync.rs`, the eager path is described as "dormant." Seven thousand lines
of incremental parsing that production does not call.

Worse, look at what `try_advanced_reuse_parse` does — this is the *primary*
strategy in `incremental_v2.rs:348`:

```rust
// Parse the new source to get target tree structure
let mut parser = Parser::new(source);
let new_tree = match parser.parse() { ... };
// ...then analyze how much of the old tree "could have been" reused
```

It performs a **full parse**, then computes reuse statistics against it, then
returns the freshly parsed tree (`incremental_v2.rs:380`). The metrics report
70–90% node reuse. The wall clock does not improve, because a full parse
already happened. The reuse numbers are real and the speedup is zero.

That is not a criticism of the authors — it is the single most valuable
finding in the prior art, and it is why this chapter is opinionated. **Measure
elapsed time, never reuse percentage.** A reuse-rate metric can be gamed by an
implementation that does all the work anyway. Chapter 6's acceptance test is a
stopwatch.

The older engine (`src/incremental.rs`, ~420 lines: checkpoint-bounded token
replay, `CHECKPOINT_INTERVAL = 256` bytes, `MAX_INCREMENTAL_EDIT_BYTES = 4096`,
prefix/suffix token splicing; its own header at `incremental.rs:5-7` says the
AST is still rebuilt from scratch) teaches the other lesson. Its
`FallbackReason::RecoveryDiagnosticsUnstable` (`incremental.rs:249-252`) is

```rust
if !diagnostics.is_empty() { /* fall back to a full parse */ }
```

**Any file with even one diagnostic never takes the incremental path.** A file
being actively typed in almost always has a diagnostic, so the path fires only
on already-clean files — precisely the files where a full re-parse was
affordable anyway. Decide up front which case you are optimising and check
that your invalidation rules do not exclude it: the mid-error file is the case
that matters. Gate on structure (§6.5.1), never on "the file has no errors".
§6.8.5's rule is narrower and different — never reuse a *subtree* that
contains an error.

The genuinely reusable ideas from perl-lsp are narrower and better:

- The **checkpoint** concept (`checkpoint.rs:1-27`, `lex.rs:6-42`) — snapshot
  lexer state at safe byte offsets so re-lexing can start mid-file.
- The **complete lexer state struct** (`perl-lexer/src/checkpoint/core.rs:5-35`)
  — an inventory of exactly what Perl lexing state must be captured.
- The **token resync loop** (`reparse.rs:70-108`) — re-lex forward until a new
  token matches a shifted old token in start, end, *and* kind; splice there.
- The **`shift_offset` underflow guard** (`reparse.rs:16-26`) — a real,
  numbered bug (#2471) where `(offset as isize + shift) as usize` wrapped to a
  huge value on deletions near byte 0. Go has the identical hazard with
  `uint32`.
- The **hard size cap** (`strategy.rs:1`, `MAX_EDIT_SIZE = 64 * 1024`) and the
  `mod.rs:44-56` policy: >64 KB touched, >1 KB single edit, or >10 newlines →
  full reparse, no attempt.

Take those five. Leave the rest.

### 6.1.4 Do you need two IRs? (HIR and PIR)

The task asks specifically about `perl-parser-core/src/hir/` and
`.../src/pir/`. The numbers first: HIR is ~294 KB of Rust across
`model.rs` (106 KB), `lower.rs` (114 KB), `disposition.rs` (47 KB),
`body.rs` (24 KB). PIR adds ~102 KB.

Their stated jobs, from the module docs:

- **HIR** (`hir/mod.rs:1-5`) — "the first compiler-substrate layer above raw
  parser nodes. It keeps stable language constructs, parser anchors, source
  ranges, and scope graph proof data." It carries a scope graph, a stash graph,
  package inheritance edges, prototype tables, bareword tables, and compile-effect
  facts. This is the layer that answers "what does `use` do to the parse."
- **PIR** (`pir/mod.rs:1-9`) — "a source-anchored tooling intermediate
  representation... models data access, calls, and control flow for editor
  tooling and static analysis." Its own doc is emphatic about what it is not:
  "PIR v0 is a compiler-substrate data layer only: it does not replace HIR,
  does not promote provider behavior, does not claim determinism."

Why two? HIR is *name- and scope-resolution* shaped; PIR is *dataflow* shaped.
Reasonable for a system that wants call-hierarchy, dead-code detection, and
control-flow analysis.

**Recommendation: neither HIR nor PIR as perl-lsp built them. One lowering
pass, specified in §6.1.6, and no dataflow IR until a feature demands one.**

An earlier draft of this section recommended building no IR at all, on one
argument: PSC walks CST nodes directly — `internal/infer/infer.go:22-37`
iterates `node.ChildCount()`, tests `child.Kind() == "package"`, calls
`child.Text(source)` — and interposing an IR meant rewriting all 4,265 lines
of it (5,012 non-test lines at 3b9422eb) before a keystroke got faster, while
PIR's own documentation says it "does not promote provider behavior". The
second half stands: PIR did not make anything better yet. The first half is
void, for two reasons, and the recommendation it carried goes with it.

First, PSC is being converted to the new tree regardless. The parser is a
component of PSC, not a sibling project, and `internal/infer/` will consume
`internal/parse/` — so the rewrite the argument was avoiding happens either
way.

Second, the argument measured the wrong interface. §6.4.1 counted the ten
*methods* PSC calls; it did not count the *values* `Kind()` returns.
`internal/infer/*.go` switches on 46 distinct tree-sitter-perl grammar names
(`subroutine_declaration_statement`, `ambiguous_function_call_expression`,
`array_element_expression`, `hash_deref_expression`,
`func0op_call_expression`, `refgen_expression`, ...; measured at 971b9031
as the distinct string literals in the `case` arms of a `switch` on
`x.Kind()`, or on a variable assigned from one, across `infer.go`,
`symbols.go`, `project.go` and `diagnostics.go`, the method chapter 4 §4.14
states), and
`internal/parse/parse.go:146-201` emits 23 coarse kinds with none of those
names (`statement`, `term`, `binary`, `index`, `declaration`, `call`, ...).
Porting the façade hands PSC a `Kind()` that returns `"binary"` where it
expects `"assignment_expression"`. Ten methods port; the vocabulary does not.

What survives of the earlier reasoning is the part about scope. The first
thing PSC needs beyond the tree is Class B invalidation (§6.1.2): to know
that editing `sub f(\@)` invalidates every call site of `f` needs a
declaration-to-use index, and that is a side table, not a replacement tree.
Issue 01a0ad3f is already building the first piece of it — a prototype table
resolved by parsing module source. Two corrections to how the earlier draft
described that table:

- **Not keyed by byte offset.** rust-analyzer considered exactly that key
  and rejected it by name (`docs/book/src/contributing/guide.md`: "We *could*
  use a text offset for the location of a particular item, but that would
  play badly with salsa: offsets change after edits. So, as a rule of thumb,
  we avoid using offsets, text ranges or syntax trees as keys and values for
  queries"). §6.4.3 records the same defect in PSC's
  former `map[uint32]types.Type`. Key facts by node identity — §6.1.6 item 4.
- **Not a poor man's incremental engine.** A side table with no dependency
  edges is a cache: adequate while `Analyze` re-runs over the whole file
  (§6.4.3 says it should, and at 0.5-2 ms it can), and structurally unable
  to express "this sub's types did not change, so its callers need not be
  rechecked". When that is needed, the cheapest documented step is
  rust-analyzer's `ItemTree` — a per-file summary of items that is "stable
  over modifications to function bodies" (`docs/book/src/contributing/architecture.md`)
  and needs no query engine — not a port of salsa, whose distinguishing
  feature (backdating: a dependent is not re-run when its input's value did
  not change) is the part a hand-rolled table cannot replicate. Nothing in
  the measured budget asks for it yet (§6.10.5).

Give up: pre-built dataflow analysis for call hierarchy. Accept it.

### 6.1.5 Prior art: what type checkers that serve an LSP actually do

Until 2026-09-17 this specification cited two Perl projects and reasoned
from first principles about everything else; a grep of `docs/specs/perl-parser/`
for TypeScript, rust-analyzer, Roslyn, Sorbet or Pyright returned nothing.
This section records what a five-angle research pass found in those systems'
own sources, and what this design takes and refuses from each. TypeScript
citations were re-checked against local clones on 2026-09-17
(microsoft/TypeScript `68cead182`, v5.8.3; microsoft/typescript-go
`89d5d5b2`). The rust-analyzer, Roslyn, Sorbet, Pyright and Swift citations
are the research pass's, read from those repositories by that pass and not
re-read here; treat them with the README's caution.

| System | Borrowed | Rejected |
|---|---|---|
| **TypeScript** (`src/compiler/`; Go port `internal/`) | The separate control-flow graph over the syntax tree: `FlowNode{Flags, Node, Antecedent, Antecedents}` (`internal/ast/flow.go:27`), built by the binder, walked backwards by `getFlowTypeOfReference` (`checker.ts:28264`) — §6.1.7. Flow nodes attached only to narrowable nodes, not every node (`types.ts:950-951`, PR 51682). The 24 `narrowTypeBy*` shapes as the transfer map for Perl guards. Byte offsets as the one internal position unit (typescript-go `CHANGES.md:100`). The Go port's decision to shrink the recognised vocabulary rather than grow the checker (`CHANGES.md:11`). | Canonicalise-on-read as the primary mechanism — see the refutation below. Zero-width error nodes (`utilities.ts:1069 nodeIsMissing`): this tree's `Unknown` spans real bytes, which needs the inverse discipline (§6.1.6 item 7). |
| **rust-analyzer** (`crates/`) | Lowering syntax to a canonical store that inference reads exclusively (`hir-ty/src/infer.rs`: `InferenceContext` holds an `ExpressionStore`, results keyed `ArenaMap<ExprId, _>`, no `ast::` in the file). The bidirectional source map as a side table, not a field on the lowered node (`hir-def/src/expr_store.rs:168-186`, `expr_map` / `expr_map_back`). Forward-mapping an erased node to its survivor so position lookup still resolves (`expr_store/lower.rs:1630`: "make the paren expr point to the inner expression as well for IDE resolution"). Rejecting byte offsets as fact keys (`docs/book/src/contributing/guide.md`). Lower to erase variation, not structure (`hir-def/src/item_tree.rs:24-28`: "it is usually a bad idea to desugar a syntax-level construct to something that is structurally different"). `ItemTree` as the cheapest incrementality step (§6.1.4). Every lowered field optional, because the source is routinely broken (`docs/book/src/contributing/syntax.md`). | rowan's green/red tree. Its distinctive win is position-independent, shareable subtrees under edit; Go lacks the interning tricks that make it cheap, and §6.5.2's sub-anchored splice gets the reuse this design needs. salsa, for the reason §6.1.4 gives. |
| **Roslyn** (`src/Compilers/CSharp/Portable/`) | The definition of full fidelity this spec already uses: byte-exact round-trip (`docs/wiki/Roslyn-Overview.md`). Two trees with a back-pointer: `BoundNode.Syntax` (`BoundTree/BoundNode.cs:24`), parens erased at binding (`Binder_Expressions.cs:1397 BindParenthesizedExpression` returns the inner expression; there is no `BoundParenthesizedExpression`). Flow-sensitive lattice analysis by visiting the *bound* tree to a fixed point with no CFG (`FlowAnalysis/AbstractFlowPass.cs:36`) — §6.1.7. | Nothing structural. Note only that the wiki overview describes "the semantic model" and never mentions a bound tree; the source does. Roslyn read from its docs alone lands in the wrong camp. |
| **Sorbet** (`sorbet/`) | The stated reason for lowering, as a rationale rather than a regret (`docs/internals.md`, CFG section): "We only have to implement typechecking for ~11 kinds of instructions (+ control flow) instead of all ~98 kinds of nodes in `parser::Node`". Measured funnel 125 → 31 → 17 kinds (`parser/tools/generate_ast.cc`, `ast/Trees.h:39`, `cfg/Instructions.h:37`). Desugaring the exact variation this chapter worries about: `until` → `while` + `!` (`ast/desugar/Desugar.cc:1737`), `case` → if/else (`:2316`). Single-pass, non-SSA, no-fixpoint inference with loop-carried variables pinned (`docs/internals.md`, Infer section) as the cheapest working point for §6.1.7. Location preservation on every lowered node as the invariant that lets an LSP work over lowered IR. A written `unsupported.md` of refused dynamic features, with rationale. | Its LSP query direction as the only option: Sorbet installs a global `Query::Loc` and re-runs typecheck (`main/lsp/LSPTypechecker.cc:1035`), matching at CFG bindings (`infer/environment.cc:1027`). Verified working; it costs a typecheck per query and degrades exactly where inference does not run (`website/docs/lsp-typed-level.md`). Open — §6.1.8. |
| **Pyright** (`packages/pyright-internal/`) | Normalising parens at *parse* time into a flag rather than a node (`parser/parseNodes.ts:1453 hasParens`; no `Parenthesized` member of `ParseNodeType`) — the cheapest known answer for `($x)` vs `$x`, and a candidate for this design's lowering (§6.1.8). `affectedExpressions` on flow labels (`analyzer/codeFlowTypes.ts`), a precomputed set that lets lazy flow evaluation skip branches that cannot affect an expression. Hover by offset → node → evaluator (`languageService/hoverProvider.ts:329-334`, `parseTreeUtils.ts:102 findNodeByOffset`) as the pull-shaped alternative to Sorbet's push. | Its tree as a fidelity model: trivia is dropped at tokenisation (`docs/internals.md`), so Pyright never had this project's LSP problem. |

**The claim this section was commissioned to test: "TypeScript proves that
one full-fidelity tree plus canonicalising accessors scales, because the
checker walks the `SourceFile` with `skipParens()`". Refuted.**

- The helper is `skipParentheses` (`utilities.ts:4957`); `skipParens` does
  not exist. It has 30 call sites in the 53,015-line `checker.ts`, 24
  `narrowTypeBy*` functions share the load, and `PartiallyEmittedExpression`
  — cited alongside it — has zero occurrences in the checker.
- TypeScript's tree is not a full-fidelity CST. A `Node` carries only
  `pos`/`end` (`types.ts:28-31`); trivia is never a node but the gap before
  `getStart()`, re-derived by `skipTrivia` over the raw text
  (`scanner.ts:640`, `services.ts:410-435`). Punctuation tokens are not all
  in the tree either: `services.ts:531 addSyntheticNodes` re-lexes gaps on
  demand, and the Go port memoises the result in a `tokenCache`
  (`internal/ast/ast.go:2525`). The Go package is named `ast`. This tree
  retains every byte as structure and passes a 620/620 round-trip; TypeScript
  never handed its checker such a tree, so it never tested the proposition.
- Where canonicalise-on-read is real, the cost is the one option (c)
  predicted. The paren case is re-implemented inline in `isMatchingReference`
  (`checker.ts:27096`, twice — once per operand), `getReferenceCandidate`
  (`:27689`), `narrowType`, `binder.ts isNarrowingExpression` and
  `isNarrowableReference`. Across `src/compiler` there are 12 `skip*`/`walkUp*`
  helpers and 14 `getEffective*` normalisers, and `skipTrivia` takes the source
  string — the leak out of the tree abstraction, at exactly the scale feared.
- TypeScript is not a one-tree system in the sense that mattered. Its binder
  builds a separate flow graph with back-pointers into the tree (§6.1.7), and
  the Go rewrite added a genuine lowering pass — `internal/parser/reparser.go`
  desugars JSDoc into synthetic typed nodes flagged `NodeFlagsReparsed`, with
  a `ReparsedClones` list because (`reparser.go:22-24`) "the binder attaches
  symbols to reparsed nodes and not to JSDoc nodes, we need the mapping when
  obtaining symbols and types from JSDoc nodes" — for the case closest to
  this project's: recovering types from an untyped surface. The port was
  otherwise constrained to be faithful (`CHANGES.md` lists "intentional
  changes"; the team's FAQ, discussions/411, describes keeping "semantics and
  structure compatible"), which cuts both ways: the reparser is a change they
  thought worth making despite that rule, and `getReferenceCandidate` being
  ported verbatim (`internal/checker/flow.go:1861`) is faithfulness, not an
  endorsement.

**The survey result.** Every surveyed system that does type inference either
lowers to a canonical tree first (Roslyn, rust-analyzer, Sorbet, Swift's
ASTGen) or never had a lossless tree to begin with (TypeScript, Pyright). The
systems that keep exactly one lossless tree — tree-sitter consumers, Biome,
LibCST — are linters, formatters and refactoring tools; none does type
inference on that tree. tree-sitter's own documentation concedes the gap:
"some types of code analysis are easier to perform using an *abstract*
syntax tree" (`docs/src/using-parsers/2-basic-parsing`). No production
checker walks a node-per-token lossless tree for inference.

One caution from the other direction. CPython deleted its CST layer (PEP 617)
because "the whole tree is kept in memory, keeping many branches that consist
of chains of nodes with a single child" and the intermediate "is not used by
anything else". The disanalogy is the second clause: this CST has a consumer
(the LSP: positions for every byte, trees for broken code, §6.8) that
CPython's did not. The single-child-chain memory cost transfers; watch it in
§6.10.4.

**Not imported.** The research pass looked for and could not verify the
following; this chapter does not claim them, and nobody should re-derive them
from memory:

- That the TypeScript team regards `checker.ts` as a mistake. No first-party
  statement was found. The artifact evidence is that the Go port split it
  into 30 files under `internal/checker/`, with flow narrowing in its own
  `flow.go` — a refusal to reproduce the monolith, not a documented regret.
- That the TypeScript team learned hand-written declarations drift. Only
  secondary sources say so. The documented position is weaker: the publishing
  handbook ranks types generated from source first and DefinitelyTyped as the
  fallback.
- That rust-analyzer's authors regret nothing about HIR. Its author's public
  regrets (rowan's uniform application to code nobody edits, salsa for every
  dependency, proc-macro bulk — `matklad.github.io/2026/08/21/rust-glancer.html`)
  do not mention the lowering layer, but the search was not exhaustive, and
  absence is not evidence.

### 6.1.6 The decision: what inference walks

The question as posed — does inference walk the CST, or is there a lowering
step — had three candidates: (a) walk the CST directly; (b) lower CST → AST
with a back-pointer; (c) one CST with canonicalising accessors that normalise
on read. The evidence answers it, and also says the question was one axis of
two: the flow structure is a separate decision (§6.1.7).

**Decision: (b), scoped as follows.**

1. **`internal/parse/` stays the only tree the parser produces, and the only
   tree the LSP's structural features read.** Positions for every byte,
   trees for broken code, round-trip, folding, rename, format, and any future
   extract-variable refactor read the CST. Sorbet marks exactly that boundary:
   its `preserveConcreteSyntax` mode (`ast/desugar/Desugar.cc:32`) exists for
   one caller, extract-variable (`main/lsp/ExtractVariable.cc:314`), because
   source-rewriting features must see what the user typed. Type queries do
   not need it.

2. **A lowering pass produces the tree inference reads, and its vocabulary is
   chapter 4 §4.14.** That list was written as "the node kinds a Go
   implementation should produce" for PSC — `Var{Sigil}`, `Deref{Sigil}`,
   `Index{Arrow}`, `Key{Arrow}`, `Bareword{Quoted}`, `Repeat{ListRepeat}`,
   `Assign{ListAssign}`, `Call{Parenthesized, Ampersand, Resolved}` — and
   the parser shipped a coarser 23-kind tree instead. The gap between them is
   the lowering. It erases variation, not structure (rust-analyzer's
   `item_tree.rs:24-28`): `$h{k}` and `$h->{k}` become one `Key` with
   `Arrow` set, where today they are `[index]` and `[binary "->" [anon_hash]]`;
   trivia disappears; `unless` may become a negated `Conditional` (Sorbet's
   `Desugar.cc:1737`, and the CST keeps the spelling for the LSP). Where Perl's
   semantics depend on the surface — §4.10's `('a') x 3`, §4.12.2's
   `my ($x) = f()` — the lowered node keeps the distinction as a `Paren` bit
   on every node (chapter 4 §4.14.6), and only the rules that care read it.

3. **The map between the two is a side table in both directions, and an
   erased CST node maps forward to its survivor.** rust-analyzer's
   `expr_map`/`expr_map_back` (`expr_store.rs:168-186`) and Roslyn's
   `BoundNode.Syntax` (`BoundNode.cs:24`) are the two shapes; the side table
   is chosen so the lowered tree carries no syntax pointers. The `($x)` case
   is `lower.rs:1630` verbatim: the paren's CST node maps to the inner
   expression's AST node, so hover on the paren resolves a type. The map is
   many-CST-to-one-AST, which is also what trivia needs.

4. **Facts — inferred types, the prototype table of issue 01a0ad3f, scopes —
   are keyed by node identity, never by byte offset and never stored on CST
   nodes.** rust-analyzer rejected offsets by name (§6.1.4) and states the
   invariant this design adopts: "syntax tree is a value type ... Using the
   tree as a store for semantic info is convenient in traditional compilers,
   but doesn't work nicely in the IDE" (`architecture.md`). §6.10.3 item 5
   and §6.10.5 item 3 already ask for identity keys in PSC's annotation map;
   this makes it the rule for every side table. The identity must be stable
   under edits above the node — an enumeration index by kind, name and
   disambiguator in the manner of `crates/span/src/ast_id.rs`, not a slice
   index — or the tables invalidate wholesale on every keystroke, which is the
   defect §6.4.3 describes for `map[uint32]types.Type` with a different key.

5. **Every lowered field is optional.** The CST the pass reads is routinely
   broken (§6.8), and a lowering that assumes well-formed input panics on the
   half-typed code the LSP exists for. rust-analyzer's accessors are all
   optional for this reason (`syntax.md`: "to accommodate incomplete and/or
   erroneous source code").

6. **Flow analysis reads the lowered tree, never the CST** (§6.1.7).

   The type vocabulary that flows over it is shared with Chalk rather than
   invented here: B::SoN's stamp lattice (`Int <: Num <: Str <: Scalar <:
   List`, with `Undef`/`Ref`/`CodeRef` as siblings of `Str`) and its
   `TypeLibrary` signature format, which is already a side table keyed by
   operator and sub name rather than anything attached to that project's
   nodes. Both are independent of the graph they were built for, which is
   what makes them portable. They should be a versioned contract between PSC
   and Chalk rather than something copied once and left to drift: a divergent
   lattice is pure cost to both projects. Reported by the B::SoN session
   2026-09-17; the lattice and format have not been re-read here. That
   session notes nothing on its side would resist being versioned, and that
   committing Chalk to a contract is not its call — so this is a proposal
   with one party's assent, not an agreement.

7. **`Unknown` is checked at the head of every rule that descends.**
   TypeScript's recovery nodes are zero-width (`nodeIsMissing`), so no rule
   can consume garbage as a subexpression by accident. This tree's `Unknown`
   spans real bytes (`parse.go:14-26`), which is better for the LSP and needs
   the inverse guard: a rule that walks into an `Unknown`'s span types garbage
   silently. `Call{Resolved:false}` already has the field consumers must test
   — tree-sitter found that a recovery node shaped like a normal node is
   missed by consumers (`api.h:581-602`, `ts_node_is_missing` vs
   `ts_node_is_error`; MISSING nodes "are not captured by `(ERROR)` queries").

   As shipped, one node replaces this per-rule discipline (§6.1.8; chapter
   4 §4.14.2): the lowering emits a `Hole` per `Unknown`, a `Hole` may
   stand anywhere a `Node` may, inference types it `Any` and never
   inspects it, and `TestHoleStandsAnywhereANodeMay` holds every `Node`
   field to that.

**Why not (a).** The per-rule variation tax is not hypothetical at 200 rules;
it is observable at 23 kinds, and it was observable in this parser rather
than in the abstract. Measured at 3b9422eb: `$h{k}` was
`[index "{" [term "$h"] [call "k"]]` and `$h->{k}` was
`[binary "->" [term "$h"] [anon_hash [call "k"]]]` — two shapes for one
operation, where §4.14 specifies one shape with an `Arrow` flag.

That case is fixed (d2ebe02a), and the fix is the argument rather than a
refutation of it. One missing branch in the led loop produced two unrelated
trees for one operation, and it survived a corpus of 1,935 files, four
prior commits and a round-trip test, because nothing in a
walk-the-CST-directly design ever has to state what shape an operation takes.
A lowering pass has a vocabulary, so the second shape has nowhere to go.
Every surface distinction the parser has not yet unified is a rule that must
match two shapes for as long as it goes unnoticed.

The standing tax is trivia: it is a sibling node (`parse.go:28-30`), so every
`Children` walk filters it, and the parser already pays that internally
(`isTrivia`, `peekSignificant`, `parse.go:626`). Context flows down the tree
while types flow up (§4.12), so a single walk carries both directions. And no
surveyed checker walks a lossless tree for inference.

**Why not (c).** The one system cited for it is not doing it on a
full-fidelity tree, and where it does normalise on read the cost leaks into
every predicate (§6.1.5). More decisively, (c) presumes a structure-rich CST
to normalise, and this CST is structure-*poor* in exactly the places
inference reads. Measured at 3b9422eb: `($x);` parses to `[term "$x"]` — the
`Paren` node is erased at `internal/parse/term.go:284-293` ("`($x)` is
`$x`"), against §4.10's instruction and §4.16's checkbox; `${$h->{k}};` and
`@{[ 1, 2 ]};` each lex to one `Variable` token and parse to a single
childless `[term]` (issue 01a0ad52); `{k}` in `$h{k}` is a `Call{Resolved:false}`
rather than a quoted bareword. No `StripParens()` un-strips a paren the
parser discarded. Round-trip passes all three, because it proves no byte was
lost, not that the parse was right (chapter 7 §7.2(c)).

**Why not B::SoN's graph as the lowering target.** The obvious candidate,
since that IR exists, already carries type annotations, and is a sibling
project. Measured answers from the B::SoN session, 2026-09-17, not re-read
here:

- The vocabulary is **not** an obstacle, which was the first guess and it was
  wrong. The 97 concrete node kinds (102 classes less 5 abstract bases,
  counted in chapter 4 §4.14) are language-shaped, not op-shaped — `Add`,
  `Concat`, `Subscript`, `RegexMatch`, `PadAccess` vs `EntryDef` (lexical vs
  package, a language distinction). The optree is *input* to `OpMap.pm`, not
  the vocabulary. That session's own read is that the vocabulary would
  survive lowering from a syntax tree, and that doing so drops 639K of
  optree-quirk recovery (rpeep artifacts, nulled ex-list wrappers, OPf_STACKED
  fusions) rather than reimplementing it.
- Neither is the language boundary: the JSON wire is a serialization format,
  so Go consuming it is not a Perl dependency.
- **Circularity is a real obstacle.** `->targ` (58 uses, pad slot indices
  resolved by perl's compiler), `STASH->NAME` (37 uses, the live symbol
  table), and `_walk_package` (classes discovered by walking stashes) mean
  the IR presupposes name resolution and class hierarchy — exactly what a
  static parser computes. This is the same circularity that ruled out reading
  perl's symbol table for imports (§4.8.3 and issue 01a0ad3f).
- **Identity is the structural one, and it is a choice rather than a law.**
  A node's ID *is* its content hash, so GVN happens at construction: `my $a =
  1 + 2; my $b = 1 + 2;` produces one `Add` node, and `$a` and `$b` do not
  exist. That serves a compiler, which wants two identical expressions to be
  one thing. It does not serve position queries, which want them to be two.

The last point is not a reason the shape is unusable, and an earlier draft of
this section overstated it. A merged node does not need *a* source location;
it needs a **set** of them. Position-to-type runs in the direction merging
permits: find the node whose origin set contains this offset, read its stamp.
What a set cannot answer is which origin produced a particular result, which
costs only over-reported diagnostics — a type error on a merged node is
reported at every origin, some possibly fine.

**Provenance is captured at construction or not at all**, because that is the
only moment both origins exist; after hash-consing the information is gone and
no later pass reconstructs it. This is worth recording as a shared design
element rather than a PSC accommodation: error messages that name every site,
debug-info line mapping that survives optimisation (the problem DWARF exists
for), profiling attribution after code motion, and explaining a merge to a
reader all need it, and all of them are compiler concerns. B::SoN currently
reads `nextstate` for control-flow structure and discards its line numbers on
the same walk that could capture them (8 sites; `B::COP` exposes `line`,
`file` and `stashpv`).

**An origin is a span, not a line**, and the contract should say so rather
than being shaped around what an optree can supply. Measured by the B::SoN
session 2026-09-17: from an optree, origins are per-*statement*, because
`nextstate` carries a line and never a column, and because perl constant-folds
before that walker runs. Statement granularity is enough for the four compiler
arguments above and not enough for column-accurate hover. Lowering from a CST
has neither limit — real spans with columns, and folding at a time of this
project's choosing — so the shared shape is `(file, start, end)` with a
line-only span as the degenerate case. The merge itself is real at
sub-statement scale either way: `($x+1) * ($x+1)` reaches that walker as two
`add` ops with distinct temporaries and hash-conses to one node, so the origin
set has two members that an optree cannot label and this CST can.

**What PSC does with that.** The lowering is a separate Go implementation
against this CST, not a consumer of that graph. What is shared is design:
the node-kind names, the type contract of item 6, and origin-sets-at-
construction. PSC's lowering does not intern in v1 — the compilation payoff
that justifies merging is one PSC does not collect — which makes every origin
set a single element, the degenerate case of the same representation. If
interning later proves worth it for memory, nothing in the shape has to
change.

**These are parser defects under any option.** The lowering does not repair
them; the parser must stop erasing structure before there is structure to
lower. Issue 01a0ad52 tracks the dereference leaf. The paren erasure and the
subscript-shape inconsistency have no issue as of 3b9422eb (§6.1.8). The
subscript shape was fixed at d2ebe02a; the paren erasure is still open.

**What the repo already decided in miniature.** Issue 01a0ad52 proposes
splitting `Term` into `Var{Sigil, Name}` and `Deref{Sigil, Expr}` and cites
the consumer's vocabulary as the reason — "`internal/infer` already switches
on `scalar` (10 uses), `array` (10), `hash` (9) as distinct kinds". That is
consumer-driven tree design, which is the lowering argument applied inside
the parser. This section extends it: the sigil split belongs in the CST
because it is structure the lexer already knows; the `Arrow` flag and the
trivia filter belong in the lowering because they are variation.

**What would change the decision.**

- A measurement that the lowering pass does not fit the budget. §6.10.2
  leaves `Analyze` at 0.5-2 ms against a 10 ms p99; a pass that walks 40,000
  CST nodes once and allocates a smaller tree should cost the same order. If
  it does not, the fallback is Pyright's shape — normalise at parse time into
  flags on CST nodes (§6.1.8) — not (c).
- A measurement that the LSP's type-at-position path through the side table
  is too slow or too fragile. rust-analyzer's syntax→HIR direction is the
  expensive one (`hir/src/semantics.rs analyze_impl`: find the enclosing body,
  lower it, reverse-map; it panics on a node from a different tree instance).
  The mitigation is that the unit is a sub, which is small; if that fails,
  Sorbet's push-the-query design is the verified alternative.
- The parser stopping at a vocabulary close enough to §4.14 that the lowering
  is the identity plus trivia removal. Then the pass is a filter, and the
  side table degenerates to a same-index map. That is a cheaper (b), not a
  reason for (a).

### 6.1.7 Flow analysis over a tree, not a sea of nodes

The stated architecture for PSC is a CST, then type inference over it with
Chalk's lattice, "using a flow analysis system similar to B::SoN" but over a
syntax tree rather than a sea-of-nodes graph. The research pass asked every
system whether a tree-shaped checker builds a separate flow graph anyway.

| System | Flow structure | Built over | Walk |
|---|---|---|---|
| TypeScript | Separate graph: `FlowNode{Flags, Node, Antecedent, Antecedents}` (`ast/flow.go:27`; `binder.ts:506 createFlowNode`, `:1351 createBranchLabel`, `:1355 createLoopLabel`) | The parser's tree, with `Node` back-pointers | Backwards over `antecedent` from a reference (`checker.ts:28264 getFlowTypeOfReference`) |
| Pyright | Separate "reverse code flow graph" per scope (`docs/internals.md`; `codeFlowTypes.ts`: "largely based on the code flow engine in the TypeScript compiler"), 20 flag kinds | The parse tree | Backwards, lazily, with `affectedExpressions` to skip unaffected branches |
| Sorbet | A CFG of 17 instruction kinds (`cfg/Instructions.h:37`) | The lowered `ast::Expression` | Forward, single pass, topologically sorted blocks, no fixpoint, non-SSA (`docs/internals.md`) |
| rust-analyzer | MIR basic blocks (`hir-ty/src/mir.rs:1034-1067`) for flow-sensitive checks only | HIR | Ordinary inference stays a tree walk over HIR (`diagnostics/unsafe_check.rs:266 walk_expr`) |
| Roslyn | **None.** `AbstractFlowPass : BoundTreeVisitor` (`FlowAnalysis/AbstractFlowPass.cs:36`), "a fairly conventional lattice flow analysis" | The *bound* (canonical, paren-erased) tree | Re-visit to a fixed point: `do { Scan() } while (stateChangedAfterUse)`, non-local edges via `PendingBranches` (`:392-403`, `:421-451`, `:773-792`). Scales to definite assignment and nullable analysis (`DefiniteAssignment.cs`, 125 KB) |
| B::SoN | **None separate.** `control_in` edges, `Region` nodes for joins, `Phi` for merges, all inline in the graph | The graph itself | Fixpoint over the graph (`B::SoN.pm`). Reported by that project 2026-09-17 |

Four of six build a flow structure separate from the syntax tree; two do
lattice dataflow by visiting their own IR to a fixed point. The second of
those is this project's sibling: B::SoN's inference is a fixpoint over its
graph with control flow inline, which is Roslyn's shape rather than
TypeScript's. What that means for this design:

1. **The syntax-tree-vs-sea-of-nodes framing is a false choice.** Every
   system above does flow analysis over a syntax tree; the flow structure is
   either a separate graph whose nodes point back into the tree (four cases)
   or a fixed-point visitor over the tree itself (one case). A graph-shaped
   flow analysis layered over a tree is the mainstream design, not a
   compromise. B::SoN's sea-of-nodes is the *IR* being analysed; here the IR
   is the lowered tree of §6.1.6 and the flow structure is whichever of the
   two shapes measures better.

2. **Whichever shape, it is built over the lowered tree, not the CST.**
   Roslyn's tree-shaped analysis works because it runs on the bound tree,
   after paren erasure; Sorbet's CFG is built from the desugared tree. The
   canonical vocabulary is what makes a tree walk tractable. This is the
   strongest single reason the lowering and the flow analysis are
   complements, not alternatives.

3. **Attach flow state narrowly.** TypeScript moved `flowNode` off the base
   `Node` onto a `FlowContainer` interface that ~30 node kinds opt into
   (`types.ts:950-951`, `:970-976`; PR 51682) — a memory lesson recorded as a
   code comment. This CST has more nodes than a TypeScript AST because it
   holds trivia, so a per-node field costs proportionally more. If flow
   nodes attach to anything, they attach to variable references,
   conditionals, loops and `return`/`last`/`next` in the lowered tree.

4. **The cheapest working point is documented.** Sorbet visits each binding
   once, with "no backsolving for types or iterating until a fixed point, and
   ... no constraint generation plus unification step", and accepts that a
   loop-carried variable cannot change type (error 7001). That is production
   value from a single forward walk. Start there; reach for Roslyn's
   fixed-point re-scan or SSA only when a measured case needs it.

5. **The narrowing shapes transfer, unevenly.** TypeScript's binder triggers
   narrowing on a hard-coded set of syntactic shapes (`binder.ts:1237
   isNarrowingExpression`) and the checker dispatches to 24 `narrowTypeBy*`
   rules (`checker.ts:29425 narrowType`). For Perl:

   | Perl guard | TypeScript analogue | Transfer |
   |---|---|---|
   | `defined $x`, `if ($x)` | `narrowTypeByTruthiness` (`:28839`) + `TypeFacts` | Clean |
   | `ref $x eq 'ARRAY'` | `narrowTypeByTypeof` (`:29051`), which pattern-matches `typeof x === "s"` in *both* operand orders (`:28898`, `:28901`) | Clean; `'ARRAY' eq ref $x` must be handled too |
   | `$x->isa('Foo')`, `->DOES` | `narrowTypeByTypePredicate` (`:29397`) | Weakest: TypeScript needs the callee's signature annotated `x is T`; no Perl module declares that. `isa`/`DOES` are special-cased as built-in predicates, the way `hasOwnProperty` is (`:29386`) |
   | `unless`, statement modifiers, `//=`, `||=` | No analogue; `narrowTypeByBinaryExpression` handles `&&`/`||`/`??=` shapes | Erased or normalised by the lowering before flow sees them |

6. **Flow analysis is its own package.** The Go port of TypeScript, under a
   faithful-port constraint, still put narrowing in its own `flow.go`. Start
   `internal/infer/`'s flow layer as a distinct package from both the parser
   and the lowering.

Whether the flow structure is a separate graph or a fixed-point visit over
the lowered tree is not decided here; §6.1.8 lists it. The owner's stated
preference — tree-shaped — has one production precedent, and that precedent
runs on a canonical tree, which this design now has.

### 6.1.8 Open after the research

The research pass settled the syntax axis and left the following open. None
is resolved by fiat here; each names what would settle it.

| Question | State | What settles it |
|---|---|---|
| Flow structure: separate graph (TypeScript, Pyright, Sorbet, rust-analyzer MIR) or fixed-point visitor over the lowered tree (Roslyn)? | Open. Both verified in production; the owner prefers tree-shaped. | Implement the single-pass Sorbet-style walk over the §4.14 tree first; measure whether the narrowing rules in §6.1.7 item 5 need antecedent edges the tree walk cannot express. |
| LSP type-at-position: pull (offset → CST node → side table → AST node → type; Roslyn, rust-analyzer, Pyright) or push (install a location query and re-run inference; Sorbet)? | Open. Pull is the default under §6.1.6 item 3; Sorbet's push is the verified fallback. | The measurement §6.1.6 names under "what would change the decision". |
| The exact lowered vocabulary: which surface distinctions survive as nodes, which as flags, which are erased. | **Closed** by the reconciliation in chapter 4 §4.14.2-§4.14.5 at 3c997a6b: every kind carries a status (SHIPPED / LOWERING / PARSER / SPECULATIVE) and its evidence; §4.14.3 does the same for the flags; §4.14.5 is the erasure list, each item VARIATION or STRUCTURE. `unless` is VARIATION as a branch swap (`B::Deparse` gives `if (c) {B} else {A}`); the `if`/`unless` modifier is VARIATION by child order; `do {} until` is STRUCTURE (perl: runs once where the modifier form runs zero times). The count asked for here was taken for `Arrow`: zero `internal/infer` rules read it by that name, and `inferElementType` (`infer.go:1270-1280`) reads the deref-ness it stands in for. | What §4.14.6 still lists: `CmpChain`'s typing rule (01a1335d-9738); `Interp`/`Match` pieces (01a12059-3011); `field` as a `Decl.Kind` inference uses (01a14248-520e); the statements lowered to a `Hole` (01a14248-51aa). |
| Reconciling §4.14 against the shared naming vocabulary (§6.1.6). | **Closed** by chapter 4 §4.14.4, from a direct read of `~/dev/perl5-son` rather than a report. The test stands: a kind is a *language* fact if a Perl programmer can point at source and say "that's what this is", an *artifact* if it exists only because the IR is in SSA/graph form. `ArgsSource`, the `Cell*` group and `Wantarray` are language, as recorded here before. `Coerce` is **not**: B::SoN's own `docs/plans/2026-09-17-the-gap-shape-is-inverted.md` uses `Coerce(Scalar -> Code)` as the stand-in for "compile and run arbitrary perl ... there is no node for that" — it is where that IR puts what it cannot name. The artifact list is ten, not five: `Phi`, `Region`, `Proj`, `Start`, `MemStart`, `StructRef`, `StructFieldAccess`, `ExpressionList`, `ListAppend`, `Coerce`. Result: one exact name agreement (`Wantarray`); one B::SoN name adopted by ruling (`ArrayLiteral`/`HashLiteral`); `Index`/`Key`/`Slice{Kind}` and `Paren` kept where the sibling is coarser or silent, as phase differences; `StrLit.Backtick` and a `tr`/`y` note on `Match` added from the sibling's `BacktickExpr` and `Transliterate`. | Whether `TypeLibrary` keys on node kinds, UNVERIFIED, which would make the `Subscript`-vs-`Index`/`Key` split a contract question. |
| Incremental `Analyze`: needed, or does whole-file re-run at 0.5-2 ms hold? | Deferred, per §6.10.5 item 3 and step 8 of §6.11. | Step 4 of §6.11. If it is needed, the first step is an `ItemTree`-shaped per-file summary (§6.1.4), not a query engine. |
| The stable node-identity scheme for §6.1.6 item 4. | **Closed**: `lower.ID` (`internal/lower/id.go`), an anchor naming the innermost enclosing named sub by kind, name and disambiguator in the manner of `crates/span/src/ast_id.rs`, plus a preorder index local to that anchor and a `Role` for a node synthesized beside its parent. Identity is per *syntax site*, as chapter 4 §4.14.1 rule 6 fixes from B::SoN's `MakeCell.pm`, so a table keyed on a captured value keys on (id, capture instance). `internal/infer` keys its annotations on it. History, measured at 3c997a6b on the tree-sitter path: the byte-offset key it replaced let the last writer win at a shared offset — on `my $x = 5; my @y = ($x); my $z = $y[0];` byte 0 carried four named nodes, and the `number` inside `(1..5)` read back as its parent's `Str`. | Nothing; the acceptance test asked for is `TestIDStableUnderEditAbove` and `TestAnnotationsSurviveEditAbove` (one statement inserted above two subs changes no id inside either), with `TestAnnotationsKeyedByIdentity` for the collision. |
| §6.1.6 item 7, "`Unknown` is checked at the head of every rule that descends". | The hazard it guarded did not occur when measured at 3c997a6b: every one of T1's 4,944 `Unknown` nodes is a childless leaf (0 with children, 0 nested), covering 29.1% of corpus bytes, standing under `source_file` (3,851), `block` (690), `binary` (281), `index` (74), `loop` (36) and five rarer parents. Nothing descends into a leaf. Chapter 4 §4.14.2 replaces the per-rule discipline with one node: the lowering emits a `Hole` per `Unknown`, inference types it `Any` and never inspects it, and a position query inside its span returns the `Hole` so the LSP reports "no information" rather than a type. A `Hole` may stand anywhere a `Node` may — the census shows it as a `binary` operand and an `index` base. | Done: item 7 now says so, and `TestHoleStandsAnywhereANodeMay` checks the `Node` fields (§4.14.6). |
| Three consumer defects the lowering must fix rather than inherit. | **Fixed** on the lowered path. History, measured at 3c997a6b on the tree-sitter path: `my @r = (1..5);` inferred `@r` as `Str` (`findOperatorText` matched the anonymous `.` before `..`); `my @a=(1,2); my @s=@a[0,1];` gave `@s elem=Unknown` where `@a elem=Int`, because no slice grammar name appeared in `internal/infer`; and the byte-offset collision above. Today `@r` and `@s` both carry element type `Int`, and annotations are keyed by `lower.ID`. | Closing 01a121c3-b0bc-7e12-a196-1d1fd563ab80, which still tracks the first against the retired path. |
| Latency of `internal/parse/`, the lowering and `Analyze` on the §6.10.2 files. | Unmeasured. The parse column there is the retired tree-sitter parser's. | A Go benchmark of `parse.Parse`, `lower.Lower` and `infer.Analyze` over `charnames.pm`, `FileHandle.pm`, `overload.pm`, `sigtrap.pm`, `_charnames.pm` (01a14248-51da-7d76-bc68-4f7dbeb842b1). |
| Paren erasure at `term.go:284-293`; `{k}` in `$h{k}` as `Call{Resolved:false}` rather than a quoted bareword. | Live defects, measured at 3b9422eb and re-measured at 3c997a6b; no issue tracks either. Canonical re-emission (issue 01a0a8d8), the only planned check that would catch the first, is blocked behind the M1 gate (01a0a70f). The third defect of this row, `$h{k}` and `$h->{k}` producing different node shapes, was fixed at d2ebe02a: `->` before a subscript opener now routes to `parseSubscript`, so both forms are one `Index`. Measured exactly neutral — 856 clean and 8,432 `Unknown` before and after on the corpus as then counted; 408 and 4,944 on T1 proper, also unchanged — which is the expected result for a change to the shape of trees that already parsed. **But the fix dropped the bit its own comment promised** (`expr.go:73`): the two forms are now byte-identical trees, `subscript_test.go:34-37` asserts they must be, and they read different variables (`%h` versus `$h`; chapter 4 §4.14.1 rule 2). §4.14 now classifies both defects as PARSER — the lowering cannot recover a bit the parser discarded — and names the fix: `Paren`, `Arrow` and a filehandle-slot bit on `parse.Node`, with `Resolved` as precedent. The same reconciliation found five more parser-side losses of the same kind: the fat comma inside `list`/`anon_array`/`anon_hash` (`(a, 1)` and `(a => 1)` parse identically); `1..2` lexed as one `Number`; `<<>>` not lexed; `qx//` not lexed; `delete`/`exists` absent from `keyword.go`. And four wrong trees that round-trip: `sub { 1 };` in statement position, `{a=>1};` at statement start, `*{"foo"}`, and a signature body under `use v5.36`. | Issues filed for each; the §4.16 checkboxes turned into tests — parenthesisation survives, `$h{k}`/`$h->{k}` differ in exactly `Arrow`, `(a, 1)`/`(a => 1)` differ. |
| The dereference leaf (issue 01a0ad52), attempted and withdrawn. | A `DerefSigil` token puts the interior of `${$h->{k}}`, `@{[ ... ]}` and `$$x` in the tree, and that part works. Measured on T1 it is worse on both axes: 408 → 406 clean, 4,944 → 4,986 `Unknown`. An earlier measurement of the same change looked favourable (−277 `Unknown`) and was taken against the wrong corpus — 1,935 files including `module/`, where three Class-MethodMaker files supplied 335 of 345 improvements. None of them is in T1. Splitting the sigil off also exposed a latent bug: `expect.go`'s CloseBracket case applies perl's right-paren lookahead to EVERY closer, though its own comment cites `yyl_rightparen` (toke.c:7156), which handles `)` alone, so `${$y}{Keys}` classifies its subscript brace as a block. That is brace classification, which is 01a0ac52's subject. | 01a0ac52 first. The work is stashed and the issue carries the SHA, both measurements, and acceptance criteria requiring that no T1 file regress. |
| Whether import resolution (issue 01a0ad3f) should keep an offline symbol-table dump as a last resort. | The issue rejects reading Perl's symbol table as circular. Sorbet shipped exactly that design (`srb rbi hidden-definitions`), found it not circular when run as a separate offline step, and retired it for operational reasons instead — slow, side-effecting, fragile to one bad file, and it "never generates types" (`website/blog/2022-07-27-srb-tapioca.md`) — while keeping it in maintenance mode because "sometimes there's simply no other tool". Pyright's `docs/type-stubs.md` gives the XS-shaped reason for declarations over inference: "Some libraries are thin shims on top of native (C++) libraries. Little or no type information would be inferable". | The issue's rationale rewritten on operational grounds, which are the ones that hold; whether to keep the dump as a last resort is a product decision. |


---

## 6.2 The Document Model

### 6.2.1 Recommendation: `[]byte` plus a line table. Not a rope.

perl-lsp uses `ropey::Rope` (`state.rs:9`, `reparse.rs:8`). Go's standard
library has no rope, and you would be writing one.

Do not. Use this:

```go
// Document is the text of one open buffer and the byte offset of each line
// start: Lines[0] == 0, and only '\n' starts a line.
type Document struct {
    Text  []byte
    Lines []int32
}
```

**Implemented** in `internal/psc/document.go` as `Document`, built by
`NewDocument(text []byte) *Document`. The LSP server's per-uri `document`
(`internal/psc/lsp.go`) holds one, and the uri lives there, so `Document`
carries no `URI`; it carries no `Version` until a version fence (§6.2.3)
needs one. `Text` need not be valid UTF-8: invalid bytes count as one
character each.

A rope buys O(log n) edits instead of O(n). Three reasons that is the wrong
trade here: a 5,000-line file is ~150 KB, so the memmove is 5–10 µs against a
10,000 µs budget; every consumer downstream wants a flat slice anyway (PSC
takes `source []byte`, `infer.go:22`), so a rope only adds a flatten; and a
rope of 150 KB is thousands of interior objects for the GC to trace per open
document, versus one.

Apply an edit with a single three-way splice:

```go
// ApplyEdit replaces Text[start:oldEnd] with newText and returns the new
// end offset. start and oldEnd MUST be on UTF-8 boundaries (see §6.6).
func (d *Document) ApplyEdit(start, oldEnd int, newText []byte) int {
    newEnd := start + len(newText)
    tail := d.Text[oldEnd:]
    need := start + len(newText) + len(tail)

    if cap(d.Text) >= need {
        // In-place: shift the tail, then write the new bytes.
        d.Text = d.Text[:need]
        copy(d.Text[newEnd:], tail)
        copy(d.Text[start:], newText)
    } else {
        buf := make([]byte, 0, need+need/4) // 25% headroom for growth
        buf = append(buf, d.Text[:start]...)
        buf = append(buf, newText...)
        buf = append(buf, tail...)
        d.Text = buf
    }
    d.reindexLinesFrom(start)
    return newEnd
}
```

Note the `cap` check. Typing one character at a time repeatedly hits the
in-place branch, so the steady-state cost of a keystroke is one memmove of the
tail and zero allocations.

### 6.2.2 The line table, and why you rebuild only the suffix

perl-lsp's `LineIndex::new` (`perl-line-index/src/lib.rs:22-30`) rebuilds the
entire `line_starts` vector on every construction, and
`apply_text_edit_to_state` (`reparse.rs:39-40`) calls it on every edit,
alongside `Rope::from_str(&state.source)` — a full rope rebuild too. Both are
O(n) per keystroke. That is affordable at 150 KB, but it is free to do better,
because line starts before the edit never change:

```go
// reindexLinesFrom rebuilds line starts from the line containing `from`.
// Line starts before that point are unaffected by an edit at `from`.
func (d *Document) reindexLinesFrom(from int) {
    // Find the first line whose start is > from; keep everything before it.
    keep := sort.Search(len(d.Lines), func(i int) bool {
        return d.Lines[i] > int32(from)
    })
    d.Lines = d.Lines[:keep]

    scan := 0
    if keep > 0 {
        scan = int(d.Lines[keep-1])
    } else {
        d.Lines = append(d.Lines, 0)
        scan = 0
    }
    for i := scan; i < len(d.Text); i++ {
        if d.Text[i] == '\n' {
            d.Lines = append(d.Lines, int32(i+1))
        }
    }
}
```

**Implemented** as `(*Document).reindexLinesFrom` in
`internal/psc/document.go`, with this signature, scanning with
`bytes.IndexByte`; `ApplyEdit` keeps the signature above.

`int32` line starts cap a document at 2 GB. Take the memory: a 5,000-line file
costs 20 KB of index instead of 40 KB, and it stays in L2.

**CRLF:** follow perl-lsp exactly. Only `\n` starts a line; `\r` is an ordinary
byte belonging to the preceding line (`perl-incremental-parsing/src/lib.rs:66-73`
asserts this). But columns must *exclude* the `\r`, or a click at end-of-line
on Windows lands one column too far right — see `line_content_end`
(`convert.rs:7-10`) and the regression test
`utf16_helpers_exclude_crlf_from_line_columns` (`convert.rs:171-179`). Strip a
trailing `\r` when computing column width, never when computing line starts.

### 6.2.3 Snapshots and the version fence

The parser must never read a `Document` that a `didChange` is mutating. Solve
it by never mutating a shared document from two goroutines — see §6.7 — and by
handing the parser an immutable snapshot:

```go
// Snapshot is an immutable view of a document at one version. Safe to share
// across goroutines; never mutated after construction.
type Snapshot struct {
    URI     string
    Version int32
    Text    []byte // aliased, never written after publication
    Lines   []int32
    Tree    *Tree  // parse result for exactly this Text
    Tokens  []Token
}
```

The rule that makes this sound: **once a `Snapshot` is published it is
frozen.** `ApplyEdit` mutates the *live* `Document`; producing a snapshot
copies the header and, when `ApplyEdit` took the in-place branch,
copies `Text`. To keep that copy off the hot path, use copy-on-write: mark the
document dirty on edit, and copy only when a snapshot is actually taken. In
practice most keystrokes are superseded before any snapshot is requested
(§6.7.3), so the copy rarely happens.

---

## 6.3 The Token Cache

Everything incremental hangs off the token cache. Get it right and node reuse
is a bonus; get it wrong and nothing else can help.

```go
// Token is one lexical token. It is the unit of incremental reuse.
type Token struct {
    Kind  TokenKind
    Start int32 // byte offset, inclusive
    End   int32 // byte offset, exclusive
    // Flags carries lexer decisions that Kind alone does not capture,
    // e.g. "this `/` was a regex, not division". Two tokens with equal
    // Kind and different Flags are NOT interchangeable for reuse.
    Flags TokenFlags
}
```

`Flags` is the Perl-specific part and it is not optional. perl-lsp's resync
loop compares only `token_type`, start, and end (`reparse.rs:86-96`). For Perl
that is not enough: a `/` lexed as division and a `/` lexed as the start of a
regex can share a kind in a coarse token enum and must never be treated as
equal. Encode the disambiguating decision in `Flags` and compare it.

### 6.3.1 Lexer state, and the checkpoint

`LexState` is the complete resumable state of the Perl lexer. perl-lsp's
version (`perl-lexer/src/checkpoint/core.rs:5-35`) is the best available
*inventory*, but it carries a two-state mode plus ad-hoc flags (`after_sub`,
`after_arrow`, `hash_brace_depth`, `after_var_subscript`) that the
eleven-state `Expect` and the bracket stack replace (chapter 3 §3.9.4), and it
omits the heredoc queue and the sublex stack (chapter 2 §2.14.1). The union,
in Go:

```go
// LexState is everything the lexer needs to resume mid-file. If two lex
// states are Equal, lexing forward from the same byte produces identical
// tokens. This is the entire correctness basis for incremental re-lexing.
type LexState struct {
    Expect         Expect        // PL_expect, all eleven states — ch3 §3.1
    BrackStack     []Expect      // PL_lex_brackstack: what `}` restores — ch3 §3.1.4
    LexMode        LexMode       // PL_lex_state: NORMAL / INTERP* / FORMLINE — ch2 §2.2.3
    SublexStack    []SublexFrame // interpolation contexts — ch2 §2.2.4
    DelimiterStack []rune        // nested delimiters in s{}{} etc.
    ParenDepth     int32         // guards heredoc-vs-bitshift
    HeredocQueue   []HeredocSpec // pending heredocs, in order — ch2 §2.9.4
    LastLopOp      OpCode        // PL_last_lop_op: the sort / filehandle cases — ch3 §3.4.4
    AfterLop       bool          // the token just consumed was a list operator (PL_oldoldbufptr == PL_last_lop): `{` — ch3 §3.1.4 default row, §3.3 step 3; `$fh` — §3.1.2 XOPERATOR
    AfterUni       bool          // ... or a named unary (PL_last_uni): `(` — ch3 §3.1.2 XTERM
    UTF8           bool          // `use utf8` in effect — ch2 §2.3.1
    Features       FeatureBits   // the `'` package separator reads one — ch2 §2.6.3
    Context        LexContext    // Normal | Heredoc | POD | Format | Regex | QuoteLike
}

// Equal reports whether two states will produce identical forward lexing.
func (a *LexState) Equal(b *LexState) bool { /* field-wise, incl. slices */ }

// Checkpoint is a resumable lexer position.
type Checkpoint struct {
    Byte  int32
    State LexState
}
```

**Where to place checkpoints.** perl-lsp records them after `;`, `{`, `}`, and
before `sub`/`package` (`lex.rs:16-40`). That is the right instinct, but its
implementation hardcodes `mode: LexerMode::ExpectTerm` at every checkpoint
rather than capturing the lexer's real state — which is exactly why it can only
ever be a heuristic. Capture the *actual* state, and add the strongest
condition:

> A byte offset is a **safe checkpoint** if and only if `Context == Normal`,
> `Expect == XSTATE`, and `BrackStack`, `SublexStack`, `HeredocQueue` and
> `DelimiterStack` are all empty.

Record a checkpoint at every such offset that follows a `;` or `}` at brace
depth 0, subject to a minimum spacing so a 5,000-line file holds hundreds, not
thousands:

```go
const checkpointStride = 512 // bytes; ~10 lines of Perl

func (tc *TokenCache) recordCheckpoints(toks []Token, states []LexState) {
    tc.Checkpoints = tc.Checkpoints[:0]
    tc.Checkpoints = append(tc.Checkpoints, Checkpoint{Byte: 0, State: initialLexState()})
    last := int32(-checkpointStride)
    for i, t := range toks {
        if !isStatementEnd(t.Kind) || !states[i].isSafe() {
            continue
        }
        if t.End-last < checkpointStride {
            continue
        }
        tc.Checkpoints = append(tc.Checkpoints, Checkpoint{Byte: t.End, State: states[i].Clone()})
        last = t.End
    }
}
```

Roughly 300 checkpoints for 150 KB, each maybe 100 bytes with its slices —
30 KB per document. Cheap.

`isSafe()` is where the Class A triggers get their teeth: inside a heredoc body,
inside POD, inside `s{}{}`, the state is not safe and no checkpoint is
recorded. Re-lexing therefore always restarts from a point where the lexer's
world is fully known.

**As landed** (`internal/lexer/checkpoint.go`). The lexer `State` is not the
field list above but the lexer struct itself with its cursor, source and output
cleared, so every field the lexer grows is state without a hand-kept list;
`Clone` deep-copies it (a reflect test fails on a reference field it misses)
and `Equal` compares clones, an empty collection equal to a missing one.
`StartState`, `StartStateBarewords` and `StartStateTyped` are the byte-0 states
of the three `Tokenize*` entry points, which start from them. This lexer has no
`SublexStack`, `DelimiterStack` or `Context`: quotes, regexes, `s{}{}`, POD,
heredoc bodies and format bodies are each one token scanned within one step,
so no step boundary falls inside them. `Checkpoints(src, start)` returns the
tokens with the checkpoints passed — byte 0, then at most one per 512 bytes —
and `Resume(src, cp)` lexes from one to EOF, equal to the full stream's suffix.
No checkpoint is recorded unless the token just emitted is a `;` or a `}` and
`Expect == XState` with the bracket stack empty, no heredoc queued (so never
with one pending on the line, nor inside its body), and no format body pending
or open, and no `[` still open by count (`openSquares`) — any closer pops the
bracket stack, so `[ )` empties it while typed Perl's `containerTypeSigil`,
walking back over raw `[`/`]` with no stop at `;`, would still reach that `[`;
nothing follows a `DataSection`, so none falls after `__END__` or
`__DATA__`. Nor is one recorded before `peekEnd`, the furthest byte a
failed lookahead read and decided a token by: a `<` in term position seeking
its `>` to the end of the line, a prototype seeking its `)` to the end of the
line, a quoted heredoc terminator seeking its quote. Each reads past the
token it emits and may cross a `;`, so the tokens before such a byte depend
on the bytes after it, and an edit there would change them (review of #510).
Every other lookahead stops at the first byte of the next token, or reads
inside a bracket or heredoc still open, where no checkpoint falls. Resume needs the bytes before the checkpoint unchanged: the other
scanners look back a token or two, and none looks past a `;` or a block's `}`.

### 6.3.2 The cache

```go
type TokenCache struct {
    Tokens      []Token
    States      []LexState   // States[i] is the state BEFORE Tokens[i]
    Checkpoints []Checkpoint // sorted by Byte
}

// CheckpointBefore returns the last safe checkpoint at or before byte.
func (tc *TokenCache) CheckpointBefore(byte int32) Checkpoint {
    i := sort.Search(len(tc.Checkpoints), func(i int) bool {
        return tc.Checkpoints[i].Byte > byte
    })
    if i == 0 {
        return Checkpoint{Byte: 0, State: initialLexState()}
    }
    return tc.Checkpoints[i-1]
}
```

Storing `States` per token doubles token memory. Do it anyway — it is what lets
the resync loop (§6.5.3) compare lexer state and not just token kind, and it is
the difference between an incremental lexer that is correct and one that is
usually correct.

**As landed** (`internal/lexer/tokencache.go`). `TokenCache` holds `Tokens`
and `Checkpoints` and no per-token states: resync is tried only at old
checkpoints, which already carry the state, so the comparison §6.5.3 needs
costs nothing per token. `NewTokenCache(src, start)` lexes as `Checkpoints`
does. `Edit(src, start, oldEnd, newEnd)` re-lexes the new bytes from the last
checkpoint strictly before `start` (a scanner may peek one byte past its
token, so a checkpoint at the edit is not trusted), recording fresh
checkpoints as it goes, and resyncs at the first old checkpoint past the edit
where all of these hold: the old token ending there lies wholly after
`oldEnd`, it equals the new stream's last token in `Kind`, shifted
`Start`/`End` and Flags (the token's `OpensBlock` and `TermFollows`; a regex
and a division `/` already differ in `Kind` here), the lexer `State` equals
the checkpoint's, and no lookahead in the new stream has read past it — the
§6.3.1 `peekEnd` rule, which `State.Equal` cannot see because `peekEnd` is an
offset and is not state. So every checkpoint the cache holds, kept or fresh,
has tokens before it that depend on no byte after it. The cache is then
the old prefix, the re-lexed tokens and the old tail shifted by the edit's
delta, checkpoints likewise; a stream that never resyncs is lexed to EOF
once. `Edit` returns the count of tokens re-lexed. The tests hold the cache
to a fresh `Tokenize*` after every edit in all three start states, including
the §7.7.3 edits, the lookahead counterexamples and edit sequences, with
every checkpoint resuming to the fresh stream's suffix and no fresh lex's
lookahead reading past any of them; `FuzzTokenCacheEdit` holds it on any input
after an edit, a second at an independent offset and the second's undo. On
edit5k.pl the latency benchmarks' 1-char edit re-lexes 261 of 37,070 tokens.

---

## 6.4 The Parse-Node Cache and Dependency Tracking

### 6.4.1 The node representation PSC already needs

The migration surface is fixed by what `internal/infer/` actually calls. Here
is the measured usage across `internal/infer/*.go`:

| Method | Call sites |
|---|---|
| `Kind()` | 99 |
| `Child(i)` | 72 |
| `ChildCount()` | 61 |
| `IsNamed()` | 51 |
| `Text(source)` | 43 |
| `StartByte()` | 26 |
| `EndByte()` | 12 |
| `Parent()` | 6 |
| `IsError()` | 1 |
| `HasError()` | 1 |

That is the whole contract. Ten methods, all already declared in
`internal/parser/parser.go`. `NamedChild`, `NamedChildCount`,
`ChildByFieldName`, and `SExpr` exist in the wrapper but PSC barely touches
them — `ChildByFieldName` is worth keeping because it is the readable way to
express structure, `SExpr` is worth keeping for tests.

Independently recounted across `internal/infer/*.go` including tests, the
ordering and the conclusion hold: `Child` 119, `Kind` 112, `ChildCount` 109,
`IsNamed` 50, `Text` 47, `StartByte` 29, `EndByte` 13, `Parent` 6. The absolute
counts differ with scoping; the contract does not.

**The migration surface is ten methods and one vocabulary, and the vocabulary
is the larger of the two.** `internal/parser/parser.go` is a thin façade over
`gotreesitter` (its ABOUTME says exactly that), and the ten methods port to
any tree. But `Kind()` is one method returning 46 distinct tree-sitter-perl
grammar names across `internal/infer/*.go` (at 971b9031, by §6.1.4's
method), and `internal/parse/parse.go:146-201`
emits 23 kinds, none with those names (§6.1.4). An earlier draft of this
section concluded from the method count that PSC would not change — "not one
of its 4,265 lines" — and that was the wrong count. PSC's walkers are
rewritten against the lowered vocabulary §6.1.6 specifies; the ten methods
below are what the *CST* node keeps exposing, for the LSP's structural
features and for the lowering pass itself.

Kind-name strings in this chapter's sketches (`"package"` in `infer.go`,
`"subroutine_declaration"` in §6.5.2) are tree-sitter-perl's; the shipped
enum is `internal/parse/parse.go:146-201`, and the shipped node is
`parse.go:212-231` — `Kind`, `Start`/`End` as `int`, `Children []*Node`,
`Text`, `Resolved`, with no parent pointer, no flags and no arena. The layout
below is what §6.4–6.5 still ask of it:

```go
// Node is a CST node. The layout is chosen so PSC's ten hot methods are
// field reads, not computations.
type Node struct {
    kind     KindID // interned; Kind() indexes a string table
    start    int32
    end      int32
    parent   *Node
    children []*Node

    // flags packs isNamed, isError, isMissing, hasError, and the
    // "subtree contains an error" summary bit.
    flags NodeFlags
}

func (n *Node) Kind() string { if n == nil { return "" }; return kindNames[n.kind] }
func (n *Node) StartByte() uint32 { if n == nil { return 0 }; return uint32(n.start) }
func (n *Node) ChildCount() int { if n == nil { return 0 }; return len(n.children) }
func (n *Node) Child(i int) *Node {
    if n == nil || i < 0 || i >= len(n.children) { return nil }
    return n.children[i]
}
func (n *Node) Text(src []byte) string {
    if n == nil || int(n.end) > len(src) { return "" }
    return string(src[n.start:n.end])
}
```

Keep the nil-safety on every method. `internal/parser/parser.go` already does
this and `infer.go:26-29` relies on it (`if child == nil { continue }`).

**Allocate nodes from an arena**, one `[]Node` per tree, and hand out pointers
into it. A 5,000-line file is maybe 40,000 nodes; that is one allocation of
about 2 MB instead of 40,000 small ones, and it drops GC scan time by more than
the parse itself costs. `gotreesitter` does this (`arena.go`, and the
`ownerArena` refcounting in `parser_incremental_support.go:5-14`) precisely
because reused subtrees keep an old arena alive. If you splice subtrees across
trees, you inherit that problem: an old arena stays alive while any spliced
node points into it. The simple fix is to **copy spliced subtrees into the new
arena** rather than aliasing — a memcpy of the affected nodes, bounded by the
damage region, and it makes tree lifetime trivially correct.

### 6.4.2 Byte-shifting reused nodes: the underflow trap

perl-lsp hit this and filed it as #2471 (`reparse.rs:16-26`). Go's version:

```go
// shiftOffset applies a possibly-negative delta without wrapping.
// A naive uint32(int32(off) + delta) wraps to ~4 billion when a deletion
// near the start of the file makes delta more negative than off.
func shiftOffset(off int32, delta int32) int32 {
    v := int64(off) + int64(delta)
    if v < 0 {
        return 0
    }
    return int32(v)
}
```

Write the test perl-lsp wrote
(`reparse_offset_tests::shift_offset_clamps_negative_underflow_to_zero`,
`reparse.rs:165`): delete the first statement of a file, then assert every
resulting token and node offset satisfies `0 <= start <= end <= len(source)`.
That invariant catches the whole class.

### 6.4.3 Dependency tracking

Three caches — text, tokens, nodes — plus PSC's annotations. The dependency
edges are strictly one-directional, which is the only reason this is tractable:

```
Document.Text  --(re-lex from checkpoint)-->  TokenCache
TokenCache     --(re-parse from boundary)-->  Tree
Tree + Text    --(PSC Analyze)------------->  Annotations, Diagnostics, SymbolTable
```

Invalidation flows the same way. Never invalidate backwards.

```go
// DamageSet is what an edit invalidated, in each cache's terms.
type DamageSet struct {
    TextStart, TextEnd int32 // widened edit range in NEW coordinates
    FirstToken         int   // index into TokenCache.Tokens
    ReparseStart       int32 // byte offset of enclosing statement/sub start
    ReparseEnd         int32 // byte offset where resync succeeded
    FullFile           bool  // a Class A/B trigger fired; discard everything
}
```

For **PSC** specifically: its annotation map is keyed by node identity,
`map[lower.ID]types.Type` (`infer.Analyze`). An edit renumbers only the
named sub it lands in, so the keys inside every other sub survive it
(`TestAnnotationsSurviveEditAbove`). The tree-sitter path keyed the map by
`StartByte`, which every edit invalidated wholesale because every key after
the edit shifted; rust-analyzer's contributor guide rejects that key by
name, for this reason (§6.1.4).

**PSC re-lowers and re-analyzes per sub.** The step-4 numbers (plan §5 M2)
put lowering at 35 ms and `Analyze` at 21 ms p50 on edit5k.pl, so the
whole-tree re-run does not fit the 10 ms gate, and the LSP no longer does it
for an edit that stays inside subs. `analyzeEdit`
(`internal/psc/incremental_analyze.go`) reads the edit's damage off the old
and new trees (`damageFromTreeDiff`, §6.10.5 item 4): a set of damaged
top-level subs, or full. On a set, `lower.Relower` lowers only those subs and
shares the rest of the old lowered tree, and `infer.Reanalyze` walks only
those subs and the code outside every sub, which threads state into them;
each other sub is replayed from what its last walk recorded -- its scopes,
diagnostics and writes to the scopes around it -- and its `lower.ID`-keyed
annotations are kept. A replayed span is read off the edited tree, never
computed from the edit: a symbol's is its declaration's span there, and a
diagnostic moves as its sub's declaration did. Arithmetic cannot tell, at a
boundary, whether a node took the inserted byte in -- `my $e = (;` becoming
`my $e = (;;` grows the declaration -- so the code outside every sub is
walked even when no sub is damaged, as for trivia between two subs. Full --
a Class B edit, a non-trivia top-level edit, a parse-error change -- falls
back to the whole-tree path, as does a sub whose walk does not replay (one
in a loop or an unreached scope, a method). `TestIncrementalAnalyzeMatchesFresh`
and `TestIncrementalAnalyzeRealEditsMatchFresh` hold the result equal to a
fresh open: lowered tree and ids, annotations, diagnostics and symbols. The
damage source is the tree diff until Repair's damage region replaces it;
the interface is the set of subs or full either way. See §6.10.2 and
00-findings §0.8.

The exception is `ProjectIndex` (`internal/infer/project.go`): cross-file
analysis triggered from `useStmt` (`internal/infer/flow.go`). Those results
are cached per file path and must survive a keystroke in an unrelated buffer.
Invalidate a `ProjectIndex` entry only when *that file's* content changes.

---

## 6.5 Damage and Repair

The core loop. Given an edit, produce a correct new tree in bounded time.

### 6.5.1 Triage first, always

Before computing any damage region, classify. This is `mod.rs:38-56`'s policy,
tightened for Perl:

```go
// Strategy decides how much work an edit requires.
type Strategy int

const (
    StrategyReuseAll  Strategy = iota // whitespace/comment interior: shift offsets only
    StrategyLocal                     // re-lex + re-parse a bounded region
    StrategyFull                      // full reparse; incrementality abandoned
)

const (
    maxEditBytes  = 64 * 1024 // matches perl-lsp MAX_EDIT_SIZE (strategy.rs:1)
    maxLocalEdit  = 1024      // single edits larger than this go full
    maxLocalLines = 10        // newline count in inserted text
)

func triage(doc *Document, snap *Snapshot, e Edit) Strategy {
    if e.touchedBytes() > maxEditBytes {
        return StrategyFull
    }
    // Class A/B triggers: a full reparse is the only sound answer.
    if introducesGlobalTrigger(snap.Text, doc.Text, e) {
        return StrategyFull
    }
    if e.touchedBytes() > maxLocalEdit ||
        bytes.Count(e.NewText, []byte{'\n'}) > maxLocalLines {
        return StrategyFull
    }
    if isInteriorTrivia(snap, e) {
        return StrategyReuseAll
    }
    return StrategyLocal
}
```

`introducesGlobalTrigger` is the Perl-specific gate and it should be
deliberately over-eager. Scan **both** the removed old text and the inserted
new text — perl-lsp learned this the hard way; `is_edit_non_structural`
(`incremental_v2.rs:475-493`) checks old and new sides precisely because
*deleting* a `#` is as structural as adding one:

```go
// globalTriggers are byte sequences whose appearance or disappearance can
// change the parse of text arbitrarily far away. Over-approximate freely:
// a false positive costs one full reparse; a false negative costs a wrong
// tree, which is unbounded damage.
var globalTriggers = [][]byte{
    []byte("<<"),        // heredoc introducer (A2, A3)
    []byte("__END__"),   // A8
    []byte("__DATA__"),  // A8
    []byte("=cut"),      // A7
    []byte("use "),      // B2, B3, B7, B8
    []byte("no "),       // B2
    []byte("package"),   // B5
    []byte("sub"),       // B1, B4 — prototypes and predeclaration
    []byte("format"),    // A10
    []byte("BEGIN"),     // B6
}

func introducesGlobalTrigger(oldText, newText []byte, e Edit) bool {
    // Widen to full lines: `=pod` and `format` are only special at line start,
    // and a heredoc introducer's effect begins at the next newline.
    oldWin := lineWindow(oldText, e.Start, e.OldEnd)
    newWin := lineWindow(newText, e.Start, e.NewEnd)
    for _, t := range globalTriggers {
        if bytes.Contains(oldWin, t) != bytes.Contains(newWin, t) {
            return true
        }
    }
    // Quote and delimiter parity changed on this line → Class A1/A4/C1/C2.
    return quoteParity(oldWin) != quoteParity(newWin) ||
        braceDelta(oldWin) != braceDelta(newWin)
}
```

Yes, `my $sub_count = 1;` takes the full-reparse path because it contains
`sub`. Correct trade: a needless reparse costs milliseconds, a wrong tree costs
a bug report. Add word boundaries later, once a benchmark says the
false-positive rate matters.

`StrategyReuseAll` handles the genuinely common case: typing inside a comment
or a string body, or adding whitespace. perl-lsp's
`incremental_parse_whitespace` (`incremental_v2.rs:538-549`) shifts every node
offset and reuses the entire tree. That is right, with one correction — it must
verify the edit is interior to a single trivia token, not merely that the edited
*bytes* are whitespace, or inserting a newline that terminates a `#` comment
sneaks through.

**As landed** (`internal/parse/incremental.go`). Triage decides on
**tokens, not text**: matching source text always misses a spelling
(`use\tconstant`, `q|a|`). `triage(old, src, e)` lexes the old source and
the edited one, and `diffTokens` finds the longest prefix the two token
streams share (ending at or before the edit) and the longest suffix they
share once the old tokens are shifted by the edit's delta (starting at or
after it). Only the tokens between, on each side, are examined. It also reads
the tree the document's `Session` parsed, resolver and all. `Strategy` is
`StrategyReuseAll`, `StrategyLocal`, `StrategyFull`. The limits are
`maxEditBytes` = 64 KiB and `maxLocalEdit` = 1,024 bytes, each measured as the
larger of the bytes deleted and the bytes inserted, and `maxLocalLines` = 10
newlines in the inserted text: 1,024 bytes and 10 newlines are still local,
1,025 and 11 are full. In place of `introducesGlobalTrigger`, an edit is full
when the changed tokens on either side:

- **hold a trigger word**, a `Word` token reading `use`, `no`, `require`,
  `BEGIN`, `sub`, `package`, `class`, `format`, `__END__` or `__DATA__`
  (B1-B7, A8, A10), whatever whitespace follows it;
- **hold a trigger kind**: a comment (A9), POD (A7), a heredoc introducer or
  body (A2, A3), a format body, the data section, a prototype (B1), or a
  construct left unterminated;
- **do not nest**: their `(`, `[`, `{` and closers, counted as
  `refuseUnclosedOpener` counts them, must balance without ever closing more
  than they opened, which catches `}{` (C1, C2, C4);
- **reach outside the anchor** `widen` would choose (old side) or its shifted
  span (new side). The suffix runs to end of file, so this is the token
  stream resyncing at or before the anchor's end; a quote-like or regex that
  loses its closer (A1, A4, A5) re-lexes past it.

An edit **touching** (or abutting) a `use`/`no`/`require`, a `BEGIN` block, a
prototype, or a `sub`, `package` or `class` declaration anywhere but strictly
inside its block is also full (chapter 7 §7.7.3's prototype row): renaming a
sub changes only a `Word`, which no token rule sees.

**`StrategyReuseAll`** is exactly one token changed on each side, strictly
inside, into one of the same kind: whitespace, a comment or POD. Every other
token only shifts, so a newline that ends a comment, or whitespace that joins
or splits tokens, is not reuse-all. A string's body is **not** reuse-all,
though the paragraph above lists it: the string node's `Text` changes, so
shifting offsets alone would not equal a fresh parse.

### 6.5.2 Widening to a safe boundary

For `StrategyLocal`, widen the raw edit range outward until both ends sit on
something the parser can restart from.

**Backward** to the last safe lexer checkpoint (§6.3.1), then further back to
the start of the enclosing safe boundary. Chapter 5 §5.14 says which Perl
constructs those are — sub, method and `class` bodies, bare and control-flow
blocks, `package NAME BLOCK`; never the file top level, a `use`, a `BEGIN` or
a `package NAME;`. `isReparseAnchor` below is the node-kind projection of that
table; do not re-derive the table here.

**Forward** is not a fixed offset. It is wherever the resync loop succeeds
(§6.5.3) — you cannot know it until you have re-lexed.

```go
// widen expands an edit to a re-parseable region. Returns the byte offset to
// restart lexing from, its lexer state, and the enclosing node to replace.
func widen(snap *Snapshot, e Edit) (cp Checkpoint, enclosing *Node) {
    cp = snap.Tokens.CheckpointBefore(e.Start)

    // Find the smallest statement- or sub-level node containing the edit.
    enclosing = smallestContaining(snap.Tree.Root(), e.Start, e.OldEnd)
    for enclosing != nil && !isReparseAnchor(enclosing.Kind()) {
        enclosing = enclosing.Parent()
    }
    if enclosing == nil {
        enclosing = snap.Tree.Root()
    }

    // The lexer must start no later than the node we intend to replace.
    if int32(enclosing.StartByte()) < cp.Byte {
        cp = snap.Tokens.CheckpointBefore(int32(enclosing.StartByte()))
    }
    return cp, enclosing
}

// isReparseAnchor names the node kinds we are willing to re-parse whole:
// sub bodies (ch5 §5.14; §6.11 step 4 decides whether blocks follow), with
// "source_file" as the full-reparse fallback. Kind names are §6.4.1's.
func isReparseAnchor(kind string) bool {
    switch kind {
    case "subroutine_declaration", "source_file":
        return true
    }
    return false
}
```

The first implementation anchors at the enclosing **sub**, falling back to
`source_file` — a full reparse — for edits at top level. Subs are the unit
PSC scopes over, so a sub-level splice keeps the symbol table's shape stable,
and re-parsing a whole sub rather than its innermost block costs microseconds
against a budget measured in milliseconds (§6.10.2). Chapter 5 §5.14's table
also admits bare and control-flow blocks; narrow to them only if step 4 of
§6.11 shows sub-level re-parse missing the budget on oversized subs. Do not
anchor at single statements — chapter 5 §5.14 says why.

**As landed** (`internal/parse/incremental.go`). `widen(old, src, e)` returns
a `lexer.Checkpoint` and the anchor node. The anchor is the **top-level
statement** holding the edit (a run of trivia between statements counts as
one) when the edit touches no declaration in it (`touchesDeclaration`: a
`use`, a `BEGIN` block, a prototype, or a sub, package or class declaration
but strictly inside its block); otherwise it is the `SourceFile` root, a
full reparse. So an edit in plain
top-level code, a `package NAME BLOCK`'s own code or a sub body anchors at
its top-level statement, and one inside a `use`, a `BEGIN` (even inside a
sub declared there), a declaration's head or a `package NAME;` widens to the
whole file, as triage's `touchesDeclaration` already makes it
`StrategyFull`. A sub inside a top-level `package NAME BLOCK`, bare block or
`if` anchors at that whole statement, past the block: a nested parser
snapshot (§6.5.3) resumes only to the block's `}` and returns no post-block
state, while a top-level one resumes to end of file. Every anchor but the
root therefore holds a top-level parser snapshot at its start.

This departs from the design above, which anchors only at subs and says not
to anchor at single statements (issue 01a182e2-9623). The bookkeeping that
rule avoided is already paid: the tracked parse snapshots the parser before
every top-level statement whichever statements anchor, so a top-level
anchor adds no snapshot. Chapter 5 §5.14's syntactic test, that the
statement cannot alter parse state visible outside, is checked rather than
assumed: check 2 below compares the whole parser state after the statement
with the old parse's, so a statement that changes what later code reads
returns ok=false, and check 1 refuses one that reads into the next
statement or a statement typed where there was none. What neither check
sees is the statement *before* the anchor, which read the anchor's first
significant token to decide where it ends: `}` then `else` continues an
`if`, and a statement ended by end of file reads on into text typed after
it (a FuzzIncremental find, `0 ` edited to `0!0`). So triage returns
`StrategyFull` when the edit changes the anchor's first significant token
(`keepsLead`), and a run of trivia, which has none, is in effect never an
anchor. A sub anchor never trips this: its edit is strictly inside the body. The checkpoint is the last the lexer recorded at or before the
anchor's start, never after it; it comes from re-lexing the old source until
the token cache (§6.3.2) exists. A call whose prototype a module installed is
widened as the Session parsed it, and lies wholly inside its anchor.

**What Repair must still check.** Triage over-approximates, but it sees only
tokens, so the reuse it permits is safe only together with Repair's own
checks; Repair returns ok=false when any fails:

0. **A top-level parser snapshot starts at the anchor.** Repair resumes
   from it; when none starts at the anchor's start, it returns ok=false.
1. **The re-parsed anchor ends where the old one did.** After one statement
   is read from that snapshot over the edited source, the first unconsumed
   token must start at the anchor's shifted end (`End` plus the edit's
   delta), or the parser must be at end of input when the anchor's shifted
   end is the end of the edited source. A parse error can make a statement
   swallow the sub's `}` and what follows while every token is unchanged
   (`return` edited to `re%n` inside a sub does), and no token rule sees
   that.
2. **The parser state there equals the old one.** The parser after that
   statement must equal the old parse's snapshot after the anchor
   (`sameState`, with its token index moved by the change in token count),
   or the statements after the anchor would read differently: this is what
   carries a `use`, a sub declaration or a lexical declared inside the
   anchor to the rest of the file. Both parses use the Session's resolver,
   since that is the tree triage read: under one, a sub body edit can change
   the state a module's prototype left (`floor 1, 2` to `floor --, 2` after
   `use POSIX qw(floor)`) where a bare parse sees no difference. The token
   index in that comparison already implies check 1, except when the anchor
   is the last top-level statement and no snapshot follows it.
3. **Modules read the same bytes** (chapter 7 §7.7.3's prototype row): any
   module the old parse loaded that now resolves to different bytes. The
   exception is a tree with no snapshots, or snapshots taken under another
   resolver key (`tops == nil || old.inc.key != key`). There Repair
   re-parses the old source under the current modules to take them; that
   re-parse already reads the new bytes, so Repair goes on only if its tree
   equals the old one. Once checks 1, 2 and 4 and the `consulted` check
   also pass, the tree Repair returns equals a fresh parse.
4. **The EOF passes run again** over the spliced tree (`finish`, §6.5.3).

`TestTriageSweepNoLeak` and `FuzzTriage` (`internal/parse`) hold the
property on Session trees: for a random edit triaged `StrategyLocal` whose
Repair checks 0 to 2 pass (`repairAccepts`), or triaged `StrategyReuseAll`,
every top-level statement before the
anchor equals the fresh parse's and every one after it equals the fresh
parse's shifted by the delta (`TreeDiff`, `TreeDiffShifted`).

### 6.5.3 The repair loop

Now the heart of it. Re-lex from the checkpoint, resync against the old token
stream, re-parse the region, splice.

```go
// Repair performs a local incremental update. It returns the new snapshot,
// or ok=false if any step failed — in which case the caller does a full
// parse. Falling back is always correct; never try to salvage.
func (s *Server) Repair(ctx context.Context, snap *Snapshot, e Edit) (*Snapshot, bool) {
    cp, enclosing := widen(snap, e)
    delta := int32(len(e.NewText)) - (e.OldEnd - e.Start)

    // 1. Re-lex forward from the checkpoint over the NEW text.
    lx := NewLexer(newText, cp.Byte, cp.State)
    var fresh []Token
    var freshStates []LexState

    // Old tokens at or after the checkpoint are the ones being replaced.
    firstOld := sort.Search(len(snap.Tokens.Tokens), func(i int) bool {
        return snap.Tokens.Tokens[i].Start >= cp.Byte
    })
    // Candidates for resync start after the edit's old end.
    oldSyncFrom := sort.Search(len(snap.Tokens.Tokens), func(i int) bool {
        return snap.Tokens.Tokens[i].Start >= e.OldEnd
    })

    editEndNew := e.Start + int32(len(e.NewText))
    syncOld := -1
    last := cp.Byte
    budget := 0

    for {
        if budget++; budget > maxResyncTokens {
            return nil, false // ran too far; a full parse is cheaper
        }
        if ctx.Err() != nil {
            return nil, false // superseded by a newer edit
        }
        tok, st, eof := lx.Next()
        if eof {
            break
        }
        if tok.End <= last {
            return nil, false // lexer failed to advance — bail, do not loop
        }
        last = tok.End
        fresh = append(fresh, tok)
        freshStates = append(freshStates, st)

        // Only look for resync once we are past the edited region.
        if tok.Start < editEndNew {
            continue
        }
        for j := oldSyncFrom; j < len(snap.Tokens.Tokens); j++ {
            old := snap.Tokens.Tokens[j]
            if shiftOffset(old.Start, delta) != tok.Start {
                continue
            }
            if shiftOffset(old.End, delta) != tok.End ||
                old.Kind != tok.Kind || old.Flags != tok.Flags {
                continue
            }
            // THE PERL-SPECIFIC CONDITION: token identity is not enough.
            // Lexer state must match too, or forward lexing diverges.
            if !st.Equal(&snap.Tokens.States[j]) {
                continue
            }
            syncOld = j
            break
        }
        if syncOld >= 0 {
            break
        }
    }
    if syncOld < 0 {
        return nil, false // never resynced: the edit reaches EOF
    }

    // 2. Splice the token stream: [0:firstOld] + fresh + shifted tail.
    toks := make([]Token, 0, firstOld+len(fresh)+len(snap.Tokens.Tokens)-syncOld)
    toks = append(toks, snap.Tokens.Tokens[:firstOld]...)
    toks = append(toks, fresh...)
    for _, old := range snap.Tokens.Tokens[syncOld:] {
        old.Start = shiftOffset(old.Start, delta)
        old.End = shiftOffset(old.End, delta)
        toks = append(toks, old)
    }
    // (States spliced identically.)

    // 3. Re-parse just the anchor node's span from the new token stream.
    lo := shiftIfAfter(int32(enclosing.StartByte()), e.Start, delta)
    hi := maxInt32(shiftOffset(int32(enclosing.EndByte()), delta), last)
    sub, ok := ParseRegion(toks, newText, lo, hi, enclosing.Kind())
    if !ok || sub.HasError() != enclosing.HasError() {
        // Error status flipped: the damage is larger than we assumed.
        return nil, false
    }

    // 4. Splice the tree, copying into a fresh arena, shifting the tail.
    tree, ok := spliceTree(snap.Tree, enclosing, sub, delta, e.Start)
    if !ok {
        return nil, false
    }
    return &Snapshot{ /* ...new version, text, lines, tree, tokens... */ }, true
}
```

Five things in that loop earn their place:

1. **`maxResyncTokens`** (say 4,000). Without it, an unterminated string
   re-lexes the entire file token by token *and* runs the O(n) resync search at
   each one — quadratic, and it will find your 50,000-line file. Bail out and
   full-parse instead; that is strictly faster.
2. **`tok.End <= last`** — perl-lsp's non-advance guard (`reparse.rs:80-82`).
   A lexer bug on malformed input becomes an infinite loop without it. In a
   language server that is a hung editor.
3. **`st.Equal(...)`** — the condition perl-lsp omits and the one Perl requires.
   Two identical `/` tokens in different `Expect` states are not
   interchangeable.
4. **`ctx.Err()`** — checked in the loop, so a superseded parse dies promptly
   rather than finishing work nobody wants (§6.7). The implementation departs
   from the sketch here: the real `Repair` takes no `context` (§6.5.5).
5. **`sub.HasError() != enclosing.HasError()`** — if re-parsing the region
   changed whether it contains an error, the damage estimate was wrong.
   Bail out.

**The parser's state at the re-parse point.** The lexer's checkpoint (§6.3.1)
is half of what step 3 resumes from. The other half is the parser's own
file-wide state — the package, the features and lexical subs in scope, the
imports, the known packages and the declared subs — which it threads from
statement to statement. Step 3 restores it from a `parserSnapshot`
(`internal/parse/parsestate.go`), a copy of the `parser` struct taken before
each statement at a re-parse point. The re-parse points are each top-level
statement and each statement directly inside a file-level `package NAME BLOCK`,
`class NAME BLOCK` or bare block, which are the nested anchors widening may
choose; there is none inside a `BEGIN` block, a `use` or a `format`. A
snapshot holds copies of every map and slice, never aliases, so one snapshot
can be resumed from any number of times. `parserFields` classifies every field
of `parser` as snapshotted, per-statement scratch, or the resuming parse's own
(source, tokens and resolver: a re-parse uses its Session's resolver, never
the snapshot's), and `TestParserStateSnapshotCoversEveryField` fails until a
new field is classified. A resumed parse stops at end of file, or at the `}`
of a nested snapshot's block.

Two passes run over the whole file at EOF, and an edit anywhere can change
their result, so Repair runs them again over the spliced tree (`finish`):
`refuseUnclosedOpener`, which wraps the file's last token in
`Unknown{UnclosedOpener}` when any opener anywhere is unclosed, and the pass
that lifts the imports and every declared sub's prototype (`readModule`) onto
the root.

**As landed** (`internal/parse/incremental.go`). The signatures are
`Repair(old *Node, e Edit, src []byte) (tree *Node, ok bool)` and
`(*Session).Repair(old *Node, e Edit, path string, src []byte, roots ...string)
(tree *Node, ok bool)`, where `src` is the **edited** source. The resolver
form's tree equals a fresh `NewSession().ParseSourceFrom(path, src,
roots...)`; neither takes a `context` (§6.5.5). On ok=false the tree is nil,
never a partial splice, and `old` is never modified.

- **What it resumes from.** A parse's root carries an unexported cache
  (`inc`, which `TreeDiff` skips): a copy of the bytes it was parsed from,
  so Repair can check that `src` is exactly those bytes with `e` applied.
  The first Repair of a parse's tree parses the old bytes again, with the
  form's fresh resolver, for the top-level parser snapshots plus the parser
  at EOF, and goes on only if that parse equals `old` under `TreeDiff`. A
  tree Repair builds carries its own snapshots, the ones after the anchor
  shifted, so the next Repair re-parses one statement and nothing else. A
  cache built under a different resolver (path, roots) is rebuilt.
- **The re-lex.** The new token stream is the one triage lexed. There is no
  separate resync loop: triage compares whole token streams, and Repair
  returns ok=false when more than `maxResyncTokens` (4,000) tokens changed,
  so a runaway edit costs two linear lexes and no parse. Both still lex the
  whole source; the token cache (§6.3.2) replaces that.
- **The anchor.** `StrategyLocal` uses widening's anchor; `StrategyReuseAll`
  the top-level statement holding the edit. The whole file as anchor, or none,
  is ok=false.
- **ok=false cases.** `old` has no cache, or `src` is not its bytes with `e`
  applied (an edit outside the buffer included); triage says
  `StrategyFull`; any top-level statement other than the anchor holds an
  error (ruling R-f: no subtree with an error is reused); no top-level
  snapshot starts at the anchor and another at its end (check 0); the
  statement read from it does not end at the shifted end or leaves a
  different parser state (checks 1, 2); a module the cached parse read now
  reads different bytes, or is found where it was not (check 3, by SHA-256
  of every `Loader` call); where `old` has no snapshots under this resolver
  key there are no earlier reads, the SHA-256 comparison does not run, and
  check 3 is only that re-parsing the old source under the current modules
  gives `old`'s tree; the anchor's statement consulted the resolver at
  all, since a fresh parse would have read those modules with the prefix's
  already seen; or the re-parsed statement gains or loses an error.
- **The splice.** Every reused top-level statement is deep-copied, shifted by
  the delta when it lies after the anchor. No offset can underflow: only
  nodes at or after the anchor's end move, and that end, shifted, is at or
  after the edit's new end. Then `finish` runs over the spliced root (check
  4) from the EOF snapshot, and `loaded`, `declErrs` and `declWarns` are the
  old root's: the anchor read no module, so a fresh parse reads the same ones.
  Whether reused nodes are copied, shifted in place, or shared is ruling
  R-e's, measured by 01a17a40-4915.

### 6.5.4 The correctness test, and it is not optional

perl-lsp got this exactly right and it is the single most important test in the
chapter (`supported_batch_edits_match_fresh_parse`,
`incremental_boundary_regressions.rs:99`):

```go
// After ANY incremental update, the spliced tree must be identical to a
// fresh parse of the same bytes. Not "close". Identical.
func assertSpliceMatchesFresh(t *testing.T, incremental *Tree, src []byte) {
    t.Helper()
    fresh, err := NewParser().Parse(src)
    if err != nil {
        t.Fatalf("fresh parse failed: %v", err)
    }
    if got, want := incremental.Root().SExpr(), fresh.Root().SExpr(); got != want {
        t.Fatalf("incremental tree diverged from fresh parse\n got: %s\nwant: %s", got, want)
    }
}
```

Run it from a fuzzer: random Perl corpus, random edits, assert after every one.
Go's built-in `testing.F` gives you this with no dependency. The corpus should
include heredocs, POD, `s{}{}`, prototypes, and `__END__` — the Class A and B
triggers — because those are where a splice will be wrong.

Wire it into a property test as well: **for any document and any edit,
`Repair` either returns a tree equal to the fresh parse, or returns
`ok=false`.** There is no third outcome. Every bug in this subsystem is a
violation of that one sentence.

In this repository that test is `FuzzIncremental` and
`FuzzIncrementalSequence` in `internal/parse/incremental_fuzz_test.go`
(chapter 7 §7.7.2-§7.7.4), with the tree comparison done by `parse.TreeDiff`.

### 6.5.5 The flag: `parse.IncrementalRepair`

Damage/repair ships switched off and is switched on only once the fuzzers above
have earned it (§6.11 step 7).

`parse.IncrementalRepair` is an exported package-level `bool` in `internal/parse/incremental.go`, by default `false` until the flip.

- **What it switches.** `parse.ReparseIncremental` (both forms) reads it.
  When it is false, `ReparseIncremental` is a full parse of the new bytes, as
  §6.11 step 3 does on every change. When it is true, `ReparseIncremental`
  calls `Repair`; when `Repair` returns ok=false it falls back to a full
  parse of the new bytes, so its result always equals a fresh parse.
- **What it does not switch.** `Repair` itself ignores the flag. The fuzzers
  and the seed tests set the flag locally and restore it (`repairOn`), so
  `ReparseIncremental` goes through `Repair` there, and they call `Repair`
  directly too: repair is tested before it is trusted.
- **The flip condition.** `FuzzIncremental` and `FuzzIncrementalSequence` each
  run 10,000,000 execs (`make fuzz-incremental`) with zero counterexamples,
  recorded in plan §5 M2. The commit that flips the default to `true` cites
  those two runs and turns `TestIncrementalRepairFlagOn`, which asserts the
  default is `true`, green; it rewrites this section to say `true`.
- **No `context`.** `Repair` and `ReparseIncremental` take no `context`, unlike
  §6.5.3's sketch: cancellation is between stages only. The document's owner
  goroutine (§6.7.1, §6.7.3) checks for a superseding edit between parse,
  lowering and analysis, never inside a stage. The cost is accepted: a
  superseded full-parse fallback runs to completion before the owner sees the
  newer edit.

---

## 6.6 Position Tracking

The classic bug generator. Three coordinate systems, and the protocol speaks
the one your parser does not.

### 6.6.1 The three systems

| System | Unit | Who uses it |
|---|---|---|
| **Byte offset** | UTF-8 bytes | Your lexer, parser, tree, PSC (`infer.go` uses `uint32` byte offsets throughout) |
| **UTF-16 code units** | 16-bit units | **LSP `Position.character` by default** |
| **Runes** | Unicode scalars | Nothing in this system. Do not introduce it. |

In `my $café = 1;` the `=` is at byte 11 but UTF-16 column 10, because `é` is
two UTF-8 bytes and one UTF-16 unit. perl-lsp shipped exactly that bug and
fixed it in `edit.rs:29-33` (issue #750) — the LSP column was being used as a
raw byte count.

`💖` is worse: 4 UTF-8 bytes and **two** UTF-16 units, a surrogate pair, so a
client can send a column pointing *between* the surrogates. perl-lsp clamps to
the character start in one place (`convert.rs:81-84`, tested at
`convert.rs:140-148`) and returns `None` in another
(`perl-line-index/src/lib.rs:105-108`). **Clamp; do not error** — a malformed
column should never fail a request.

### 6.6.2 The Go implementation

Internally, byte offsets. Only convert at the JSON-RPC boundary.

**Implemented** in `internal/psc/document.go` with UTF-16 alone, step 1 of
§6.6.3, so the signatures that landed take no `PosEnc` and use the
handler's `lspPosition` and the `uint32` offsets of `infer`:

```go
func (d *Document) ByteToPosition(off uint32) lspPosition
func (d *Document) PositionToByte(p lspPosition) uint32
```

`ByteToPosition` also clamps an offset in a line ending to the end of the
line's content, so the `\n` of a CRLF line has the column of its `\r`.
Every position in `internal/psc/lsp_handler.go` goes through these two.
The sketch below is the shape once `PosEnc` is threaded (§6.6.3 step 2).

```go
// PosEnc is the negotiated LSP position encoding.
type PosEnc int

const (
    EncUTF16 PosEnc = iota // LSP default and mandatory baseline
    EncUTF8                // preferred when the client supports it
)

// ByteToPosition converts a byte offset to an LSP position.
func (d *Document) ByteToPosition(off int32, enc PosEnc) Position {
    if off < 0 { off = 0 }
    if int(off) > len(d.Text) { off = int32(len(d.Text)) }

    line := int32(sort.Search(len(d.Lines), func(i int) bool {
        return d.Lines[i] > off
    }) - 1)
    if line < 0 { line = 0 }
    lineStart := d.Lines[line]

    // Clamp to a UTF-8 boundary: a byte offset inside a multi-byte
    // character must not produce a half-character column.
    for off > lineStart && !utf8.RuneStart(d.Text[off]) {
        off--
    }
    seg := d.Text[lineStart:off]
    if enc == EncUTF8 {
        return Position{Line: line, Character: int32(len(seg))}
    }
    return Position{Line: line, Character: utf16Len(seg)}
}

// utf16Len counts UTF-16 code units in UTF-8 bytes without allocating.
// A rune above the BMP needs a surrogate pair, hence two units.
func utf16Len(b []byte) int32 {
    var n int32
    for i := 0; i < len(b); {
        r, size := utf8.DecodeRune(b[i:])
        i += size
        if r > 0xFFFF {
            n += 2
        } else {
            n++
        }
    }
    return n
}

// PositionToByte converts an LSP position to a byte offset. Out-of-range
// input is clamped, never rejected: editors send stale positions routinely.
func (d *Document) PositionToByte(p Position, enc PosEnc) int32 {
    if p.Line < 0 { return 0 }
    if int(p.Line) >= len(d.Lines) { return int32(len(d.Text)) }

    lineStart := d.Lines[p.Line]
    lineEnd := int32(len(d.Text))
    if int(p.Line)+1 < len(d.Lines) {
        lineEnd = d.Lines[p.Line+1] - 1 // exclude '\n'
    }
    // Exclude a trailing '\r' so CRLF columns match the client's view.
    if lineEnd > lineStart && d.Text[lineEnd-1] == '\r' {
        lineEnd--
    }
    if enc == EncUTF8 {
        off := lineStart + p.Character
        if off > lineEnd { return lineEnd }
        for off > lineStart && !utf8.RuneStart(d.Text[off]) { off-- }
        return off
    }
    var units int32
    for i := lineStart; i < lineEnd; {
        if units >= p.Character {
            return i
        }
        r, size := utf8.DecodeRune(d.Text[i:lineEnd])
        w := int32(1)
        if r > 0xFFFF {
            w = 2
        }
        // Column lands inside a surrogate pair: clamp to the char start.
        if units+w > p.Character {
            return i
        }
        units += w
        i += int32(size)
    }
    return lineEnd
}
```

Note `unicode/utf16` is not imported. You never need to *build* UTF-16 — only
to count units — and `r > 0xFFFF` does that in one comparison. `utf8.RuneStart`
does the boundary clamping. Both are stdlib, both are one line.

### 6.6.3 Capability negotiation, and the trap

LSP 3.17 added `positionEncoding`. The client advertises
`general.positionEncodings` (e.g. `["utf-8","utf-16"]`); the server picks one
and echoes it in `ServerCapabilities.positionEncoding`. `utf-16` is mandatory
and the default.

perl-lsp negotiates correctly (`runtime/lifecycle/capabilities.rs:366-389`) —
and then does this (`capabilities.rs:605-616`):

```rust
// Phase 1 (this PR) only negotiates and stores the client's preferred
// position encoding ... `text_sync` and every feature provider still compute
// positions in UTF-16 code units. Per the LSP 3.17 spec, client and server
// MUST agree on one encoding or offsets are misinterpreted, so the
// *advertised* `positionEncoding` MUST stay pinned to "utf-16" ... until
// phase 2 threads the negotiated encoding through the providers.
capabilities["positionEncoding"] = json!("utf-16");
```

That comment is the most useful thing in the crate. **Advertising an encoding
is a promise every provider must keep.** Negotiating UTF-8 while one hover
handler still counts UTF-16 silently corrupts positions for anyone with a
non-ASCII identifier — and it will not show up in an ASCII test suite.

The Go plan:

1. Ship advertising `utf-16` only. Store the negotiated value but do not use it.
   (Landed: `handleInitialize` advertises `utf-16`; nothing reads the
   client's offer yet, so nothing stores it.)
2. Thread `PosEnc` through **every** conversion site — one type, one parameter,
   no defaults.
3. Only then advertise `utf-8` when offered. The win is real: with UTF-8
   negotiated, `PositionToByte` is `lineStart + Character` with a boundary
   clamp, and the hot path stops decoding runes entirely.

Test with a corpus containing `é` (2 bytes/1 unit), `→` (3/1), and `💖` (4/2),
on both LF and CRLF. perl-lsp's `convert.rs:97-179` test block is a ready-made
table to port.

---

## 6.7 Concurrency in Go

### 6.7.1 The model: one owner goroutine per document

Two tempting models, both wrong. A **mutex over a shared document map** is
what perl-lsp does (`self.documents.lock()` throughout `text_sync.rs`), and its
own comment records the cost: eager incremental maintenance "needs its own
parse to run synchronously under the same `documents` lock as the text-state
update, so it is incompatible with the async worker path today"
(`text_sync.rs:48-56`) — a lock held across a parse serializes everything. A
**goroutine per request** races two parses of one document, and the loser may
publish after the winner.

Instead: **one goroutine owns each document; all mutation happens there;
readers get immutable snapshots.**

```go
// docWorker owns one document. Its fields are goroutine-confined: no other
// goroutine touches them, so no locks are needed on the parse path.
type docWorker struct {
    doc  *Document
    snap atomic.Pointer[Snapshot] // published for readers

    edits    chan Edit        // from didChange
    requests chan *docRequest // from every other LSP method
    quit     chan struct{}

    parseGen atomic.Uint64 // increments on every new edit
    cancel   context.CancelFunc
}

func (w *docWorker) run() {
    debounce := time.NewTimer(time.Hour)
    defer debounce.Stop()
    pending := false

    for {
        select {
        case e := <-w.edits:
            w.doc.ApplyEdit(int(e.Start), int(e.OldEnd), e.NewText)
            w.parseGen.Add(1)
            if w.cancel != nil {
                w.cancel() // supersede any in-flight parse
            }
            pending = true
            if !debounce.Stop() {
                select { case <-debounce.C: default: }
            }
            debounce.Reset(debounceInterval)

        case <-debounce.C:
            if pending {
                w.reparse()
                pending = false
            }

        case req := <-w.requests:
            if pending { // a request forces the parse to happen now
                if !debounce.Stop() {
                    select { case <-debounce.C: default: }
                }
                w.reparse()
                pending = false
            }
            req.serve(w.snap.Load())

        case <-w.quit:
            return
        }
    }
}
```

Why this shape:

- `didChange` returns immediately after a memmove. Text sync never blocks on a
  parse, which is what keeps typing smooth.
- `snap` is `atomic.Pointer[Snapshot]` — readers never lock, and a stale
  snapshot is *fine* for hover or documentSymbol. A tree one keystroke old
  beats a 200 ms wait.
- Requests jump the debounce queue. A `hover` that arrives 5 ms after a
  keystroke gets a current answer rather than waiting out the timer.

**As landed** (`internal/psc/owner.go`, issue 01a17969-4541): the owner is
`owner`, one per open URI in `LSPServer.documents`, whose `sync.Mutex` now
guards only that map — no parse runs under it, so two documents parse at once.
Every operation on an owner is a closure run on its goroutine (`owner.call`),
so `didChange` returns once the text is swapped and the timers reset. Requests
are served through the owner rather than from an `atomic.Pointer`: a request
forces a pending parse and waits for the tree of the latest version, which
lands before the waiters are served. A `didChange` whose version is not newer
than the owner's is dropped; one for a URI that is not open makes no owner;
`didClose` stops the owner and waits for its goroutine to return; `shutdown`
stops every owner, so nothing is published after it. A panic in parse, lower
or analyze is that version's error, published as one error diagnostic over the
document; the owner keeps the last good tree and keeps serving. An open is
parsed at once and its diagnostics published when it lands, with no debounce.

### 6.7.2 Debouncing

Two different debounces with different intervals — this is important and
perl-lsp separates them too (`runtime/diagnostic_debounce.rs`):

| Work | Interval | Rationale |
|---|---|---|
| Parse | **15 ms** | Below perceptual threshold; coalesces fast typing bursts |
| Diagnostics (PSC) | **300 ms** | Errors appearing mid-word are noise; wait for a pause |

perl-lsp's debouncer runs a dedicated thread with a `HashMap<uri, deadline>`
and `recv_timeout`, coalescing per URI. In Go, one `time.Timer` per document
worker is simpler and needs no map at all, because the worker *is* the per-URI
partition.

The timer-reset dance above is the correct idiom and it is easy to get wrong.
`Stop()` returning false means the timer already fired, so its value may be
sitting in the channel; the non-blocking drain prevents an immediate spurious
fire after `Reset`.

**As landed** (`internal/psc/owner.go`): `parseDebounce` is 15 ms and
`diagnosticsDebounce` 300 ms, each an owner timer (`parseT`, `diagT`). The
intervals and the clock the timers come from are `LSPServer` fields, so tests
shorten them or drive a fake clock; an interval of zero acts at once.
`stopTimer` is the Stop/drain idiom, and every `Reset` follows it.

### 6.7.3 Cancellation

Every parse gets a `context.Context` cancelled when a newer edit arrives:

```go
func (w *docWorker) reparse() {
    if w.cancel != nil {
        w.cancel()
    }
    ctx, cancel := context.WithCancel(context.Background())
    w.cancel = cancel

    gen := w.parseGen.Load()
    snap := w.snap.Load()

    // Text is copied here so the parse reads a frozen buffer.
    text := make([]byte, len(w.doc.Text))
    copy(text, w.doc.Text)

    go func() {
        newSnap, ok := parseWithRepair(ctx, snap, text, w.doc.Version)
        if !ok || ctx.Err() != nil {
            return // superseded or failed; the next parse will cover it
        }
        // Only publish if no newer edit landed while we worked.
        if w.parseGen.Load() == gen {
            w.snap.Store(newSnap)
            w.publishDiagnostics(newSnap)
        }
    }()
}
```

The generation check is the belt to the context's braces. `ctx.Err()` catches
most supersessions, but a parse that finished microseconds before the cancel
would otherwise publish a stale tree over a fresh one. Compare generations
before storing.

Cancellation must be *checked*, not merely signalled. The two places to poll:
the resync loop (§6.5.3) and the parser's statement loop. Once per statement is
plenty — `ctx.Err()` is an atomic load, roughly a nanosecond, and statements
are microseconds apart.

Also honour LSP `$/cancelRequest`. Map request IDs to `CancelFunc`s in the
server's dispatcher and cancel on receipt.

**As landed** (`internal/psc/owner.go`, `LSPServer.analyzeSource`, `analyze`):
each parse gets a `context.Context` that a newer edit cancels, and the
generation check is the only gate on landing — a result whose generation is not
the owner's is dropped, whether or not it saw the cancel. `ctx` is polled
before the parse, before lowering and before analysis; the parser's statement
loop does not poll it yet, so a cancelled parse runs to the end of the current
parse stage. `$/cancelRequest` is not handled yet.

### 6.7.4 Bounding total work

`GOMAXPROCS` goroutines parsing 20 open documents at once will thrash. Gate
parses through a semaphore:

```go
// parseSlots bounds concurrent parses. A buffered channel is Go's
// semaphore; there is no reason to import one.
var parseSlots = make(chan struct{}, max(2, runtime.NumCPU()/2))

func acquireParseSlot(ctx context.Context) bool {
    select {
    case parseSlots <- struct{}{}:
        return true
    case <-ctx.Done():
        return false
    }
}
func releaseParseSlot() { <-parseSlots }
```

Half the CPUs, minimum two. The editor's own UI needs the rest.

**As landed** (`internal/psc/owner.go`): `parseSlots` has capacity
`max(2, runtime.NumCPU()/2)`, and `owner.parse` acquires a slot with the
parse's `ctx` around the whole parse, lower and analyze.

---

## 6.8 Error Recovery for a Live Buffer

While typing, the buffer is almost always invalid. `my $x = ` is not a
sentence in Perl, but the user typing it still wants completion after `=`.

### 6.8.1 What a live tree must guarantee

1. **Total.** Every byte of source is covered by some node. No gaps.
2. **Local.** An error in one sub does not destroy the tree for the others.
3. **Stable.** Node identity for untouched regions survives across parses, or
   the editor's folding ranges and semantic-token deltas flicker.
4. **Quiet.** One syntax error produces one diagnostic, not two hundred.

### 6.8.2 Error and missing nodes

Two distinct kinds, and the distinction matters to PSC:

```go
// KindError covers source text the parser could not fit into the grammar.
// It has children (the tokens it swallowed) and a real byte span.
// KindMissing is a zero-width node for a token the grammar required and the
// source did not supply — e.g. the `;` in `my $x = 1`. It has no children
// and start == end.
const (
    KindError   KindID = /* ... */
    KindMissing KindID = /* ... */
)

func (n *Node) IsError() bool   { return n != nil && n.kind == KindError }
func (n *Node) IsMissing() bool { return n != nil && n.kind == KindMissing }
func (n *Node) HasError() bool  { return n != nil && n.flags&FlagSubtreeError != 0 }
```

`HasError` must be a **precomputed flag propagated up at construction**, not a
tree walk. PSC calls it, and `gotreesitter`'s reuse cursor
(`incremental.go:168`) tests `cur.hasError` on every candidate — an O(subtree)
walk there would make reuse slower than reparsing.

**Prefer missing nodes over error nodes.** `my $x = 1` with a synthesized
missing `;` gives PSC a well-formed `variable_declaration` it can type; the
same text wrapped in an error node gives PSC nothing. Missing-token insertion
is where most live-buffer usability comes from.

The tokens worth synthesizing in Perl:

| Missing | Insert when | Result |
|---|---|---|
| `;` | Next token starts a new statement, or EOF | Statement completes |
| `}` | EOF with open blocks | Blocks close; outer structure survives |
| `)` | A `;` or `}` arrives with open parens | Expression completes |
| Expression | Binary operator with no RHS (`$x = `) | Placeholder so `=` typechecks |
| Identifier | `sub` with no name, `$` with no name | Declaration is still recognizable |

**As landed (`internal/parse/recovery.go`).** The kinds are `parse.Error` and
`parse.Missing`, and the flag is the field `Node.HasError`, set by
`markErrors` on each statement as it is finished (only on one that built a
recovery node) and by `refuseUnclosedOpener` along the path it inserts into.
An Error node stands where a catch-all Unknown stood -- `not_a_term`,
`trailing_tokens`, `not_an_expression` -- carries that code, and covers the
same bytes; `parse.CatchAll` names those three, and their nodes are structure
only, reporting nothing (perigrin, R-a′). The missing operand is a Missing
node carrying `missing_operand` (R-g), at the operator's end, before `;`, `}`
or end of input. The other Missing nodes carry no code and Text names the
token: `;` (only before `my`/`our` on a later line, not after a word), each
closer still open at end of input (a child of the `unclosed_opener` or
specific refusal holding the last token; none past a `nesting_too_deep`
decline), and an empty Text for an identifier after `sub` or `$`, inside the
`not_a_term` Error. Specific refusals stay `Unknown`. Every reader asks
`(*Node).Refused()` (`internal/parse/refused.go`) rather than comparing a
Kind with `Unknown`, so ratchets count an Error node as they counted its
Unknown.

### 6.8.3 Panic-mode recovery, bounded

When no missing token works, skip to a synchronization point. The sync-point
tiers (column-0 `sub`/`package` as a hard reset, line-start keywords,
punctuation at the current depth), the recovery loop, the nearest-point rule
and the 100-token cap are chapter 5 §5.13.4; this chapter adds nothing to
them. What the live buffer needs from that loop is only that everything
skipped becomes **one** error node (§6.8.2), so that §6.8.4's filter sees one
diagnostic.

**As landed (`internal/parse/recovery.go`, `panicSkip`).** Panic mode starts
at a catch-all refusal at statement level: `trailing_tokens`,
`not_an_expression`, and a statement that is one `not_a_term`.
`missing_operand` starts none, since its Missing node fits. From there the
skip stops at the nearest of §5.13.4's points: after a `;`, or after a `{
... }` the skip read (a block form takes no `;`); before the `}` closing the
enclosing block, a column-0 `sub`, `package`, `class`, `format` or phaser
name (tier 1), a line-start tier-2 keyword where a statement can start, or
the 101st token (the cap, `maxSkip`). A sync point counts only once the
statement holds a token, so every recovery advances. A heredoc body, POD or
a string is one token, so a `sub` inside one is never a sync point.

Three choices §5.13.4 leaves open are fixed here.

- **Where a statement can start.** A tier-2 word syncs after a `;`, `{` or
  `}`. After a term (a variable, number, string or closing bracket), only a
  declaration word syncs (`my our local state sub package use`), since the
  statement before it lost its `;`. A modifier word (`if unless elsif else
  while until for foreach`) there continues the statement, as in `croak
  "x"\n\t\tunless $loaded;`. After a word or an operator, nothing syncs:
  `$cb =\n sub { ... }` is one statement.
- **Brackets.** The skip counts brackets from the statement's first token,
  so a closer the parse stopped short of, as in `++$h{refaddr $x} > 1`, is
  the statement's and not the block's. Inside a bracket only tier 1 and the
  enclosing block's `}` stop it. The cap waits for the bracket to close:
  catch-alls fire on Perl perl compiles (`trailing_tokens` on 318 of 620
  perl.git `t/` files), and a cap that cut a group split such statements
  into fragments. A closer closes the innermost opener of its kind. A `)`
  or `]` closing an opener read before the skip ends the skip after it.
  That opener may be in the statement, or before it and still open
  (`enclosedBy`, back to the nearest `;`, `{` or `}`). So an error in `if
  (foo 'x') { ... }` stays inside the condition, and the block after it is
  parsed. A closer with none open and no block to close is stray and
  skipped, so a cascade of one-token refusals is one Error node.
- **Only the tokens already read decide.** Nothing in a prefix's parse
  depends on text after it, which is what resuming a snapshot on edited
  source (Repair) needs. A statement still open at end of input reads on
  into whatever is appended, so no snapshot is taken after one
  (`openToEOF`).

The skipped region is one Error node carrying the code that started it, and
a specific refusal in what parsed before the skip stays its child, with its
code. The heredoc bodies after the skip are taken into it.

Tier 1's hard reset for a lost brace is decided in `parseBlock`, where the
block is seen to end. A block that reaches end of input with its `{` open
closes at the first column-0 anchor its own statement loop met directly
inside it (`tierOneMark`). The parser is put back as it stood there, its
snapshots since the brace are dropped, and the block ends with a zero-width
Missing `}` before the anchor. `refuseUnclosedOpener` leaves that brace out
of its count (`closedEarly`). A sub missing its `}` leaves every column-0 sub
after it at top level, each equal to its error-free parse.

A column-0 `sub` the expression parser takes, as an anonymous sub or inside
a paren, is not a statement of the block, so it closes nothing: the brace
gains its Missing `}` at end of input as before. What follows a tier-1 close
depends on there being no `}` later, so no snapshot is taken after one
(`rescued`). A balanced file is never touched.

### 6.8.4 The anti-cascade rule

**One syntax error produces at most one diagnostic. Non-negotiable.**

A missing `}` can make every subsequent sub look nested and produce a
diagnostic per line. That is the single most common way a language server
becomes unusable.

Three mechanisms, applied together:

```go
// Diagnostics from parse errors are filtered before publication.
type diagFilter struct {
    lastErrorEnd int32
    count        int
}

const (
    maxParseDiagnostics = 20  // hard cap per document
    errorCooldownBytes  = 40  // suppress errors within N bytes of the last
)

func (f *diagFilter) admit(d Diagnostic) bool {
    if f.count >= maxParseDiagnostics {
        return false
    }
    // Cascade suppression: an error immediately after another is almost
    // always a consequence of it, not an independent problem.
    if d.StartByte < uint32(f.lastErrorEnd+errorCooldownBytes) && f.count > 0 {
        return false
    }
    f.lastErrorEnd = int32(d.EndByte)
    f.count++
    return true
}
```

1. **Cooldown.** An error within 40 bytes of the previous one is suppressed.
2. **Hard cap.** After 20 diagnostics, stop. Report "20 errors; further errors
   suppressed" if you must say anything.
3. **PSC gating.** This is the important one: **when the tree has parse errors,
   do not publish PSC type diagnostics at all** — or publish only from subtrees
   with `HasError() == false`. A type error derived from a tree the parser
   already knows is wrong is noise with a false-confidence label.

PSC's `Diagnostic` struct (`internal/infer/diagnostics.go:39-47`) has
`StartByte`/`EndByte`, so subtree gating is a range test against the error
nodes' spans. Cheap.

**As landed (`internal/psc/diagfilter.go`).** `diagFilter` is the filter
above, with the 40-byte cooldown (`errorCooldownBytes`) measured from the end
of the last *admitted* error and the cap of 20 (`maxParseDiagnostics`). One
is made per analysis, so neither carries over between publications or
documents; psc check and the LSP both report through it (`analyze`). Its
input, `parseErrors`, is the reported errors only: each Missing node for a
token, each Error or Missing node whose code `parse.CatchAll` does not
classify catch-all (`missing_operand` among them, R-g), and each `Unknown`
whose code perl always shares (`perlRefuses`). Nothing inside a reported or
catch-all node is reported again, so a Missing `}` inside an
`unclosed_brace` refusal is one diagnostic, and a catch-all node publishes
none (R-a′). Gating is the range test, against the innermost statement
holding each reported error, so a type error in `$u + 1 + ;` is gated though
its Missing operand is zero-width after it. Catch-all nodes gate nothing.
`HasError` is not the gate, since it also covers catch-all nodes.

### 6.8.5 Recovery interacts with incremental parsing

Two rules that are easy to miss:

- **Never reuse a subtree containing an error.** `gotreesitter` enforces this
  (`incremental.go:168`: `if cur.hasError || ... { continue }`). An error node's
  extent is an artifact of where recovery happened to stop; it is not a
  meaningful structural fact and it will not be correct at a new position.
- **A parse that changes error status invalidates the damage estimate.** The
  check in §6.5.3 (`sub.HasError() != enclosing.HasError()`) exists because
  gaining or losing an error means recovery ran differently, and recovery's
  effects are not local. Fall back to a full parse.

---

## 6.9 The LSP Surface

### 6.9.1 Requests and what each needs from the parser

| Request | Parser capability | Snapshot freshness | Budget |
|---|---|---|---|
| `initialize` | none | — | < 50 ms |
| `textDocument/didOpen` | full parse | fresh | < 100 ms |
| `textDocument/didChange` | text apply only | — | **< 1 ms** |
| `textDocument/semanticTokens/full` | full token stream + positions | fresh | < 50 ms |
| `textDocument/semanticTokens/full/delta` | token stream + previous result id | fresh | < 20 ms |
| `textDocument/documentSymbol` | tree; sub/package nodes | stale OK | < 20 ms |
| `textDocument/definition` | tree + symbol table | stale OK | < 30 ms |
| `textDocument/hover` | tree + PSC annotations | stale OK | < 30 ms |
| `textDocument/completion` | tree at cursor, error-tolerant | **fresh** | < 50 ms |
| `textDocument/publishDiagnostics` | tree + PSC `Analyze` | fresh, debounced | < 300 ms |
| `textDocument/foldingRange` | tree; block extents | stale OK | < 20 ms |

Three observations that shape the design:

- **`didChange` must not parse.** It is a notification; nothing waits on a
  reply, but the client's next request queues behind it. Apply text, bump
  version, return. §6.7's worker does exactly this.
- **`semanticTokens` is the token cache's customer.** It needs *every* token
  with a position and a classification — including comments and POD, which the
  tree may drop as trivia. Keep them in `TokenCache.Tokens` with a trivia flag
  rather than discarding them at lex time. The delta variant needs the previous
  result kept per document, keyed by a result id.
- **`completion` is the only request that genuinely needs a fresh parse**, and
  the buffer is guaranteed invalid at the cursor. It is the reason §6.8 exists.

**Text sync follows the flag.** The server advertises `TextDocumentSync` 2
(incremental) exactly when `parse.IncrementalRepair` is true, and 1 (full)
otherwise. Under 2, each content change's range is in UTF-16 line and character
positions, converted to byte offsets through the document's line table —
clamped as §6.6 clamps any position — and the changes of one `didChange` apply
in order, each to the text the one before it left. Under 1, the first change's
text replaces the document, as before. With the flag set, the owner re-parses
through `ReparseIncremental` from the last tree that landed, its `parse.Edit`
the common prefix and suffix of that tree's bytes and the current text, so a
coalesced burst or a cancelled parse never pairs a tree with other bytes; a
`file:` URI goes through `Session.ReparseIncremental`, each re-parse in a fresh
Session sharing the document's `parse.ModuleCache`.

### 6.9.2 JSON-RPC framing by hand

stdlib-only means writing the transport: about 80 lines. HTTP-like headers,
then a JSON body:

```
Content-Length: 123\r\n
\r\n
{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
```

```go
// Conn is an LSP JSON-RPC 2.0 connection over a stream (stdin/stdout).
type Conn struct {
    r  *bufio.Reader
    w  *bufio.Writer
    mu sync.Mutex // serializes writes; reads are single-threaded by design
}

func NewConn(r io.Reader, w io.Writer) *Conn {
    return &Conn{r: bufio.NewReaderSize(r, 64*1024), w: bufio.NewWriter(w)}
}

// Message is both request and response; absent fields stay nil.
type Message struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      json.RawMessage `json:"id,omitempty"`     // number OR string
    Method  string          `json:"method,omitempty"`
    Params  json.RawMessage `json:"params,omitempty"` // decoded per method
    Result  json.RawMessage `json:"result,omitempty"`
    Error   *RespError      `json:"error,omitempty"`
}

type RespError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

// Read reads one framed message.
func (c *Conn) Read() (*Message, error) {
    var length int
    for {
        line, err := c.r.ReadString('\n')
        if err != nil {
            return nil, err
        }
        line = strings.TrimRight(line, "\r\n")
        if line == "" {
            break // end of headers
        }
        name, value, ok := strings.Cut(line, ":")
        if !ok {
            continue // tolerate junk headers
        }
        if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
            length, err = strconv.Atoi(strings.TrimSpace(value))
            if err != nil {
                return nil, fmt.Errorf("bad Content-Length: %w", err)
            }
        }
    }
    if length <= 0 || length > maxMessageBytes {
        return nil, fmt.Errorf("invalid Content-Length: %d", length)
    }
    buf := make([]byte, length)
    if _, err := io.ReadFull(c.r, buf); err != nil {
        return nil, err
    }
    var m Message
    if err := json.Unmarshal(buf, &m); err != nil {
        return nil, err
    }
    return &m, nil
}

// Write writes one framed message. Safe for concurrent use.
func (c *Conn) Write(m *Message) error {
    m.JSONRPC = "2.0"
    body, err := json.Marshal(m)
    if err != nil {
        return err
    }
    c.mu.Lock()
    defer c.mu.Unlock()
    if _, err := fmt.Fprintf(c.w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
        return err
    }
    if _, err := c.w.Write(body); err != nil {
        return err
    }
    return c.w.Flush()
}
```

Details that bite:

- **`ID` is `json.RawMessage`.** The spec permits number or string. Decoding to
  `int` breaks clients that send strings; decoding to `interface{}` turns 64-bit
  ids into lossy `float64`. Keep it raw and echo it back verbatim.
- **`Params` is `json.RawMessage`.** Decode per method into a typed struct. Do
  not decode into `map[string]interface{}` and then fish with string keys —
  perl-lsp does this (`params.pointer("/textDocument/uri")`) and it is how
  `.character` ends up as a `float64`.
- **`Content-Length` is bytes, not runes.** `len(body)` on a `[]byte`. Getting
  this wrong desynchronizes the stream permanently.
- **Write must be mutex-guarded.** Diagnostics are published from parse
  goroutines while the main loop writes responses. Two interleaved writes
  corrupt the framing.
- **`bufio.Reader` sized at 64 KB.** `didOpen` on a large file arrives as one
  message; the default 4 KB buffer makes it many syscalls.
- **`maxMessageBytes`** (say 64 MB) so a malformed header cannot make you
  allocate the heap.

`encoding/json` is fast enough. A 150 KB `didOpen` unmarshals in a few hundred
microseconds. Do not reach for a faster JSON library until a profile says so.

### 6.9.3 The dispatch loop

```go
func (s *Server) Serve(ctx context.Context) error {
    for {
        msg, err := s.conn.Read()
        if err == io.EOF {
            return nil
        }
        if err != nil {
            return err
        }
        switch msg.Method {
        // Notifications with ordering requirements run inline: they must be
        // applied in the order received, and they are all O(memmove).
        case "textDocument/didOpen", "textDocument/didChange",
            "textDocument/didClose", "textDocument/didSave":
            s.handleSync(msg)

        case "$/cancelRequest":
            s.cancelRequest(msg.Params)

        case "shutdown":
            s.respond(msg.ID, nil)

        case "exit":
            return nil

        default:
            // Everything else is a request: handle off the read loop so a
            // slow provider cannot stall document sync.
            go s.handleRequest(ctx, msg)
        }
    }
}
```

The split matters. Sync notifications must be ordered and are cheap, so they run
inline. Requests may be slow, so they run concurrently — and because they only
read published snapshots (§6.7.1), no locking is required.

---

## 6.10 Memory and Latency Budget

### 6.10.1 The target, and the real corpus

**A keystroke in a 5,000-line Perl file yields a PSC-consumable tree in under
10 ms at p99.**

`TestLatencyEditBudget` (`internal/psc`) holds that gate on
`internal/psc/testdata/latency/edit5k.pl`, a generated 5,000-line fixture: it
drives at least 200 1-char edits through the real `didChange` path and fails
unless the nearest-rank p99 of the re-parse, from the document's owner starting
it (§6.7.1) to the PSC-consumable tree, is under 10 ms. The target is on the
tree, not on analysis. The edit-to-analysis interval is a separate test,
`TestLatencyEditAnalysisBudget`, whose budget and interval are set by
perigrin's ruling on the step-4 numbers (issue 01a17969-45ba). Neither counts the
§6.7.2 debounce: each timer starts when the owner starts the re-parse (or the
analysis), not when the keystroke arrives (perigrin, 2026-10-27).

The corpus is not hypothetical. In `perl5/lib` and on CPAN you will find:

| File | Approx. lines | Why it hurts |
|---|---|---|
| `overload.pm` | ~1,500 | Dense operator tables |
| `CPAN.pm` | ~10,000 | Long subs, heavy heredoc use |
| `Encode` tables | ~50,000 | Generated, enormous data structures |
| `DBI.pm` | ~9,000 | Prototypes, XS boundary, deep nesting |
| `Moose` internals | ~2,000/file | Metaprogramming; barewords everywhere |
| `bigint.pm` / `Math::BigInt` | ~7,000 | Overload-heavy |

5,000 lines is the *median* real file, not the worst case. Design for the
median and degrade explicitly beyond it.

### 6.10.2 The budget, itemized

| Stage | Budget | Notes |
|---|---|---|
| Apply edit to `[]byte` | 10 µs | One memmove of ~150 KB |
| Reindex line table suffix | 20 µs | Only lines after the edit |
| Triage / trigger scan | 5 µs | `bytes.Contains` over one line window |
| Re-lex from checkpoint | 200 µs | ≤ 512 bytes typical, plus resync overshoot |
| Re-parse anchor region | 500 µs | One statement or sub |
| Splice tree + copy arena | 100 µs | Bounded by damage region |
| **Parse subtotal** | **~0.9 ms** | |
| PSC `Analyze` (full tree) | 0.5–2 ms | Measured, see below |
| Diagnostic formatting | 100 µs | |
| **Total** | **~3 ms** | |

#### Measured on the tree-sitter path, 2026-09-04: the reverse of the intuition

This measurement is history. The budget above is the *target* for a
hand-written Go parser. Against the parser PSC consumed then — tree-sitter,
through `internal/parser`, since retired — the ratio was inverted: parse
dominated analysis by 20-50×. PSC now reads `internal/parse/` through
`internal/lower`, and neither has been benchmarked on these files
(01a14248-51da-7d76-bc68-4f7dbeb842b1):

| File | Lines | Parse | `Analyze` | Ratio |
|---|---:|---:|---:|---:|
| `charnames.pm` | 484 | 6.5 ms | 0.48 ms | 14× |
| `FileHandle.pm` | 262 | 69.7 ms | 2.8 ms | 25× |
| `overload.pm` | 1701 | 99.6 ms | 2.2 ms | 45× |
| `sigtrap.pm` | 327 | 350.0 ms | 7.0 ms | 50× |
| `_charnames.pm` | 858 | 521.7 ms | 19.3 ms | 27× |

`go test -bench`, perl5/lib, 2026-09-04. Note that time tracks grammar
pathology rather than file length — `sigtrap.pm` at 327 lines costs 5× what
`overload.pm` costs at 1701.

So PSC was **not** the bottleneck: at 0.5-2 ms it already fit the budget, while
a single tree-sitter parse blew it by one to two orders of magnitude. The consequence for
build order is direct — replacing the parser is what buys the interactive
budget, and optimising `Analyze` first would be optimising the cheap half.

`Analyze`'s incremental story was a real problem then, just not a latency
one: its annotation map was `map[uint32]types.Type` keyed by `StartByte`, so
every key after an edit shifted and nothing survived a re-parse. The map is
now keyed by `lower.ID` (§6.4.3), and an edit renumbers only the sub it
lands in.

### 6.10.3 What the budget forces

1. **No allocation on the edit path.** Reuse buffers. `ApplyEdit`'s `cap`
   check exists for this. Set `GOGC=400` for the server process — heap is
   plentiful and GC pauses come straight out of budget.
2. **Interned kind strings.** PSC calls `Kind()` at 95 sites and once per node
   visited, so it runs constantly. Returning a `string` from a `[]string`
   indexed by a `KindID` is a pointer load; building a string per call would
   dominate the walk.
3. **Arena-allocated nodes.** 40,000 individual allocations per parse is
   ~4 ms of GC pressure alone. One slice, one allocation.
4. **No rope.** §6.2.1.
5. **PSC's annotation map is keyed by node identity**, `lower.ID`, so the
   annotations of every sub an edit does not touch survive a splice. The
   byte-keyed `map[uint32]types.Type` of the tree-sitter path did not: every
   key after an edit shifted. See §6.4.3 and §6.10.5.
6. **Degrade explicitly above a threshold.** perl-lsp has
   `max_file_size_bytes()` and skips parsing beyond it
   (`text_sync.rs:126-140`). Do the same, but degrade rather than refuse:

```go
const (
    fullServiceBytes = 512 * 1024 // below: everything
    degradedBytes    = 4 << 20    // below: syntax only, no PSC
    // above degradedBytes: tokens only — semantic tokens and folding work,
    // nothing else. Never refuse to open a file.
)
```

Tell the user which tier they are in. Silent degradation reads as a broken
server.

### 6.10.4 Memory

Per open 5,000-line (150 KB) document:

| Structure | Size |
|---|---|
| `Text` | 150 KB (+25% headroom) |
| `Lines` | 20 KB |
| `Tokens` (~30k × 16 B) | 480 KB |
| `States` (~30k × ~80 B) | 2.4 MB |
| Checkpoints (~300) | 30 KB |
| Node arena (~40k × 48 B) | 1.9 MB |
| PSC annotations | ~500 KB |
| **Total** | **~5.5 MB** |

20 open documents is ~110 MB. Acceptable for an editor plugin, and `States`
is the obvious first cut if it is not — store a state only at checkpoints
rather than per token, and recompute by re-lexing forward from the checkpoint
when the resync loop needs one. That trades ~2.4 MB per document for a few
microseconds per resync comparison. Make that trade if memory complaints
arrive; not before.

### 6.10.5 The biggest risk

**The parser. Measured, not assumed.**

The obvious suspect is PSC: `Analyze` walks every node, every time, and its
cost grows with file size independent of edit size. Measured, though, it is 0.5-2 ms on real perl5/lib files — inside
budget, and 20-50× *cheaper* than a single parse with the tree-sitter parser
PSC read on 2026-09-04 (§6.10.2, history). The intuition that the type checker must be the expensive half is
simply wrong here.

**The measured risk is the parser.** A single parse of `sigtrap.pm` — 327
lines — costs 350 ms, which is not 3× over an interactive budget but 35×. Cost
tracks grammar pathology rather than file size, so it cannot be predicted from
line count and cannot be debounced away: the user is typing in the file that is
slow. This is the strongest available argument for the whole specification,
and it is a stopwatch reading rather than a design preference.

Mitigations, revised to match the measurement:

1. **Replace the parser.** This is the whole budget. Everything else is
   rounding error until it is done.
2. **Debounce diagnostics at 300 ms** while parsing at 15 ms. Cheap, and worth
   doing regardless — it decouples PSC from the keystroke path even though PSC
   is not currently the problem.
3. **Key annotations by node identity, not byte offset.** Done when PSC
   moved to the lowered tree: `Analyze`'s map is `map[lower.ID]types.Type`
   (§6.4.3). The byte-keyed map before it lost every annotation after an
   edit.
4. **Analyze only subs intersecting the damage region.** Sound when nothing in
   that region changes a declaration, which the Class B trigger scan already
   tells you. Implemented in `internal/psc/incremental_analyze.go`: the
   damaged subs are re-lowered and re-analyzed, every other sub keeps its
   lowering and analysis, and a Class B edit falls back to the full path
   (§6.4.3). The step-4 numbers asked for it: lowering and `Analyze` were
   35 ms and 21 ms p50 on edit5k.pl, against a 10 ms gate.

Ship (1) first — the numbers leave no other reading. (2) is free. (3) and (4)
are done; (4) waited for step 4's evidence of §6.11, and got it.

---

## 6.11 Implementation Order

Build in this order. Each step is independently testable and shippable. This
is the spec's only build order: the per-chapter checklists (§2.15, §3.12,
§4.16, §5.16) are acceptance lists, not sequences, and the milestone ladder
(chapter 7 §7.8, held in `docs/plans/2026-09-05-parser-conformance-plan.md`
§5) is the set of gates the steps must pass. Step 0, before any of it, is the
oracle self-check test (plan §6).

1. **`Document` + line table + position conversion** (§6.2, §6.6). Test with a
   non-ASCII corpus first — this is where the silent bugs live.
2. **JSON-RPC transport** (§6.9.2). Verify against a real editor with only
   `initialize` and `didOpen` implemented.
3. **Full-parse-on-every-change server.** No incrementality. This is what
   perl-lsp actually ships, and it is *usable* under 500 KB. Ship it. This
   step presupposes PSC consumes the new tree — the lowering pass and the
   rewrite of `internal/infer/`'s walkers against §4.14 (§6.1.6) — which is
   a prerequisite of this list rather than a step in it; the M1 gate (issue
   01a0a70f) sits before it.
4. **Measure.** Instrument every stage (perl-lsp's
   `FIRST_USE_LATENCY_METHODS` list at `runtime/latency.rs:6-16` names the
   right hot paths). Do not optimize before this step produces numbers.
5. **Error recovery** (§6.8). Biggest quality win per line of code —
   completion inside a half-typed statement is what users notice.
6. **Token cache + checkpoints** (§6.3). Now incrementality has a foundation.
7. **Damage/repair** (§6.5), gated behind the flag `parse.IncrementalRepair`
   (§6.5.5; `internal/parse/incremental.go`, default `false`), with the
   splice-equals-fresh-parse fuzzer (§6.5.4) green before the flag flips. While
   it is false the LSP re-parses fully on every change; when true it calls
   `ReparseIncremental`. The flag flips when `FuzzIncremental` and
   `FuzzIncrementalSequence` have each run 10,000,000 execs
   (`make fuzz-incremental`) with zero counterexamples, recorded in plan §5 M2;
   `TestIncrementalRepairFlagOn` asserts the default is `true` and goes green
   in the commit that flips it.
8. **PSC incrementality** (§6.10.5), when the numbers from step 4 demand it.

The ordering is deliberate. perl-lsp built step 7 before step 3 was fast enough
to matter, and the result was 7,000 lines that production does not call. Do not
repeat that.

---

## 6.12 Summary of Recommendations

| Decision | Recommendation | Given up |
|---|---|---|
| Document storage | `[]byte` + `[]int32` line table | O(log n) edits; irrelevant at 150 KB |
| Line reindex | Suffix-only rebuild | Nothing |
| Token cache | Tokens + per-token lexer state + safe checkpoints | ~2.4 MB/doc |
| Reuse unit | Sub-anchored splice (§6.5.2) | Fine-grained node reuse; block- and statement-level anchoring until measured |
| Trigger detection | Over-eager substring scan on old and new text | Some needless full reparses |
| IR layers | Neither HIR nor PIR as perl-lsp built them; one lowering pass CST → §4.14 vocabulary (§6.1.6) | Prebuilt dataflow analysis |
| What inference walks | The lowered tree, never the CST; CST ↔ AST side table in both directions, erased nodes mapped forward (§6.1.6) | A single tree; the "PSC is unchanged" claim of the earlier draft |
| Fact keys | Stable node identity, for every side table | Byte-offset keys (rust-analyzer rejected them by name, §6.1.4) |
| Flow analysis | Over the lowered tree; single-pass Sorbet-style first (§6.1.7) | Deciding graph-vs-fixpoint-visit before measuring |
| Node API | The 10 methods PSC calls, on the CST, for the LSP's structural features | The tree-sitter-perl vocabulary — 46 grammar names PSC switches on are rewritten against §4.14 (§6.4.1) |
| Node allocation | Arena per tree, copy on splice | Zero-copy subtree sharing |
| Positions | Bytes internally; convert at the wire | Nothing |
| `positionEncoding` | Advertise `utf-16` until every provider is threaded | UTF-8 fast path, temporarily |
| Concurrency | One owner goroutine per document; atomic snapshots | Multi-goroutine parse of one doc |
| Cancellation | `context.Context` + generation counter | Nothing |
| Debounce | 15 ms parse / 300 ms diagnostics | Nothing |
| Error recovery | Missing tokens first, bounded panic mode second | Some recovery precision |
| Diagnostics | 40-byte cooldown, cap 20, gate PSC on parse errors | Some real errors hidden behind earlier ones |
| Correctness gate | Splice must equal fresh parse, fuzz-verified | Nothing |

The single most important sentence in this chapter: **an incremental parse that
does not produce exactly the tree a fresh parse would produce is not an
optimization, it is a bug.** Everything else is negotiable.
