# PVM End-to-End Tests

This directory contains end-to-end tests for the PVM (Perl Version Manager) system. These tests verify that the complete system works as expected, including:

- Version installation and management
- Shell integration
- Shim creation and execution
- Configuration handling

## Running Tests

From the project root directory:

```bash
# Run all E2E tests
go test ./test/e2e/... -v

# Run specific test file
go test ./test/e2e/version_test.go -v

# Skip slow tests (e.g., actual Perl installation)
go test ./test/e2e/... -v -short

# Preserve test environment for debugging (don't clean up)
go test ./test/e2e/... -v -preserve
```

## Test Environment

Each test creates an isolated environment with:

- Temporary HOME directory
- Isolated XDG directories (config, data, cache)
- Fresh PVM binary built from current source
- Custom PATH with shims directory
- No interference with host system

## Test Structure

- `helpers/`: Common test utilities
  - `env.go`: Test environment setup and cleanup
  - `assertions.go`: Test assertions and helpers
  - `commands.go`: Functions to run PVM commands
- `fixtures/`: Test input files and fixtures
- `version_test.go`: Tests for version installation and management
- `shell_test.go`: Tests for shell integration
- `shim_test.go`: Tests for shim creation and execution
- `config_test.go`: Tests for configuration system
- `main_test.go`: Main test setup and common utilities

## Writing New Tests

1. Create a new test file for a specific feature area
2. Use the `NewTestEnv()` function to create an isolated test environment
3. Run commands with `env.RunCommand()` or `env.RunPVM()`
4. Make assertions with `helpers.AssertX()` functions
5. Defer `env.Cleanup()` to ensure proper teardown

## Important Notes

- Actual Perl installation tests are marked as `t.Skip()` when running with `-short` flag
- Tests ensure they clean up after themselves, even on failure
- Each test function gets its own isolated environment
