---
title: "`use if` leaves the sub table incomplete: assume its import fails"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-30T00:17:03.289387514Z
updated: 2026-09-30T00:24:32.396273755Z
sessions:
- start_sha: 162befda1aebda817232e707fd151ff547382530
  end_sha: f59ce40ba15896278e591b95e50c273f2a36ebfe
  commits: 1
  started_at: 2026-09-30T00:17:03.799603818Z
  ended_at: 2026-09-30T00:24:32.396273755Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T00:17:03.799603818Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-30T00:24:32.396273755Z
observed_paths:
- internal/parse/indirectvar_test.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/use.go
---

`use if COND, MODULE, LIST` imports only when COND holds at compile time, and the parser cannot evaluate COND. Today if.pm's own `import` marks the sub table incomplete, so in perl.git t/op/coreamp.t:842 (`use if !is_miniperl, File::Spec::Functions, qw "catfile";`) the later `catfile $dir, 'aaa'` refuses.

perigrin decided (2026-09-30) that the import is assumed to fail. `use if` defines nothing and leaves the table complete, so `catfile` is an unknown word and reads as perl reads one: `$dir->catfile, 'aaa'` (toke.c:8216). This is NOT what a full perl does here: `!is_miniperl` is true there, `catfile` is imported, and the call is `catfile($dir, 'aaa')`. The assumption trades that for a definite reading.

A second fix is needed to reach the call: a computed import list on a pragma that defines no caller subs (`use overload '%{}' => sub {...}`, line 23) marked the table incomplete. It no longer does.

## Acceptance criteria

- [ ] `use if` is assumed to import nothing (`go test ./internal/parse/ -run '^TestUseIfImportFails$' -count=1 -v | grep -q '^--- PASS: TestUseIfImportFails'`)
