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
	rm -f pvm-linux-amd64 pvm-linux-arm64 pvm-darwin-amd64 pvm-darwin-arm64 pvm-windows-amd64.exe
	go clean ./...

cross-compile:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pvm-linux-amd64 ./cmd/pvm/
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o pvm-linux-arm64 ./cmd/pvm/
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o pvm-darwin-arm64 ./cmd/pvm/
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o pvm-windows-amd64.exe ./cmd/pvm/
