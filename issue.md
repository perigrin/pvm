---
title: "A handle inside print's parens refuses: print($fh \"x\") is canon's own spelling and does not re-parse"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T03:26:39.956195905Z
updated: 2026-09-29T03:26:39.956195905Z
---

A filehandle inside print's parens refuses unless it is the `{$fh}` block
form, and canon writes a handle as `print($fh "x")` -- so canon's own
emission does not re-parse:

    print($fh "x");        Unknown=1
    print(STDERR "y");     Unknown=1

perl reads both as handles, by the same no-comma rule as without parens.
Measured on 5.42.0 with -MO=Deparse:

    print($fh "x");      print $fh 'x';
    print(STDERR "y");   print STDERR 'y';
    print($a, "z");      print $a, 'z';

The parenthesised path called parseFilehandleBlock, whose comment held that
inside parens a scalar or bareword first is an ordinary argument. perl
disagrees; the parenless path's parseFilehandleSlot already reads all three
spellings with the comma test.

Found when the lexer's filehandle-slot heuristic (for `print $fh <<"EOF"`)
made T1's bare_block_return.t canon `print($fh <<'EOF')`, which then re-parsed
with the heredoc body attached to the NEXT statement.

## Acceptance criteria

- [ ] A scalar, bareword or block handle inside print's parens is a handle when no comma follows, and canons to a fixpoint (`go test ./internal/parse/ -run '^TestParenthesisedFilehandle$' -count=1 -v | grep -q '^--- PASS: TestParenthesisedFilehandle'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
