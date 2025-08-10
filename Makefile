.PHONY: all test clean install tree-sitter vendor cross-compile release install-tools check-tools
.PHONY: build-dev build-release lint check fmt check-generate generate
.PHONY: test-performance test-stress test-integration test-unit profile optimize performance-analysis test-repository-consistency
.PHONY: install-test-perl test-with-binary-perl

# Define binaries - we only build pvm, others are symlinks
BINARIES := pvm
BUILDDIR := build
LIBDIR := lib

# Build configuration
BUILD_TAGS ?=
LDFLAGS := -s -w
DEBUG_LDFLAGS :=

# Development build uses debug symbols and development tags
build-dev: BUILD_TAGS += debug,noembed
build-dev: LDFLAGS := $(DEBUG_LDFLAGS)
build-dev: $(BUILDDIR) $(LIBDIR) tree-sitter $(BINARIES)

# Release build uses optimized flags and embed tags
build-release: BUILD_TAGS += release,embed
build-release: $(BUILDDIR) $(LIBDIR) tree-sitter $(BINARIES)

# Default target
all: build-dev

# Set CGO flags for tree-sitter integration
export CGO_ENABLED=1
export CGO_CFLAGS=-I$(PWD)/tree-sitter-typed-perl/include -I$(PWD)/tree-sitter-typed-perl -I$(PWD)/tree-sitter-typed-perl/src
export CGO_LDFLAGS=-L$(PWD)/tree-sitter-typed-perl -Wl,-rpath,$(PWD)/tree-sitter-typed-perl

$(BUILDDIR):
	mkdir -p $(BUILDDIR)

$(LIBDIR):
	mkdir -p $(LIBDIR)

# Ensure vendor directory exists
vendor:
	go mod vendor

# Build tree-sitter-typed-perl library
tree-sitter: $(LIBDIR)
	@echo "Building tree-sitter-typed-perl parser..."
	cd tree-sitter-typed-perl && $(MAKE) generate && $(MAKE) libtree-sitter-typed-perl.a
	@echo "Tree-sitter-typed-perl build complete"

# Tool management
install-tools:
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/stringer@latest
	go install github.com/matryer/moq@latest
	go install gotest.tools/gotestsum@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install golang.org/x/tools/cmd/godoc@latest
	@echo "Development tools installed successfully"

check-tools:
	@echo "Checking development tools..."
	@GOBIN_PATH=$$(go env GOPATH)/bin; \
	if [ ! -z "$$(go env GOBIN)" ]; then \
		GOBIN_PATH=$$(go env GOBIN); \
	fi; \
	missing_tools=""; \
	[ -f "$$GOBIN_PATH/stringer" ] || missing_tools="$$missing_tools stringer"; \
	[ -f "$$GOBIN_PATH/moq" ] || missing_tools="$$missing_tools moq"; \
	[ -f "$$GOBIN_PATH/gotestsum" ] || missing_tools="$$missing_tools gotestsum"; \
	[ -f "$$GOBIN_PATH/staticcheck" ] || missing_tools="$$missing_tools staticcheck"; \
	[ -f "$$GOBIN_PATH/govulncheck" ] || missing_tools="$$missing_tools govulncheck"; \
	if [ ! -z "$$missing_tools" ]; then \
		echo "Missing tools:$$missing_tools"; \
		echo "Installing missing tools automatically..."; \
		$(MAKE) install-tools; \
	fi
	@echo "All tools are available"

# Build rules for each binary
pvm: LDFLAGS := $(DEBUG_LDFLAGS)
pvm: $(BUILDDIR) tree-sitter
	go build -mod=mod -tags="debug,noembed" -ldflags="$(LDFLAGS)" -o $(BUILDDIR)/pvm ./cmd/pvm
	@echo "Creating symlinks for other components..."
	cd $(BUILDDIR) && ./pvm symlinks create

# Legacy targets for backward compatibility - now just create symlinks
pvx pm psc: pvm
	@echo "Using pvm binary with symlinks for $@"

# Enhanced testing with gotestsum
# Auto-detects CPU count for optimal parallelization
# Default test runs in short mode for fast feedback (10% sampling of long tests)
test: tree-sitter install-tools
	@PARALLEL_JOBS=$$(go run scripts/cpu_count.go); \
	echo "Running tests with $$PARALLEL_JOBS parallel workers..."; \
	LD_LIBRARY_PATH=$(PWD)/tree-sitter-typed-perl:$$LD_LIBRARY_PATH PLENV_VERSION= PVM_PERL_VERSION= PVM_TEST_MODE=integration PVM_SKIP_NETWORK_CALLS=1 go run gotest.tools/gotestsum@latest --format=short --jsonfile=test-results.json -- -mod=mod -timeout=3m -parallel=$$PARALLEL_JOBS ./...; \
	TEST_EXIT_CODE=$$?; \
	if [ $$TEST_EXIT_CODE -ne 0 ]; then \
		echo ""; \
		echo "📊 TEST SUMMARY REPORT"; \
		echo "======================"; \
		go run scripts/test_summary.go test-results.json; \
		echo ""; \
		echo "💡 TIP: Run 'make test-full' for comprehensive testing"; \
		echo "💡 TIP: Run 'go test -v ./path/to/package -run TestName' for specific test debugging"; \
	fi; \
	exit $$TEST_EXIT_CODE

