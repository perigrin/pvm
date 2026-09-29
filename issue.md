---
title: "`CORE::__DATA__` is not the data marker"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T19:31:46.576572764Z
updated: 2026-09-29T19:49:30.518101967Z
sessions:
- start_sha: 0427fe15cc8c6fee517453d3a8c364be2453ac8d
  end_sha: a5d54ef27561683c56b464076aa0c5cf21adaded
  commits: 1
  started_at: 2026-09-29T19:31:47.982186484Z
  ended_at: 2026-09-29T19:49:30.518101967Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T19:31:47.982186484Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T19:49:30.518101967Z
observed_paths:
- internal/lexer/pod.go
- internal/lexer/pod_test.go
- internal/parse/testdata/perlgitt.ratchet
---

`CORE::__DATA__` at the start of a line is not seen as the data marker, so the lexer reads the data after it as Perl. toke.c reaches KEY___DATA__ through the CORE:: lookup too (toke.c:8295). Measured: a file with `print <DATA>;` followed by `CORE::__DATA__` prints the line that comes after it.

perl.git t/op/coreamp.t:1254.

## Acceptance criteria

- [ ] `CORE::__END__` and `CORE::__DATA__` lex as a DataSection (`go test ./internal/lexer/ -run '^TestCoreDataSection$' -count=1 -v | grep -q '^--- PASS: TestCoreDataSection'`)
