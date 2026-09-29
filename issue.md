---
title: "An empty slot inside brackets refuses: `[\"x\",, \"y\"]`"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T15:23:27.484454088Z
updated: 2026-09-29T15:34:40.931504489Z
sessions:
- start_sha: 475b9fc5a315964dff925731d2cb7cf34f2421f7
  end_sha: 59c71e8af0366ea66f124dc3c4c574467e1e02b8
  commits: 1
  started_at: 2026-09-29T15:23:29.258229082Z
  ended_at: 2026-09-29T15:34:40.931504489Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T15:23:29.258229082Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T15:34:40.931504489Z
observed_paths:
- internal/parse/bracketslot_test.go
- internal/parse/term.go
- internal/parse/testdata/perlgitt.ratchet
---

An empty slot inside `[...]` or `{...}` refuses: `["x",, "y"]`. parseParenList drops a separator with no element after it, and parseBracketed does not.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my $a = ["x",, "y"]; my $h = {a => 1,, b => 2}; my $c = [1, => 2];'
    my $a = ['x', 'y'];
    my $h = {'a', 1, 'b', 2};
    my $c = [1, 2];

perl.git t/porting/bench.t:274-282, `"--bisect=Ir",,`.

## Acceptance criteria

- [ ] An empty slot in brackets is dropped, and canon is a fixpoint (`go test ./internal/parse/ -run '^TestBracketEmptySlot$' -count=1 -v | grep -q '^--- PASS: TestBracketEmptySlot'`)
