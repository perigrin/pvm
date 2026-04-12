# Shell-integration test container

Runs the pvm shell-integration tests against bash, zsh, and fish in a
reproducible environment.

## Why

Several pvm bugs only manifest in specific shells (e.g., fish's read-only
`$version`, POSIX `||` leaking into fish eval). Local dev machines typically
have one or two of the three shells; tests that exec the missing shells skip
silently. This container installs all three so the skips never hide
regressions.

## Usage

From the repo root:

```sh
docker build -t pvm-shell-ci -f test/docker/shell-integration/Dockerfile .
docker run --rm pvm-shell-ci
```

Override the default suite with additional args:

```sh
docker run --rm pvm-shell-ci go test ./...
```

## What runs by default

The container runs the tests whose names match
`TestPvmUseRefreshesPathAcrossShells|TestShellTemplates|TestDetectShell|TestGenerateShellUseSystem`
under `./internal/perl/...` and `./test/e2e/...`. That covers:

- cross-shell runtime PATH-refresh after `pvm use`
- static template scans (export `PVM_SHELL`, call `_pvm_update_perl_path`)
- `DetectShell` precedence (`PVM_SHELL` over `$SHELL`)
- `GenerateShellUse` system-path behavior (injection rejection, fish `set -e`,
  PowerShell `Remove-Item`)

## CI wiring

A typical GitHub Actions job:

```yaml
jobs:
  shell-integration:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: docker build -t pvm-shell-ci -f test/docker/shell-integration/Dockerfile .
      - run: docker run --rm pvm-shell-ci
```
