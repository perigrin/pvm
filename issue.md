---
title: "Heredocs: the pending queue and six terminator forms"
state: in-progress
urgency: normal
milestone: m0-lexer-round-trip
created: 2026-09-14T22:45:28.161521094Z
updated: 2026-09-15T05:53:11.3434952Z
sessions:
- start_sha: 7fc3a20a3439e693a8bcf59134ea0ea3d15cfa50
  end_sha: ""
  commits: 0
  started_at: 2026-09-15T05:53:11.3434952Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-15T05:53:11.3434952Z
---

Heredocs, after the expect-state issue because perl gates their recognition
on it. `toke.c:7173-7177`:

    if (PL_expect != XOPERATOR) {
        if (s[1] != '<' && !memchr(s,'>', PL_bufend - s))
            check_unary();
        if (s[1] == '<' && s[2] != '>')
            s = scan_heredoc(s);

`<<` in operator position is a left shift. In term position it opens a
heredoc. The same two bytes, and only the expect state separates them — which
is why an earlier draft of this chain had heredocs in the quote-scanner
issue and created a dependency cycle.

## The pending queue

The terminator is read at the `<<`, but the body starts after the current
LINE ends. So the lexer carries a queue: on reaching the newline, consume
each pending heredoc's body in the order they were opened.

`t/base/lex.t:56` opens TWO on one line. Spec §2.9.5 and §2.15.1.

Worse, `t/base/lex.t:114`:

    $foo =~ s/^not /substr(<<EOF, 0, 0)/e;
      Ignored
    EOF

A heredoc opened inside a substitution's replacement code, with the body on
the lines that follow. That case belongs to the sublexing issue, but it is
the reason the queue cannot be local to one scan.

## The six forms

Spec §2.9.3 lists six, not four — an earlier draft named four:

    <<EOF        interpolating
    <<"EOF"      interpolating, explicit; note <<"EOF" and << "EOF" differ
    <<'EOF'      non-interpolating
    <<\EOF       non-interpolating, backslash-quoted
    <<`EOF`      command
    <<~EOT       indentation-stripping (any of the above may be tilde'd)

`t/base/lex.t` uses three: `<<'EOF'` (36), bare `<<EOF` (41), and `<<\EOE`
(45), the last mid-expression: `eval <<\EOE, print $@;`. It contains no
`<<~` at all, so that form needs its own fixture rather than a claim about
lex.t.

The whitespace asymmetry is real and testable: `<<"EOF"` is a heredoc,
`<< "EOF"` is a left shift against a string.

## Acceptance

- [ ] `<<` in operator position is a left shift, in term position a heredoc (`go test ./internal/lexer/ -run '^TestHeredocVersusLeftShift$' -count=1 -v | grep -q '^--- PASS: TestHeredocVersusLeftShift'`)
- [ ] All six §2.9.3 forms lex, including `<<~` from its own fixture (`go test ./internal/lexer/ -run '^TestHeredocSixForms$' -count=1 -v | grep -q '^--- PASS: TestHeredocSixForms'`)
- [ ] Two heredocs opened on one line take their bodies in order (`go test ./internal/lexer/ -run '^TestHeredocQueueOrder$' -count=1 -v | grep -q '^--- PASS: TestHeredocQueueOrder'`)
- [ ] A heredoc opened mid-expression takes its body after the line ends (`go test ./internal/lexer/ -run '^TestHeredocMidExpression$' -count=1 -v | grep -q '^--- PASS: TestHeredocMidExpression'`)
- [ ] `<<"EOF"` is a heredoc and `<< "EOF"` is not (`go test ./internal/lexer/ -run '^TestHeredocWhitespaceAsymmetry$' -count=1 -v | grep -q '^--- PASS: TestHeredocWhitespaceAsymmetry'`)

Spec: §2.9.3, §2.9.5, §2.15.1. toke.c:7173-7177.
