// ABOUTME: The lowered vocabulary of §4.14: one Go type per kind inference reads, and Hole for an Unknown.
// ABOUTME: Pure declaration plus a generic child walk; every Node field is optional because the CST is routinely broken.

package lower

import "fmt"

// Node is one lowered node. Every field of type Node or []Node on every kind
// is optional (§6.1.6 item 5): the CST this tree comes from is routinely
// broken, so a nil child is a normal value, never a bug. A Hole may stand in
// any of those fields (§4.14.2, Hole row).
//
// The method is Meta, not the Base of the §4.14 listing: Index, Key and
// Slice have a field named Base, and a field hides a promoted method of the
// same name, so those kinds could not be Nodes.
type Node interface{ Meta() *nodeBase }

// nodeBase is what every node carries besides its kind's own fields.
//
// No span: positions live in the position map, keyed by ID (§6.1.6 item 4).
// No syntax pointer either, so the lowered tree is a value of its own.
type nodeBase struct {
	ID ID

	// Paren is true when the source wrapped this node in parentheses, read
	// from parse.Node.Paren. §4.14.6 asked whether that fact is a Paren node
	// or a bit; it is a bit, here on every node, because:
	//   - a Paren node must be seen through by every rule that matches on an
	//     operand's kind, and a rule that forgets silently loses the operand,
	//     the same tax §4.14.5 refuses for a synthesised Not under `unless`;
	//   - a flag on only the kinds known to read it (Repeat, Assign, Decl)
	//     misses the readers nobody listed: `\(@a)` makes one reference per
	//     element and `\@a` one reference (perl: 3 against 1 for a
	//     three-element @a), so Ref reads it too;
	//   - parse.Node carries it the same way, set at the one site that
	//     erases the parens, so the lowering copies a field and invents
	//     nothing.
	// The paren's bytes still resolve, forward to this node, through the
	// position map. Rule 1 holds: the fact is never folded away.
	Paren bool
}

// Meta returns the node's identity and shared flags.
func (b *nodeBase) Meta() *nodeBase { return b }

// Hole is one Unknown, opaque (§4.14.2). Inference types it Any and never
// inspects it. An Unknown can have children -- `if ($x) {` gives
// `[unknown refusal=unclosed_opener [block]]` -- and a Hole does not lower
// them: it has no children. The position map still covers their bytes,
// because the Hole's span covers them, so a query there answers "no
// information" rather than a type.
type Hole struct{ nodeBase }

// --- Literals and names ---

type NumLit struct {
	nodeBase
	Text  string
	IsInt bool
}

type StrLit struct {
	nodeBase
	Value        string
	Interpolated bool
	Backtick     bool // `cmd` / qx// -- the value is a command
}

// Interp is an interpolating string in pieces: "a$b c" -> [StrLit, Var, StrLit].
type Interp struct {
	nodeBase
	Parts []Node
}

type QwList struct {
	nodeBase
	Words []string
}

type Bareword struct {
	nodeBase
	Name   string
	Quoted bool // from => or {k}
}

// --- Variables and dereference ---

// Var is $x @x %x *x &x $#x. (Sigil, Name) is the identity pair: $_ and @_
// are different variables (§4.14.1 rule 7).
type Var struct {
	nodeBase
	Sigil string
	Name  string
}

// Deref is ${$x} / $$x / $x->@*; Postfix marks ->@*.
type Deref struct {
	nodeBase
	Sigil   string
	Expr    Node
	Postfix bool
}

// Index is $x[0] / $x->[0].
type Index struct {
	nodeBase
	Base  Node
	Idx   Node
	Arrow bool
}

// Key is $h{k} / $x->{k}. Arrow selects WHICH variable: $h{k} reads %h,
// $h->{k} reads $h (§4.14.1 rule 2).
type Key struct {
	nodeBase
	Base  Node
	Key   Node
	Arrow bool
}

// Slice is @a[..] @h{..} %a[..] %h{..} (LIST)[..].
type Slice struct {
	nodeBase
	Base Node
	Idx  Node
	Kind SliceKind
}

type SliceKind uint8

