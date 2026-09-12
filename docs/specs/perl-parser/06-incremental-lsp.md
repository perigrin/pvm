<!-- ABOUTME: Chapter 6 of the Perl parser spec — incremental re-parsing and LSP server architecture in pure Go. -->
<!-- ABOUTME: Damage/repair from an edit range, UTF-8/UTF-16 position tracking, live-buffer error recovery, and the concurrency model. -->

# Chapter 6: Incremental Parsing and LSP Architecture

This chapter is about architecture, not grammar. Chapters 1–5 tell you how to
parse Perl once. This one tells you how to parse it again, ten milliseconds
later, after the user typed one character — and how to wrap that in a language
server that stdlib-only Go can actually ship.

The target consumer is PSC, the type checker already living in
`internal/infer/`. Everything here is shaped by one constraint: **PSC must not
notice that the tree came from an incremental parse.** A tree assembled by
splicing is either indistinguishable from a fresh parse of the same bytes, or
it is a bug. Chapter 6 exists to make that guarantee cheap.

**Target: a keystroke in a 5,000-line Perl file produces a new, PSC-consumable
tree in under 10 ms at p95.** Every recommendation below is downstream of that
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

**Recommendation for the Go port: build neither, at first.**

PSC does not consume an IR. It walks CST nodes directly. Read
`internal/infer/infer.go:22-37` — it iterates `node.ChildCount()`, tests
`child.Kind() == "package"`, calls `child.Text(source)`. There are 4,265 lines
of that. Interposing an IR means rewriting all of it before a single keystroke
gets faster, and PIR's own documentation says it "does not promote provider
behavior" — that is, it did not make anything better yet.

Build the IR when you have a concrete need the CST cannot serve, and build only
one. The likely first need is Class B invalidation (§6.1.2): to know that
editing `sub f(\@)` invalidates every call site of `f`, you need a
declaration-to-use index. That is HIR's scope graph, and it is a *side table*
keyed by byte offset — not a replacement tree.

Give up: pre-built dataflow analysis for call hierarchy. Accept it. Add a
`facts` side table when a feature demands it, keyed by byte offset so it
invalidates with the same damage region as the tree.

---

## 6.2 The Document Model

### 6.2.1 Recommendation: `[]byte` plus a line table. Not a rope.

perl-lsp uses `ropey::Rope` (`state.rs:9`, `reparse.rs:8`). Go's standard
library has no rope, and you would be writing one.

Do not. Use this:

```go
// Document is the authoritative text of one open buffer.
type Document struct {
    URI     string
    Version int32   // LSP document version; monotonically increasing
    Text    []byte  // full UTF-8 source, always valid UTF-8
    Lines   []int32 // byte offset of each line start; Lines[0] == 0
}
```

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

**This is the good news of the whole chapter: the migration surface is ten
methods.** `internal/parser/parser.go` is already a thin façade over
`gotreesitter` (its ABOUTME says exactly that). Reimplement the façade over a
hand-written parser and PSC does not change — not one of its 4,265 lines.

Design the node accordingly:

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

For **PSC** specifically: its annotation map is keyed by `StartByte`
(`infer.go:57-60`, `map[uint32]types.Type`). Byte-keyed annotations are
invalidated wholesale by any edit, because every key after the edit shifts.

**Do not try to incrementally update PSC's annotation map.** Re-run `Analyze`
on the whole tree after each parse. Justification: `Analyze` is a bottom-up
walk over an in-memory tree with no I/O, and for 40,000 nodes that is a
few milliseconds. It fits the budget. If it stops fitting, the fix is to make
`Analyze` incremental per-sub — a much later project — not to patch a byte-keyed
map. Measured, `Analyze` is 0.5-2 ms against a parse of 6-500 ms on the same
files, so it is nowhere near the tightest constraint; see §6.10.2 and
00-findings §0.8.

The exception is `ProjectIndex` (`internal/infer/project.go`): cross-file
analysis triggered from `walkUseStatement` (`infer.go:22-53`). Those results
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
   rather than finishing work nobody wants (§6.7).
5. **`sub.HasError() != enclosing.HasError()`** — if re-parsing the region
   changed whether it contains an error, the damage estimate was wrong.
   Bail out.

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

### 6.8.3 Panic-mode recovery, bounded

When no missing token works, skip to a synchronization point. The sync-point
tiers (column-0 `sub`/`package` as a hard reset, line-start keywords,
punctuation at the current depth), the recovery loop, the nearest-point rule
and the 100-token cap are chapter 5 §5.13.4; this chapter adds nothing to
them. What the live buffer needs from that loop is only that everything
skipped becomes **one** error node (§6.8.2), so that §6.8.4's filter sees one
diagnostic.

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
10 ms at p95.**

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

#### Measured today, and it is the reverse of the intuition

The budget above is the *target* for a hand-written Go parser. Against the
parser we actually ship, the ratio is inverted — parse dominates analysis by
20-50×:

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

So PSC is **not** the bottleneck: at 0.5-2 ms it already fits the budget, while
a single parse blows it by one to two orders of magnitude. The consequence for
build order is direct — replacing the parser is what buys the interactive
budget, and optimising `Analyze` first would be optimising the cheap half.

`Analyze`'s incremental story is still a real problem, just not a latency one:
its annotation map is `map[uint32]types.Type` keyed by `StartByte`
(`infer.go:57-60`), so every key after an edit shifts and nothing survives a
re-parse. That costs correctness of caching, not milliseconds.

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
5. **PSC's annotation map is byte-keyed, so nothing survives a splice.** The
   fix is to key by *node identity* rather than byte offset, changing
   `infer.go`'s `map[uint32]types.Type`. This is a correctness-of-caching
   problem, not a latency one — at 0.5-2 ms `Analyze` already fits — so it is
   worth doing only once the parser is fast enough for reuse to matter. See
   §6.10.5.
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
budget, and 20-50× *cheaper* than a single parse with the parser we ship today
(§6.10.2). The intuition that the type checker must be the expensive half is
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
3. **Key annotations by node identity, not byte offset.** `Analyze`'s map is
   `map[uint32]types.Type` keyed by `StartByte` (`infer.go:57-60`), so every
   key after an edit shifts and no annotation survives a re-parse. This is a
   correctness-of-caching problem, not a latency one, and it only becomes
   worth fixing once (1) has made re-parsing cheap enough that reuse matters.
4. **Analyze only subs intersecting the damage region.** Sound when nothing in
   that region changes a declaration, which the Class B trigger scan already
   tells you. Do this last; at 2 ms there is nothing to win yet.

Ship (1) first — the numbers leave no other reading. (2) is free. (3) and (4)
wait for evidence, per step 4 of §6.11.

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
   perl-lsp actually ships, and it is *usable* under 500 KB. Ship it.
4. **Measure.** Instrument every stage (perl-lsp's
   `FIRST_USE_LATENCY_METHODS` list at `runtime/latency.rs:6-16` names the
   right hot paths). Do not optimize before this step produces numbers.
5. **Error recovery** (§6.8). Biggest quality win per line of code —
   completion inside a half-typed statement is what users notice.
6. **Token cache + checkpoints** (§6.3). Now incrementality has a foundation.
7. **Damage/repair** (§6.5), gated behind a flag, with the
   splice-equals-fresh-parse fuzzer (§6.5.4) green before the flag flips.
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
| IR layers | Neither HIR nor PIR initially | Prebuilt dataflow analysis |
| Node API | The 10 methods PSC already calls | Nothing — PSC is unchanged |
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
