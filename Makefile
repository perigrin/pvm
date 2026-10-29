.PHONY: all build test smoke-test clean pvm pvx pm psc cross-compile fuzz-lexer fuzz-parser fuzz-incremental fuzz-chart-recognize chart-ratchet chart-tables

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

# The lexer fuzz target. Separate from `test` because -fuzz takes ONE package
# and runs until its budget is spent; it cannot live inside `go test ./...`.
#
# 1M executions is the M0 target (spec §7.6.2). Crashers land in
# internal/lexer/testdata/fuzz/ and are replayed by the ordinary suite
# afterwards, which is the regression half.
fuzz-lexer:
	go test ./internal/lexer/ -run '^$$' -fuzz '^FuzzLexer$$' -fuzztime 1000000x

# The parser fuzz target, same shape and same reasoning as fuzz-lexer.
# Crashers land in internal/parse/testdata/fuzz/ and are replayed by
# TestFuzzSeeds afterwards.
fuzz-parser:
	go test ./internal/parse/ -run '^$$' -fuzz '^FuzzParser$$' -fuzztime 1000000x

# The incremental re-parse property (spec §7.7): one fuzzer per edit shape,
# run one after the other because -fuzz takes one target. 10M executions
# each. Their seeds, the §7.7.3 edits among them, are replayed by `test`.
fuzz-incremental:
	go test ./internal/parse/ -run '^$$' -fuzz '^FuzzIncremental$$' -fuzztime 10000000x
	go test ./internal/parse/ -run '^$$' -fuzz '^FuzzIncrementalSequence$$' -fuzztime 10000000x

# The chart recognizer's fuzz target (internal/chart/recognize): it never
# panics, and acceptance agrees with and without the transition cache and
# lookahead. Seeds in internal/chart/recognize/testdata/fuzz/ replay in make
# test. FUZZTIME=10000x for a short run.
fuzz-chart-recognize: FUZZTIME ?= 1000000x
fuzz-chart-recognize:
	go test ./internal/chart/recognize/ -run '^$$' -fuzz '^FuzzRecognize$$' -fuzztime $(FUZZTIME)

# The chart spike's ratchet (internal/chart/README.md has its output
# contract). Mainline does not gate on it: `test` never runs it, and its CI
# job, chart-ratchet.yml, is advisory. It fails only on a regression against
# internal/chart/testdata/ratchet/baseline.txt.
chart-ratchet:
	go run ./internal/chart/cmd/ratchet

# The chart grammar's generated tables (internal/chart/README.md describes
# them). They are committed; the ratchet's tables-drift key holds them to
# what this writes.
chart-tables:
	go run ./internal/chart/grammar/cmd/gen

# Docker-based user-journey smoke test. Simulates a fresh install against
# bash/zsh/fish and asserts the original #432/#433 bug scenarios stay fixed.
# Requires docker (or a docker-compatible runtime).
smoke-test:
	./test/docker/smoke-test/run-smoke.sh

clean:
	rm -f $(BINARIES)
	rm -f pvm-linux-amd64 pvm-linux-arm64 pvm-darwin-amd64 pvm-darwin-arm64 pvm-windows-amd64.exe
	go clean ./...

cross-compile:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pvm-linux-amd64 ./cmd/pvm/
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o pvm-linux-arm64 ./cmd/pvm/
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o pvm-darwin-arm64 ./cmd/pvm/
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o pvm-windows-amd64.exe ./cmd/pvm/
