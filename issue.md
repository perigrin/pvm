---
title: "print $fh <<'EOF' lexes << as a shift: the filehandle slot never leaves operator position"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T03:39:15.639427053Z
updated: 2026-09-29T03:44:35.994027606Z
sessions:
- start_sha: dc1cf141bf6b22beb8acb2b3d552f14ed5840878
  end_sha: 6f2cd1d339c41288f07145a861f08949e579b862
  commits: 1
  started_at: 2026-09-29T03:39:16.949676433Z
  ended_at: 2026-09-29T03:44:35.994027606Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T03:39:16.949676433Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T03:44:35.994027606Z
observed_paths:
- internal/lexer/fhslot_test.go
- internal/lexer/lexer.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
---

`print $fh <<'EOF'` lexes the `<<` as a left shift, so the heredoc body is read
as code and the statement after it refuses:

    print $f <<'EOT';
    ...body...
    EOT
    close $f;                 refuses: the body swallowed it

After a scalar the lexer expects an operator. perl resolves the filehandle slot
in toke.c's yyl_dollar: a scalar straight after a list operator, followed by
whitespace, is a handle when the next character can only start a term -- a
quote, a digit, a sigil and name, a word other than a binary operator, `-1`,
`/re/`, or `<<` followed by something other than space or `=`. It applies after
`print(` too. Measured on 5.42.0 with -MO=Deparse:

    print $f <<"EOT"; ...    print $f "x\n";
    print($f <<"EOT"); ...   print $f "x\n";
    print $f /2/;            print $f /2/;
    print $x - 1;            print $x - 1;
    print $x == 1;           print $x == 1;

perl tests PL_last_lop, so the rule holds after ANY list operator, a user's
included. The lexer cannot know a user's list operators, so this applies after
the builtins whose slot this is: print, printf, say, exec, system.

Reach: 7 perl.git t/ files print a heredoc to a scalar handle; 5 of the dirty
ones first refused at the statement after it (op/caller.t, op/goto.t,
run/switchd.t, run/switches.t, win32/system.t).

## Acceptance criteria

- [ ] After print/printf/say/exec/system, a handle scalar followed by a term-starting character leaves term position, with or without the call's parens (`go test ./internal/lexer/ -run '^TestFilehandleSlotLeavesTermPosition$' -count=1 -v | grep -q '^--- PASS: TestFilehandleSlotLeavesTermPosition'`)
- [ ] NEGATIVE: `print $x - 1`, `print $x == 1` and `$y = $x <<2` keep their operators (`go test ./internal/lexer/ -run '^TestFilehandleSlotLeavesTermPosition$' -count=1 -v | grep -q '^--- PASS: TestFilehandleSlotLeavesTermPosition'`)
- [ ] REGRESSION GUARD: the lexer's golden token stream is unchanged (`go test ./internal/lexer/ -count=1`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
