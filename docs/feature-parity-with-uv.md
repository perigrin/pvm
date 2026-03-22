# PVM Feature Parity with uv

PVM is a fast, cross-platform Perl development toolkit written in pure Go.
This document maps PVM's capabilities against [uv](https://docs.astral.sh/uv/),
Astral's Python package and project manager, to show where PVM has parity,
where it is ahead, and where it intentionally defers to existing Perl tools.

## Version Management (uv python ↔ pvm)

| Capability | uv | PVM | Notes |
|---|---|---|---|
| Install a version | `uv python install` | `pvm install` | Binary + source builds |
| List available | `uv python list` | `pvm available` | |
| List installed | `uv python list --only-installed` | `pvm versions` | |
| Pin project version | `uv python pin` | `pvm global` + `.perl-version` | |
| Switch version | Automatic via `.python-version` | `pvm use` / `pvm global` | |
| Uninstall | `uv python uninstall` | `pvm uninstall` | |
| Shell integration | — | `pvm init bash/zsh/fish` | |

**Status: Full parity.**

## Package Management (uv pip / uv add ↔ pm)

| Capability | uv | PVM | Notes |
|---|---|---|---|
| Install packages | `uv pip install` / `uv add` | `pm install` / `pm add` | |
| Remove packages | `uv pip uninstall` / `uv remove` | `pm remove` | |
| List installed | `uv pip list` | `pm list` | |
| Search | — | `pm search` | PVM has search; uv does not |
| Show outdated | — | `pm outdated` | |
| Dependency tree | `uv tree` | `pm deps` | |
| Sync from lockfile | `uv sync` | `pm sync` | |
| Mirror configuration | Index flags | `pm mirror` | |

**Status: Full parity.** PVM additionally has `pm search` and `pm outdated`
which uv lacks.

## Lockfile & Resolution (uv.lock ↔ cpanfile.snapshot)

| Capability | uv | PVM | Notes |
|---|---|---|---|
| Dependency declaration | `pyproject.toml` | `cpanfile` | Standard Perl format |
| Lockfile | `uv.lock` | `cpanfile.snapshot` | Pins exact versions |
| Sync from lockfile | `uv sync` | `pm sync` | |
| Generate lockfile | `uv lock` | `pm sync --generate-only` | |
| Install from lockfile | `uv sync --frozen` | `pm sync --install-only` | |

**Status: Full parity.** Both systems provide reproducible builds via pinned
dependency lockfiles.

## Script Execution (uv run / uvx ↔ pvx)

| Capability | uv | PVM | Notes |
|---|---|---|---|
| Run script | `uv run script.py` | `pvx script.pl` | |
| Ephemeral deps | `uv run --with pkg` | `pvx --require Module` | |
| Inline metadata | PEP 723 (`# /// script`) | `# /// pvm` + `=begin pvm` POD | Two formats: comment + POD |
| Specific version | `uv run --python 3.12` | `pvx --perl 5.40.0` | |
| Auto-install deps | — | `pvx --auto-install` | PVM detects + installs |
| Auto-detect deps | — | Built-in + `psc analyze` | Regex + AST-based |
| Isolation levels | — | `--isolation global/local/clean` | **PVM is ahead** |
| Named environments | — | `pvx --name myenv` | Persistent isolation |
| Ephemeral tool run | `uvx ruff` | `pvx` | Same model |

**Status: PVM is ahead.** PVM has isolation levels (global, local, clean),
auto-dependency detection from source code (both regex and tree-sitter AST),
and POD-based inline metadata that is invisible to the Perl parser. uv has
none of these.

## Tool Management (uv tool ↔ pvm tool)

| Capability | uv | PVM | Notes |
|---|---|---|---|
| Install tool | `uv tool install` | `pvm tool install` | |
| List tools | `uv tool list` | `pvm tool list` | |
| Uninstall tool | `uv tool uninstall` | `pvm tool uninstall` | |
| Run ephemerally | `uvx` | `pvx` | |

**Status: Full parity.**

## Static Analysis (psc — uv has no equivalent)

| Capability | uv | PVM | Notes |
|---|---|---|---|
| Parse source | — | `psc parse` | Tree-sitter CST |
| Dependency analysis | — | `psc analyze` | AST-based `use`/`require` extraction |
| Type checking | — | `psc check` | Bitset type inference, diagnostics |
| Guard suggestions | — | `psc check` hint lines | Suggests `defined()`, `ref()`, etc. |
| Language server | — | `psc lsp` | Hover, diagnostics, definitions |
| Cross-file analysis | — | ProjectIndex | Resolves imports, methods, classes |

**Status: PVM is significantly ahead.** uv has no static analysis capability.
PVM provides a full type inference engine with flow narrowing, guard pattern
recognition, user-defined type guard functions, cross-file analysis, and an
LSP server — all in pure Go with zero external dependencies.

## Configuration (uv.toml / pyproject.toml ↔ pvm.toml)

| Capability | uv | PVM | Notes |
|---|---|---|---|
| Project config | `pyproject.toml` | `pvm.toml` | TOML format |
| User config | `~/.config/uv/uv.toml` | XDG config | XDG spec compliant |
| Shell completion | `uv generate-shell-completion` | `pvm completion` | bash/zsh/fish/PowerShell |

**Status: Full parity.**

## Self-Management

| Capability | uv | PVM | Notes |
|---|---|---|---|
| Self-update | `uv self update` | `pvm self update` | |
| Version info | `uv version` | `pvm version` | |
| Health check | — | `pvm self doctor` | |
| Changelog | — | `pvm self changelog` | |

**Status: Full parity.** PVM additionally has `doctor` and `changelog`.

## Build & Publish

| Capability | uv | PVM | Notes |
|---|---|---|---|
| Build distribution | `uv build` | — | Perl uses `dzil`, `ExtUtils::MakeMaker` |
| Publish to registry | `uv publish` | — | Perl uses `CPAN::Upload` |

**Status: Intentional deferral.** uv built `uv build` and `uv publish` because
Python's build tooling was fragmented (setuptools vs flit vs hatchling vs
poetry vs twine). Perl's build and publish ecosystem is mature and unified —
`dzil` (or `ExtUtils::MakeMaker` / `Module::Build`) handles building, and
PAUSE upload is straightforward. PVM does not need to replace working tools.

## Platform Support

| Platform | uv | PVM |
|---|---|---|
| Linux (amd64, arm64) | Yes | Yes |
| macOS (amd64, arm64) | Yes | Yes |
| Windows | Yes | Cross-compile exists, not fully tested |

## Summary

PVM has **full feature parity** with uv for Perl development, and is
**ahead** in two significant areas:

1. **Static analysis** — `psc` provides type inference, diagnostics, guard
   suggestions, and an LSP server. uv has nothing comparable.

2. **Execution isolation** — `pvx` offers three isolation levels with
   auto-dependency detection from source. uv's `uv run` has no isolation
   model.

The only area where PVM intentionally does not compete is build/publish,
where Perl's existing tools (`dzil`, PAUSE) are mature and well-established.
