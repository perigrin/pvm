<!-- ABOUTME: Chapter 4 of the Perl parser specification: the expression grammar — operators, precedence, associativity, and every term form. -->
<!-- ABOUTME: Ground truth is perly.y's %left/%right/%nonassoc block; this chapter turns it into a binding-power table a Go Pratt parser can encode directly. -->

# 4. Expression Grammar

## 4.0 Ground truth and how to read this chapter

Perl's precedence is not a matter of opinion or of `perlop`'s prose. It is the
ordered block of `%left` / `%right` / `%nonassoc` declarations in
`perly.y:150-183`. Bison assigns increasing precedence in declaration order, so
the first declaration is the loosest-binding and the last is the tightest. That
block, verbatim from `perly.y`, is:

```
%nonassoc <ival> PREC_LOW              /* 150 */
%nonassoc LOOPEX
%nonassoc <pval> PLUGIN_LOW_OP
%left <ival> OROP  <pval> PLUGIN_LOGICAL_OR_LOW_OP
%left <ival> ANDOP <pval> PLUGIN_LOGICAL_AND_LOW_OP
%right <ival> NOTOP
%nonassoc LSTOP LSTOPSUB BLKLSTOP
%left PERLY_COMMA
%right <ival> ASSIGNOP <pval> PLUGIN_ASSIGN_OP
%right <ival> PERLY_QUESTION_MARK PERLY_COLON
%nonassoc DOTDOT
%left <ival> OROR DORDOR <pval> PLUGIN_LOGICAL_OR_OP
%left <ival> ANDAND <pval> PLUGIN_LOGICAL_AND_OP
%left <ival> BITOROP
%left <ival> BITANDOP
%left <ival> CHEQOP NCEQOP
%left <ival> CHRELOP NCRELOP
%nonassoc <pval> PLUGIN_REL_OP
%nonassoc UNIOP UNIOPSUB
%nonassoc KW_REQUIRE
%left <ival> SHIFTOP
%left ADDOP <pval> PLUGIN_ADD_OP
%left MULOP <pval> PLUGIN_MUL_OP
%left <ival> MATCHOP
%right <ival> PERLY_EXCLAMATION_MARK PERLY_TILDE UMINUS REFGEN
%right POWOP <pval> PLUGIN_POW_OP
%nonassoc <ival> PREINC PREDEC POSTINC POSTDEC POSTJOIN
%nonassoc <pval> PLUGIN_HIGH_OP
%left <ival> ARROW
%nonassoc <ival> PERLY_PAREN_CLOSE
%left <ival> PERLY_PAREN_OPEN
%left PERLY_BRACKET_OPEN PERLY_BRACE_OPEN     /* 183 */
```

Everything in §4.1 is derived from those 32 lines. Where PerlOnJava or perl-lsp
disagrees, this chapter follows `perly.y` and flags the disagreement — those
flags are collected in §4.13.

Three things about the list surprise people, and all three matter to an
implementer:

1. **Precedence is a property of the token, not of the spelling.** `toke.c`
   emits `ADDOP` for `+`, `-` *and* `.`; `MULOP` for `*`, `/`, `%` *and* `x`.
   The precedence table has one entry each; the operator identity rides along
   in `pl_yylval.ival`. A Go implementation should do the same: classify to a
   token class, carry the concrete operator as a field.
2. **Named unary operators and list operators are precedence levels**, not
   special-cases bolted on. `UNIOP` sits between `PLUGIN_REL_OP` and
   `SHIFTOP`; `LSTOP` sits between `NOTOP` and `PERLY_COMMA`. Everything
   `perlop` says about "named unary operators" and "list operators (rightward)"
   falls out of those two lines.
3. **The last four lines are not really binary operators.** `ARROW`,
   `PERLY_PAREN_OPEN`, `PERLY_BRACKET_OPEN` and `PERLY_BRACE_OPEN` are
   declared at the top so that shift/reduce conflicts on postfix constructs
   resolve in favour of shifting — i.e. so `$x->[0]{k}(1)` chains. In a Pratt
   parser they are postfix/led handlers at the highest binding power, not
   infix operators.

---

## 4.1 The complete precedence table

Levels are numbered from loosest (1) to tightest (32), matching `perly.y`
declaration order. The **BP** column is the binding power recommended for a Go
Pratt parser (§4.2); it is `level * 10`, leaving room to insert.

