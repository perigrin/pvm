---
title: "system { \"ls\" } LIST orphans its list: system and exec have print's indirect-object slot"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T04:34:19.05212845Z
updated: 2026-09-29T04:34:19.05212845Z
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
