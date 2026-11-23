# crochet:assess round 6: codebase lens

**Subject.** `docs/plans/2026-11-23-chart-stateless-precedence.md`, blob **60327eb7d7e379443281e0623b5d5f586a382f9c** (`git hash-object`). The worktree HEAD is 0d2b31610, the minute on top of 788778ed0. This is revision 6.

**Order of reading.** The subject first; then `INDEX.md`, `harness/fresh4/`, the Makefile, `ci.yml`, `cmd/ratchet` (`lex.go`, `perlgit.go`, `absence_reference_test.go`), `forest.go` and `readings.go`. The plans diff last.

**What I ran.** One timing probe on the combined prototype through `go test -overlay`. The worktree is untouched.

## 1. Disposition of C30 to C33

| Id | Disposition |
|---|---|
| C30 | **Released.** §3.4's inputs drop "the port's refused cases" and add "precedence-refused sentences, from a generator committed with a fixed seed, which reproduces the 185 and 989 refused sentences of the evidence". |
| C31 | **Released.** §3.4: "Today's `Constrain` stays, exported as a reference implementation, `forest.ConstrainReference`. It is documented as used only by tests … It cannot move out of `forest`". The aggregate runs from `cmd/ratchet` behind a build tag. |
| C32 | **Released.** §3.1: "The tables are generated when `forest.NewPrecedence` runs … They are stored on the `Precedence` value." §3.3: "Per-call state lives on the call, never on the shared `Precedence`." |
| C33 | **Issue-level, carried to refinement.** It covers the root's per-alternative spine top, canonicalisation, the `pi = -1` allow on a mixed pair, the empty-right-child exemption, cached-slice aliasing and the per-call memo heads. None belongs in the decision. |

I hold these findings and I am also the one who releases them, so I may be biased toward finding the bar met. Each release above quotes the text it rests on.

## 2. The reference and where the differential runs

**Workable, with one real consequence (C34).**

- **Package placement.**
  - A `//go:build precedencereference` test in `internal/chart/cmd/ratchet` (package `main`) can reach `aggregateFiles` and `corpus()`. It can call the exported `forest.Build` and `forest.ConstrainReference`. And it can copy a `Forest` by hand, since `Root`, `Nodes` and `Alts` are exported. This is exactly how `absence_reference_test.go` is placed (`//go:build absencereference`, package `main`).
  - The referee is in package `chart` and the toys are in package `forest`. Both can call the exported reference in the ordinary suite.
  - No import cycle results: `cmd/ratchet` imports `chart` and `forest`, and `chart` imports `forest`.
- **The constraint's input.** In `parseStream`, `libraryOperators` runs between `Build` and `Constrain`, and it only reads the forest. So `Build`'s output is the constraint's input, and the boundary is reachable from outside `forest`.

### C34: Partial. "Follows the precedent of chart-absence-reference" changes what `make test` needs

The precedent's corpus is in the tree. ci.yml says, above its step: "Its corpus is in the tree (T1's in-tree default), so it needs no env." The precedence reference's aggregate needs more:

- **perl.git `t/`, pinned and clean.** `perlGit` refuses any root that is missing, dirty or not at 94e50866 (`perlgit.go`), from `PERL5_CORPUS` or `/opt/corpora/perl5`.
- **The full superset-population pass.** `aggregateFiles` calls `supersetPopulation` over RD verdicts and perl labels.

Under §3.4 ("`make test` runs it after the suite … When its corpora are missing, the target fails; it never skips"), three things follow:

- Every `make test` run now needs the pinned perl5 checkout. CLAUDE.md requires 100% of `make test` before every commit, so a contributor or agent without `/opt/corpora/perl5` can no longer commit. The precedent never imposed that.
- `make test` also absorbs a full population pass.
- In CI, the absence-reference step sets no `PERL5_CORPUS`. The Ratchets job sets it only on its own `go test` step. The new step must set it.