# Full test suite for comprehensive validation (all tests including slow ones)
test-full: tree-sitter install-tools
	@PARALLEL_JOBS=$$(go run scripts/cpu_count.go); \
	echo "Running full test suite with $$PARALLEL_JOBS parallel workers..."; \
	LD_LIBRARY_PATH=$(PWD)/tree-sitter-typed-perl:$$LD_LIBRARY_PATH PLENV_VERSION= PVM_PERL_VERSION= PVM_TEST_MODE=full PVM_SKIP_NETWORK_CALLS=1 go run gotest.tools/gotestsum@latest --format=standard-verbose --jsonfile=test-results-full.json -- -mod=mod -timeout=10m -parallel=$$PARALLEL_JOBS ./...; \
	TEST_EXIT_CODE=$$?; \
	if [ $$TEST_EXIT_CODE -ne 0 ]; then \
		echo ""; \
		echo "📊 COMPREHENSIVE TEST SUMMARY REPORT"; \
		echo "===================================="; \
		go run scripts/test_summary.go test-results-full.json; \
	else \
		echo ""; \
		echo "🎉 ALL TESTS PASSING! Comprehensive test suite complete."; \
	fi; \
	exit $$TEST_EXIT_CODE

# Legacy compatibility
test-short: test

# Baseline testing for regression prevention
test-baselines: tree-sitter install-tools
	@echo "Running baseline tests..."
	go run gotest.tools/gotestsum@latest --format=short -- -mod=mod -run=TestParser_Baselines ./internal/parser/
	go run gotest.tools/gotestsum@latest --format=short -- -mod=mod -run=TestTypeChecker_Baselines ./internal/typechecker/

test-baselines-update: tree-sitter install-tools
	@echo "Updating baseline tests..."
	UPDATE_BASELINES=1 go run gotest.tools/gotestsum@latest --format=short -- -mod=mod -run=TestParser_Baselines ./internal/parser/
	UPDATE_BASELINES=1 go run gotest.tools/gotestsum@latest --format=short -- -mod=mod -run=TestTypeChecker_Baselines ./internal/typechecker/

test-performance-baseline: tree-sitter install-tools
	@echo "Updating performance baselines..."
	UPDATE_PERFORMANCE_BASELINE=1 go test -mod=mod -bench=BenchmarkParser_Performance -benchmem ./internal/parser/
	UPDATE_PERFORMANCE_BASELINE=1 go test -mod=mod -bench=BenchmarkTypeChecker_Performance -benchmem ./internal/typechecker/

# Coverage reporting
test-coverage: tree-sitter install-tools
	@echo "Running tests with coverage..."
	@PARALLEL_JOBS=$$(go run scripts/cpu_count.go); \
	echo "Using $$PARALLEL_JOBS parallel workers for coverage..."; \
	go run gotest.tools/gotestsum@latest --format=short -- -mod=mod -short -timeout=3m -parallel=$$PARALLEL_JOBS -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

# Full coverage reporting
test-coverage-full: tree-sitter install-tools
	@echo "Running full test suite with coverage..."
	@PARALLEL_JOBS=$$(go run scripts/cpu_count.go); \
	echo "Using $$PARALLEL_JOBS parallel workers for full coverage..."; \
	go run gotest.tools/gotestsum@latest --format=short -- -mod=mod -timeout=10m -parallel=$$PARALLEL_JOBS -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Full coverage report saved to coverage.html"

test-coverage-report: test-coverage
	@echo "Coverage Summary:"
	go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'

# Integration testing
test-integration: tree-sitter install-tools
	@echo "Running integration tests..."
	@PARALLEL_JOBS=$$(go run scripts/cpu_count.go); \
	go run gotest.tools/gotestsum@latest --format=short -- -mod=mod -tags=integration -parallel=$$PARALLEL_JOBS ./test/e2e/

# Performance optimization targets
optimize-performance: tree-sitter check-tools
	@echo "Running performance optimization analysis..."
	go run scripts/optimize_performance.go

profile-performance: tree-sitter check-tools
	@echo "Profiling performance bottlenecks..."
	go run scripts/profile_performance.go

