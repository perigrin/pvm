.PHONY: all build test clean

all: build

build:
	go build ./...

test:
	go test ./... -count=1

clean:
	go clean ./...
