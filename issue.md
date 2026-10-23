---
title: Split the faithful comparator out, tag canon's parens by role, and enumerate the forgiveness rules, with no behaviour change
state: done
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5e5e-73cc-a210-0341ea517914
blocks:
- 01a1664d-75d4-76eb-aad2-10b322f366a9
- 01a1690b-4745-7ef7-9285-0a9f97d86924
- 01a1690b-1d10-7a84-be7e-0cbc31f36812
created: 2026-10-23T12:56:46.617002164Z
updated: 2026-10-23T19:54:59.82861176Z
sessions:
- start_sha: 5b88c65c7bdc2f2969d00476bef99270f1695b3c
  end_sha: 863b9896136941d0b13806459168c863fa094653
  commits: 3
  started_at: 2026-10-23T18:03:40.049352127Z
  ended_at: 2026-10-23T19:54:59.82861176Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-23T18:03:40.049352127Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-23T19:54:59.82861176Z
observed_paths:
- .github/workflows/parse-fidelity.yml
- internal/parse/canon.go
- internal/parse/export_test.go
- internal/parse/fidelity.go
- internal/parse/fidelity_test.go
- internal/parseoracle/ci_test.go
- internal/parseoracle/cmd/receipt/main.go
- internal/parseoracle/gate_test.go
- internal/parseoracle/opcount_test.go
- internal/parseoracle/ratchet_test.go
- internal/parseoracle/receipt.go
- internal/parseoracle/receipt_test.go
- internal/parseoracle/siteratchet.go
- internal/parseoracle/sites_test.go
- internal/parseoracle/testdata/opcount/README.md
- internal/parseoracle/testdata/opcount/opcount.md
- internal/parseoracle/testdata/ratchet/fixture-sites.txt
- internal/parseoracle/testdata/ratchet/m3start.txt
- internal/parseoracle/testdata/ratchet/sites.txt
---

**The title is stale and git-zhi cannot retitle.** Read it as "Split the faithful comparator out and enumerate the forgiveness rules, with no behaviour change". The role tags it names were dropped under perigrin's round-7 ruling: the call-paren forgiveness rule, their only consumer, is removed by the `not`/list-op issue, so canon records one bit per paren pair instead (below).

## Context

