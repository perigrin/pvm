---
title: "`CORE::__DATA__` is not the data marker"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T19:31:46.576572764Z
updated: 2026-09-29T19:31:46.576572764Z
---

`CORE::__DATA__` at the start of a line is not seen as the data marker, so the lexer reads the data after it as Perl. toke.c reaches KEY___DATA__ through the CORE:: lookup too (toke.c:8295). Measured: a file with `print <DATA>;` followed by `CORE::__DATA__` prints the line that comes after it.

perl.git t/op/coreamp.t:1254.

## Acceptance criteria

- [ ] `CORE::__END__` and `CORE::__DATA__` lex as a DataSection (`go test ./internal/lexer/ -run '^TestCoreDataSection$' -count=1 -v | grep -q '^--- PASS: TestCoreDataSection'`)
