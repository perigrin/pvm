# Dockerfile for pvm shell-integration tests.
#
# Runs the test/e2e suite inside a container that has bash, zsh, AND fish
# installed so the cross-shell PATH-refresh tests don't skip the shells that
# aren't available on the host.
#
# Usage (from repo root):
#   docker build -t pvm-shell-ci -f test/docker/shell-integration/Dockerfile .
#   docker run --rm pvm-shell-ci

FROM golang:1.24-bookworm

# - bash: default in Debian
# - zsh:  for the zsh template test case
# - fish: for the fish template test case (the shell most likely to break
#         the POSIX-centric template)
# - build-essential, libssl-dev: needed by a few transitive test helpers
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash \
        zsh \
        fish \
        build-essential \
        libssl-dev \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# PVM_SKIP_NETWORK_CALLS prevents the resolver from fetching MetaCPAN during
# tests (it has a 72h cache but may still block the first run in a fresh
# container).
ENV PVM_SKIP_NETWORK_CALLS=1
# PVM_REQUIRE_ALL_SHELLS=1 makes the e2e tests fail (instead of silently
# skipping) if bash/zsh/fish are missing — the whole point of this image
# is that all three are installed; a broken image should surface loudly.
ENV PVM_REQUIRE_ALL_SHELLS=1

# Default to running the shell-integration subset; callers can override.
CMD ["go", "test", "-count=1", "-run", "TestPvmUseRefreshesPathAcrossShells|TestShellTemplates|TestDetectShell|TestGenerateShellUseSystem", "./internal/perl/...", "./test/e2e/..."]
