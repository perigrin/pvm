# Smoke-test container

Exercises the real user-facing journey against a fresh-install PVM on a
Debian-based system with bash, zsh, and fish available. Runs four suites
per invocation, covering the commands a typical user touches:

## `core-commands.sh` (shell-agnostic, ~38 assertions)

Command-surface guarantees that don't depend on which shell invokes pvm.
Covers `pvm version`, `pvm --help`, `pvm list`/`versions`, `pvm available`
(offline), `pvm completion {bash,zsh,fish}`, `pvm config`, `pvm env list`,
`pvm remote list`, `pvm detect-version`, `pvm self doctor`, `pvm current`
/ `pvm current --bare`, plus `--help` for every top-level command to
catch cobra registration bugs.

## Per-shell journeys (~29–32 assertions each)

`bash-journey.sh` / `zsh-journey.sh` / `fish-journey.fish` each source
the real shell integration and exercise workflows that depend on the
shell (env-var exports, PATH rewriting). Eight sections per shell:

1. **Top-level command presence** — root `pvm --help` lists `local`,
   `global`, `use [version[@library]]` (regression guard for #432).
2. **Setup** — `pvm import-system` + fake version bin via
   `smoke_setup_stub_perl`.
3. **`pvm use`** — PATH rewrite, `PVM_PERL_VERSION` export, `perl -v`
   reflects the switch, bogus version exits non-zero AND short-circuits
   `&&` / `; and` chains (regression guard for #433).
4. **`pvm local`** — writes `.perl-version`, `detect-version` sees it,
   `--unset` removes it.
5. **`pvm global`** — writes global config, `pvm current --bare` picks
   it up, `--unset` clears it.
6. **`pvm use system`** — clears `PVM_PERL_VERSION`.
7. **Library environments** — `pvm env list` shows a manually-created
   env, `pvm use X@testlib` exports `PVM_PERL_LIBRARY`, library bin
   lands on PATH, library `lib/perl5` lands on `PERL5LIB`, clearing
   via `pvm use X` unexports the library, `pvm env remove` tears it
   down.
8. **Doctor sanity** — `pvm self doctor` exits 0 on a clean container
   and does NOT falsely flag stale installs (regression guard for
   PR #436's disambiguation logic).

Fish additionally asserts: no `Unknown command` errors from the template
(PR #434 fish-template regressions), and completions actually register.

## Total coverage

~128 assertions across the four suites, covering roughly 20 of the 31
top-level pvm commands. The six commands that require real Perl install
(`install`, `build-perl`, `install-perl`, `module`, `run`/pvx, `psc`)
intentionally stay out of smoke scope — they belong in the Go e2e suite
(`test/docker/shell-integration/`) or a separate slow-tier CI job.

## Running

From the repo root, with Docker or a Docker-compatible runtime (podman works
if aliased) installed:

```sh
./test/docker/smoke-test/run-smoke.sh
```

The script builds a Linux `pvm` binary, bakes it into the image, then runs
one container invocation per shell. Each journey is a self-contained script
under `journeys/`; a failure is visible as a `[SMOKE FAIL]` marker in the
output and exits non-zero.

## Layout

- `Dockerfile` — minimal Debian image with bash/zsh/fish + the pvm binary.
  No Go toolchain; this is a user-journey test, not a unit-test container.
- `journeys/common.sh` / `common.fish` — assertion helpers.
- `journeys/bash-journey.sh` / `zsh-journey.sh` / `fish-journey.fish` —
  per-shell journey scripts. Each sources `pvm init $shell`, runs real
  user commands, and asserts observable behavior.
- `run-smoke.sh` — orchestration (build + run + aggregate).

## Relationship to the shell-integration test container

`../shell-integration/Dockerfile` runs the Go e2e test suite in a container.
It's a *CI-parity* check — the tests themselves live in `test/e2e/`.

This smoke-test container is different: it has no Go toolchain and does
not run `go test`. It runs the same commands a real user types, and asserts
on the shell output. If a future change to the templates or the `pvm`
binary breaks the user-visible contract, this catches it even if the
unit-test env (which uses the built-in-test `env.PVMBinary`) misses it
because a stale `pvm` in PATH shadows the fresh build.

## Offline

The image runs with `PVM_SKIP_NETWORK_CALLS=1` set by default — the smoke
test does not reach MetaCPAN or any other external service. It's safe to
run in CI environments with no internet access.
