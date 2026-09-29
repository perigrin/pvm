---
title: "system { \"ls\" } LIST orphans its list: system and exec have print's indirect-object slot"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T04:34:19.05212845Z
updated: 2026-09-29T04:40:09.86324255Z
sessions:
- start_sha: 01a2c320ae90cbb6b06d417640bfc830f33b0d91
  end_sha: 2ff17048eb0761ee20df9012bdcda59e36ef6136
  commits: 1
  started_at: 2026-09-29T04:34:20.566784227Z
  ended_at: 2026-09-29T04:40:09.86324255Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T04:34:20.566784227Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T04:40:09.86324255Z
observed_paths:
- internal/lexer/keyword.go
- internal/parse/call.go
- internal/parse/indirectprog_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1canon.ratchet
---

`system` and `exec` take their program in an indirect-object slot, the same
slot print has for a handle, and the parser gave that slot only to print,
printf and say. So the block became the call's only argument and the list a
statement of its own, at Unknown=0:

    my $rc = system { "ls" } "ls", "-l";    canon  my $rc = system({"ls"});"ls" , "-l";
    exec { $p } @args;                      Unknown=1

Measured on 5.42.0 with -MO=Deparse:

    system { "ls" } "ls", "-l"      system({'ls';} 'ls', '-l')
    exec { $p } @args;              exec {$p;} @args;
    system $shell "-sh";            system $shell '-sh';

Both are builtins, so the lexer's blockTaking table -- which draws the line at
builtins because an arbitrary word's parenthesised block is a syntax error --
takes them too.

Reach: op/exec.t and op/taint.t in perl.git t/; three T1 files.

## Acceptance criteria

- [ ] system and exec read a block, scalar or bareword program slot with no comma after it, and canon to a fixpoint (`go test ./internal/parse/ -run '^TestIndirectProgramSlot$' -count=1 -v | grep -q '^--- PASS: TestIndirectProgramSlot'`)
- [ ] REGRESSION GUARD: a plain `system LIST` is unchanged (`go test ./internal/parse/ -run '^TestIndirectProgramSlot$' -count=1 -v | grep -q '^--- PASS: TestIndirectProgramSlot'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
