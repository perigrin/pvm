---
title: "Declared syntax: keywords and sub prefixes from a declaration file (Future::AsyncAwait)"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-01T05:37:08.107276027Z
updated: 2026-10-01T05:37:08.107276027Z
---

Second slice of type libraries (01a0f450-9ded-7c8a-8c8b-c04734752b2d): declared syntax. A keyword plugin is a grammar extension, not a name, but plugins built on XS::Parse::Keyword and XS::Parse::Sublike state their grammar as data. Future::AsyncAwait 0.71's AsyncAwait.xs registers `async` as a sublike with XS_PARSE_SUBLIKE_FLAG_PREFIX, `await` with one XPK_TERMEXPR_SCALARCTX, and CANCEL with one XPK_ANONSUB whose build returns KEYWORD_PLUGIN_STMT.

A declaration file restates that with three attributes: `sub async :sublike(PREFIX);`, `sub await :keyword(TERMEXPR);`, `sub CANCEL :statement(ANONSUB);`. The pieces are XS::Parse::Keyword's, minus XPK_. Importing the module brings them into scope file-wide, as features are, and `use M ()` brings none. The parser reads TERMEXPR, BLOCK and ANONSUB; a declaration naming any other piece declares nothing rather than guessing.

Measured on 5.42.0: `await $f + 1` is await($f + 1), `(await $f, 3)` is two elements, `CANCEL { 1 } print "x";` compiles, and `my $c = async sub { 1 };` is an anonymous async sub. PerlOnJava unit/custom_warning_async_suspend_state.t:27.

## Acceptance criteria

- [ ] Declared keywords parse once imported (`go test ./internal/parse/ -run '^TestDeclaredKeywords$' -count=1 -v | grep -q '^--- PASS: TestDeclaredKeywords'`)
