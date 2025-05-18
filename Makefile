.PHONY: all test clean install

# Define binaries
BINARIES := pvm pvx pvi psc
BUILDDIR := build

all: $(BUILDDIR) $(BINARIES)

$(BUILDDIR):
	mkdir -p $(BUILDDIR)

# Build rules for each binary
pvm: $(BUILDDIR)
	go build -o $(BUILDDIR)/pvm ./cmd/pvm

pvx: $(BUILDDIR)
	go build -o $(BUILDDIR)/pvx ./cmd/pvx

pvi: $(BUILDDIR)
	go build -o $(BUILDDIR)/pvi ./cmd/pvi

psc: $(BUILDDIR)
	go build -o $(BUILDDIR)/psc ./cmd/psc

# Run all tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf $(BUILDDIR)

# Install binaries
install: $(BINARIES)
	cp $(BUILDDIR)/pvm $(GOPATH)/bin/pvm
	cp $(BUILDDIR)/pvx $(GOPATH)/bin/pvx
	cp $(BUILDDIR)/pvi $(GOPATH)/bin/pvi
	cp $(BUILDDIR)/psc $(GOPATH)/bin/psc