# Complete test suite with all validations
test-all: tree-sitter install-tools
	@echo "Running complete test suite..."
	make test-baselines
	make test-performance
	make test-integration
	make test-coverage-full

# Run all tests (with tree-sitter support) - legacy compatibility
test-go: tree-sitter
	go test -mod=mod -v ./...

# Run tests without vendor (for better tree-sitter compatibility)
test-novendor: tree-sitter
	rm -rf vendor || true
	go test -mod=mod -v ./...

# Run repository consistency tests to prevent configuration regressions
test-repository-consistency:
	@echo "Running repository consistency tests..."
	go run gotest.tools/gotestsum@latest --format=short -- \
		./internal/config \
		./internal/updater \
		./internal/version \
		-run="RepositoryConsistency"
	@echo "✅ Repository consistency tests passed"

test-parser: tree-sitter install-tools
	go run gotest.tools/gotestsum@latest --format=short -- -mod=mod ./internal/parser/...

test-ast: tree-sitter install-tools
	go run gotest.tools/gotestsum@latest --format=short -- -mod=mod ./internal/ast/... ./internal/astnav/...

# Binary Perl test infrastructure
install-test-perl:
	@echo "Ensuring Perl 5.40.0 is available for testing..."
	@if ! ./build/pvm list | grep -q "5.40.0"; then \
		echo "Installing Perl 5.40.0..."; \
		./build/pvm install --prefer-binary 5.40.0; \
	else \
		echo "Perl 5.40.0 already installed"; \
	fi

test-with-binary-perl: tree-sitter install-tools install-test-perl
	@echo "Running tests with binary Perl infrastructure..."
	@PARALLEL_JOBS=$$(go run scripts/cpu_count.go); \
	PVM_TEST_PERL_VERSION=5.40.0 PVM_USE_BINARY_PERL=1 \
	LD_LIBRARY_PATH=$(PWD)/tree-sitter-typed-perl:$$LD_LIBRARY_PATH \
	go run gotest.tools/gotestsum@latest --format=short --jsonfile=test-results-binary.json -- \
	-mod=mod -timeout=5m -parallel=$$PARALLEL_JOBS ./test/e2e/...; \
	TEST_EXIT_CODE=$$?; \
	if [ $$TEST_EXIT_CODE -ne 0 ]; then \
		echo ""; \
		echo "📊 BINARY PERL TEST SUMMARY"; \
		echo "==========================="; \
		go run scripts/test_summary.go test-results-binary.json; \
	fi; \
	exit $$TEST_EXIT_CODE

# Benchmarking
bench: tree-sitter
	go test -mod=mod -bench=. -benchmem ./...

bench-compare: tree-sitter
	@echo "Running benchmark comparison..."
	go test -mod=mod -bench=. -benchmem -count=5 ./...

# Code quality
lint: check-tools
	staticcheck ./...
	go vet ./...

fmt:
	go fmt ./...

check: lint
	go mod tidy
	go mod verify

# Security scanning
security: check-tools
	@echo "🔍 Running comprehensive security scan..."
	@echo "📊 Running vulnerability check..."
	govulncheck ./...
	@echo "🛡️ Running static security analysis..."
	@command -v gosec >/dev/null 2>&1 || go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...

# Code generation
generate:
	go generate ./...

check-generate: generate
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Generated code is out of date. Please run 'make generate' and commit the changes."; \
		git status --porcelain; \
		exit 1; \
	fi

# Development environment setup
setup: install-tools
	go mod download
	go mod tidy

# Build performance monitoring
build-monitor:
	@echo "🔍 Running monitored build..."
	./scripts/build-monitor.sh all

build-monitor-component:
	@if [ -z "$(COMPONENT)" ]; then \
		echo "Usage: make build-monitor-component COMPONENT=<pvm|pvx|pm|psc>"; \
		exit 1; \
	fi
	./scripts/build-monitor.sh $(COMPONENT)

# Performance reporting
performance-report:
	@echo "📊 Build Performance Report"
	@echo "============================"
	@if [ -d .build-metrics ]; then \
		echo "Recent build metrics:"; \
		ls -la .build-metrics/ | tail -5; \
	else \
		echo "No build metrics found. Run 'make build-monitor' first."; \
	fi

# Clean build artifacts
clean:
	rm -rf $(BUILDDIR)
	rm -rf $(LIBDIR)
	cd tree-sitter-typed-perl && $(MAKE) clean || true

# Install binaries
install: $(BINARIES)
	cp $(BUILDDIR)/pvm $(GOPATH)/bin/pvm
	@echo "Creating symlinks in $(GOPATH)/bin..."
	cd $(GOPATH)/bin && ./pvm symlinks create

