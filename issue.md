---
title: Re-measure §4.14.2 against HEAD and rewrite the status table
state: done
urgency: normal
milestone: m2-lower-the-tree
blocks:
- 01a0c10f-37f8-7a43-b3ce-ea194f7defe0
created: 2026-09-20T23:02:22.60535129Z
updated: 2026-10-11T15:23:28.754402863Z
sessions:
- start_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  end_sha: 1a558ba8c76f9795fdb80f2f14ad28ecca94aee4
  commits: 1
  started_at: 2026-10-11T10:36:37.013354411Z
  ended_at: 2026-10-11T15:23:28.754402863Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T10:36:37.013354411Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-11T15:23:28.754402863Z
observed_paths:
- docs/specs/perl-parser/04-expressions.md
- docs/specs/perl-parser/06-incremental-lsp.md
- internal/lower/id.go
- internal/lower/id_test.go
- internal/parse/statustable_test.go
---

## Context

- paths: `docs/specs/perl-parser/04-expressions.md` (§4.14 preamble, §4.14.2, §4.14.3, §4.14.6), `internal/parse/parse.go`, `internal/parse/statustable_test.go` (new)
- commands: `go test ./internal/parse/ -run "^TestStatusTableAtHead$" -count=1 -v`
- entrypoints: `docs/specs/perl-parser/04-expressions.md:1428` (the "at 3c997a6b" preamble) and the six `| Kind | CST at 3c997a6b |` table headers at `:1773-1830`; the `parse.Node` flags at `internal/parse/parse.go:308-487`

§4.14.2 is the decomposition axis for the whole lowering: each lowering slice
takes its rows and statuses from it. Its baseline is 3c997a6b, and the table
is stale in the direction of LESS work.

This issue now blocks only the vocabulary issue (01a0c10f-37f8), whose
§4.14.6 decisions rest on these facts. Every lowering slice depends on the
vocabulary issue, so each still runs after this one.

Measured at 24bfac4a by parsing each fragment with `parse.Parse` (a throwaway
test run through `go test -overlay`, nothing committed). This list supersedes
the one measured at b6e7c351. Every shape below was re-checked at e18b59c8
(after the code-deref, builtin-bundle, apostrophe and bare-& merges) and is
unchanged:

	map { $_ } @a;            0 Unknown; call "map" resolved, block child
	sort { $a <=> $b } @a;    0 Unknown
	grep { $_ } @a;           0 Unknown
	delete $h{k};             0 Unknown; call "delete" over index "{"
	sub { 1 };                0 Unknown; declaration "sub" with a block child
	my @r = (1..5);           0 Unknown; binary ".." paren over term "1", term "5"
	$h->{k};                  index "{" arrow; the key is term "k"
	(a => 1, b => 2);         list paren; the barewords are call "a" fat, call "b" fat
	("a") x 3;                binary "x" over term "\"a\"" paren
	print $fh "x";            call "print", term "$fh" handle
	print {$fh} "x";          0 Unknown NOW: call "print", block handle over the statement $fh
	&$code(1);                0 Unknown NOW: index "(" over term "&$code", term "1"
	new Foo(1);               1 Unknown (trailing_tokens) when Foo is undeclared;
	                          call "new" resolved indirect over term "Foo", term "1"
	                          after "package Foo; sub new {1} package main;"
	do 'file.pl';             0 Unknown NOW: call "do" resolved over a term
	<*.c>;  <<>>;  $::x;      0 Unknown NOW: one term each
	goto &foo;                0 Unknown NOW: loop_control "goto" over term "&foo"
	{a=>1};                   0 Unknown NOW: anon_hash at statement start
	${$x}; $$x; @$x; %$x;     0 Unknown NOW: unary "$" / "@" / "%" over term "$x"
	$#{$x}; $#$x;             still one childless term each
	sub f :prototype($) {1}   0 Unknown NOW: an attribute child ":prototype($)"
	my $x :shared;            attribute ":shared", not a term
	qx/ls/;                   one term "qx/ls/"; splitting is still PARSER
	f(1,2);                   call "f" over ONE binary "," child, not two arguments