const (
	SliceArray   SliceKind = iota // @a[...]
	SliceHash                     // @h{...}
	SliceKVArray                  // %a[...]
	SliceKVHash                   // %h{...}
	SliceList                     // (LIST)[...]
)

// --- Operators ---

// Unary is ! ~ - \ ++ --.
type Unary struct {
	nodeBase
	Op      string
	Operand Node
	Postfix bool
}

type Binary struct {
	nodeBase
	Op   string
	L, R Node
}

// Repeat is x (§4.10).
type Repeat struct {
	nodeBase
	Left, Count Node
	ListRepeat  bool
}

// Assign's ListAssign comes from the LHS's syntactic shape.
type Assign struct {
	nodeBase
	Op         string
	L, R       Node
	ListAssign bool
}

type Ternary struct {
	nodeBase
	Cond, Then, Else Node
}

// CmpChain is a 5.32+ chained comparison (§4.3).
type CmpChain struct {
	nodeBase
	Operands []Node
	Ops      []string
	Class    TokenClass
}

// TokenClass is the precedence class of a chained comparison.
type TokenClass uint8

const (
	ChRelOp TokenClass = iota // CHRELOP: < > <= >= lt gt le ge
	ChEqOp                    // CHEQOP: == != eq ne
)

// Range is .. vs ...
type Range struct {
	nodeBase
	Lo, Hi    Node
	Exclusive bool
}

// List's Fat[i] is true when element i was followed by =>.
type List struct {
	nodeBase
	Elems []Node
	Fat   []bool
}

// --- Declarations ---

type Decl struct {
	nodeBase
	Kind  DeclKind
	Vars  []Node // Var, or a List of Var for the parenthesised form
	Attrs []string
	Init  Node // nil if none
	// ListAssign marks `my (...) =`, which puts its Init in list context
	// where `my $x =` puts it in scalar (§4.12.2 rule 2).
	ListAssign bool
	RefAlias   bool // my \$x = ...
}

type DeclKind uint8

const (
	DeclMy DeclKind = iota
	DeclOur
	DeclState
	DeclLocal
	DeclField
)

// --- Calls ---

type Call struct {
	nodeBase
	Name string // "" if Fn is set
	Fn   Node   // $coderef->() or &{$x}()
	// Handle is the slot before the list in print $fh LIST / print {$fh}
	// LIST, which is NOT argument 1.
	Handle        Node
	Args          []Node
	Parenthesized bool // affects the paren cliff (§4.8.1) and prototypes
	Ampersand     bool // &foo -- prototypes bypassed
	BareAmpersand bool // &foo with no parens -- inherits caller's @_
	Resolved      bool // false if the parser could not tell call from bareword
}

// MethodCall is $obj->name(...) and indirect object syntax.
//
// Indirect is kept: parse.Node.Indirect sets it for `new Foo(1)` when the
// file has seen package Foo, so it carries signal. There is no Ambiguous
// field: nothing in the CST would ever set it, and §4.14.6 drops a field
// that is always false.
type MethodCall struct {
	nodeBase
	Invocant  Node
	Name      string // "" if Dynamic is set
	Dynamic   Node   // $obj->$name / $obj->$coderef
	Args      []Node
	HasParens bool
	Indirect  bool // "new Foo @args" -- §4.5.3
}

// --- Special forms ---

// MapGrepSort has no BlockGuessed: the lexer runs perl's own intuit_curly, so
// the brace's reading is perl's and there is no guess to record (§4.14.1
// rule 4).
type MapGrepSort struct {
	nodeBase
	Op        string // "map" | "grep" | "sort"
	Block     Node   // block form
	FirstExpr Node   // expression form (map/grep) or comparator (sort)
	List      []Node
}

// DoBlock is do { ... }.
type DoBlock struct {
	nodeBase
	Body Node
}

// DoFile is do EXPR.
type DoFile struct {
	nodeBase
	Path Node
}

// EvalBlock is eval { ... }, statically analysable.
type EvalBlock struct {
	nodeBase
	Body Node
}

// EvalStr is eval EXPR: the soundness boundary, whose result is Any.
type EvalStr struct {
	nodeBase
	Code Node
}

type AnonSub struct {
	nodeBase
	Sig   []Node
	Body  Node
	Attrs []string
}

// ArrayLiteral is [ ... ], lowered from parse.AnonArray. The name is
// B::SoN's (perigrin, 2026-10-11): name agreement with the sibling, and the
// parser's own kind keeps its name.
type ArrayLiteral struct {
	nodeBase
	Elems []Node
}

// HashLiteral is { ... }, lowered from parse.AnonHash; named as ArrayLiteral is.
type HashLiteral struct {
	nodeBase
	Elems []Node
}

// Ref is \EXPR.
type Ref struct {
	nodeBase
	Operand Node
}

// FileTest is -e -d -f ...; a nil Operand means $_.
type FileTest struct {
	nodeBase
	Op      byte
	Operand Node
}

// Readline is <FH> <> <<>>.
type Readline struct {
	nodeBase
	Handle Node
	Magic  bool
}

// Glob is <*.c> / glob EXPR.
type Glob struct {
	nodeBase
	Pattern Node
}

// LoopEx is last/next/redo/goto.
type LoopEx struct {
	nodeBase
	Op    string
	Label Node
}

// Wantarray has the same name in B::SoN.
type Wantarray struct{ nodeBase }

// Match is m// s/// tr/// qr// (§4.11).
type Match struct {
	nodeBase
	Op      string
	Pattern []Node // when Op is tr or y: a character SET, not a pattern
	Replace []Node // likewise; and Flags are c/d/s/r, not regex flags
	Flags   string
	Target  Node // nil means $_
	Negated bool // !~
}

// --- Statements (§5.17's kinds, without spans; lowered in stmt.go) ---

// Block is `{ ... }` holding statements. It is reached from an expression
// too -- the body of do, eval and map -- so expr lowers it.
type Block struct {
	nodeBase
	Stmts []Node
}

// ExprStmt is an expression evaluated as a statement.
type ExprStmt struct {
	nodeBase
	X Node
}

// IfStmt is every conditional: the block form, the modifier form, `unless`
// and `elsif`. `unless` swaps the branches rather than negating Cond
// (§4.14.5), and an `elsif` is an IfStmt in the Else slot. Then and Else
// are a Block, an expression (the modifier form's body) or an IfStmt.
type IfStmt struct {
	nodeBase
	Cond, Then, Else Node
}

// BlockStmt is a bare block, a loop that runs once: last and next work in
// it. Labels are Barewords carrying the label nodes' ids, as LoopEx's target
// is; they stack, `A: B: { }` giving two.
type BlockStmt struct {
	nodeBase
	Labels         []Node
	Body, Continue Node
}

// WhileStmt is `while` and `until`, block form and modifier form. A plain
// `until` is Negated: one loop spelled two ways. PostCond marks the
// `do BLOCK while`/`until` form, which tests after its body and so runs it
// at least once (perl: `$m++ until 1` leaves 0, `do { $n++ } until 1`
// leaves 1). A nil Cond is an infinite loop.
type WhileStmt struct {
	nodeBase
	Labels               []Node
	Cond, Body, Continue Node
	Negated, PostCond    bool
}

// CForStmt is `for (INIT; COND; STEP)`; an empty slot is nil.
type CForStmt struct {
	nodeBase
	Labels                 []Node
	Init, Cond, Post, Body Node
}

// ForeachStmt is `for`/`foreach` over a list, and the modifier form. Var is
// the loop variable as written -- `my $i` a declaration, `$i` a Var, `\my
// $x` a Ref -- and nil means $_.
type ForeachStmt struct {
	nodeBase
	Labels                    []Node
	Var, List, Body, Continue Node
}

// UseStmt is `use`, `no` and `require` (§5.9.4). Module is the module's
// name, "" for `use VERSION` and for `require EXPR`; Version is the version
// to check, "" if none; Expr is require's operand when it is not a module
// name -- `require "foo.pl"`, `require $f`; Args is the import list, nil if
// none, and `use Foo ()`'s empty list is kept because it suppresses import.
type UseStmt struct {
	nodeBase
	Op      string // "use" | "no" | "require"
	Module  string
	Version string
	Expr    Node
	Args    Node
}