The decision should choose one of two options:
- **(a)** keep `chart-precedence-reference` out of `make test` (CI's Ratchets job plus the named target);
- **(b)** split it, so `make test` runs only an in-tree part (T1, Guacamole, std5k) and CI runs the perl.git part.

Either way, it should say what a developer without the corpus runs.

## 3. Is "verdict, exact kept set, readings" at the boundary sufficient and runnable?

**Sufficient.** Same built forest, same verdict and identical kept set mean an identical constrained forest. Resolve and readings are deterministic functions of that forest, so everything downstream is equal. §3.4's argument that recognition, refusal byte and `Expected` cannot differ is correct: nothing before the constraint changes, and a precedence refusal is always `{Byte:0, Reason:precedence}`.

### C35: Low. The readings count is not reachable at the boundary from `cmd/ratchet`

`countReadings` is unexported. `ParseWith` calls it after resolve (`forest.go`; readings.go l.32). From the build-tagged test in `cmd/ratchet` there are only two routes: an export, or a switch inside `ParseWith`. The prototype used the second, a production global `ConstrainVariant`.

Equal kept sets already imply equal readings, so one of two sentences suffices:
- state that readings follow from the kept set, or
- say they are compared in `forest`'s own tests.

Either keeps a test switch out of production code.

## 4. The seeded generator

**Feasible.** `harness/fresh4/zz_fresh4_test.go.txt` (`TestZZFresh4Crafted`) holds 83 hand-written lines plus `rand.New(rand.NewSource(4))`, with `F4_N` defaulting to 20,000. That is the 20,083 operator sentences. `zz_fresh4d_test.go.txt` (`TestZZFresh4Chains`) generates the 15,000 chains from `rand.NewSource(44)`.

`math/rand`'s seeded `Source` sequence is stable across Go releases, so the sentences reproduce exactly. The harness around them (`fresh4Run`) drives the prototype's V0/V1/V2 switch and must be rewritten against production and `ConstrainReference`.

### C36: Low. Don't make 185 and 989 assertions

Those counts belong to today's grammar and lexer. A table or lexer change moves them legitimately. As asserted numbers they would be change detectors, contrary to the ratchet's own rules. The test should:
- assert that the reference and production agree on every sentence;
- assert that some sentences are refused, which is the passes-today half of the README's own rule.

It should not assert the exact counts.

**Runtime.** 35,083 parses, each with two constraints, in the ordinary suite. Measure it. If it is slow, run it behind the build tag with the aggregate, and keep the 1,174 refused sentences, plus a sample, in the ordinary suite.

## 5. Tables on `Precedence` against `grammar.Derived`

**Consistent.**

- `forest.Parse` reaches `Precedence` only through `grammar.Derived(tab, buildParsePrecedence)` (readings.go l.116-131). That is once per tables value, so the generated tables are built once and shared read-only, which is safe for concurrent readers.
- The referee builds its own `Precedence` once per run (`newPrecedence`; `perlyPrecedence` is a `sync.OnceValues`).
- The toys build theirs once per toy load (precedence_test.go l.68).
- The generation error reaches `ParseWith` as `prec.err`, which becomes `Refusal{Reason: err}` for every file (forest.go l.134-137). The ratchet's `pipelineReason` treats any other reason as an error, so it is loud.

### C37: Low. Report the generation cost

On the committed grammar, the prototype's `NewStateless` takes about **90 ms** per build. Today's `buildParsePrecedence` takes about 310 ms, including loading the committed grammar (`TestZZAssess6GenTime`, overlay).

The cost is once per process, and the speed keys exclude derived tables. It adds to every cold start and to each test that builds a `Precedence`. Report it with the change. It is not a stop.

## 6. What I did not examine

- The diff `2782fd181..788778ed0` beyond §3, and §§4–6, except where they are cited above.
- The runtime of the full generator.
- The superset-population cost inside `make test`.
- The CI pin's relation to 94e50866. I did not open `perl542.pin`.
- The no-map prototype.
- The other assessors' round-5 reports.

## 7. Recommendation: **modify**

One text change: **C34.** Decide whether `chart-precedence-reference` runs inside `make test`, given that its aggregate needs the pinned perl.git checkout, which the precedent never required. Name the CI environment it needs.

Low: C35, C36, C37. C33 is carried to refinement. The rest of §3 is accurate and workable on the code as it stands.
