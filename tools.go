//go:build tools

// ABOUTME: Build tool dependency management for development
// ABOUTME: Ensures consistent tool versions across development environments

package tools

import (
	// Code generation tools
	_ "golang.org/x/tools/cmd/stringer"

	// Mock generation for testing
	_ "github.com/matryer/moq"

	// Enhanced test runner
	_ "gotest.tools/gotestsum"

	// Static analysis and linting
	_ "honnef.co/go/tools/cmd/staticcheck"

	// Security scanning
	_ "golang.org/x/vuln/cmd/govulncheck"

	// Benchmarking and profiling (using go test built-in benchstat)

	// Documentation generation
	_ "golang.org/x/tools/cmd/godoc"
)