- paths: `internal/parse/canon.go` (records which `(`/`)` pairs it wrote from a node's `Paren` flag), `internal/parse/fidelity.go` (`faithfulTokens` split out of `Faithful`), `internal/parse/export_test.go` (exports `faithfulTokens` to `package parse_test`), `internal/parse/fidelity_test.go`, `internal/parse/testdata/t1fidelity.ratchet` and `internal/parse/testdata/perlgittfaithful.ratchet` (read; must not move)
- commands: `go test ./internal/parse/ -run '^(TestFaithfulTokens|TestCanonMarksParenNodePairs|TestFaithfulForgivenessIsClosed)$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (the strings.Builder emission), `internal/parse/fidelity.go` (`Faithful`, `isName`, `matchParen`, `matchForgivingSemicolons`, `significantTokens`)

Split out of the Paren-once issue (K5, chain-review round 6) so that issue is not the largest single-session risk on the critical chain. This issue makes NO behaviour change. Every faithful verdict and both ratchets stay exactly as they are. It lands BEFORE the named-unary canon issue (P30, H12, L13, M9, round 8): both edit canon.go, and this issue's edit is small (byte offsets at the paren emission sites), while that issue changes `emit`'s signature at every call site, so it rebases over this one and never the other way round.

**What it builds.**
- **The comparator.** `Faithful(n *Node, src []byte)` takes a tree, so a test that says "source A against canon B is unfaithful" needs a comparator over token sequences. This issue splits one out of `Faithful`, with this signature (K5):

  `faithfulTokens(src []string, canon []CanonToken) (ok bool, forgiven []Forgiveness, reason string)`

  Here `Forgiveness{CanonIndex int, SrcIndex int, Rule string}` records each forgiven difference and the rule that forgave it. `Faithful` calls `faithfulTokens` and keeps its current signature. `faithfulTokens` is unexported; fidelity_test.go is `package parse_test`, so `internal/parse/export_test.go` gains `var FaithfulTokens = faithfulTokens` beside its existing exports (L8, chain-review round 7).
- **One bit per paren pair, not four role tags (round-7 ruling).** `Canon` writes into a strings.Builder and `Faithful` re-tokenizes the string, so canon records, by byte offset, each `(`/`)` pair it wrote because the node's `Paren` flag is set, and `CanonToken{Text string; ParenNode bool}` carries that after re-tokenizing. Today (measured at 054e75e6, M3) canon writes such a pair on two paths: the `List` case writes its own `(` and `)`; a `Term` with `Paren` is written as its SOURCE TEXT (`writesParens`, canon.go:~709, `n.Kind == Term`), which reproduces the source's nesting, so `my $z = (($x))` re-emits `(($x))` and `($x) x 3` re-emits `($x) x 3`. For the Term path the pair marked ParenNode is the outermost pair of that source text; any inner pair is unmarked. Every other `Paren` node (the CmpChain in `ok(($y eq "x"))`, a Binary) gets no pair at all today; those are the Paren-once issue's. Nothing consumes the bit in this issue. The Paren-once issue's extra-nesting rule does: it forgives a source pair nested directly inside a `ParenNode` pair, and nothing else. The round-6 tags `call-args`, `operand` and `binding` had one consumer, the call-paren rule, which the `not`/list-op issue removes under perigrin's round-7 ruling, so they are not built. That also retires K9 (a pair matching two roles) and L7 (the `my($x)` → `operand` fixture): a pair written from `Paren` is a `ParenNode` pair wherever it sits, and every other pair is not.
- **The closed-set test.** TestFaithfulForgivenessIsClosed enumerates the rules in force at its commit, each with a perl-measured fixture, and fails on any other rule name. At this commit those are `call-parens` and `last-semicolon`. The `not`/list-op issue removes `call-parens` from the code and the enumeration; the Paren-once issue adds `extra-nesting` and the extent negatives. The end state is perigrin's round-7 set: exactly `last-semicolon` and `extra-nesting`. The test also pins `significantTokens`' dropped token kinds.

**What fidelity.go forgives today (read 2026-10-23 at 054e75e6).** It has exactly the two rules its comment names:
1. call parens, keyed on `(` right after a name in the emission (`isName`), with the arguments matched unparenthesised;
2. a block's last `;`, before the emission's `}`.

Rule 2 is also applied inside a forgiven call paren (`matchForgivingSemicolons`), for a block argument such as `print({$fh;} "x")`. That is the same rule one level down, not a third. Separately, `significantTokens` drops Whitespace, Comment, Pod and DataSection tokens before comparing. That is token selection, not a forgiven difference, but a skip added there would bypass the rule list. So TestFaithfulForgivenessIsClosed also pins the dropped token kinds to exactly those four.

## Steps

- [ ] RED: write TestFaithfulTokens over a table of (source tokens, canon tokens, verdict, forgiven rules), copied from today's Faithful behaviour: the T1 and perl.git t/ examples in fidelity_test.go plus `print "x"` / `print("x")` and `sub f { 1 }` / `sub f { 1; }`. It calls `parse.FaithfulTokens` from export_test.go. Run it and watch it fail (no `faithfulTokens` yet).
- [ ] GREEN: split `faithfulTokens` out with the signature above, returning each forgiven difference's rule, and add the export. Run it and TestCanonFidelityRatchet, and watch both pass, with the ratchet unmoved.
- [ ] RED/GREEN: write TestCanonMarksParenNodePairs. For a fixture set, every `(`/`)` pair canon wrote from a node's `Paren` flag carries `ParenNode`, and no other pair does: `my ($x) = @_` (the List's pair, marked), `($x) x 3` (the Term's source-text pair, marked), `my $z = (($x))` (the outer source-text pair marked, the inner not, as canon writes both today), `print("x")` (not marked), `(1 + 2) * 3` (binding power, not marked). The `ok(($y eq "x"))` fixture belongs to the Paren-once issue, because canon writes no inner pair for it today (M3).
- [ ] RED/GREEN: write TestFaithfulForgivenessIsClosed with the two existing rules' fixtures. Measure each with the pinned perl 5.44.0: `print "x"` and `print("x")` print the same; `perl -e 'sub f { 1 } sub g { 1; } print f(), g()'` prints `11`. Pin the dropped token kinds to Whitespace, Comment, Pod and DataSection. Add TestFaithfulForgivenessIsClosedRejectsUnlistedRule: a fixture comparator reporting an unlisted rule name is flagged.
- [ ] Run TestPerlGitTFaithfulRatchet and TestCanonFidelityRatchet. Neither may move; if one does, the split changed behaviour, so fix it.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer checks that no verdict moved.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The token-sequence comparator reproduces Faithful's verdicts and names the rule behind each forgiven difference (`go test ./internal/parse/ -run '^TestFaithfulTokens$' -count=1 -v | grep -q '^--- PASS: TestFaithfulTokens'`)
- [ ] Every pair canon writes from a node's Paren flag, on the List path or the Term source-text path, is marked ParenNode, and no other pair is (`go test ./internal/parse/ -run '^TestCanonMarksParenNodePairs$' -count=1 -v | grep -q '^--- PASS: TestCanonMarksParenNodePairs'`)
- [ ] The forgiven set is enumerated, each rule with a perl-measured fixture, and the dropped token kinds are pinned (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosed'`)
- [ ] The comparator is reachable from the external test package (`grep -q '^var FaithfulTokens = faithfulTokens' internal/parse/export_test.go`)
- [ ] Regression guard: the T1 fidelity ratchet does not move (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
- [ ] Regression guard: the perl.git t/ faithful ratchet does not move (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)

### Negative Scenarios

- [ ] A comparator that reports a rule outside the enumerated set fails the closed-set test (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosedRejectsUnlistedRule$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosedRejectsUnlistedRule'`)
