---
title: "`use Test::More tests => 9` imports nothing: the plan list is read as names"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T16:29:04.338119751Z
updated: 2026-09-30T16:29:05.325303549Z
sessions:
- start_sha: 2149d9471a9dd509899efbe7860716bf6b6daeeb
  end_sha: ""
  commits: 0
  started_at: 2026-09-30T16:29:05.325303549Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T16:29:05.325303549Z
---

`use Test::More tests => 9;` imports nothing. The import reader takes `tests => 9` for a computed name list and gives up, so `is` and `ok` are unknown and every parenless call refuses. Test::More's import is Test::Builder::Module's (perl 5.42's Test/Builder/Module.pm:75-122):
- the names under `import => [...]` are the import list;
- everything else is handed to plan();
- Exporter::import gets the stripped list, which means @EXPORT when it is empty.

Test::More's import_extra only records no_diag, merges `import` lists and exports $TODO. A module whose base is Test::Builder::Module was also counted as dynamic (a non-Exporter base), which left the sub table incomplete.

Measured on the T1 corpus with perl 5.42's own @INC as loader: 281 refusals in 83 dirty files, 81 of them on this. After the fix: 46 refusals in 24 files.

## Acceptance criteria

- [ ] A Test::Builder::Module subclass imports @EXPORT or its `import => [...]` names (`go test ./internal/parse/ -run '^TestBuilderModuleImport$' -count=1 -v | grep -q '^--- PASS: TestBuilderModuleImport'`)
