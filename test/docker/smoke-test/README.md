# Smoke-test container

Exercises the real user-facing journey against a fresh-install PVM on a
Debian-based system with bash, zsh, and fish available. Asserts the two
originally-reported bug scenarios stay fixed:

1. **`pvm local` / `pvm global` are top-level commands** (issue #432).
2. **`pvm use X` actually changes the active `perl`** — PATH refresh works,
   `PVM_PERL_VERSION` exports, and `perl -v` reflects the new version in
   the same shell (issue #433). Also exits non-zero for bogus versions.

Fish gets extra assertions covering the bugs uncovered during PR #434
development: no fish parse errors in the templates, completions actually
register after `pvm init fish | source`.

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