| # | BP | `perly.y` token | Assoc | Concrete operators | Notes |
|---|---|---|---|---|---|
| 1 | 10 | `PREC_LOW` | nonassoc | — | Pseudo-token. Used only via `%prec` to force a reduction (`perly.y:1262`, `1273`, `1436`). Never lexed. |
| 2 | 20 | `LOOPEX` | nonassoc | `goto` `last` `next` `redo` `dump` | `perly.y:1663-1667`. Takes an optional term. |
| 3 | 30 | `PLUGIN_LOW_OP` | nonassoc | (XS infix plugins) | `PL_infix_plugin`. Ignore unless supporting plugins. |
| 4 | 40 | `OROP` | **left** | `or` `xor` | `toke.c:8824` (`or`), `toke.c:9224-9226` (`xor`). Both are `OROP`; `xor` sets `pl_yylval.ival = OP_XOR`. **`xor` is the same level as `or`, not its own.** |
| 5 | 50 | `ANDOP` | **left** | `and` | `toke.c:8349`. |
| 6 | 60 | `NOTOP` | **right** | `not` | `toke.c:8796`. Prefix only. Takes a `listexpr` (`perly.y:1669`) — it swallows commas. |
| 7 | 70 | `LSTOP` `LSTOPSUB` `BLKLSTOP` | nonassoc | `print` `push` `join` `die` `warn` `sort` `map` `grep` `return` … and any user sub with a list prototype | The "list operators (rightward)" level. `toke.c:2152 #define LOP(f,x)`. |
| 8 | 80 | `PERLY_COMMA` | **left** | `,` `=>` | `perly.y:1266-1273`. `=>` is a comma that autoquotes its left bareword (§4.5.4). |
| 9 | 90 | `ASSIGNOP` | **right** | `=` `+=` `-=` `*=` `/=` `.=` `%=` `**=` `x=` `\|\|=` `&&=` `//=` `\|=` `&=` `^=` `<<=` `>>=` `\|.=` `&.=` `^.=` `^^=` | `toke.c:250 #define AOPERATOR`. All assignment ops are one token class. |
| 10 | 100 | `PERLY_QUESTION_MARK` `PERLY_COLON` | **right** | `?:` | `perly.y:1568`. Right associative: `$a ? $b : $c ? $d : $e` groups as `$a ? $b : ($c ? $d : $e)`. |
| 11 | 110 | `DOTDOT` | **nonassoc** | `..` `...` | `perly.y:1440`. **Non-associative**: `1 .. 2 .. 3` is a syntax error. |
| 12 | 120 | `OROR` `DORDOR` | **left** | `\|\|` `//` `^^` | `toke.c:6970` (`\|\|`), `toke.c:7065` (`//`), `toke.c:6436-6441` (`^^`). **`^^` is emitted as `OROR`** — it shares `\|\|`'s level exactly. |
| 13 | 130 | `ANDAND` | **left** | `&&` | `toke.c:6913`. |
| 14 | 140 | `BITOROP` | **left** | `\|` `^` `\|.` `^.` | `toke.c:265 BOop`, `toke.c:6984`, `toke.c:6452`. |
| 15 | 150 | `BITANDOP` | **left** | `&` `&.` | `toke.c:266 BAop`, `toke.c:6938-6940`. |
| 16 | 160 | `CHEQOP` `NCEQOP` | **left** | chaining: `==` `!=` `eq` `ne` — non-chaining: `<=>` `cmp` `~~` | `toke.c:275-276`. See §4.3 on comparison chaining. `~~` lexes only while the `smartmatch` feature is on — off under `use v5.42` (chapter 3 §3.2). |
| 17 | 170 | `CHRELOP` `NCRELOP` | **left** | chaining: `<` `>` `<=` `>=` `lt` `gt` `le` `ge` — non-chaining: `isa` | `toke.c:277-278`, `toke.c:8697 NCRop(OP_ISA)`. **`isa` is at relational level and is non-chaining.** |
| 18 | 180 | `PLUGIN_REL_OP` | nonassoc | (plugins) | |
| 19 | 190 | `UNIOP` `UNIOPSUB` | **nonassoc** | named unaries: `defined` `ref` `scalar` `lc` `uc` `length` `int` `abs` `exists` `delete` `each` `keys` `values` `shift` `pop` `chr` `ord` `hex` `oct` `log` `sqrt` `sin` `cos` `rand` `srand` `quotemeta` `readline` `caller` `sleep` `exit` `chdir` `rmdir` `stat` `lstat` `undef` `local`(§4.6) `my`(§4.6) `do EXPR`(§4.7); **plus all file test operators** `-e -f -d -r -w -x -s -z -l -p -S -b -c -t -u -g -k -T -B -A -M -C -o -R -W -X -O` | `toke.c:296 #define UNI(f)`; file tests via `toke.c:6255-6261 FTST(ftst)` which returns `UNIOP`. |
| 20 | 200 | `KW_REQUIRE` | nonassoc | `require` | Its own level purely so `require Foo::Bar` parses the bareword specially (`perly.y:1673-1676`). |
| 21 | 210 | `SHIFTOP` | **left** | `<<` `>>` | `toke.c:270`-adjacent; emitted from `yyl_leftpointy`/`yyl_rightpointy`. |
| 22 | 220 | `ADDOP` | **left** | `+` `-` `.` | `toke.c:272 #define Aop`, `toke.c:6314`, `toke.c:6343`. **String concat `.` is an ADDOP** — same level as `+`. |
| 23 | 230 | `MULOP` | **left** | `*` `/` `%` `x` | `toke.c:274 #define Mop`, `toke.c:6387`, `toke.c:6401`, `toke.c:9216 Mop(OP_REPEAT)`. **Repetition `x` is a MULOP.** |
| 24 | 240 | `MATCHOP` | **left** | `=~` `!~` | `toke.c:271 #define PMop`, `toke.c:7018`. |
| 25 | 250 | `!` `~` `UMINUS` `REFGEN` | **right** | `!` `~` `~.` unary `-` unary `+` `\` | `perly.y:1497-1504` (`UMINUS`), `toke.c:7297 OPERATOR(REFGEN)`. Unary `+` is a no-op that exists solely to disambiguate (`perly.y:1499`). |
| 26 | 260 | `POWOP` | **right** | `**` | `toke.c:270 PWop`, `toke.c:6376`. Right associative *and tighter than unary minus*: `-2**2 == -4`. |
| 27 | 270 | `PREINC` `PREDEC` `POSTINC` `POSTDEC` `POSTJOIN` | **nonassoc** | `++` `--` (both fixities) | `perly.y:1509-1528`. `POSTJOIN` is the implicit `join` after an interpolated `->@*` (`perly.y:1515-1522`). |
| 28 | 280 | `PLUGIN_HIGH_OP` | nonassoc | (plugins) | `toke.c:9287`. |
| 29 | 290 | `ARROW` | **left** | `->` | Postfix in practice. §4.4, §4.5. |
| 30 | 300 | `PERLY_PAREN_CLOSE` | nonassoc | `)` | Conflict-resolution only. |
| 31 | 310 | `PERLY_PAREN_OPEN` | **left** | `(` | Postfix call. Also the `%prec` used to make terms bind maximally tight (`perly.y:1582-1590`). |
| 32 | 320 | `PERLY_BRACKET_OPEN` `PERLY_BRACE_OPEN` | **left** | `[` `{` | Postfix subscript. Tightest thing in the grammar. |

### 4.1.1 The `and or not xor` versus `&& || !` cliff

The single most consequential fact in the table: levels 4-6 (`or`, `and`,
`not`) sit **below** the comma at level 8, while levels 12-13 (`||`, `&&`) and
25 (`!`) sit **above** assignment at level 9. This is why:

```perl
my $x = $a || $b;      # my $x = ($a || $b);   -- || is tighter than =
my $x = $a or $b;      # (my $x = $a) or $b;   -- or is looser than =
                       # $x always gets $a. This is the classic bug.

open my $fh, '<', $f or die "no: $!";   # correct: `or` is looser than the
                                        # list operator `open`, so `open` gets
                                        # all three args and `or` sees the
                                        # whole call as its left operand.

open my $fh, '<', $f || die "no: $!";   # WRONG: `||` is tighter than `,`, so
                                        # this is open(my $fh, '<', ($f || die))
```

And symmetrically for `not` versus `!`:

```perl
print not 1, 2;        # not takes a *listexpr* (perly.y:1669), so it is
                       # not((1,2)) -> !2 in scalar context. Deparses to
                       # `print((!1))` because the comma expression's value
                       # in scalar context is its last element, constant-folded.
print !1, 2;           # ! binds tighter than , : print((!1), 2)
```

### 4.1.2 Named unary operators: the level-19 rule

Level 19 sits *below* `SHIFTOP`/`ADDOP`/`MULOP` and *above* the comparisons.
That single placement explains every named-unary surprise:

```perl
length $x + 1          # length($x + 1)      -- ADDOP is tighter, verified
                       #                        by B::Deparse -p
length($x) + 1         # (length($x)) + 1    -- parens make it FUNC1
defined $x && $y       # defined($x) && $y   -- ANDAND is looser than UNIOP
ref $x eq 'HASH'       # ref($x) eq 'HASH'   -- CHEQOP is looser. This is why
                       #                        the idiom works at all.
-e $file . '.bak'      # -e ($file . '.bak') -- file tests are named unaries.
                       #    B::Deparse: `-e 'fx'` for `-e "f" . "x"`.
```

The last one is a genuine trap and the reason `-e` sits at level 19 rather than
at a level of its own: `-s $file > 1024` parses as `(-s $file) > 1024` (good),
but `-e $dir . '/f'` parses as `-e ($dir . '/f')` (surprising, but consistent).

### 4.1.3 Operators absent from the table

`perlop` documents some things that are not precedence levels:

- **`sub{}`, `[]`, `{}`, `qw()`, `<FH>`, literals, variables** — these are
  terms (nuds), not operators. They have no binding power.
- **`local`, `my`, `our`, `state`** — `my` reduces via `myattrterm %prec UNIOP`
  (`perly.y:1572`) and `local` via `KW_LOCAL term %prec UNIOP`
  (`perly.y:1574`). They are level 19 by `%prec`, not by their own declaration.
- **`do BLOCK`** — `%prec PERLY_PAREN_OPEN`, i.e. level 31, a term
  (`perly.y:1560`). `do EXPR` is `%prec UNIOP`, level 19 (`perly.y:1558`).

---

## 4.2 The Go binding-power table

This is the artifact to encode. A Pratt parser needs, per token, a *left
binding power* (how tightly it grabs the expression to its left) and, for
right-associative operators, a slightly lower power to pass down when parsing
the right operand.

```go
// ABOUTME: Operator binding powers for the Perl expression parser.
// ABOUTME: Derived from perly.y:150-183; declaration order is precedence order.

package parser

type Assoc uint8

const (
	AssocLeft Assoc = iota
	AssocRight
	AssocNone // %nonassoc: a second occurrence at the same level is an error
)

// OpInfo is the Pratt table entry for an infix or postfix operator.
type OpInfo struct {
	BP    int    // left binding power
	Assoc Assoc
	Class TokenClass // ADDOP, MULOP, ... the perly.y token this maps to
}

// RightBP returns the power to pass to parseExpr when parsing the right
// operand. Left-associative operators pass BP so an equal-power operator to
// the right stops; right-associative pass BP-1 so it continues.
func (o OpInfo) RightBP() int {
	if o.Assoc == AssocRight {
		return o.BP - 1
	}
	return o.BP
}

// infixBP is the complete infix/postfix table. Prefix operators are handled in
// the nud dispatch (see prefixBP below) and do not appear here.
var infixBP = map[string]OpInfo{
	// level 4-5: lowest logical. perly.y:154-155
	"or":  {40, AssocLeft, OROP},
	"xor": {40, AssocLeft, OROP}, // NOT its own level -- toke.c:9224
	"and": {50, AssocLeft, ANDOP},

	// level 8: comma. perly.y:158
	",":  {80, AssocLeft, COMMA},
	"=>": {80, AssocLeft, COMMA}, // autoquotes LHS bareword

	// level 9: assignment, right assoc. perly.y:159 / toke.c:250
	"=": {90, AssocRight, ASSIGNOP}, "+=": {90, AssocRight, ASSIGNOP},
	"-=": {90, AssocRight, ASSIGNOP}, "*=": {90, AssocRight, ASSIGNOP},
	"/=": {90, AssocRight, ASSIGNOP}, ".=": {90, AssocRight, ASSIGNOP},
	"%=": {90, AssocRight, ASSIGNOP}, "**=": {90, AssocRight, ASSIGNOP},
	"x=": {90, AssocRight, ASSIGNOP}, "||=": {90, AssocRight, ASSIGNOP},
	"&&=": {90, AssocRight, ASSIGNOP}, "//=": {90, AssocRight, ASSIGNOP},
	"|=": {90, AssocRight, ASSIGNOP}, "&=": {90, AssocRight, ASSIGNOP},
	"^=": {90, AssocRight, ASSIGNOP}, "<<=": {90, AssocRight, ASSIGNOP},
	">>=": {90, AssocRight, ASSIGNOP}, "|.=": {90, AssocRight, ASSIGNOP},
	"&.=": {90, AssocRight, ASSIGNOP}, "^.=": {90, AssocRight, ASSIGNOP},
	"^^=": {90, AssocRight, ASSIGNOP},

	// level 10: ternary, right assoc. perly.y:160
	"?": {100, AssocRight, QUESTION},

	// level 11: range, NONASSOC. perly.y:161
	"..": {110, AssocNone, DOTDOT}, "...": {110, AssocNone, DOTDOT},

	// level 12-13. perly.y:162-163
	"||": {120, AssocLeft, OROR}, "//": {120, AssocLeft, DORDOR},
	"^^": {120, AssocLeft, OROR}, // toke.c:6441 -- shares || level
	"&&": {130, AssocLeft, ANDAND},

	// level 14-15: bitwise. perly.y:164-165
	"|": {140, AssocLeft, BITOROP}, "^": {140, AssocLeft, BITOROP},
	"|.": {140, AssocLeft, BITOROP}, "^.": {140, AssocLeft, BITOROP},
	"&": {150, AssocLeft, BITANDOP}, "&.": {150, AssocLeft, BITANDOP},

	// level 16: equality. perly.y:166. Chaining handled separately (4.3).
	"==": {160, AssocLeft, CHEQOP}, "!=": {160, AssocLeft, CHEQOP},
	"eq": {160, AssocLeft, CHEQOP}, "ne": {160, AssocLeft, CHEQOP},
	"<=>": {160, AssocLeft, NCEQOP}, "cmp": {160, AssocLeft, NCEQOP},
	"~~": {160, AssocLeft, NCEQOP},

	// level 17: relational. perly.y:167
	"<": {170, AssocLeft, CHRELOP}, ">": {170, AssocLeft, CHRELOP},
	"<=": {170, AssocLeft, CHRELOP}, ">=": {170, AssocLeft, CHRELOP},
	"lt": {170, AssocLeft, CHRELOP}, "gt": {170, AssocLeft, CHRELOP},
	"le": {170, AssocLeft, CHRELOP}, "ge": {170, AssocLeft, CHRELOP},
	"isa": {170, AssocLeft, NCRELOP}, // toke.c:8697 -- NON-chaining

	// level 21-23. perly.y:171-173
	"<<": {210, AssocLeft, SHIFTOP}, ">>": {210, AssocLeft, SHIFTOP},
	"+": {220, AssocLeft, ADDOP}, "-": {220, AssocLeft, ADDOP},
	".": {220, AssocLeft, ADDOP}, // concat is an ADDOP
	"*": {230, AssocLeft, MULOP}, "/": {230, AssocLeft, MULOP},
	"%": {230, AssocLeft, MULOP},
	"x": {230, AssocLeft, MULOP}, // repetition is a MULOP

	// level 24: binding. perly.y:174
	"=~": {240, AssocLeft, MATCHOP}, "!~": {240, AssocLeft, MATCHOP},

	// level 26: exponentiation, right assoc, TIGHTER than unary minus.
	// perly.y:176
	"**": {260, AssocRight, POWOP},

	// level 27: postfix inc/dec (as led). perly.y:177
	"++": {270, AssocNone, POSTINC}, "--": {270, AssocNone, POSTDEC},

	// level 29-32: postfix chain. perly.y:180-183
	"->": {290, AssocLeft, ARROW},
	"(":  {310, AssocLeft, PAREN_OPEN},
	"[":  {320, AssocLeft, BRACKET_OPEN},
	"{":  {320, AssocLeft, BRACE_OPEN},
}

// prefixBP is the binding power a prefix operator passes down when parsing its
// operand. These come from the %prec annotations and the unary rules.
var prefixBP = map[string]int{
	"not": 60,  // perly.y:156, takes a listexpr -- swallows commas
	"!":   250, // perly.y:1501
	"~":   250, // perly.y:1503
	"~.":  250,
	"-":   250, // UMINUS, perly.y:1497
	"+":   250, // UMINUS, perly.y:1499 -- a no-op that forces term parsing
	"\\":  250, // REFGEN, perly.y:1571
	"++":  270, // PREINC, perly.y:1523
	"--":  270, // PREDEC, perly.y:1526
}

// Named unaries and list operators are not in prefixBP because their operand
// power depends on whether a '(' immediately follows. See 4.8.
const (
	bpNamedUnary  = 190 // UNIOP, perly.y:169
	bpListOp      = 70  // LSTOP, perly.y:157
	bpFuncCall    = 310 // a '(' immediately after the name promotes to FUNC
)
```

### 4.2.1 The driving loop

```go
// parseExpr parses an expression, consuming infix operators whose binding
// power exceeds minBP.
func (p *Parser) parseExpr(minBP int) Node {
	left := p.parseTerm() // the nud: literals, variables, prefix ops, ...

	for {
		tok := p.peek()
		op, ok := infixBP[tok.Text]
		if !ok || op.BP <= minBP {
			break
		}
		// Postfix forms consume their own closer and loop again.
		switch tok.Text {
		case "->", "(", "[", "{", "++", "--":
			left = p.parsePostfix(left)
			continue
		}
		p.next()
		if op.Assoc == AssocNone && sameLevel(left, op) {
			p.errorf(tok.Pos, "syntax error: %q is non-associative", tok.Text)
		}
		if tok.Text == "?" {
			// The 'then' branch is bracketed by ':' so it parses at power 0.
			then := p.parseExpr(0)
			p.expect(":")
			els := p.parseExpr(op.RightBP())
			left = &Ternary{Cond: left, Then: then, Else: els}
			continue
		}
		right := p.parseExpr(op.RightBP())
		left = p.buildBinary(tok, left, right)
	}
	return left
}
```

Three things this loop must get right, and each is a place implementations go
wrong:

1. **`op.BP <= minBP` breaks, `>` continues.** With `RightBP() = BP` for
   left-assoc and `BP-1` for right-assoc, this yields correct grouping for
   both. Do not use `<` here; `a - b - c` would become `a - (b - c)`.
2. **The ternary `then` branch parses at power 0**, because `:` terminates it
   unambiguously. The `else` branch parses at `RightBP()` = 99, which is what
   makes `?:` right-associative. (PerlOnJava does exactly this,
   `ParseInfix.java:279-283`, and it is correct.)
3. **`AssocNone` needs an actual check.** `1 .. 2 .. 3` must be a syntax error,
   not silently left-associative. The `sameLevel` helper inspects whether
   `left` is already a binary node at the same declared level. `perly.y` gets
   this from bison for free; a hand-written parser must do it explicitly.

### 4.2.2 What the table cannot express

The binding-power table is necessary but not sufficient. Three constructs
defeat any pure precedence table and require parser state:

- **List operators (§4.8)** whose right extent depends on whether `(`
  immediately followed the name — a *lexer* decision (`toke.c:S_lop`).
- **`sort`/`map`/`grep` block-vs-expression first argument (§4.9)** — requires
  disambiguating `{` as a block versus an anon-hash.
- **Every term-vs-operator ambiguity** (`/` as divide or as `m//` start, `<` as
  less-than or as `<FH>` start, `%`/`&`/`*` as binop or as sigil). This is the
  `PL_expect` state machine and belongs to the lexer chapter; the expression
  parser must feed it back the expected-token state after every reduction.

---

## 4.3 Chained comparisons

Perl 5.32+ has chained comparisons. `perly.y:1462-1495` splits comparison
tokens into two classes:

| Class | Operators | Chains? |
|---|---|---|
| `CHRELOP` | `<` `>` `<=` `>=` `lt` `gt` `le` `ge` | yes |
| `NCRELOP` | `isa` | no |
| `CHEQOP` | `==` `!=` `eq` `ne` | yes |
| `NCEQOP` | `<=>` `cmp` `~~` | no |

Chaining builds a single n-ary comparison with each operand evaluated once:

```perl
if ($lo <= $x && $x <= $hi) { }   # old
if ($lo <= $x <= $hi) { }          # 5.32+, $x evaluated once
```

The grammar rules are worth reading directly (`perly.y:1466-1478`):

```
relopchain:	term CHRELOP term       { cmpchain_start(...) }
	|	relopchain CHRELOP term  { cmpchain_extend(...) }
	;
termrelop:	relopchain %prec PREC_LOW { cmpchain_finish(...) }
	|	term NCRELOP term         { newBINOP(...) }
	|	termrelop NCRELOP  { yyerror("syntax error"); YYERROR; }
	|	termrelop CHRELOP  { yyerror("syntax error"); YYERROR; }
	;
```

The two `yyerror` productions are the entire non-chaining rule: **once you have
reduced a `termrelop`, no further relational operator may follow.** So:

```perl
$a < $b < $c        # OK -- chained, one CmpChain node
$a <=> $b <=> $c    # syntax error -- NCEQOP cannot chain
$a isa Foo isa Bar  # syntax error -- NCRELOP cannot chain
$a < $b isa Foo     # syntax error -- termrelop followed by NCRELOP
```

Note that a chain and a non-chaining operator at the *same* level still
conflict, which is why the table alone is insufficient. In Go:

```go
// After building any comparison node, mark it. A subsequent comparison at the
// same level either extends the chain (both chainable, same class) or errors.
type CmpChain struct {
	Operands []Node   // n+1 operands
	Ops      []string // n operators
	Class    TokenClass // CHRELOP or CHEQOP -- may not mix
}
```

Verified against perl 5.42: `perl -MO=Deparse,-p -e 'my $q = $a < $b < $c'`
produces a chained form, not nested binaries.

---

## 4.4 Term forms

`perly.y:1565-1712` enumerates every term. This section walks that list.

### 4.4.1 Literals

Literals arrive from the lexer as `THING` (`perly.y:1634`, `%prec
PERLY_PAREN_OPEN`) — numbers, single-quoted strings, and fully-constant
double-quoted strings. Interpolated strings arrive already parsed into a
concatenation/`join` tree by `toke.c`'s sublexer; a Go implementation should
produce an explicit `Interp` node instead (§4.11).

```perl
42        0x2a      0b101010     052      1_000_000    1.5e3
'literal'  "interp $x"  `backtick`  q{}  qq{}  qx{}
__FILE__  __LINE__  __PACKAGE__  __SUB__     # FUNC0OP, perly.y:1685
```

### 4.4.2 Variables and sigils

Every sigil form reduces through the same `indirob` non-terminal
(`perly.y:1884-1894`):

```
indirob : BAREWORD | scalar %prec PREC_LOW | block | PRIVATEREF ;
scalar  : PERLY_DOLLAR        indirob  { newSVREF }
ary     : PERLY_SNAIL         indirob  { newAVREF }
hsh     : PERLY_PERCENT_SIGN  indirob  { newHVREF }
star    : PERLY_STAR          indirob  { newGVREF }
amper   : PERLY_AMPERSAND     indirob  { newCVREF }
arylen  : DOLSHARP            indirob  { newAVREF }
```

That one production generates all of these uniformly:

```perl
$x     @x     %x     *x     &x     $#x        # indirob = BAREWORD
$$x    @$x    %$x    *$x    &$x    $#$x       # indirob = scalar
${$x}  @{$x}  %{$x}  *{$x}  &{$x}  $#{$x}     # indirob = block
${ $h->{k} }  @{ [ 1, 2 ] }                    # block, arbitrary expression
$$$x   @$$x                                    # nests: indirob = scalar = ...
```

The `%prec PREC_LOW` on `scalar` in `indirob` (`perly.y:1888`) is what makes
`$$x[0]` parse as `${$x}[0]` (element of `@$x`) and not `${$x[0]}`. A Go
implementation gets this by parsing the sigil's operand as a *tightly bound*
term — a single variable, a `{...}` block, or another sigil chain — and then
letting the postfix loop attach subscripts to the whole thing.

```go
// parseSigil handles $ @ % * & $#, each with the same operand grammar.
func (p *Parser) parseSigil(sigil string) Node {
	switch {
	case p.at("{"):
		return &Deref{Sigil: sigil, Expr: p.parseBlockAsExpr()}
	case p.atSigil(): // $$x, @$x, ...
		return &Deref{Sigil: sigil, Expr: p.parseSigil(p.next().Text)}
	default:
		return &Var{Sigil: sigil, Name: p.parseVarName()}
	}
}
```

Special variables (`$_`, `$0`, `$1`, `$@`, `$!`, `$/`, `${^GLOBAL_PHASE}`) are
lexer concerns; the parser sees them as `Var` nodes with unusual names.

### 4.4.3 Declarations: `my` / `our` / `state` / `local` / `field`

```
myattrterm : KW_MY myterm attrlist  | KW_MY myterm
           | KW_MY REFGEN myterm attrlist | KW_MY REFGEN term      # perly.y:1717-1725
myterm     : '(' expr ')' | '()' | scalar | hsh | ary               # perly.y:1728-1741
term       : myattrterm %prec UNIOP | KW_LOCAL term %prec UNIOP     # perly.y:1572-1574
```

`myterm` is deliberately restrictive: only a parenthesised list or a single
`$`/`@`/`%` variable. `local`, by contrast, takes an arbitrary `term`
(`perly.y:1574`), which is why `local $h{key}` and `local $INC[0]` are legal
but `my $h{key}` is not.

```perl
my $x;                    my ($a, $b) = @_;      my @list;   my %h;
my $x :shared;                                   # attrlist
my \$ref = \$other;                              # KW_MY REFGEN, 5.26+ aliasing
our $pkg_var;             state $counter = 0;
local $/ = undef;         local $h{key};         local @a[1,2];   # arbitrary term
field $x :param = 1;                             # perly.y:1758-1780, 5.38+
```

`our` and `state` are lexed to the same `KW_MY` token with a flag in
`toke.c`; the grammar does not distinguish them. A Go AST **should**
distinguish them, because PSC needs to know scope and initialisation
semantics (`state` initialises once).

### 4.4.4 Anonymous constructors

```
anonymous : '[' optexpr ']'                       { newANONLIST }   # perly.y:1534
          | HASHBRACK optexpr ';' '}' %prec '('   { newANONHASH }   # perly.y:1536
          | KW_SUB_anon startanonsub proto subattrlist subbody      # perly.y:1538
          | KW_SUB_anon_sig ... sigsubbody                          # perly.y:1543
          | KW_METHOD_anon ... sigsubbody                           # perly.y:1548
```

```perl
[ 1, 2, 3 ]              # arrayref
{ a => 1, b => 2 }       # hashref -- IF the lexer decided HASHBRACK
sub { $_[0] * 2 }        # coderef
sub ($x, $y) { $x + $y } # coderef with signature (KW_SUB_anon_sig)
method ($x) { ... }      # 5.38+ anon method
```

`HASHBRACK` versus a bare block is decided in `yyl_leftcurly`
(`toke.c:6643`; the heuristic at `6698-6842`) by `PL_expect` plus a lookahead heuristic that chapter 3
§3.3 specifies step by step — including the lowercase-bareword asymmetry a
one-line paraphrase loses. It is a documented heuristic, not a rule, and
Perl's own advice is to disambiguate with `+{...}` or `{; ...}`:

```perl
map { $_ => 1 } @list      # perl guesses BLOCK (ch3 §3.3 step 6: `$` does not move the scan); with "$_" it guesses HASH and fails
map { ($_ => 1) } @list    # forced block
map { +{ $_ => 1 } } @list # block returning a hashref
map {; $_ => 1 } @list     # forced block via leading semicolon
```

A Go parser must reproduce the heuristic *and* record that it guessed, so the
LSP can surface the ambiguity and PSC can widen its type conclusion.

### 4.4.5 References: `\`

`REFGEN` is level 25, right associative with the other unaries
(`perly.y:176`, rule at `perly.y:1571`):

```perl
\$x      \@a      \%h      \&code     \*glob
\ ( $a, $b )     # distributes! yields (\$a, \$b), not a ref to a list
\\$x             # ref to a ref -- right associative
```

The distribution over a parenthesised list is a runtime property of
`OP_REFGEN`, not a parse property. The parse is `Ref{ Paren{ List } }`; PSC
must know that `\(LIST)` has type `List of Ref`, while `\@a` is a single
`Ref[Array]`.

### 4.4.6 Dereference syntax, all forms

Perl offers three spellings of the same operation, plus postfix deref (5.24+):

| Operation | Circumfix | Prefix | Arrow | Postfix deref |
|---|---|---|---|---|
| whole scalar | `${$x}` | `$$x` | — | `$x->$*` |
| whole array | `@{$x}` | `@$x` | — | `$x->@*` |
| whole hash | `%{$x}` | `%$x` | — | `$x->%*` |
| whole code | `&{$x}` | `&$x` | — | `$x->&*` |
| whole glob | `*{$x}` | `*$x` | — | `$x->**` |
| last index | `${#$x}`* | `$#$x` | — | `$x->$#*` |
| array elem | `${$x}[0]` | `$$x[0]` | `$x->[0]` | — |
| hash elem | `${$x}{k}` | `$$x{k}` | `$x->{k}` | — |
| call | `&{$x}(1)` | `&$x(1)` | `$x->(1)` | — |
| array slice | `@{$x}[0,1]` | `@$x[0,1]` | — | `$x->@[0,1]` |
| hash slice | `@{$x}{qw/a b/}` | `@$x{qw/a b/}` | — | `$x->@{'a','b'}` |

\* `$#{$x}` is the real circumfix spelling; `${#$x}` is not valid.

The postfix forms have dedicated grammar rules (`perly.y:1647-1660`):

```
term : term ARROW PERLY_DOLLAR       PERLY_STAR  { newSVREF }
     | term ARROW PERLY_SNAIL        PERLY_STAR  { newAVREF }
     | term ARROW PERLY_PERCENT_SIGN PERLY_STAR  { newHVREF }
     | term ARROW PERLY_AMPERSAND    PERLY_STAR  { ENTERSUB(newCVREF) }
     | term ARROW PERLY_STAR         PERLY_STAR %prec '(' { newGVREF }
arylen : term ARROW DOLSHARP PERLY_STAR          { newAVREF }        # perly.y:1866
sliceme: ary | term ARROW PERLY_SNAIL                                # perly.y:1871
kvslice: hsh | term ARROW PERLY_PERCENT_SIGN                         # perly.y:1876
```

Note the lexer cooperation: `toke.c:6285-6293` sets `PL_expect = XPOSTDEREF`
when it sees `->` followed by `$*`, `&*`, `$#*`, `@*`, `@[`, `@{`, `%*`, `%{`,
`**`, or `*{`. Without that state, `->%*` would lex `%` as modulus.

### 4.4.7 Slices and key/value slices

`perly.y:1594-1633` gives four slice rules:

```perl
@a[1, 2]        # array slice     -> OP_ASLICE     (sliceme '[' expr ']')
@h{'a', 'b'}    # hash slice      -> OP_HSLICE     (sliceme '{' expr ';' '}')
%a[1, 2]        # kv array slice  -> OP_KVASLICE   (kvslice '[' expr ']')  5.20+
%h{'a', 'b'}    # kv hash slice   -> OP_KVHSLICE   (kvslice '{' expr ';' '}') 5.20+
```

The sigil determines the *result shape*, not the container. `@h{...}` returns
values; `%h{...}` returns interleaved key/value pairs. `@a[...]` returns
elements; `%a[...]` returns index/element pairs. This is a substantial fact for
PSC: the same container with different sigils has different result types.

```perl
my @vals  = @h{qw/a b/};     # (v_a, v_b)          -- List
my %pairs = %h{qw/a b/};     # (a => v_a, b => v_b) -- List of even length
my @elems = @a[0, 1];        # ($a[0], $a[1])
my %idx   = %a[0, 1];        # (0 => $a[0], 1 => $a[1])
```

Note the `';'` in the hash-slice rules. `toke.c` inserts a phantom semicolon
before the closing `}` of a hash subscript (see the comment at
`perly.y:1355-1357`) so the grammar can tell `$h{...}` from a block. A hand-
written parser does not need the phantom token; it needs the *state* — after
`{` in subscript position, parse an expression, not a block.

The list-slice forms are separate (`perly.y:1401-1407`):

```perl
(localtime)[5, 4, 3]     # '(' expr ')' '[' expr ']'  -> newSLICEOP
qw(a b c)[1]             # QWLIST '[' expr ']'
()[0]                    # '(' ')' '[' expr ']'  -- empty list slice
```

---

## 4.5 Method calls

### 4.5.1 The forms that exist

From `listop` (`perly.y:1289-1319`) and `subscripted` (`perly.y:1382-1400`):

```perl
$obj->method                  # term ARROW methodname               (perly.y:1295)
$obj->method(@args)           # term ARROW methodname '(' optexpr ')' (perly.y:1289)
Class->method(@args)          # same; the bareword is the invocant
$obj->$name                   # methodname : scalar                  (perly.y:1322-1324)
$obj->$name(@args)            # dynamic method name from a string
$obj->$coderef(@args)         # same rule; a coderef in $name bypasses lookup
$obj->&name                   # term ARROW '&' subname              (perly.y:1301)
$obj->&name(@args)            #                                      (perly.y:1300)
$obj->SUPER::method(@args)    # methodname is the bareword "SUPER::method"
$obj->Some::Class::method()   # fully qualified
$coderef->(@args)             # term ARROW '(' expr ')'             (perly.y:1385-1393)
```

`methodname` is exactly two things (`perly.y:1322-1324`):

```
methodname : METHCALL0 | scalar ;
```

`METHCALL0` is a bareword forced by the lexer at `toke.c:6297-6300`
(`force_word(s, METHCALL0, ALLOW_PACKAGE)`), which is why `$obj->method` does
not require the method to be declared and why `SUPER::` and `Foo::Bar::` work
without special grammar.

### 4.5.2 Chaining

`subscripted` is left-recursive over itself (`perly.y:1372-1381`), and
`listop`'s method rules take an arbitrary `term` as invocant. Combined with
`ARROW` at level 29 and `[`/`{`/`(` at 31-32, chains parse greedily left to
right:

```perl
$x->{a}[0]{b}(1)->method->@*
# ((((($x->{a})->[0])->{b})->(1))->method)->@*
```

**The arrow is optional between consecutive subscripts.** `$x->{a}{b}` and
`$x->{a}->{b}` are the same tree, because `subscripted '{' expr ';' '}'` is a
production in its own right (`perly.y:1380`). It is *not* optional before `(`
in a method call, nor before the first subscript after a scalar variable
(`$x[0]` is `@x`'s element, `$x->[0]` is a deref).

```go
func (p *Parser) parsePostfix(left Node) Node {
	for {
		switch {
		case p.at("->"):
			p.next()
			left = p.parseArrowTail(left) // method, subscript, call, or ->@* etc.
		case p.at("[") && subscriptable(left):
			left = &Index{Base: left, Idx: p.parseBracketed("[", "]")}
		case p.at("{") && subscriptable(left):
			left = &Key{Base: left, Key: p.parseBracketed("{", "}")}
		case p.at("(") && callable(left):
			left = &Call{Fn: left, Args: p.parseBracketed("(", ")")}
		default:
			return left
		}
	}
}
```

`subscriptable(left)` is true after a subscript, an arrow-deref, or a scalar
variable. It is false after a bareword or a completed call without an arrow —
`foo{a}` is not a subscript.

### 4.5.3 Indirect object syntax, and why it is a disaster

```
listop : METHCALL0 indirob optlistexpr                  # perly.y:1307  "new Class @args"
       | METHCALL  indirob '(' optexpr ')'              # perly.y:1313  "method $obj (@args)"
```

This is the rule that permits:

```perl
my $obj = new Foo::Bar(1, 2);       # == Foo::Bar->new(1, 2)
print STDERR "message\n";           # the filehandle slot in list operators
print {$fh} "message\n";            # block form of the same
```

It is a disaster for four reasons, and a Go implementation should reproduce it
only for `print`/`printf`/`say` filehandle slots and behind a diagnostic
elsewhere:

1. **It is not decidable without a symbol table.** `new Foo(1)` is a method
   call if `Foo` is a package and a function call `new(Foo(1))` if `new` and
   `Foo` are subs. `toke.c:10002` calls the resolution point "disambiguate
   between method and sub call" and consults the symbol table at parse time.
2. **`indirob` accepts a block** (`perly.y:1890`), so `print {$fh} @args`
   parses; but a bare `{` after a list operator is also an anon hash, so
   `print {1,2}` is ambiguous with the same heuristic as §4.4.4.
3. **It changes the arity silently.** `new Foo 1, 2, 3` swallows to the end of
   the statement, `new Foo(1), 2` does not (§4.8).
4. **It is gated by the `indirect` feature, not by `use strict`.** Measured on
   5.42: `use strict` does *not* disable it — `my $o = new Foo;` still parses
   and runs under `use strict`. What disables it is turning off the `indirect`
   feature, which `use v5.36` and later bundles do; `S_intuit_method` then
   returns 0 immediately (`toke.c:5071`) and `new Foo` becomes a syntax error.
   See 00-findings §0.2. So the pragma that matters is the version declaration
   or an explicit `no feature 'indirect'`, and a lexer must track that bit.

Recommendation: parse it, produce an `IndirectMethod` node, and mark it
`Ambiguous: true`. Let PSC decline to infer through it rather than guess.

### 4.5.4 The fat comma `=>`

`=>` is a comma at level 8 that autoquotes a bareword to its immediate left.
The autoquoting happens in the lexer, not the grammar — by the time `perly.y`
sees it, the left operand is already a `THING`. PerlOnJava does it in the
parser (`ParseInfix.java:252-263`) and additionally strips a trailing `::`,
matching perl's `Foo::Bar:: => 1` behaviour.

```perl
a => 1              # ('a', 1)   -- bareword autoquoted
-a => 1             # ('-a', 1)  -- leading minus is kept and quoted
Foo::Bar:: => 1     # ('Foo::Bar', 1)  -- trailing :: stripped
$a => 1             # ($a, 1)    -- not a bareword, no quoting
a() => 1            # (a(), 1)   -- a call, not a bareword
```

`=>` also suppresses the "Bareword not allowed under strict subs" error for its
left operand. The AST should record `Comma{Fat: true}` so a formatter can
preserve the spelling and PSC can treat the LHS as a string literal.

---

## 4.6 `local`, `my`, and lvalue positions

`local` is level 19 and takes an arbitrary term (`perly.y:1574`), which is
looser than most people expect:

```perl
local $x = 1;                # local($x) = 1  -- ASSIGNOP (90) < UNIOP (190)?
                             # No: `local term` reduces at UNIOP, then '=' at 90
                             # applies to the whole `local $x`. Correct.
local $x, $y;                # local($x), $y  -- the comma is looser than local
local ($x, $y);              # localises both -- the parens make one term
```

The rule `KW_LOCAL term %prec UNIOP` means `local` grabs one *term*, and a
parenthesised list is one term. Without parens, only the first variable is
localised. Same for `my`, but `my`'s operand grammar (`myterm`,
`perly.y:1728-1741`) refuses anything but a paren-list or a plain variable, so
`my $x, $y` produces a warning rather than silently declaring only `$x`.

---

## 4.7 `do`, `eval`, and block-versus-expression operators

Four operators take either a BLOCK or an EXPR, and the two spellings parse at
different levels and mean different things:

| Form | `perly.y` | Level | Meaning |
|---|---|---|---|
| `do BLOCK` | `1560` `%prec '('` | 31 (term) | Run the block, yield last value. Not a loop. |
| `do EXPR` | `1558` `%prec UNIOP` | 19 | `dofile()` — read and execute a file. |
| `eval BLOCK` | `1671` `UNIOP block` | 19 | Compile-time-known block, runtime exception trap. |
| `eval EXPR` | `1673` `UNIOP term` | 19 | String eval — a full compiler invocation at runtime. |
| `sub BLOCK` | `1538` `%prec '('` | 31 (term) | Anonymous sub. |
| `sort/map/grep BLOCK LIST` | §4.9 | 7 (LSTOP) | Block is the first argument. |

```perl
my $v = do { 1; 2; 3 };      # 3        -- do BLOCK is a term
do 'config.pl';               # dofile   -- do EXPR is a named unary
do $file;                     # dofile! Not "run the coderef in $file".
                              # Use $file->() or &$file for that.
eval { risky() };  if ($@) {} # block eval -- statically analysable
eval "1 + $user_input";       # string eval -- NOT statically analysable
eval $code;                   # same; PSC must treat the result as Any
```

The `do BLOCK`/`do EXPR` split at levels 31 versus 19 is the reason
`do { ... } while $cond;` is a statement modifier on a term and executes at
least once, while `do $file while $cond` is a named unary in a while loop.

`eval BLOCK` versus `eval EXPR` is decided purely by whether the next token is
`{` (`perly.y:1671` vs `1673`) — with the same anon-hash ambiguity, though in
practice `eval {` is always read as a block because `UNIOP block` is tried
first.

---

## 4.8 Function and list-operator calls

### 4.8.1 The paren cliff

This is the most-cited Perl gotcha and it lives in the *lexer*, at
`toke.c:S_lop` (around `toke.c:2160`):

```c
static int
S_lop(pTHX_ enum yytokentype t, I32 f, U8 x, char *s)
{
    ...
    if (*s == '(')
        return REPORT(FUNC);        /* '(' immediately follows: bounded call */
    s = skipspace(s);
    if (*s == '(')
        return REPORT(FUNC);        /* '(' after whitespace: still bounded */
    else {
        return REPORT(t);           /* no paren: LSTOP, greedy to the right */
    }
}
```

`FUNC` is not in the precedence table at all — it is consumed by
`FUNC '(' optexpr ')'` (`perly.y:1327`), a bounded production. `LSTOP` is level
7, which is looser than everything except `not`, `and`, `or`, `xor`. Hence:

```perl
print (1+2)*3;      # print(3) * 3   -- verified: B::Deparse gives (print(3) * 3)
print((1+2)*3);     # print(9)
print +(1+2)*3;     # print(9)  -- the classic unary-plus workaround
```

The same rule applies to *named unaries* via `UNI3` (`toke.c:285-296`), which
returns `FUNC1` if a `(` follows and `UNIOP` otherwise:

```perl
length ($x) + 1     # length($x) + 1  -- FUNC1, bounded
length $x + 1       # length($x + 1)  -- UNIOP, level 19
```

A Go implementation must reproduce this at the point where it decides how to
parse a bareword's arguments:

```go
// callArity decides how far a named operator's arguments extend.
// This mirrors toke.c:S_lop and toke.c:UNI3 -- the '(' test happens BEFORE
// any precedence consideration and is why `print (1+2)*3` prints 3.
func (p *Parser) parseNamedOp(name string, kind OpKind) Node {
	if p.peekAfterSpace().Text == "(" {
		// Bounded: FUNC / FUNC1. Arguments are exactly what is inside.
		return &Call{Name: name, Args: p.parseParenList(), Parenthesized: true}
	}
	switch kind {
	case NamedUnary: // UNIOP, level 19
		if p.startsTerm() {
			return &Call{Name: name, Args: []Node{p.parseExpr(bpNamedUnary)}}
		}
		return &Call{Name: name, Args: nil} // implicit $_
	case ListOp: // LSTOP, level 7 -- swallows commas and everything above
		return &Call{Name: name, Args: p.parseCommaList(bpListOp)}
	}
}
```

### 4.8.2 How far a list operator swallows

`LSTOP optlistexpr` (`perly.y:1325`) with `LSTOP` at level 7 means a list
operator without parens consumes everything up to — but not including — the
next `and`/`or`/`xor`/`not`, or a statement terminator. Everything tighter,
including the comma and assignment, becomes part of its argument list.

```perl
print $a, $b or die;         # print($a, $b) or die     -- `or` is looser
print $a, $b || die;         # print($a, ($b || die))   -- `||` is tighter
return $x if $cond;          # `if` modifier terminates the list
my @s = sort @a, @b;         # sort(@a, @b) -- one flat list, both slurped
join ':', map { uc } @list;  # join(':', map({uc} @list)) -- map is itself
                             # an LSTOP inside join's list. Both greedy;
                             # the inner one ends at join's own end.
```

`toke.c:S_lop` also lowers `PL_lex_fakeeof` to `LEX_FAKEEOF_LOWLOGIC` when it
emits `LSTOP`, which is the mechanism that stops the swallow at `and`/`or`.

### 4.8.3 `&foo` and friends

```
term : amper                        { ENTERSUB }         # perly.y:1636  &foo;
     | amper '(' ')'                { ENTERSUB STACKED } # perly.y:1638  &foo()
     | amper '(' expr ')'           { ENTERSUB STACKED } # perly.y:1640  &foo(@a)
     | NOAMP subname optlistexpr    { ENTERSUB STACKED } # perly.y:1645  foo @a
amper: PERLY_AMPERSAND indirob      { newCVREF }         # perly.y:1841
```

Four distinct meanings, and the differences are semantic, not just syntactic:

```perl
foo(@args)    # normal call, prototype applied
foo @args     # NOAMP: call without parens, prototype applied,
              #        only if foo is already declared
&foo(@args)   # call, prototype IGNORED
&foo          # call, passes the CALLER'S @_ through unchanged
\&foo         # a code reference, no call
&$ref(@args)  # call through a reference
&{$ref}(@args)# same
$ref->(@args) # same, preferred spelling
```

`&foo` with no parens passing `@_` through is the one worth flagging in an
LSP — it is almost always a mistake in modern code, and PSC cannot infer the
argument types because they come from the enclosing sub.

`NOAMP` (`perly.y:1645`) is emitted by the lexer only when the bareword is a
known sub. This is the fundamental undecidability: whether `foo @args` parses
as a call or is a syntax error depends on whether a `BEGIN` block declared
`foo`. Chapter 1 covers the consequence; the expression parser's job is to
produce a `Call` node with `Resolved: false` and let a later pass decide.

---

## 4.9 `sort` / `map` / `grep`

These three are `LSTOP`s (`toke.c:8585` `LOP(OP_GREPSTART, XREF)`,
`toke.c:8756` `LOP(OP_MAPSTART, XREF)`, `toke.c:9043` `LOP(OP_SORT, XREF)`) but
their first argument has three mutually ambiguous forms.

### 4.9.1 The three shapes

```perl
sort { $a <=> $b } @x     # BLOCK  LIST      -- comparator block
sort $coderef @x          # SCALAR LIST      -- comparator in a scalar. NO COMMA.
sort sortsub @x           # BAREWORD LIST    -- named comparator. NO COMMA.
sort keys %h              # LIST             -- `keys %h` is the list; no
                          #                     comparator; default string sort
sort @x, @y               # LIST             -- flat list of both
```

The critical asymmetry: **when the first argument is a comparator, there is no
comma after it.** When it is part of the list, there is. So `sort $x, @y` sorts
`($x, @y)` while `sort $x @y` sorts `@y` using `$x` as comparator.

`toke.c:9038-9043` handles this by calling `force_word(s, BAREWORD,
CHECK_KEYWORD | ALLOW_PACKAGE)` before emitting `LOP(OP_SORT, XREF)` — it
speculatively forces the next bareword, but `CHECK_KEYWORD` means `keys`,
`values`, `map` etc. are *not* forced, so `sort keys %h` correctly parses as a
list. Verified: `B::Deparse` renders it `sort(keys %h)`.

`map` and `grep` have a second shape:

```perl
map  { $_ * 2 } @x        # BLOCK LIST
map  BLOCK LIST           #   -- no comma after the block
grep { /x/ } @lines       # BLOCK LIST
grep /x/, @lines          # EXPR, LIST  -- comma REQUIRED after the expression
map  $_ * 2, @x           # EXPR, LIST  -- comma REQUIRED
```

So `map`/`grep` require a comma after an expression first argument and forbid
one after a block first argument. `sort` forbids the comma after either a block
or a comparator, and requires it between plain list elements.

### 4.9.2 The `{` decision

The whole ambiguity reduces to: after `sort`/`map`/`grep`, is `{` a block or an
anonymous hash? `toke.c` sets `PL_expect = XREF` (`toke.c:8756`) and
`yyl_leftcurly` falls through to the lookahead heuristic of chapter 3 §3.3.
The practical rule for a Go implementation:

```go
// After sort/map/grep, a '{' is a BLOCK unless the lookahead says otherwise.
// Mirrors toke.c's XREF handling plus the anon-hash heuristic.
func (p *Parser) parseMapGrepSortFirstArg(name string) (Node, bool /*isBlock*/) {
	if p.at("{") {
		if p.looksLikeAnonHash() { // ch3 §3.3, verbatim
			return p.parseTerm(), false
		}
		return p.parseBlock(), true
	}
	if name == "sort" {
		// A lone scalar or bareword NOT followed by a comma is the comparator.
		if save := p.mark(); p.tryParseComparatorNoComma() {
			return p.comparator(), true
		} else {
			p.reset(save)
		}
	}
	return nil, false // fall through to plain list parsing
}
```

`looksLikeAnonHash` must be the same predicate used in §4.4.4 and in chapter
5 §5.12; keep one implementation, specified by chapter 3 §3.3. Record the guess on the node — `MapGrep{BlockGuessed: true}` —
so the LSP can offer the `+{` / `{;` disambiguation as a quick fix.

### 4.9.3 Context inside the block

`map`'s block is in **list** context; `grep`'s and `sort`'s are in **boolean**
(scalar) and **scalar** context respectively. This matters to PSC:

```perl
my @pairs = map { ($_, 1) } @keys;   # block in list ctx: 2 values per input
my @count = map { $_, 1 } @keys;     # same
my @n     = map { scalar @$_ } @rows;# block forced to scalar by `scalar`
my @hits  = grep { @$_ } @rows;      # block in boolean ctx: @$_ is a count
```

---

## 4.10 `x`: string repetition versus list repetition

`x` is a `MULOP` (`toke.c:9216`), level 23, left associative. But its *meaning*
depends on the syntactic shape of its left operand, decided at
`perly.y:1425-1428`:

```
| term MULOP term { if ($MULOP != OP_REPEAT) scalar($lhs);
                    $$ = newBINOP($MULOP, 0, $lhs, scalar($rhs)); }
```

For `OP_REPEAT` the left operand is *not* forced to scalar context. Whether it
becomes a list repeat is then decided by whether the left operand was
parenthesised (`sawparens`, `perly.y:1577`) or is a `qw()` list:

```perl
my $s = 'ab' x 3;         # 'ababab'          -- scalar repeat
my @a = ('a') x 3;        # ('a','a','a')     -- LIST repeat, parens matter
my @a = 'a' x 3;          # ('aaa')           -- scalar repeat, then assigned
my @a = (1, 2) x 3;       # (1,2,1,2,1,2)     -- verified via B::Deparse
my @a = qw(a b) x 2;      # ('a','b','a','b') -- QWLIST counts as parenthesised
my @m = ([]) x 3;         # THREE REFS TO THE SAME ARRAY -- repeat copies the
                          # value, and the value is one reference.
```

The AST must therefore carry the parenthesisation:

```go
type Repeat struct {
	Left      Node
	Count     Node
	ListRepeat bool // true iff Left was syntactically parenthesised or a qw()
}
```

Do not normalise away `Paren` nodes before this decision is made. This is a
common source of bugs in Perl parsers that eagerly unwrap parentheses.

Measured 2026-09-17 at 3b9422eb: `internal/parse/term.go:284-293` does
exactly this — `finishList` returns the single item with the parens' span,
and `($x);` parses to `[statement [term "$x"]]` with no `Paren` node. It
round-trips, because round-trip cannot see structure (chapter 7 §7.2(c)).
No issue tracked it at the time of measurement; chapter 6 §6.1.8 lists it.

---

## 4.11 Quote-like operators as expressions

`toke.c` reduces all quote-likes to `PMFUNC` (match/subst/trans) or `THING`
(strings) or `QWLIST`. `perly.y:1692-1704` shows `PMFUNC` re-entering the
lexer for the pattern and replacement:

```
| PMFUNC { ... start_subparse if PMf_HAS_CV ... }
    SUBLEXSTART listexpr optrepl SUBLEXEND
        { pmruntime($PMFUNC, $listexpr, $optrepl, 1, ...); }
```

| Operator | Token | Scalar-context value | List-context value |
|---|---|---|---|
| `m/PAT/` | `PMFUNC` | boolean match | capture groups, or `(1)` if none |
| `m/PAT/g` | `PMFUNC` | iterated boolean (advances `pos`) | all matches |
| `s/PAT/REP/` | `PMFUNC` | count of substitutions | same |
| `s/PAT/REP/r` | `PMFUNC` | the modified copy | same |
| `tr/A/B/` | `PMFUNC` | count of chars translated | same |
| `tr/A/B/r` | `PMFUNC` | the modified copy | same |
| `qr/PAT/` | `PMFUNC` | a `Regexp` object | same |
| `qw(a b)` | `QWLIST` | last element | the list |
| `q()` `qq()` | `THING` | the string | the string |
| `qx()` `` `` `` | — | all output, one string | output split on `$/` |

Two parse-level facts:

1. **`m//` and `s///` bind via `=~`** (level 24) or default to `$_`. A bare
   `/foo/` in expression position is `$_ =~ m/foo/`. Whether `/` starts a match
   or is division is `PL_expect` state, not precedence.
2. **`s///e` and `s///ee` make the replacement a code expression**, parsed as
   Perl, not as a string. The `PMf_HAS_CV` branch above is exactly this. A Go
   parser must recursively parse the replacement when `/e` is present, and must
   refuse (or mark `Ambiguous`) on `/ee`, which is a nested string eval.

The AST should keep the pattern as source text plus a parsed interpolation
tree, not as a compiled regex: the LSP needs spans inside the pattern, and PSC
needs to know which captures exist.

```go
type Match struct {
	Op      string // "m", "s", "tr", "qr"
	Pattern []Node // interpolation parts: Str, Var, Expr
	Replace []Node // s/// only; parsed as Perl code when /e
	Flags   string
	Target  Node   // nil means $_
	Negated bool   // true for !~
}
```

---

## 4.12 Context propagation

This section exists because PSC consumes it. Perl has three contexts — **list**,
**scalar**, and **void** — plus scalar sub-flavours (boolean, numeric, string)
that affect overloading but not parsing. Context flows *down* the tree from the
consumer to the producer, which is the opposite direction from type inference,
so PSC needs it recorded explicitly on each node.

### 4.12.1 What imposes which context

| Construct | Context imposed on children | Source |
|---|---|---|
| `my @a = EXPR` / `my %h = EXPR` | **list** on EXPR | assignment to aggregate |
| `my $x = EXPR` | **scalar** on EXPR | assignment to scalar |
| `my ($x) = EXPR` | **list** on EXPR | parens make the LHS a list |
| `($a, $b) = EXPR` | **list** | |
| `return EXPR` | **caller's context** (`wantarray`) | undecidable statically |
| `scalar EXPR` | **scalar** | `perly.y` `UNIOP` with `$` prototype |
| `EXPR ? A : B` | cond→**boolean**; A,B→**inherited** | `perly.y:1568` |
| `if/while/unless (EXPR)` | **boolean** (a scalar flavour) | |
| `EXPR1 && EXPR2` | E1→**boolean**, E2→**inherited** | `perly.y:1442` `newLOGOP` |
| `EXPR1 \|\| EXPR2`, `//` | same as `&&` | `perly.y:1446`, `1450` |
| `not EXPR`, `!EXPR` | **boolean** | `perly.y:1501`, `1669` |
| arithmetic `+ - * / % **` | **scalar** on both, numeric | `scalar($lhs)` in the actions |
| `.` concat | **scalar** on both, string | `perly.y:1431` calls `scalar()` |
| `x` repeat | LHS: **inherited**; RHS: **scalar** | `perly.y:1425` skips `scalar($lhs)` |
| comparisons | **scalar** on both | `perly.y:1465` etc. call `scalar()` |
| `,` (comma) | **inherited** by every element | list flattening |
| `[ EXPR ]`, `{ EXPR }` | **list** on EXPR | `newANONLIST`/`newANONHASH` |
| `\ EXPR` | **inherited**, no flattening | `OP_REFGEN` |
| subscript `$a[EXPR]`, `$h{EXPR}` | **scalar** on EXPR | `scalar($expr)` at `perly.y:1364` |
| slice `@a[EXPR]`, `@h{EXPR}` | **list** on EXPR | `list($expr)` at `perly.y:1597` |
| function args (no prototype) | **list** | |
| function args (prototype `$`) | **scalar** per `$` | prototypes, Chapter 5 |
| named unary operand | **scalar** | `newUNOP` after `scalar()` |
| `print`, `push`, `join` args | **list** | `op_convert_list` |
| `sort`/`map`/`grep` list arg | **list** | |
| `map` block | **list** | |
| `grep` block | **boolean** | |
| `sort` block | **scalar** (numeric or string) | |
| statement in void position | **void** | last-statement-of-block excepted |
| last statement of a sub/`do` block | **caller's context** | |
| `wantarray` | — | *reports* the current context at runtime |

### 4.12.2 The rules an implementation must encode

1. **Context is a property of the edge, not the node.** The same `@a` is a
   count in scalar context and a list of elements in list context. Store the
   imposed context on the child link, or as a field the parent sets.
2. **Assignment decides context by the *syntactic shape* of its LHS**, before
   any type information exists. `my ($x) = f()` and `my $x = f()` differ only
   in parentheses and call `f` in different contexts. This is why parentheses
   must survive into the AST (see also §4.10).
3. **`return` and the last statement of a sub inherit the caller's context**,
   which is not statically known. PSC should model this as a context variable,
   not resolve it to one branch.
4. **Void context propagates only to statement position.** Every expression in
   a statement-list except the last is in void context; the last inherits.
5. **List context flattens; scalar context does not.** `my @a = (@b, @c)` is
   one flat list. `my $n = (@b, @c)` is `$c` in scalar context (comma operator),
   not a count. This trips up both humans and parsers.

```go
type Context uint8

const (
	CtxVoid Context = iota
	CtxScalar
	CtxList
	CtxBoolean // a scalar flavour; affects overloading, not arity
	CtxInherit // resolved from the parent; used for return, ?:, ||-rhs
)

// Node carries the context its parent imposes on it. Set during a downward
// pass after parsing; the parser itself does not need to know contexts, but
// it must preserve the syntax (parens, sigils) that determine them.
type nodeBase struct {
	Ctx  Context
	Span Span
}
```

---

## 4.13 Where the implementations disagree with `perly.y`

Each of these was checked against `perly.y`'s declaration block and, where
observable, against `perl -MO=Deparse` on perl 5.42.0.

### 4.13.1 PerlOnJava: `^^` has its own precedence level

`ParserTables.java:337` places `^^` at level 9 alongside `||` and `//`, which
happens to be right. But `ParserTables.java:342` also lists `^^=` at level 6
with the other assignment ops, and `INFIX_OP` (`ParserTables.java:15`) lists
`^^` separately. `toke.c:6436-6441` shows `^^` is literally emitted as the
`OROR` token with `pl_yylval.ival = OP_XOR` — same token class, same level. The
net behaviour matches; the table just describes it as a coincidence rather than
an identity. Low severity, but it means a future edit to the `||` level would
not automatically move `^^`.

### 4.13.2 PerlOnJava: file test operators get their own level

`ParserTables.java:331` — `addOperatorsToMap(16, "-d")` — puts file tests at a
level between `isa` (15) and shift (17). `toke.c:6255-6261` emits `FTST(ftst)`,
which is `#define FTST(f) return (pl_yylval.ival=f, ..., REPORT((int)UNIOP))`
(`toke.c:261`). **File tests are `UNIOP`, level 19 — the same level as every
other named unary.** They belong above shift, add and mul, not below.

Consequence, confirmed against perl:

```perl
-e $file . '.bak'
# perly.y:  -e ($file . '.bak')   -- ADDOP (22) is tighter than UNIOP (19)
# perl:     B::Deparse of `-e "f" . "x"` gives `-e 'fx'` -- confirms perly.y
# PerlOnJava's table (level 16 < 18 for '.') would give (-e $file) . '.bak'
```

Also, only `-d` is in the map. Every other file test (`-e`, `-f`, `-r`, `-s`,
`-z`, `-M`, …) has no entry, so `precedenceMap.get()` returns null for them and
they fall through to a different path. This is a real bug.

### 4.13.3 PerlOnJava: `\` at the same level as `!` and `~`

`ParserTables.java:336` — `addOperatorsToMap(21, "!", "~", "~.", "\\")`. That
matches `perly.y:175` (`%right ! ~ UMINUS REFGEN`) exactly. Correct — noted
here only because `perlop` lists `\` on a separate line from `!` and `~`, and
several parsers follow `perlop` rather than the grammar. `perlop`'s grouping is
presentational; `perly.y` puts them on one `%right` line, which means
`\!$x` and `!\$x` both parse without needing a level between them.

### 4.13.4 PerlOnJava: `print` as a precedence level

`ParserTables.java:329` — `addOperatorsToMap(4, "print")` — invents a level for
`print` between `not` (3) and `,` (5). In `perly.y` this is the `LSTOP` level
(`perly.y:157`), which is shared by *every* list operator: `print`, `push`,
`join`, `die`, `warn`, `sort`, `map`, `grep`, `return`, `open`, and any
user-defined sub with a list prototype. Naming the level after one member is
harmless for `print` itself but means the other list operators are not in the
table at all and must be special-cased elsewhere. A Go implementation should
name the level `LSTOP` and route every list operator through it.

### 4.13.5 perl-lsp: `isa` is treated as chainable and left-associative

`precedence.rs:780-806` handles `isa` inside `parse_relational_with`'s `while`
loop, in the same iteration structure as `<`, `>`, `<=`, `>=`. That makes
`$a isa Foo isa Bar` parse as `(($a isa Foo) isa Bar)`.

`perly.y:1467-1476` classifies `isa` as `NCRELOP` — non-chaining — and has an
explicit error production:

```
termrelop : term NCRELOP term  { newBINOP(...) }
          | termrelop NCRELOP  { yyerror("syntax error"); YYERROR; }
          | termrelop CHRELOP  { yyerror("syntax error"); YYERROR; }
```

`$a isa Foo isa Bar` and `$a < $b isa Foo` are both syntax errors in real perl.
perl-lsp accepts both. For an LSP this is arguably a benign over-acceptance
(better to keep parsing than to bail), but the node it produces is wrong and
PSC would infer through a tree that perl would have rejected. Recommendation:
parse it, but emit a diagnostic.

### 4.13.6 perl-lsp: no chained comparison support

`precedence.rs:752-810` builds left-nested `Binary` nodes for `<`, `>`, `<=`,
`>=`. Perl 5.32+ chains these into a single n-ary comparison with
single-evaluation semantics (`perly.y:1466-1478`, `cmpchain_start` /
`cmpchain_extend` / `cmpchain_finish`). `$lo <= $x() <= $hi` calls `$x()`
**once** in real perl and **twice** under a left-nested reading — and the
left-nested reading is also numerically wrong, since `($lo <= $x) <= $hi`
compares a boolean against `$hi`.

This is the most consequential of the disagreements, because the tree shape
differs, not just the acceptance. Verified: `perl -MO=Deparse,-p -e 'my $q =
$a < $b < $c'` on 5.42 produces a chain, not nested binaries.

### 4.13.7 perl-lsp: recursive-descent cascade instead of a table

`precedence.rs` implements one function per level — `parse_or`, `parse_and`,
`parse_bitwise_or`, `parse_bitwise_xor`, `parse_range`, `parse_bitwise_and`,
`parse_equality`, `parse_relational`, `parse_shift`, `parse_additive`,
`parse_multiplicative`, `parse_power` — each calling the next. This is correct
where the cascade order matches `perly.y`, and it does. Two observations:

- `parse_bitwise_xor` is a *separate level from* `parse_bitwise_or`
  (`precedence.rs:442`, `543`). In `perly.y:164`, `|` and `^` are both
  `BITOROP` on one `%left` line — **the same level**. The cascade gives `^`
  tighter binding than `|`, so `$a | $b ^ $c` parses as `$a | ($b ^ $c)` in
  perl-lsp and as `($a | $b) ^ $c` in perl. This is a genuine
  associativity/grouping bug for mixed `|`/`^` expressions.
- `parse_range` (`precedence.rs:448`, `567`) is called by `parse_bitwise_and`
  and itself calls `parse_equality`, so in the cascade `..` binds *tighter*
  than every bitwise operator and *looser* than equality. In `perly.y:161`,
  `DOTDOT` is **looser than `||`**, which is in turn looser than all the
  bitwise operators — the opposite end of the range. So `$a | $b .. $c` should
  parse as `($a | $b) .. $c`; perl-lsp's cascade groups it the other way. Also
  `DOTDOT` is `%nonassoc` in `perly.y`; `parse_range_with`'s `while` loop makes
  it left associative, so perl-lsp silently accepts `1 .. 2 .. 3`, which real
  perl rejects with `syntax error ... near "2 .."` (verified on 5.42).

For a Go implementation, prefer the table (§4.2) over a cascade: 32 levels is
32 functions to keep in the right order, and the two bugs above are exactly the
kind that a cascade invites and a table prevents.

### 4.13.8 Summary

| Finding | Where | Severity |
|---|---|---|
| File tests given their own level below `.`; only `-d` in the map | `ParserTables.java:331` | **high** — changes grouping of `-e $f . '.x'` |
| `isa` chainable and left-assoc | `precedence.rs:780-806` | medium — over-accepts, wrong tree |
| No chained comparisons | `precedence.rs:752-810` | **high** — wrong tree and wrong evaluation count |
| `^` split from `\|` into its own level | `precedence.rs:442`, `543` | medium — regroups `$a \| $b ^ $c` |
| `..` placed among the bitwise levels, and made left-assoc | `precedence.rs:448`, `567` | medium — wrong grouping, over-accepts `1..2..3` |
| `print` named as a level instead of `LSTOP` | `ParserTables.java:329` | low — other list ops absent from the table |
| `^^` described as separate from `\|\|` rather than identical | `ParserTables.java:337` | low — behaviour currently correct |

---

## 4.14 The lowered vocabulary

This is the vocabulary inference reads, not the vocabulary the parser emits.
Chapter 6 §6.1.6 places a lowering pass between `internal/parse`'s CST and
this list; the CST stays the only tree the LSP's structural features read.
The list was written before the parser existed. This revision reconciles it,
at 3c997a6b, against four sources, and every entry below names which of
them it rests on:

- **P** — what `internal/parse` ships: 23 kinds plus `SourceFile`
  (`parse.go:146-201`), and a `Node` carrying `Kind`, `Start`, `End`,
  `Children`, `Text` and exactly one flag, `Resolved` (`parse.go:213-234`).
  Measured on T1 (PerlOnJava `unit/*.t` at top level, 986 files): 408 clean,
  4,944 `Unknown`, 0 round-trip failures. Tree shapes quoted below were
  dumped from `parse.Parse` at that commit with a throwaway test, since
  deleted; each is one line to reproduce.
- **C** — what `internal/infer` consumes today, which is still the
  tree-sitter tree via `internal/parser/`: 57 grammar names. Where a rule
  re-derives structure from source *text*, that is evidence the lowered node
  must carry the fact as a field, because the text path is where the
  measured defects live.
- **perl** — perl 5.42.0, run.
- **SoN** — B::SoN at `~/dev/perl5-son`, read directly: 102 classes under
  `lib/SoN/IR/Node/`, of which 5 are abstract bases with no `operation()`
  and no construction site (`Access`, `Aggregate`, `BinOp`, `UnaryOp`,
  `Regex`), so 97 concrete kinds. Chapter 6 §6.1.6 said 101, from a report
  rather than a count.

Each kind carries a status:

| Status | Meaning |
|---|---|
| SHIPPED | the CST has the node; the lowering renames or passes it through |
| LOWERING | the CST has no such node but carries every fact needed to build one |
| PARSER | the parser discarded the fact, or does not parse the construct; the lowering cannot build the node until the parser changes |
| SPECULATIVE | no consumer reads it, no parser support exists, and no perl or SoN fact requires it |

The guiding principle is unchanged — **preserve syntax that determines
semantics** and record every place the parser had to guess — but the list
is no longer justified as a smaller tree. Measured over T1: 272,455 CST
nodes, of which 43,250 (15.9%) are trivia; the lowering benchmark at the
same commit lowered them to 229,205 (ratio 0.841), and the reduction is the
trivia. Its justification is §6.1.6's: one shape per
operation, so a rule matches one shape.

```go
// Every node carries the context its parent imposes (§4.12) and an identity.
//
// Identity is a stable id, never a byte offset. internal/infer/infer.go:184
// keys its annotation map on StartByte, and a parent written after its
// children overwrites the child at the same offset: measured on
// `my $x = 5; my @y = ($x); my $z = $y[0];`, byte 0 carries four nodes and
// the `number` inside `(1..5)` reads back as its parent's type. Spans live
// in the side table of §6.1.6 item 3, keyed by this id; the table must also
// preserve total source ORDER, because nearestPatternBefore
// (infer.go:1724-1751) resolves $1 by "the last match before this point".
type Node interface{ Base() *nodeBase }

// Hole is one Unknown, opaque. See §4.14.2.
type Hole struct{ nodeBase }

// --- Literals and names ---
type NumLit    struct{ nodeBase; Text string; IsInt bool }
type StrLit    struct{ nodeBase; Value string; Interpolated bool; Backtick bool }
                                     // Backtick: `cmd` / qx// -- the value is a command
type Interp    struct{ nodeBase; Parts []Node } // "a$b c" -> [StrLit, Var, StrLit]
type QwList    struct{ nodeBase; Words []string }
type Bareword  struct{ nodeBase; Name string; Quoted bool } // Quoted: from => or {k}

// --- Variables and dereference ---
type Var    struct{ nodeBase; Sigil string; Name string } // $x @x %x *x &x $#x
                                     // (Sigil, Name) is the identity pair: $_ and @_
                                     // are different variables. See rule 7.
type Deref  struct{ nodeBase; Sigil string; Expr Node; Postfix bool }
                                     // ${$x} / $$x / $x->@*   Postfix marks ->@*
type Index  struct{ nodeBase; Base Node; Idx Node; Arrow bool }  // $x[0] $x->[0]
type Key    struct{ nodeBase; Base Node; Key Node; Arrow bool }  // $h{k} $x->{k}
                                     // Arrow selects WHICH variable: $h{k} reads %h,
                                     // $h->{k} reads $h. See rule 2.
type Slice  struct{ nodeBase; Base Node; Idx Node; Kind SliceKind }
                                     // @a[..] @h{..} %a[..] %h{..}

type SliceKind uint8
const (
	SliceArray SliceKind = iota // @a[...]
	SliceHash                    // @h{...}
	SliceKVArray                 // %a[...]
	SliceKVHash                  // %h{...}
	SliceList                    // (LIST)[...]
)

// --- Operators ---
type Unary   struct{ nodeBase; Op string; Operand Node; Postfix bool } // ! ~ - \ ++ --
type Binary  struct{ nodeBase; Op string; L, R Node }
type Repeat  struct{ nodeBase; Left, Count Node; ListRepeat bool }   // x -- see 4.10
type Assign  struct{ nodeBase; Op string; L, R Node; ListAssign bool }
                                    // ListAssign from the LHS's syntactic shape
type Ternary struct{ nodeBase; Cond, Then, Else Node }
type CmpChain struct {              // 5.32+ chained comparison -- see 4.3
	nodeBase
	Operands []Node
	Ops      []string
	Class    TokenClass // CHRELOP or CHEQOP
}
type Range   struct{ nodeBase; Lo, Hi Node; Exclusive bool } // .. vs ...
type List    struct{ nodeBase; Elems []Node; Fat []bool }    // Fat[i]: elem i
                                                              // was followed by =>
type Paren   struct{ nodeBase; Inner Node }  // KEEP THIS. 4.10 and 4.12 need it.
                                             // The CST erases it today; rule 1.

// --- Declarations ---
type Decl struct {
	nodeBase
	Kind  DeclKind // My, Our, State, Local, Field
	Vars  []Node   // Var or Paren{List{Var}}
	Attrs []string
	Init  Node     // nil if none
	RefAlias bool  // my \$x = ...
}

// --- Calls ---
type Call struct {
	nodeBase
	Name          string // "" if Fn is set
	Fn            Node   // $coderef->() or &{$x}()
	Handle        Node   // print $fh LIST / print {$fh} LIST -- the slot before
	                     // the list, which is NOT argument 1 (infer.go:1155-1165)
	Args          []Node
	Parenthesized bool // affects the paren cliff (4.8.1) and prototypes
	Ampersand     bool // &foo -- prototypes bypassed
	BareAmpersand bool // &foo with no parens -- inherits caller's @_
	Resolved      bool // false if the parser could not tell call from bareword
}

type MethodCall struct {
	nodeBase
	Invocant Node
	Name     string // "" if Dynamic is set
	Dynamic  Node   // $obj->$name / $obj->$coderef
	Args     []Node
	HasParens bool
	Indirect bool // "new Foo @args" -- see 4.5.3; no CST source yet (§4.14.6)
	Ambiguous bool
}

// --- Special forms ---
type MapGrepSort struct {
	nodeBase
	Op           string // "map" | "grep" | "sort"
	Block        Node   // block form
	FirstExpr    Node   // expression form (map/grep) or comparator (sort)
	List         []Node
	BlockGuessed bool // the '{' was disambiguated heuristically -- see 4.9.2
}
type DoBlock  struct{ nodeBase; Body Node }              // do { ... }
type DoFile   struct{ nodeBase; Path Node }              // do EXPR
type EvalBlock struct{ nodeBase; Body Node }             // statically analysable
type EvalStr  struct{ nodeBase; Code Node }              // PSC: result is Any
type AnonSub  struct{ nodeBase; Sig []Node; Body Node; Attrs []string }
type AnonList struct{ nodeBase; Elems []Node }           // [ ... ]  (SoN: ArrayLiteral)
type AnonHash struct{ nodeBase; Elems []Node }           // { ... }  (SoN: HashLiteral)
type Ref      struct{ nodeBase; Operand Node }           // \EXPR
type FileTest struct{ nodeBase; Op byte; Operand Node }  // -e -d -f ...; nil = $_
type Readline struct{ nodeBase; Handle Node; Magic bool } // <FH> <> <<>>
type Glob     struct{ nodeBase; Pattern Node }           // <*.c> / glob EXPR
type LoopEx   struct{ nodeBase; Op string; Label Node }  // last/next/redo/goto
type Wantarray struct{ nodeBase }                        // same name in SoN

type Match struct { // m// s/// tr/// qr// -- see 4.11
	nodeBase
	Op      string
	Pattern []Node // when Op is tr or y: a character SET, not a pattern
	Replace []Node // likewise; and Flags are c/d/s/r, not regex flags
	Flags   string
	Target  Node // nil means $_
	Negated bool // !~
}
```

### 4.14.1 Node-kind rules that are not obvious

1. **Never fold `Paren` away.** `(1,2) x 3` versus `1,2 x 3`, and
   `my ($x) = f()` versus `my $x = f()`, both turn on it. A pass that unwraps
   parens is a pass that introduces bugs.

   The parser is that pass today. `finishList` (`internal/parse/term.go:288`)
   returns the single item with the parens' *span* and no node: `($x);` is
   `[term "$x"]` spanning `[0,4)`, and `((($x)));` is one `term "$x"` spanning
   `[0,8)` — three paren levels, zero nodes. The information is not gone: for
   `('a') x 3` the left operand is `term "'a'"` spanning `[0,5)` while for
   `'ab' x 3` it spans `[0,4)`, so a lowering can test `src[n.Start]=='('`.
   That test is unsound on its own. Census over T1: 6,574 non-trivia nodes
   have a span starting with `(`, and 721 of them (664 `binary`, 43 `index`,
   7 `ternary`, 6 `cmp_chain`, 1 `postfix`) have their *first child* starting
   at the same offset — `(1,2)[0] x 3` is an `index` over a `list`, not a
   parenthesised operand, and perl agrees (`my @a = ((1,2)[0]) x 3` gives
   `1 1 1`). A guard — the first child starts later than the node, or the
   node is a leaf whose `Text` is shorter than its span — removes those 721.
   Whether anything else slips past the guard was not checked beyond one
   known case: `PrototypeNode` leaves (62 in T1) carry their parens in
   `Text` and must be excluded by kind. The smaller and exact fix is a
   `Paren bool` on `parse.Node`, set at the one site that erases the node;
   `Resolved` (`parse.go:234`) is the precedent for a kind-specific flag.
   Status: PARSER, with a LOWERING fallback that needs the guard and has not
   been proven sound. `qw(a b) x 2` is the other list-repeat spelling and
   its `term "qw(a b)"` starts with `q`, so it is a separate case keyed on
   the `qw` prefix.

2. **`Index`/`Key` carry `Arrow`, and `Arrow` names a variable, not a
   spelling.** The earlier text of this rule said `$x->[0]` and `$$x[0]`
   "produce the same runtime op but different source" and gave the LSP as the
   consumer. That undersells it. `$h{k}` reads `%h`; `$h->{k}` reads `$h`:

   ```perl
   my %h = (k => "hash"); my $h = { k => "ref" };
   print $h{k}, $h->{k};    # hashref   -- perl 5.42.0
   ```

   Since d2ebe02a routed `->` before a subscript opener into `parseSubscript`
   (`internal/parse/expr.go:83-88`), `$h{k};` and `$h->{k};` parse to
   byte-identical trees — `[index "{" [term "$h"] [call "k"]]` — differing
   only in the `index` span, and `subscript_test.go:34-37` asserts that they
   must. 553 of T1's 2,186 `index` nodes (25.3%) are the arrow form. The
   comment at `expr.go:73` promises the flag §4.14 asks for;
   `parse.Node` does not have it. Status: PARSER. The arrow bytes survive in
   the gap between the base child's `End` and the opener, but re-scanning
   source there is the leak §6.1.6 criticises. Add `Arrow bool` to
   `parse.Node`, set in the `->` branch, and turn the test into an assertion
   that the two forms differ in exactly that field.

   What inference actually branches on is one step removed: "is the base a
   reference being dereferenced". `inferElementType` (`infer.go:1270-1280`)
   reads the old grammar's `container_variable`-vs-`scalar` split for
   exactly this, and `$$x[0]` is a dereference with no arrow. So the lowering
   derives *deref-ness* from `Arrow || base is a Deref`, and keeps `Arrow`
   itself for the rename case. No rule in `internal/infer` reads an arrow
   property by that name today (grep over `internal/infer/*.go`: comments
   only), which ranks it below the sigil split for inference and above it
   for the LSP.

3. **`Call.Resolved` and `MethodCall.Ambiguous` are the honest-uncertainty
   channel.** Chapter 1 establishes that some Perl cannot be parsed without
   running it. Rather than guessing silently, mark the node and let PSC widen
   to `Any` at exactly those points.

   `Resolved` ships (`parse.go:229-234`) and the old grammar's
   `ambiguous_function_call_expression` was carrying the same bit. But
   `Call{Resolved:false}` is overloaded four ways in the CST, all measured:
   an unresolved call; a hash key (`$h{k}` → `[index "{" [term "$h"]
   [call "k"]]`); the left of a fat comma (`a => 1` → `[binary "=>"
   [call "a"] [term "1"]]`); a class name as invocant (`Foo->new` →
   `[binary "->" [call "Foo"] [call "new"]]`); and the stat-buffer bareword
   (`-r _` → `[call "-r" [call "_"]]`, where §4.15.1 asks for a `Bareword`).
   The lowering can separate these by position — second child of an
   `Index{"{"}`, left operand of `Binary "=>"`, left operand of `Binary "->"`,
   operand of a file test — and produce `Bareword{Quoted:true}` for the
   first two. Status: LOWERING, but `parseWordTerm`'s `default:` arm already
   knows the position at construction, and fixing it there is the smaller
   diff. `MethodCall.Ambiguous` and `Indirect` have no CST source at all:
   `new Foo(1)` is `Unknown`. Status: PARSER; whether to keep the fields is
   §4.14.6.

4. **`MapGrepSort.BlockGuessed` likewise.** The `{`-is-a-block heuristic is
   documented as a heuristic in perl's own source; record when it fired.

   Nothing to record yet: `map { $_ } @a;`, `grep { $_ } @a;` and
   `sort { $a <=> $b } @a;` are each one `Unknown`. The expression forms
   parse (`map $_+1, @a` → `[call "map" resolved=true [binary "," ...]]`).
   This is the brace classification of §4.9.2, owned by issue 01a0ac52,
   which §6.1.8 records as the cause of the withdrawn dereference fix too.
   Status: PARSER, blocked on 01a0ac52 — not an open vocabulary question.

5. **Keep `Slice.Kind` explicit.** The four slice kinds produce four
   different result types (§4.4.7); collapsing them to "slice" loses the
   information PSC needs most. Re-measured on perl 5.42.0: `@a[0,2]` gives
   elements, `@h{qw(x y)}` values, `%a[0,1]` index/element pairs, `%h{'x'}`
   key/value pairs.

   The CST carries the discriminant: every slice is `Index` (the sole
   `Kind: Index` construction site is `parseSubscript`, `expr.go`), and the
   pair (base child's first byte, `Index.Text`) selects the kind — measured
   `@a[0,1]` → `index "["` over `term "@a"`; `@h{'a','b'}` → `index "{"`
   over `term "@h"`; `%h{'a'}` → `index "{"` over `term "%h"`; `%a[0,1]` →
   `index "["` over `term "%a"`; `(1,2)[0]` → `index "["` over a `list`
   (SliceList). Status: LOWERING, one five-way switch, and the largest single
   block of this list the lowering gets for free.

   Inference branches only two ways on it today — KV or not, because only
   the element type is at stake — and does not branch at all yet: the old
   grammar's `slice_expression`, `slice_container_variable` and
   `anonymous_slice_expression` appear nowhere in `internal/infer`, and
   measured, `my @a=(1,2); my @s=@a[0,1];` gives `@a elem=Int` but
   `@s elem=Unknown`. The other three values serve the LSP and diagnostics.
   B::SoN maps `aslice`, `kvaslice`, `hslice`, `kvhslice` all to one `Slice`
   (`FromOptree/OpMap.pm:209-231`) and `aelem`, `helem`, `multideref`,
   `gelem` all to one `Subscript` (`:205-208`), because the optree it reads
   has already typed the result. That is a phase difference, not a
   disagreement; PSC is upstream of it and must not follow.

6. **A captured variable is a shared mutable cell, not a value.** This is a
   fact about Perl rather than about any node kind, and it constrains every
   consumer of this vocabulary:

   ```perl
   my $n = 5; my $c = sub { $n }; $n = 99;   # $c->() is 99, not 5
   for my $i (1..3) { push @s, sub { $i } }  # 1, 2, 3 -- not 3, 3, 3
   ```

   Both measured on perl 5.42.0, 2026-09-17.

   The first line says a closure reads the cell at call time, so inference
   cannot type `$n` inside the sub from its value at the point of capture.
   The second says each loop iteration creates a *distinct* cell, so the
   three `$i` references are three cells that happen to share a name and a
   source span. Any scheme that identifies captured variables by name, by
   declaration site, or by source position collapses them and gets 3,3,3.

   B::SoN reached the same conclusion from the other end and encodes it in
   its node vocabulary. `lib/SoN/IR/Node/MakeCell.pm`, read directly: "NEVER
   HASH-CONSED. Two structurally identical MakeCells are two DIFFERENT
   cells ... One node with three executions yields three cells." That
   phrasing settles what §6.1.6 item 4's identity scheme must mean: node
   identity is per *syntax site* — one id for the `sub { $i }`, which is
   correct for the tree and for the position map — and any table keyed on a
   captured *value* must key on (node id, capture instance), not on node id
   alone. `AnonSub.pm` splits the two the same way: "the name addresses one
   shared body, the inputs are the per-closure environment." This list has
   no cell node and does not need one: the cell is a property of scope
   resolution, not of syntax.

7. **`(Sigil, Name)` is `Var`'s identity, and the sigil is load-bearing.**
   `$_`, `@_` and the file-test bareword `_` are three different things.
   B::SoN's `PadAccess.pm` and `EntryDef.pm` both die on a missing sigil,
   and record why: a name-only identity hash-consed `$_` and `@_` into one
   node, so a `shift @_` and a regex subject shared a definition. A syntax
   tree does not hash-cons, so the failure cannot recur here in that form —
   but three paths in `internal/infer` re-derive `@_` and `$_[0]` from
   source text (`infer.go:3659`, `:3753`, `:4193`), which is the same
   mistake in a different coat. The lowering produces `Var{"@","_"}` and
   `Index{Var{"$","_"}, 0}` and those rules compare fields. No
   `ImplicitTopic` or `ArgsSource` node: B::SoN's `ArgsSource` exists to
   stop the hash-consing, which is a graph problem, and `nil` remains the
   spelling for an elided `$_` on `FileTest` and `Match`.

### 4.14.2 The vocabulary, entry by entry

Kind by kind: what the CST gives the lowering, who consumes it, and the
status. "C" counts are uses in `internal/infer` of the old grammar name.

**Literals and names**

| Kind | CST at 3c997a6b | Consumer / perl / SoN | Status |
|---|---|---|---|
| `NumLit` | `term "42"`, `term "0x1f"`, `term "1_000"`; `IsInt` from `Text` | C: `number` (2) | SHIPPED |
| `StrLit` | `term "'ab'"`, `term "\"ls\""`; delimiter is `Text[0]` | C: `string_literal` (2), `string_content` (2, text only) | SHIPPED |
| `StrLit.Backtick` | a backtick string is one `term` whose `Text` keeps the backtick delimiters, so its first byte separates it from `"ls"`. `qx/ls/;` → `Unknown`: `internal/lexer/quote.go:22-31` has no `qx` row | SoN: `BacktickExpr`, a language fact | LOWERING for backticks; PARSER for `qx//` |
| `Interp` | `"a$b c"` is one `term`; `lexer.go`'s `Quote` doc defers interpolation "when there is a parser to consume the pieces" | C: `string_content` (1, `infer.go:1569`) reads text only | PARSER |
| `QwList` | `term "qw(a b)"`, opaque; the lowering splits the body | perl: `(qw(a b)) x 2` and `("a","b") x 2` both `a b a b`; scalar `qw(a b c)` and `("a","b","c")` both `c` | LOWERING, and the body split is assigned to the lowering here |
| `Bareword{Quoted}` | positional over `Call{Resolved:false}` (rule 3) | C: `bareword` (7); old grammar has `autoquoted_bareword` | LOWERING; parser fix preferred |

**Variables and dereference**

| Kind | CST at 3c997a6b | Consumer / perl / SoN | Status |
|---|---|---|---|
| `Var{Sigil,Name}` | `term "$x"`, `"@a"`, `"%h"`, `"*foo"`, `"$#a"`, `"&foo"`, `"$Foo::x"`; sigil is the leading bytes. `$::x;` → `Unknown` | C: `scalar` (17), `array` (9), `hash` (10), `varname` — and at 8 of 9 switch sites the three arms are identical modulo the sigil string (`infer.go:1669-1673`, `:3560-3564`, `:3851-3855`, `:4072-4077`, `:1942-1946`, `:3088-3090`, `:1402-1406`); `lookupNarrowedType` (`:1890-1896`) and `sigildName` (`symbols.go:403`) take the sigil as a parameter. Sigil is a field, not three kinds; issue 01a0ad52's counts measure matching, not distinguishing. SoN: `(sigil, symbol)` identity, rule 7 | SHIPPED; `$::x` PARSER |
| `Deref{Sigil,Expr}` | `${$x}`, `$$x`, `@{$x}`, `@$x`, `%$x`, `$#{$x}`, `$#$x` are each one childless `term` (issue 01a0ad52). The held `DerefSigil` fix measured worse on T1 (406 clean / 4,986 `Unknown`) and is blocked on 01a0ac52. `*{"foo"}` → `term "*"` over `anon_hash` — a wrong tree that round-trips | C: `array_deref_expression`, `hash_deref_expression` (1 each) | PARSER |
| `Deref{Postfix}` | `$x->@*` → `[binary "->" [term "$x"] [term "@*"]]`, likewise `->%*`; `$x->$#*` → `binary "*"` with an `Unknown`; `$x->@[0,1]` → `Unknown` | §4.4.6 | LOWERING for `->@*`/`->%*`; PARSER for `->$#*` and postfix slices |
| `Index`, `Key` | `index "["` / `index "{"` over base and subscript; chains nest (`$x->{a}[0]{b}`) | C: `array_element_expression` (4), `hash_element_expression` (1) | SHIPPED; `Arrow` PARSER (rule 2) |
| `Slice{Kind}` | rule 5 | rule 5 | LOWERING |
| `Call.Fn` (call through a reference) | `$code->(1)` → `[binary "->" [term "$code"] [term "1"]]` (the `(1)` is `finishList`'s rewrite; `(1,2)` is a `list`); `&{$code}(1)` → `[index "(" [index "{" [term "&"] [term "$code"]] [term "1"]]`; `&$code(1);` → `Unknown` | §4.8.3 | LOWERING for the first two shapes; PARSER for `&$code()` |

**Operators**

| Kind | CST at 3c997a6b | Consumer / perl / SoN | Status |
|---|---|---|---|
| `Unary{Op,Postfix}` | `unary "!"`, `"neg"`, `"~"`, `"++"`, `"ref"` (for `\`), `"pos"` (for unary `+`); `postfix "++"` | C: `unary_expression` (1); `extractNegatedGuard` (`infer.go:2285`) tests text `== "!"` | SHIPPED |
| `Binary{Op}` | `binary` with `Text` the operator | C: `binary_expression` (3), `and`/`or` (1 each); `findOperatorText` (`infer.go:645-659`) re-derives the operator from anonymous tokens. SoN: `NumEq`/`StrEq`/`NumLt`/`StrLt` are split by perl's *op name* (`OpMap.pm:134-151`) — a total function of `Op`, and every SoN BinOp carries `op_str()` with the spelling. Keep one `Binary{Op}`; the table `==`→`NumEq`, `eq`→`StrEq`, `<`→`NumLt`, `lt`→`StrLt`, `.`→`Concat`, `x`→`Repeat`, `//`→`DefinedOr` is the bridge | SHIPPED; the lowering *sets* `Op` from `Text` |
| `Repeat{ListRepeat}` | `binary "x"`; `ListRepeat` from rule 1 or the `qw` prefix | §4.10 | LOWERING once rule 1's bit exists |
| `Assign{ListAssign}` | `binary "="`, `"+="`, ...; `($a,$b) = (1,2)` has a `list` LHS; `@a = (1,2)` an aggregate `term` | C: `assignment_expression` (3) | LOWERING; `my (...) =` is under `Decl` |
| `Ternary` | `ternary "?:"` | C: `conditional_expression` (1) | SHIPPED |
| `CmpChain` | `cmp_chain "<"` with flat operands | C: none; no chained-comparison rule exists | SHIPPED; the typing rule is unwritten (§4.14.6) |
| `Range{Exclusive}` | `$a..$b` → `binary ".."`, `$a...$b` → `binary "..."`; **`1..2` → one `term "1..2"`**: the lexer's number scan takes the range as one `Number` token, so literal endpoints have no `Lo`/`Hi` | C: measured defect — `my @r = (1..5)` infers `@r` as `Str`, because `findOperatorText` matches the anonymous `.` before ever seeing `..` while `signatures.go:182` has `"..": {Result: List}`; perl: `$r[0]+1` is 2 | LOWERING for variable endpoints; PARSER (lexer) for literal ones. Required, not optional |
| `List{Fat}` | `(a => 1, b => 2)` → `list` with four flat children, identical to `(a, 1, b, 2)`; `parseList` consumes `,` and `=>` alike (`term.go:275-276`). Inside call arguments the fat comma survives: `f(a => 1)` → `[call "f" [binary "=>" ...]]` | §4.5.4: `=>` quotes the word to its left | SHIPPED as `List`; **`Fat` PARSER** — inside a `List`, `AnonArray` or `AnonHash` the fact is destroyed |
| `Paren` | rule 1 | rule 1 | PARSER |

**Declarations**

| Kind | CST at 3c997a6b | Consumer / perl / SoN | Status |
|---|---|---|---|
| `Decl{Kind}` | `declaration` with `Text` the declarator; T1 census of `Declaration.Text`: my 4,905, sub 1,775, package 569, local 398, our 318, state 9, field 7, class 5, method 4 | C: `variable_declaration` (3); `local`/`state`/`our`/`field` unhandled | SHIPPED |
| `Decl{Vars,Init}` | two shapes: `my $x = f()` → `[declaration "my" [term "$x"] [call "f"]]` (two children, no `=`); `my ($x) = f()` → `[declaration "my" [binary "=" [term "$x"] [call "f"]]]` (one child, the `=`); `my ($x,$y) = f()` → the same with a `list` LHS. The list-vs-scalar fact §4.12.2 rule 2 needs is *the child count* — perl: with `sub f{(1,2,3)}`, `my ($x)=f()` gives 1 and `my $y=f()` gives 3 | C: `extractVarNameFromDecl` (`infer.go:3560`) | LOWERING, by the predicate "one child and it is `binary "="`"; normalising the parser to one shape is recommended before the lowering is written |
| `Decl{Attrs}` | `my $x :shared` → `term ":shared"` as a second child | — | LOWERING |
| `Decl{RefAlias}` | `my \$x = $y` → `[declaration "my" [binary "=" [unary "ref" [term "$x"]] [term "$y"]]]` | C: none | LOWERING; SPECULATIVE consumer |
| sub declarations | `sub f { 1 }` → `[declaration "sub" [term "f"] [block ...]]`; `sub f;` likewise without a block; `sub f :prototype($) { 1 }` → `declaration "sub"` spanning `sub f` then `Unknown` | C: `subroutine_declaration_statement` (8) | SHIPPED; attributes PARSER |
| signatures | without the feature, `sub f ($x, $y) { 1 }` → `[declaration "sub" [term "f"] [prototype "($x, $y)"] [block ...]]`, which is what perl does too; with `use v5.36;` the same text → `declaration "sub"` spanning `sub f`, then a separate statement `[index "{" [list [term "$x"] [term "$y"]] [term "1"]]` — the body read as a hash slice of the signature. `internal/lexer/proto.go` returns false when `l.signatures` is set and nothing downstream picks the signature up | C: `signature` (2), `mandatory_parameter` (2), both unwrapped to their `scalar` | PARSER |

**Calls**

| Kind | CST at 3c997a6b | Consumer / perl / SoN | Status |
|---|---|---|---|
| `Call{Name,Args,Resolved}` | `call` with `Text` the name, `Resolved` set for builtins and subs declared in the file; `f 1,2` and `foo 1,2` → `Unknown` by design (§4.8.3) | C: five sites do the same `GetBuiltin` + arity + argument-type check and differ only in where the name hides — `infer.go:693` (`function_call_expression`, `ambiguous_function_call_expression`), `:881` (`func1op_call_expression`), `:1009` (`func0op_call_expression`), `:4004`, `:4048` (`function` or `bareword` child). All five lower to this one kind | SHIPPED |
| `Call.Parenthesized` | `f(1,2)` → `call "f"` whose span includes the `(`; `map $_+1, @a` does not | §4.8.1 | LOWERING |
| `Call.Ampersand`, `.BareAmpersand` | `&foo;` → `term "&foo"`; `&foo();` → `[index "(" [term "&foo"]]` | §4.8.3 | LOWERING (a `term` whose `Text` starts with `&`; bare iff no enclosing `index "("`) |
| `Call.Handle` | `print $fh "x"` → `[call "print" [term "$fh"] [term "\"x\""]]` (two children); `print $fh, "x"` → `[call "print" [binary "," ...]]` (one); `print STDERR "x"` likewise two children. `parseFilehandleSlot` (`call.go:167`) decides the slot at construction and emits a plain `term`. `print {$fh} "x"` → `Unknown`, though `call.go:177` has the branch for it — the lexer is not marking that `{` as `OpensBlock` | C: `infer.go:1155-1165` documents that counting the handle as argument 1 made every typed-handle `print` a false `Str` mismatch, and skips `indirect_object` for it | LOWERING by child shape, fragile; a bit at `call.go:167` is the right fix; `{$fh}` PARSER |
| `MethodCall` | three shapes: `$obj->method(1)` → `[binary "->" [term "$obj"] [call "method" [term "1"]]]`; `Foo->new` → `[binary "->" [call "Foo"] [call "new"]]`; `$obj->$name(1)` → `[binary "->" [term "$obj"] [index "(" [term "$name"] [term "1"]]]`. `HasParens` from the `call` span | C: `method_call_expression` (1), `method` (1) | LOWERING; `Indirect`/`Ambiguous` PARSER (`new Foo(1)` → `Unknown`) |

**Special forms**

| Kind | CST at 3c997a6b | Consumer / perl / SoN | Status |
|---|---|---|---|
| `MapGrepSort` | block forms all `Unknown`; expression forms are `call "map"`/`"sort"` `resolved=true` | C: `map_grep_expression`, `sort_expression` (2), `map`, `sort`; `callElementType` (`infer.go:1313`) tells map from grep by `strings.HasPrefix(... "grep")` on source text — two operations with opposite element-type rules, separated by string matching. `Op` fixes exactly that | expression form LOWERING; block form PARSER, blocked on 01a0ac52 (rule 4) |
| `DoBlock` | `do { 1 }` → `[call "do" [block ...]]` | C: `blockResultType` (`infer.go:1417`) already computes a block's value for `map` | LOWERING (rename) |
| `DoFile` | `do 'file.pl';` → `Unknown` | C: none | PARSER; SPECULATIVE consumer |
| `EvalBlock`, `EvalStr` | `eval { 1 }` → `[call "eval" [block ...]]`; `eval "1"` and `eval $s` → `[call "eval" [term ...]]` | C: none. Needed anyway: `EvalStr` is the soundness boundary where the result is `Any` | LOWERING |
| `AnonSub` | in expression position, `my $f = sub { 1 }` → `[declaration "my" [term "$f"] [declaration "sub" [block ...]]]`. **In statement position `sub { 1 };` → two statements**, `[declaration "sub"]` spanning only `sub` and `[anon_hash [term "1"]]` — a wrong tree that round-trips, which is the failure `Unknown` exists to prevent (`parse.go:14-26`). `statement()` reaches `parseDeclaration` before `parseTerm`'s anon-sub check, and `finishBodyOrSemicolon` sees a `{` the lexer classified as a term. `Sig`: PARSER (signatures row above). With `use v5.36;`, `my $f = sub ($x) { 1 };` → `[index "{" [call "sub" [term "$x"]] [term "1"]]` | C: `anonymous_subroutine_expression` (1) | SHIPPED in expression position; PARSER in statement position and for `Sig` |
| `AnonList`, `AnonHash` | `anon_array`, `anon_hash`; `+{a=>1}` → `[unary "pos" [anon_hash ...]]`. **`{a=>1};` at statement start → `block`** holding a `=>` statement, then an `Unknown` for the `;`. perl disagrees: `B::Deparse` gives `+{'a', 1};` — the `toke.c:6698-6842` heuristic fires at statement start too. A wrong tree that round-trips | C: `anonymous_array_expression` (3), `anonymous_hash_expression` (1). SoN: `ArrayLiteral`/`HashLiteral`, renamed from `ArrayRef`/`HashRef` after the name asserted reference-ness the stamp contradicted (37 corpus cases). Here `[...]` *is* unambiguously a reference constructor, so the local names are not unsound; the rename is §4.14.6 | SHIPPED; statement-start `{` PARSER |
| `Ref` | `\@a` → `[unary "ref" [term "@a"]]`; `\($a,$b)` → `[unary "ref" [list ...]]` | C: `refgen_expression` (1), `ref` (1) | LOWERING |
| `FileTest` | `-e $f` → `[call "-e" resolved=true [term "$f"]]`; `-r _` → operand `call "_"` (rule 3) | C: none as a kind; the old grammar folds `-e` into `func1op_call_expression` and the builtin table returns `Bool`, which works | LOWERING (rename); SPECULATIVE as a distinct kind — no file-test-specific rule exists |
| `Readline{Magic}` | `<STDIN>`, `<$fh>`, `<>` → one `term`; **`<<>>;` → `[binary ">>" [unknown "<<"] [unknown ">"]]`** — the lexer does not know the 5.22 form | C: none; `infer.go:78-84` uses `my $line = <$T>` as the canonical irreducible `Unknown`, which is what the kind lets it name | LOWERING for `<FH>`/`<$fh>`/`<>`; PARSER (lexer) for `<<>>` |
| `Glob` | `<*.c>;` → `Unknown`; `glob("*.c")` → `call "glob" resolved=false` | C: none | PARSER for `<...>`; SPECULATIVE consumer |
| `LoopEx` | `last FOO` → `[loop_control "last" [label "FOO"]]`; `next;` likewise; `return` is also `loop_control` (`return (1,2)` → `[loop_control "return" [list ...]]`); `goto &foo;` → `Unknown` | C: `return_expression` (3); `blockAlwaysExits` (`infer.go:3312-3342`) detects `return`/`die`/`exit` by text and will need `last`/`next` for correct branch joins | SHIPPED; `goto` PARSER |
| `Wantarray` | `wantarray;` → `call "wantarray" resolved=true` | C: `NarrowByContext` (`infer.go:1291`) exists and has no context source. SoN: `Wantarray`, the one exact-name agreement, and language by SoN's own test (`Wantarray.pm`: perl made it a runtime function because an interpreter cannot see a sub's callers) | LOWERING (rename) |
| `Match{Op}` | `s/a/b/g`, `m/x/`, `qr/x/`, `/a/`, `tr/a/b/`, `y/a/b/` → one `term` each with the raw operator as `Text` | C: `match_regexp`, `quoted_regexp`, `regexp_content` (1 each); `nearestPatternBefore` (`infer.go:1724-1751`) needs source order (§4.14 preamble) | SHIPPED as an opaque leaf |
| `Match{Pattern,Replace,Flags}` | inside the one `Quote` token, by design (`quote.go`) | SoN: `Transliterate.pm` records that reading `tr`'s operands as a regex compiles something the source never wrote — hence the field note in the listing | PARSER |
| `Match{Target,Negated}` | `$x =~ /a/` → `binary "=~"`, `$x !~ /a/` → `binary "!~"` | — | LOWERING |
| `Hole` | `Unknown`: 4,944 in T1, every one a childless leaf (0 with children, 0 nested); parents: `source_file` 3,851, `block` 690, `binary` 281, `index` 74, `loop` 36, `loop_control` 6, `cmp_chain` 2, `call` 2, `unary` 1, `declaration` 1. Spans p50 30 bytes, p90 345, p99 1,532, max 14,794; 29.1% of corpus bytes | §6.1.6 item 7 asks every descending rule to check for `Unknown`; nothing can descend into a leaf. The rule is instead: a `Hole` may stand anywhere a `Node` may (the census shows it as a `binary` operand and an `index` base or subscript), inference types it `Any` and never inspects it, and a position query inside its span returns the `Hole` so the LSP reports "no information" rather than a type | LOWERING; new in this list |

Named builtins that B::SoN gives their own kind — `Delete`, `Exists`,
`Chomp` (chomp/chop), `Count`, `Length` — are `Call{Name}` here, and the
sibling's reasons for splitting them (perl has distinct `schomp`/`schop` ops;
`exists` is membership, not definedness — measured, `exists $h{u}` is true
for `u => undef`) are facts a `Call`'s name already carries. Two of them do
not parse: `delete $h{k};` and `exists $h{k};` are each `Unknown`, because
neither word is in `internal/parse/keyword.go`. `chomp $x;` parses. That
is a keyword-table gap, PARSER, and independent of this list.

### 4.14.3 The flags

§4.14 specifies fifteen flags. `parse.Node` has one. For each: does the CST
carry the fact, does the lowering re-derive it, or must the parser change.

| Flag | Carried how | Status |
|---|---|---|
| `Call.Resolved` | `parse.Node.Resolved` | SHIPPED |
| `Unary.Postfix` | `Kind == Postfix` vs `Kind == Unary` | SHIPPED (a kind split) |
| `Repeat.ListRepeat` | rule 1's paren fact, or `Text` prefix `qw` | LOWERING after the `Paren` bit; unsound from spans alone |
| `Assign.ListAssign` | LHS is a `list`, or an aggregate-sigil `term`, or rule 1's paren fact | LOWERING |
| `Decl`'s list form | child count: one child that is `binary "="` | LOWERING |
| `Decl.RefAlias` | `unary "ref"` as the assignment's LHS | LOWERING |
| `Bareword.Quoted` | position (rule 3) | LOWERING; parser fix preferred |
| `Call.Parenthesized` | `(` inside the `call` span after the name | LOWERING |
| `Call.Ampersand` | `term` `Text` starts with `&` | LOWERING |
| `Call.BareAmpersand` | that `term` with no enclosing `index "("`; `&$code()` is `Unknown` | LOWERING, partial |
| `Slice.Kind` | base sigil × `Index.Text` | LOWERING |
| `Range.Exclusive` | `binary ".."` vs `"..."` | LOWERING; literal endpoints PARSER |
| `Match.Negated` | `binary "=~"` vs `"!~"` | LOWERING |
| `Readline.Magic` | `term "<>"`; `<<>>` does not lex | LOWERING for `<>`; PARSER for `<<>>` |
| `Index.Arrow`, `Key.Arrow` | not in the tree; identical shapes asserted by test | PARSER |
| `List.Fat` | not in the tree inside `list`/`anon_array`/`anon_hash` | PARSER |
| `MapGrepSort.BlockGuessed` | no node to flag | PARSER, blocked on 01a0ac52 |
| `MethodCall.Indirect`, `.Ambiguous` | no node to flag | PARSER; keep-or-cut is §4.14.6 |
| `StrLit.Backtick` | `Text[0] == '`'` | LOWERING |

Three of these ask for a bit on `parse.Node` — `Paren`, `Arrow`, and the
filehandle slot — and `Resolved` (`parse.go:234`) is the precedent for
each. That is a smaller change than either re-deriving from source bytes
(the leak §6.1.6 criticises) or building a richer CST.

### 4.14.4 Shared names with B::SoN

The test adopted from that project: a kind is a *language* fact if a Perl
programmer can point at source and say "that is what this is", an
*artifact* if it exists only because the IR is in SSA/graph form.

**Same operation, same name.** `Wantarray`. That is the whole list, and it
needed no work.

**Same operation, different name, keep ours.** `AnonList`/`AnonHash` against
`ArrayLiteral`/`HashLiteral` (§4.14.6); `Binary{Op}` against the
`NumEq`/`StrEq`/`Concat`/`Repeat`/`DefinedOr` family, where the mapping is a
total table from `Op` (§4.14.2, Operators); `Match` against `RegexMatch`/`RegexSubst`/`RegexCapture`;
`Var` against `PadAccess`/`EntryDef`/`EntryWrite`, where the lexical-vs-
package split is a scope-resolution result (`PadAccess` carries a pad
`targ`, `EntryDef` a package) that belongs in §6.1.6 item 4's side table,
not in a node kind.

**Ours is finer, and correctly so.** `Index`/`Key`/`Slice{Kind}` against
one `Subscript` and one `Slice` (rule 5). `Paren`, which has no SoN
counterpart at all — `grep -rn Paren lib/SoN/IR/Node/` returns nothing; the
only trace is `Call.paren_form`, scoped to whether a bare list builtin
absorbs the remaining list. Perl's parser resolves parens into optree shape
before B::SoN sees anything. Both are phase differences: the optree has
already been typed, and a syntax tree has not. A future reader should not
"harmonise" either away.

**Theirs names something ours lacks.** `BacktickExpr` — now
`StrLit.Backtick`. `Transliterate` — a note on `Match`, not a node, until a
consumer needs the split. `Delete`, `Exists`, `Chomp`, `Count`, `Length` —
`Call` by name, above. `ArgsSource`, `Parameter`, the `Cell*` group — rules
6 and 7: language facts, already covered by `Var{"@","_"}`, `AnonSub.Sig`
and scope resolution respectively.

**Artifacts, excluded.** Ten, not the five earlier text named: `Phi`,
`Region`, `Proj`, `Start`, `MemStart` (graph and memory-SSA); `StructRef`
and `StructFieldAccess` (both `ABOUTME` lines say "from the StructPromotion
optimizer"); `ExpressionList` (a call's argument list as a node —
`Call.Args` holds it structurally); `ListAppend` ("the value a map/grep LOOP
CARRIES", input 0 a loop `Phi`); and `Coerce`. Chapter 6 §6.1.8 classified
`Coerce` as language. B::SoN's own `docs/plans/2026-09-17-the-gap-shape-is-
inverted.md` says otherwise: "Coerce(Scalar -> Code) is standing in for
'compile and run arbitrary perl'. There is no node for that" — `Coerce` is
where that IR puts what it cannot name, which is the opposite of something
a programmer can point at. A syntax tree has nothing to put there.

Adopt no B::SoN name wholesale. The audit found one exact agreement and no
case where the sibling's name is better than the one it would replace. What
is shared is the stamp lattice and `TypeLibrary` format of §6.1.6 item 6,
and the mechanical `Op`-to-kind table above. UNVERIFIED and out of scope
here: whether `TypeLibrary` keys on node kinds — if it does, the
`Subscript`-vs-`Index`/`Key` split above is a contract question rather than
a free local choice. Settled by reading `lib/SoN/IR/Stamp.pm` and the
`TypeLibrary` loader together; note that `Stamp.pm:11-40` gives a DAG with
`Boolean` and `Undef` as children of `Scalar` beside `Str`, which is richer
than the linear chain §6.1.6 item 6 quotes.

### 4.14.5 What the lowering erases

Each candidate erasure, whether it is VARIATION (two spellings, one
meaning; erase) or STRUCTURE (the spelling *is* the meaning; keep), and what
breaks if a STRUCTURE item is erased anyway.

| Erasure | Class | Evidence and consequence |
|---|---|---|
| Trivia | VARIATION | 43,250 CST nodes (15.9%); the only material size reduction the lowering makes |
| `$h{k}` / `$h->{k}` to one `Key` | STRUCTURE unless `Arrow` survives | Rule 2: they read different variables. Erased today, and the parser did it |
| `$$x[0]` / `${$x}[0]` / `$x->[0]` to one `Index` | VARIATION for the spelling, STRUCTURE for deref-ness | Same runtime op; the lowering keeps "base is a `Deref`" and the LSP keeps `Arrow` |
| `($x)` to `$x` | STRUCTURE | Rule 1: `('a') x 3` vs `'a' x 3`; `my ($x) = f()` vs `my $x = f()`. Erased today at `finishList` |
| `unless (c) {A} else {B}` to `if` | VARIATION, as a branch swap | `perl -MO=Deparse` gives `if ($c) {B} else {A}` — a swap, not `if (!c)`. Specify the swap; a synthesised `Not` is a node flow narrowing would have to see through |
| `until` to `while` | VARIATION for the modifier and block forms; **STRUCTURE for `do {} until`** | perl: `$m++ until 1` leaves 0; `do { $n++ } until 1` leaves 1. Both are `loop` nodes; the do-form is `loop "while"`/`"until"` whose child 0 is `call "do"` over a `block`. The lowering needs a `PostCond` bit or a distinct kind — this list has neither and chapter 5 owns the statement kinds |
| Statement-modifier `if`/`unless` to block form | VARIATION | `$x = 1 if $c` → `conditional "if"` with children `[body, cond]`, the reverse of the block form's `[cond, block]`; the lowering dispatches on child order, not kind |
| `qw(a b)` to `List` of `StrLit` | VARIATION | perl: identical to `("a","b")` under `x` and in scalar context. `QwList` may still be kept for the LSP |
| `a => 1` to `"a", 1` | STRUCTURE for the left word | §4.5.4; without `Fat` the lowering cannot tell `(a, 1)` (a call) from `(a => 1)` (a string). Destroyed today inside `list`/`anon_array`/`anon_hash` |
| `1..2` to `Range` | STRUCTURE | The lexer destroys it for literal endpoints; inference measurably needs it |
| `&foo` / `&foo()` / `foo()` to one `Call` | STRUCTURE for the bare form | §4.8.3: `&foo` passes the caller's `@_`; `Ampersand`/`BareAmpersand` keep it |
| `-e $f` to `Call{"-e"}` | VARIATION | The consumer already handles it as a builtin with a signature |
| `do {}`, `eval {}`, `wantarray` to `Call` | VARIATION in the CST, STRUCTURE in this list | Each has a semantic rule of its own (a block value; a soundness boundary; context polymorphism); the lowering renames them |
| Five call spellings to one `Call` | VARIATION | The consumer's five identical `GetBuiltin` sites are the proof |
| The `print $fh` handle to argument 1 | STRUCTURE | `infer.go:1155-1165`: measured false `Str` mismatch on every typed-handle print |
| `Unknown` to `Hole` | VARIATION | A leaf to a leaf; the span and the "never inspect" rule are what matter |

### 4.14.6 Open after the reconciliation

Not resolved here; each names what would settle it.

- **`Paren`: node or bit.** Rule 1 says the information survives in spans
  and that a `Paren bool` on `parse.Node` is the exact fix. Whether the
  *lowered* tree then carries a `Paren` node or a `Parenthesized` flag on
  `Repeat`/`Assign`/`Decl` (Pyright's `hasParens`) is a choice between the
  position map's convenience (rust-analyzer's `lower.rs:1630` maps the paren
  forward to the inner node) and one fewer kind. Only two rules read it.
  Settled by writing those two rules and the position-map entry.
- **`MethodCall.Indirect` / `.Ambiguous`.** `new Foo(1)` is `Unknown`, so
  the honest-uncertainty channel of rule 3 carries no signal for indirect
  object syntax. §4.5.3 calls the syntax a disaster. If it is permanently
  out of M1 scope, drop both fields rather than ship fields that are always
  false. Settled by a scope decision, not a measurement.
- **`AnonList`/`AnonHash` versus `ArrayLiteral`/`HashLiteral`.** The only
  rename on which the audit is not unilateral. B::SoN paid for `ArrayRef`
  with 37 miscompiled corpus cases; the local names are not unsound, since
  a syntax tree's `[...]` is unambiguously a reference constructor. Name
  agreement against a correct local name: perigrin's call.
- **`CmpChain`'s typing rule.** The parser ships the kind, so the lowering
  must accept it, and `internal/infer` has no chained-comparison logic.
  Settled by writing the rule and seeing whether it wants a flat operand
  list or nested `Binary`s.
- **`Interp.Parts` and `Match.Pattern`.** The consumer reads
  `string_content` and `regexp_content` once each. Whether it wants the
  pieces or only the text is settled by reading those two sites
  (`infer.go:1569`, `:1741`); one use each is weak evidence either way.
- **`local`/`state`/`our`/`field` as `Decl.Kind` values for inference.**
  The consumer handles none of the four. `local` restores on scope exit,
  which is a flow fact; whether flow analysis must model dynamic scoping is
  a §6.1.7 question and §6.1.7 is unbuilt.
- **Whether `Arrow` is a regression.** d2ebe02a unified the shapes and
  dropped the bit. Before it, `$h->{k}` was `[binary "->" ...]` and the
  operator's `Text` held the arrow. UNVERIFIED that the pre-fix shape was
  otherwise usable; settled by `git checkout 3b9422eb -- internal/parse`
  and one dump. Either way the fix is the bit, not a revert.
- **The `Hole` invariants.** The census shows `Unknown` as a `binary`
  operand (281), an `index` base or subscript (74), and once each under
  `unary` and `declaration`. Whether every §4.14 field that is typed `Node`
  tolerates a `Hole` — `Decl.Init`, `Index.Base`, `Match.Target` — is a
  property of the rules, not of the list, and is settled when they are
  written.
- **Whether the three measured consumer defects are tracked.** Range as
  `Str`, slice element-type loss, and the `StartByte` collision: no issue
  search was made. If any is open, this section should reference it.

---

## 4.15 File tests, readline, and glob

### 4.15.1 File test operators

`toke.c:6203-6262` recognises `-` followed by a single letter followed by a
non-identifier character, and emits `FTST(op)` — which is `UNIOP`, level 19
(§4.1.2). The letters: `r w x o R W X O e z s f d l p S b c t u g k T B A M C`.

```perl
-e $file            # exists
-d $dir             # is directory
-s $file            # size in bytes, or undef/0
-M $file            # age in days since script start
-e $file && -r _    # `_` reuses the previous stat buffer -- a magic bareword
-e -f -r $file      # STACKED (5.10+): -e($file) && -f($file) && -r($file)
```

The stacked form is a parse consequence of level 19 being `%nonassoc` combined
with `-f $file` being a valid operand for `-e`. `toke.c:6252` falls back to
"minus followed by a one-letter sub call" when the letter is not a known test,
so `-x` where `x` is a user sub is `-(x())`, not a file test — another place
the parse depends on the symbol table.

`_` as the operand is a bareword that names the special `*_` filehandle
holding the last `stat` result. Parse it as a `Bareword`, not a `Var`.

### 4.15.2 Readline and glob

`toke.c:12051-12200` (`S_scan_inputsymbol`) handles every `<...>` form. The
lexer decides between readline and glob by inspecting the contents:

```perl
<FH>          # readline on the FH filehandle       -- bareword contents
<$fh>         # readline on the handle in $fh       -- scalar contents
<STDIN>       # readline on STDIN
<>            # readline on ARGV -- toke.c:12136 rewrites <> to <ARGV>
<<>>          # 5.22+: readline ARGV without magic open (no 2-arg open,
              #        so filenames with '|' or '<' are safe) -- toke.c:12060
<*.c>         # glob -- toke.c:12123 sets pl_yylval.ival = OP_GLOB
<$dir/*>      # glob with interpolation
glob('*.c')   # the same thing spelled as a function
```

The rule (`toke.c:12070-12200`): if the contents are empty, a bareword
filehandle name (`\w+` possibly with `::`), or a single scalar variable, it is
readline; otherwise it is glob. A Go implementation must reproduce this,
because `<$x>` and `<$x/*>` differ in meaning by one character.

Context matters here more than for most operators:

```perl
my $line  = <$fh>;      # scalar context: one line
my @lines = <$fh>;      # list context: ALL lines, unbounded memory
while (my $l = <$fh>) { }   # scalar; also gets the implicit `defined` wrapper
while (<$fh>) { }           # implicit assignment to $_ AND implicit defined()
```

That last implicit `defined` wrapper is applied by the compiler when a
`readline` is the entire condition of a `while`. It is a statement-level
transformation, covered in the statements chapter, but the expression parser
must mark the node (`Readline`) distinctly enough for that pass to find it.

### 4.15.3 `wantarray`

`toke.c` emits `FUNC0` for `wantarray` (`perly.y:1679`). It takes no arguments
and reports the context of the current sub call at runtime: true in list
context, false (defined) in scalar, undef in void. It is the observable face of
§4.12's `CtxInherit`, and PSC should treat a sub containing `wantarray` as
context-polymorphic rather than trying to pick a branch.

---

## 4.16 Checklist for the implementer

Acceptance list; build order is chapter 6 §6.11.

- [ ] Encode the 32 levels of §4.2 as a table, not a function cascade.
- [ ] Left-assoc passes `BP`, right-assoc passes `BP-1`; loop breaks on `<=`.
- [ ] `AssocNone` levels (`..`, `...`, named unaries, `++`/`--`) reject a
      second occurrence at the same level rather than silently associating.
- [ ] `.` is at the `+`/`-` level; `x` is at the `*`/`/` level.
- [ ] `isa` is at the relational level and does not chain.
- [ ] `^^` shares `||`'s level exactly.
- [ ] File tests are named unaries at level 19, above `+` and `.`.
- [ ] Chained comparisons build one `CmpChain`, not nested `Binary`s.
- [ ] The `(`-immediately-follows test happens in the lexer, before precedence,
      for both list operators and named unaries (§4.8.1).
- [ ] Parenthesisation survives into the tree as a node or a `Paren` bit on
      `parse.Node`, set where `finishList` erases the node (§4.14.1 rule 1);
      `('a') x 3` and `'a' x 3` must differ in it (§4.10, §4.12).
- [ ] `$h{k}` and `$h->{k}` differ in exactly one field, `Arrow`
      (§4.14.1 rule 2); `subscript_test.go` asserts the difference, not the
      identity.
- [ ] The fat comma survives inside `list`, `anon_array` and `anon_hash`
      (§4.14.2, `List{Fat}`); `(a, 1)` and `(a => 1)` must differ.
- [ ] One shared `looksLikeAnonHash` predicate (chapter 3 §3.3), used by every
      `{` site (§4.4.4, §4.9.2, chapter 5 §5.12), recording that it guessed.
- [ ] Postfix deref (`->@*`, `->%*`, `->$#*`, `->&*`, `->**`) requires lexer
      state; `->%*` must not lex `%` as modulus.
- [ ] The arrow is optional between consecutive subscripts, required before a
      method call and before the first deref of a scalar.
- [ ] Context is recorded per edge during a downward pass (§4.12), and
      `return` / last-statement / `wantarray` stay context-polymorphic.
- [ ] Every guess is recorded on the node: `Resolved`, `Ambiguous`,
      `BlockGuessed`. Only `Resolved` ships; the other two have no node to
      sit on until 01a0ac52 and indirect-object syntax are parsed
      (§4.14.3).
- [ ] Nothing that round-trips is a wrong tree. Measured wrong-but-round-
      tripping trees at 3c997a6b (§4.14.2): `sub { 1 };` in statement
      position, `{a=>1};` at statement start, `*{"foo"}`, and a signature
      body under `use v5.36`.

---

## 4.17 Source citations

| Fact | Citation |
|---|---|
| Precedence declaration block | `perly.y:150-183` |
| `expr` / `listexpr` / comma | `perly.y:1249-1273` |
| `listop` (list ops, method calls, indirect object) | `perly.y:1275-1320` |
| `methodname` | `perly.y:1322-1324` |
| `subscripted` (subscripts, calls, list slices) | `perly.y:1327-1407` |
| `termbinop` | `perly.y:1409-1461` |
| `termrelop` / `relopchain` / `termeqop` / `eqopchain` | `perly.y:1462-1495` |
| `termunop` (unary, pre/post inc/dec, POSTJOIN) | `perly.y:1497-1530` |
| `anonymous` (`[]`, `{}`, `sub{}`, `method{}`) | `perly.y:1532-1555` |
| `termdo` (`do BLOCK` vs `do EXPR`) | `perly.y:1557-1562` |
| `term` (the full term list) | `perly.y:1565-1712` |
| `myattrterm` / `myterm` / `fieldvar` | `perly.y:1714-1785` |
| `amper` / `scalar` / `ary` / `hsh` / `star` / `arylen` | `perly.y:1841-1870` |
| `sliceme` / `kvslice` / `gelem` / `indirob` | `perly.y:1871-1897` |
| `S_lop` — the paren cliff for list operators | `toke.c:2160`+ |
| `UNI3` / `UNI` / `UNIDOR` — the paren cliff for named unaries | `toke.c:285-303` |
| `FTST` — file tests are `UNIOP` | `toke.c:261`, `toke.c:6203-6262` |
| `Aop` / `Mop` / `BOop` / `BAop` / `PWop` / `PMop` | `toke.c:265-274` |
| `ChEop` / `NCEop` / `ChRop` / `NCRop` | `toke.c:275-278` |
| `isa` is `NCRELOP` | `toke.c:8697` |
| `xor` is `OROP` | `toke.c:9222-9226` |
| `^^` is `OROR` | `toke.c:6429-6441` |
| `x` is `MULOP` | `toke.c:9216` |
| `sort` forces a bareword with `CHECK_KEYWORD` | `toke.c:9038-9043` |
| `map` / `grep` are `LOP(..., XREF)` | `toke.c:8585`, `toke.c:8756` |
| `->` postfix-deref lexer state (`XPOSTDEREF`) | `toke.c:6285-6293` |
| `S_scan_inputsymbol` — `<FH>`, `<>`, `<<>>`, `<*.c>` | `toke.c:12051-12200` |
| anon-hash vs block heuristic | `toke.c:6698-6842` |
| PerlOnJava precedence map | `ParserTables.java:326-349` |
| PerlOnJava right-assoc set | `ParserTables.java:48-51` |
| PerlOnJava driving loop | `Parser.java:245-340` |
| PerlOnJava infix dispatch (`,` `=>` `?` `->`) | `ParseInfix.java:248-500` |
| PerlOnJava comparison-chaining validation | `ParseInfix.java:740-800` |
| perl-lsp level cascade | `precedence.rs:424-470`, `815-836` |
| perl-lsp relational level and `isa` | `precedence.rs:752-810` |
| perl-lsp postfix deref | `postfix.rs:132-190` |
