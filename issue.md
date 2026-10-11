---
title: "builtin version bundles are not imported: `use builtin \":5.39\"` and `use v5.40` bring nothing into scope"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T07:47:15.291952848Z
updated: 2026-10-11T01:54:43.281832372Z
sessions:
- start_sha: 24bfac4a5293052e7c38dd54a56a58aef818e639
  end_sha: ""
  commits: 0
  started_at: 2026-10-11T00:57:42.520154202Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T00:57:42.520154202Z
---

`use builtin LIST` imports literal names (see importBuiltins in use.go). Two other routes bring builtin functions into scope and are not read:

1. A version bundle in the list: `use builtin ':5.39';` imports every function builtin.c tags SHORTVER(5,39) or earlier. perl.git t/ has it twice.
2. A version declaration: `use v5.39` or later imports the same bundle. Measured with perl 5.42:

       $ perl -MO=Deparse -e 'use v5.40; my $x = blessed $obj;'
       my $x = builtin::blessed(${'obj'});

Without them, `blessed $obj` after `use v5.40;` is `WORD $var` with an unknown WORD and refuses. builtin.c's `builtins[]` carries the bundle version per function; NO_BUNDLE ones (inf, nan, is_bool, stringify, created_as_*, load_module, export_lexically) never come in by bundle.

### Fixed (2026-10-11)

Measured on 5.42: `use builtin ':5.39'` through `':5.42'`, and `use v5.39`
and later, import the same 13 names (true false weaken unweaken is_weak
blessed refaddr reftype ceil floor is_tainted trim indexed); `use v5.36`
and `v5.38` import none; no bundle imports the NO_BUNDLE names (inf, nan,
is_bool, ...). Fixed in 713d1b16, whose test is named
TestBuiltinBundlesImportTheirNames, not the TestBuiltinBundleImports the
criterion first named; it holds the negatives too.

## Acceptance criteria

- [ ] A builtin bundle, by `use builtin ':5.39'` or `use v5.40`, brings its functions into scope, and a NO_BUNDLE name or `use v5.38` brings none (`go test ./internal/parse/ -run '^TestBuiltinBundlesImportTheirNames$' -count=1 -v | grep -q '^--- PASS: TestBuiltinBundlesImportTheirNames'`)
