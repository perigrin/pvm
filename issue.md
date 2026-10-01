---
title: "Discuss type libraries: declaration files for what a module defines"
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-30T21:55:19.14961642Z
updated: 2026-10-01T04:01:38.346580915Z
---

perigrin (2026-09-30): "eventually we should discuss type libraries similar to TypeScript". Some modules define their interface in ways no static reading of their source can recover:
- XS subs in an installed perl with no .xs (prototypes, packages);
- imports built by a module's own `import`: Mojolicious::Lite's `get`/`plugin` (monkey_patch into the caller), Moose::Role's `requires` (Moose::Exporter), Filter::Simple's `FILTER`;
- keyword plugins: Future::AsyncAwait's `async sub` / `await` syntax.

These are the last refusals in PerlOnJava T1: mojolicious_mount_lifecycle_regression.t, mojolicious_skipped_closure_lifecycle_regression.t, moose_role_constant_method.t, source_filter_subroutine_lexical_storage.t, custom_warning_async_suspend_state.t.

A declaration file per module, like TypeScript's `.d.ts`, could state what a `use` defines (names, prototypes, packages, keywords) for the parser to read instead of the module's code. This issue is the discussion; the design is open.

## Prior design (found 2026-10-01)

perl5-son owns the type-library design; start from it rather than inventing one.

- ~/dev/perl5-son/docs/plans/2026-09-16-a-declaration-syntax-for-signatures.md: perigrin's spelling `sub :infix + (Num $x, Num $y) Num;`. perl 5.42 can't parse it (typed params, return type, bodyless signature), so a declaration file is not a perl file and needs its own parser, which is unwritten. Its later sections record the pvm/bson exchange: a prototype string determines parse shape, and it is a separate column from `operands` (a coercion ceiling), never merged with it.
- ~/dev/perl5-son/lib/B/SoN/TypeLibrary.pm, commit 2688e2d: the TypeScript layering. %SIGNATURES/%BUILTIN_SIGNATURES are perl's own (lib.d.ts); %DECLARED via `declare($name, {operands, result})` is asserted over code never compiled (@types/Foo); sub records are derived. A declaration is an assertion, marked by `is_declared`, and measured rows win over declared ones. perl5-son's stash "TypeLibrary answers the question -- delete the three resolvers" may reshape it.
- ~/dev/perl5-son/docs/plans/2026-09-16-require-as-a-graph-edge-makes-closure-decidable.md (c8627c1): why .d.ts and not OCaml/Rust. CPAN is code the compiler never sees, and `*f = sub {...}` makes any declaration a claim that can be wrong: sound with respect to the declarations, not the program.
- pvm 2026-09-17 (01a0ad3f): imports resolve by parsing the module's own source, with no perl subprocess at parse or build time. A declaration file covers what that reading cannot; it does not replace it.

Open, raised by rebel: the residue is two shapes. Runtime-built imports (Mojolicious::Lite, Moose::Exporter, Filter::Simple) are names and prototypes a declaration can state. A keyword plugin (Future::AsyncAwait) is a grammar extension, not a name; whether one format carries both is the load-bearing question.

## Proposed answer to the open question (2026-10-01)

Grammar extensions are already declarative in Perl. Future::AsyncAwait's .xs registers `register_xs_parse_sublike("async", ...)` (flags PREFIX|BODY_OPTIONAL|ALLOW_PKGNAME) and `register_xs_parse_keyword("await", ...)` (piece XPK_TERMEXPR_SCALARCTX, permit_hintkey "Future::AsyncAwait/async"). Syntax::Keyword::Try, Object::Pad and others use the same XS::Parse::Keyword/Sublike/Infix registries. So a declaration file can carry three kinds of entry: subs (name, prototype, later types), keywords (name plus a piece sequence in XS::Parse::Keyword's vocabulary, lexically gated like `isa`/`defer`), and sublikes/infix operators. One generic piece interpreter in the parser, not a rule per module. Our own feature-gated syntax (defer, isa, try, class/field/method) would be the first entries, the lib.d.ts row, to prove the format on syntax already parsed correctly. Start with explicit declaration files; reading XS::Parse hook structs straight from .xs is a later optimisation.

Decided by perigrin (2026-10-01): source filters (Filter::Simple, Filter::Util::Call) and Devel::Declare rewrite source text, and not parsing them is fine. They stay refusals; source_filter_subroutine_lexical_storage.t is out of scope for type libraries.

## Acceptance criteria

- [ ] A design for module declaration files is written and agreed (`sh -c 'ls docs/plans/*type-librar* >/dev/null'`)
