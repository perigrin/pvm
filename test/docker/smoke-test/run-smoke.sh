#!/usr/bin/env bash
# ABOUTME: Build the smoke-test image and run the three shell journeys
# ABOUTME: (bash, zsh, fish). Fails if any shell's journey fails.

set -eu

here="$(cd "$(dirname "$0")" && pwd)"
repo_root="$(cd "$here/../../.." && pwd)"

cd "$repo_root"

echo "==> Building pvm binary for the smoke-test image"
mkdir -p "$here/build"
(
    cd "$repo_root"
    # Build for Linux since the container runs Debian. CGO_ENABLED=0 matches
    # the Makefile's cross-compile convention and guarantees the binary runs
    # regardless of the host's libc.
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$here/build/pvm" ./cmd/pvm/
)

echo "==> Building smoke-test container"
# Copy the binary next to the Dockerfile so the build context is minimal.
cp "$here/build/pvm" "$here/pvm"
trap 'rm -f "$here/pvm"' EXIT
docker build -t pvm-smoke:latest "$here"

echo ""
failures=0

# Run the shell-agnostic suite once (from bash) before the three per-shell
# journeys. Cheaper than running it three times and keeps the per-shell
# journey files focused on integration-dependent workflows.
echo "==> Running core-commands suite (shell-agnostic)"
if docker run --rm pvm-smoke:latest bash "/smoke/journeys/core-commands.sh"; then
    echo "   ✓ core-commands suite passed"
else
    echo "   ✗ core-commands suite FAILED"
    failures=$((failures + 1))
fi
echo ""

for shell_script in bash:bash-journey.sh zsh:zsh-journey.sh fish:fish-journey.fish; do
    shell="${shell_script%%:*}"
    script="${shell_script#*:}"
    echo "==> Running $shell journey"
    if docker run --rm pvm-smoke:latest "$shell" "/smoke/journeys/$script"; then
        echo "   ✓ $shell journey passed"
    else
        echo "   ✗ $shell journey FAILED"
        failures=$((failures + 1))
    fi
    echo ""
done

if [ "$failures" -gt 0 ]; then
    echo "==> $failures suite(s) failed"
    exit 1
fi

echo "==> All smoke suites passed"
