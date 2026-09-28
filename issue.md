---
title: "Canon drops the parens around a declaration: (my $s = \"abc\") =~ /x/ re-emits as a different program"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T22:31:49.085135865Z
updated: 2026-09-28T22:32:03.158875841Z
---

Canon drops the parentheses around a declaration, at Unknown=0, and emits a
different program:

    () = (my $str = "abc") =~ /(...)/;    ->  () = my $str = "abc" =~ /(...)/;
    (my $str) =~ /x/;                     ->  my $str =~ /x/;

The first is WRONG: without the parens the match binds to "abc" before the
assignment, so `$str` gets the match result instead of "abc". perl's own
re/pat_advanced.t line 23 is this statement. The tree may already hold the
paren (List and Term carry `Paren`); the Declaration arm in canon.go does not
consult it.

Found while bisecting re/pat_advanced.t for the nested-block issue, 01a0ea25-4fe7-76cc-b87a-f67a331a5ef7. An
Unknown count cannot see it; only canon can.

## Acceptance criteria

- [ ] A parenthesised declaration keeps its parens through canon, and the canon re-parses to itself (`go test ./internal/parse/ -run '^TestParenthesisedDeclarationCanon$' -count=1 -v | grep -q '^--- PASS: TestParenthesisedDeclarationCanon'`)
- [ ] REGRESSION GUARD: the canon ratchet holds (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
