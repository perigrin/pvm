---
title: Split the faithful comparator out, tag canon's parens by role, and enumerate the forgiveness rules, with no behaviour change
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5e5e-73cc-a210-0341ea517914
created: 2026-10-23T12:56:46.617002164Z
updated: 2026-10-23T12:56:49.329772351Z
---

## Context

- paths: `internal/parse/canon.go` (role tags on every paren it writes), `internal/parse/fidelity.go` (`faithfulTokens` split out of `Faithful`), `internal/parse/fidelity_test.go`, `internal/parse/testdata/t1fidelity.ratchet` and `internal/parse/testdata/perlgittfaithful.ratchet` (read; must not move)
- commands: `go test ./internal/parse/ -run '^(TestFaithfulTokens|TestCanonTagsEveryParen|TestFaithfulForgivenessIsClosed)$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (the strings.Builder emission), `internal/parse/fidelity.go` (`Faithful`, `isName`, `matchParen`, `matchForgivingSemicolons`, `significantTokens`)

Split out of the Paren-once issue (K5, chain-review round 6) so that issue is not the largest single-session risk on the critical chain. This issue makes NO behaviour change. Every faithful verdict and both ratchets stay exactly as they are; the Paren-once issue then changes behaviour on top of it.

**What it builds.**
- **The comparator.** `Faithful(n *Node, src []byte)` takes a tree, so a test that says "source A against canon B is unfaithful" needs a comparator over token sequences. This issue splits one out of `Faithful`, with this signature (K5):

  `faithfulTokens(canon []CanonToken, src []string) (ok bool, forgiven []Forgiveness, reason string)`

  Here `Forgiveness{CanonIndex int, SrcIndex int, Rule string}` records each forgiven difference and the rule that forgave it. `Faithful` calls `faithfulTokens` and keeps its current signature.
- **Role tags.** `Canon` writes into a strings.Builder, and `Faithful` re-tokenizes the string. So canon records a role for every `(`/`)` it writes, by byte offset, and `CanonToken` carries it after re-tokenizing:
  - `paren-node`: a node with `Paren` set;
  - `call-args`: the argument list of a list operator, or of a user or method call;
  - `operand`: the operand of a declarator (`my`/`our`/`local`/`state`), a named unary op, `not`, or a word infix op such as `x`;
  - `binding`: required by binding power.

  In this issue the call-paren rule still keys on `isName`, so behaviour does not change. The Paren-once issue narrows it to `call-args`, per perigrin's round-6 ruling.
- **The closed-set test.** TestFaithfulForgivenessIsClosed enumerates the rules that exist today, `call-parens` and `last-semicolon`, each with a perl-measured fixture. It fails on any other rule name, and it pins `significantTokens`' dropped token kinds. The Paren-once issue adds the third rule, `extra-nesting`, and the extent negatives.

**What fidelity.go forgives today (read 2026-10-23 at 054e75e6).** It has exactly the two rules its comment names:
1. call parens, keyed on `(` right after a name in the emission, with the arguments matched unparenthesised;
2. a block's last `;`, before the emission's `}`.

Rule 2 is also applied inside a forgiven call paren (`matchForgivingSemicolons`), for a block argument such as `print({$fh;} "x")`. That is the same rule one level down, not a fourth. Separately, `significantTokens` drops Whitespace, Comment, Pod and DataSection tokens before comparing. That is token selection, not a forgiven difference, but a skip added there would bypass the rule list. So TestFaithfulForgivenessIsClosed also pins the dropped token kinds to exactly those four.

## Steps

- [ ] RED: write TestFaithfulTokens over a table of (canon tokens, source tokens, verdict, forgiven rules), copied from today's Faithful behaviour: the T1 and perl.git t/ examples in fidelity_test.go plus `print "x"` / `print("x")` and `sub f { 1 }` / `sub f { 1; }`. Run it and watch it fail (no `faithfulTokens` yet).
- [ ] GREEN: split `faithfulTokens` out with the signature above, returning each forgiven difference's rule. Run it and TestCanonFidelityRatchet, and watch both pass, with the ratchet unmoved.
- [ ] RED/GREEN: write TestCanonTagsEveryParen. Every `(` and `)` canon writes for a fixture set carries exactly one role tag, and the tags for `print("x")`, `my($x)`, `defined($x)`, `($x)` with Paren, and `(1 + 2) * 3` are `call-args`, `operand`, `operand`, `paren-node` and `binding`.
- [ ] RED/GREEN: write TestFaithfulForgivenessIsClosed with the two existing rules' fixtures. Measure each with the pinned perl 5.44.0: `print "x"` and `print("x")` print the same; `perl -e 'sub f { 1 } sub g { 1; } print f(), g()'` prints `11`. Pin the dropped token kinds to Whitespace, Comment, Pod and DataSection. Add TestFaithfulForgivenessIsClosedRejectsUnlistedRule: a fixture comparator reporting an unlisted rule name is flagged.
- [ ] Run TestPerlGitTFaithfulRatchet and TestCanonFidelityRatchet. Neither may move; if one does, the split changed behaviour, so fix it.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer checks that no verdict moved.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The token-sequence comparator reproduces Faithful's verdicts and names the rule behind each forgiven difference (`go test ./internal/parse/ -run '^TestFaithfulTokens$' -count=1 -v | grep -q '^--- PASS: TestFaithfulTokens'`)
- [ ] Every paren canon writes carries one role tag (`go test ./internal/parse/ -run '^TestCanonTagsEveryParen$' -count=1 -v | grep -q '^--- PASS: TestCanonTagsEveryParen'`)
- [ ] The forgiven set is enumerated, each rule with a perl-measured fixture, and the dropped token kinds are pinned (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosed'`)
- [ ] Regression guard: the T1 fidelity ratchet does not move (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
- [ ] Regression guard: the perl.git t/ faithful ratchet does not move (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)

### Negative Scenarios

- [ ] A comparator that reports a rule outside the enumerated set fails the closed-set test (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosedRejectsUnlistedRule$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosedRejectsUnlistedRule'`)
