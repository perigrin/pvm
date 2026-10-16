---
title: "Ponytail cuts in lower, parse, lexer and @INC from the PR #486 review"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T01:16:38.093747943Z
updated: 2026-10-16T01:16:38.093747943Z
---

## Context

- found by: the agentic review of PR #486, ponytail lens (half A); report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`

perigrin ruled 2026-10-16: the ponytail simplifications are filed as refactor issues, not applied on PR #486. Apply them after #486 merges, each in its own commit with tests unchanged in behaviour. Items a criterion-named test depends on stay. The specialist's findings:

#### Ponytail review, half A (PR #486 at 8a0c85c6)

Each finding gives its location, what to cut, what replaces it, the lines saved, and how safe the removal is (0-100).

Fields that a criteria-named test asserts were checked first and left out: Deref.Postfix (TestLowerPostfixDeref), Decl.Attrs/RefAlias (TestLowerDeclAttrsAndRefAlias), Call.Parenthesized/Ampersand/BareAmpersand (TestLowerCallFlags), Bareword.Quoted (TestLowerBarewordByPosition), List.Fat (TestLowerListFat), MethodCall.Indirect/HasParens (TestLowerMethodCall), and IsDeref (TestLowerDerefness). Those fields are also write-only: nothing in internal/infer or internal/psc reads them (their only reader is the reflective printer). That is a spec-level question for §4.14, not a code cut. Also noted: IsDeref (internal/lower/node.go:601) has no non-test caller.

#### Findings

1. **internal/lower/node.go:384-393, 561-563 (Match.Pattern, Match.Replace, Match.Flags)**
   - Cut: all three fields, the `add(n.Pattern...)` and `add(n.Replace...)` lines in Children, and the comments on lines 388-389, which describe content those fields never hold.
   - Why: nothing ever sets them. special.go:206 and :237 build `&Match{Op, Target, Negated}` only, and special.go:199 says the pattern "stays inside the one quote token". Positions.Quote (positions.go:220) is the ruled way to read the pattern (perigrin, 2026-10-14). No reader exists: the `.Pattern` hits in infer are Glob.Pattern, and the "Flags" hit is a comment.
   - Replace with: nothing.
   - Saves: about 6 lines.
   - Confidence: 85.

2. **internal/lower/special.go:68-74 (operand)**
   - Cut: `operand`. It is exactly `l.nth(n.Children, 0)` (stmt.go:432).
   - Replace with: `l.nth(n.Children, 0)` at its 4 call sites (special.go:57, 84, 102, 113).
   - Saves: 7 lines.
   - Confidence: 95.

3. **internal/parse/term.go:323-348 (`&` branch)**
   - Cut: the third `p.peekSignificant()` of the same token and the second, identical `&Node{Kind: Term, ...}` block.
   - Replace with:
     ```go
     name, ok := p.peekSignificant()
     if !ok { return p.refuseNotATerm(tok) }
     if p.ampTakes(name) || p.ampTakesPunct(name) { p.advanceTo(name); return &Node{Kind: Term, Text: text + p.text(name), Start: tok.Start, End: name.End} }
     ```
   - Behaviour is the same: `ampTakes` was only reached when `ok` was true.
   - Saves: about 8 lines.
   - Confidence: 90.

4. **internal/lower/ops.go:316-317, 326-328 and node.go:201, 204-210 (CmpChain.Class, TokenClass, ChRelOp, ChEqOp, eqOps)**
   - Cut: all of them.
   - Why: nothing outside lower reads Class, and infer has no `*lower.CmpChain` case at all. eqOps also re-declares parse's cmpClasses (parse/chain.go). Only TestLowerUnaryTernaryAndChain asserts Class (ops_test.go:281, 286), and that test is not criteria-named; drop those two lines from it. §4.14 lists the field, so update the spec line too.
   - Replace with: nothing.
   - Saves: about 14 lines.
   - Confidence: 75.

5. **internal/lower/print.go:310-331 vs node.go:487-593 (two walks of the same children)**
   - Cut: either one of the two walks of the same field set. `roles` (reflection) and `Children` (a 95-line switch) produce the same list. TestPrintedChildrenAreChildren (print_test.go:43) and the `PrintedChild`/`PrintedChildren` aliases in export_test.go:23-27 exist only to keep the two in sync.
   - Replace with: `Children` built as `roles(n)` with the names dropped. That deletes the switch, its panic, the sync test and the export aliases. A kind can then never ship without its walk, so the comment at node.go:484-486 goes as well.
   - Cost: reflection on the inference hot path (16 call sites). That cost has not been measured; per CLAUDE.md, measure before rejecting this.
   - Saves: about 120 lines.
   - Confidence: 55.

6. **internal/lower/node.go:90-94 (Interp)**
   - Cut: the kind. Nothing ever constructs it, because lowering emits StrLit{Interpolated}; spec row §4.14 marks it PARSER, i.e. future. It is scaffolding for later.
   - Also cut: its Children case (node.go:498-499), the `allKinds` entry, and the dead arms in infer (infer.go:113-115 and the `*lower.Interp` in vars.go:342).
   - Replace with: nothing. Add it back when the lexer splits interpolation.
   - Saves: about 10 lines.
   - Confidence: 60.

7. **internal/lower/positions.go:128, 138-143, 182-185, 239 (Positions.spans)**
   - Cut: the `spans` map. Every entry is `cst[id].Start/End`, except the root's override.
   - Replace with: a `span()` method that reads `p.cst[id]`, special-cases `n == p.root` to `statements(root)`, and is used by At as well.
   - Saves: about 4 lines and one map per file.
   - Confidence: 80.

8. **internal/lower/lower.go:11-21 vs stmt.go:440-448 (Lower duplicates block)**
   - Cut: the statement loop in Lower, which is the loop in `block`.
   - Replace with: `out := &File{Stmts: l.block(root).(*Block).Stmts}`.
   - Saves: 4 lines.
   - Confidence: 85.

9. **internal/lower/ops.go:245-252 (commaChain)**
   - Cut: the body, which is elementList's loop over one node.
   - Replace with: `if !isComma(n) { return nil }; return l.elementList([]*parse.Node{n})`.
   - Saves: 2 lines.
   - Confidence: 90.

10. **internal/lower/special.go:118, 129-135 and node.go:362-366 (Readline.Magic)**
    - Cut: the field. No consumer reads it; `Handle == nil` already says the same thing for every Readline that lowering builds. The `t == "<<>>"` branch then folds into `body == "" || t == "<<>>"`.
    - Test impact: TestLowerSpecialRenames (criteria-named) asserts the field at special_test.go:169 and 172. The test's body changes; its name stays.
    - Saves: about 4 lines.
    - Confidence: 65.

11. **internal/lower/id.go:93-103 vs decl.go:330 (two "is this a named sub" predicates)**
    - Cut: the second of two predicates that test the same thing: anchorKey uses `Text != ""` and sub() uses `word.MatchString`. Where they disagree, a declaration opens an id anchor yet lowers as an AnonSub, or the reverse.
    - Replace with: one `subName(n) (string, bool)` helper used by both.
    - Saves: about 3 lines, and removes the drift.
    - Confidence: 70.

12. **internal/lexer/scan.go:1109 vs lexer.go:590-591 (the same "term-position star" test written twice)**
    - Cut: the second copy of `l.expect != XOperator && l.expect != XPostDeref && l.attrArgDepth == 0`. It appears in both scanOperator and emit's globStar, and the two must stay in lockstep.
    - Replace with: one `starIsSigil()` method.
    - Saves: about 0 lines, but removes a lockstep hazard.
    - Confidence: 70.

13. **internal/lower/literals.go:101, call.go:212 (childless term matched against `word`, spelled out by hand)**
    - Cut: the hand-written checks.
    - Replace with: `bareTerm(n, word)`, which already exists (stmt.go:372).
    - Saves: 0 lines, but gives one spelling.
    - Confidence: 90.

14. **internal/perl/inc.go:43-47 (loop-invariant work inside the PERL5LIB loop)**
    - Cut: the `version, arch :=` lookup and the `version == "" || arch == ""` break, both done on every pass.
    - Replace with: one hoisted guard, `if version != "" && arch != ""`, around the subdirectory loop.
    - Saves: 1 line, and the code reads straight.
    - Confidence: 95.

15. **internal/lower/node.go:274-279 (MethodCall comment)**
    - "Indirect is kept: ... so it carries signal": nothing outside the lower tests reads Indirect. The comment argues for a field that is never used.
    - The field stays, because TestLowerMethodCall is criteria-named. Cut the comment, or let it name the reader that is planned.
    - Saves: 4 lines.
    - Confidence: 60.

#### Plain list

- node.go:384-393: Match.Pattern/Replace/Flags are never set. Delete them. About 6 lines. 85.
- special.go:68-74: operand duplicates nth. Delete it. 7 lines. 95.
- term.go:323-348: the `&` branch peeks three times and builds the same node twice. Collapse it. About 8 lines. 90.
- ops.go:316-328 and node.go:201-210: CmpChain.Class, TokenClass and eqOps are unread. Delete them. About 14 lines. 75.
- print.go:310 and node.go:487: two child walks. Derive Children from roles. About 120 lines. 55 (perf unmeasured).
- node.go:90-94: the Interp kind is never produced. Delete it and its infer arms. About 10 lines. 60.
- positions.go:128: the spans map duplicates cst. Delete it. About 4 lines. 80.
- lower.go:14-18: Lower re-implements block. Delete the loop. 4 lines. 85.
- ops.go:245-252: commaChain re-implements elementList. 2 lines. 90.
- special.go:129-135: Readline.Magic is unread. Delete it. About 4 lines. 65.
- id.go:99 and decl.go:330: two named-sub predicates. Unify them. About 3 lines. 70.
- scan.go:1109 and lexer.go:590: the star predicate is written twice. Extract it. 0 lines. 70.
- literals.go:101 and call.go:212: use bareTerm. 0 lines. 90.
- inc.go:43-47: hoist the invariant guard. 1 line. 95.
- node.go:274-279: the comment defends a field nothing reads. Delete the comment. 4 lines. 60.

## Steps

- [ ] One commit per cut; behaviour unchanged (the infer goldens, precision and the m1 gates hold); `make test`; commit

## Acceptance criteria

- [ ] the cuts are applied with behaviour unchanged (`go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$' -count=1 -v | grep -q '^--- PASS: TestLoweredGoldenRatchet'`)
