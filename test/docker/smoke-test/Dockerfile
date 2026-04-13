# Smoke-test container: simulates a fresh user system with bash, zsh, and
# fish installed alongside a minimal Perl toolchain. Accepts a pre-built pvm
# binary via COPY. No Go toolchain — this is NOT a unit-test container, it
# tests the user-facing journey against a release-style binary.
#
# Build + run is driven by ./run-smoke.sh from the repo root.

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash \
        zsh \
        fish \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Put pvm where fresh-install instructions tell users to put it.
# /usr/local/bin is on the default Debian PATH for bash, zsh, and fish.
COPY pvm /usr/local/bin/pvm
RUN chmod +x /usr/local/bin/pvm

# Per-shell journey scripts exercise the user-reported bug scenarios.
COPY journeys /smoke/journeys

# Offline-safe: MetaCPAN isn't reachable from CI in many environments and
# we don't need it for the smoke test (no real Perl installs happen).
ENV PVM_SKIP_NETWORK_CALLS=1

# Give each run its own scratch XDG dirs — mimics a fresh user account.
ENV HOME=/root \
    XDG_CONFIG_HOME=/root/.config \
    XDG_DATA_HOME=/root/.local/share \
    XDG_CACHE_HOME=/root/.cache \
    XDG_STATE_HOME=/root/.local/state

WORKDIR /smoke
