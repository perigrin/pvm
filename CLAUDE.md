# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Test Commands

All test files MUST pass 100% before committing, even tests that appear
unrelated to the current work.

### Basic Commands
- Build all: `make`
- Build individual components:
  - PVM: `make pvm`
  - PVX: `make pvx`
  - PM: `make pm`
  - PSC: `make psc`
- Test all: `make test`
- Clean: `make clean`

### Understanding Test Results

**Always use `make test` for comprehensive test status.**

Individual package testing:
- `go test ./internal/parser` - Parser functionality
- `go test ./internal/psc` - PSC command functionality
- `go test ./test/e2e` - End-to-end integration tests

### Cross-Platform Build
- Cross-compile all platforms: `make cross-compile`

Supported platforms:
- Linux (AMD64, ARM64)
- macOS (AMD64, ARM64)

### Pure-Go Build

This branch uses gotreesitter (pure Go) for Perl parsing. No CGO, no C
compiler, no tree-sitter CLI, no Node.js required. Cross-compilation works
with standard `GOOS`/`GOARCH` environment variables.

## Problem-Solving Philosophy
- **NEVER create workarounds** - If you find yourself needing a workaround, STOP
- Think deeply about the root cause of the problem
- If you can't find a proper solution, ask perigrin for advice
- Workarounds create technical debt and mask the real issues that should be fixed

## Code Style Guidelines

## Test Data Format Preference
- **When encountering JSON-based test files, consider migrating them to Markdown format**
- Markdown test files are more readable, easier to maintain, and support better documentation

## Repository Configuration Protection

**Critical**: The PVM update system must point to the correct GitHub repository (`perigrin/pvm`), not `perigrin/pvm-dev`.

### Files Protected
- `internal/config/types.go` - Default update repository and binary mirrors
- `internal/updater/updater.go` - Updater default options
- `internal/updater/auto_update.go` - Auto-update defaults
- `internal/version/types.go` - Version check defaults
- `internal/pvm/config.go` - PVM config defaults

**If repository configuration tests fail, DO NOT bypass them - fix the underlying configuration issue.**

## Pre-commit Hook Compliance

**CRITICAL**: Never bypass pre-commit hooks using `--no-verify` or `SKIP=` environment variables.

### Why Pre-commit Hooks Are Non-Negotiable

1. **Technical Debt Prevention**: Hooks catch issues early before they accumulate
2. **Code Quality Maintenance**: Enforces consistent formatting, linting, and style
3. **Regression Prevention**: Catches breaking changes and test failures immediately
4. **Team Standards**: Ensures all contributors follow the same quality standards

### When Hooks Fail on Unrelated Files

**DO**: Fix the underlying issues in those files (formatting, linting, etc.)
**DON'T**: Skip the hooks to avoid dealing with unrelated technical debt

## PVM Project Patterns

### Test Failure Protocol
- ALWAYS run `make test` before and after changes
- NO compromises on 100% test pass rate - this is non-negotiable

### Build Dependencies Management
- Use the Makefile for ALL builds - do not create workarounds
- Cross-platform builds require consistent dependency resolution

### Performance Philosophy
- Measure before optimizing - use concrete benchmarks
- Avoid premature optimization in favor of correct implementation
- Correctness > Performance always
