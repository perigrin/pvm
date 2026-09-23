---
title: "The #line directive is lexed as a comment, and the corpus file passes anyway"
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-23T23:53:42.258650321Z
updated: 2026-09-23T23:53:42.258650321Z
---

`#line 200 "bzzzt"` at column zero is not a comment -- it rewrites
`__LINE__` and `__FILE__` for everything after it. Our lexer treats it
as one, and `conformance/12_packages/11_line_directive.t` passes
anyway.

Found by the gate on `01a0cf3e`, which fixed the unrelated refusal that
had been masking this.

## Measured

`internal/lexer/scan.go`'s `scanComment` consumes `#` to end of line
with no special case, and nothing else in the lexer looks for the
directive:

    $ grep -rn 'line.*directive' internal/lexer/*.go     # nothing

So the directive is skipped as trivia and our `__LINE__`/`__FILE__`
would report the real position rather than the rewritten one -- which
is exactly the failure the corpus file's own header warns about:

> a lexer that treats every `#` as a comment skips the directive and
> reports `2` and the real path -- silently.

## Why the corpus file does not catch it

- `--- expect output` is what PERL prints, so it tests perl, not us.
- `--- expect parses` is satisfied by treating the directive as a
  comment, which is what makes the file parse.
- There is no token fact that could reach it: the glossary has twelve
  categories and none names a line directive.

The file now carries a paragraph saying so, so this is recorded rather
than silent, but a warning in a header is not a check.

## What this needs

A `line directive` glossary category and a lexer token for it, so the
corpus can assert the directive was RECOGNISED rather than skipped.
That is the same shape as the heredoc opener and body categories: a
token kind whose existence is the claim.

Whether `__LINE__` and `__FILE__` should then report the rewritten
values is a separate question and probably a later one -- recognising
the directive is what makes the claim assertable at all.

## Acceptance criteria

- [ ] The glossary names a line-directive category (`grep -q 'line directive' conformance/GLOSSARY.md`)
- [ ] The lexer emits a distinct token for a column-zero `#line` (`go test ./internal/lexer/ -run '^TestLineDirective' -count=1 -v | grep -q '^--- PASS'`)
- [ ] The corpus file asserts the directive was recognised rather than skipped (`awk '/^--- expect tokens/{s=1;next} /^--- /{s=0} s' conformance/12_packages/11_line_directive.t | grep -q 'line directive'`)
- [ ] The whole suite is green (`make test`)
