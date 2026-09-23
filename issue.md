---
title: Six keywords that silently reparse when their feature is off
state: in-progress
urgency: normal
milestone: corpus-complete
created: 2026-09-23T02:07:02.559443716Z
updated: 2026-09-23T17:08:38.870184806Z
sessions:
- start_sha: dba3a2a8f33826ff834672eb4a84f0faf854bc36
  end_sha: 1afd0f341b23e124940e7ed14440f2338370dd79
  commits: 1
  started_at: 2026-09-23T16:33:23.880128542Z
  ended_at: 2026-09-23T16:45:28.172532695Z
- start_sha: df9f271d0a74bb508132371df7b7f0d40a632f0b
  end_sha: ""
  commits: 0
  started_at: 2026-09-23T17:08:38.870184806Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T16:33:23.880128542Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-23T16:45:28.172532695Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-09-23T17:08:38.696523757Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T17:08:38.870184806Z
observed_paths:
- conformance/11_oo/14_state_ungated.t
- conformance/11_oo/15_class_ungated.t
- conformance/11_oo/16_classname_ungated.t
- conformance/11_oo/17_defer_ungated.t
- internal/conformance/testdata/corpus.ratchet
- internal/conformance/tier11_test.go
---

Every version-gated feature in the corpus is tested ONLY in its enabled
state. Every such file turns the feature on at the top. Not one pins what
the same bytes do with it off.

That is half of what "version-gated" means, and it is the half where both
of the corpus's previously-found bugs lived: `say` without its feature is
a method call, `isa` without its feature is a filehandle print. Both were
found by asking what happens with the feature OFF.

## Six keywords that silently reparse, measured under 5.42.0

    $ perl -MO=Deparse -e 'sub c { state $n = 0; return ++$n }'
    -e syntax OK
    sub c { $n->state = 0; return ++$n; }

    $ perl -MO=Deparse -e 'class Foo { }'
    'Foo'->class({});

    $ perl -MO=Deparse -e 'my $x; field $x;'
    $x->field;

    $ perl -MO=Deparse -e 'print __CLASS__;'
    print __CLASS__ $_;

    $ perl -e 'sub f { defer { print "D\n" } print "body\n" } f();'
    D
    body                          <- and dies after running the block

    $ perl -ce 'sub f($x,$y) { 1 }'
    -e syntax OK                  <- then "Malformed prototype" at runtime

None is a syntax error. Each compiles clean and means something else.

`__CLASS__` is the `isa` finding in a different keyword -- it becomes a
FILEHANDLE. `defer` is worse than any of them: not a different structure
but an INVERTED EXECUTION ORDER, failing at runtime after the block has
already run.

## What catches these

`parses` plus `output`. Both readings parse, so `parses` alone says
nothing; what separates them is that one dies or prints in a different
order. `defer` is pinned by ordering, `state` by the runtime death,
signatures by the prototype warning.

These are not token-fact files. The bytes are identical by construction
-- that is the point -- so no token claim can differ between the pair.

## Files

Six pairs' worth, but written as six files rather than twelve: each
existing enabled file already pins one half, so the new file is the
DISABLED half and cites its partner. `05_scoping/03_state.t`,
`11_oo/07_class_empty.t`, `11_oo/08_class_field_method.t`,
`10_io/07_say.t` and `11_oo/10_isa_infix.t` are the existing halves;
`defer`, `__CLASS__` and signatures need both halves written.

Placement follows the enabled half's tier, so the pair sits together.

## The one that cannot be written yet

`class Foo { }` EMPTY reparses silently, but `class Foo { field $x;
method m {...} }` is a hard syntax error without the feature. So the
disabled half is writable only for the empty form -- which is exactly
what `11_oo/07_class_empty.t` holds. Worth recording in the file rather
than discovering twice.

## Acceptance criteria

- [ ] A corpus file pins what state does with its feature disabled -- at tier 11, since the ungated reading needs method dispatch (`awk '/^--- source/{s=1;next} /^--- expect/{s=0} s' conformance/11_oo/*.t | grep -q 'state \$n'`)
- [ ] The featureless defer file exists and its refusal is recorded rather than described (`go test ./internal/conformance/ -run '^TestCorpus/11_oo' -count=1 -v | grep -q '^    --- SKIP: TestCorpus/11_oo/17_defer_ungated'`)
- [ ] Every new file is validated against pinned perl 5.42.0, which is what makes a gate pair a measurement (`go test ./internal/conformance/ -run '^TestCorpusLints$' -count=1 -v | grep -q '^--- PASS: TestCorpusLints'`)
- [ ] Each new file's refusal, if it has one, cites a code rather than a message (`go test ./internal/conformance/ -run '^TestCorpusSkipsAreDocumented$' -count=1 -v | grep -q '^--- PASS: TestCorpusSkipsAreDocumented'`)
- [ ] The whole conformance suite is green with the new files in it (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