Four facts the table does not yet carry:

- **§4.14.3's bits all shipped, and more.** `Arrow` (`parse.go:330`), `Paren`
  (`:341`), `Fat` (`:350`), `Handle` (`:359`), plus `Comparator` (`:366`) and
  `Indirect` (`:373`). Every §4.14.3 row that reads PARSER for one of these
  is wrong.
- **`Indirect` carries signal.** §4.14.6 said to drop `MethodCall.Indirect`
  if it was always false. It is no longer always false: the parser sets it
  when the class is a package the file has seen. The vocabulary issue
  decides from this.
- **An Unknown can have children.** The census says every T1 Unknown is a
  childless leaf. At 24bfac4a, `if ($x) {` gives
  `[unknown refusal=unclosed_opener [block]]`. Record the count of Unknowns
  with children over T1 next to the census.
- **Call arguments are one comma tree.** `f(1,2)` and `print $fh, "x"` carry
  their arguments as one `binary ","` child. The calls slice flattens it.

The grammar-name count is contested: the parent brief says 34, §4.14 says
57, and §6.1.4 and §6.4.1 say 34. Fix the method so it stops drifting:
**distinct string literals in a `case` arm of a switch on `node.Kind()`
across the four non-test files of `internal/infer`.** Measure it, state the
method next to the number, and reconcile all three sites. A raw `case "..."`
grep over `infer.go`, `symbols.go`, `project.go` and `diagnostics.go` gives
an upper bound: 65 distinct strings at 24bfac4a, 56 of them lower-case
word names. It includes switches on operator text and sigils.

## Steps

- [ ] Write `internal/parse/statustable_test.go` with `TestStatusTableAtHead`:
      a table of (fragment, expected Unknown count, expected flags) covering
      every row whose status this issue changes. Re-measure the list above
      at execution time; do not copy it.
- [ ] Run it and confirm it fails:
      `go test ./internal/parse/ -run "^TestStatusTableAtHead$" -count=1 -v`
- [ ] Set the expectations to the measured truth, not the parser. The
      test pins reality; it does not change it. Re-run to green. Commit.
      (perigrin, 2026-10-11: no second test pins the rows that still
      refuse. Such a test would turn red whenever another milestone's
      parser work closes a row, tying m2's final verify to that progress.)
- [ ] Count the grammar names by the stated method.
- [ ] Rewrite §4.14.2's status column at HEAD. Change the preamble at
      `:1428` and each table header from `CST at 3c997a6b` to the HEAD sha.
      Every changed row cites its fragment and the measured tree.
- [ ] Rewrite §4.14.3: all six bits are SHIPPED. Strike the "Three of these
      ask for a bit on parse.Node" paragraph.
- [ ] Add the Unknown-with-children count to the Hole row, and the comma
      tree to the Call row.
- [ ] Reconcile the grammar-name count across §4.14, §6.1.4, §6.4.1 and the
      parent brief 01a0c0e7, with the method stated.
- [ ] Run the parse suite: `go test ./internal/parse/ -count=1`
- [ ] Commit the spec update.

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance Criteria

### Positive Scenarios
- [ ] The status table is pinned by a test that measures the real parser (`go test ./internal/parse/ -run '^TestStatusTableAtHead$' -count=1 -v | grep -q '^--- PASS: TestStatusTableAtHead'`)

### Negative Scenarios
- [ ] No §4.14.2 table still claims the stale 3c997a6b baseline, and all six are re-measured at a named sha (`! grep -q 'CST at 3c997a6b' docs/specs/perl-parser/04-expressions.md && test "$(grep -c '^| Kind | CST at [0-9a-f]\{8\}' docs/specs/perl-parser/04-expressions.md)" -eq 6`)