// PhaserStmt is BEGIN, END, CHECK, INIT or UNITCHECK and its block.
type PhaserStmt struct {
	nodeBase
	Kind string
	Body Node
}

// Children returns a node's non-nil children in field order. A nil field or
// a nil slice element is skipped, never returned. A kind with no case here
// panics: that is a kind declared without its walk, a bug in this file, and
// TestZeroValueNodeIsUsable lists every kind so it cannot ship.
func Children(n Node) []Node {
	var kids []Node
	add := func(ns ...Node) {
		for _, c := range ns {
			if c != nil {
				kids = append(kids, c)
			}
		}
	}
	switch n := n.(type) {
	case *Hole, *NumLit, *StrLit, *QwList, *Bareword, *Var, *Wantarray:
	case *Interp:
		add(n.Parts...)
	case *Deref:
		add(n.Expr)
	case *Index:
		add(n.Base, n.Idx)
	case *Key:
		add(n.Base, n.Key)
	case *Slice:
		add(n.Base, n.Idx)
	case *Unary:
		add(n.Operand)
	case *Binary:
		add(n.L, n.R)
	case *Repeat:
		add(n.Left, n.Count)
	case *Assign:
		add(n.L, n.R)
	case *Ternary:
		add(n.Cond, n.Then, n.Else)
	case *CmpChain:
		add(n.Operands...)
	case *Range:
		add(n.Lo, n.Hi)
	case *List:
		add(n.Elems...)
	case *Decl:
		add(n.Vars...)
		add(n.Init)
	case *Call:
		add(n.Fn, n.Handle)
		add(n.Args...)
	case *MethodCall:
		add(n.Invocant, n.Dynamic)
		add(n.Args...)
	case *MapGrepSort:
		add(n.Block, n.FirstExpr)
		add(n.List...)
	case *DoBlock:
		add(n.Body)
	case *DoFile:
		add(n.Path)
	case *EvalBlock:
		add(n.Body)
	case *EvalStr:
		add(n.Code)
	case *AnonSub:
		add(n.Sig...)
		add(n.Body)
	case *ArrayLiteral:
		add(n.Elems...)
	case *HashLiteral:
		add(n.Elems...)
	case *Ref:
		add(n.Operand)
	case *FileTest:
		add(n.Operand)
	case *Readline:
		add(n.Handle)
	case *Glob:
		add(n.Pattern)
	case *LoopEx:
		add(n.Label)
	case *Match:
		add(n.Pattern...)
		add(n.Replace...)
		add(n.Target)
	case *Block:
		add(n.Stmts...)
	case *ExprStmt:
		add(n.X)
	case *IfStmt:
		add(n.Cond, n.Then, n.Else)
	case *BlockStmt:
		add(n.Labels...)
		add(n.Body, n.Continue)
	case *WhileStmt:
		add(n.Labels...)
		add(n.Cond, n.Body, n.Continue)
	case *CForStmt:
		add(n.Labels...)
		add(n.Init, n.Cond, n.Post, n.Body)
	case *ForeachStmt:
		add(n.Labels...)
		add(n.Var, n.List, n.Body, n.Continue)
	case *UseStmt:
		add(n.Expr, n.Args)
	case *PhaserStmt:
		add(n.Body)
	default:
		panic(fmt.Sprintf("lower.Children: no case for %T", n))
	}
	return kids
}

// IsDeref reports whether n is a subscript that reads through a reference:
// it is spelled with -> or its base is itself a dereference -- a Deref, as
// in $$x[0], or another subscript, as in $h{a}[0], where perl implies the
// arrow. Arrow alone is not this: $$x[0] has no arrow (§4.14.1 rule 2).
//
// It lives beside Children because it switches on the kinds declared here.
func IsDeref(n Node) bool {
	var base Node
	switch n := n.(type) {
	case *Index:
		if n.Arrow {
			return true
		}
		base = n.Base
	case *Key:
		if n.Arrow {
			return true
		}
		base = n.Base
	}
	switch base.(type) {
	case *Deref, *Index, *Key:
		return true
	}
	return false
}