# Set version info for release builds
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
GITHUB_TOKEN ?=
RELEASE_LDFLAGS = -s -w -X 'tamarou.com/pvm/internal/version.Version=$(VERSION)' -X 'tamarou.com/pvm/internal/version.BuildTime=$(BUILD_TIME)' -X 'tamarou.com/pvm/internal/version.CommitHash=$(COMMIT)' -X 'tamarou.com/pvm/internal/version.GitHubToken=$(GITHUB_TOKEN)'

# Cross-compile for supported platforms
cross-compile: tree-sitter
	@echo "Cross-compiling for supported platforms..."
	mkdir -p $(BUILDDIR)/release

	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -ldflags="$(RELEASE_LDFLAGS)" -o $(BUILDDIR)/release/pvm-linux-amd64 ./cmd/pvm

	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(RELEASE_LDFLAGS)" -o $(BUILDDIR)/release/pvm-darwin-arm64 ./cmd/pvm

	@echo "Cross-compilation complete. Single pvm binary for each platform in $(BUILDDIR)/release/"
	@echo "Use 'pvm symlinks create' after installation to create pvx, pm, psc symlinks"

# Create release archives
release: cross-compile
	@echo "Creating release archives..."
	cd $(BUILDDIR)/release && \
	tar -czf pvm-linux-amd64.tar.gz pvm-linux-amd64 && \
	tar -czf pvm-darwin-arm64.tar.gz pvm-darwin-arm64
	@echo "Release archives created in $(BUILDDIR)/release/"
	@echo "Each archive contains a single pvm binary - use 'pvm symlinks create' after installation"

# Test mode targets for different types of testing
test-unit: tree-sitter install-tools
	@PARALLEL_JOBS=$$(go run scripts/cpu_count.go); \
	echo "Running unit tests only with $$PARALLEL_JOBS parallel workers..."; \
	LD_LIBRARY_PATH=$(PWD)/tree-sitter-typed-perl:$$LD_LIBRARY_PATH PLENV_VERSION= PVM_PERL_VERSION= PVM_TEST_MODE=unit go run gotest.tools/gotestsum@latest --format=short --jsonfile=test-results-unit.json -- -mod=mod -timeout=2m -parallel=$$PARALLEL_JOBS ./...

test-integration: tree-sitter install-tools
	@PARALLEL_JOBS=$$(go run scripts/cpu_count.go); \
	echo "Running integration tests with $$PARALLEL_JOBS parallel workers..."; \
	LD_LIBRARY_PATH=$(PWD)/tree-sitter-typed-perl:$$LD_LIBRARY_PATH PLENV_VERSION= PVM_PERL_VERSION= PVM_TEST_MODE=integration go run gotest.tools/gotestsum@latest --format=short --jsonfile=test-results-integration.json -- -mod=mod -timeout=5m -parallel=$$PARALLEL_JOBS ./...

test-performance: tree-sitter install-tools
	@PARALLEL_JOBS=$$(go run scripts/cpu_count.go); \
	echo "Running performance tests with $$PARALLEL_JOBS parallel workers..."; \
	LD_LIBRARY_PATH=$(PWD)/tree-sitter-typed-perl:$$LD_LIBRARY_PATH PLENV_VERSION= PVM_PERL_VERSION= PVM_TEST_MODE=performance go run gotest.tools/gotestsum@latest --format=standard-verbose --jsonfile=test-results-performance.json -- -mod=mod -timeout=10m -parallel=$$PARALLEL_JOBS ./...

test-stress: tree-sitter install-tools
	@PARALLEL_JOBS=$$(go run scripts/cpu_count.go); \
	echo "Running stress tests with $$PARALLEL_JOBS parallel workers..."; \
	LD_LIBRARY_PATH=$(PWD)/tree-sitter-typed-perl:$$LD_LIBRARY_PATH PLENV_VERSION= PVM_PERL_VERSION= PVM_TEST_MODE=stress go run gotest.tools/gotestsum@latest --format=standard-verbose --jsonfile=test-results-stress.json -- -mod=mod -timeout=15m -parallel=$$PARALLEL_JOBS ./...

# Performance optimization targets
test-benchmarks: tree-sitter
	@echo "🚀 Running performance benchmarks..."
	go test -bench=BenchmarkParser_Performance -benchmem ./internal/parser/ -run=^$

profile: tree-sitter
	@echo "🔍 Generating performance profiles..."
	go run scripts/profile_performance.go
	@echo "📊 Profiles saved to performance-profiles/"
	@echo "🔧 Analyze with: go tool pprof performance-profiles/<profile>.prof"

optimize: tree-sitter
	@echo "🧪 Running optimization validation..."
	go run scripts/test_optimizations.go

performance-analysis: profile optimize
	@echo "📈 Comprehensive performance analysis complete"
	@echo "📂 Check performance-profiles/ for detailed profiling data"
	@echo "💡 Use 'make test-performance' for regular benchmarking"
