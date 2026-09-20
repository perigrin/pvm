---
title: Re-measure §4.14.2 against HEAD and rewrite the status table
state: pending
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c10f-9671-733f-8b98-30a78f634b0f
- 01a0c110-839f-77fb-8a3d-30fa5a8d6e96
- 01a0c111-8c3e-7451-97d9-10307441ba2d
- 01a0c112-8a65-7d5c-94ec-69859b2928d3
- 01a0c112-a8bb-7d7f-b156-d4ad99474e87
- 01a0c113-c3b2-75d8-ae0e-4f146d610c49
created: 2026-09-20T23:02:22.60535129Z
updated: 2026-09-20T23:12:32.663992571Z
---

## Context

- paths: `docs/specs/perl-parser/04-expressions.md` (§4.14.2, §4.14.3, §4.14.6), `internal/parse/parse.go`, `internal/parse/statustable_test.go` (new)
- commands: `go test ./internal/parse/ -run "^TestStatusTableAtHead$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md` §4.14.2; `internal/parse/parse.go:236-299`

§4.14.2 is the decomposition axis for the whole lowering (seam D slices by its
SHIPPED / LOWERING / PARSER status column). Its baseline is commit 3c997a6b and
HEAD is 39+ commits past it. The table is measurably STALE, in the direction of
LESS work: several rows marked PARSER now parse clean.

Verified at HEAD (b6e7c351) by parsing each fragment and dumping the tree:

	map { $_ } @a;          0 Unknowns; call(map) resolved=true with a block child
	sort { $a <=> $b } @a;  0 Unknowns
	grep { $_ } @a;         0 Unknowns
	delete $h{k};           0 Unknowns; call(delete) with an index child
	exists $h{k};           0 Unknowns
	sub { 1 };              0 Unknowns; declaration(sub) with a block child
	my @r = (1..5);         0 Unknowns; binary ".." -- the "Range infers Str" row is GONE
	$h->{k};                index "{" with Arrow SET
	(a => 1, b => 2);       list with Fat SET on the two barewords
	($x); ("a") x 3;        term with Paren SET
	print $fh "x";          term with Handle SET
	print {$fh} "x";        1 Unknown  -- STILL BLOCKED
	new Foo(1);             1 Unknown  -- STILL BLOCKED
	qx/ls/;                 parses as term "qx/ls/" -- splitting still PARSER
	&$code(1);              1 Unknown  -- STILL BLOCKED

Bigger than the row edits: **§4.14.3 asks for three bits on `parse.Node` --
`Paren`, `Arrow`, and the filehandle slot -- and all three SHIPPED, plus `Fat`.**
`parse.go:236-299` documents `Arrow`, `Paren`, `Fat` and `Handle` with the same
"meaningful on one kind, false elsewhere" contract as `Resolved`. Every §4.14.3
row that reads "PARSER" for one of those four is wrong at HEAD.

The grammar-name count is also contested: the parent brief says 34, §4.14 says
57, §6.1.4/§6.4.1 say 34. The count that matters is stated here so it stops
drifting: **distinct string literals appearing in a `case` arm of a switch on
`node.Kind()` across the four non-test files of `internal/infer`.** Measure it,
state the method inline in the spec next to the number, and reconcile the three
sites to that one figure. (A raw `case "..."` grep over the same four files
yields 83, which is an upper bound -- it includes switches on operator text and
on sigils, not just `Kind()`.)

This issue produces a spec update plus a REGRESSION TEST that fails the next
time the parser makes a row stale, so the table cannot silently rot again.

## Steps

- [ ] Write `internal/parse/statustable_test.go` with `TestStatusTableAtHead`: a
      table of (source fragment, expected Unknown count, expected flags set)
      covering every §4.14.2 row whose status this issue changes, plus the four
      rows that stay PARSER. Assert against `parse.Parse`.
- [ ] Run to verify failure: `go test ./internal/parse/ -run "^TestStatusTableAtHead$" -count=1 -v`
- [ ] Fix the EXPECTATIONS to the measured truth, not the parser -- this test
      pins reality, it does not change it. Re-run to green.
- [ ] Commit the test.
- [ ] Count the grammar names by the stated method over `internal/infer/infer.go`,
      `symbols.go`, `project.go`, `diagnostics.go`.
- [ ] Rewrite §4.14.2 status column at HEAD. Change the preamble baseline from
      3c997a6b to the HEAD sha. Every changed row cites the fragment and the
      measured tree.
- [ ] Rewrite §4.14.3: `Paren`, `Arrow`, `Handle`, `Fat` are SHIPPED; strike the
      "Three of these ask for a bit on `parse.Node`" paragraph, which is done.
- [ ] Reconcile the grammar-name count across §4.14, §6.1.4, §6.4.1 and the
      parent brief 01a0c0e7 to one number with the method stated.
- [ ] Run the full parse suite: `go test ./internal/parse/ -count=1`
- [ ] Commit the spec update.

## Acceptance Criteria

### Positive Scenarios
- [ ] The status table is pinned by an executable test that measures the real parser (`go test ./internal/parse/ -run "^TestStatusTableAtHead$" -count=1 -v | grep -q "^--- PASS: TestStatusTableAtHead"`)
- [ ] The parse package is green with the new test in it (`go test ./internal/parse/ -count=1`)
- [ ] §4.14.2 no longer claims its 3c997a6b baseline (`test -z "$(grep -c "against 3c997a6b" docs/specs/perl-parser/04-expressions.md | grep -v "^0$")"`)
