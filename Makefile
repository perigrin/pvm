.PHONY: all test clean install tree-sitter vendor cross-compile release

# Define binaries
BINARIES := pvm pvx pvi psc
BUILDDIR := build
LIBDIR := lib

all: $(BUILDDIR) $(LIBDIR) tree-sitter $(BINARIES)

$(BUILDDIR):
	mkdir -p $(BUILDDIR)

$(LIBDIR):
	mkdir -p $(LIBDIR)

# Ensure vendor directory exists
vendor:
	go mod vendor

# Build tree-sitter-perl library
tree-sitter: $(LIBDIR) vendor
	./bin/build_tree_sitter.sh

# Build rules for each binary
pvm: $(BUILDDIR)
	go build -o $(BUILDDIR)/pvm ./cmd/pvm

pvx: $(BUILDDIR)
	go build -o $(BUILDDIR)/pvx ./cmd/pvx

pvi: $(BUILDDIR)
	go build -o $(BUILDDIR)/pvi ./cmd/pvi

psc: $(BUILDDIR) tree-sitter
	CGO_CFLAGS="-I$(shell pwd)/include" go build -o $(BUILDDIR)/psc ./cmd/psc

# Run all tests (excluding tree-sitter dependent tests)
test:
	go test -v $(shell go list ./... | grep -v treesitter)

# Run PSC tests with proper CGO flags
test-psc: tree-sitter
	CGO_CFLAGS="-I$(shell pwd)/include" go test -v ./internal/parser/treesitter/...

# Clean build artifacts
clean:
	rm -rf $(BUILDDIR)
	rm -rf $(LIBDIR)

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
