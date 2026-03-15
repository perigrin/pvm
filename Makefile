.PHONY: all build test clean pvm pvx pm psc cross-compile

BINARIES = pvm pvx pm psc

all: build

build: $(BINARIES)

pvm:
	go build -o pvm ./cmd/pvm/

pvx:
	go build -o pvx ./cmd/pvx/

pm:
	go build -o pm ./cmd/pm/

psc:
	go build -o psc ./cmd/psc/

test:
	go test ./... -count=1

clean:
	rm -f $(BINARIES)
	rm -f pvm-linux-amd64 pvm-linux-arm64 pvm-darwin-amd64 pvm-darwin-arm64
	go clean ./...

cross-compile:
	GOOS=linux GOARCH=amd64 go build -o pvm-linux-amd64 ./cmd/pvm/
	GOOS=linux GOARCH=arm64 go build -o pvm-linux-arm64 ./cmd/pvm/
	GOOS=darwin GOARCH=amd64 go build -o pvm-darwin-amd64 ./cmd/pvm/
	GOOS=darwin GOARCH=arm64 go build -o pvm-darwin-arm64 ./cmd/pvm/
