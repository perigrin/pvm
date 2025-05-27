.PHONY: all test clean install tree-sitter vendor cross-compile release install-tools check-tools
.PHONY: build-dev build-release lint check fmt check-generate generate
.PHONY: test-performance profile optimize performance-analysis

# Define binaries
BINARIES := pvm pvx pvi psc
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
export CGO_CFLAGS=-I$(PWD)/tree-sitter-typed-perl/include -I$(PWD)/tree-sitter-typed-perl

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
	cd tree-sitter-typed-perl && $(MAKE) generate
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
	@command -v stringer >/dev/null 2>&1 || (echo "stringer not found, run 'make install-tools'" && exit 1)
	@command -v moq >/dev/null 2>&1 || (echo "moq not found, run 'make install-tools'" && exit 1)
	@command -v gotestsum >/dev/null 2>&1 || (echo "gotestsum not found, run 'make install-tools'" && exit 1)
	@command -v staticcheck >/dev/null 2>&1 || (echo "staticcheck not found, run 'make install-tools'" && exit 1)
	@command -v govulncheck >/dev/null 2>&1 || (echo "govulncheck not found, run 'make install-tools'" && exit 1)
	@echo "All tools are available"

# Build rules for each binary
pvm: $(BUILDDIR)
	go build -mod=mod -tags="$(BUILD_TAGS)" -ldflags="$(LDFLAGS)" -o $(BUILDDIR)/pvm ./cmd/pvm

pvx: $(BUILDDIR)
	go build -mod=mod -tags="$(BUILD_TAGS)" -ldflags="$(LDFLAGS)" -o $(BUILDDIR)/pvx ./cmd/pvx

pvi: $(BUILDDIR)
	go build -mod=mod -tags="$(BUILD_TAGS)" -ldflags="$(LDFLAGS)" -o $(BUILDDIR)/pvi ./cmd/pvi

psc: $(BUILDDIR) tree-sitter
	go build -mod=mod -tags="$(BUILD_TAGS)" -ldflags="$(LDFLAGS)" -o $(BUILDDIR)/psc ./cmd/psc

# Enhanced testing with gotestsum
test: tree-sitter check-tools
	gotestsum --format=standard-verbose -- -mod=mod ./...

test-short: tree-sitter check-tools
	gotestsum --format=short -- -mod=mod -short ./...

# Baseline testing for regression prevention
test-baselines: tree-sitter check-tools
	@echo "Running baseline tests..."
	gotestsum --format=short -- -mod=mod -run=TestParser_Baselines ./internal/parser/
	gotestsum --format=short -- -mod=mod -run=TestTypeChecker_Baselines ./internal/typechecker/

test-baselines-update: tree-sitter check-tools
	@echo "Updating baseline tests..."
	UPDATE_BASELINES=1 gotestsum --format=short -- -mod=mod -run=TestParser_Baselines ./internal/parser/
	UPDATE_BASELINES=1 gotestsum --format=short -- -mod=mod -run=TestTypeChecker_Baselines ./internal/typechecker/

test-performance-baseline: tree-sitter check-tools
	@echo "Updating performance baselines..."
	UPDATE_PERFORMANCE_BASELINE=1 go test -mod=mod -bench=BenchmarkParser_Performance -benchmem ./internal/parser/
	UPDATE_PERFORMANCE_BASELINE=1 go test -mod=mod -bench=BenchmarkTypeChecker_Performance -benchmem ./internal/typechecker/

# Coverage reporting
test-coverage: tree-sitter check-tools
	@echo "Running tests with coverage..."
	gotestsum --format=short -- -mod=mod -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report saved to coverage.html"

test-coverage-report: test-coverage
	@echo "Coverage Summary:"
	go tool cover -func=coverage.out | grep total | awk '{print "Total Coverage: " $$3}'

# Integration testing
test-integration: tree-sitter check-tools
	@echo "Running integration tests..."
	gotestsum --format=short -- -mod=mod -tags=integration ./test/e2e/

# Performance optimization targets
optimize-performance: tree-sitter check-tools
	@echo "Running performance optimization analysis..."
	go run scripts/optimize_performance.go

profile-performance: tree-sitter check-tools
	@echo "Profiling performance bottlenecks..."
	go run scripts/profile_performance.go

# Complete test suite with all validations
test-all: tree-sitter check-tools
	@echo "Running complete test suite..."
	make test-baselines
	make test-performance
	make test-integration
	make test-coverage-report

# Run all tests (with tree-sitter support) - legacy compatibility
test-go: tree-sitter
	go test -mod=mod -v ./...

# Run tests without vendor (for better tree-sitter compatibility)
test-novendor: tree-sitter
	rm -rf vendor || true
	go test -mod=mod -v ./...

# Run specific component tests
test-scanner: tree-sitter
	gotestsum --format=short -- -mod=mod ./internal/scanner/...

test-parser: tree-sitter
	gotestsum --format=short -- -mod=mod ./internal/parser/...

test-ast: tree-sitter
	gotestsum --format=short -- -mod=mod ./internal/ast/... ./internal/astnav/...

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
	govulncheck ./...

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
		echo "Usage: make build-monitor-component COMPONENT=<pvm|pvx|pvi|psc>"; \
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
	cp $(BUILDDIR)/pvx $(GOPATH)/bin/pvx
	cp $(BUILDDIR)/pvi $(GOPATH)/bin/pvi
	cp $(BUILDDIR)/psc $(GOPATH)/bin/psc

# Cross-compile for all platforms (development)
cross-compile: tree-sitter
	@echo "Cross-compiling for all platforms..."
	mkdir -p $(BUILDDIR)/release

	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILDDIR)/release/pvm-linux-amd64 ./cmd/pvm
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILDDIR)/release/pvx-linux-amd64 ./cmd/pvx
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILDDIR)/release/pvi-linux-amd64 ./cmd/pvi

	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILDDIR)/release/pvm-darwin-amd64 ./cmd/pvm
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILDDIR)/release/pvx-darwin-amd64 ./cmd/pvx
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILDDIR)/release/pvi-darwin-amd64 ./cmd/pvi

	# macOS ARM64
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILDDIR)/release/pvm-darwin-arm64 ./cmd/pvm
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILDDIR)/release/pvx-darwin-arm64 ./cmd/pvx
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILDDIR)/release/pvi-darwin-arm64 ./cmd/pvi

	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILDDIR)/release/pvm-windows-amd64.exe ./cmd/pvm
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILDDIR)/release/pvx-windows-amd64.exe ./cmd/pvx
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILDDIR)/release/pvi-windows-amd64.exe ./cmd/pvi

	@echo "Cross-compilation complete. Binaries in $(BUILDDIR)/release/"

# Create release archives
release: cross-compile
	@echo "Creating release archives..."
	cd $(BUILDDIR)/release && \
	tar -czf pvm-linux-amd64.tar.gz pvm-linux-amd64 pvx-linux-amd64 pvi-linux-amd64 && \
	tar -czf pvm-darwin-amd64.tar.gz pvm-darwin-amd64 pvx-darwin-amd64 pvi-darwin-amd64 && \
	tar -czf pvm-darwin-arm64.tar.gz pvm-darwin-arm64 pvx-darwin-arm64 pvi-darwin-arm64 && \
	zip pvm-windows-amd64.zip pvm-windows-amd64.exe pvx-windows-amd64.exe pvi-windows-amd64.exe
	@echo "Release archives created in $(BUILDDIR)/release/"

# Performance optimization targets
test-performance: tree-sitter
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
