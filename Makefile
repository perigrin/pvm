.PHONY: all test clean install tree-sitter

# Define binaries
BINARIES := pvm pvx pvi psc
BUILDDIR := build
LIBDIR := lib

all: $(BUILDDIR) $(LIBDIR) tree-sitter $(BINARIES)

$(BUILDDIR):
	mkdir -p $(BUILDDIR)

$(LIBDIR):
	mkdir -p $(LIBDIR)

# Build tree-sitter-perl library
tree-sitter: $(LIBDIR)
	./bin/build_tree_sitter.sh

# Build rules for each binary
pvm: $(BUILDDIR)
	go build -o $(BUILDDIR)/pvm ./cmd/pvm

pvx: $(BUILDDIR)
	go build -o $(BUILDDIR)/pvx ./cmd/pvx

pvi: $(BUILDDIR)
	go build -o $(BUILDDIR)/pvi ./cmd/pvi

psc: $(BUILDDIR) tree-sitter
	go build -o $(BUILDDIR)/psc ./cmd/psc

# Run all tests
test: tree-sitter
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf $(BUILDDIR)
	rm -rf $(LIBDIR)
	rm -rf vendor/tree-sitter-perl

# Install binaries
install: $(BINARIES)
	cp $(BUILDDIR)/pvm $(GOPATH)/bin/pvm
	cp $(BUILDDIR)/pvx $(GOPATH)/bin/pvx
	cp $(BUILDDIR)/pvi $(GOPATH)/bin/pvi
	cp $(BUILDDIR)/psc $(GOPATH)/bin/psc
