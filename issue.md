---
title: "Discuss type libraries: declaration files for what a module defines"
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-30T21:55:19.14961642Z
updated: 2026-10-01T16:07:47.396313835Z
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

## Landed (2026-10-01)

- a70a294e (01a0f5c5): declaration files for subs, embedded under internal/parse/declarations/ as `Module/Name.pmt` (renamed from .pmd in a3095b73), read in place of the module's source whether or not it is installed. Moose, Moose::Role, Mojolicious::Lite. A literal `M->import` in BEGIN now applies as `use M`.
- 15fb2815 (01a0f5f7): declared syntax, `:sublike(PREFIX)`, `:keyword(PIECES)` and `:statement(PIECES)`, pieces TERMEXPR, BLOCK and ANONSUB. Future::AsyncAwait.
- 36df6a2c (01a0f808): declarations/CORE.pmt, the lib.d.ts row: the 188 prototyped builtins as `sub NAME :prototype(PROTO);`, parsed to build the default builtin table (coreprotos.go deleted); types go on these lines. `:prototype(...)` is read as a prototype everywhere.
- 5374dda4: parse.Session, which made the corpus ratchets fast enough to keep adding to (T1 241s -> 4s).

T1 is 985 of 986; the one left is the source filter, out of scope by decision.

Still open: types (perigrin's `sub :infix + (Num $x, Num $y) Num;` -- the attribute-after-name spelling used here, `sub await :keyword(...)`, is what our Perl parser already reads; measured on 5.42, perl's order is name, attributes, then signature (`sub f :lvalue ($x)` is OK; `sub :lvalue f` is a syntax error), so an operator declaration fits as `sub + :infix (Num $x, Num $y) Num;`, the order of perigrin's 2026-09-17 spelling, and only the `+` name and the types need a parser of their own); a user search path for declaration files beyond the embedded set; the remaining XS::Parse pieces and non-PREFIX sublikes; lexical rather than file-wide scope for declared syntax.

Decided by perigrin (2026-10-01): declaration files keep perl's order -- name, attributes, then signature. An operator is `sub + :infix (Num $x, Num $y) Num;`, not `sub :infix + (...)`.

The fully typed form (perigrin, 2026-10-01): `sub NAME :attributes (Type $var, ...) Type { ... }`; a declaration ends in `;` where the body would be, and an anonymous sub omits NAME. perl 5.42 already parses the name, attributes and plain signature in that order; a typed parameter and the return type after `)` are the two pieces the parser must add.

Decided by perigrin (2026-10-01): typed parameters and return types are valid only in .pmt files, for now. Measured on 5.42: without the signatures feature `sub f (Ref $x) {...}` is valid Perl (a prototype string), so accepting types in ordinary source would misread it. The parser reads them in a declaration mode used for .pmt; types/signatures.go's 27 typed builtins then move into CORE.pmt, comments included, and signatures.go goes.

Decided by perigrin (2026-10-01), builtins with no prototype: CORE.pmt declares their types and a shape attribute (`:unary` for the 8 unary ones; no prototype means list operator), and their special forms (filehandle slot, BLOCK first, pattern first, `defined &f`) stay in the parser. parse/keyword.go's tables then derive from CORE.pmt and are deleted, after a test shows the derived shapes agree with them. A type on a slurpy names its elements: `sub print (FileHandle $fh = select(), Str @args = ($_)) Bool;` -- the default handle is the selected one, not STDOUT, and print stringifies (measured on 5.42).

Decided by perigrin (2026-10-01), multi: a .pmt declaration may be `multi sub NAME ...`, several signatures for one name, and a call site selects one by arity, by argument types that are not subtypes of each other, or by call context. When the call site cannot decide, the consumer joins the candidates. Bounded polymorphism needs no multi: `abs (Num $x) Num` narrows by meet(join(args), declared). Measured examples: `each %h` gives (Str, v) and `each @a` gives (Int, v); `select(FH)` returns the previous handle and `select(r,w,e,t)` a count; `localtime` is a Str in scalar context and a List[Int] in list context, and the scalar value is not one of the list's.

Blocked on the paper (~/dev/types-paper, perigrin: "we need to add parametric types to the paper"): parametric containers (List[T] covariant vs Array[T]; Hash[T] flattening to List[Str|T]; ArrayRef[T]; lvalue lists for chomp), tuples (List[Str, T] for each), and multi/context-indexed function types. Until then a type on a slurpy in a .pmt is the element type of the caller's flattened list.

Decided by perigrin (2026-10-01), `:context(...)`: an attribute like `:prototype` naming the contexts a declaration answers for, from wantarray's three, spelled in prototype sigils (perigrin): `:context($)` scalar, `:context(@)` list, `:context()` void, `:context($@)` scalar or list. As with :prototype, absent and empty differ: no :context answers for every context, `:context()` for void alone. Without it a declaration answers for every context. Its presence signals wantarray-like behaviour. A call in a context no declaration answers for has no type (Unknown): `sub sort :context(@) (...) List;` leaves scalar sort undeclared, as perldoc leaves it undefined. Example:

    multi sub localtime :prototype(;$) :context($) (Int $time = time) Str;
    multi sub localtime :prototype(;$) :context(@) (Int $time = time) List[Int];

## Acceptance criteria

- [ ] A design for module declaration files is written and agreed (`sh -c 'ls docs/plans/*type-librar* >/dev/null'`)
